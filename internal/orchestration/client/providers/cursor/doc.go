// Package cursor provides a Go interface to headless Cursor Agent CLI sessions.
//
// Cursor is an AI-powered code editor that includes a CLI agent for
// non-interactive execution of AI-assisted coding tasks. This package
// implements the client.HeadlessClient interface to enable Cursor as a
// provider in the orchestration system.
//
// # Usage
//
// Import this package to register the Cursor client with the client registry:
//
//	import _ "github.com/zjrosen/perles/internal/orchestration/client/providers/cursor"
//
// Then create a client using the registry:
//
//	client, err := client.NewClient(client.ClientCursor)
//
// # CLI Requirements
//
// The "cursor-agent" command must be available in PATH. Install from:
// https://cursor.com/install
//
// # Headless Mode
//
// Cursor runs in headless mode using the --print flag with stream-json output:
//
//	cursor-agent --print --output-format stream-json --model composer-1 "your prompt here"
//
// Key flags:
//   - --print: Non-interactive mode for scripting/automation
//   - --output-format stream-json: Structured JSONL output for parsing
//   - --model: Model selection (e.g., composer-1)
//   - --resume: Resume existing session by ID
//   - --force: Allow direct file modifications without confirmation
//
// # System Prompt
//
// Cursor does not support --append-system-prompt. System prompts are
// prepended to the main prompt with a separator (same pattern as OpenCode).
//
// # MCP Configuration
//
// Cursor CLI does not accept --mcp-config as a flag. Instead, it reads MCP
// server configuration from .cursor/mcp.json in the project directory. Before
// spawning, this package writes the orchestration MCP server config to
// {workDir}/.cursor/mcp.json, merging with any existing user-defined servers.
package cursor
