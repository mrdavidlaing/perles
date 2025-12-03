// Package nobeads provides the empty state view shown when no .beads directory exists.
package nobeads

import (
	"perles/internal/ui/styles"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Chain link color definitions.
var (
	link1Color     = lipgloss.Color("#54A0FF") // Blue
	link2Color     = lipgloss.Color("#73F59F") // Green
	brokenColor    = lipgloss.Color("#696969") // Grey - broken/inactive
	link4Color     = lipgloss.Color("#FECA57") // Yellow
	link5Color     = lipgloss.Color("#7D56F4") // Purple
	connectorColor = lipgloss.Color("#CCCCCC") // Light - consistent connectors
)

// Chain link pieces (rendered separately for coloring).
// Each link is rendered independently then joined horizontally.
var (
	// First link (left side open for chain start)
	link1Lines = []string{
		"╔═══════╗",
		"║       ╠",
		"║       ╠",
		"╚═══════╝",
	}

	// Middle links (open on both sides)
	link2Lines = []string{
		"╔═══════╗",
		"╣       ╠",
		"╣       ╠",
		"╚═══════╝",
	}

	// Broken link with radiating crack lines
	brokenLines = []string{
		"    \\│/    ",
		"╔════╲   │   ╱════╗",
		"╣     ╲  │  ╱     ╠",
		"╣     ╱  │  ╲     ╠",
		"╚════╱   │   ╲════╝",
		"    /│\\    ",
	}

	link4Lines = []string{
		"╔═══════╗",
		"╣       ╠",
		"╣       ╠",
		"╚═══════╝",
	}

	// Last link (right side closed for chain end)
	link5Lines = []string{
		"╔═══════╗",
		"╣       ║",
		"╣       ║",
		"╚═══════╝",
	}

	// Connectors between links
	connectorLines = []string{
		"",
		"═══",
		"═══",
		"",
	}
)

// Model holds the nobeads view state.
type Model struct {
	width  int
	height int
}

// New creates a new nobeads view.
func New() Model {
	return Model{}
}

// Init returns the initial command.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		}
	}
	return m, nil
}

// View renders the empty state.
func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}

	// Chain link styles - each link gets a different color
	link1Style := lipgloss.NewStyle().Foreground(link1Color)
	link2Style := lipgloss.NewStyle().Foreground(link2Color)
	brokenStyle := lipgloss.NewStyle().Foreground(brokenColor)
	link4Style := lipgloss.NewStyle().Foreground(link4Color)
	link5Style := lipgloss.NewStyle().Foreground(link5Color)
	connectorStyle := lipgloss.NewStyle().Foreground(connectorColor)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(styles.TextPrimaryColor).
		MarginTop(1)

	messageStyle := lipgloss.NewStyle().
		Foreground(styles.TextDescriptionColor)

	hintStyle := lipgloss.NewStyle().
		Foreground(styles.TextMutedColor).
		Italic(true).
		MarginTop(2)

	// Build chain art by rendering each link with its color
	chainArt := buildChainArt(
		link1Style, link2Style, brokenStyle, link4Style, link5Style, connectorStyle,
	)

	// Build content
	var content strings.Builder

	content.WriteString(chainArt)
	content.WriteString("\n\n")
	content.WriteString(titleStyle.Render("Oh no! Looks like there's a break in the chain!"))
	content.WriteString("\n\n")
	content.WriteString(messageStyle.Render("No .beads directory found in the current directory."))
	content.WriteString("\n\n")
	content.WriteString(messageStyle.Render("Try one of these options:"))
	content.WriteString("\n\n")
	content.WriteString(messageStyle.Render("  1. (Recommended) Run perles from a directory containing .beads/"))
	content.WriteString("\n")
	content.WriteString(messageStyle.Render("  2. Use the --beads-dir flag: perles --beads-dir /path/to/project"))
	content.WriteString("\n")
	content.WriteString(messageStyle.Render("  3. Run 'perles init' to create a local config file, then set beads_dir"))
	content.WriteString("\n")
	content.WriteString(messageStyle.Render("  4. Set beads_dir in your config file (~/.config/perles/config.yaml)"))
	content.WriteString("\n\n")
	content.WriteString(hintStyle.Render("Press q to quit"))

	// Center the content
	containerStyle := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Center)

	return containerStyle.Render(content.String())
}

// SetSize updates the view dimensions.
func (m Model) SetSize(width, height int) Model {
	m.width = width
	m.height = height
	return m
}

// buildChainArt constructs the colored chain ASCII art.
// The broken link is taller than the regular links, so we need to
// align them properly by adding padding to the shorter links.
func buildChainArt(
	link1Style, link2Style, brokenStyle, link4Style, link5Style, connectorStyle lipgloss.Style,
) string {
	// Broken link is 6 lines, regular links are 4 lines
	// Add 1 empty line at top and 1 at bottom of regular links for alignment
	paddedLink1 := padLines(link1Lines)
	paddedLink2 := padLines(link2Lines)
	paddedLink4 := padLines(link4Lines)
	paddedLink5 := padLines(link5Lines)
	paddedConnector := padLines(connectorLines)

	// Render each piece with its style
	link1Rendered := renderLines(paddedLink1, link1Style)
	conn1Rendered := renderLines(paddedConnector, connectorStyle)
	link2Rendered := renderLines(paddedLink2, link2Style)
	conn2Rendered := renderLines(paddedConnector, connectorStyle)
	brokenRendered := renderLines(brokenLines, brokenStyle)
	conn3Rendered := renderLines(paddedConnector, connectorStyle)
	link4Rendered := renderLines(paddedLink4, link4Style)
	conn4Rendered := renderLines(paddedConnector, connectorStyle)
	link5Rendered := renderLines(paddedLink5, link5Style)

	// Join horizontally
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		link1Rendered,
		conn1Rendered,
		link2Rendered,
		conn2Rendered,
		brokenRendered,
		conn3Rendered,
		link4Rendered,
		conn4Rendered,
		link5Rendered,
	)
}

// padLines adds empty lines to center the content vertically within 6 lines.
func padLines(lines []string) []string {
	const targetHeight = 6
	if len(lines) >= targetHeight {
		return lines
	}

	// Find the width of the widest line for padding
	maxWidth := 0
	for _, line := range lines {
		if w := lipgloss.Width(line); w > maxWidth {
			maxWidth = w
		}
	}

	// Calculate padding
	diff := targetHeight - len(lines)
	topPad := diff / 2
	bottomPad := diff - topPad

	emptyLine := strings.Repeat(" ", maxWidth)
	result := make([]string, 0, targetHeight)

	// Add top padding
	for range topPad {
		result = append(result, emptyLine)
	}

	// Add original lines
	result = append(result, lines...)

	// Add bottom padding
	for range bottomPad {
		result = append(result, emptyLine)
	}

	return result
}

// renderLines joins lines with newlines and applies a style.
func renderLines(lines []string, style lipgloss.Style) string {
	rendered := strings.Join(lines, "\n")
	return style.Render(rendered)
}
