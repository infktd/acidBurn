package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// AlertsPanel manages the alerts modal.
type AlertsPanel struct {
	styles  *Styles
	alerts  *AlertHistory
	visible bool
	width   int
	height  int
}

// NewAlertsPanel creates an alerts panel.
func NewAlertsPanel(styles *Styles, alerts *AlertHistory, width, height int) *AlertsPanel {
	return &AlertsPanel{
		styles:  styles,
		alerts:  alerts,
		visible: false,
		width:   width,
		height:  height,
	}
}

// Show makes the alerts panel visible.
func (a *AlertsPanel) Show() {
	a.visible = true
}

// Hide closes the alerts panel.
func (a *AlertsPanel) Hide() {
	a.visible = false
}

// IsVisible returns whether the panel is shown.
func (a *AlertsPanel) IsVisible() bool {
	return a.visible
}

// SetSize updates the panel dimensions.
func (a *AlertsPanel) SetSize(width, height int) {
	a.width = width
	a.height = height
}

// Update handles input for the alerts panel.
func (a *AlertsPanel) Update(msg tea.Msg) (*AlertsPanel, tea.Cmd) {
	if !a.visible {
		return a, nil
	}

	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return a, nil
	}

	// Close on Esc or H
	switch keyMsg.String() {
	case "esc", "H":
		a.visible = false
	}

	return a, nil
}

// View renders the alerts panel.
func (a *AlertsPanel) View() string {
	if !a.visible {
		return ""
	}

	content := ""

	// Title
	titleStyle := lipgloss.NewStyle().
		Width(76).
		Align(lipgloss.Center).
		Bold(true).
		Foreground(a.styles.theme.Primary)
	content += titleStyle.Render("ALERTS") + "\n\n"

	// Get recent alerts (limit to fit modal height)
	alerts := a.alerts.Recent(18) // Reduced from 20 to fit better
	if len(alerts) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Width(76).
			Align(lipgloss.Center).
			Foreground(a.styles.theme.Muted)
		content += emptyStyle.Render("No alerts yet") + "\n"
	} else {
		// Render alerts (newest first)
		for _, alert := range alerts {
			content += a.renderAlert(alert) + "\n"
		}
	}

	content += "\n"

	// Footer
	footerText := "[Esc] or [H] to close"
	footerStyle := lipgloss.NewStyle().
		Width(76).
		Align(lipgloss.Center)
	content += footerStyle.Render(footerText)

	// Fixed size modal box (80 cols x 28 rows)
	modalStyle := a.styles.ModalBorder.
		Width(80).
		Height(28).
		Padding(1, 2)

	return modalStyle.Render(content)
}

// renderAlert renders a single alert line.
func (a *AlertsPanel) renderAlert(alert Alert) string {
	// Timestamp
	timestamp := alert.Timestamp.Format("15:04:05")
	timeStyle := lipgloss.NewStyle().Foreground(a.styles.theme.Muted)

	// Alert type badge
	var badge string
	var badgeStyle lipgloss.Style

	switch alert.Type {
	case AlertServiceCrashed:
		badgeStyle = lipgloss.NewStyle().
			Foreground(a.styles.theme.Error).
			Bold(true)
		badge = "[CRASH]"
	case AlertServiceRecovered:
		badgeStyle = lipgloss.NewStyle().
			Foreground(a.styles.theme.Success).
			Bold(true)
		badge = "[RECOVER]"
	case AlertProjectStarted:
		badgeStyle = lipgloss.NewStyle().
			Foreground(a.styles.theme.Success).
			Bold(true)
		badge = "[START]"
	case AlertProjectStopped:
		badgeStyle = lipgloss.NewStyle().
			Foreground(a.styles.theme.Warning).
			Bold(true)
		badge = "[STOP]"
	case AlertCritical:
		badgeStyle = lipgloss.NewStyle().
			Foreground(a.styles.theme.Error).
			Bold(true)
		badge = "[CRITICAL]"
	case AlertInfo:
		badgeStyle = lipgloss.NewStyle().
			Foreground(a.styles.theme.Primary).
			Bold(true)
		badge = "[INFO]"
	default:
		badgeStyle = lipgloss.NewStyle().
			Foreground(a.styles.theme.Muted).
			Bold(true)
		badge = "[UNKNOWN]"
	}

	// Message (truncate if too long)
	message := alert.Message
	if alert.Service != "" {
		message = fmt.Sprintf("%s: %s", alert.Service, message)
	}

	// Calculate max message length: 76 - timestamp(8) - badge(~10) - spacing(4) = ~54
	maxLen := 54
	if len(message) > maxLen {
		message = message[:maxLen-3] + "..."
	}

	// Build line: [timestamp] [badge] message
	line := timeStyle.Render(timestamp) + " " +
		badgeStyle.Render(badge) + " " +
		message

	lineStyle := lipgloss.NewStyle().Width(76)
	return lineStyle.Render(line)
}
