package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewAlertsPanel(t *testing.T) {
	styles := NewStyles(GetTheme("matrix"))
	alerts := NewAlertHistory(100)
	panel := NewAlertsPanel(styles, alerts, 100, 50)
	if panel == nil {
		t.Fatal("NewAlertsPanel returned nil")
	}
	if panel.IsVisible() {
		t.Error("new alerts panel should not be visible")
	}
}

func TestAlertsPanelShowHide(t *testing.T) {
	styles := NewStyles(GetTheme("matrix"))
	alerts := NewAlertHistory(100)
	panel := NewAlertsPanel(styles, alerts, 100, 50)

	panel.Show()
	if !panel.IsVisible() {
		t.Error("IsVisible() should return true after Show()")
	}

	panel.Hide()
	if panel.IsVisible() {
		t.Error("IsVisible() should return false after Hide()")
	}
}

func TestAlertsPanelUpdate(t *testing.T) {
	styles := NewStyles(GetTheme("matrix"))
	alerts := NewAlertHistory(100)
	panel := NewAlertsPanel(styles, alerts, 100, 50)
	panel.Show()

	// Test key press handling - should close on 'H'
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'H'}}
	_, cmd := panel.Update(msg)

	// Should close on 'H' or escape
	if cmd != nil {
		// Command returned, likely to close
	}

	// Panel should be hidden after pressing H
	if panel.IsVisible() {
		t.Error("panel should be hidden after pressing H")
	}
}

func TestAlertsPanelUpdateEscape(t *testing.T) {
	styles := NewStyles(GetTheme("matrix"))
	alerts := NewAlertHistory(100)
	panel := NewAlertsPanel(styles, alerts, 100, 50)
	panel.Show()

	// Test escape key
	msg := tea.KeyMsg{Type: tea.KeyEsc}
	panel.Update(msg)

	if panel.IsVisible() {
		t.Error("panel should be hidden after pressing Esc")
	}
}

func TestAlertsPanelView(t *testing.T) {
	styles := NewStyles(GetTheme("matrix"))
	alerts := NewAlertHistory(100)
	panel := NewAlertsPanel(styles, alerts, 100, 50)

	// Should return empty when hidden
	view := panel.View()
	if view != "" {
		t.Error("View() should return empty string when hidden")
	}

	panel.Show()
	view = panel.View()
	if view == "" {
		t.Error("View() should return content when visible")
	}
}
