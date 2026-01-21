package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/infktd/acidburn/internal/config"
	"github.com/infktd/acidburn/internal/registry"
	"github.com/infktd/acidburn/internal/scanner"
	"github.com/infktd/acidburn/internal/ui"
)

func main() {
	// Load config
	cfg, err := config.Load(config.Path())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Load registry
	reg, err := registry.Load(registry.Path())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading registry: %v\n", err)
		os.Exit(1)
	}

	// Auto-discover projects if enabled
	if cfg.Projects.AutoDiscover {
		projects, _ := scanner.Scan(cfg.Projects.ScanPaths, cfg.Projects.ScanDepth)
		for _, path := range projects {
			reg.AddProject(path)
		}
		// Save updated registry
		_ = registry.Save(registry.Path(), reg)
	}

	// Run TUI
	p := tea.NewProgram(ui.New(cfg, reg), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
