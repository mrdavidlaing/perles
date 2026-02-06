package cursor

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/zjrosen/perles/internal/orchestration/client"
)

// errTest is a sentinel error for testing
var errTest = errors.New("test error")

// =============================================================================
// Executable Discovery Tests (using ExecutableFinder)
// =============================================================================

func TestDefaultKnownPaths_ContainsExpectedPaths(t *testing.T) {
	require.Len(t, defaultKnownPaths, 3)
	require.Equal(t, "~/.local/bin/{name}", defaultKnownPaths[0], "First path should be ~/.local/bin/{name}")
	require.Equal(t, "/opt/homebrew/bin/{name}", defaultKnownPaths[1], "Second path should be /opt/homebrew/bin/{name}")
	require.Equal(t, "/usr/local/bin/{name}", defaultKnownPaths[2], "Third path should be /usr/local/bin/{name}")
}

func TestDefaultKnownPaths_PriorityOrder(t *testing.T) {
	// Verify paths are checked in priority order (most common install locations first)
	// ~/.local/bin is highest priority
	// /opt/homebrew/bin for Apple Silicon Macs
	// /usr/local/bin for Intel Macs and Linux
	finder := client.NewExecutableFinder("cursor-agent",
		client.WithKnownPaths(defaultKnownPaths...),
	)
	require.NotNil(t, finder)
}

func TestExecutableFinder_LocalBinPath(t *testing.T) {
	tempDir := t.TempDir()
	localBinDir := filepath.Join(tempDir, ".local", "bin")
	require.NoError(t, os.MkdirAll(localBinDir, 0755))

	execName := "cursor-agent"
	if os.PathSeparator == '\\' {
		execName = "cursor-agent.exe"
	}
	execPath := filepath.Join(localBinDir, execName)
	require.NoError(t, os.WriteFile(execPath, []byte("#!/bin/bash\necho test"), 0755))

	t.Setenv("HOME", tempDir)
	t.Setenv("USERPROFILE", tempDir)

	path, err := client.NewExecutableFinder("cursor-agent",
		client.WithKnownPaths(defaultKnownPaths...),
	).Find()
	require.NoError(t, err)
	require.Equal(t, execPath, path)
}

func TestExecutableFinder_PathFallback(t *testing.T) {
	t.Setenv("HOME", "/non-existent-path")
	t.Setenv("USERPROFILE", "/non-existent-path")

	path, err := client.NewExecutableFinder("cursor-agent",
		client.WithKnownPaths(defaultKnownPaths...),
	).Find()
	if err != nil {
		require.True(t, errors.Is(err, client.ErrExecutableNotFound))
		require.Contains(t, err.Error(), "cursor-agent")
		require.Contains(t, err.Error(), "PATH")
	} else {
		require.NotEmpty(t, path)
	}
}

func TestExecutableFinder_NotFound_ErrorContainsPaths(t *testing.T) {
	t.Setenv("HOME", "/non-existent-path-for-test")
	t.Setenv("USERPROFILE", "/non-existent-path-for-test")
	t.Setenv("PATH", "")

	path, err := client.NewExecutableFinder("cursor-agent-nonexistent-test-12345",
		client.WithKnownPaths(defaultKnownPaths...),
	).Find()
	require.Error(t, err)
	require.True(t, errors.Is(err, client.ErrExecutableNotFound))
	require.Empty(t, path)

	errMsg := err.Error()
	require.Contains(t, errMsg, "cursor-agent-nonexistent-test-12345", "Error should mention executable name")
	require.Contains(t, errMsg, "PATH", "Error should mention PATH was checked")
}

func TestDefaultKnownPaths_WindowsExeSuffix(t *testing.T) {
	for i, path := range defaultKnownPaths {
		require.Contains(t, path, "{name}",
			"Path %d (%s) should use {name} template for cross-platform .exe support", i, path)
	}

	for i, path := range defaultKnownPaths {
		require.NotContains(t, path, ".exe",
			"Path %d (%s) should not hardcode .exe - ExecutableFinder handles this", i, path)
		require.NotContains(t, path, "cursor-agent",
			"Path %d (%s) should use {name} template, not hardcoded name", i, path)
	}
}

// =============================================================================
// Lifecycle Tests - Process struct behavior without actual subprocess spawning
// =============================================================================

// newTestProcess creates a Process struct for testing without spawning a real subprocess.
func newTestProcess() *Process {
	ctx, cancel := context.WithCancel(context.Background())
	bp := client.NewBaseProcess(
		ctx,
		cancel,
		nil, // no cmd for test
		nil, // no stdout for test
		nil, // no stderr for test
		"/test/project",
		client.WithProviderName("cursor"),
		client.WithStderrCapture(true),
	)
	bp.SetSessionRef("cursor-test-session-12345")
	bp.SetStatus(client.StatusRunning)
	return &Process{BaseProcess: bp}
}

func TestProcess_ChannelBufferSizes(t *testing.T) {
	p := newTestProcess()

	require.Equal(t, 100, cap(p.EventsWritable()))
	require.Equal(t, 10, cap(p.ErrorsWritable()))
}

func TestProcess_StatusTransitions_PendingToRunning(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bp := client.NewBaseProcess(
		ctx,
		cancel,
		nil, nil, nil,
		"/test",
		client.WithProviderName("cursor"),
	)
	p := &Process{BaseProcess: bp}

	require.Equal(t, client.StatusPending, p.Status())
	require.False(t, p.IsRunning())

	p.SetStatus(client.StatusRunning)
	require.Equal(t, client.StatusRunning, p.Status())
	require.True(t, p.IsRunning())
}

func TestProcess_StatusTransitions_RunningToCompleted(t *testing.T) {
	p := newTestProcess()

	require.Equal(t, client.StatusRunning, p.Status())
	require.True(t, p.IsRunning())

	p.SetStatus(client.StatusCompleted)
	require.Equal(t, client.StatusCompleted, p.Status())
	require.False(t, p.IsRunning())
}

func TestProcess_StatusTransitions_RunningToFailed(t *testing.T) {
	p := newTestProcess()

	p.SetStatus(client.StatusFailed)
	require.Equal(t, client.StatusFailed, p.Status())
	require.False(t, p.IsRunning())
}

func TestProcess_StatusTransitions_RunningToCancelled(t *testing.T) {
	p := newTestProcess()

	err := p.Cancel()
	require.NoError(t, err)
	require.Equal(t, client.StatusCancelled, p.Status())
	require.False(t, p.IsRunning())
}

func TestProcess_Cancel_TerminatesAndSetsStatus(t *testing.T) {
	p := newTestProcess()

	require.Equal(t, client.StatusRunning, p.Status())

	err := p.Cancel()
	require.NoError(t, err)
	require.Equal(t, client.StatusCancelled, p.Status())

	select {
	case <-p.Context().Done():
		// Expected - context was cancelled
	default:
		require.Fail(t, "Context should be cancelled after Cancel()")
	}
}

func TestProcess_Cancel_RacePrevention(t *testing.T) {
	// Verifies Cancel() sets status BEFORE calling cancelFunc,
	// preventing race conditions with goroutines that check status.
	for i := 0; i < 100; i++ {
		ctx, cancel := context.WithCancel(context.Background())
		bp := client.NewBaseProcess(
			ctx,
			cancel,
			nil, nil, nil,
			"/test",
			client.WithProviderName("cursor"),
		)
		bp.SetStatus(client.StatusRunning)
		p := &Process{BaseProcess: bp}

		var observedStatus client.ProcessStatus
		var wg sync.WaitGroup
		wg.Add(1)

		go func() {
			defer wg.Done()
			<-p.Context().Done()
			observedStatus = p.Status()
		}()

		time.Sleep(time.Microsecond)

		err := p.Cancel()
		require.NoError(t, err)

		wg.Wait()

		require.Equal(t, client.StatusCancelled, observedStatus,
			"Goroutine should see StatusCancelled after context cancel (iteration %d)", i)
	}
}

func TestProcess_Cancel_DoesNotOverrideTerminalState(t *testing.T) {
	tests := []struct {
		name           string
		initialStatus  client.ProcessStatus
		expectedStatus client.ProcessStatus
	}{
		{
			name:           "does not override completed",
			initialStatus:  client.StatusCompleted,
			expectedStatus: client.StatusCompleted,
		},
		{
			name:           "does not override failed",
			initialStatus:  client.StatusFailed,
			expectedStatus: client.StatusFailed,
		},
		{
			name:           "does not override already cancelled",
			initialStatus:  client.StatusCancelled,
			expectedStatus: client.StatusCancelled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			bp := client.NewBaseProcess(
				ctx,
				cancel,
				nil, nil, nil,
				"/test",
				client.WithProviderName("cursor"),
			)
			bp.SetStatus(tt.initialStatus)
			p := &Process{BaseProcess: bp}

			err := p.Cancel()
			require.NoError(t, err)
			require.Equal(t, tt.expectedStatus, p.Status())
		})
	}
}

func TestProcess_ContextTimeout_TriggersCancellation(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	bp := client.NewBaseProcess(
		ctx,
		cancel,
		nil, nil, nil,
		"/test",
		client.WithProviderName("cursor"),
	)
	bp.SetStatus(client.StatusRunning)
	p := &Process{BaseProcess: bp}

	<-p.Context().Done()

	require.Equal(t, context.DeadlineExceeded, p.Context().Err())
}

// =============================================================================
// Process Property Tests
// =============================================================================

func TestProcess_SessionRef_ReturnsSessionID(t *testing.T) {
	p := newTestProcess()
	require.Equal(t, "cursor-test-session-12345", p.SessionRef())
}

func TestProcess_SessionRef_InitiallyEmpty(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bp := client.NewBaseProcess(
		ctx,
		cancel,
		nil, nil, nil,
		"/test/project",
		client.WithProviderName("cursor"),
	)
	bp.SetStatus(client.StatusRunning)
	p := &Process{BaseProcess: bp}

	require.Equal(t, "", p.SessionRef())
}

func TestProcess_SessionID_Convenience(t *testing.T) {
	p := newTestProcess()

	require.Equal(t, p.SessionRef(), p.SessionID())
	require.Equal(t, "cursor-test-session-12345", p.SessionID())
}

func TestProcess_WorkDir(t *testing.T) {
	p := newTestProcess()
	require.Equal(t, "/test/project", p.WorkDir())
}

func TestProcess_PID_NilProcess(t *testing.T) {
	p := newTestProcess()
	require.Equal(t, -1, p.PID())
}

// =============================================================================
// Channel Tests
// =============================================================================

func TestProcess_Wait_BlocksUntilCompletion(t *testing.T) {
	p := newTestProcess()

	p.WaitGroup().Add(1)

	done := make(chan bool)
	go func() {
		err := p.Wait()
		require.NoError(t, err)
		done <- true
	}()

	select {
	case <-done:
		require.Fail(t, "Wait should be blocking")
	case <-time.After(10 * time.Millisecond):
		// Expected - still waiting
	}

	p.WaitGroup().Done()

	select {
	case <-done:
		// Expected - Wait completed
	case <-time.After(time.Second):
		require.Fail(t, "Wait should have completed after wg.Done()")
	}
}

func TestProcess_SendError_NonBlocking(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	bp := client.NewBaseProcess(
		ctx,
		cancel,
		nil, nil, nil,
		"/test",
		client.WithProviderName("cursor"),
	)
	p := &Process{BaseProcess: bp}

	// Fill the channel (capacity is 10)
	for i := 0; i < 10; i++ {
		p.ErrorsWritable() <- errTest
	}

	// Channel is now full - SendError should not block
	done := make(chan bool)
	go func() {
		p.SendError(errors.New("overflow error"))
		done <- true
	}()

	select {
	case <-done:
		// Expected - SendError returned without blocking
	case <-time.After(100 * time.Millisecond):
		require.Fail(t, "SendError blocked on full channel - should have dropped error")
	}

	require.Len(t, p.ErrorsWritable(), 10)
}

func TestProcess_SendError_SuccessWhenSpaceAvailable(t *testing.T) {
	p := newTestProcess()

	p.SendError(errTest)

	select {
	case err := <-p.Errors():
		require.Equal(t, errTest, err)
	default:
		require.Fail(t, "Error should have been sent to channel")
	}
}

func TestProcess_EventsChannel(t *testing.T) {
	p := newTestProcess()

	eventsCh := p.Events()
	require.NotNil(t, eventsCh)

	go func() {
		p.EventsWritable() <- client.OutputEvent{Type: client.EventSystem, SubType: "init"}
	}()

	select {
	case event := <-eventsCh:
		require.Equal(t, client.EventSystem, event.Type)
		require.Equal(t, "init", event.SubType)
	case <-time.After(time.Second):
		require.Fail(t, "Timeout waiting for event")
	}
}

func TestProcess_ErrorsChannel(t *testing.T) {
	p := newTestProcess()

	errorsCh := p.Errors()
	require.NotNil(t, errorsCh)

	go func() {
		p.ErrorsWritable() <- errTest
	}()

	select {
	case err := <-errorsCh:
		require.Equal(t, errTest, err)
	case <-time.After(time.Second):
		require.Fail(t, "Timeout waiting for error")
	}
}

// =============================================================================
// Interface Compliance Tests
// =============================================================================

func TestProcess_ImplementsHeadlessProcess(t *testing.T) {
	var hp client.HeadlessProcess = newTestProcess()
	require.NotNil(t, hp)

	_ = hp.Events()
	_ = hp.Errors()
	_ = hp.SessionRef()
	_ = hp.Status()
	_ = hp.IsRunning()
	_ = hp.WorkDir()
	_ = hp.PID()
}

// =============================================================================
// Cursor-Specific: Session Extraction from Init Events
// =============================================================================

func TestExtractSession_FromInitEvent(t *testing.T) {
	event := client.OutputEvent{Type: client.EventSystem, SubType: "init"}
	rawLine := []byte(`{"type":"system","subtype":"init","session_id":"ses_cursor_abc123"}`)

	result := extractSession(event, rawLine)
	require.Equal(t, "ses_cursor_abc123", result)
}

func TestExtractSession_IgnoresNonInitEvents(t *testing.T) {
	tests := []struct {
		name    string
		event   client.OutputEvent
		rawLine []byte
	}{
		{
			name:    "assistant event",
			event:   client.OutputEvent{Type: client.EventAssistant},
			rawLine: []byte(`{"type":"assistant","session_id":"ses_123"}`),
		},
		{
			name:    "tool_use event",
			event:   client.OutputEvent{Type: client.EventToolUse},
			rawLine: []byte(`{"type":"tool_use","session_id":"ses_123"}`),
		},
		{
			name:    "result event",
			event:   client.OutputEvent{Type: client.EventResult},
			rawLine: []byte(`{"type":"result","session_id":"ses_123"}`),
		},
		{
			name:    "system non-init event",
			event:   client.OutputEvent{Type: client.EventSystem, SubType: "heartbeat"},
			rawLine: []byte(`{"type":"system","subtype":"heartbeat","session_id":"ses_123"}`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractSession(tt.event, tt.rawLine)
			require.Equal(t, "", result, "extractSession should only return session from init events")
		})
	}
}

func TestExtractSession_EmptySessionID(t *testing.T) {
	event := client.OutputEvent{Type: client.EventSystem, SubType: "init"}
	rawLine := []byte(`{"type":"system","subtype":"init","session_id":""}`)

	result := extractSession(event, rawLine)
	require.Equal(t, "", result)
}

func TestExtractSession_MissingSessionID(t *testing.T) {
	event := client.OutputEvent{Type: client.EventSystem, SubType: "init"}
	rawLine := []byte(`{"type":"system","subtype":"init"}`)

	result := extractSession(event, rawLine)
	require.Equal(t, "", result)
}

func TestExtractSession_InvalidJSON(t *testing.T) {
	event := client.OutputEvent{Type: client.EventSystem, SubType: "init"}
	rawLine := []byte(`not json`)

	result := extractSession(event, rawLine)
	require.Equal(t, "", result)
}

// =============================================================================
// Provider Configuration Tests
// =============================================================================

func TestProcess_StderrCapture_Enabled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bp := client.NewBaseProcess(
		ctx,
		cancel,
		nil, nil, nil,
		"/test",
		client.WithProviderName("cursor"),
		client.WithStderrCapture(true),
	)

	require.True(t, bp.CaptureStderr())
}

func TestProcess_ProviderName_IsCursor(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bp := client.NewBaseProcess(
		ctx,
		cancel,
		nil, nil, nil,
		"/test",
		client.WithProviderName("cursor"),
	)

	require.Equal(t, "cursor", bp.ProviderName())
}
