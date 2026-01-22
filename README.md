# devdash

[![Made with Nix](https://img.shields.io/badge/Made_with-Nix-5277C3?logo=nixos&logoColor=white)](https://nixos.org)
[![Made with devenv](https://img.shields.io/badge/Made_with-devenv-00D9FF?logo=data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMTAwIiBoZWlnaHQ9IjEwMCIgdmlld0JveD0iMCAwIDEwMCAxMDAiIGZpbGw9Im5vbmUiIHhtbG5zPSJodHRwOi8vd3d3LnczLm9yZy8yMDAwL3N2ZyI+PHBhdGggZD0iTTUwIDEwTDkwIDUwTDUwIDkwTDEwIDUwTDUwIDEwWiIgZmlsbD0id2hpdGUiLz48L3N2Zz4=&logoColor=white)](https://devenv.sh)
[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![Test Coverage](https://img.shields.io/badge/coverage-59.8%25-brightgreen?logo=go&logoColor=white)](https://github.com/infktd/devdash/actions)
[![GitHub Stars](https://img.shields.io/github/stars/infktd/devdash?style=social)](https://github.com/infktd/devdash)

A terminal user interface (TUI) for managing devenv.sh environments. Think of it as "Docker Desktop for Nix" - a unified control plane for all your devenv projects.

## Overview

devdash provides a centralized dashboard for monitoring and controlling multiple devenv projects. It communicates with process-compose daemons via Unix sockets to display service status, stream logs, and manage service lifecycles.

### Key Features

- **Project Discovery**: Automatically scans configured directories for devenv projects
- **Multi-Project Management**: View and switch between all your devenv projects in one place
- **Service Control**: Start, stop, and restart individual services or entire projects
- **Log Streaming**: Real-time log viewing with search, filtering, and level-based colorization
- **Health Monitoring**: Track service crashes, recoveries, and state changes
- **System Notifications**: Desktop notifications for critical events (crashes, recoveries)
- **Theming**: Multiple color themes (matrix, nord, dracula)
- **Customizable UI**: Configurable sidebar width, timestamps, and log behavior

## Requirements

- Go 1.21 or later
- devenv.sh installed and configured
- process-compose (included with devenv)

## Installation

```bash
go install github.com/infktd/devdash@latest
```

Or build from source:

```bash
git clone https://github.com/infktd/devdash
cd devdash
go build -o devdash .
```

## Usage

```bash
devdash
```

On first run, devdash will:
1. Create a default configuration at `~/.config/devdash/config.yaml`
2. Scan configured directories for devenv projects
3. Create a project registry at `~/.config/devdash/registry.yaml`

## Keybindings

### Global

| Key | Action |
|-----|--------|
| `q` | Quit (detach from projects) |
| `Ctrl+X` | Shutdown all services and quit |
| `S` | Open settings |
| `H` | View alert history |
| `?` | Show help |
| `Tab` | Cycle focus between panes |

### Sidebar (Projects)

| Key | Action |
|-----|--------|
| `Up/Down` | Navigate projects |
| `/` | Search projects |
| `Enter` | Select project |
| `s` | Start project |
| `x` | Stop project |
| `Esc` | Clear search |

### Services Pane

| Key | Action |
|-----|--------|
| `Up/Down` | Navigate services |
| `Enter` | Filter logs to service |
| `s` | Start service |
| `x` | Stop service |
| `r` | Restart service |
| `Esc` | Return to sidebar |

### Logs Pane

| Key | Action |
|-----|--------|
| `Up/Down` | Scroll logs |
| `f` | Toggle follow mode |
| `g` | Jump to top |
| `G` | Jump to bottom |
| `/` | Start search |
| `n` | Next search match |
| `N` | Previous search match |
| `Ctrl+f` | Toggle filter mode (show only matches) |
| `Esc` | Clear search/filter |

## Configuration

Configuration is stored at `~/.config/devdash/config.yaml`:

```yaml
projects:
  scan_paths:
    - ~/code
    - ~/projects
  auto_discover: true
  scan_depth: 3

notifications:
  system_enabled: true
  tui_alerts: true
  critical_only: false
  overrides:
    - service: postgres
      system: true
      critical_only: false

ui:
  theme: matrix        # matrix, nord, dracula
  default_log_view: focused
  log_follow: true
  show_timestamps: true
  dim_timestamps: true
  sidebar_width: 25

polling:
  focused_project: 2       # seconds
  background_project: 10   # seconds
```

## Project States

devdash detects and displays project states:

| State | Description |
|-------|-------------|
| Running | process-compose socket is active and responding |
| Degraded | Some services have crashed (exit code != 0) |
| Idle | No process-compose daemon running |
| Stale | Socket file exists but daemon is not responding |
| Missing | Project directory no longer exists |

## Architecture

```
devdash
├── main.go                 # Entry point
└── internal/
    ├── compose/            # process-compose API client
    │   ├── client.go       # Unix socket HTTP client
    │   └── types.go        # API response types
    ├── config/             # Configuration management
    │   ├── config.go       # Load/save YAML config
    │   └── types.go        # Config struct definitions
    ├── health/             # Service health monitoring
    │   └── monitor.go      # Crash detection, recovery tracking
    ├── notify/             # Desktop notifications
    │   └── notify.go       # Cross-platform notifications
    ├── registry/           # Project registry
    │   ├── registry.go     # Load/save project list
    │   └── types.go        # Project struct, state detection
    ├── scanner/            # Project discovery
    │   └── scanner.go      # Filesystem scanning for devenv.nix
    └── ui/                 # Terminal UI (Bubble Tea)
        ├── model.go        # Main application model
        ├── keys.go         # Keybinding definitions
        ├── styles.go       # Lipgloss style definitions
        ├── theme.go        # Color themes
        ├── logview.go      # Log viewport with search
        ├── logbuffer.go    # Circular log buffer, level detection
        ├── settings.go     # Settings panel (huh forms)
        ├── splash.go       # Startup splash screen
        ├── toast.go        # Toast notifications
        ├── alerthistory.go # Alert history overlay
        └── progress.go     # Progress bar component
```

## How It Works

1. **Project Discovery**: On startup, devdash scans configured directories for `devenv.nix` files, identifying devenv projects.

2. **Socket Communication**: For each running project, devdash connects to the process-compose Unix socket at `.devenv/run/pc.sock` and communicates via the REST API.

3. **State Detection**: Project state is determined by attempting to connect to the socket:
   - Connection succeeds: Running
   - Socket file exists but connection fails: Stale
   - No socket file: Idle
   - Directory missing: Missing

4. **Log Streaming**: Logs are fetched via the process-compose API and displayed with automatic level detection (ERROR, WARN, INFO, DEBUG) and colorization.

5. **Health Monitoring**: devdash tracks service state changes and emits events for crashes and recoveries, triggering toast notifications and optional desktop notifications.

## Themes

### matrix (default)
Classic hacker aesthetic with bright green on black.

### nord
Arctic, bluish color palette based on the Nord theme.

### dracula
Purple-based dark theme with pink accents.

## ASCII Art Presets

The splash screen supports multiple ASCII art styles:
- `default` - Slant font
- `block` - Block letters with Unicode box characters
- `small` - Compact figlet style
- `minimal` - Simple bordered text
- `hacker` - Bold block style

## License

MIT

## Acknowledgments

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) - Style definitions
- [huh](https://github.com/charmbracelet/huh) - Form components
- [devenv](https://devenv.sh) - Developer environments
- [process-compose](https://github.com/F1bonacc1/process-compose) - Process orchestration
