package styles

import "testing"

func TestFormatCommentIndicator(t *testing.T) {
	tests := []struct {
		name     string
		count    int
		expected string
	}{
		{"zero comments", 0, ""},
		{"negative count", -1, ""},
		{"one comment", 1, "1\U0001F4AC"},
		{"few comments", 3, "3\U0001F4AC"},
		{"many comments", 99, "99\U0001F4AC"},
		{"lots of comments", 999, "999\U0001F4AC"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatCommentIndicator(tt.count)
			if got != tt.expected {
				t.Errorf("FormatCommentIndicator(%d) = %q, want %q",
					tt.count, got, tt.expected)
			}
		})
	}
}
