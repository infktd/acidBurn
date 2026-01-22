package ui

import (
	"fmt"
	"strings"
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

func TestParseLogTimestamp(t *testing.T) {
	tests := []struct {
		name    string
		line    string
		wantNil bool
	}{
		{
			name:    "ISO8601 timestamp",
			line:    "2026-01-21T10:30:45Z info: message",
			wantNil: false,
		},
		{
			name:    "timestamp with milliseconds",
			line:    "2026-01-21T10:30:45.123Z message",
			wantNil: false,
		},
		{
			name:    "no timestamp",
			line:    "plain log message",
			wantNil: true,
		},
		{
			name:    "invalid timestamp format",
			line:    "not-a-timestamp message",
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := ParseLogTimestamp(tt.line)
			if tt.wantNil && ok {
				t.Errorf("ParseLogTimestamp() = %v, want not found", result)
			}
			if !tt.wantNil && !ok {
				t.Error("ParseLogTimestamp() not found, want timestamp")
			}
		})
	}
}

func TestDetectLogLevel(t *testing.T) {
	tests := []struct {
		line string
		want LogLevel
	}{
		{"ERROR: something failed", LevelError},
		{"error: something failed", LevelError},
		{"WARN: be careful", LevelWarn},
		{"warning: be careful", LevelWarn},
		{"INFO: normal message", LevelInfo},
		{"info: normal message", LevelInfo},
		{"DEBUG: detailed info", LevelDebug},
		{"debug: detailed info", LevelDebug},
		{"plain message", LevelInfo}, // default
	}

	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			got := DetectLogLevel(tt.line)
			if got != tt.want {
				t.Errorf("DetectLogLevel(%q) = %v, want %v", tt.line, got, tt.want)
			}
		})
	}
}

func TestLogBufferTail(t *testing.T) {
	buf := NewLogBuffer(100)

	// Add some lines
	for i := 1; i <= 10; i++ {
		buf.Add(LogEntry{
			Message: fmt.Sprintf("line %d", i),
			Service: "test",
		})
	}

	// Tail 5 lines
	lines := buf.Tail(5)
	if len(lines) != 5 {
		t.Fatalf("Tail(5) returned %d lines, want 5", len(lines))
	}

	// Should be lines 6-10
	if !strings.Contains(lines[0].Message, "line 6") {
		t.Errorf("first tailed line = %q, want line 6", lines[0].Message)
	}
	if !strings.Contains(lines[4].Message, "line 10") {
		t.Errorf("last tailed line = %q, want line 10", lines[4].Message)
	}
}

func TestLogBufferCapacity(t *testing.T) {
	buf := NewLogBuffer(50)
	cap := buf.Capacity()
	if cap != 50 {
		t.Errorf("Capacity() = %d, want 50", cap)
	}
}
