package vimtextarea

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

// errClipboard is used in clipboard error tests
var errClipboard = errors.New("clipboard unavailable")

// ============================================================================
// YankLineCommand Tests (yy)
// ============================================================================

// TestYankLineCommand_Execute tests yanking entire current line
func TestYankLineCommand_Execute(t *testing.T) {
	m := newTestModelWithContent("hello world")

	cmd := &YankLineCommand{}
	result := cmd.Execute(m)

	require.Equal(t, Executed, result)
	require.Equal(t, "hello world", m.lastYankedText)
	require.True(t, m.lastYankWasLinewise, "yy should set lastYankWasLinewise = true")
}

// TestYankLineCommand_SetsLinewiseTrue verifies lastYankWasLinewise is true
func TestYankLineCommand_SetsLinewiseTrue(t *testing.T) {
	m := newTestModelWithContent("test line")
	m.lastYankWasLinewise = false // Start with false to verify it's set

	cmd := &YankLineCommand{}
	cmd.Execute(m)

	require.True(t, m.lastYankWasLinewise)
}

// TestYankLineCommand_EmptyLine tests yanking an empty line
func TestYankLineCommand_EmptyLine(t *testing.T) {
	m := newTestModelWithContent("")

	cmd := &YankLineCommand{}
	result := cmd.Execute(m)

	require.Equal(t, Executed, result)
	require.Equal(t, "", m.lastYankedText)
	require.True(t, m.lastYankWasLinewise)
}

// TestYankLineCommand_CursorUnchanged tests that cursor doesn't move after yy
func TestYankLineCommand_CursorUnchanged(t *testing.T) {
	m := newTestModelWithContent("hello world")
	m.cursorRow = 0
	m.cursorCol = 5

	cmd := &YankLineCommand{}
	cmd.Execute(m)

	require.Equal(t, 0, m.cursorRow, "cursor row should not change")
	require.Equal(t, 5, m.cursorCol, "cursor col should not change")
}

// TestYankLineCommand_MultiLine tests yanking from middle of multi-line content
func TestYankLineCommand_MultiLine(t *testing.T) {
	m := newTestModelWithContent("line1", "line2", "line3")
	m.cursorRow = 1 // On "line2"

	cmd := &YankLineCommand{}
	cmd.Execute(m)

	require.Equal(t, "line2", m.lastYankedText)
}

// TestYankLineCommand_Keys tests command keys
func TestYankLineCommand_Keys(t *testing.T) {
	cmd := &YankLineCommand{}
	require.Equal(t, []string{"yy"}, cmd.Keys())
}

// TestYankLineCommand_Mode tests command mode
func TestYankLineCommand_Mode(t *testing.T) {
	cmd := &YankLineCommand{}
	require.Equal(t, ModeNormal, cmd.Mode())
}

// TestYankLineCommand_ID tests command ID
func TestYankLineCommand_ID(t *testing.T) {
	cmd := &YankLineCommand{}
	require.Equal(t, "yank.line", cmd.ID())
}

// TestYankLineCommand_IsUndoable tests yank is not undoable
func TestYankLineCommand_IsUndoable(t *testing.T) {
	cmd := &YankLineCommand{}
	require.False(t, cmd.IsUndoable(), "yank should not be undoable")
}

// TestYankLineCommand_ChangesContent tests yank doesn't change content
func TestYankLineCommand_ChangesContent(t *testing.T) {
	cmd := &YankLineCommand{}
	require.False(t, cmd.ChangesContent(), "yank should not change content")
}

// ============================================================================
// YankWordCommand Tests (yw)
// ============================================================================

// TestYankWordCommand_Execute tests yanking from cursor to word boundary
func TestYankWordCommand_Execute(t *testing.T) {
	m := newTestModelWithContent("hello world")
	m.cursorCol = 0 // At 'h'

	cmd := &YankWordCommand{}
	result := cmd.Execute(m)

	require.Equal(t, Executed, result)
	require.Equal(t, "hello ", m.lastYankedText)
	require.False(t, m.lastYankWasLinewise, "yw should set lastYankWasLinewise = false")
}

// TestYankWordCommand_SetsLinewiseFalse verifies lastYankWasLinewise is false
func TestYankWordCommand_SetsLinewiseFalse(t *testing.T) {
	m := newTestModelWithContent("hello world")
	m.lastYankWasLinewise = true // Start with true to verify it's set to false

	cmd := &YankWordCommand{}
	cmd.Execute(m)

	require.False(t, m.lastYankWasLinewise)
}

// TestYankWordCommand_AtEndOfLine tests yw at end of line yanks remaining characters
func TestYankWordCommand_AtEndOfLine(t *testing.T) {
	m := newTestModelWithContent("hello world")
	m.cursorCol = 6 // At 'w' in "world"

	cmd := &YankWordCommand{}
	result := cmd.Execute(m)

	require.Equal(t, Executed, result)
	require.Equal(t, "world", m.lastYankedText) // Should yank to end of line
}

// TestYankWordCommand_MiddleOfWord tests yw from middle of word
func TestYankWordCommand_MiddleOfWord(t *testing.T) {
	m := newTestModelWithContent("hello world")
	m.cursorCol = 2 // At 'l' in "hello"

	cmd := &YankWordCommand{}
	cmd.Execute(m)

	require.Equal(t, "llo ", m.lastYankedText) // Should yank from 'l' to start of next word
}

// TestYankWordCommand_EmptyLine tests yw on empty line
func TestYankWordCommand_EmptyLine(t *testing.T) {
	m := newTestModelWithContent("")

	cmd := &YankWordCommand{}
	result := cmd.Execute(m)

	require.Equal(t, Executed, result)
	require.Equal(t, "", m.lastYankedText)
	require.False(t, m.lastYankWasLinewise)
}

// TestYankWordCommand_CursorAtEOL tests yw when cursor is past end of line
func TestYankWordCommand_CursorAtEOL(t *testing.T) {
	m := newTestModelWithContent("hello")
	m.cursorCol = 5 // Past end of line

	cmd := &YankWordCommand{}
	result := cmd.Execute(m)

	require.Equal(t, Executed, result)
	require.Equal(t, "", m.lastYankedText)
}

// TestYankWordCommand_CursorUnchanged tests cursor doesn't move after yw
func TestYankWordCommand_CursorUnchanged(t *testing.T) {
	m := newTestModelWithContent("hello world")
	m.cursorRow = 0
	m.cursorCol = 3

	cmd := &YankWordCommand{}
	cmd.Execute(m)

	require.Equal(t, 0, m.cursorRow, "cursor row should not change")
	require.Equal(t, 3, m.cursorCol, "cursor col should not change")
}

// TestYankWordCommand_LastWord tests yanking the last word on line
func TestYankWordCommand_LastWord(t *testing.T) {
	m := newTestModelWithContent("hello world")
	m.cursorCol = 6 // At 'w' in "world"

	cmd := &YankWordCommand{}
	cmd.Execute(m)

	// Last word should yank to end of line
	require.Equal(t, "world", m.lastYankedText)
}

// TestYankWordCommand_Keys tests command keys
func TestYankWordCommand_Keys(t *testing.T) {
	cmd := &YankWordCommand{}
	require.Equal(t, []string{"yw"}, cmd.Keys())
}

// TestYankWordCommand_Mode tests command mode
func TestYankWordCommand_Mode(t *testing.T) {
	cmd := &YankWordCommand{}
	require.Equal(t, ModeNormal, cmd.Mode())
}

// TestYankWordCommand_ID tests command ID
func TestYankWordCommand_ID(t *testing.T) {
	cmd := &YankWordCommand{}
	require.Equal(t, "yank.word", cmd.ID())
}

// ============================================================================
// YankToEOLCommand Tests (y$, Y)
// ============================================================================

// TestYankToEOLCommand_Execute tests yanking from cursor to end of line
func TestYankToEOLCommand_Execute(t *testing.T) {
	m := newTestModelWithContent("hello world")
	m.cursorCol = 6 // At 'w' in "world"

	cmd := &YankToEOLCommand{}
	result := cmd.Execute(m)

	require.Equal(t, Executed, result)
	require.Equal(t, "world", m.lastYankedText)
	require.False(t, m.lastYankWasLinewise, "y$ should set lastYankWasLinewise = false")
}

// TestYankToEOLCommand_SetsLinewiseFalse verifies lastYankWasLinewise is false
func TestYankToEOLCommand_SetsLinewiseFalse(t *testing.T) {
	m := newTestModelWithContent("hello world")
	m.lastYankWasLinewise = true // Start with true to verify it's set to false

	cmd := &YankToEOLCommand{}
	cmd.Execute(m)

	require.False(t, m.lastYankWasLinewise)
}

// TestYankToEOLCommand_FromStart tests y$ from start of line
func TestYankToEOLCommand_FromStart(t *testing.T) {
	m := newTestModelWithContent("hello world")
	m.cursorCol = 0

	cmd := &YankToEOLCommand{}
	cmd.Execute(m)

	require.Equal(t, "hello world", m.lastYankedText)
}

// TestYankToEOLCommand_AtEOL tests y$ when cursor is at/past end of line
func TestYankToEOLCommand_AtEOL(t *testing.T) {
	m := newTestModelWithContent("hello")
	m.cursorCol = 5 // Past end of line

	cmd := &YankToEOLCommand{}
	result := cmd.Execute(m)

	require.Equal(t, Executed, result)
	require.Equal(t, "", m.lastYankedText)
}

// TestYankToEOLCommand_EmptyLine tests y$ on empty line
func TestYankToEOLCommand_EmptyLine(t *testing.T) {
	m := newTestModelWithContent("")

	cmd := &YankToEOLCommand{}
	result := cmd.Execute(m)

	require.Equal(t, Executed, result)
	require.Equal(t, "", m.lastYankedText)
}

// TestYankToEOLCommand_CursorUnchanged tests cursor doesn't move after y$
func TestYankToEOLCommand_CursorUnchanged(t *testing.T) {
	m := newTestModelWithContent("hello world")
	m.cursorRow = 0
	m.cursorCol = 3

	cmd := &YankToEOLCommand{}
	cmd.Execute(m)

	require.Equal(t, 0, m.cursorRow, "cursor row should not change")
	require.Equal(t, 3, m.cursorCol, "cursor col should not change")
}

// TestYankToEOLCommand_Keys tests command keys (y$ and Y alias)
func TestYankToEOLCommand_Keys(t *testing.T) {
	cmd := &YankToEOLCommand{}
	// This command has both y$ and Y as keys (Y is alias for y$)
	keys := cmd.Keys()
	require.Contains(t, keys, "y$")
	require.Contains(t, keys, "Y")
}

// TestYankToEOLCommand_Mode tests command mode
func TestYankToEOLCommand_Mode(t *testing.T) {
	cmd := &YankToEOLCommand{}
	require.Equal(t, ModeNormal, cmd.Mode())
}

// TestYankToEOLCommand_ID tests command ID
func TestYankToEOLCommand_ID(t *testing.T) {
	cmd := &YankToEOLCommand{}
	require.Equal(t, "yank.to_eol", cmd.ID())
}

// ============================================================================
// Y Command (alias for y$) Tests
// ============================================================================

// TestYCommand_WorksAsAliasForY tests that Y works as alias for y$
func TestYCommand_WorksAsAliasForY(t *testing.T) {
	m := newTestModelWithContent("hello world")
	m.cursorCol = 6 // At 'w'

	// Y is registered as YankToEOLCommand
	cmd, ok := DefaultRegistry.Get(ModeNormal, "Y")
	require.True(t, ok, "Y command should be registered")
	require.NotNil(t, cmd)

	result := cmd.Execute(m)
	require.Equal(t, Executed, result)
	require.Equal(t, "world", m.lastYankedText)
	require.False(t, m.lastYankWasLinewise)
}

// ============================================================================
// Registry Tests
// ============================================================================

// TestDefaultPendingRegistry_HasYankCommands verifies yank commands are registered
func TestDefaultPendingRegistry_HasYankCommands(t *testing.T) {
	// yy should be registered
	cmd, ok := DefaultPendingRegistry.Get('y', "y")
	require.True(t, ok, "yy should be registered")
	require.Equal(t, "yank.line", cmd.ID())

	// yw should be registered
	cmd, ok = DefaultPendingRegistry.Get('y', "w")
	require.True(t, ok, "yw should be registered")
	require.Equal(t, "yank.word", cmd.ID())

	// y$ should be registered
	cmd, ok = DefaultPendingRegistry.Get('y', "$")
	require.True(t, ok, "y$ should be registered")
	require.Equal(t, "yank.to_eol", cmd.ID())
}

// TestDefaultRegistry_HasYPendingOperator verifies y operator is registered
func TestDefaultRegistry_HasYPendingOperator(t *testing.T) {
	cmd, ok := DefaultRegistry.Get(ModeNormal, "y")
	require.True(t, ok, "y operator should be registered as pending command")
	require.NotNil(t, cmd)
	require.Equal(t, "pending.y", cmd.ID())
}

// TestDefaultRegistry_HasYAlias verifies Y (alias for y$) is registered
func TestDefaultRegistry_HasYAlias(t *testing.T) {
	cmd, ok := DefaultRegistry.Get(ModeNormal, "Y")
	require.True(t, ok, "Y command should be registered")
	require.NotNil(t, cmd)
	require.Equal(t, "yank.to_eol", cmd.ID())
}

// ============================================================================
// Content Unchanged Tests
// ============================================================================

// TestYankCommands_ContentUnchanged verifies yank commands don't modify content
func TestYankCommands_ContentUnchanged(t *testing.T) {
	tests := []struct {
		name    string
		cmd     Command
		content []string
	}{
		{"yy", &YankLineCommand{}, []string{"hello world"}},
		{"yw", &YankWordCommand{}, []string{"hello world"}},
		{"y$", &YankToEOLCommand{}, []string{"hello world"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Model{
				content:        append([]string{}, tt.content...),
				cursorRow:      0,
				cursorCol:      0,
				mode:           ModeNormal,
				history:        NewCommandHistory(),
				pendingBuilder: NewPendingCommandBuilder(),
			}

			tt.cmd.Execute(m)

			require.Equal(t, tt.content, m.content, "content should not be modified")
		})
	}
}

// ============================================================================
// Clipboard Integration Tests
// ============================================================================

// TestYankLineCommand_CopiesTextToClipboard tests yy copies to system clipboard
func TestYankLineCommand_CopiesTextToClipboard(t *testing.T) {
	m := newTestModelWithContent("hello world")
	clipboard := &mockClipboard{}
	m.clipboard = clipboard

	cmd := &YankLineCommand{}
	cmd.Execute(m)

	require.True(t, clipboard.copyCalled, "clipboard.Copy should be called")
	require.Equal(t, "hello world", clipboard.copiedText)
}

// TestYankLineCommand_StillSetsInternalRegister tests yy sets internal register even with clipboard
func TestYankLineCommand_StillSetsInternalRegister(t *testing.T) {
	m := newTestModelWithContent("hello world")
	clipboard := &mockClipboard{}
	m.clipboard = clipboard

	cmd := &YankLineCommand{}
	cmd.Execute(m)

	require.Equal(t, "hello world", m.lastYankedText, "internal register should still be set")
	require.True(t, m.lastYankWasLinewise, "lastYankWasLinewise should be true")
}

// TestYankLineCommand_WithNilClipboard_NoPanic tests yy doesn't panic with nil clipboard
func TestYankLineCommand_WithNilClipboard_NoPanic(t *testing.T) {
	m := newTestModelWithContent("hello world")
	// clipboard is nil by default

	cmd := &YankLineCommand{}
	require.NotPanics(t, func() {
		cmd.Execute(m)
	})

	// Internal register should still work
	require.Equal(t, "hello world", m.lastYankedText)
}

// TestYankLineCommand_ClipboardError_StillSetsRegister tests yy sets register even on clipboard error
func TestYankLineCommand_ClipboardError_StillSetsRegister(t *testing.T) {
	m := newTestModelWithContent("hello world")
	clipboard := &mockClipboard{copyErr: errClipboard}
	m.clipboard = clipboard

	cmd := &YankLineCommand{}
	cmd.Execute(m)

	// Clipboard was called (even though it failed)
	require.True(t, clipboard.copyCalled)
	// Internal register should still be set
	require.Equal(t, "hello world", m.lastYankedText)
}

// TestYankWordCommand_CopiesTextToClipboard tests yw copies to system clipboard
func TestYankWordCommand_CopiesTextToClipboard(t *testing.T) {
	m := newTestModelWithContent("hello world")
	clipboard := &mockClipboard{}
	m.clipboard = clipboard
	m.cursorCol = 0 // At 'h'

	cmd := &YankWordCommand{}
	cmd.Execute(m)

	require.True(t, clipboard.copyCalled, "clipboard.Copy should be called")
	require.Equal(t, "hello ", clipboard.copiedText)
}

// TestYankWordCommand_StillSetsInternalRegister tests yw sets internal register even with clipboard
func TestYankWordCommand_StillSetsInternalRegister(t *testing.T) {
	m := newTestModelWithContent("hello world")
	clipboard := &mockClipboard{}
	m.clipboard = clipboard
	m.cursorCol = 0

	cmd := &YankWordCommand{}
	cmd.Execute(m)

	require.Equal(t, "hello ", m.lastYankedText, "internal register should still be set")
	require.False(t, m.lastYankWasLinewise, "lastYankWasLinewise should be false")
}

// TestYankWordCommand_WithNilClipboard_NoPanic tests yw doesn't panic with nil clipboard
func TestYankWordCommand_WithNilClipboard_NoPanic(t *testing.T) {
	m := newTestModelWithContent("hello world")
	// clipboard is nil by default
	m.cursorCol = 0

	cmd := &YankWordCommand{}
	require.NotPanics(t, func() {
		cmd.Execute(m)
	})

	require.Equal(t, "hello ", m.lastYankedText)
}

// TestYankWordCommand_ClipboardError_StillSetsRegister tests yw sets register even on clipboard error
func TestYankWordCommand_ClipboardError_StillSetsRegister(t *testing.T) {
	m := newTestModelWithContent("hello world")
	clipboard := &mockClipboard{copyErr: errClipboard}
	m.clipboard = clipboard
	m.cursorCol = 0

	cmd := &YankWordCommand{}
	cmd.Execute(m)

	require.True(t, clipboard.copyCalled)
	require.Equal(t, "hello ", m.lastYankedText)
}

// TestYankWordCommand_EmptyLine_CopiesEmptyToClipboard tests yw on empty line copies empty string
func TestYankWordCommand_EmptyLine_CopiesEmptyToClipboard(t *testing.T) {
	m := newTestModelWithContent("")
	clipboard := &mockClipboard{}
	m.clipboard = clipboard

	cmd := &YankWordCommand{}
	cmd.Execute(m)

	require.True(t, clipboard.copyCalled, "clipboard.Copy should be called even for empty string")
	require.Equal(t, "", clipboard.copiedText)
}

// TestYankToEOLCommand_CopiesTextToClipboard tests y$ copies to system clipboard
func TestYankToEOLCommand_CopiesTextToClipboard(t *testing.T) {
	m := newTestModelWithContent("hello world")
	clipboard := &mockClipboard{}
	m.clipboard = clipboard
	m.cursorCol = 6 // At 'w' in "world"

	cmd := &YankToEOLCommand{}
	cmd.Execute(m)

	require.True(t, clipboard.copyCalled, "clipboard.Copy should be called")
	require.Equal(t, "world", clipboard.copiedText)
}

// TestYankToEOLCommand_StillSetsInternalRegister tests y$ sets internal register even with clipboard
func TestYankToEOLCommand_StillSetsInternalRegister(t *testing.T) {
	m := newTestModelWithContent("hello world")
	clipboard := &mockClipboard{}
	m.clipboard = clipboard
	m.cursorCol = 6

	cmd := &YankToEOLCommand{}
	cmd.Execute(m)

	require.Equal(t, "world", m.lastYankedText, "internal register should still be set")
	require.False(t, m.lastYankWasLinewise, "lastYankWasLinewise should be false")
}

// TestYankToEOLCommand_WithNilClipboard_NoPanic tests y$ doesn't panic with nil clipboard
func TestYankToEOLCommand_WithNilClipboard_NoPanic(t *testing.T) {
	m := newTestModelWithContent("hello world")
	// clipboard is nil by default
	m.cursorCol = 6

	cmd := &YankToEOLCommand{}
	require.NotPanics(t, func() {
		cmd.Execute(m)
	})

	require.Equal(t, "world", m.lastYankedText)
}

// TestYankToEOLCommand_ClipboardError_StillSetsRegister tests y$ sets register even on clipboard error
func TestYankToEOLCommand_ClipboardError_StillSetsRegister(t *testing.T) {
	m := newTestModelWithContent("hello world")
	clipboard := &mockClipboard{copyErr: errClipboard}
	m.clipboard = clipboard
	m.cursorCol = 6

	cmd := &YankToEOLCommand{}
	cmd.Execute(m)

	require.True(t, clipboard.copyCalled)
	require.Equal(t, "world", m.lastYankedText)
}

// TestYankToEOLCommand_AtEOL_CopiesEmptyToClipboard tests y$ at end of line copies empty string
func TestYankToEOLCommand_AtEOL_CopiesEmptyToClipboard(t *testing.T) {
	m := newTestModelWithContent("hello")
	clipboard := &mockClipboard{}
	m.clipboard = clipboard
	m.cursorCol = 5 // Past end of line

	cmd := &YankToEOLCommand{}
	cmd.Execute(m)

	require.True(t, clipboard.copyCalled, "clipboard.Copy should be called even for empty string")
	require.Equal(t, "", clipboard.copiedText)
}

// TestYankToEOLCommand_YAlias_CopiesTextToClipboard tests Y (alias for y$) copies to clipboard
func TestYankToEOLCommand_YAlias_CopiesTextToClipboard(t *testing.T) {
	m := newTestModelWithContent("hello world")
	clipboard := &mockClipboard{}
	m.clipboard = clipboard
	m.cursorCol = 6 // At 'w'

	// Y is registered as YankToEOLCommand
	cmd, ok := DefaultRegistry.Get(ModeNormal, "Y")
	require.True(t, ok, "Y command should be registered")

	cmd.Execute(m)

	require.True(t, clipboard.copyCalled, "clipboard.Copy should be called")
	require.Equal(t, "world", clipboard.copiedText)
}
