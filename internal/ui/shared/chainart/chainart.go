// Package chainart provides the shared chain ASCII art used by nobeads and outdated views.
package chainart

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Chain link color definitions (internal to package).
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

// BuildChainArt constructs the colored chain ASCII art.
// The broken link is taller than the regular links, so we need to
// align them properly by adding padding to the shorter links.
func BuildChainArt() string {
	// Create styles from the color definitions
	link1Style := lipgloss.NewStyle().Foreground(link1Color)
	link2Style := lipgloss.NewStyle().Foreground(link2Color)
	brokenStyle := lipgloss.NewStyle().Foreground(brokenColor)
	link4Style := lipgloss.NewStyle().Foreground(link4Color)
	link5Style := lipgloss.NewStyle().Foreground(link5Color)
	connectorStyle := lipgloss.NewStyle().Foreground(connectorColor)
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
