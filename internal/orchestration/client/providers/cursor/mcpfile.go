package cursor

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// mcpFileConfig mirrors the mcp.MCPConfig structure for reading/writing .cursor/mcp.json.
// We use a local type to avoid import cycles with the mcp package.
type mcpFileConfig struct {
	MCPServers map[string]json.RawMessage `json:"mcpServers"`
}

// writeMCPConfigFile writes the MCP server configuration to .cursor/mcp.json
// in the given work directory. If the file already exists, perles-managed
// servers (perles-orchestrator, perles-worker, perles-observer) are merged
// into the existing config to preserve user-defined servers.
func writeMCPConfigFile(workDir, mcpConfigJSON string) error {
	if workDir == "" || mcpConfigJSON == "" {
		return nil
	}

	// Parse the incoming config to extract our server entries
	var incoming mcpFileConfig
	if err := json.Unmarshal([]byte(mcpConfigJSON), &incoming); err != nil {
		return fmt.Errorf("parsing MCP config: %w", err)
	}

	cursorDir := filepath.Join(workDir, ".cursor")
	mcpPath := filepath.Join(cursorDir, "mcp.json")

	// Read existing config if it exists, so we can merge
	existing := mcpFileConfig{MCPServers: make(map[string]json.RawMessage)}
	if data, err := os.ReadFile(mcpPath); err == nil {
		// File exists — parse it for merging
		if err := json.Unmarshal(data, &existing); err != nil {
			// Existing file is malformed; overwrite it
			existing.MCPServers = make(map[string]json.RawMessage)
		}
	}

	// Merge: add/overwrite our server entries into the existing config
	for name, serverCfg := range incoming.MCPServers {
		existing.MCPServers[name] = serverCfg
	}

	// Write the merged config
	if err := os.MkdirAll(cursorDir, 0o755); err != nil {
		return fmt.Errorf("creating .cursor directory: %w", err)
	}

	data, err := json.MarshalIndent(existing, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling merged MCP config: %w", err)
	}

	if err := os.WriteFile(mcpPath, data, 0o644); err != nil {
		return fmt.Errorf("writing .cursor/mcp.json: %w", err)
	}

	return nil
}
