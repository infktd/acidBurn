package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ProgressBar renders a visual progress indicator.
type ProgressBar struct {
	styles   *Styles
	width    int
	progress float64 // 0.0 to 1.0
	label    string
	showPct  bool
}

// NewProgressBar creates a new progress bar.
func NewProgressBar(styles *Styles, width int) *ProgressBar {
	return &ProgressBar{
		styles:  styles,
		width:   width,
		showPct: true,
	}
}

// SetProgress sets the progress value (0.0 to 1.0).
func (p *ProgressBar) SetProgress(value float64) {
	if value < 0 {
		value = 0
	}
	if value > 1 {
		value = 1
	}
	p.progress = value
}

// SetLabel sets the progress bar label.
func (p *ProgressBar) SetLabel(label string) {
	p.label = label
}

// SetWidth sets the width of the progress bar.
func (p *ProgressBar) SetWidth(width int) {
	p.width = width
}

// SetShowPercentage toggles percentage display.
func (p *ProgressBar) SetShowPercentage(show bool) {
	p.showPct = show
}

// Progress returns the current progress value.
func (p *ProgressBar) Progress() float64 {
	return p.progress
}

// View renders the progress bar.
func (p *ProgressBar) View() string {
	// Calculate bar width (accounting for brackets and percentage)
	pctText := ""
	if p.showPct {
		pctText = " 100%"
	}

	barWidth := p.width - 2 - len(pctText) // 2 for [ ]
	if barWidth < 10 {
		barWidth = 10
	}

	// Calculate filled portion
	filled := int(float64(barWidth) * p.progress)
	if filled > barWidth {
		filled = barWidth
	}

	// Build the bar
	filledChar := "█"
	emptyChar := "░"

	filledPart := strings.Repeat(filledChar, filled)
	emptyPart := strings.Repeat(emptyChar, barWidth-filled)

	// Get theme for colors
	theme := GetTheme("acid-green")
	if p.styles != nil {
		// Use style colors
	}

	filledStyle := lipgloss.NewStyle().Foreground(theme.Primary)
	emptyStyle := lipgloss.NewStyle().Foreground(theme.Muted)

	bar := "[" + filledStyle.Render(filledPart) + emptyStyle.Render(emptyPart) + "]"

	// Add percentage
	if p.showPct {
		pct := int(p.progress * 100)
		pctStyle := lipgloss.NewStyle().Foreground(theme.Secondary)
		bar += pctStyle.Render(" " + padLeft(pct, 3) + "%")
	}

	// Add label if present
	if p.label != "" {
		labelStyle := lipgloss.NewStyle().Foreground(theme.Secondary)
		return labelStyle.Render(p.label) + "\n" + bar
	}

	return bar
}

// padLeft pads a number with spaces on the left.
func padLeft(n, width int) string {
	s := ""
	for i := 0; i < width; i++ {
		s = " " + s
	}
	ns := string(rune('0'+n%10)) + s
	n /= 10
	for i := 1; i < width && n > 0; i++ {
		ns = string(rune('0'+n%10)) + ns[1:]
		n /= 10
	}
	return ns[:width]
}

// Spinner provides animated loading indicators.
type Spinner struct {
	frames  []string
	current int
}

// NewSpinner creates a new spinner with default frames.
func NewSpinner() *Spinner {
	return &Spinner{
		frames: []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
	}
}

// Frame returns the current spinner frame and advances.
func (s *Spinner) Frame() string {
	frame := s.frames[s.current]
	s.current = (s.current + 1) % len(s.frames)
	return frame
}

// Reset resets the spinner to the first frame.
func (s *Spinner) Reset() {
	s.current = 0
}
