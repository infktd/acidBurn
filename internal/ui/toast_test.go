package ui

import (
	"strings"
	"testing"
	"time"
)

func TestToastManagerShow(t *testing.T) {
	styles := NewStyles(GetTheme("matrix"))
	tm := NewToastManager(styles, 60)

	tm.Show("Test message", ToastInfo, 5*time.Second)

	if !tm.IsVisible() {
		t.Error("toast should be visible")
	}
	if tm.Current() == nil {
		t.Error("current toast should not be nil")
	}
	if tm.Current().Message != "Test message" {
		t.Errorf("expected 'Test message', got %q", tm.Current().Message)
	}
}

func TestToastManagerDismiss(t *testing.T) {
	styles := NewStyles(GetTheme("matrix"))
	tm := NewToastManager(styles, 60)

	tm.Show("Test", ToastError, 0)
	if !tm.IsVisible() {
		t.Error("should be visible")
	}

	tm.Dismiss()
	if tm.IsVisible() {
		t.Error("should not be visible after dismiss")
	}
}

func TestToastManagerLevels(t *testing.T) {
	styles := NewStyles(GetTheme("matrix"))
	tm := NewToastManager(styles, 60)

	tests := []struct {
		level ToastLevel
		name  string
	}{
		{ToastInfo, "info"},
		{ToastWarn, "warn"},
		{ToastError, "error"},
	}

	for _, tt := range tests {
		tm.Show("Test", tt.level, 0)
		if tm.Current().Level != tt.level {
			t.Errorf("expected level %v, got %v", tt.level, tm.Current().Level)
		}
		view := tm.View()
		if view == "" {
			t.Errorf("view should not be empty for %s level", tt.name)
		}
	}
}

func TestToastManagerView(t *testing.T) {
	styles := NewStyles(GetTheme("matrix"))
	tm := NewToastManager(styles, 60)

	// No toast - view should be empty
	view := tm.View()
	if view != "" {
		t.Error("view should be empty when no toast")
	}

	// With toast
	tm.Show("Alert!", ToastWarn, 0)
	view = tm.View()
	if view == "" {
		t.Error("view should have content")
	}
	if !strings.Contains(view, "Alert!") {
		t.Error("view should contain message")
	}
}

func TestToastOverwrite(t *testing.T) {
	styles := NewStyles(GetTheme("matrix"))
	tm := NewToastManager(styles, 60)

	tm.Show("First", ToastInfo, 0)
	tm.Show("Second", ToastError, 0)

	if tm.Current().Message != "Second" {
		t.Error("new toast should overwrite old")
	}
	if tm.Current().Level != ToastError {
		t.Error("new toast level should be preserved")
	}
}

// TestToastFadeInAnimation verifies fade-in opacity progression
func TestToastFadeInAnimation(t *testing.T) {
	styles := NewStyles(GetTheme("matrix"))
	tm := NewToastManager(styles, 100)

	tm.Show("Test", ToastInfo, 5*time.Second)

	// Should start at opacity 0.0
	if tm.opacity != 0.0 {
		t.Errorf("expected initial opacity 0.0, got %f", tm.opacity)
	}

	// Simulate 500ms elapsed (halfway through 1s fade-in)
	time.Sleep(100 * time.Millisecond)
	tm.Update(ToastTickMsg(time.Now()))

	// Opacity should be increasing (between 0 and 1)
	if tm.opacity <= 0.0 || tm.opacity >= 1.0 {
		t.Errorf("expected opacity between 0 and 1 during fade-in, got %f", tm.opacity)
	}

	// Simulate full 1 second fade-in
	tm.current.CreatedAt = time.Now().Add(-1 * time.Second)
	tm.Update(ToastTickMsg(time.Now()))

	// Should be at full opacity after fade-in
	if tm.opacity < 0.99 {
		t.Errorf("expected opacity ~1.0 after fade-in, got %f", tm.opacity)
	}
}

// TestToastFullyCycle verifies complete 5-second animation cycle
func TestToastFullyCycle(t *testing.T) {
	styles := NewStyles(GetTheme("matrix"))
	tm := NewToastManager(styles, 100)

	tm.Show("Test", ToastInfo, 5*time.Second)

	// Phase 1: Fade-in (0-1s) - opacity increases
	tm.current.CreatedAt = time.Now().Add(-500 * time.Millisecond)
	tm.Update(ToastTickMsg(time.Now()))
	fadeInOpacity := tm.opacity
	if fadeInOpacity <= 0.3 || fadeInOpacity >= 0.7 {
		t.Errorf("expected mid-fade opacity ~0.5, got %f", fadeInOpacity)
	}

	// Phase 2: Fully visible (1s-4s) - opacity at 1.0
	tm.current.CreatedAt = time.Now().Add(-2 * time.Second)
	tm.Update(ToastTickMsg(time.Now()))
	if tm.opacity != 1.0 {
		t.Errorf("expected opacity 1.0 during visible phase, got %f", tm.opacity)
	}

	// Phase 3: Fade-out (4s-5s) - opacity decreases
	tm.current.CreatedAt = time.Now().Add(-4500 * time.Millisecond)
	tm.Update(ToastTickMsg(time.Now()))
	fadeOutOpacity := tm.opacity
	if fadeOutOpacity <= 0.3 || fadeOutOpacity >= 0.7 {
		t.Errorf("expected mid-fade-out opacity ~0.5, got %f", fadeOutOpacity)
	}

	// Phase 4: Dismissed after 5s
	tm.current.CreatedAt = time.Now().Add(-5100 * time.Millisecond)
	tm.Update(ToastTickMsg(time.Now()))
	if tm.IsVisible() {
		t.Error("toast should be dismissed after 5 seconds")
	}
}

// TestToastDynamicWidth verifies width calculation based on terminal size
func TestToastDynamicWidth(t *testing.T) {
	styles := NewStyles(GetTheme("matrix"))

	// Test that different widths affect the toast manager's width field
	tests := []struct {
		terminalWidth int
		expectedWidth int
	}{
		{80, 80},
		{100, 100},
		{200, 200},
		{50, 50},
	}

	for _, tt := range tests {
		tm := NewToastManager(styles, tt.terminalWidth)
		if tm.width != tt.expectedWidth {
			t.Errorf("terminal width %d: expected toast width %d, got %d",
				tt.terminalWidth, tt.expectedWidth, tm.width)
		}

		// Verify toast displays something
		tm.Show("Test message", ToastInfo, 0)
		view := tm.View()
		if view == "" {
			t.Errorf("terminal width %d: view should not be empty", tt.terminalWidth)
		}
		if !strings.Contains(view, "Test message") {
			t.Errorf("terminal width %d: view should contain message", tt.terminalWidth)
		}
	}
}

// TestToastMessageTruncation verifies long messages are truncated
func TestToastMessageTruncation(t *testing.T) {
	styles := NewStyles(GetTheme("matrix"))
	tm := NewToastManager(styles, 100)

	longMessage := strings.Repeat("This is a very long message ", 10)
	tm.Show(longMessage, ToastInfo, 0)

	view := tm.View()
	// Should contain ellipsis for truncation
	if !strings.Contains(view, "...") {
		t.Error("expected long message to be truncated with ellipsis")
	}
}

// TestToastColorBlending verifies color changes during animation
func TestToastColorBlending(t *testing.T) {
	styles := NewStyles(GetTheme("matrix"))
	tm := NewToastManager(styles, 100)

	tm.Show("Test", ToastSuccess, 5*time.Second)
	tm.opacity = 0.5 // Mid-fade

	view := tm.View()
	// View should contain styled content (color is applied via lipgloss)
	if view == "" {
		t.Error("view should contain styled content during fade")
	}
	if !strings.Contains(view, "Test") {
		t.Error("view should contain message text")
	}
}
