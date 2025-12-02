package styles

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestApplyTheme_Default(t *testing.T) {
	err := ApplyTheme(ThemeConfig{})
	assert.NoError(t, err)
	// Should apply default preset colors
	assert.Equal(t, DefaultPreset.Colors[TokenTextPrimary], TextPrimaryColor.Dark)
}

func TestApplyTheme_Preset(t *testing.T) {
	// First add a test preset
	TestPreset := Preset{
		Name:        "test",
		Description: "Test preset",
		Colors: map[ColorToken]string{
			TokenTextPrimary: "#FF0000",
		},
	}
	Presets["test"] = TestPreset
	defer delete(Presets, "test")

	err := ApplyTheme(ThemeConfig{Preset: "test"})
	assert.NoError(t, err)
	assert.Equal(t, "#FF0000", TextPrimaryColor.Dark)
}

func TestApplyTheme_ColorOverride(t *testing.T) {
	err := ApplyTheme(ThemeConfig{
		Colors: map[string]string{
			"text.primary": "#00FF00",
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, "#00FF00", TextPrimaryColor.Dark)
}

func TestApplyTheme_PresetWithOverride(t *testing.T) {
	// Color override should take precedence over preset
	TestPreset := Preset{
		Name:        "test2",
		Description: "Test preset 2",
		Colors: map[ColorToken]string{
			TokenTextPrimary:   "#FF0000",
			TokenTextSecondary: "#0000FF",
		},
	}
	Presets["test2"] = TestPreset
	defer delete(Presets, "test2")

	err := ApplyTheme(ThemeConfig{
		Preset: "test2",
		Colors: map[string]string{
			"text.primary": "#00FF00", // Override preset
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, "#00FF00", TextPrimaryColor.Dark)   // Overridden
	assert.Equal(t, "#0000FF", TextSecondaryColor.Dark) // From preset
}

func TestApplyTheme_InvalidPreset(t *testing.T) {
	err := ApplyTheme(ThemeConfig{Preset: "nonexistent"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown theme preset")
}

func TestApplyTheme_InvalidToken(t *testing.T) {
	err := ApplyTheme(ThemeConfig{
		Colors: map[string]string{
			"invalid.token": "#FF0000",
		},
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown color token")
}

func TestApplyTheme_InvalidHexColor(t *testing.T) {
	err := ApplyTheme(ThemeConfig{
		Colors: map[string]string{
			"text.primary": "not-a-color",
		},
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid hex color")
}

func TestIsValidToken(t *testing.T) {
	tests := []struct {
		token ColorToken
		valid bool
	}{
		{TokenTextPrimary, true},
		{TokenStatusError, true},
		{ColorToken("invalid.token"), false},
		{ColorToken(""), false},
	}
	for _, tt := range tests {
		t.Run(string(tt.token), func(t *testing.T) {
			assert.Equal(t, tt.valid, isValidToken(tt.token))
		})
	}
}

func TestIsValidHexColor(t *testing.T) {
	tests := []struct {
		color string
		valid bool
	}{
		{"#FFF", true},
		{"#FFFFFF", true},
		{"#abc", true},
		{"#AbCdEf", true},
		{"#123456", true},
		{"FFFFFF", false},   // Missing #
		{"#FF", false},      // Too short
		{"#FFFFFFF", false}, // Too long
		{"#GGGGGG", false},  // Invalid chars
		{"not-color", false},
		{"", false},
	}
	for _, tt := range tests {
		t.Run(tt.color, func(t *testing.T) {
			assert.Equal(t, tt.valid, isValidHexColor(tt.color))
		})
	}
}
