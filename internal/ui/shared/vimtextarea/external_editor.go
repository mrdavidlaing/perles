package vimtextarea

import (
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// ExternalEditorFinishedMsg is sent when the external editor closes.
// Contains the edited content from the temp file.
type ExternalEditorFinishedMsg struct {
	Content string
	Err     error
}

// OpenExternalEditorCommand opens the current content in $EDITOR.
// When the editor closes, the content is read back and sent as ExternalEditorFinishedMsg.
type OpenExternalEditorCommand struct {
	MotionBase
}

// Execute returns PassThrough - actual execution happens via tea.ExecProcess.
// The command's work is done in CreateEditorCmd which returns a tea.Cmd.
func (c *OpenExternalEditorCommand) Execute(m *Model) ExecuteResult {
	return Executed
}

// Keys returns the trigger keys for this command.
func (c *OpenExternalEditorCommand) Keys() []string {
	return []string{"<ctrl+g>"}
}

// Mode returns the mode this command operates in.
func (c *OpenExternalEditorCommand) Mode() Mode {
	return ModeNormal
}

// ID returns the hierarchical identifier for this command.
func (c *OpenExternalEditorCommand) ID() string {
	return "external.editor"
}

// CreateEditorCmd creates the tea.Cmd that opens the external editor.
// This must be called separately because it needs the current content.
func CreateEditorCmd(content string) tea.Cmd {
	return func() tea.Msg {
		// Get editor from environment (VISUAL > EDITOR > vi)
		editor := os.Getenv("VISUAL")
		if editor == "" {
			editor = os.Getenv("EDITOR")
		}
		if editor == "" {
			editor = "vi"
		}

		// Create temp file with current content
		tmpFile, err := os.CreateTemp("", "perles-edit-*.md")
		if err != nil {
			return ExternalEditorFinishedMsg{Err: err}
		}
		tmpPath := tmpFile.Name()

		// Write content and close
		if _, err := tmpFile.WriteString(content); err != nil {
			os.Remove(tmpPath)
			return ExternalEditorFinishedMsg{Err: err}
		}
		tmpFile.Close()

		// Create the editor command
		cmd := exec.Command(editor, tmpPath)

		// Return an ExecProcess message that will be handled by Bubble Tea
		// We wrap this in a function that returns the final message after exec
		return ExternalEditorExecMsg{
			cmd:     cmd,
			tmpPath: tmpPath,
		}
	}
}

// ExternalEditorExecMsg is sent when the external editor command is ready to execute.
// The parent component handles this via tea.ExecProcess by calling ExecCmd().
type ExternalEditorExecMsg struct {
	cmd     *exec.Cmd
	tmpPath string
}

// ExecCmd returns the tea.ExecProcess command and cleanup function.
// Call this from the parent Update when receiving ExternalEditorExecMsg.
func (msg ExternalEditorExecMsg) ExecCmd() tea.Cmd {
	return tea.ExecProcess(msg.cmd, func(err error) tea.Msg {
		// Cleanup: always remove temp file
		defer os.Remove(msg.tmpPath)

		if err != nil {
			return ExternalEditorFinishedMsg{Err: err}
		}

		// Read back content
		content, readErr := os.ReadFile(msg.tmpPath)
		if readErr != nil {
			return ExternalEditorFinishedMsg{Err: readErr}
		}

		// Trim all trailing newlines that editors like vim add on save
		text := strings.TrimRight(string(content), "\n")

		return ExternalEditorFinishedMsg{Content: text}
	})
}
