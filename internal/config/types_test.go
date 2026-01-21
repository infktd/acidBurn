package config

import (
	"testing"
)

func TestDefaultConfigHasScanPaths(t *testing.T) {
	cfg := Default()
	if len(cfg.Projects.ScanPaths) == 0 {
		t.Fatal("Default config should have at least one scan path")
	}
}

func TestDefaultConfigHasTheme(t *testing.T) {
	cfg := Default()
	if cfg.UI.Theme == "" {
		t.Fatal("Default config should have a theme")
	}
}

func TestDefaultConfigHasPollingIntervals(t *testing.T) {
	cfg := Default()
	if cfg.Polling.FocusedProject == 0 {
		t.Fatal("Default config should have focused project polling interval")
	}
	if cfg.Polling.BackgroundProject == 0 {
		t.Fatal("Default config should have background project polling interval")
	}
}
