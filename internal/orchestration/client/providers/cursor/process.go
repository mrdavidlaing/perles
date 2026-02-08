package cursor

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/zjrosen/perles/internal/log"
	"github.com/zjrosen/perles/internal/orchestration/client"
)

// defaultKnownPaths defines the priority-ordered paths to check for the cursor-agent executable.
// These are checked before falling back to PATH lookup.
var defaultKnownPaths = []string{
	"~/.local/bin/{name}",      // Common binary location
	"/opt/homebrew/bin/{name}", // Apple Silicon Mac (Homebrew)
	"/usr/local/bin/{name}",    // Intel Mac / Linux
}

// Process represents a headless Cursor Agent CLI process.
// Process implements client.HeadlessProcess by embedding BaseProcess.
type Process struct {
	*client.BaseProcess
}

// extractSession extracts the session ID from an init event.
func extractSession(event client.OutputEvent, rawLine []byte) string {
	if event.Type == client.EventSystem && event.SubType == "init" {
		var initData struct {
			SessionID string `json:"session_id"`
		}
		if err := json.Unmarshal(rawLine, &initData); err == nil && initData.SessionID != "" {
			return initData.SessionID
		}
	}
	return ""
}

// Spawn creates and starts a new headless Cursor process.
// Context is used for cancellation and timeout control.
func Spawn(ctx context.Context, cfg Config) (*Process, error) {
	return spawnProcess(ctx, cfg)
}

// Resume continues an existing Cursor session using --resume flag.
func Resume(ctx context.Context, sessionID string, cfg Config) (*Process, error) {
	cfg.SessionID = sessionID
	return spawnProcess(ctx, cfg)
}

// spawnProcess is the internal implementation for both Spawn and Resume.
func spawnProcess(ctx context.Context, cfg Config) (*Process, error) {
	// Write .cursor/mcp.json if MCP config is provided.
	// Cursor CLI reads MCP server configuration from this file (not from CLI flags).
	if cfg.MCPConfig != "" {
		log.Debug(log.CatOrch, "writing MCP config to .cursor/mcp.json",
			"subsystem", "cursor", "workDir", cfg.WorkDir)
		if err := writeMCPConfigFile(cfg.WorkDir, cfg.MCPConfig); err != nil {
			return nil, fmt.Errorf("cursor: writing MCP config: %w", err)
		}
	}

	// Find the cursor-agent executable using ExecutableFinder
	execPath, err := client.NewExecutableFinder("cursor-agent",
		client.WithKnownPaths(defaultKnownPaths...),
	).Find()
	if err != nil {
		return nil, err
	}
	log.Debug(log.CatOrch, "found cursor-agent executable",
		"subsystem", "cursor", "path", execPath)

	args := buildArgs(cfg)

	// Build environment variables (BEADS_DIR if set)
	env := client.BuildEnvVars(client.Config{BeadsDir: cfg.BeadsDir})

	log.Debug(log.CatOrch, "spawning cursor-agent process",
		"subsystem", "cursor", "workDir", cfg.WorkDir,
		"model", cfg.Model, "sessionID", cfg.SessionID)

	base, err := client.NewSpawnBuilder(ctx).
		WithExecutable(execPath, args).
		WithWorkDir(cfg.WorkDir).
		WithSessionRef(cfg.SessionID).
		WithTimeout(cfg.Timeout).
		WithParser(NewParser()).
		WithSessionExtractor(extractSession).
		WithStderrCapture(true).
		WithProviderName("cursor").
		WithEnv(env).
		Build()
	if err != nil {
		return nil, fmt.Errorf("cursor: %w", err)
	}

	return &Process{BaseProcess: base}, nil
}

// SessionID returns the session ID (may be empty until init event is received).
// This is a convenience method that wraps SessionRef for backwards compatibility.
func (p *Process) SessionID() string {
	return p.SessionRef()
}

// Ensure Process implements client.HeadlessProcess at compile time.
var _ client.HeadlessProcess = (*Process)(nil)
