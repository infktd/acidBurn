package ui

import (
	"testing"

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
