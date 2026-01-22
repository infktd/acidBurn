package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewHelpPanel(t *testing.T) {
	styles := NewStyles(GetTheme("matrix"))
	panel := NewHelpPanel(styles, 100, 50)
	if panel == nil {
		t.Fatal("NewHelpPanel returned nil")
	}
	if panel.IsVisible() {
		t.Error("new help panel should not be visible")
	}
}

func TestHelpPanelShowHide(t *testing.T) {
	styles := NewStyles(GetTheme("matrix"))
	panel := NewHelpPanel(styles, 100, 50)

	panel.Show()
	if !panel.IsVisible() {
		t.Error("IsVisible() should return true after Show()")
	}

	panel.Hide()
	if panel.IsVisible() {
		t.Error("IsVisible() should return false after Hide()")
	}
}

func TestHelpPanelUpdate(t *testing.T) {
	styles := NewStyles(GetTheme("matrix"))
	panel := NewHelpPanel(styles, 100, 50)
	panel.Show()

	// Test key handling - should close on '?'
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}
	_, cmd := panel.Update(msg)

	// Should handle close on ? or escape
	_ = cmd

	// Panel should be hidden after pressing ?
	if panel.IsVisible() {
		t.Error("panel should be hidden after pressing ?")
	}
}

func TestHelpPanelUpdateEscape(t *testing.T) {
	styles := NewStyles(GetTheme("matrix"))
	panel := NewHelpPanel(styles, 100, 50)
	panel.Show()

	// Test escape key
	msg := tea.KeyMsg{Type: tea.KeyEsc}
	panel.Update(msg)

	if panel.IsVisible() {
		t.Error("panel should be hidden after pressing Esc")
	}
}

func TestHelpPanelView(t *testing.T) {
	styles := NewStyles(GetTheme("matrix"))
	panel := NewHelpPanel(styles, 100, 50)

	view := panel.View()
	if view != "" {
		t.Error("View() should return empty when hidden")
	}

	panel.Show()
	view = panel.View()
	if view == "" {
		t.Error("View() should return content when visible")
	}
}
