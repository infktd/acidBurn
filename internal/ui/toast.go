package ui

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ToastLevel represents the severity of a toast notification.
type ToastLevel int

const (
	ToastInfo ToastLevel = iota
	ToastSuccess
	ToastWarn
	ToastError
)

// Toast represents a notification banner.
type Toast struct {
	Message   string
	Level     ToastLevel
	CreatedAt time.Time
	Duration  time.Duration // Auto-dismiss after this duration, 0 = no auto-dismiss
}

// ToastManager manages toast notification display.
type ToastManager struct {
	current   *Toast
	styles    *Styles
	width     int
	visible   bool
	dismissed bool
}

// NewToastManager creates a toast manager.
func NewToastManager(styles *Styles, width int) *ToastManager {
	return &ToastManager{
		styles: styles,
		width:  width,
	}
}

// Show displays a toast notification.
func (tm *ToastManager) Show(msg string, level ToastLevel, duration time.Duration) {
	tm.current = &Toast{
		Message:   msg,
		Level:     level,
		CreatedAt: time.Now(),
		Duration:  duration,
	}
	tm.visible = true
	tm.dismissed = false
}

// Dismiss hides the current toast.
func (tm *ToastManager) Dismiss() {
	tm.visible = false
	tm.dismissed = true
}

// IsVisible returns whether a toast is currently shown.
func (tm *ToastManager) IsVisible() bool {
	return tm.visible && tm.current != nil
}

// Current returns the current toast (may be nil).
func (tm *ToastManager) Current() *Toast {
	return tm.current
}

// View renders the toast notification.
func (tm *ToastManager) View() string {
	if !tm.IsVisible() {
		return ""
	}

	toast := tm.current

	// Get icon and border color based on level
	var icon string
	var borderColor lipgloss.Color

	theme := GetTheme("acid-green") // Default theme for colors
	if tm.styles != nil {
		// Use theme colors from styles
		switch toast.Level {
		case ToastInfo:
			icon = "ℹ"
			borderColor = lipgloss.Color("#88C0D0") // Blue
		case ToastSuccess:
			icon = "✓"
			borderColor = lipgloss.Color("#00FF41") // Green
		case ToastWarn:
			icon = "⚠"
			borderColor = lipgloss.Color("#FFD700") // Yellow
		case ToastError:
			icon = "✗"
			borderColor = lipgloss.Color("#FF4136") // Red
		default:
			icon = "ℹ"
			borderColor = theme.Primary
		}
	}

	// Calculate content width (accounting for border and padding)
	contentWidth := tm.width - 4 // 2 for borders, 2 for padding

	// Build the message content
	dismissHint := "[x] dismiss"
	messageSpace := contentWidth - len(icon) - 2 - len(dismissHint) - 2 // icon + space + dismiss + spaces

	message := toast.Message
	if len(message) > messageSpace {
		message = message[:messageSpace-3] + "..."
	}

	// Pad message to fill space
	padding := messageSpace - len(message)
	if padding < 0 {
		padding = 0
	}
	paddedMessage := message + strings.Repeat(" ", padding)

	content := icon + " " + paddedMessage + "  " + dismissHint

	// Create the styled box
	boxStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(0, 1).
		Width(tm.width)

	return boxStyle.Render(content)
}

// ToastTickMsg is a tick message for auto-dismiss.
type ToastTickMsg time.Time

// TickCmd returns a command that ticks every 100ms for toast management.
func (tm *ToastManager) TickCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return ToastTickMsg(t)
	})
}

// Update handles toast tick messages.
func (tm *ToastManager) Update(msg tea.Msg) (*ToastManager, tea.Cmd) {
	switch msg.(type) {
	case ToastTickMsg:
		if tm.IsVisible() && tm.current != nil && tm.current.Duration > 0 {
			if time.Since(tm.current.CreatedAt) > tm.current.Duration {
				tm.Dismiss()
				return tm, nil
			}
			// Keep ticking
			return tm, tm.TickCmd()
		}
	}
	return tm, nil
}
