package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ConfirmDialog is a small confirmation dialog.
type ConfirmDialog struct {
	visible  bool
	message  string
	onYes    func() tea.Msg
	onNo     func() tea.Msg
	selected int // 0 = Yes, 1 = No
	styles   *Styles
}

// NewConfirmDialog creates a new confirmation dialog.
func NewConfirmDialog(styles *Styles) *ConfirmDialog {
	return &ConfirmDialog{
		styles:   styles,
		selected: 1, // Default to "No" for safety
	}
}

// Show displays the confirmation dialog.
func (c *ConfirmDialog) Show(message string, onYes, onNo func() tea.Msg) {
	c.visible = true
	c.message = message
	c.onYes = onYes
	c.onNo = onNo
	c.selected = 1 // Always default to No
}

// Hide closes the dialog.
func (c *ConfirmDialog) Hide() {
	c.visible = false
}

// IsVisible returns whether the dialog is shown.
func (c *ConfirmDialog) IsVisible() bool {
	return c.visible
}

// Update handles input for the dialog.
func (c *ConfirmDialog) Update(msg tea.Msg) (*ConfirmDialog, tea.Cmd) {
	if !c.visible {
		return c, nil
	}

	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return c, nil
	}

	switch keyMsg.String() {
	case "left", "h", "right", "l", "tab":
		// Toggle between Yes and No
		c.selected = 1 - c.selected

	case "enter":
		c.visible = false
		if c.selected == 0 && c.onYes != nil {
			return c, func() tea.Msg { return c.onYes() }
		} else if c.onNo != nil {
			return c, func() tea.Msg { return c.onNo() }
		}

	case "esc":
		c.visible = false
		if c.onNo != nil {
			return c, func() tea.Msg { return c.onNo() }
		}

	case "y":
		// Quick yes
		c.visible = false
		if c.onYes != nil {
			return c, func() tea.Msg { return c.onYes() }
		}

	case "n":
		// Quick no
		c.visible = false
		if c.onNo != nil {
			return c, func() tea.Msg { return c.onNo() }
		}
	}

	return c, nil
}

// View renders the confirmation dialog as a small box.
func (c *ConfirmDialog) View() string {
	if !c.visible {
		return ""
	}

	var content string

	// Message
	msgStyle := lipgloss.NewStyle().
		Width(46).
		Align(lipgloss.Center).
		Bold(true)
	content += msgStyle.Render(c.message) + "\n\n"

	// Buttons
	yesButton := "[ Yes ]"
	noButton := "[ No ]"

	if c.selected == 0 {
		yesButton = c.styles.SelectedItem.Render(yesButton)
	} else {
		noButton = c.styles.SelectedItem.Render(noButton)
	}

	buttonLine := lipgloss.JoinHorizontal(
		lipgloss.Center,
		yesButton,
		"    ",
		noButton,
	)

	buttonStyle := lipgloss.NewStyle().Width(46).Align(lipgloss.Center)
	content += buttonStyle.Render(buttonLine) + "\n"

	// Help text
	helpStyle := lipgloss.NewStyle().
		Width(46).
		Align(lipgloss.Center).
		Foreground(c.styles.theme.Muted)
	content += "\n" + helpStyle.Render("←/→ or Tab to switch  [Enter] Confirm  [Esc] Cancel")

	// Wrap in a box (50 wide x 10 tall)
	boxStyle := c.styles.FocusedBorder.
		Width(50).
		Height(10).
		Padding(1, 2)

	return boxStyle.Render(content)
}
