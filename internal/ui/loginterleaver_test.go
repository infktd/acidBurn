package ui

import (
	"testing"
	"time"
)

func TestLogInterleaverSortsByTimestamp(t *testing.T) {
	output := NewLogBuffer(100)
	li := NewLogInterleaver(output)

	now := time.Now()

	// Add entries out of order
	li.Add(LogEntry{Timestamp: now.Add(2 * time.Millisecond), Service: "c", Message: "third"})
	li.Add(LogEntry{Timestamp: now, Service: "a", Message: "first"})
	li.Add(LogEntry{Timestamp: now.Add(1 * time.Millisecond), Service: "b", Message: "second"})

	// Manually trigger flush (don't start background goroutine for this test)
	li.flush()

	lines := output.Lines()
	if len(lines) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(lines))
	}

	// Verify order
	if lines[0].Message != "first" {
		t.Errorf("expected first, got %s", lines[0].Message)
	}
	if lines[1].Message != "second" {
		t.Errorf("expected second, got %s", lines[1].Message)
	}
	if lines[2].Message != "third" {
		t.Errorf("expected third, got %s", lines[2].Message)
	}
}

func TestLogInterleaverConcurrentAdd(t *testing.T) {
	output := NewLogBuffer(1000)
	li := NewLogInterleaver(output)

	done := make(chan struct{})

	// Spawn multiple goroutines adding entries
	for i := 0; i < 10; i++ {
		go func(service string) {
			for j := 0; j < 100; j++ {
				li.Add(LogEntry{
					Timestamp: time.Now(),
					Service:   service,
					Message:   "test",
				})
			}
			done <- struct{}{}
		}(string(rune('a' + i)))
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	li.flush()

	if output.Len() != 1000 {
		t.Errorf("expected 1000 entries, got %d", output.Len())
	}
}

func TestLogInterleaverStartStop(t *testing.T) {
	output := NewLogBuffer(100)
	li := NewLogInterleaver(output)

	li.Start()

	// Add entries
	li.Add(LogEntry{Timestamp: time.Now(), Service: "a", Message: "test"})

	// Wait for at least one flush cycle
	time.Sleep(100 * time.Millisecond)

	li.Stop()

	if output.Len() == 0 {
		t.Error("expected entries to be flushed")
	}
}

func TestLogInterleaverFlushClearsPending(t *testing.T) {
	output := NewLogBuffer(100)
	li := NewLogInterleaver(output)

	li.Add(LogEntry{Timestamp: time.Now(), Service: "a", Message: "first"})
	li.flush()

	li.Add(LogEntry{Timestamp: time.Now(), Service: "b", Message: "second"})
	li.flush()

	if output.Len() != 2 {
		t.Errorf("expected 2 entries, got %d", output.Len())
	}
}
