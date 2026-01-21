package ui

import (
	"testing"

	"github.com/infktd/acidburn/internal/config"
)

func TestSettingsPanelCreate(t *testing.T) {
	cfg := config.Default()
	sp := NewSettingsPanel(cfg)

	if sp == nil {
		t.Fatal("settings panel should not be nil")
	}
	if sp.IsVisible() {
		t.Error("should not be visible by default")
	}
}

func TestSettingsPanelShowHide(t *testing.T) {
	cfg := config.Default()
	sp := NewSettingsPanel(cfg)

	sp.Show()
	if !sp.IsVisible() {
		t.Error("should be visible after Show()")
	}

	sp.Hide()
	if sp.IsVisible() {
		t.Error("should not be visible after Hide()")
	}
}

func TestSettingsPanelSave(t *testing.T) {
	cfg := config.Default()
	sp := NewSettingsPanel(cfg)

	// Modify form values directly
	sp.theme = "nord"
	sp.scanDepth = 5
	sp.autoDiscover = false

	sp.Save()

	if !sp.WasSaved() {
		t.Error("should be marked as saved")
	}
	if sp.config.UI.Theme != "nord" {
		t.Errorf("theme should be nord, got %s", sp.config.UI.Theme)
	}
	if sp.config.Projects.ScanDepth != 5 {
		t.Errorf("scan depth should be 5, got %d", sp.config.Projects.ScanDepth)
	}
	if sp.config.Projects.AutoDiscover != false {
		t.Error("auto discover should be false")
	}
}

func TestSettingsPanelCancel(t *testing.T) {
	cfg := config.Default()
	originalTheme := cfg.UI.Theme

	sp := NewSettingsPanel(cfg)
	sp.Show()

	// Modify
	sp.theme = "dracula"

	// Cancel
	sp.Cancel()

	if sp.WasSaved() {
		t.Error("should not be saved after cancel")
	}
	if sp.theme != originalTheme {
		t.Errorf("theme should be reset to %s, got %s", originalTheme, sp.theme)
	}
}

func TestSettingsPanelForm(t *testing.T) {
	cfg := config.Default()
	sp := NewSettingsPanel(cfg)

	form := sp.Form()
	if form == nil {
		t.Error("form should not be nil")
	}
}

func TestSettingsPanelView(t *testing.T) {
	cfg := config.Default()
	sp := NewSettingsPanel(cfg)

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
