package ui

import (
	"regexp"
	"strings"
	"sync"
	"time"
)

// Common timestamp formats to try when parsing log lines
var timestampFormats = []string{
	// ISO formats
	"2006-01-02T15:04:05.000Z",
	"2006-01-02T15:04:05Z",
	"2006-01-02T15:04:05",
	"2006-01-02 15:04:05.000",
	"2006-01-02 15:04:05",
	// Date with timezone
	"Mon Jan 2 03:04:05 PM MST 2006",
	"Mon Jan 02 03:04:05 PM MST 2006",
	"Jan 2 15:04:05 2006",
	"Jan 02 15:04:05 2006",
	// Time only (assume today)
	"15:04:05.000",
	"15:04:05",
	"03:04:05 PM",
}

// Regex patterns to find timestamps in log lines
var timestampPatterns = []*regexp.Regexp{
	// ISO: 2006-01-02T15:04:05 or 2006-01-02 15:04:05
	regexp.MustCompile(`\d{4}-\d{2}-\d{2}[T ]\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:Z|[+-]\d{2}:?\d{2})?`),
	// Date: Mon Jan 2 03:04:05 PM MST 2006
	regexp.MustCompile(`(?:Mon|Tue|Wed|Thu|Fri|Sat|Sun)\s+(?:Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec)\s+\d{1,2}\s+\d{1,2}:\d{2}:\d{2}\s+(?:AM|PM)\s+\w+\s+\d{4}`),
	// Time only: 15:04:05
	regexp.MustCompile(`\b\d{2}:\d{2}:\d{2}(?:\.\d+)?\b`),
}

// ParseLogTimestamp attempts to extract a timestamp from a log line.
// Returns the parsed time and true if successful, or zero time and false if not found.
func ParseLogTimestamp(line string) (time.Time, bool) {
	for _, pattern := range timestampPatterns {
		match := pattern.FindString(line)
		if match == "" {
			continue
		}

		for _, format := range timestampFormats {
			if t, err := time.Parse(format, match); err == nil {
				// For time-only formats, use today's date
				if t.Year() == 0 {
					now := time.Now()
					t = time.Date(now.Year(), now.Month(), now.Day(),
						t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), now.Location())
				}
				return t, true
			}
		}
	}

	return time.Time{}, false
}

// LogLevel represents the severity of a log entry.
type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelWarn
	LevelError
)

func (l LogLevel) String() string {
	switch l {
	case LevelDebug:
		return "debug"
	case LevelInfo:
		return "info"
	case LevelWarn:
		return "warn"
	case LevelError:
		return "error"
	default:
		return "unknown"
	}
}

// Log level detection patterns (case-insensitive keywords)
var levelPatterns = map[LogLevel][]string{
	LevelError: {"error", "err", "fatal", "critical", "panic", "exception", "fail"},
	LevelWarn:  {"warn", "warning", "caution"},
	LevelDebug: {"debug", "trace", "verbose"},
	LevelInfo:  {"info", "notice"},
}

// DetectLogLevel attempts to detect the log level from a message.
// Returns LevelInfo if no level can be detected.
func DetectLogLevel(message string) LogLevel {
	msgLower := strings.ToLower(message)

	// Check for explicit level indicators like [ERROR], ERROR:, level=error
	// Priority: Error > Warn > Debug > Info

	// Check error first (highest priority)
	for _, pattern := range levelPatterns[LevelError] {
		if strings.Contains(msgLower, pattern) {
			return LevelError
		}
	}

	// Check warn
	for _, pattern := range levelPatterns[LevelWarn] {
		if strings.Contains(msgLower, pattern) {
			return LevelWarn
		}
	}

	// Check debug
	for _, pattern := range levelPatterns[LevelDebug] {
		if strings.Contains(msgLower, pattern) {
			return LevelDebug
		}
	}

	// Default to info
	return LevelInfo
}

// LogEntry represents a single log line.
type LogEntry struct {
	Timestamp time.Time
	Service   string
	Level     LogLevel
	Message   string
}

// LogBuffer is a circular buffer for log entries.
type LogBuffer struct {
	entries  []LogEntry
	capacity int
	head     int
	size     int
	mu       sync.RWMutex
}

// NewLogBuffer creates a new log buffer with the given capacity.
func NewLogBuffer(capacity int) *LogBuffer {
	return &LogBuffer{
		entries:  make([]LogEntry, capacity),
		capacity: capacity,
	}
}

// Add appends a log entry to the buffer.
func (b *LogBuffer) Add(entry LogEntry) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.entries[b.head] = entry
	b.head = (b.head + 1) % b.capacity
	if b.size < b.capacity {
		b.size++
	}
}

// Lines returns all log entries in order (oldest first).
func (b *LogBuffer) Lines() []LogEntry {
	b.mu.RLock()
	defer b.mu.RUnlock()

	result := make([]LogEntry, b.size)
	if b.size == 0 {
		return result
	}

	start := 0
	if b.size == b.capacity {
		start = b.head
	}

	for i := 0; i < b.size; i++ {
		idx := (start + i) % b.capacity
		result[i] = b.entries[idx]
	}

	return result
}

// Tail returns the last n entries.
func (b *LogBuffer) Tail(n int) []LogEntry {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if n > b.size {
		n = b.size
	}
	if n == 0 {
		return []LogEntry{}
	}

	result := make([]LogEntry, n)
	start := (b.head - n + b.capacity) % b.capacity
	if b.size < b.capacity {
		start = b.size - n
	}

	for i := 0; i < n; i++ {
		idx := (start + i) % b.capacity
		if b.size < b.capacity {
			idx = start + i
		}
		result[i] = b.entries[idx]
	}

	return result
}

// Len returns the number of entries in the buffer.
func (b *LogBuffer) Len() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.size
}

// Clear removes all entries from the buffer.
func (b *LogBuffer) Clear() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.head = 0
	b.size = 0
}

// Capacity returns the maximum capacity of the buffer.
func (b *LogBuffer) Capacity() int {
	return b.capacity
}
