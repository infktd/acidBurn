<div align="center">

# devdash

**A Terminal Dashboard for Managing devenv.sh Projects**

*Think Docker Desktop for Nix—a unified control plane for all your development environments*

[![Made with Nix](https://img.shields.io/badge/Made_with-Nix-5277C3?logo=nixos&logoColor=white)](https://nixos.org)
[![Made with devenv](https://img.shields.io/badge/Made_with-devenv-00D9FF?logo=data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMTAwIiBoZWlnaHQ9IjEwMCIgdmlld0JveD0iMCAwIDEwMCAxMDAiIGZpbGw9Im5vbmUiIHhtbG5zPSJodHRwOi8vd3d3LnczLm9yZy8yMDAwL3N2ZyI+PHBhdGggZD0iTTUwIDEwTDkwIDUwTDUwIDkwTDEwIDUwTDUwIDEwWiIgZmlsbD0id2hpdGUiLz48L3N2Zz4=&logoColor=white)](https://devenv.sh)
[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![Test Coverage](https://img.shields.io/badge/coverage-59.8%25-brightgreen?logo=go&logoColor=white)](https://github.com/infktd/devdash/actions)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![GitHub Stars](https://img.shields.io/github/stars/infktd/devdash?style=social)](https://github.com/infktd/devdash)

[Features](#-features) •
[Installation](#-installation) •
[Quick Start](#-quick-start) •
[Documentation](#-documentation) •
[Contributing](#-contributing)

</div>

---

## Overview

**devdash** is a terminal user interface (TUI) for managing [devenv.sh](https://devenv.sh) development environments. It provides centralized monitoring and control for multiple devenv projects, eliminating the need to juggle terminal windows and manual process management.

### The Problem

When working with multiple devenv projects, developers face several challenges:

- 📁 **Scattered projects**: Manually navigating between project directories
- 🔄 **Process management**: Running separate `devenv up` commands for each project
- 📊 **No visibility**: Checking service status requires multiple terminal tabs
- 📝 **Log chaos**: Monitoring logs across different projects and services
- 🚨 **Silent failures**: Service crashes go unnoticed until something breaks

### The Solution

devdash provides a **unified control tower** for all your devenv projects:

```
┌─ Projects ────────────────┐┌─ Services (my-api) ─────────────────────────┐
│ > my-api         [Running]││  postgres    Running  ↑ 2h 15m  CPU: 2%    │
│   blog-site      [Idle]   ││  redis       Running  ↑ 2h 15m  CPU: 1%    │
│   data-pipeline  [Stopped]││  api-server  Running  ↑ 2h 14m  CPU: 15%   │
└───────────────────────────┘│  worker      Stopped  ↓ 5m ago             │
                             └──────────────────────────────────────────────┘
┌─ Logs (api-server) ──────────────────────────────────────────────────────┐
│ 2026-01-21 21:30:45 INFO  Server starting on port 8080                   │
│ 2026-01-21 21:30:46 INFO  Connected to database                          │
│ 2026-01-21 21:31:02 DEBUG Handling request GET /api/users               │
└──────────────────────────────────────────────────────────────────────────┘
```

<!--
TODO: Add actual screenshot here
![devdash interface](docs/screenshots/main-view.png)
-->

---

## ✨ Features

### Core Capabilities

- **🔍 Automatic Discovery** - Scans configured directories to find all devenv projects
- **🎛️ Multi-Project Management** - Switch between projects instantly from a unified sidebar
- **⚡ Service Control** - Start, stop, and restart services with single keystrokes
- **📊 Real-Time Monitoring** - Live service status, uptime, and resource usage
- **📝 Advanced Logging** - Streaming logs with search, filtering, and syntax highlighting
- **💚 Health Monitoring** - Automatic crash detection and recovery tracking
- **🔔 Smart Notifications** - Desktop alerts for critical events (configurable per service)
- **🎨 Beautiful Themes** - Multiple color schemes (Matrix, Nord, Dracula)
- **⚙️ Highly Configurable** - Customize scan paths, polling intervals, UI behavior

### Why devdash?

| Without devdash | With devdash |
|----------------|--------------|
| `cd ~/project1 && devenv up` | **One unified dashboard** |
| `cd ~/project2 && devenv up` | **All projects in one view** |
| *Check logs in terminal 1* | **Aggregated log streaming** |
| *Check logs in terminal 2* | **Instant project switching** |
| *Service crashed? Didn't notice* | **Desktop notifications** |
| *Manually grep logs for errors* | **Built-in search & filtering** |

---

## 📦 Installation

### Nix Flakes (Recommended)

```bash
# Run directly
nix run github:infktd/devdash

# Install to profile
nix profile install github:infktd/devdash

# Add to your flake.nix
{
  inputs.devdash.url = "github:infktd/devdash";
  # ...
}
```

### Go Install

```bash
go install github.com/infktd/devdash@latest
```

### Build from Source

```bash
git clone https://github.com/infktd/devdash
cd devdash
nix build    # or: go build -o devdash .
./result/bin/devdash
```

### System Requirements

- **Operating System**: Linux, macOS
- **Dependencies**:
  - [devenv.sh](https://devenv.sh) (the projects you're managing)
  - process-compose (included with devenv)
- **Optional**: `beeep` for desktop notifications

---

## 🚀 Quick Start

### 1. First Launch

```bash
devdash
```

On first run, devdash will:
- Create default configuration at `~/.config/devdash/config.yaml`
- Scan `~/coding` for devenv projects (configurable)
- Build a project registry at `~/.config/devdash/registry.yaml`

### 2. Configure Scan Paths

Edit `~/.config/devdash/config.yaml` to customize where devdash looks for projects:

```yaml
projects:
  scan_paths:
    - ~/code
    - ~/work
    - ~/projects
  scan_depth: 3        # How deep to search for devenv.nix
  auto_discover: true  # Automatically find new projects
```

### 3. Start Using

| Action | Keybinding |
|--------|-----------|
| Navigate projects | `↑` / `↓` or `j` / `k` |
| Select project | `Enter` |
| Start project services | `s` |
| Stop project services | `x` |
| Restart service | `r` (in services pane) |
| Search logs | `/` (in logs pane) |
| Toggle log follow | `f` |
| Open settings | `S` |
| Show help | `?` |
| Quit | `q` |

---

## 📖 Documentation

### Project States

devdash automatically detects and displays project states:

| State | Icon | Description |
|-------|------|-------------|
| **Running** | 🟢 | process-compose daemon active and responding |
| **Degraded** | 🟡 | Some services crashed (exit code ≠ 0) |
| **Idle** | ⚪ | No daemon running (project stopped) |
| **Stale** | 🔴 | Socket exists but daemon not responding |
| **Missing** | ❌ | Project directory no longer exists |

### Complete Keybindings

<details>
<summary><strong>Global Shortcuts</strong></summary>

| Key | Action |
|-----|--------|
| `q` | Quit (detach from projects, leave running) |
| `Ctrl+X` | Emergency shutdown (stop all services and quit) |
| `S` | Open settings panel |
| `H` | View alert history |
| `?` | Show help overlay |
| `Tab` | Cycle focus between panes |
| `Shift+Tab` | Cycle focus backwards |

</details>

<details>
<summary><strong>Projects Sidebar</strong></summary>

| Key | Action |
|-----|--------|
| `↑` / `↓` or `j` / `k` | Navigate projects |
| `/` | Search/filter projects |
| `Enter` | Select and switch to project |
| `s` | Start entire project (all services) |
| `x` | Stop entire project |
| `Esc` | Clear search filter |

</details>

<details>
<summary><strong>Services Pane</strong></summary>

| Key | Action |
|-----|--------|
| `↑` / `↓` or `j` / `k` | Navigate services |
| `Enter` | Filter logs to selected service |
| `s` | Start service |
| `x` | Stop service |
| `r` | Restart service |
| `Esc` | Return focus to sidebar |

</details>

<details>
<summary><strong>Logs Pane</strong></summary>

| Key | Action |
|-----|--------|
| `↑` / `↓` or `j` / `k` | Scroll logs line by line |
| `Page Up` / `Page Down` | Scroll logs by page |
| `g` | Jump to top of logs |
| `G` | Jump to bottom of logs |
| `f` | Toggle follow mode (auto-scroll) |
| `/` | Start search |
| `n` | Next search match |
| `N` | Previous search match |
| `Ctrl+F` | Toggle filter mode (show only matches) |
| `Esc` | Clear search/filter |

</details>

### Configuration Reference

<details>
<summary><strong>Full config.yaml Example</strong></summary>

```yaml
projects:
  scan_paths:
    - ~/code
    - ~/projects
    - ~/work
  auto_discover: true
  scan_depth: 3

notifications:
  system_enabled: true     # Desktop notifications
  tui_alerts: true         # In-app toast messages
  critical_only: false     # Only notify on critical events
  overrides:               # Per-service overrides
    - service: postgres
      system: true
      critical_only: false
    - service: redis
      system: false        # Disable desktop notifications for redis

ui:
  theme: matrix           # matrix | nord | dracula
  default_log_view: focused
  log_follow: true
  show_timestamps: true
  dim_timestamps: true
  sidebar_width: 25

polling:
  focused_project: 2      # Poll active project every 2 seconds
  background_project: 10  # Poll background projects every 10 seconds
```

</details>

### Themes

| Theme | Description | Best For |
|-------|-------------|----------|
| `matrix` | Classic green-on-black hacker aesthetic | Terminal purists |
| `nord` | Arctic, bluish color palette | Modern, clean look |
| `dracula` | Purple with pink accents | Vibrant, playful style |

<!--
TODO: Add theme screenshots
![Themes comparison](docs/screenshots/themes.png)
-->

---

## 🏗️ Architecture

devdash is built with Go and the [Bubble Tea](https://github.com/charmbracelet/bubbletea) TUI framework.

### How It Works

```
┌─────────────┐
│   devdash   │
└──────┬──────┘
       │
       ├─ Scans filesystem for devenv.nix files
       ├─ Connects to Unix sockets (.devenv/state/process-compose/pc.sock)
       ├─ Queries process-compose REST API
       ├─ Streams logs and service status
       └─ Monitors health, triggers notifications
```

**Key Components:**

1. **Scanner** - Discovers devenv projects by searching for `devenv.nix`
2. **Registry** - Maintains list of known projects and their states
3. **Compose Client** - Communicates with process-compose via Unix sockets
4. **Health Monitor** - Tracks service state changes, detects crashes
5. **UI Layer** - Renders interactive terminal interface with Bubble Tea

<details>
<summary><strong>Project Structure</strong></summary>

```
devdash/
├── main.go                    # Application entry point
├── internal/
│   ├── compose/               # process-compose API client
│   │   ├── client.go          # Unix socket HTTP client
│   │   └── types.go           # API response structures
│   ├── config/                # Configuration management
│   │   ├── config.go          # YAML config loading/saving
│   │   └── types.go           # Configuration structs
│   ├── health/                # Service health monitoring
│   │   └── monitor.go         # Crash detection, recovery tracking
│   ├── notify/                # Desktop notifications
│   │   └── notify.go          # Cross-platform notification delivery
│   ├── packages/              # Nix package scanning
│   │   └── scanner.go         # Parse and categorize installed packages
│   ├── registry/              # Project registry
│   │   ├── registry.go        # Project list persistence
│   │   └── types.go           # Project metadata, state detection
│   ├── scanner/               # Project discovery
│   │   └── scanner.go         # Filesystem scanning for devenv.nix
│   └── ui/                    # Terminal user interface
│       ├── model.go           # Main Bubble Tea model
│       ├── keys.go            # Keybinding definitions
│       ├── styles.go          # Lipgloss styles
│       ├── theme.go           # Color theme definitions
│       ├── logview.go         # Log viewport with search
│       ├── logbuffer.go       # Circular log buffer
│       ├── settings.go        # Settings panel (huh forms)
│       ├── splash.go          # Startup animation
│       ├── toast.go           # Toast notifications
│       ├── alerthistory.go    # Alert history overlay
│       ├── packages.go        # Nix packages pane
│       └── progress.go        # Progress indicators
├── docs/
│   └── plans/                 # Development planning documents
├── flake.nix                  # Nix flake for package distribution
└── README.md                  # This file
```

</details>

---

## 🧪 Testing

devdash has comprehensive test coverage (59.8%):

```bash
# Run all tests
go test ./...

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run specific package tests
go test ./internal/ui -v
```

**Test Suite:**
- 317 total tests across all packages
- Unit tests for all core components
- Integration tests for API clients
- UI component tests

---

## 🤝 Contributing

Contributions are welcome! Whether it's bug reports, feature requests, or pull requests.

### Reporting Issues

Found a bug? [Open an issue](https://github.com/infktd/devdash/issues/new) with:
- devdash version (`devdash --version` or git commit)
- Operating system and version
- Steps to reproduce
- Expected vs. actual behavior
- Relevant logs (check `~/.config/devdash/devdash.log`)

### Development Setup

```bash
# Clone the repository
git clone https://github.com/infktd/devdash
cd devdash

# Option 1: Using Nix (recommended)
nix develop   # Enters development shell with all dependencies
go build

# Option 2: Using Go directly
go mod download
go build -o devdash .
./devdash
```

### Pull Requests

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Add tests if applicable
5. Ensure tests pass (`go test ./...`)
6. Commit with clear messages
7. Push to your fork
8. Open a Pull Request

**Guidelines:**
- Follow existing code style
- Add tests for new features
- Update documentation as needed
- Keep commits focused and atomic

---

## 📄 License

This project is licensed under the **MIT License** - see the [LICENSE](LICENSE) file for details.

---

## 🙏 Acknowledgments

Built with excellent open-source tools:

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) - Terminal styling
- [huh](https://github.com/charmbracelet/huh) - Interactive forms
- [devenv](https://devenv.sh) - Fast, declarative development environments
- [process-compose](https://github.com/F1bonacc1/process-compose) - Process orchestration

Special thanks to the Nix and devenv communities for building amazing developer tools.

---

## 🔗 Links

- **Homepage**: [github.com/infktd/devdash](https://github.com/infktd/devdash)
- **Issues**: [Report bugs or request features](https://github.com/infktd/devdash/issues)
- **devenv.sh**: [Learn about devenv](https://devenv.sh)
- **Nix**: [Learn about Nix](https://nixos.org)

---

<div align="center">

**Made with ❤️ by developers, for developers**

If devdash helps you manage your devenv projects, consider giving it a ⭐!

</div>
