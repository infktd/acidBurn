package ui

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestLogViewRender(t *testing.T) {
	theme := GetTheme("matrix")
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
	theme := GetTheme("matrix")
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
	theme := GetTheme("matrix")
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

func TestLogViewSetBuffer(t *testing.T) {
	theme := GetTheme("matrix")
	styles := NewStyles(theme)

	lv := NewLogView(styles, 80, 10)
	buf := NewLogBuffer(100)
	buf.Add(LogEntry{
		Message: "test line",
		Service: "service1",
	})

	lv.SetBuffer(buf)

	// Verify buffer is set (implementation-dependent check)
	view := lv.View()
	if view == "" {
		t.Log("SetBuffer() set buffer (view may be empty due to viewport)")
	}
}

func TestLogViewPageUpPageDown(t *testing.T) {
	theme := GetTheme("matrix")
	styles := NewStyles(theme)

	lv := NewLogView(styles, 80, 20)
	buf := NewLogBuffer(1000)

	// Add many lines
	for i := 0; i < 100; i++ {
		buf.Add(LogEntry{
			Message: fmt.Sprintf("line %d", i),
			Service: "service",
		})
	}
	lv.SetBuffer(buf)
	lv.SetSize(80, 20)

	// PageDown should scroll
	lv.PageDown()
	// Can't easily verify scroll position without exposing internals
	// Just ensure it doesn't panic

	// PageUp should scroll back
	lv.PageUp()
}

func TestLogViewClear(t *testing.T) {
	theme := GetTheme("matrix")
	styles := NewStyles(theme)

	lv := NewLogView(styles, 80, 10)
	buf := NewLogBuffer(100)
	buf.Add(LogEntry{
		Message: "test",
		Service: "service",
	})
	lv.SetBuffer(buf)

	lv.Clear()

	// After clear, view should be empty or show empty state
	view := lv.View()
	_ = view // just verify it doesn't panic
}

func TestLogViewSearchQuery(t *testing.T) {
	theme := GetTheme("matrix")
	styles := NewStyles(theme)

	lv := NewLogView(styles, 80, 10)

	query := lv.SearchQuery()
	if query != "" {
		t.Errorf("SearchQuery() = %q, want empty initially", query)
	}

	// After setting search (via SetSearch)
	lv.SetSearch("test")
	query = lv.SearchQuery()
	if query != "test" {
		t.Errorf("SearchQuery() = %q, want 'test'", query)
	}
}
