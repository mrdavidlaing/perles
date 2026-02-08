//go:build cursor_integration

package cursor

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/zjrosen/perles/internal/orchestration/client"
)

const cursorIntegrationModel = "composer-1"

// skipIfCursorNotAvailable skips the test if cursor-agent CLI is not installed.
func skipIfCursorNotAvailable(t *testing.T) {
	t.Helper()
	_, err := client.NewExecutableFinder("cursor-agent",
		client.WithKnownPaths(defaultKnownPaths...),
	).Find()
	if err != nil {
		t.Skip("cursor-agent CLI not available, skipping integration test")
	}
}

// skipIfCursorNotConfigured skips the test if cursor-agent is not properly configured.
func skipIfCursorNotConfigured(t *testing.T) {
	t.Helper()
	path, err := client.NewExecutableFinder("cursor-agent",
		client.WithKnownPaths(defaultKnownPaths...),
	).Find()
	if err != nil {
		t.Skip("cursor-agent CLI not available")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, path, "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		out := string(output)
		if strings.Contains(out, "unknown option") ||
			strings.Contains(out, "unknown flag") ||
			strings.Contains(out, "flag provided but not defined") {
			return
		}
		t.Skipf("cursor-agent CLI not properly configured: %v", err)
	}
}

// =============================================================================
// Integration Tests - Require cursor-agent CLI
// =============================================================================

// TestIntegration_Spawn_ReceivesInitEvent tests that spawning a Cursor process
// receives an init event with session ID.
func TestIntegration_Spawn_ReceivesInitEvent(t *testing.T) {
	skipIfCursorNotAvailable(t)
	skipIfCursorNotConfigured(t)

	workDir := t.TempDir()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cfg := Config{
		WorkDir: workDir,
		Prompt:  "Say 'test' and nothing else",
		Model:   cursorIntegrationModel,
		Timeout: 30 * time.Second,
	}

	process, err := Spawn(ctx, cfg)
	require.NoError(t, err, "Failed to spawn cursor-agent process")
	require.NotNil(t, process)

	defer func() {
		_ = process.Cancel()
		_ = process.Wait()
	}()

	require.Equal(t, client.StatusRunning, process.Status())
	require.True(t, process.IsRunning())
	require.Greater(t, process.PID(), 0, "Process should have a valid PID")

	eventCh := process.Events()
	foundInit := false
	timeout := time.After(25 * time.Second)

	for !foundInit {
		select {
		case event, ok := <-eventCh:
			if !ok {
				if !foundInit {
					t.Fatal("Events channel closed without receiving init event")
				}
				return
			}
			t.Logf("Received event: type=%s subtype=%s", event.Type, event.SubType)
			if event.Type == client.EventSystem && event.SubType == "init" {
				foundInit = true
				sessionID := process.SessionRef()
				t.Logf("Init event received, session ID: %s", sessionID)
			}
		case <-timeout:
			if !foundInit {
				t.Fatal("Timeout waiting for init event")
			}
			return
		case <-ctx.Done():
			t.Fatal("Context cancelled while waiting for init event")
		}
	}

	require.True(t, foundInit, "Should have received init event")
}

// TestIntegration_Spawn_ReceivesAssistantEvents tests that the process receives
// assistant response events.
func TestIntegration_Spawn_ReceivesAssistantEvents(t *testing.T) {
	skipIfCursorNotAvailable(t)
	skipIfCursorNotConfigured(t)

	workDir := t.TempDir()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	cfg := Config{
		WorkDir: workDir,
		Prompt:  "Reply with exactly one word: hello",
		Model:   cursorIntegrationModel,
		Timeout: 60 * time.Second,
	}

	process, err := Spawn(ctx, cfg)
	require.NoError(t, err)
	defer func() {
		_ = process.Cancel()
		_ = process.Wait()
	}()

	eventCh := process.Events()
	var receivedEvents []client.OutputEvent
	timeout := time.After(55 * time.Second)

eventLoop:
	for {
		select {
		case event, ok := <-eventCh:
			if !ok {
				break eventLoop
			}
			receivedEvents = append(receivedEvents, event)
			t.Logf("Event: type=%s subtype=%s", event.Type, event.SubType)

			if event.Type == client.EventResult {
				break eventLoop
			}
		case <-timeout:
			break eventLoop
		case <-ctx.Done():
			t.Fatal("Context cancelled")
		}
	}

	_ = process.Wait()

	require.NotEmpty(t, receivedEvents, "Should have received at least one event")
	t.Logf("Process completed with status: %s", process.Status())
	t.Logf("Total events received: %d", len(receivedEvents))
}

// TestIntegration_ProcessCompletion tests that the process completes gracefully.
func TestIntegration_ProcessCompletion(t *testing.T) {
	skipIfCursorNotAvailable(t)
	skipIfCursorNotConfigured(t)

	workDir := t.TempDir()

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	cfg := Config{
		WorkDir: workDir,
		Prompt:  "Say 'done'",
		Model:   cursorIntegrationModel,
		Timeout: 45 * time.Second,
	}

	process, err := Spawn(ctx, cfg)
	require.NoError(t, err)

	go func() {
		for range process.Events() {
		}
	}()

	err = process.Wait()
	require.NoError(t, err)

	status := process.Status()
	require.True(t, status.IsTerminal(), "Process should be in terminal state, got: %s", status)
}

// TestIntegration_Cancel_TerminatesProcess tests that Cancel() properly terminates
// the running process.
func TestIntegration_Cancel_TerminatesProcess(t *testing.T) {
	skipIfCursorNotAvailable(t)
	skipIfCursorNotConfigured(t)

	workDir := t.TempDir()

	ctx := context.Background()

	cfg := Config{
		WorkDir: workDir,
		Prompt:  "Count from 1 to 1000 very slowly, one number per line",
		Model:   cursorIntegrationModel,
	}

	process, err := Spawn(ctx, cfg)
	require.NoError(t, err)
	require.True(t, process.IsRunning())

	time.Sleep(500 * time.Millisecond)

	err = process.Cancel()
	require.NoError(t, err)

	_ = process.Wait()

	require.Equal(t, client.StatusCancelled, process.Status())
	require.False(t, process.IsRunning())
}

// TestIntegration_MCPConfig_Written tests that MCP configuration is properly
// written to .cursor/mcp.json when provided.
func TestIntegration_MCPConfig_Written(t *testing.T) {
	skipIfCursorNotAvailable(t)

	workDir := t.TempDir()

	mcpConfig := `{
  "mcpServers": {
    "perles-orchestrator": {
      "url": "http://localhost:9999/mcp"
    }
  }
}`

	err := writeMCPConfigFile(workDir, mcpConfig)
	require.NoError(t, err)

	configPath := filepath.Join(workDir, ".cursor", "mcp.json")
	data, err := os.ReadFile(configPath)
	require.NoError(t, err)

	require.Contains(t, string(data), "perles-orchestrator")
	require.Contains(t, string(data), "http://localhost:9999/mcp")
}

// TestIntegration_SessionResume tests resuming an existing session.
func TestIntegration_SessionResume(t *testing.T) {
	skipIfCursorNotAvailable(t)
	skipIfCursorNotConfigured(t)

	workDir := t.TempDir()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	cfg := Config{
		WorkDir: workDir,
		Prompt:  "Say 'first message'",
		Model:   cursorIntegrationModel,
		Timeout: 30 * time.Second,
	}

	process, err := Spawn(ctx, cfg)
	require.NoError(t, err)

	eventCh := process.Events()
	var sessionID string
	timeout := time.After(25 * time.Second)

waitForInit:
	for {
		select {
		case event, ok := <-eventCh:
			if !ok {
				break waitForInit
			}
			if event.Type == client.EventSystem && event.SubType == "init" {
				sessionID = process.SessionRef()
				if sessionID == "" && event.SessionID != "" {
					sessionID = event.SessionID
				}
				break waitForInit
			}
		case <-timeout:
			break waitForInit
		}
	}

	_ = process.Cancel()
	_ = process.Wait()

	if sessionID == "" {
		t.Skip("Could not extract session ID for resume test")
	}

	t.Logf("First session completed with ID: %s", sessionID)

	ctx2, cancel2 := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel2()

	resumeCfg := Config{
		WorkDir:   workDir,
		Prompt:    "Say 'resumed'",
		Model:     cursorIntegrationModel,
		SessionID: sessionID,
		Timeout:   30 * time.Second,
	}

	resumeProcess, err := Resume(ctx2, sessionID, resumeCfg)
	require.NoError(t, err)
	defer func() {
		_ = resumeProcess.Cancel()
		_ = resumeProcess.Wait()
	}()

	require.True(t, resumeProcess.IsRunning())
	t.Logf("Resume process started with PID: %d", resumeProcess.PID())

	resumeEventCh := resumeProcess.Events()
	timeout2 := time.After(25 * time.Second)

drainLoop:
	for {
		select {
		case event, ok := <-resumeEventCh:
			if !ok {
				break drainLoop
			}
			t.Logf("Resume event: type=%s subtype=%s", event.Type, event.SubType)
			if event.Type == client.EventResult {
				break drainLoop
			}
		case <-timeout2:
			break drainLoop
		}
	}

	_ = resumeProcess.Wait()
	t.Logf("Resume process completed with status: %s", resumeProcess.Status())
}

// TestIntegration_WorkDir_Respected tests that the working directory is properly
// set for the spawned process.
func TestIntegration_WorkDir_Respected(t *testing.T) {
	skipIfCursorNotAvailable(t)
	skipIfCursorNotConfigured(t)

	workDir := t.TempDir()

	testFile := filepath.Join(workDir, "test-marker.txt")
	err := os.WriteFile(testFile, []byte("integration test marker"), 0o644)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cfg := Config{
		WorkDir: workDir,
		Prompt:  "test",
		Model:   cursorIntegrationModel,
	}

	process, err := Spawn(ctx, cfg)
	require.NoError(t, err)

	require.Equal(t, workDir, process.WorkDir())

	_ = process.Cancel()
	_ = process.Wait()
}
