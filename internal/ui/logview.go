package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// LogView renders log entries with formatting and scroll support.
type LogView struct {
	buffer  *LogBuffer
	styles  *Styles
	width   int
	height  int
	offset  int    // Scroll offset from bottom
	follow  bool   // Auto-scroll to bottom
	service string // Filter to specific service, empty = all
	// Search fields
	searchQuery  string
	searchActive bool
	filterMode   bool  // Only show matching lines
	matches      []int // Line indices that match
	matchIndex   int   // Current match cursor (0-indexed internally)
}

// NewLogView creates a new log view component.
func NewLogView(styles *Styles, width, height int) *LogView {
	return &LogView{
		buffer: NewLogBuffer(10000),
		styles: styles,
		width:  width,
		height: height,
		follow: true,
	}
}

// AddEntry adds a log entry to the view.
func (lv *LogView) AddEntry(entry LogEntry) {
	lv.buffer.Add(entry)
	if lv.follow {
		lv.offset = 0
	}
}

// SetBuffer replaces the internal buffer (for unified view).
func (lv *LogView) SetBuffer(buf *LogBuffer) {
	lv.buffer = buf
}

// SetService filters logs to a specific service (empty = all).
func (lv *LogView) SetService(service string) {
	lv.service = service
	lv.offset = 0
}

// GetService returns the current service filter.
func (lv *LogView) GetService() string {
	return lv.service
}

// ScrollInfo returns current scroll position info (current line, total lines).
func (lv *LogView) ScrollInfo() (int, int) {
	all := lv.buffer.Lines()

	// Apply service filter
	if lv.service != "" {
		filtered := make([]LogEntry, 0)
		for _, e := range all {
			if e.Service == lv.service {
				filtered = append(filtered, e)
			}
		}
		all = filtered
	}

	total := len(all)
	if total == 0 {
		return 0, 0
	}

	// Current position (from bottom, inverted for display)
	currentLine := total - lv.offset
	if currentLine > total {
		currentLine = total
	}
	if currentLine < 1 {
		currentLine = 1
	}

	return currentLine, total
}

// SetSize updates the viewport dimensions.
func (lv *LogView) SetSize(width, height int) {
	lv.width = width
	lv.height = height
}

// View renders the log viewport.
func (lv *LogView) View() string {
	lines := lv.getVisibleLines()

	var sb strings.Builder
	for _, entry := range lines {
		sb.WriteString(lv.formatEntry(entry))
		sb.WriteString("\n")
	}

	// Pad to fill height
	lineCount := len(lines)
	for i := lineCount; i < lv.height; i++ {
		sb.WriteString("\n")
	}

	return sb.String()
}

func (lv *LogView) getVisibleLines() []LogEntry {
	all := lv.buffer.Lines()

	// Filter by service if set
	if lv.service != "" {
		filtered := make([]LogEntry, 0)
		for _, e := range all {
			if e.Service == lv.service {
				filtered = append(filtered, e)
			}
		}
		all = filtered
	}

	// Filter by search query if filter mode is active
	if lv.filterMode && lv.searchActive && lv.searchQuery != "" {
		queryLower := strings.ToLower(lv.searchQuery)
		filtered := make([]LogEntry, 0)
		for _, e := range all {
			if strings.Contains(strings.ToLower(e.Message), queryLower) {
				filtered = append(filtered, e)
			}
		}
		all = filtered
	}

	total := len(all)
	if total == 0 {
		return []LogEntry{}
	}

	// Calculate visible range
	visibleHeight := lv.height
	if visibleHeight > total {
		visibleHeight = total
	}

	end := total - lv.offset
	if end > total {
		end = total
	}
	start := end - visibleHeight
	if start < 0 {
		start = 0
	}

	return all[start:end]
}

func (lv *LogView) formatEntry(entry LogEntry) string {
	// Timestamp (when we received the log)
	timestamp := lv.styles.LogTimestamp.Render(entry.Timestamp.Format("15:04:05"))

	// Strip existing service prefix from message (e.g., "[worker] ...")
	msg := entry.Message
	if entry.Service != "" {
		prefix := fmt.Sprintf("[%s] ", entry.Service)
		msg = strings.TrimPrefix(msg, prefix)
	}

	// Format message with level colorization and search highlighting
	message := lv.formatMessageWithLevel(msg, entry.Level)

	// Add service prefix for unified view (colored tag)
	if lv.service == "" && entry.Service != "" {
		serviceTag := lipgloss.NewStyle().
			Foreground(lv.getServiceColor(entry.Service)).
			Render(fmt.Sprintf("[%s]", strings.ToUpper(entry.Service)))
		return fmt.Sprintf("%s %s %s", timestamp, serviceTag, message)
	}

	return fmt.Sprintf("%s %s", timestamp, message)
}

// getLevelStyle returns the style for a log level.
func (lv *LogView) getLevelStyle(level LogLevel) lipgloss.Style {
	switch level {
	case LevelError:
		return lv.styles.LogLevelError
	case LevelWarn:
		return lv.styles.LogLevelWarn
	case LevelDebug:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#6272A4")) // Muted blue/gray
	default:
		return lv.styles.LogLine
	}
}

// formatMessageWithLevel formats a message with level colors and search highlighting.
func (lv *LogView) formatMessageWithLevel(msg string, level LogLevel) string {
	baseStyle := lv.getLevelStyle(level)

	if !lv.searchActive || lv.searchQuery == "" {
		return baseStyle.Render(msg)
	}

	// Case-insensitive search for highlighting
	queryLower := strings.ToLower(lv.searchQuery)
	msgLower := strings.ToLower(msg)

	idx := strings.Index(msgLower, queryLower)
	if idx == -1 {
		return baseStyle.Render(msg)
	}

	// Highlight style: bold with reverse video (background highlight)
	highlightStyle := lipgloss.NewStyle().
		Bold(true).
		Reverse(true)

	// Build highlighted message by finding all occurrences
	var result strings.Builder
	pos := 0
	for {
		idx = strings.Index(msgLower[pos:], queryLower)
		if idx == -1 {
			// No more matches, append remaining text
			result.WriteString(baseStyle.Render(msg[pos:]))
			break
		}
		// Append text before match
		if idx > 0 {
			result.WriteString(baseStyle.Render(msg[pos : pos+idx]))
		}
		// Append highlighted match (using original case from msg)
		matchEnd := pos + idx + len(lv.searchQuery)
		result.WriteString(highlightStyle.Render(msg[pos+idx : matchEnd]))
		pos = matchEnd
	}

	return result.String()
}

// formatMessage formats a message with search highlighting if search is active.
func (lv *LogView) formatMessage(msg string) string {
	if !lv.searchActive || lv.searchQuery == "" {
		return lv.styles.LogLine.Render(msg)
	}

	// Case-insensitive search for highlighting
	queryLower := strings.ToLower(lv.searchQuery)
	msgLower := strings.ToLower(msg)

	idx := strings.Index(msgLower, queryLower)
	if idx == -1 {
		return lv.styles.LogLine.Render(msg)
	}

	// Highlight style: bold with reverse video (background highlight)
	highlightStyle := lipgloss.NewStyle().
		Bold(true).
		Reverse(true)

	// Build highlighted message by finding all occurrences
	var result strings.Builder
	pos := 0
	for {
		idx = strings.Index(msgLower[pos:], queryLower)
		if idx == -1 {
			// No more matches, append remaining text
			result.WriteString(lv.styles.LogLine.Render(msg[pos:]))
			break
		}
		// Append text before match
		if idx > 0 {
			result.WriteString(lv.styles.LogLine.Render(msg[pos : pos+idx]))
		}
		// Append highlighted match (using original case from msg)
		matchEnd := pos + idx + len(lv.searchQuery)
		result.WriteString(highlightStyle.Render(msg[pos+idx : matchEnd]))
		pos = matchEnd
	}

	return result.String()
}

func (lv *LogView) getServiceColor(service string) lipgloss.Color {
	// Simple hash-based color assignment
	colors := []lipgloss.Color{
		"#336791", // postgres blue
		"#00D8FF", // cyan
		"#98C379", // green
		"#E06C75", // red
		"#C678DD", // purple
		"#E5C07B", // yellow
		"#56B6C2", // teal
	}

	hash := 0
	for _, c := range service {
		hash = hash*31 + int(c)
	}
	if hash < 0 {
		hash = -hash
	}

	return colors[hash%len(colors)]
}

// Scroll operations

func (lv *LogView) ScrollUp() {
	lv.follow = false
	maxOffset := lv.buffer.Len() - lv.height
	if maxOffset < 0 {
		maxOffset = 0
	}
	if lv.offset < maxOffset {
		lv.offset++
	}
}

func (lv *LogView) ScrollDown() {
	if lv.offset > 0 {
		lv.offset--
	}
	if lv.offset == 0 {
		lv.follow = true
	}
}

func (lv *LogView) ScrollToTop() {
	lv.follow = false
	lv.offset = lv.buffer.Len() - lv.height
	if lv.offset < 0 {
		lv.offset = 0
	}
}

func (lv *LogView) ScrollToBottom() {
	lv.offset = 0
	lv.follow = true
}

func (lv *LogView) PageUp() {
	lv.follow = false
	lv.offset += lv.height
	maxOffset := lv.buffer.Len() - lv.height
	if lv.offset > maxOffset {
		lv.offset = maxOffset
	}
	if lv.offset < 0 {
		lv.offset = 0
	}
}

func (lv *LogView) PageDown() {
	lv.offset -= lv.height
	if lv.offset < 0 {
		lv.offset = 0
		lv.follow = true
	}
}

// Follow mode

func (lv *LogView) ToggleFollow() {
	lv.follow = !lv.follow
	if lv.follow {
		lv.offset = 0
	}
}

func (lv *LogView) SetFollow(f bool) {
	lv.follow = f
	if f {
		lv.offset = 0
	}
}

func (lv *LogView) IsFollowing() bool {
	return lv.follow
}

// Clear removes all log entries.
func (lv *LogView) Clear() {
	lv.buffer.Clear()
	lv.offset = 0
}

// Search methods

// SetSearch sets the search query and finds all matches.
func (lv *LogView) SetSearch(query string) {
	lv.searchQuery = query
	lv.searchActive = query != ""
	lv.matches = nil
	lv.matchIndex = 0

	if !lv.searchActive {
		return
	}

	// Find all matching line indices (case-insensitive)
	queryLower := strings.ToLower(query)
	all := lv.getFilteredLines()
	for i, entry := range all {
		if strings.Contains(strings.ToLower(entry.Message), queryLower) {
			lv.matches = append(lv.matches, i)
		}
	}

	// Jump to first match if found
	if len(lv.matches) > 0 {
		lv.scrollToMatch(0)
	}
}

// ClearSearch clears the search state.
func (lv *LogView) ClearSearch() {
	lv.searchQuery = ""
	lv.searchActive = false
	lv.filterMode = false
	lv.matches = nil
	lv.matchIndex = 0
}

// NextMatch jumps to the next match.
func (lv *LogView) NextMatch() {
	if len(lv.matches) == 0 {
		return
	}
	lv.matchIndex = (lv.matchIndex + 1) % len(lv.matches)
	lv.scrollToMatch(lv.matchIndex)
}

// PrevMatch jumps to the previous match.
func (lv *LogView) PrevMatch() {
	if len(lv.matches) == 0 {
		return
	}
	lv.matchIndex--
	if lv.matchIndex < 0 {
		lv.matchIndex = len(lv.matches) - 1
	}
	lv.scrollToMatch(lv.matchIndex)
}

// ToggleFilter toggles filter mode (show only matching lines).
func (lv *LogView) ToggleFilter() {
	lv.filterMode = !lv.filterMode
	lv.offset = 0
}

// IsSearchActive returns true if search is active.
func (lv *LogView) IsSearchActive() bool {
	return lv.searchActive
}

// IsFilterMode returns true if filter mode is on.
func (lv *LogView) IsFilterMode() bool {
	return lv.filterMode
}

// SearchQuery returns the current search query.
func (lv *LogView) SearchQuery() string {
	return lv.searchQuery
}

// MatchCount returns the total number of matches.
func (lv *LogView) MatchCount() int {
	return len(lv.matches)
}

// CurrentMatchIndex returns the current match position (1-indexed for display).
func (lv *LogView) CurrentMatchIndex() int {
	if len(lv.matches) == 0 {
		return 0
	}
	return lv.matchIndex + 1
}

// scrollToMatch scrolls the view to make the match at the given index visible.
func (lv *LogView) scrollToMatch(idx int) {
	if idx < 0 || idx >= len(lv.matches) {
		return
	}

	lineIdx := lv.matches[idx]
	total := len(lv.getFilteredLines())

	// Calculate offset to center the match in view
	// offset is distance from bottom, so offset = total - lineIdx - height/2
	lv.offset = total - lineIdx - lv.height/2
	if lv.offset < 0 {
		lv.offset = 0
	}
	maxOffset := total - lv.height
	if maxOffset < 0 {
		maxOffset = 0
	}
	if lv.offset > maxOffset {
		lv.offset = maxOffset
	}
	lv.follow = false
}

// getFilteredLines returns lines filtered by service (but not by search filter mode).
func (lv *LogView) getFilteredLines() []LogEntry {
	all := lv.buffer.Lines()

	// Filter by service if set
	if lv.service != "" {
		filtered := make([]LogEntry, 0)
		for _, e := range all {
			if e.Service == lv.service {
				filtered = append(filtered, e)
			}
		}
		all = filtered
	}

	return all
}
