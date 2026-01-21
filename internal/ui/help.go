package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// HelpPanel manages the help modal.
type HelpPanel struct {
	styles  *Styles
	visible bool
	width   int
	height  int
}

// NewHelpPanel creates a help panel.
func NewHelpPanel(styles *Styles, width, height int) *HelpPanel {
	return &HelpPanel{
		styles:  styles,
		visible: false,
		width:   width,
		height:  height,
	}
}

// Show makes the help panel visible.
func (h *HelpPanel) Show() {
	h.visible = true
}

// Hide closes the help panel.
func (h *HelpPanel) Hide() {
	h.visible = false
}

// IsVisible returns whether the panel is shown.
func (h *HelpPanel) IsVisible() bool {
	return h.visible
}

// SetSize updates the panel dimensions.
func (h *HelpPanel) SetSize(width, height int) {
	h.width = width
	h.height = height
}

// Update handles input for the help panel.
func (h *HelpPanel) Update(msg tea.Msg) (*HelpPanel, tea.Cmd) {
	if !h.visible {
		return h, nil
	}

	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return h, nil
	}

	// Close on Esc or ?
	switch keyMsg.String() {
	case "esc", "?":
		h.visible = false
	}

	return h, nil
}

// View renders the help panel.
func (h *HelpPanel) View() string {
	if !h.visible {
		return ""
	}

	// Helper to highlight keybinds in brackets
	kb := func(key string) string {
		bracketStyle := lipgloss.NewStyle().Foreground(h.styles.theme.Muted)
		accentStyle := lipgloss.NewStyle().
			Foreground(h.styles.theme.Primary).
			Bold(true)
		return bracketStyle.Render("[") +
			accentStyle.Render(key) +
			bracketStyle.Render("]")
	}

	// Non-bracketed keybind (for plain keys)
	k := func(key string) string {
		return lipgloss.NewStyle().
			Foreground(h.styles.theme.Primary).
			Bold(true).
			Render(key)
	}

	content := ""

	// Title
	titleStyle := lipgloss.NewStyle().
		Width(76).
		Align(lipgloss.Center).
		Bold(true).
		Foreground(h.styles.theme.Primary)
	content += titleStyle.Render("KEYBINDINGS") + "\n\n"

	// Two-column layout
	leftCol := ""
	rightCol := ""

	// GLOBAL
	leftCol += h.styles.Title.Render("GLOBAL") + "\n"
	leftCol += "  " + k("q") + "       Quit (detach)\n"
	leftCol += "  " + k("Ctrl+X") + "  Shutdown all\n"
	leftCol += "  " + k("S") + "       Settings\n"
	leftCol += "  " + k("E") + "       Edit config\n"
	leftCol += "  " + k("H") + "       Alerts\n"
	leftCol += "  " + k("?") + "       This help\n\n"

	// NAVIGATION
	rightCol += h.styles.Title.Render("NAVIGATION") + "\n"
	rightCol += "  " + kb("↑/k") + "     Up\n"
	rightCol += "  " + kb("↓/j") + "     Down\n"
	rightCol += "  " + kb("Tab") + "     Switch pane\n"
	rightCol += "  " + kb("Enter") + "   Select/Confirm\n"
	rightCol += "  " + kb("Esc") + "     Back/Cancel\n\n\n"

	// SIDEBAR
	leftCol += h.styles.Title.Render("SIDEBAR") + "\n"
	leftCol += "  " + k("s") + "       Start project\n"
	leftCol += "  " + k("x") + "       Stop project\n"
	leftCol += "  " + k("d") + "       Delete project\n"
	leftCol += "  " + k("c") + "       Repair stale\n"
	leftCol += "  " + k("Ctrl+h") + "  Hide/show\n\n"

	// SERVICES
	rightCol += h.styles.Title.Render("SERVICES") + "\n"
	rightCol += "  " + k("s") + "       Start service\n"
	rightCol += "  " + k("x") + "       Stop service\n"
	rightCol += "  " + k("r") + "       Restart service\n"
	rightCol += "  " + kb("Enter") + "   Filter logs\n\n\n"

	// LOGS
	leftCol += h.styles.Title.Render("LOGS") + "\n"
	leftCol += "  " + k("f") + "       Toggle follow\n"
	leftCol += "  " + kb("↑/↓") + "     Scroll\n"
	leftCol += "  " + kb("g/G") + "     Top/Bottom\n"

	// SEARCH
	rightCol += h.styles.Title.Render("SEARCH (in Logs)") + "\n"
	rightCol += "  " + k("/") + "       Start search\n"
	rightCol += "  " + k("n") + "       Next match\n"
	rightCol += "  " + k("N") + "       Prev match\n"
	rightCol += "  " + k("Ctrl+f") + "  Filter mode\n"
	rightCol += "  " + kb("Esc") + "     Clear search\n"

	// Combine columns
	columns := lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftCol,
		rightCol,
	)

	content += columns + "\n\n"

	// Footer
	footerText := kb("Esc") + " or " + kb("?") + " to close"
	footerStyle := lipgloss.NewStyle().
		Width(76).
		Align(lipgloss.Center)
	content += footerStyle.Render(footerText)

	// Fixed size modal box (80 cols x 28 rows)
	modalStyle := h.styles.FocusedBorder.
		Width(80).
		Height(28).
		Padding(1, 2)

	return modalStyle.Render(content)
}
