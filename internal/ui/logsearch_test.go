package ui

import (
	"testing"
	"time"
)

func TestLogViewSearch(t *testing.T) {
	styles := NewStyles(GetTheme("matrix"))
	lv := NewLogView(styles, 80, 24)

	// Add some entries
	lv.AddEntry(LogEntry{Timestamp: time.Now(), Service: "api", Level: LevelInfo, Message: "Starting server"})
	lv.AddEntry(LogEntry{Timestamp: time.Now(), Service: "api", Level: LevelInfo, Message: "Listening on port 8080"})
	lv.AddEntry(LogEntry{Timestamp: time.Now(), Service: "api", Level: LevelError, Message: "Connection error"})
	lv.AddEntry(LogEntry{Timestamp: time.Now(), Service: "api", Level: LevelInfo, Message: "Request received"})

	// Search for "error"
	lv.SetSearch("error")

	if !lv.IsSearchActive() {
		t.Error("search should be active")
	}
	if lv.MatchCount() != 1 {
		t.Errorf("expected 1 match, got %d", lv.MatchCount())
	}
}

func TestLogViewSearchCaseInsensitive(t *testing.T) {
	styles := NewStyles(GetTheme("matrix"))
	lv := NewLogView(styles, 80, 24)

	lv.AddEntry(LogEntry{Timestamp: time.Now(), Service: "api", Level: LevelInfo, Message: "ERROR happened"})
	lv.AddEntry(LogEntry{Timestamp: time.Now(), Service: "api", Level: LevelInfo, Message: "error again"})

	lv.SetSearch("error")

	if lv.MatchCount() != 2 {
		t.Errorf("expected 2 matches (case-insensitive), got %d", lv.MatchCount())
	}
}

func TestLogViewNextPrevMatch(t *testing.T) {
	styles := NewStyles(GetTheme("matrix"))
	lv := NewLogView(styles, 80, 24)

	lv.AddEntry(LogEntry{Timestamp: time.Now(), Service: "a", Level: LevelInfo, Message: "match one"})
	lv.AddEntry(LogEntry{Timestamp: time.Now(), Service: "b", Level: LevelInfo, Message: "no hit"})
	lv.AddEntry(LogEntry{Timestamp: time.Now(), Service: "c", Level: LevelInfo, Message: "match two"})
	lv.AddEntry(LogEntry{Timestamp: time.Now(), Service: "d", Level: LevelInfo, Message: "match three"})

	lv.SetSearch("match")

	if lv.MatchCount() != 3 {
		t.Fatalf("expected 3 matches, got %d", lv.MatchCount())
	}

	// Should start at first match
	if lv.CurrentMatchIndex() != 1 {
		t.Errorf("expected current match 1, got %d", lv.CurrentMatchIndex())
	}

	lv.NextMatch()
	if lv.CurrentMatchIndex() != 2 {
		t.Errorf("expected current match 2, got %d", lv.CurrentMatchIndex())
	}

	lv.NextMatch()
	if lv.CurrentMatchIndex() != 3 {
		t.Errorf("expected current match 3, got %d", lv.CurrentMatchIndex())
	}

	// Wrap around
	lv.NextMatch()
	if lv.CurrentMatchIndex() != 1 {
		t.Errorf("expected wrap to 1, got %d", lv.CurrentMatchIndex())
	}

	lv.PrevMatch()
	if lv.CurrentMatchIndex() != 3 {
		t.Errorf("expected wrap back to 3, got %d", lv.CurrentMatchIndex())
	}
}

func TestLogViewFilterMode(t *testing.T) {
	styles := NewStyles(GetTheme("matrix"))
	lv := NewLogView(styles, 80, 24)

	lv.AddEntry(LogEntry{Timestamp: time.Now(), Service: "a", Level: LevelInfo, Message: "match"})
	lv.AddEntry(LogEntry{Timestamp: time.Now(), Service: "b", Level: LevelInfo, Message: "no"})
	lv.AddEntry(LogEntry{Timestamp: time.Now(), Service: "c", Level: LevelInfo, Message: "match again"})

	lv.SetSearch("match")
	lv.ToggleFilter()

	if !lv.IsFilterMode() {
		t.Error("expected filter mode on")
	}

	// In filter mode, only matching lines shown
	// View() should only render matching entries
	view := lv.View()
	if len(view) == 0 {
		t.Error("view should have content")
	}
}

func TestLogViewClearSearch(t *testing.T) {
	styles := NewStyles(GetTheme("matrix"))
	lv := NewLogView(styles, 80, 24)

	lv.AddEntry(LogEntry{Timestamp: time.Now(), Service: "a", Level: LevelInfo, Message: "test"})
	lv.SetSearch("test")

	if !lv.IsSearchActive() {
		t.Error("search should be active")
	}

	lv.ClearSearch()

	if lv.IsSearchActive() {
		t.Error("search should be cleared")
	}
	if lv.MatchCount() != 0 {
		t.Errorf("matches should be empty, got %d", lv.MatchCount())
	}
	if lv.IsFilterMode() {
		t.Error("filter mode should be off")
	}
}
