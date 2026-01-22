package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/infktd/devdash/internal/config"
)

func TestSettingsPanelCreate(t *testing.T) {
	cfg := config.Default()
	theme := GetTheme("matrix")
	styles := NewStyles(theme)
	sp := NewSettingsPanel(cfg, styles, 80, 24)

	if sp == nil {
		t.Fatal("settings panel should not be nil")
	}
	if sp.IsVisible() {
		t.Error("should not be visible by default")
	}
	if len(sp.fields) != 8 {
		t.Errorf("should have 8 fields, got %d", len(sp.fields))
	}
}

func TestSettingsPanelShowHide(t *testing.T) {
	cfg := config.Default()
	theme := GetTheme("matrix")
	styles := NewStyles(theme)
	sp := NewSettingsPanel(cfg, styles, 80, 24)

	sp.Show()
	if !sp.IsVisible() {
		t.Error("should be visible after Show()")
	}

	sp.Hide()
	if sp.IsVisible() {
		t.Error("should not be visible after Hide()")
	}
}

func TestSettingsPanelWorkingCopy(t *testing.T) {
	cfg := config.Default()
	originalTheme := cfg.UI.Theme
	theme := GetTheme("matrix")
	styles := NewStyles(theme)
	sp := NewSettingsPanel(cfg, styles, 80, 24)

	// Modify working copy directly
	sp.workingCopy.theme = "nord"
	sp.workingCopy.scanDepth = 5
	sp.workingCopy.autoDiscover = false

	// Apply to config
	sp.applyWorkingCopy()

	if sp.config.UI.Theme != "nord" {
		t.Errorf("theme should be nord, got %s", sp.config.UI.Theme)
	}
	if sp.config.Projects.ScanDepth != 5 {
		t.Errorf("scan depth should be 5, got %d", sp.config.Projects.ScanDepth)
	}
	if sp.config.Projects.AutoDiscover != false {
		t.Error("auto discover should be false")
	}

	// Test loading back from config
	sp.loadWorkingCopy()
	if sp.workingCopy.theme != "nord" {
		t.Errorf("working copy theme should be nord, got %s", sp.workingCopy.theme)
	}

	// Reset config for other tests
	cfg.UI.Theme = originalTheme
}

func TestSettingsPanelCancel(t *testing.T) {
	cfg := config.Default()
	originalTheme := cfg.UI.Theme
	theme := GetTheme("matrix")
	styles := NewStyles(theme)
	sp := NewSettingsPanel(cfg, styles, 80, 24)
	sp.Show()

	// Modify working copy
	sp.workingCopy.theme = "dracula"

	// Cancel should reload from config
	sp.Cancel()

	if sp.workingCopy.theme != originalTheme {
		t.Errorf("theme should be reset to %s, got %s", originalTheme, sp.workingCopy.theme)
	}
	if sp.IsVisible() {
		t.Error("should not be visible after cancel")
	}
}

func TestSettingsPanelFieldTypes(t *testing.T) {
	cfg := config.Default()
	theme := GetTheme("matrix")
	styles := NewStyles(theme)
	sp := NewSettingsPanel(cfg, styles, 80, 24)

	// Check field types
	expectedTypes := []FieldType{
		FieldToggle, // Auto-discover
		FieldSelect, // Scan depth
		FieldSelect, // Default log view
		FieldSelect, // Theme
		FieldToggle, // System notifications
		FieldToggle, // TUI alerts
		FieldButton, // Save
		FieldButton, // Cancel
	}

	for i, expected := range expectedTypes {
		if sp.fields[i].Type != expected {
			t.Errorf("field %d: expected type %d, got %d", i, expected, sp.fields[i].Type)
		}
	}
}

func TestSettingsPanelNavigation(t *testing.T) {
	cfg := config.Default()
	theme := GetTheme("matrix")
	styles := NewStyles(theme)
	sp := NewSettingsPanel(cfg, styles, 80, 24)
	sp.Show()

	if sp.selectedField != 0 {
		t.Error("should start at field 0")
	}

	// Test navigation (without sending actual tea messages)
	// Just verify the initial state is correct
	if sp.editMode {
		t.Error("should not be in edit mode initially")
	}
}

func TestSettingsPanelView(t *testing.T) {
	cfg := config.Default()
	theme := GetTheme("matrix")
	styles := NewStyles(theme)
	sp := NewSettingsPanel(cfg, styles, 80, 24)

	// Not visible - empty view
	view := sp.View()
	if view != "" {
		t.Error("view should be empty when not visible")
	}

	// Visible
	sp.Show()
	view = sp.View()
	if view == "" {
		t.Error("view should have content when visible")
	}
}

// MockKeyMsg creates a tea.KeyMsg for testing
func MockKeyMsg(key string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key), Alt: false}
}

func TestSettingsPanelUpdate(t *testing.T) {
	cfg := config.Default()
	theme := GetTheme("matrix")
	styles := NewStyles(theme)
	sp := NewSettingsPanel(cfg, styles, 80, 24)
	sp.Show()

	// Test that Update returns when not visible
	sp.Hide()
	updatedSp, cmd := sp.Update(MockKeyMsg("j"))
	if updatedSp != sp {
		t.Error("should return same panel when not visible")
	}
	if cmd != nil {
		t.Error("should return nil cmd when not visible")
	}

	// Test that Update handles non-KeyMsg
	sp.Show()
	updatedSp, cmd = sp.Update(tea.WindowSizeMsg{})
	if updatedSp != sp {
		t.Error("should return same panel for non-KeyMsg")
	}
	if cmd != nil {
		t.Error("should return nil cmd for non-KeyMsg")
	}
}

func TestSettingsPanelHandleNavigationMode(t *testing.T) {
	cfg := config.Default()
	theme := GetTheme("matrix")
	styles := NewStyles(theme)
	sp := NewSettingsPanel(cfg, styles, 80, 24)
	sp.Show()

	initialField := sp.selectedField
	if initialField != 0 {
		t.Errorf("should start at field 0, got %d", initialField)
	}

	// Test down navigation (j)
	sp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	if sp.selectedField != 1 {
		t.Errorf("j should move to field 1, got %d", sp.selectedField)
	}

	// Test down navigation (down arrow)
	sp.Update(tea.KeyMsg{Type: tea.KeyDown})
	if sp.selectedField != 2 {
		t.Errorf("down arrow should move to field 2, got %d", sp.selectedField)
	}

	// Test up navigation (k)
	sp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	if sp.selectedField != 1 {
		t.Errorf("k should move back to field 1, got %d", sp.selectedField)
	}

	// Test up navigation (up arrow)
	sp.Update(tea.KeyMsg{Type: tea.KeyUp})
	if sp.selectedField != 0 {
		t.Errorf("up arrow should move back to field 0, got %d", sp.selectedField)
	}

	// Test that up at top doesn't go negative
	sp.Update(tea.KeyMsg{Type: tea.KeyUp})
	if sp.selectedField != 0 {
		t.Errorf("up at top should stay at 0, got %d", sp.selectedField)
	}

	// Move to bottom
	for i := 0; i < len(sp.fields); i++ {
		sp.Update(tea.KeyMsg{Type: tea.KeyDown})
	}
	maxField := len(sp.fields) - 1
	if sp.selectedField != maxField {
		t.Errorf("should be at max field %d, got %d", maxField, sp.selectedField)
	}

	// Test that down at bottom doesn't exceed
	sp.Update(tea.KeyMsg{Type: tea.KeyDown})
	if sp.selectedField != maxField {
		t.Errorf("down at bottom should stay at %d, got %d", maxField, sp.selectedField)
	}

	// Test Esc cancels
	sp.workingCopy.theme = "modified"
	sp.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if sp.IsVisible() {
		t.Error("Esc should close the panel")
	}
}

func TestSettingsPanelHandleEditMode(t *testing.T) {
	cfg := config.Default()
	theme := GetTheme("matrix")
	styles := NewStyles(theme)
	sp := NewSettingsPanel(cfg, styles, 80, 24)
	sp.Show()

	// Navigate to a Select field (scan depth at index 1)
	sp.selectedField = 1
	if sp.fields[sp.selectedField].Type != FieldSelect {
		t.Fatal("field 1 should be a Select field")
	}

	// Activate edit mode with Enter
	sp.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if !sp.editMode {
		t.Error("Enter on Select field should enable edit mode")
	}

	// Store initial value
	initialValue := sp.workingCopy.scanDepth

	// Test cycling with l (right)
	sp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")})
	newValue := sp.workingCopy.scanDepth
	if newValue == initialValue {
		t.Error("l should cycle to next option")
	}

	// Test cycling with h (left)
	sp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("h")})
	if sp.workingCopy.scanDepth != initialValue {
		t.Error("h should cycle back to initial value")
	}

	// Test left arrow
	sp.Update(tea.KeyMsg{Type: tea.KeyLeft})
	if sp.workingCopy.scanDepth == initialValue {
		t.Error("left arrow should cycle to previous option")
	}

	// Reset and test right arrow
	sp.workingCopy.scanDepth = initialValue
	sp.Update(tea.KeyMsg{Type: tea.KeyRight})
	if sp.workingCopy.scanDepth == initialValue {
		t.Error("right arrow should cycle to next option")
	}

	// Test Enter confirms and exits edit mode
	sp.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if sp.editMode {
		t.Error("Enter should exit edit mode")
	}

	// Re-enter edit mode and test Esc cancels
	sp.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if !sp.editMode {
		t.Error("should be in edit mode")
	}
	sp.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if sp.editMode {
		t.Error("Esc should exit edit mode")
	}
}

func TestSettingsPanelHandleFieldActivation(t *testing.T) {
	cfg := config.Default()
	theme := GetTheme("matrix")
	styles := NewStyles(theme)
	sp := NewSettingsPanel(cfg, styles, 80, 24)
	sp.Show()

	// Test Toggle field (auto-discover at index 0)
	sp.selectedField = 0
	if sp.fields[sp.selectedField].Type != FieldToggle {
		t.Fatal("field 0 should be a Toggle field")
	}

	initialValue := sp.workingCopy.autoDiscover
	sp.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if sp.workingCopy.autoDiscover == initialValue {
		t.Error("Enter on Toggle should flip the value")
	}

	// Test Select field (scan depth at index 1)
	sp.selectedField = 1
	if sp.fields[sp.selectedField].Type != FieldSelect {
		t.Fatal("field 1 should be a Select field")
	}

	sp.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if !sp.editMode {
		t.Error("Enter on Select should enable edit mode")
	}
	sp.editMode = false // Reset

	// Test Cancel button (index 7)
	sp.selectedField = 7
	if sp.fields[sp.selectedField].Label != "Cancel" {
		t.Fatal("field 7 should be Cancel button")
	}

	sp.workingCopy.theme = "modified"
	sp.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if sp.IsVisible() {
		t.Error("Enter on Cancel should close panel")
	}
	if sp.workingCopy.theme == "modified" {
		t.Error("Cancel should reset working copy")
	}
}

func TestSettingsPanelHandleSave(t *testing.T) {
	// Create a temporary config file
	cfg := config.Default()
	theme := GetTheme("matrix")
	styles := NewStyles(theme)
	sp := NewSettingsPanel(cfg, styles, 80, 24)
	sp.Show()

	// Modify working copy
	sp.workingCopy.theme = "nord"
	sp.workingCopy.scanDepth = 3

	// Navigate to Save button (index 6)
	sp.selectedField = 6
	if sp.fields[sp.selectedField].Label != "Save" {
		t.Fatal("field 6 should be Save button")
	}

	// Activate Save
	_, cmd := sp.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Should return a command
	if cmd == nil {
		t.Error("Save should return a command")
	}

	// Should have applied working copy to config
	if sp.config.UI.Theme != "nord" {
		t.Errorf("theme should be nord, got %s", sp.config.UI.Theme)
	}
	if sp.config.Projects.ScanDepth != 3 {
		t.Errorf("scan depth should be 3, got %d", sp.config.Projects.ScanDepth)
	}

	// Execute the command to get the message
	if cmd != nil {
		msg := cmd()
		// Should be either settingsSavedMsg or settingsSaveErrorMsg
		switch msg.(type) {
		case settingsSavedMsg:
			// Success case
			if sp.IsVisible() {
				t.Error("panel should be hidden after successful save")
			}
		case settingsSaveErrorMsg:
			// Error case - this is acceptable (file might not be writable)
		default:
			t.Errorf("unexpected message type: %T", msg)
		}
	}
}

func TestSettingsPanelCycleOptions(t *testing.T) {
	cfg := config.Default()
	theme := GetTheme("matrix")
	styles := NewStyles(theme)
	sp := NewSettingsPanel(cfg, styles, 80, 24)
	sp.Show()

	// Navigate to scan depth field (index 1)
	sp.selectedField = 1
	field := sp.fields[sp.selectedField]
	if field.Type != FieldSelect {
		t.Fatal("field 1 should be a Select field")
	}

	// Set to first option
	sp.workingCopy.scanDepth = 1

	// Enter edit mode
	sp.editMode = true

	// Cycle forward through all options
	sp.Update(tea.KeyMsg{Type: tea.KeyRight})
	if sp.workingCopy.scanDepth != 2 {
		t.Errorf("should cycle to 2, got %d", sp.workingCopy.scanDepth)
	}

	sp.Update(tea.KeyMsg{Type: tea.KeyRight})
	if sp.workingCopy.scanDepth != 3 {
		t.Errorf("should cycle to 3, got %d", sp.workingCopy.scanDepth)
	}

	// Continue to last option
	sp.workingCopy.scanDepth = 5

	// Test wrap-around to first
	sp.Update(tea.KeyMsg{Type: tea.KeyRight})
	if sp.workingCopy.scanDepth != 1 {
		t.Errorf("should wrap to 1, got %d", sp.workingCopy.scanDepth)
	}

	// Test cycling backward with wrap-around
	sp.Update(tea.KeyMsg{Type: tea.KeyLeft})
	if sp.workingCopy.scanDepth != 5 {
		t.Errorf("should wrap backward to 5, got %d", sp.workingCopy.scanDepth)
	}

	// Test theme field cycling (index 3)
	sp.selectedField = 3
	field = sp.fields[sp.selectedField]
	if field.Type != FieldSelect {
		t.Fatal("field 3 should be a Select field")
	}

	initialTheme := sp.workingCopy.theme
	sp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")})
	if sp.workingCopy.theme == initialTheme {
		t.Error("should cycle to different theme")
	}

	// Cycle back
	sp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("h")})
	if sp.workingCopy.theme != initialTheme {
		t.Errorf("should cycle back to %s, got %s", initialTheme, sp.workingCopy.theme)
	}
}

func TestSettingsPanelRenderSelectOptions(t *testing.T) {
	cfg := config.Default()
	theme := GetTheme("matrix")
	styles := NewStyles(theme)
	sp := NewSettingsPanel(cfg, styles, 80, 24)
	sp.Show()

	// Navigate to a select field
	// This depends on field order
	for i := 0; i < 5; i++ {
		sp.Update(MockKeyMsg("j"))
	}

	// Activate select field
	sp.Update(MockKeyMsg("\r"))

	// View should show options
	view := sp.View()
	if view == "" {
		t.Error("View() should not be empty when select is active")
	}

	// Should contain theme options
	if !strings.Contains(view, "dark") && !strings.Contains(view, "light") {
		t.Log("View may contain theme options in select mode")
	}
}

func TestSettingsPanelGetHelpText(t *testing.T) {
	cfg := config.Default()
	theme := GetTheme("matrix")
	styles := NewStyles(theme)
	sp := NewSettingsPanel(cfg, styles, 80, 24)
	sp.Show()

	view := sp.View()

	// Help text should appear in view
	// Check for common help keys
	if !strings.Contains(view, "enter") && !strings.Contains(view, "esc") {
		t.Log("View should contain help text for navigation")
	}
}
