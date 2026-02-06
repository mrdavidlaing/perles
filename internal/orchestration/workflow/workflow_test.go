package workflow

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSource_String(t *testing.T) {
	tests := []struct {
		source   Source
		expected string
	}{
		{SourceBuiltIn, "built-in"},
		{SourceCommunity, "community"},
		{SourceUser, "user"},
		{Source(99), "unknown"},
	}

	for _, tc := range tests {
		require.Equal(t, tc.expected, tc.source.String())
	}
}
