package ui

import (
	"strings"
	"testing"
)

func TestSplashScreenCreate(t *testing.T) {
	styles := NewStyles(GetTheme("matrix"))
	splash := NewSplashScreen(styles, 80, 24)

	if splash == nil {
		t.Fatal("splash should not be nil")
	}
	if !splash.IsVisible() {
		t.Error("splash should be visible by default")
	}
}

func TestSplashScreenProgress(t *testing.T) {
	styles := NewStyles(GetTheme("matrix"))
	splash := NewSplashScreen(styles, 80, 24)

	splash.SetProgress(0.5)
	if splash.Progress() != 0.5 {
		t.Errorf("expected 0.5, got %f", splash.Progress())
	}

	// Clamp to valid range
	splash.SetProgress(1.5)
	if splash.Progress() != 1.0 {
		t.Errorf("expected 1.0 (clamped), got %f", splash.Progress())
	}

	splash.SetProgress(-0.5)
	if splash.Progress() != 0.0 {
		t.Errorf("expected 0.0 (clamped), got %f", splash.Progress())
	}
}

func TestSplashScreenMessage(t *testing.T) {
	styles := NewStyles(GetTheme("matrix"))
	splash := NewSplashScreen(styles, 80, 24)

	splash.SetMessage("Scanning projects...")
	if splash.Message() != "Scanning projects..." {
		t.Errorf("expected 'Scanning projects...', got %q", splash.Message())
	}
}

func TestSplashScreenShowHide(t *testing.T) {
	styles := NewStyles(GetTheme("matrix"))
	splash := NewSplashScreen(styles, 80, 24)

	splash.Hide()
	if splash.IsVisible() {
		t.Error("should not be visible after Hide()")
	}

	splash.Show()
	if !splash.IsVisible() {
		t.Error("should be visible after Show()")
	}
}

func TestSplashScreenView(t *testing.T) {
	styles := NewStyles(GetTheme("matrix"))
	splash := NewSplashScreen(styles, 80, 24)

	view := splash.View()
	if view == "" {
		t.Error("view should have content when visible")
	}

	// Should contain tagline
	if !strings.Contains(view, "fleet control") {
		t.Error("view should contain 'fleet control' tagline")
	}

	// Should contain progress bar characters
	if !strings.Contains(view, "░") && !strings.Contains(view, "█") {
		t.Error("view should contain progress bar")
	}
}

func TestSplashScreenViewWhenHidden(t *testing.T) {
	styles := NewStyles(GetTheme("matrix"))
	splash := NewSplashScreen(styles, 80, 24)

	splash.Hide()
	view := splash.View()
	if view != "" {
		t.Error("view should be empty when hidden")
	}
}

func TestSplashScreenCustomAsciiArt(t *testing.T) {
	styles := NewStyles(GetTheme("matrix"))
	splash := NewSplashScreen(styles, 80, 24)

	customArt := "=== CUSTOM ==="
	splash.SetAsciiArt(customArt)

	view := splash.View()
	if !strings.Contains(view, customArt) {
		t.Error("view should contain custom ASCII art")
	}
}

func TestSplashScreenProgressBar(t *testing.T) {
	styles := NewStyles(GetTheme("matrix"))
	splash := NewSplashScreen(styles, 80, 24)

	// At 0%
	splash.SetProgress(0)
	view := splash.View()
	if !strings.Contains(view, "  0%") {
		t.Error("should show 0% at start")
	}

	// At 100%
	splash.SetProgress(1.0)
	view = splash.View()
	if !strings.Contains(view, "100%") {
		t.Error("should show 100% at end")
	}
}
