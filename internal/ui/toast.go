package ui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lucasb-eyer/go-colorful"
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
	opacity   float64 // 0.0 to 1.0 for fade-in animation
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
	tm.opacity = 0.0 // Start invisible for fade-in
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

	// Get icon and color based on level
	var icon string
	var color lipgloss.Color

	if tm.styles != nil {
		switch toast.Level {
		case ToastInfo:
			icon = "ℹ"
			color = lipgloss.Color("#88C0D0") // Blue
		case ToastSuccess:
			icon = "✓"
			color = lipgloss.Color("#00FF41") // Green
		case ToastWarn:
			icon = "⚠"
			color = lipgloss.Color("#FFD700") // Yellow
		case ToastError:
			icon = "✗"
			color = lipgloss.Color("#FF4136") // Red
		default:
			icon = "ℹ"
			color = tm.styles.theme.Primary
		}
	}

	// Dynamic width based on terminal, leaving room for other UI elements
	// Use 70% of terminal width, with min 60 and max 120
	maxWidth := tm.width * 70 / 100
	if maxWidth < 60 {
		maxWidth = 60
	}
	if maxWidth > 120 {
		maxWidth = 120
	}
	messageSpace := maxWidth - len(icon) - 2 // icon + space

	message := toast.Message
	if len(message) > messageSpace {
		message = message[:messageSpace-3] + "..."
	}

	content := icon + " " + message

	// Apply fade-in effect by blending color with background
	finalColor := color
	if tm.opacity < 1.0 {
		// Blend from dark gray to target color
		bgColor, _ := colorful.Hex("#1a1a1a")
		targetColor, _ := colorful.Hex(string(color))
		blended := bgColor.BlendLuv(targetColor, tm.opacity)
		finalColor = lipgloss.Color(blended.Hex())
	}

	// Simple inline style with color
	style := lipgloss.NewStyle().
		Foreground(finalColor).
		Bold(true)

	return style.Render(content)
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
		if tm.IsVisible() && tm.current != nil {
			elapsed := time.Since(tm.current.CreatedAt)

			fadeInDuration := 1000 * time.Millisecond   // 1 second fade-in
			visibleDuration := 3000 * time.Millisecond  // 3 seconds fully visible
			fadeOutDuration := 1000 * time.Millisecond  // 1 second fade-out
			totalDuration := fadeInDuration + visibleDuration + fadeOutDuration

			if elapsed < fadeInDuration {
				// Fade-in phase: 0 -> 1
				tm.opacity = float64(elapsed) / float64(fadeInDuration)
			} else if elapsed < fadeInDuration+visibleDuration {
				// Fully visible phase
				tm.opacity = 1.0
			} else if elapsed < totalDuration {
				// Fade-out phase: 1 -> 0
				fadeOutElapsed := elapsed - (fadeInDuration + visibleDuration)
				tm.opacity = 1.0 - (float64(fadeOutElapsed) / float64(fadeOutDuration))
			} else {
				// Animation complete, dismiss
				tm.Dismiss()
				return tm, nil
			}

			// Keep ticking
			return tm, tm.TickCmd()
		}
	}
	return tm, nil
}
