package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewConfirmDialog(t *testing.T) {
	styles := NewStyles(GetTheme("matrix"))
	dialog := NewConfirmDialog(styles)
	if dialog == nil {
		t.Fatal("NewConfirmDialog returned nil")
	}
	if dialog.IsVisible() {
		t.Error("new confirm dialog should not be visible")
	}
}

func TestConfirmDialogShowHide(t *testing.T) {
	styles := NewStyles(GetTheme("matrix"))
	dialog := NewConfirmDialog(styles)

	dialog.Show("Test message", func() tea.Msg { return nil }, func() tea.Msg { return nil })
	if !dialog.IsVisible() {
		t.Error("IsVisible() should return true after Show()")
	}

	dialog.Hide()
	if dialog.IsVisible() {
		t.Error("IsVisible() should return false after Hide()")
	}
}

func TestConfirmDialogCallbackYes(t *testing.T) {
	styles := NewStyles(GetTheme("matrix"))
	dialog := NewConfirmDialog(styles)

	dialog.Show("Test", func() tea.Msg {
		return nil
	}, func() tea.Msg { return nil })

	// Simulate 'y' key press
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}}
	dialog.Update(msg)

	// Dialog should be hidden
	if dialog.IsVisible() {
		t.Error("dialog should be hidden after 'y' key")
	}
}

func TestConfirmDialogCallbackNo(t *testing.T) {
	styles := NewStyles(GetTheme("matrix"))
	dialog := NewConfirmDialog(styles)

	dialog.Show("Test", func() tea.Msg { return nil }, func() tea.Msg {
		return nil
	})

	// Simulate 'n' key press
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
	dialog.Update(msg)

	// Dialog should be hidden
	if dialog.IsVisible() {
		t.Error("dialog should be hidden after 'n' key")
	}
}

func TestConfirmDialogCallbackEnter(t *testing.T) {
	styles := NewStyles(GetTheme("matrix"))
	dialog := NewConfirmDialog(styles)

	dialog.Show("Test", func() tea.Msg { return nil }, func() tea.Msg { return nil })

	// Default selection is 1 (No), pressing enter should trigger onNo
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	dialog.Update(msg)

	// Dialog should be hidden
	if dialog.IsVisible() {
		t.Error("dialog should be hidden after Enter key")
	}
}

func TestConfirmDialogToggleSelection(t *testing.T) {
	styles := NewStyles(GetTheme("matrix"))
	dialog := NewConfirmDialog(styles)

	dialog.Show("Test", func() tea.Msg { return nil }, func() tea.Msg { return nil })

	// Toggle selection with left/right
	leftMsg := tea.KeyMsg{Type: tea.KeyLeft}
	dialog.Update(leftMsg)

	// Press enter should now trigger onYes (selection toggled to 0)
	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	dialog.Update(enterMsg)

	if dialog.IsVisible() {
		t.Error("dialog should be hidden after Enter")
	}
}

func TestConfirmDialogView(t *testing.T) {
	styles := NewStyles(GetTheme("matrix"))
	dialog := NewConfirmDialog(styles)

	view := dialog.View()
	if view != "" {
		t.Error("View() should return empty when hidden")
	}

	dialog.Show("Test message?", func() tea.Msg { return nil }, func() tea.Msg { return nil })
	view = dialog.View()
	if view == "" {
		t.Error("View() should return content when visible")
	}
	if !strings.Contains(view, "Test message?") {
		t.Error("View() should contain the confirmation message")
	}
}
