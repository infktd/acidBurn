package ui

import (
	"testing"
)

func TestGetThemeReturnsDefault(t *testing.T) {
	theme := GetTheme("acid-green")
	if theme.Primary == "" {
		t.Fatal("Theme should have a primary color")
	}
}

func TestGetThemeFallsBackToDefault(t *testing.T) {
	theme := GetTheme("nonexistent-theme")
	if theme.Primary == "" {
		t.Fatal("Unknown theme should fall back to default")
	}
}

func TestAllThemesHaveRequiredColors(t *testing.T) {
	for name, theme := range Themes {
		if theme.Primary == "" {
			t.Errorf("Theme %q missing Primary", name)
		}
		if theme.Secondary == "" {
			t.Errorf("Theme %q missing Secondary", name)
		}
		if theme.Background == "" {
			t.Errorf("Theme %q missing Background", name)
		}
		if theme.Muted == "" {
			t.Errorf("Theme %q missing Muted", name)
		}
	}
}
