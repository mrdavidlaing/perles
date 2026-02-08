package cursor

// buildArgs constructs command line arguments for the Cursor Agent CLI.
//
// For new sessions:
//
//	cursor-agent --print --output-format stream-json --model <model> "prompt"
//
// For resume sessions:
//
//	cursor-agent --print --output-format stream-json --resume <id> --model <model> "prompt"
//
// Note: Cursor CLI does not support --append-system-prompt, --allowed-tools,
// --disallowed-tools, or --mcp-config. System prompt is prepended to the
// main prompt in configFromClient instead. MCP config is written to
// .cursor/mcp.json in the work directory before spawning.
func buildArgs(cfg Config) []string {
	args := []string{
		"--print",
		"--output-format", "stream-json",
	}

	// Session resume flag
	if cfg.SessionID != "" {
		args = append(args, "--resume", cfg.SessionID)
	}

	// Model selection
	if cfg.Model != "" {
		args = append(args, "--model", cfg.Model)
	}

	// Skip permissions maps to --force for direct file modifications
	if cfg.SkipPermissions {
		args = append(args, "--force")
	}

	// Auto-approve MCP servers when MCP config is provided.
	// Without this, cursor-agent silently skips unapproved servers.
	if cfg.MCPConfig != "" {
		args = append(args, "--approve-mcps")
	}

	// Prompt as final positional argument
	if cfg.Prompt != "" {
		args = append(args, cfg.Prompt)
	}

	return args
}
