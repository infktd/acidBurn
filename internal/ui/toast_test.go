package ui

import (
	"strings"
	"testing"
	"time"
)

func TestToastManagerShow(t *testing.T) {
	styles := NewStyles(GetTheme("acid-green"))
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
	styles := NewStyles(GetTheme("acid-green"))
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
	styles := NewStyles(GetTheme("acid-green"))
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
	styles := NewStyles(GetTheme("acid-green"))
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
	styles := NewStyles(GetTheme("acid-green"))
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
