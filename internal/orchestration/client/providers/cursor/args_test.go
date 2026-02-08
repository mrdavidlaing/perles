package cursor

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildArgs(t *testing.T) {
	tests := []struct {
		name string
		cfg  Config
		want []string
	}{
		{
			name: "minimal config with prompt only",
			cfg: Config{
				Prompt: "hello world",
			},
			want: []string{"--print", "--output-format", "stream-json", "hello world"},
		},
		{
			name: "with model",
			cfg: Config{
				Prompt: "hello world",
				Model:  "composer-1",
			},
			want: []string{"--print", "--output-format", "stream-json", "--model", "composer-1", "hello world"},
		},
		{
			name: "resume session",
			cfg: Config{
				Prompt:    "follow up",
				SessionID: "ses_abc123",
			},
			want: []string{"--print", "--output-format", "stream-json", "--resume", "ses_abc123", "follow up"},
		},
		{
			name: "resume with model",
			cfg: Config{
				Prompt:    "continue work",
				SessionID: "ses_xyz789",
				Model:     "composer-1",
			},
			want: []string{"--print", "--output-format", "stream-json", "--resume", "ses_xyz789", "--model", "composer-1", "continue work"},
		},
		{
			name: "skip permissions maps to --force",
			cfg: Config{
				Prompt:          "do it",
				SkipPermissions: true,
			},
			want: []string{"--print", "--output-format", "stream-json", "--force", "do it"},
		},
		{
			name: "empty prompt omits prompt arg",
			cfg:  Config{},
			want: []string{"--print", "--output-format", "stream-json"},
		},
		{
			name: "MCPConfig adds --approve-mcps",
			cfg: Config{
				Prompt:    "do it",
				MCPConfig: `{"mcpServers":{"perles-orchestrator":{"url":"http://localhost:9000/mcp"}}}`,
			},
			want: []string{"--print", "--output-format", "stream-json", "--approve-mcps", "do it"},
		},
		{
			name: "full config",
			cfg: Config{
				Prompt:          "build it",
				Model:           "composer-1",
				SessionID:       "ses_full",
				SkipPermissions: true,
				MCPConfig:       `{"mcpServers":{}}`,
			},
			want: []string{
				"--print", "--output-format", "stream-json",
				"--resume", "ses_full",
				"--model", "composer-1",
				"--force",
				"--approve-mcps",
				"build it",
			},
		},
		{
			name: "prompt with special characters is preserved",
			cfg: Config{
				Prompt: `Create a function that checks --flag "value" and handles $variables`,
			},
			want: []string{"--print", "--output-format", "stream-json", `Create a function that checks --flag "value" and handles $variables`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildArgs(tt.cfg)
			require.Equal(t, tt.want, got)
		})
	}
}
