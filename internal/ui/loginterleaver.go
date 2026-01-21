package ui

import (
	"sort"
	"sync"
	"time"
)

// LogInterleaver collects log entries from multiple sources,
// sorts them by timestamp, and outputs to a unified buffer.
type LogInterleaver struct {
	buffer    []LogEntry
	output    *LogBuffer
	flushTick *time.Ticker
	mu        sync.Mutex
	done      chan struct{}
	wg        sync.WaitGroup
	started   bool
}

// NewLogInterleaver creates an interleaver that outputs to the given buffer.
func NewLogInterleaver(output *LogBuffer) *LogInterleaver {
	return &LogInterleaver{
		buffer: make([]LogEntry, 0, 100),
		output: output,
		done:   make(chan struct{}),
	}
}

// Add queues an entry for interleaving.
func (li *LogInterleaver) Add(entry LogEntry) {
	li.mu.Lock()
	defer li.mu.Unlock()
	li.buffer = append(li.buffer, entry)
}

// Start begins the background flush goroutine.
func (li *LogInterleaver) Start() {
	li.mu.Lock()
	if li.started {
		li.mu.Unlock()
		return
	}
	li.started = true
	li.mu.Unlock()

	li.flushTick = time.NewTicker(50 * time.Millisecond)
	li.wg.Add(1)
	go func() {
		defer li.wg.Done()
		for {
			select {
			case <-li.flushTick.C:
				li.flush()
			case <-li.done:
				li.flushTick.Stop()
				li.flush() // Final flush
				return
			}
		}
	}()
}

// Stop halts the background goroutine.
func (li *LogInterleaver) Stop() {
	li.mu.Lock()
	if !li.started {
		li.mu.Unlock()
		return
	}
	li.mu.Unlock()

	close(li.done)
	li.wg.Wait()
}

// flush sorts pending entries by timestamp and writes to output.
func (li *LogInterleaver) flush() {
	li.mu.Lock()
	if len(li.buffer) == 0 {
		li.mu.Unlock()
		return
	}

	// Sort by timestamp
	sort.Slice(li.buffer, func(i, j int) bool {
		return li.buffer[i].Timestamp.Before(li.buffer[j].Timestamp)
	})

	// Copy entries to output
	entries := li.buffer
	li.buffer = make([]LogEntry, 0, 100)
	li.mu.Unlock()

	// Add to output buffer (outside lock to avoid potential deadlock)
	for _, e := range entries {
		li.output.Add(e)
	}
}
