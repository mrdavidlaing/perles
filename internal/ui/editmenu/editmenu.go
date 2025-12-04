// Package editmenu provides a menu component for issue editing operations.
package editmenu

import (
	"perles/internal/ui/overlay"
	"perles/internal/ui/styles"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Option represents an edit menu option.
type Option int

const (
	OptionLabels Option = iota
	OptionPriority
	OptionStatus
)

// optionLabels maps options to their display labels.
var optionLabels = map[Option]string{
	OptionLabels:   "Edit labels",
	OptionPriority: "Change priority",
	OptionStatus:   "Change status",
}

// SelectMsg is sent when an option is selected.
type SelectMsg struct {
	Option Option
}

// CancelMsg is sent when the menu is cancelled.
type CancelMsg struct{}

// Model holds the edit menu state.
type Model struct {
	selected       Option
	viewportWidth  int
	viewportHeight int
}

// New creates a new edit menu.
func New() Model {
	return Model{
		selected: OptionLabels,
	}
}

// SetSize sets the viewport dimensions for overlay rendering.
func (m Model) SetSize(width, height int) Model {
	m.viewportWidth = width
	m.viewportHeight = height
	return m
}

// Update handles messages.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down", "ctrl+n":
			if m.selected < OptionStatus {
				m.selected++
			}
		case "k", "up", "ctrl+p":
			if m.selected > OptionLabels {
				m.selected--
			}
		case "enter":
			return m, func() tea.Msg {
				return SelectMsg{Option: m.selected}
			}
		case "esc":
			return m, func() tea.Msg {
				return CancelMsg{}
			}
		}
	}
	return m, nil
}

// View renders the menu box (without positioning).
func (m Model) View() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(styles.OverlayTitleColor).
		PaddingLeft(1)

	width := 25

	// Build options
	var options strings.Builder
	for i := OptionLabels; i <= OptionStatus; i++ {
		var line string
		if i == m.selected {
			labelStyle := lipgloss.NewStyle().Bold(true)
			line = styles.SelectionIndicatorStyle.Render(">") + labelStyle.Render(optionLabels[i])
		} else {
			line = " " + optionLabels[i]
		}
		options.WriteString(line)
		if i < OptionStatus {
			options.WriteString("\n")
		}
	}

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.OverlayBorderColor).
		Width(width)

	// Divider spans full width (no padding)
	dividerStyle := lipgloss.NewStyle().Foreground(styles.OverlayBorderColor)
	divider := dividerStyle.Render(strings.Repeat("─", width))
	content := titleStyle.Render("Edit Issue") + "\n" +
		divider + "\n" +
		options.String()

	return boxStyle.Render(content)
}

// Overlay renders the menu on top of a background view.
func (m Model) Overlay(background string) string {
	menuBox := m.View()

	if background == "" {
		return lipgloss.Place(
			m.viewportWidth, m.viewportHeight,
			lipgloss.Center, lipgloss.Center,
			menuBox,
		)
	}

	return overlay.Place(overlay.Config{
		Width:    m.viewportWidth,
		Height:   m.viewportHeight,
		Position: overlay.Center,
	}, menuBox, background)
}

// Selected returns the currently selected option.
func (m Model) Selected() Option {
	return m.selected
}
