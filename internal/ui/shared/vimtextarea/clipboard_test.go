package vimtextarea

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockClipboard is a simple mock for testing clipboard functionality.
type mockClipboard struct {
	copiedText string
	copyErr    error
	copyCalled bool
}

func (m *mockClipboard) Copy(text string) error {
	m.copyCalled = true
	if m.copyErr != nil {
		return m.copyErr
	}
	m.copiedText = text
	return nil
}

func TestSetClipboard_ReturnsNewModelWithClipboardSet(t *testing.T) {
	// Arrange
	m := New(Config{VimEnabled: true})
	clipboard := &mockClipboard{}

	// Act
	m2 := m.SetClipboard(clipboard)

	// Assert - new model has clipboard set
	require.NotNil(t, m2.clipboard)

	// Verify immutability - original model unchanged
	assert.Nil(t, m.clipboard)
}

func TestCopyToSystemClipboard_WithNilClipboard_DoesNotPanic(t *testing.T) {
	// Arrange
	m := newTestModelWithContent("hello world")
	// clipboard is nil by default

	// Act & Assert - should not panic
	assert.NotPanics(t, func() {
		m.copyToSystemClipboard("test")
	})
}

func TestCopyToSystemClipboard_WithValidClipboard_CallsCopy(t *testing.T) {
	// Arrange
	m := newTestModelWithContent("hello world")
	clipboard := &mockClipboard{}
	m.clipboard = clipboard

	// Act
	m.copyToSystemClipboard("test text")

	// Assert
	assert.True(t, clipboard.copyCalled)
	assert.Equal(t, "test text", clipboard.copiedText)
}

func TestCopyToSystemClipboard_WithClipboardError_LogsButDoesNotFail(t *testing.T) {
	// Arrange
	m := newTestModelWithContent("hello world")
	clipboard := &mockClipboard{
		copyErr: errors.New("clipboard unavailable"),
	}
	m.clipboard = clipboard

	// Act & Assert - should not panic, error is logged but not propagated
	assert.NotPanics(t, func() {
		m.copyToSystemClipboard("test")
	})

	// Verify Copy was still called
	assert.True(t, clipboard.copyCalled)
}

func TestSetClipboard_ImmutablePattern_ChainedCalls(t *testing.T) {
	// Arrange
	m := New(Config{VimEnabled: true})
	clipboard1 := &mockClipboard{}
	clipboard2 := &mockClipboard{}

	// Act - chain SetClipboard calls
	m1 := m.SetClipboard(clipboard1)
	m2 := m1.SetClipboard(clipboard2)

	// Assert - each model has its own clipboard
	assert.Nil(t, m.clipboard)
	assert.Same(t, clipboard1, m1.clipboard)
	assert.Same(t, clipboard2, m2.clipboard)
}

func TestCopyToSystemClipboard_WithEmptyString_StillCallsCopy(t *testing.T) {
	// Arrange
	m := newTestModelWithContent("hello")
	clipboard := &mockClipboard{}
	m.clipboard = clipboard

	// Act
	m.copyToSystemClipboard("")

	// Assert - empty string is a valid clipboard operation
	assert.True(t, clipboard.copyCalled)
	assert.Equal(t, "", clipboard.copiedText)
}
