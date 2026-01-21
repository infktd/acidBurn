package ui

import (
	"strings"
	"testing"
	"time"
)

func TestLogViewRender(t *testing.T) {
	theme := GetTheme("acid-green")
	styles := NewStyles(theme)

	lv := NewLogView(styles, 80, 10)
	lv.AddEntry(LogEntry{
		Timestamp: time.Date(2026, 1, 20, 14, 30, 0, 0, time.UTC),
		Service:   "postgres",
		Level:     LevelInfo,
		Message:   "database ready",
	})

	output := lv.View()
	if !strings.Contains(output, "database ready") {
		t.Errorf("Expected output to contain 'database ready', got: %s", output)
	}
}

func TestLogViewFollowMode(t *testing.T) {
	theme := GetTheme("acid-green")
	styles := NewStyles(theme)

	lv := NewLogView(styles, 80, 5)

	// Should start in follow mode
	if !lv.IsFollowing() {
		t.Error("Expected follow mode to be enabled by default")
	}

	// Toggle off
	lv.ToggleFollow()
	if lv.IsFollowing() {
		t.Error("Expected follow mode to be disabled after toggle")
	}
}

func TestLogViewScrolling(t *testing.T) {
	theme := GetTheme("acid-green")
	styles := NewStyles(theme)

	lv := NewLogView(styles, 80, 3) // Only 3 visible lines
	lv.SetFollow(false)

	// Add more entries than can be displayed
	for i := 0; i < 10; i++ {
		lv.AddEntry(LogEntry{Message: "line"})
	}

	// Should be able to scroll
	lv.ScrollUp()
	lv.ScrollDown()
	lv.ScrollToTop()
	lv.ScrollToBottom()
}
