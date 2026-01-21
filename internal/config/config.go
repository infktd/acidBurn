// Package config handles devdash configuration loading and persistence.
package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	configDir  = "devdash"
	configFile = "config.yaml"
)

// Path returns the default config file path.
func Path() string {
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		home, _ := os.UserHomeDir()
		configHome = filepath.Join(home, ".config")
	}
	return filepath.Join(configHome, configDir, configFile)
}

// Load reads config from path, creating default if missing.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		cfg := Default()
		if err := Save(path, cfg); err != nil {
			return cfg, nil // Return default even if save fails
		}
		return cfg, nil
	}
	if err != nil {
		return nil, err
	}

	cfg := Default() // Start with defaults
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

// Save writes config to path, creating directories as needed.
func Save(path string, cfg *Config) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
