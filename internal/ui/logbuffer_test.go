package ui

import (
	"testing"
	"time"
)

func TestLogBufferAddAndGet(t *testing.T) {
	buf := NewLogBuffer(100)

	buf.Add(LogEntry{
		Timestamp: time.Now(),
		Service:   "postgres",
		Level:     LevelInfo,
		Message:   "database ready",
	})

	lines := buf.Lines()
	if len(lines) != 1 {
		t.Fatalf("Expected 1 line, got %d", len(lines))
	}
	if lines[0].Message != "database ready" {
		t.Errorf("Expected 'database ready', got %q", lines[0].Message)
	}
}

func TestLogBufferCircular(t *testing.T) {
	buf := NewLogBuffer(3) // Small buffer

	buf.Add(LogEntry{Message: "one"})
	buf.Add(LogEntry{Message: "two"})
	buf.Add(LogEntry{Message: "three"})
	buf.Add(LogEntry{Message: "four"}) // Should push out "one"

	lines := buf.Lines()
	if len(lines) != 3 {
		t.Fatalf("Expected 3 lines, got %d", len(lines))
	}
	if lines[0].Message != "two" {
		t.Errorf("Expected first line 'two', got %q", lines[0].Message)
	}
	if lines[2].Message != "four" {
		t.Errorf("Expected last line 'four', got %q", lines[2].Message)
	}
}

func TestLogBufferClear(t *testing.T) {
	buf := NewLogBuffer(100)
	buf.Add(LogEntry{Message: "test"})
	buf.Clear()

	if buf.Len() != 0 {
		t.Errorf("Expected 0 lines after clear, got %d", buf.Len())
	}
}

func TestLogLevelString(t *testing.T) {
	if LevelInfo.String() != "info" {
		t.Errorf("Expected 'info', got %q", LevelInfo.String())
	}
	if LevelWarn.String() != "warn" {
		t.Errorf("Expected 'warn', got %q", LevelWarn.String())
	}
	if LevelError.String() != "error" {
		t.Errorf("Expected 'error', got %q", LevelError.String())
	}
}
