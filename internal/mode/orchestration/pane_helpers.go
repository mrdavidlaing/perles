package orchestration

import (
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"

	"github.com/zjrosen/perles/internal/ui/styles"
)

// ScrollablePaneConfig holds the configuration for rendering a scrollable pane.
type ScrollablePaneConfig struct {
	// Viewport is a pointer to the viewport in the pane's map.
	// CRITICAL: Must be a pointer to preserve reference semantics for scroll state persistence.
	Viewport *viewport.Model

	// ContentDirty indicates whether the content has changed since last render.
	// Used to determine if auto-scroll to bottom should occur.
	ContentDirty bool

	// HasNewContent indicates if new content arrived while scrolled up.
	// Displayed as "↓New" indicator in the right title.
	HasNewContent bool

	// MetricsDisplay is optional metrics text (e.g., "27k/200k" for context).
	// Displayed in the right title.
	MetricsDisplay string

	// LeftTitle is the title shown on the left side of the border.
	LeftTitle string

	// TitleColor is the color for the title text.
	TitleColor lipgloss.AdaptiveColor

	// BorderColor is the color for the pane border.
	BorderColor lipgloss.AdaptiveColor
}

// renderScrollablePane handles the common viewport setup, content padding, auto-scroll,
// and border rendering pattern used by all pane render functions.
//
// CRITICAL INVARIANTS (do not change the order of operations):
//  1. wasAtBottom MUST be captured BEFORE SetContent() to preserve user scroll position.
//     If checked after SetContent(), users will be forcibly scrolled to bottom on every render.
//  2. Content padding MUST be PREPENDED (not appended) to push content to the bottom of the viewport.
//     Appending padding would leave content at the top.
//  3. Viewport MUST use pointer semantics (stored in map) for scroll state to persist across renders.
//
// contentFn receives the available width (viewport width) and returns the rendered content string.
func renderScrollablePane(
	width, height int,
	cfg ScrollablePaneConfig,
	contentFn func(wrapWidth int) string,
) string {
	// Calculate viewport dimensions (subtract 2 for borders)
	vpWidth := max(width-2, 1)
	vpHeight := max(height-2, 1)

	// Build pre-wrapped content
	content := contentFn(vpWidth)

	// Pad content to push it to the bottom when it's shorter than viewport.
	// CRITICAL: Padding must be PREPENDED to push content to bottom.
	// This preserves the "latest content at bottom" behavior.
	contentLines := strings.Split(content, "\n")
	if len(contentLines) < vpHeight {
		padding := make([]string, vpHeight-len(contentLines))
		contentLines = append(padding, contentLines...) // Prepend padding
		content = strings.Join(contentLines, "\n")
	}

	// Update viewport dimensions
	cfg.Viewport.Width = vpWidth
	cfg.Viewport.Height = vpHeight

	// CRITICAL: Check if user was at bottom BEFORE SetContent() changes the viewport state.
	// This enables smart auto-scroll: only follow new content if user was at bottom.
	wasAtBottom := cfg.Viewport.AtBottom()

	cfg.Viewport.SetContent(content)

	// Smart auto-scroll: only scroll to bottom if content is dirty AND user was at bottom.
	// This preserves scroll position when user has scrolled up to read history.
	if cfg.ContentDirty && wasAtBottom {
		cfg.Viewport.GotoBottom()
	}

	// Get viewport view (handles scrolling and clipping)
	viewportContent := cfg.Viewport.View()

	// Build right title with new content indicator, scroll indicator, and metrics
	// This must happen AFTER SetContent so scroll indicator is accurate
	rightTitle := buildRightTitle(*cfg.Viewport, cfg.HasNewContent, cfg.MetricsDisplay)

	// Render pane with bordered title
	return styles.RenderWithTitleBorder(
		viewportContent,
		cfg.LeftTitle,
		rightTitle,
		width,
		height,
		false,
		cfg.TitleColor,
		cfg.BorderColor,
	)
}

// buildRightTitle constructs the right title section for pane borders.
// It combines the new content indicator, scroll indicator, and optional metrics display.
func buildRightTitle(vp viewport.Model, hasNewContent bool, metricsDisplay string) string {
	var parts []string

	// Add new content indicator if scrolled up and new content arrived
	if hasNewContent {
		parts = append(parts, newContentIndicatorStyle.Render("↓New"))
	}

	// Add scroll indicator if scrolled up from bottom
	if scrollIndicator := buildScrollIndicator(vp); scrollIndicator != "" {
		parts = append(parts, scrollIndicator)
	}

	// Add metrics display if available (e.g., "27k/200k" for context usage)
	if metricsDisplay != "" {
		parts = append(parts, scrollIndicatorStyle.Render(metricsDisplay))
	}

	return strings.Join(parts, " ")
}
