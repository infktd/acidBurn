# devdash Design Document

**Date:** 2026-01-20
**Status:** Approved
**Version:** 1.0

## Overview

devdash is a polished TUI command center for managing devenv.sh environments across macOS and Linux. It functions as "Docker Desktop for Nix" - services run as background daemons, and the TUI acts as a control plane that attaches/detaches freely.

### Tech Stack

- **Language:** Go
- **TUI Framework:** Bubble Tea (elm-architecture)
- **Styling:** Lipgloss
- **Components:** Bubbles (viewport, list, key), Huh (forms)
- **Logging:** Charm Log
- **Notifications:** beeep or go-notify for OS-level alerts

### Core Architecture

```
┌─────────────────────────────────────────────────────┐
│                    devdash TUI                      │
│  (Bubble Tea app - attaches/detaches from services) │
└──────────────────────┬──────────────────────────────┘
                       │ REST API / Socket
┌──────────────────────▼──────────────────────────────┐
│              process-compose (daemon)                │
│    (launched via `devenv up -d`, persists in bg)    │
└──────────────────────┬──────────────────────────────┘
                       │ manages
        ┌──────────────┼──────────────┐
        ▼              ▼              ▼
   [service]      [service]      [service]
   (user-defined in devenv.nix)
```

**Key Principle:** devdash never manages processes directly. It's a wrapper around devenv/process-compose, communicating via the process-compose REST API and socket at `.devenv/state/process-compose/`.

---

## Section 1: Project Discovery & Registry

### The Smart Registry Strategy

Auto-discovery populates a persistent registry.

**First Launch - Auto-Discovery:**
- Scan configured paths (default: `~/code`, `~/projects`) for `devenv.nix` files
- Limit depth to 3 levels for performance using `filepath.WalkDir`
- Populate the registry with discovered projects

**Scan Exclusions (hardcoded):**
```go
var scanExclusions = []string{
    "node_modules", ".git", ".direnv", "dist",
    "target", "vendor", ".venv", "__pycache__",
}
```

**Persistent Registry:**
```yaml
# ~/.config/devdash/projects.yaml
projects:
  - id: a1b2c3d4  # hash of path for uniqueness
    path: ~/code/my-saas
    name: my-saas  # user-editable alias
    hidden: false
    last_active: 2026-01-19T14:30:00Z
  - id: e5f6g7h8
    path: ~/code/personal/api
    name: personal-api
    hidden: true
    last_active: 2026-01-10T09:00:00Z
```

**State Detection:**

| State | Logic | UI |
|-------|-------|-----|
| Idle | No socket file | Gray |
| Running (Healthy) | Socket responds, all services OK | Green |
| Running (Degraded) | Socket responds, some services crashed | Yellow |
| Stale | Socket exists, connection refused | Red + "Repair" button |
| Missing | `os.Stat(path)` fails | Ghost icon + "Relocate/Remove" |

```go
func getProjectState(path string) ProjectState {
    socketPath := filepath.Join(path, ".devenv/state/process-compose/pc.sock")
    if conn, err := net.Dial("unix", socketPath); err == nil {
        conn.Close()
        return Running  // Socket exists and accepts connections
    }
    if _, err := os.Stat(socketPath); err == nil {
        return Stale    // Socket file exists but not responding
    }
    return Idle         // No socket, project not running
}
```

**Registry Operations:**
- **Auto-refresh:** Re-scan on launch (optional, configurable)
- **Manual add:** User can add paths not in scan directories
- **Hide/Show:** Toggle visibility without removing from registry
- **Remove:** Delete from registry entirely

---

## Section 2: UI Layout - The Cockpit

### Layout: 3-Pane Command Center

```
┌─ devdash ─── FLEET > MY-SAAS > COLLECTOR ─── PIDs: 8 ── MEM: 1.4GB ── Nix: OK ── 14:32 ● ─┐
│                                                                                              │
│  PROJECTS [2/8]         │  [Project: MY-SAAS]                                   ● HEALTHY   │
│  / search...            │ ┌────────────────────────────────────────────────────────────────┐│
│                         │ │ SERVICES [3]                        CPU    MEM    STATE        ││
│  ── ACTIVE ──           │ │ ● postgres                          0.2%   45MB   [R]          ││
│  ● my-saas     [245MB]  │ │ ● next-app                          1.4%   120MB  [R]          ││
│  ◐ data-api    [890MB]  │ │ ● collector                         8.5%   80MB   [R]          ││
│                         │ └────────────────────────────────────────────────────────────────┘│
│  ── IDLE ──             │                                                                   │
│  ○ blog                 │  LOGS: collector                                        [f]ollow  │
│  ○ side-project         │ ┌────────────────────────────────────────────────────────────────┐│
│  ✗ legacy-app           │ │ 14:32:01 [info] Fetching batch #402                            ││
│                         │ │ 14:32:05 [info] Inserted 50 rows into PG                       ││
│  ── GLOBAL ──           │ │ 14:32:10 [warn] Slow query detected (200ms)                    ││
│  ▤ ALL LOGS             │ │ 14:32:15 [info] Batch #402 complete                            ││
│                         │ └────────────────────────────────────────────────────────────────┘│
│                                                                                              │
└─ [TAB] Switch Focus  [s] Start All  [x] Stop All  [r] Restart  [/] Search  [?] Help ────────┘
```

### Top Bar (Global Telemetry)

| Section | Content |
|---------|---------|
| Left | Breadcrumbs: `FLEET > PROJECT > SERVICE` (shows focus path) |
| Center | Global stats: PIDs managed, total MEM, Nix daemon status |
| Right | Clock + Fleet Health LED (●/◐/✗) |

### Sidebar (Health Monitor)

- **Search:** `/` activates filter input at top
- **Grouped sections:** ACTIVE, IDLE, GLOBAL (dim Lipgloss headers)
- **Micro-stats:** Memory footprint per project `[245MB]`
- **Status glyphs:** `●` healthy, `◐` degraded, `○` idle, `✗` stale/missing
- **Fixed width:** 25 characters

### Main Panel (Action Center)

- **Top Box - Service Table:** High-density stats (CPU, MEM, STATE)
- **Bottom Box - Log Viewport:** Dedicated to output stream
- **Adaptive:** Services table shrinks when terminal is short

### Focus Navigation (3-Pane Model)

```
Tab cycle: Sidebar → Service Table → Log Viewport → Sidebar
```

Each pane has distinct actions available when focused. Focused pane has highlighted border (Lipgloss).

### Empty State

When no project selected, show "Fleet Overview" - aggregated stats, recent activity across all projects, or "Getting Started" for new users.

### Adaptive Layout

```go
func (m model) View() string {
    // 1. Render the Top Bar
    header := headerStyle.Render(fmt.Sprintf(" devdash ─── %s ─── Nix: %s", m.Breadcrumbs, m.NixStatus))

    // 2. Render the Sidebar (Fixed Width)
    sidebar := sidebarStyle.Width(25).Height(m.Height - 4).Render(m.Sidebar.View())

    // 3. Render the Main Area (Adaptive Width)
    mainWidth := m.Width - 25 - 4 // Subtract sidebar and borders

    servicesBox := mainStyle.Width(mainWidth).Render(m.ServiceTable.View())
    logsBox := mainStyle.Width(mainWidth).Render(m.LogViewport.View())

    // Stack the Main Area components vertically
    mainArea := lipgloss.JoinVertical(lipgloss.Left, servicesBox, logsBox)

    // 4. Join Sidebar and Main Area horizontally
    body := lipgloss.JoinHorizontal(lipgloss.Top, sidebar, mainArea)

    // 5. Render Footer
    footer := footerStyle.Render(m.Help.View())

    return lipgloss.JoinVertical(lipgloss.Left, header, body, footer)
}
```

---

## Section 3: Service Management & Background Processes

### Architecture: devdash as Control Plane

devdash never manages processes directly. It communicates with process-compose via:
- **Unix socket:** `.devenv/state/process-compose/pc.sock`
- **REST API:** process-compose exposes endpoints for status, start, stop, logs

### Lifecycle Commands

| Action | devdash executes | Result |
|--------|-------------------|--------|
| Start Project | `devenv up -d` | Launches process-compose daemon |
| Stop Project | API: `POST /project/stop` | Graceful shutdown of all services |
| Start Service | API: `POST /process/{name}/start` | Individual service start |
| Stop Service | API: `POST /process/{name}/stop` | Individual service stop |
| Restart Service | API: `POST /process/{name}/restart` | Stop + Start |

### Two-Tiered Exit

| Key | Action | Services |
|-----|--------|----------|
| `q` | Quit TUI | Keep running (detach) |
| `Ctrl+X` | Shutdown Project | Stop all, then quit |

### Re-attach on Launch

```go
func (m *model) discoverRunningProjects() {
    for _, project := range m.registry.Projects {
        socketPath := filepath.Join(project.Path, ".devenv/state/process-compose/pc.sock")
        if client, err := connectToSocket(socketPath); err == nil {
            project.State = Running
            project.Client = client  // Reuse connection for API calls
            m.syncServiceStatus(project)  // Fetch current service states
        }
    }
}
```

### Zombie Process Recovery

```go
func (m *model) repairProject(project *Project) error {
    // 1. Find orphaned processes
    orphans := findOrphanedProcesses(project.Path)
    if len(orphans) > 0 {
        // 2. Prompt user: "Found 3 orphaned processes. Kill them?"
        if m.confirmKillOrphans(orphans) {
            for _, pid := range orphans {
                syscall.Kill(pid, syscall.SIGKILL)
            }
        }
    }
    // 3. Clean up stale socket
    os.Remove(socketPath)
    // 4. Fresh start
    return exec.Command("devenv", "up", "-d").Run()
}
```

### Adaptive Polling

| Context | Interval | Rationale |
|---------|----------|-----------|
| Focused project | 2s | High-res stats for active debugging |
| Background (visible) | 10-15s | Catch crashes without overhead |
| Hidden projects | None | Poll on unhide or focus |

### Log Streaming Strategy

- **Single service:** Use process-compose API (simpler)
- **Unified "ALL LOGS":** Tail `.log` files directly, merge by timestamp (more performant for interleaving)

---

## Section 4: Log Viewing Experience

### Default: Single Focus (Deep Dive)

When a service is selected in the Service Table, logs viewport shows only that service's output.

```
┌─ LOGS: postgres ──────────────────────────────── [f]ollow ─┐
│ 14:32:01  ○  database system is ready                      │
│ 14:32:02  ○  listening on IPv4 address "0.0.0.0"           │
│ 14:32:03  ○  listening on port 5432                        │
│ 14:32:15  ○  connection received: host=127.0.0.1           │
│ █                                                          │
└────────────────────────────────────────────────────────────┘
```

### Unified Stream (System Pulse)

Selecting "▤ ALL LOGS" in sidebar switches to merged view with color-coded prefixes:

```
┌─ LOGS: ALL SERVICES ──────────────────────────── [f]ollow ─┐
│ 14:32:01 [POSTGRES]  database system is ready              │
│ 14:32:02 [NEXT-APP]  ready on http://localhost:3000        │
│ 14:32:03 [COLLECTOR] Starting batch sync...                │
│ 14:32:05 [POSTGRES]  connection received: host=127.0.0.1   │
│ 14:32:06 [COLLECTOR] Inserted 50 rows                      │
│ █                                                          │
└────────────────────────────────────────────────────────────┘
```

### Color Coding

```go
var serviceColors = map[string]lipgloss.Color{
    "postgres":  lipgloss.Color("#336791"),  // Postgres blue
    "next-app":  lipgloss.Color("#00D8FF"),  // Next.js cyan
    "collector": lipgloss.Color("#98C379"),  // Green
    // Auto-assign from palette for unknown services
}
```

### Log Level Icons

| Level | Icon | Color |
|-------|------|-------|
| info | `○` | dim gray |
| warn | `◐` | yellow |
| error | `●` | red bold |
| debug | `·` | very dim |

Timestamps are dimmed to reduce visual noise.

### Log Interactions

| Key | Action |
|-----|--------|
| `f` | Toggle follow mode (auto-scroll to bottom) |
| `pgup/pgdn` | Scroll through history |
| `/` | Search within logs (highlight matches) |
| `Ctrl+f` | Toggle filter mode (show only matching lines) |
| `g` / `G` | Jump to top / bottom |
| `w` | Wrap/unwrap long lines |
| `y` | Yank (copy) current line to clipboard |

### Sliding Window Buffer (Unified Stream)

```go
type LogInterleaver struct {
    buffer    []LogEntry
    flushTick *time.Ticker  // 50ms
}

func (li *LogInterleaver) flush() []LogEntry {
    sort.Slice(li.buffer, func(i, j int) bool {
        return li.buffer[i].Timestamp.Before(li.buffer[j].Timestamp)
    })
    entries := li.buffer
    li.buffer = nil
    return entries
}
```

### Performance: Segmented Rendering

```go
type LogBuffer struct {
    lines    [10000]string  // Circular buffer
    head     int
    rendered string         // Cache visible slice only
    dirty    bool
}

func (lb *LogBuffer) View(viewportHeight int) string {
    if !lb.dirty {
        return lb.rendered
    }
    // Only render visible lines + margin
    lb.rendered = lb.renderSlice(lb.head - viewportHeight - 50, lb.head)
    lb.dirty = false
    return lb.rendered
}
```

---

## Section 5: Alerts & Notifications

### Multi-Tiered Awareness

| Tier | Channel | When TUI is... | Purpose |
|------|---------|----------------|---------|
| In-TUI | Visual indicators | Open | Immediate feedback |
| System | OS notifications | Closed/backgrounded | Uninterrupted flow |

### In-TUI Alerts

1. **Status Glyph Changes:** `●` → `◐` → `✗` in sidebar
2. **Fleet LED:** Top-right LED changes/blinks based on fleet health
3. **Toast Notifications:** Slide-in banner for critical events
4. **Row Highlighting:** Crashed service row turns red in Service Table

```
┌─────────────────────────────────────────────────────────┐
│  ⚠ COLLECTOR crashed (exit code 1)          [x] dismiss │
└─────────────────────────────────────────────────────────┘
```

### LED Animation Hierarchy

| State | Animation | When |
|-------|-----------|------|
| Static Red | None | Service stopped (intentional) |
| Pulsing Red | Slow fade in/out via `tea.Tick` | Service crashed (non-zero exit) |
| Blinking Red | Fast toggle | Critical: Nix daemon down, OOM |
| Static Yellow | None | Degraded (some services down) |
| Static Green | None | Healthy |

### System Notifications (OS-level)

```go
func (m *model) notifyServiceCrash(service Service) {
    if m.config.Notifications.SystemEnabled {
        beeep.Alert(
            "devdash: Service Crashed",
            fmt.Sprintf("%s in %s exited unexpectedly", service.Name, service.Project),
            ""  // icon path
        )
    }
}
```

### Notification Events

| Event | In-TUI | System | Default |
|-------|--------|--------|---------|
| Service crashed | Always | Yes | On |
| Service restarted (auto) | Toast | Optional | Off |
| Project went idle | Glyph | Optional | Off |
| High memory (>80%) | Yellow | Optional | Off |
| Nix daemon down | Top bar | Yes | On |

### Per-Service Notification Config

```yaml
notifications:
  system_enabled: true
  tui_alerts: true

  overrides:
    - service: "webpack-watcher"
      system: false  # Too noisy, TUI only
    - service: "postgres"
      system: true
      critical_only: false
```

### Event Bus (Bubble Tea Pattern)

```go
type ServiceCrashedMsg struct {
    Project  string
    Service  string
    ExitCode int
}

func listenForHealthEvents(ch <-chan HealthEvent) tea.Cmd {
    return func() tea.Msg {
        event := <-ch
        switch event.Type {
        case Crashed:
            return ServiceCrashedMsg{...}
        case Recovered:
            return ServiceRecoveredMsg{...}
        }
        return nil
    }
}

// In Update()
case ServiceCrashedMsg:
    m.lastToast = formatCrashMessage(msg)
    m.showToast = true
    m.alertHistory = append(m.alertHistory, msg)
    return m, tea.Batch(
        func() tea.Msg { beeep.Alert(...); return nil },
        listenForHealthEvents(m.healthChan),
    )
```

### Notification History

- Buffer: Last 10 events
- Access: `H` key opens history overlay
- Shows: timestamp, event type, project/service, message
- Actions: `Enter` to jump to that project/service

```
┌─ ALERT HISTORY ────────────────────────────────────────────┐
│ 14:32:15  ●  CRASH    collector @ my-saas     exit code 1  │
│ 14:28:03  ◐  WARN     next-app @ my-saas      high memory  │
│ 14:15:44  ○  RECOVER  postgres @ data-api     restarted    │
│                                                            │
│ [Enter] Go to service  [Esc] Close                         │
└────────────────────────────────────────────────────────────┘
```

---

## Section 6: Configuration & Settings

### Source of Truth

`~/.config/devdash/config.yaml`

### Full Config Structure

```yaml
# ~/.config/devdash/config.yaml

# Project discovery
projects:
  scan_paths:
    - ~/code
    - ~/projects
  auto_discover: true
  scan_depth: 3

# Notifications
notifications:
  system_enabled: true
  tui_alerts: true
  critical_only: false

  overrides:
    - service: "webpack-watcher"
      system: false
    - service: "postgres"
      system: true

# UI preferences
ui:
  theme: "matrix"
  default_log_view: "focused"  # focused | unified
  log_follow: true
  show_timestamps: true
  dim_timestamps: true
  sidebar_width: 25

# Polling intervals (seconds)
polling:
  focused_project: 2
  background_project: 10

# Keybindings (future customization)
keys:
  quit: "q"
  shutdown: "ctrl+x"
  search: "/"
  help: "?"
```

### In-TUI Settings Panel (via huh)

Access: `S` key from any screen

```
┌─ SETTINGS ─────────────────────────────────────────────────┐
│                                                            │
│  ┌─ GENERAL ─────────────────────────────────────────────┐ │
│  │ Auto-discover projects    [●] On  [ ] Off             │ │
│  │ Scan depth                [3  ▼]                      │ │
│  │ Default log view          [Focused ▼]                 │ │
│  └───────────────────────────────────────────────────────┘ │
│                                                            │
│  ┌─ NOTIFICATIONS ───────────────────────────────────────┐ │
│  │ System notifications      [●] On  [ ] Off             │ │
│  │ TUI alerts                [●] On  [ ] Off             │ │
│  │ Critical only             [ ] On  [●] Off             │ │
│  └───────────────────────────────────────────────────────┘ │
│                                                            │
│  ┌─ APPEARANCE ──────────────────────────────────────────┐ │
│  │ Theme                     [Acid Green ▼]              │ │
│  │ Dim timestamps            [●] On  [ ] Off             │ │
│  └───────────────────────────────────────────────────────┘ │
│                                                            │
│  [Enter] Save  [e] Edit in $EDITOR  [Esc] Cancel           │
└────────────────────────────────────────────────────────────┘
```

### Config Lifecycle (Watch/Reload)

```go
// After huh form saves
func (m *model) saveConfig() tea.Cmd {
    data, _ := yaml.Marshal(m.config)
    os.WriteFile(configPath, data, 0644)
    return func() tea.Msg {
        return ConfigUpdatedMsg{m.config}
    }
}

// After $EDITOR returns
func (m *model) reloadConfig() tea.Cmd {
    return tea.ExecProcess(
        exec.Command(os.Getenv("EDITOR"), configPath),
        func(err error) tea.Msg {
            newConfig := loadConfigFromDisk()
            return ConfigUpdatedMsg{newConfig}
        },
    )
}

// In Update()
case ConfigUpdatedMsg:
    m.config = msg.Config
    m.applyTheme(msg.Config.UI.Theme)
    m.updatePollingIntervals()
    return m, nil
```

### Theme Definitions

```go
var themes = map[string]Theme{
    "matrix": {
        Primary:    lipgloss.Color("#39FF14"),
        Secondary:  lipgloss.Color("#00FF41"),
        Background: lipgloss.Color("#0D0D0D"),
        Muted:      lipgloss.Color("#4A4A4A"),
    },
    "nord": {
        Primary:    lipgloss.Color("#88C0D0"),
        Secondary:  lipgloss.Color("#81A1C1"),
        Background: lipgloss.Color("#2E3440"),
        Muted:      lipgloss.Color("#4C566A"),
    },
    "dracula": {
        Primary:    lipgloss.Color("#BD93F9"),
        Secondary:  lipgloss.Color("#FF79C6"),
        Background: lipgloss.Color("#282A36"),
        Muted:      lipgloss.Color("#6272A4"),
    },
}
```

### Theme Persistence

```go
func (m *model) applyTheme(themeName string) {
    theme := themes[themeName]

    m.styles.Sidebar = lipgloss.NewStyle().
        BorderForeground(theme.Primary).
        Foreground(theme.Secondary)

    m.styles.Header = lipgloss.NewStyle().
        Background(theme.Background).
        Foreground(theme.Primary).
        Bold(true)

    // Apply to huh forms
    m.settingsForm = huh.NewForm(...).
        WithTheme(huh.ThemeBase().
            Focused.Base.BorderForeground(theme.Primary).
            Blurred.Base.BorderForeground(theme.Muted))
}
```

### Splash Screen

Shown on first launch during initial project scan:

```
┌────────────────────────────────────────────────────────────┐
│                                                            │
│                    [ASCII ART HERE]                        │
│                       devdash                             │
│                                                            │
│              Scanning for devenv projects...               │
│              ████████████░░░░░░░░░  52%                    │
│                                                            │
└────────────────────────────────────────────────────────────┘
```

- Theme-colored ASCII art (customizable)
- Progress bar for scan status
- Can be skipped with any key

---

## Section 7: Keybindings & Help System

### Philosophy: Hybrid (Vim + Arrows)

Both work everywhere - zero friction for all users.

### Keybinding Layers

| Layer | Context | Navigation | Actions |
|-------|---------|------------|---------|
| **Global** | Always | `Tab` (cycle focus), `?` (help) | `q` (quit), `Ctrl+X` (shutdown), `S` (settings), `H` (alert history), `R` (refresh) |
| **Sidebar** | Projects list | `j/k` or `↑/↓`, `/` (search) | `s` (start), `x` (stop), `Enter` (select) |
| **Services** | Service table | `j/k` or `↑/↓` | `r` (restart), `x` (stop), `s` (start), `Enter` (focus logs) |
| **Logs** | Log viewport | `j/k` or `↑/↓`, `g/G` (top/bottom), `PgUp/PgDn` | `f` (follow), `/` (search), `Ctrl+f` (filter), `w` (wrap), `y` (yank) |

### KeyMap Implementation (bubbles/key)

```go
import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
    Quit     key.Binding
    Shutdown key.Binding
    Settings key.Binding
    Help     key.Binding
    Up       key.Binding
    Down     key.Binding
    Select   key.Binding
    Back     key.Binding
    Start    key.Binding
    Stop     key.Binding
    Restart  key.Binding
}

func DefaultKeyMap() KeyMap {
    return KeyMap{
        Quit: key.NewBinding(
            key.WithKeys("q"),
            key.WithHelp("q", "quit (detach)"),
        ),
        Shutdown: key.NewBinding(
            key.WithKeys("ctrl+x"),
            key.WithHelp("ctrl+x", "shutdown all"),
        ),
        Up: key.NewBinding(
            key.WithKeys("k", "up"),
            key.WithHelp("↑/k", "up"),
        ),
        Down: key.NewBinding(
            key.WithKeys("j", "down"),
            key.WithHelp("↓/j", "down"),
        ),
        // ... etc
    }
}

func (k KeyMap) ShortHelp() []key.Binding {
    return []key.Binding{k.Up, k.Down, k.Start, k.Stop, k.Help}
}
```

### Context-Sensitive Footer

Updates based on focused pane:

```
Sidebar focused:
└─ [↑/↓] Navigate  [Enter] Select  [s] Start  [x] Stop  [/] Search  [?] Help ─┘

Services focused:
└─ [↑/↓] Navigate  [r] Restart  [x] Stop  [s] Start  [Enter] View Logs  [?] Help ─┘

Logs focused:
└─ [↑/↓] Scroll  [f] Follow  [/] Search  [Ctrl+f] Filter  [g/G] Top/Bottom  [?] Help ─┘
```

### Help Overlay

```
┌─ KEYBINDINGS ──────────────────────────────────────────────┐
│                                                            │
│  GLOBAL                     NAVIGATION                     │
│  q       Quit (detach)      ↑/k     Up                     │
│  Ctrl+X  Shutdown all       ↓/j     Down                   │
│  S       Settings           Tab     Switch pane            │
│  H       Alert history      Enter   Select/Confirm         │
│  R       Refresh            Esc     Back/Cancel            │
│  ?       This help                                         │
│                                                            │
│  SIDEBAR                    SERVICES                       │
│  s       Start project      s       Start service          │
│  x       Stop project       x       Stop service           │
│  /       Search projects    r       Restart service        │
│                                                            │
│  LOGS                                                      │
│  f       Toggle follow      g/G     Top/Bottom             │
│  /       Search logs        PgUp    Page up                │
│  Ctrl+f  Filter mode        PgDn    Page down              │
│  w       Toggle wrap        y       Yank line              │
│                                                            │
│                                         [Esc] Close        │
└────────────────────────────────────────────────────────────┘
```

---

## Summary: Feature Stack

| Layer | Component | Purpose |
|-------|-----------|---------|
| 0. Core | Go + Bubble Tea | Reactive TUI engine |
| 1. Registry | Smart Discovery | Auto-find `devenv.nix`, persist to `projects.yaml` |
| 2. Interface | The Cockpit | 3-pane layout with telemetry header |
| 3. Backend | REST/Socket Control | `process-compose` as daemon, devdash as control plane |
| 4. Feedback | Log Interleaver | Color-coded unified/focused streams with search |
| 5. Awareness | Multi-Tier Alerts | OS notifications + in-TUI toasts + history |
| 6. Config | YAML + huh Forms | Reproducible config with TUI editor |
| 7. Control | Hybrid Keybindings | Vim + arrows, context-sensitive help |
| 8. Polish | Splash Screen | Themed ASCII art during scan |

---

## Next Steps

1. Initialize Go module and project structure
2. Set up git worktree for isolated development
3. Create detailed implementation plan
4. Begin with core Bubble Tea scaffolding
