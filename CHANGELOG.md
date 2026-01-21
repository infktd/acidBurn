# Changelog

All notable changes to acidBurn will be documented in this file.

This project uses [0ver](https://0ver.org/) versioning.

## [Unreleased]

### Added
- Real-time log flow indicators in services table
  - Animated braille spinner (⣾⣽⣻⢿⡿⣟⣯⣷) shows when service is actively logging
  - Appears in new ACTIVITY column between STATUS and SERVICE
  - Indicator displays when logs received within last 2 seconds
  - Smooth 100ms animation cycle for visual feedback
- Animated state transition effects
  - Status text flashes bold when service state changes (Running ↔ Stopped)
  - Smooth 1.5 second fade from bright to normal intensity
  - Green flash for Running state, yellow/orange for Stopped state
  - Immediately draws attention to state changes
- Live CPU/Memory sparklines in services table
  - Inline mini-graphs using Unicode block characters (▁▂▃▄▅▆▇█)
  - Displays last 10 data points showing resource usage trends
  - Appears next to current CPU/Memory values (e.g., "45.2% ▃▄▅▇▆▅▄▃")
  - Auto-normalizes to show relative changes clearly
  - Visible after 3+ readings collected
- Bidirectional panel navigation
  - Tab cycles forward through panels (Projects → Services → Logs)
  - Shift+Tab cycles backward through panels (Projects ← Services ← Logs)
  - Tab keybinding now visible in footer for all panes

### Changed
- Services table column order reorganized: STATUS | ACTIVITY | SERVICE | PID | CPU | MEM | EXIT
- Log activity tracked per service with timestamp precision
- Activity indicators animate independently using dedicated ticker (100ms intervals)
- Footer keybindings now centered for better visual balance

### Fixed
- Projects list navigation glitch when pressing up at top of list
- Navigation keys no longer passed to list component (prevents double-handling)

### Technical
- Added `logActivity` map tracking last log timestamp per service
- Added `activityFrame` counter for spinner animation state
- Implemented `activityTickMsg` and `activityTickCmd()` for animation updates
- Created `getActivityIndicator()` helper with 8-frame braille spinner
- Services table updates every animation frame to display current spinner state
- Added `serviceStates`, `stateChangeTime`, `stateFlashIntensity` maps for transition tracking
- Implemented `applyFlashEffect()` for state transition pulse/flash animation
- Flash intensity decays exponentially over 1.5 seconds using linear interpolation
- Added `cpuHistory` and `memHistory` maps tracking last 10 readings per service
- Implemented `renderSparkline()` and `renderMemorySparkline()` for Unicode graph generation
- Sparklines use 8 Unicode block characters (▁▂▃▄▅▆▇█) for smooth gradients
- Auto-normalization based on min/max values in history window
- CPU and MEM columns widened to 18 and 20 characters for sparkline display
- Added `ShiftTab` key binding for reverse panel navigation
- Implemented `cycleFocusReverse()` using modulo arithmetic

## [0.1.2] - 2026-01-21

### Added
- Spinner animations for project start/stop/restart operations
- Loading indicators for project state transitions
- Section title styling with `// TITLE ─────────────────` format
- Custom project filter with full theme integration (press `/` to activate)
  - Theme-colored "Filter:" prompt and input text
  - Blinking cursor during input mode
  - Case-insensitive substring search
  - Esc to clear, Enter to apply
- Five new themes: Dracula, Nord, Ayu Dark, Solarized Dark, Monokai
- ACTIVE/IDLE section headers in project list

### Changed
- Sidebar now uses bubbles/list component with built-in fuzzy search
- Services pane now uses bubbles/table component for cleaner, aligned columns
- Section titles styled with code comment format (`// TITLE ───────`)
- Section title colors dynamically change based on pane focus (primary when focused, muted when not)
- Modal borders changed to rounded style for softer appearance
- Projects visually grouped under ACTIVE and IDLE section headers
- Project items indented under section headers for clear hierarchy
- Removed all light themes, Catppuccin variants, and Oxocarbon
- Curated theme selection of 8 distinct dark themes: Acid Green, Gruvbox, Dracula, Nord, Tokyo Night, Ayu Dark, Solarized Dark, Monokai

### Fixed
- Custom project filter now uses theme colors (primary color for "Filter:" text and input)
- Filter mode now captures ALL keys (prevents keybinds like 's', 'x' from triggering while typing)
- Filter prompt "Filter:" text and cursor now properly styled with theme primary color
- Theme colors now update immediately when changed (no restart required)
- Esc key properly exits and clears filter mode
- Enter key applies filter and exits input mode (filter remains active)
- Removed "2items" counter by disabling status bar (was counting section headers as items)

### Technical
- Integrated `bubbles/spinner` for loading state animations
- Integrated `bubbles/table` for services display with proper column alignment
- Integrated `bubbles/list` for sidebar with built-in fuzzy search filtering
- Added custom `projectDelegate` implementing `list.ItemDelegate` for project rendering
- Added `sectionHeaderItem` type for ACTIVE/IDLE headers
- Section headers skip navigation (cursor jumps over them)
- Added `projectIndexToListIndex()` and `listIndexToProjectIndex()` for header-aware indexing
- Removed custom project search implementation (replaced by list's built-in search)
- Added `renderSectionTitle()` function for consistent section header styling with theme awareness
- Added `loadingOp` and `loadingProject` fields to Model for spinner state tracking
- Spinner displays in place of status glyph during project operations
- List component now receives messages for filter handling via `projectsList.Update(msg)`
- Implemented custom project filter UI (replaced bubbles/list built-in filter for full style control)
- Added `projectFilterMode` and `projectFilterInput` fields to Model for custom filter state
- Custom filter rendering in `renderSidebar()` with theme-aware styles
- Filter displays below "PROJECTS" title with blinking cursor during input mode
- Filter logic in `updateDisplayedProjects()` uses case-insensitive substring matching
- Global key handler routes keys to sidebar when in custom filter mode
- Navigation keys (Up/Down) disabled during filter mode to allow 'j'/'k' in search terms
- Disabled list status bar (`SetShowStatusBar(false)`) and list filtering (`SetFilteringEnabled(false)`)
- Backspace key removes characters from filter input

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
| 0.1.2 | 2026-01-21 | UI components: spinners, table, list with fuzzy search, styled section titles |
| 0.1.1 | 2026-01-21 | UI polish: centered modals, theme expansion, animations |
| 0.1.0 | 2026-01-20 | Initial release |
