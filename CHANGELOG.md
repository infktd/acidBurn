# Changelog

All notable changes to devdash will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [0.1.0] - 2026-01-21

Initial release of devdash - a terminal dashboard for managing devenv.sh projects.

### Features

**Project Management**
- Automatic discovery of devenv projects by scanning configured directories
- Project registry with persistent storage
- Real-time project state detection (Running, Degraded, Idle, Stale, Missing)
- Start and stop entire projects from the dashboard
- Project search and filtering
- Hide/show projects in the registry

**Service Control**
- View all services for selected project with status, uptime, CPU, and memory usage
- Start, stop, and restart individual services
- Real-time service health monitoring
- Crash detection and recovery tracking

**Log Viewing**
- Live log streaming from process-compose
- Search logs with highlighting and match navigation
- Filter logs to show only selected service
- Auto-follow mode for real-time monitoring
- Log level detection and colorization (ERROR, WARN, INFO, DEBUG)
- Scroll navigation with vim-style keybindings

**Nix Package Inspector**
- View installed packages in current devenv environment
- Categorized package display with descriptions and versions
- Adaptive layout for narrow terminals

**Notifications**
- Desktop notifications for service crashes and recoveries (macOS/Linux)
- In-app toast notifications
- Per-service notification configuration
- Critical-only mode option

**Themes**
- 8 built-in color schemes:
  - Matrix (classic green-on-black)
  - Gruvbox (retro warm colors)
  - Dracula (purple with pink accents)
  - Nord (arctic blue palette)
  - Tokyo Night (modern blue/purple)
  - Ayu Dark (golden/orange accents)
  - Solarized Dark (muted scientific palette)
  - Monokai (classic editor theme)

**Configuration**
- YAML configuration file at `~/.config/devdash/config.yaml`
- Customizable project scan paths and depth
- Adjustable polling intervals
- UI preferences (theme, sidebar width, timestamps)
- Notification preferences

### Installation

```bash
# Nix Flakes
nix run github:infktd/devdash

# Go
go install github.com/infktd/devdash@latest

# From source
git clone https://github.com/infktd/devdash
cd devdash
nix build
```

### Requirements

- devenv.sh
- process-compose (included with devenv)
- Linux or macOS

---

[0.1.0]: https://github.com/infktd/devdash/releases/tag/v0.1.0
