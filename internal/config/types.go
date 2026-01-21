package config

import (
	"os"
	"path/filepath"
)

// Config represents the acidBurn configuration.
type Config struct {
	Projects      ProjectsConfig      `yaml:"projects"`
	Notifications NotificationsConfig `yaml:"notifications"`
	UI            UIConfig            `yaml:"ui"`
	Polling       PollingConfig       `yaml:"polling"`
}

// ProjectsConfig configures project discovery.
type ProjectsConfig struct {
	ScanPaths    []string `yaml:"scan_paths"`
	AutoDiscover bool     `yaml:"auto_discover"`
	ScanDepth    int      `yaml:"scan_depth"`
}

// NotificationsConfig configures alerts and notifications.
type NotificationsConfig struct {
	SystemEnabled bool                   `yaml:"system_enabled"`
	TUIAlerts     bool                   `yaml:"tui_alerts"`
	CriticalOnly  bool                   `yaml:"critical_only"`
	Overrides     []NotificationOverride `yaml:"overrides,omitempty"`
}

// NotificationOverride allows per-service notification settings.
type NotificationOverride struct {
	Service      string `yaml:"service"`
	System       bool   `yaml:"system"`
	CriticalOnly bool   `yaml:"critical_only,omitempty"`
}

// UIConfig configures the user interface.
type UIConfig struct {
	Theme          string `yaml:"theme"`
	DefaultLogView string `yaml:"default_log_view"`
	LogFollow      bool   `yaml:"log_follow"`
	ShowTimestamps bool   `yaml:"show_timestamps"`
	DimTimestamps  bool   `yaml:"dim_timestamps"`
	SidebarWidth   int    `yaml:"sidebar_width"`
}

// PollingConfig configures polling intervals in seconds.
type PollingConfig struct {
	FocusedProject    int `yaml:"focused_project"`
	BackgroundProject int `yaml:"background_project"`
}

// Default returns a Config with sensible defaults.
func Default() *Config {
	home, _ := os.UserHomeDir()
	return &Config{
		Projects: ProjectsConfig{
			ScanPaths: []string{
				filepath.Join(home, "code"),
				filepath.Join(home, "projects"),
			},
			AutoDiscover: true,
			ScanDepth:    3,
		},
		Notifications: NotificationsConfig{
			SystemEnabled: true,
			TUIAlerts:     true,
			CriticalOnly:  false,
		},
		UI: UIConfig{
			Theme:          "acid-green",
			DefaultLogView: "focused",
			LogFollow:      true,
			ShowTimestamps: true,
			DimTimestamps:  true,
			SidebarWidth:   25,
		},
		Polling: PollingConfig{
			FocusedProject:    2,
			BackgroundProject: 10,
		},
	}
}
