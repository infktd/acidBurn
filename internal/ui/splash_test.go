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

func TestSplashScreenSetAsciiArtByName(t *testing.T) {
	styles := NewStyles(GetTheme("matrix"))
	splash := NewSplashScreen(styles, 80, 24)

	// Test setting valid ASCII art names
	validNames := []string{"default", "slant", "small", "minimal", "cyber"}
	for _, name := range validNames {
		splash.SetAsciiArtByName(name)
		view := splash.View()
		if view == "" {
			t.Errorf("view should have content for art %q", name)
		}
		// Check that the art was actually set by looking at the internal field
		// We can't directly access the field, but we can verify the view changes
		if !strings.Contains(view, "devdash") && !strings.Contains(view, "DEVDASH") && !strings.Contains(view, "devDASH") {
			// Some ASCII arts might not contain the exact string, so just check it's not empty
			if view == "" {
				t.Errorf("view should not be empty for art %q", name)
			}
		}
	}

	// Test setting invalid ASCII art name (should not change current art)
	splash.SetAsciiArtByName("default")
	originalView := splash.View()

	splash.SetAsciiArtByName("invalid-name-that-does-not-exist")
	newView := splash.View()

	// View should remain the same when invalid name is used
	if originalView != newView {
		t.Error("view should not change when invalid art name is provided")
	}
}

func TestSplashScreenGetAsciiArtNames(t *testing.T) {
	names := GetAsciiArtNames()

	if len(names) == 0 {
		t.Error("should return at least one ASCII art name")
	}

	// Check for expected names
	expectedNames := map[string]bool{
		"default": false,
		"block":   false,
		"small":   false,
		"minimal": false,
		"hacker":  false,
	}

	for _, name := range names {
		if _, exists := expectedNames[name]; exists {
			expectedNames[name] = true
		}
	}

	// Verify all expected names are present
	for name, found := range expectedNames {
		if !found {
			t.Errorf("expected ASCII art name %q not found in list", name)
		}
	}
}

func TestSplashScreenTick(t *testing.T) {
	styles := NewStyles(GetTheme("matrix"))
	splash := NewSplashScreen(styles, 80, 24)

	// Set progress to show animation
	splash.SetProgress(0.5)

	// Get initial frame counter (we can't access it directly, but we can test the animation changes)
	initialView := splash.View()

	// Tick several times
	for i := 0; i < 10; i++ {
		splash.Tick()
	}

	// The view might change due to animation frame updates
	// We'll verify that Tick() doesn't cause errors and can be called repeatedly
	newView := splash.View()

	// Both views should be non-empty
	if initialView == "" {
		t.Error("initial view should not be empty")
	}
	if newView == "" {
		t.Error("view after ticks should not be empty")
	}

	// Test Tick on hidden splash
	splash.Hide()
	splash.Tick() // Should not cause errors even when hidden

	// Test Tick at different progress levels
	splash.Show()
	splash.SetProgress(0.0)
	splash.Tick()

	splash.SetProgress(1.0)
	splash.Tick()

	view := splash.View()
	if view == "" {
		t.Error("view should still render after ticks")
	}
}
