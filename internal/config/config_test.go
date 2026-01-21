package config

import (
	"path/filepath"
	"testing"
)

func TestLoadCreatesDefaultIfMissing(t *testing.T) {
	// Use temp dir
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg == nil {
		t.Fatal("Load() returned nil config")
	}
	if cfg.UI.Theme != "acid-green" {
		t.Errorf("Expected theme 'acid-green', got %q", cfg.UI.Theme)
	}
}

func TestSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Create and modify config
	cfg := Default()
	cfg.UI.Theme = "nord"

	// Save it
	if err := Save(configPath, cfg); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	// Load it back
	loaded, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if loaded.UI.Theme != "nord" {
		t.Errorf("Expected theme 'nord', got %q", loaded.UI.Theme)
	}
}

func TestConfigPath(t *testing.T) {
	path := Path()
	if path == "" {
		t.Fatal("Path() returned empty string")
	}
	if !filepath.IsAbs(path) {
		t.Errorf("Path() should return absolute path, got %q", path)
	}
}
