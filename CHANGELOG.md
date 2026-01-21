# Changelog

All notable changes to acidBurn will be documented in this file.

This project uses [0ver](https://0ver.org/) versioning.

## [0.1.1] - 2026-01-21

### Added
- 10 new themes with light/dark variants:
  - Catppuccin family: Mocha, Macchiato, Frappé, Latte (light)
  - Tokyo Night family: Night, Storm, Day (light)
  - Gruvbox family: Dark, Light
- All themes available in settings panel theme selector
- ConfirmDialog component for user confirmations (50x10 centered modal)
  - Yes/No buttons with keyboard navigation (←/→, Tab)
  - Quick y/n shortcuts
  - Defaults to "No" for safety
- `Registry.RemoveProject()` for deleting projects from registry
- `Registry.ToggleHidden()` for hiding/showing projects
- `Project.Repair()` for cleaning up stale socket files

### Changed
- Help modal now renders as centered modal (80x28) instead of full-screen overlay
- Alerts now display in centered modal (80x28) instead of full-screen page
- Splash screen loading bar now animates smoothly from 0% to 100% over ~500ms
- Keybind brackets in help modal and footer now use theme accent colors
- Help keybinding label changed from "Alert history" to "Alerts"

### Technical
- Created `HelpPanel` component for modular help rendering
- Created `AlertsPanel` component for modular alerts rendering
- Created `ConfirmDialog` component following centered overlay pattern
- Removed `huh` dependency from settings implementation (now custom component)
- Help and alerts modals follow same centered overlay pattern as settings
- Registry methods for project lifecycle management
- Stale project repair removes entire `.devenv/run` directory

## [0.1.0] - 2026-01-20

### Added

#### Core Features
- Project discovery via filesystem scanning for `devenv.nix` files
- Project registry with persistent storage at `~/.config/acidburn/registry.yaml`
- Configuration system with YAML file at `~/.config/acidburn/config.yaml`
- Unix socket communication with process-compose daemons
- Project state detection (Running, Degraded, Idle, Stale, Missing)

#### User Interface
- Three-pane layout: sidebar (projects), services, and logs
- Keyboard-driven navigation with vim-style bindings
- Focus indicators and selection highlighting
- Startup splash screen with customizable ASCII art (5 presets)
- Settings panel for runtime configuration changes
- Toast notifications for service events
- Alert history overlay

#### Project Management
- Sidebar with active/idle project sections
- Project search and filtering
- Start idle projects with `devenv up -d`
- Stop running projects with API or `devenv down` fallback
- Automatic project state caching for consistent rendering

#### Service Control
- Service list with status, PID, CPU, memory, and exit code columns
- Start, stop, and restart individual services
- Filter logs by selected service

#### Log Viewing
- Real-time log streaming from process-compose API
- Circular buffer with 10,000 line capacity
- Log level detection (ERROR, WARN, INFO, DEBUG)
- Level-based colorization (red, yellow, default, muted)
- Timestamp parsing and display
- Follow mode (auto-scroll to new logs)
- Scroll navigation (up/down, page up/down, top/bottom)
- Log search with highlighting
- Search navigation (next/previous match)
- Filter mode (show only matching lines)
- Per-service log filtering

#### Health Monitoring
- Service crash detection with exit code tracking
- Service recovery detection
- Event history with timestamps
- Toast notifications for state changes

#### Notifications
- Desktop notifications via system notification daemon
- Configurable per-service notification overrides
- Critical-only mode option

#### Theming
- Three built-in themes: acid-green, nord, dracula
- Consistent color palette across all UI elements
- Configurable via settings panel

#### Configuration Options
- Project scan paths and depth
- Auto-discovery toggle
- Notification preferences
- UI customization (theme, sidebar width, timestamps)
- Polling intervals for focused/background projects

### Technical Details
- Built with Bubble Tea (Elm architecture for Go)
- Lip Gloss for styling
- huh for form components
- Cross-platform notification support (Linux, macOS, Windows)
- HTTP-over-Unix-socket for process-compose API

---

## Version History

| Version | Date | Summary |
|---------|------|---------|
| 0.1.1 | 2026-01-21 | UI polish: centered modals, theme expansion, animations |
| 0.1.0 | 2026-01-20 | Initial release |
