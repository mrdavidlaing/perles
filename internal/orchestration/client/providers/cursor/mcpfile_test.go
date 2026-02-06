package cursor

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteMCPConfigFile(t *testing.T) {
	t.Run("creates .cursor/mcp.json in work dir", func(t *testing.T) {
		workDir := t.TempDir()
		mcpJSON := `{"mcpServers":{"perles-orchestrator":{"url":"http://localhost:9000/mcp"}}}`

		err := writeMCPConfigFile(workDir, mcpJSON)
		require.NoError(t, err)

		mcpPath := filepath.Join(workDir, ".cursor", "mcp.json")
		data, err := os.ReadFile(mcpPath)
		require.NoError(t, err)

		var parsed mcpFileConfig
		require.NoError(t, json.Unmarshal(data, &parsed))
		assert.Contains(t, parsed.MCPServers, "perles-orchestrator")
	})

	t.Run("merges with existing user config", func(t *testing.T) {
		workDir := t.TempDir()

		// Write pre-existing user config
		cursorDir := filepath.Join(workDir, ".cursor")
		require.NoError(t, os.MkdirAll(cursorDir, 0o755))
		existingConfig := `{
  "mcpServers": {
    "user-server": {"command": "my-server", "args": ["--flag"]},
    "another-server": {"url": "http://example.com/mcp"}
  }
}`
		require.NoError(t, os.WriteFile(filepath.Join(cursorDir, "mcp.json"), []byte(existingConfig), 0o644))

		// Write our MCP config
		mcpJSON := `{"mcpServers":{"perles-orchestrator":{"url":"http://localhost:9000/mcp"}}}`
		err := writeMCPConfigFile(workDir, mcpJSON)
		require.NoError(t, err)

		// Read back and verify merge
		data, err := os.ReadFile(filepath.Join(cursorDir, "mcp.json"))
		require.NoError(t, err)

		var parsed mcpFileConfig
		require.NoError(t, json.Unmarshal(data, &parsed))

		// All three servers should be present
		assert.Len(t, parsed.MCPServers, 3)
		assert.Contains(t, parsed.MCPServers, "user-server", "existing user server should be preserved")
		assert.Contains(t, parsed.MCPServers, "another-server", "existing user server should be preserved")
		assert.Contains(t, parsed.MCPServers, "perles-orchestrator", "our server should be added")
	})

	t.Run("overwrites malformed existing config", func(t *testing.T) {
		workDir := t.TempDir()

		// Write malformed file
		cursorDir := filepath.Join(workDir, ".cursor")
		require.NoError(t, os.MkdirAll(cursorDir, 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(cursorDir, "mcp.json"), []byte("not valid json"), 0o644))

		// Should succeed and overwrite
		mcpJSON := `{"mcpServers":{"perles-orchestrator":{"url":"http://localhost:9000/mcp"}}}`
		err := writeMCPConfigFile(workDir, mcpJSON)
		require.NoError(t, err)

		data, err := os.ReadFile(filepath.Join(cursorDir, "mcp.json"))
		require.NoError(t, err)

		var parsed mcpFileConfig
		require.NoError(t, json.Unmarshal(data, &parsed))
		assert.Len(t, parsed.MCPServers, 1)
		assert.Contains(t, parsed.MCPServers, "perles-orchestrator")
	})

	t.Run("no-op for empty workDir", func(t *testing.T) {
		err := writeMCPConfigFile("", `{"mcpServers":{}}`)
		require.NoError(t, err)
	})

	t.Run("no-op for empty mcpConfig", func(t *testing.T) {
		err := writeMCPConfigFile("/some/dir", "")
		require.NoError(t, err)
	})

	t.Run("returns error for invalid mcpConfig JSON", func(t *testing.T) {
		workDir := t.TempDir()
		err := writeMCPConfigFile(workDir, "not json")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "parsing MCP config")
	})

	t.Run("worker config merges alongside coordinator", func(t *testing.T) {
		workDir := t.TempDir()

		// First write coordinator config
		coordJSON := `{"mcpServers":{"perles-orchestrator":{"url":"http://localhost:9000/mcp"}}}`
		require.NoError(t, writeMCPConfigFile(workDir, coordJSON))

		// Then write worker config to same dir
		workerJSON := `{"mcpServers":{"perles-worker":{"url":"http://localhost:9000/worker/w1"}}}`
		require.NoError(t, writeMCPConfigFile(workDir, workerJSON))

		data, err := os.ReadFile(filepath.Join(workDir, ".cursor", "mcp.json"))
		require.NoError(t, err)

		var parsed mcpFileConfig
		require.NoError(t, json.Unmarshal(data, &parsed))
		assert.Len(t, parsed.MCPServers, 2)
		assert.Contains(t, parsed.MCPServers, "perles-orchestrator")
		assert.Contains(t, parsed.MCPServers, "perles-worker")
	})
}
