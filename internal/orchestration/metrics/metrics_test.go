package metrics

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestContextUsage(t *testing.T) {
	tests := []struct {
		name        string
		tokensUsed  int
		totalTokens int
		want        float64
	}{
		{
			name:        "zero window returns zero",
			tokensUsed:  1000,
			totalTokens: 0,
			want:        0,
		},
		{
			name:        "zero tokens returns zero",
			tokensUsed:  0,
			totalTokens: 200000,
			want:        0,
		},
		{
			name:        "50% usage",
			tokensUsed:  100000,
			totalTokens: 200000,
			want:        50,
		},
		{
			name:        "85% usage (critical threshold)",
			tokensUsed:  170000,
			totalTokens: 200000,
			want:        85,
		},
		{
			name:        "70% usage (warning threshold)",
			tokensUsed:  140000,
			totalTokens: 200000,
			want:        70,
		},
		{
			name:        "27k/200k typical usage",
			tokensUsed:  27000,
			totalTokens: 200000,
			want:        13.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := TokenMetrics{
				TokensUsed:  tt.tokensUsed,
				TotalTokens: tt.totalTokens,
			}
			got := m.ContextUsage()
			require.Equal(t, tt.want, got, "ContextUsage()")
		})
	}
}

func TestFormatContextDisplay(t *testing.T) {
	tests := []struct {
		name        string
		tokensUsed  int
		totalTokens int
		want        string
	}{
		{
			name:        "zero window returns dash",
			tokensUsed:  1000,
			totalTokens: 0,
			want:        "-",
		},
		{
			name:        "typical 27k/200k",
			tokensUsed:  27000,
			totalTokens: 200000,
			want:        "27k/200k",
		},
		{
			name:        "45k/200k",
			tokensUsed:  45000,
			totalTokens: 200000,
			want:        "45k/200k",
		},
		{
			name:        "small numbers round down",
			tokensUsed:  500,
			totalTokens: 200000,
			want:        "0k/200k",
		},
		{
			name:        "100k context window",
			tokensUsed:  50000,
			totalTokens: 100000,
			want:        "50k/100k",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := TokenMetrics{
				TokensUsed:  tt.tokensUsed,
				TotalTokens: tt.totalTokens,
			}
			got := m.FormatContextDisplay()
			require.Equal(t, tt.want, got, "FormatContextDisplay()")
		})
	}
}

func TestFormatCostDisplay(t *testing.T) {
	tests := []struct {
		name         string
		totalCostUSD float64
		want         string
	}{
		{
			name:         "zero cost",
			totalCostUSD: 0,
			want:         "$0.0000",
		},
		{
			name:         "small cost",
			totalCostUSD: 0.0892,
			want:         "$0.0892",
		},
		{
			name:         "larger cost",
			totalCostUSD: 1.2345,
			want:         "$1.2345",
		},
		{
			name:         "rounds to 4 decimal places",
			totalCostUSD: 0.12345678,
			want:         "$0.1235",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := TokenMetrics{
				TotalCostUSD: tt.totalCostUSD,
			}
			got := m.FormatCostDisplay()
			require.Equal(t, tt.want, got, "FormatCostDisplay()")
		})
	}
}

func TestTokenMetrics_FullStruct(t *testing.T) {
	// Test a fully populated struct to ensure all fields work together
	now := time.Now()
	m := TokenMetrics{
		TokensUsed:    27000,
		TotalTokens:   200000,
		OutputTokens:  1000,
		TurnCostUSD:   0.0150,
		TotalCostUSD:  0.0892,
		LastUpdatedAt: now,
	}

	// Verify all calculations
	require.Equal(t, 13.5, m.ContextUsage(), "ContextUsage()")
	require.Equal(t, "27k/200k", m.FormatContextDisplay(), "FormatContextDisplay()")
	require.Equal(t, "$0.0892", m.FormatCostDisplay(), "FormatCostDisplay()")
	require.Equal(t, now, m.LastUpdatedAt, "LastUpdatedAt")
}
