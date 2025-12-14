// Package styles contains Lip Gloss style definitions.
package styles

import "fmt"

// FormatCommentIndicator returns the comment indicator string.
// Returns empty string when count is 0.
func FormatCommentIndicator(count int) string {
	if count <= 0 {
		return ""
	}
	return fmt.Sprintf("%d\U0001F4AC", count) // 💬
}
