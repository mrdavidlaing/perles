package editor

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOpenCmd_ReturnsExecMsg(t *testing.T) {
	// Set a known editor
	originalVisual := os.Getenv("VISUAL")
	originalEditor := os.Getenv("EDITOR")
	os.Setenv("VISUAL", "")
	os.Setenv("EDITOR", "cat") // Simple command that exists everywhere
	defer func() {
		os.Setenv("VISUAL", originalVisual)
		os.Setenv("EDITOR", originalEditor)
	}()

	cmd := OpenCmd("test content")
	require.NotNil(t, cmd)

	// Execute the command to get the message
	msg := cmd()

	// Should return ExecMsg (not FinishedMsg with error)
	execMsg, ok := msg.(ExecMsg)
	require.True(t, ok, "expected ExecMsg, got %T", msg)
	require.NotEmpty(t, execMsg.tmpPath)

	// Verify temp file was created with content
	content, err := os.ReadFile(execMsg.tmpPath)
	require.NoError(t, err)
	require.Equal(t, "test content", string(content))

	// Cleanup
	os.Remove(execMsg.tmpPath)
}

func TestOpenCmd_UsesVISUALFirst(t *testing.T) {
	originalVisual := os.Getenv("VISUAL")
	originalEditor := os.Getenv("EDITOR")
	os.Setenv("VISUAL", "myvisual")
	os.Setenv("EDITOR", "myeditor")
	defer func() {
		os.Setenv("VISUAL", originalVisual)
		os.Setenv("EDITOR", originalEditor)
	}()

	cmd := OpenCmd("test")
	msg := cmd()

	execMsg, ok := msg.(ExecMsg)
	require.True(t, ok)
	require.Equal(t, "myvisual", execMsg.cmd.Path)

	// Cleanup
	os.Remove(execMsg.tmpPath)
}

func TestOpenCmd_FallsBackToEDITOR(t *testing.T) {
	originalVisual := os.Getenv("VISUAL")
	originalEditor := os.Getenv("EDITOR")
	os.Setenv("VISUAL", "")
	os.Setenv("EDITOR", "myeditor")
	defer func() {
		os.Setenv("VISUAL", originalVisual)
		os.Setenv("EDITOR", originalEditor)
	}()

	cmd := OpenCmd("test")
	msg := cmd()

	execMsg, ok := msg.(ExecMsg)
	require.True(t, ok)
	require.Equal(t, "myeditor", execMsg.cmd.Path)

	// Cleanup
	os.Remove(execMsg.tmpPath)
}

func TestOpenCmd_FallsBackToVi(t *testing.T) {
	originalVisual := os.Getenv("VISUAL")
	originalEditor := os.Getenv("EDITOR")
	os.Setenv("VISUAL", "")
	os.Setenv("EDITOR", "")
	defer func() {
		os.Setenv("VISUAL", originalVisual)
		os.Setenv("EDITOR", originalEditor)
	}()

	cmd := OpenCmd("test")
	msg := cmd()

	execMsg, ok := msg.(ExecMsg)
	require.True(t, ok)
	// exec.Command resolves to full path, so check it ends with "vi"
	require.Contains(t, execMsg.cmd.Path, "vi")

	// Cleanup
	os.Remove(execMsg.tmpPath)
}

func TestOpenCmd_TempFileContainsContent(t *testing.T) {
	originalEditor := os.Getenv("EDITOR")
	os.Setenv("EDITOR", "cat")
	defer os.Setenv("EDITOR", originalEditor)

	content := "line 1\nline 2\nline 3"
	cmd := OpenCmd(content)
	msg := cmd()

	execMsg, ok := msg.(ExecMsg)
	require.True(t, ok)

	// Read temp file
	fileContent, err := os.ReadFile(execMsg.tmpPath)
	require.NoError(t, err)
	require.Equal(t, content, string(fileContent))

	// Cleanup
	os.Remove(execMsg.tmpPath)
}

func TestFinishedMsg_Fields(t *testing.T) {
	msg := FinishedMsg{
		Content: "edited content",
		Err:     nil,
	}
	require.Equal(t, "edited content", msg.Content)
	require.NoError(t, msg.Err)
}
