package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/infktd/devdash/internal/config"
	"github.com/infktd/devdash/internal/registry"
	"github.com/infktd/devdash/internal/scanner"
	"github.com/infktd/devdash/internal/ui"
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
		projects, err := scanner.Scan(cfg.Projects.ScanPaths, cfg.Projects.ScanDepth)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: error during project scan: %v\n", err)
		}
		for _, path := range projects {
			reg.AddProject(path)
		}
		// Save updated registry
		if err := registry.Save(registry.Path(), reg); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to save registry: %v\n", err)
		}
	}

	// Run TUI with mouse support
	p := tea.NewProgram(
		ui.New(cfg, reg),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(), // Enable mouse motion tracking
	)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
