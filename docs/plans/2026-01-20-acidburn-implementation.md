# acidBurn Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build a polished TUI command center for managing devenv.sh environments across macOS and Linux.

**Architecture:** acidBurn acts as a control plane over process-compose daemons. It discovers projects by scanning for `devenv.nix` files, maintains a persistent registry, and provides a 3-pane "cockpit" UI with real-time service stats, logs, and alerts.

**Tech Stack:** Go 1.21+, Bubble Tea, Lipgloss, Bubbles (viewport, list, key), Huh (forms), beeep (notifications), gopkg.in/yaml.v3

---

## Phase 1: Project Setup & Core Scaffolding

### Task 1.1: Initialize Go Module

**Files:**
- Create: `go.mod`
- Create: `main.go`

**Step 1: Initialize the Go module**

Run:
```bash
cd /home/infktd/coding/acidBurn && go mod init github.com/infktd/acidburn
```

Expected: `go.mod` created with module name

**Step 2: Create minimal main.go**

Create `main.go`:
```go
package main

import "fmt"

func main() {
	fmt.Println("acidBurn")
}
```

**Step 3: Verify it compiles**

Run: `go build -o acidburn .`
Expected: Binary `acidburn` created

**Step 4: Run it**

Run: `./acidburn`
Expected: Outputs "acidBurn"

**Step 5: Commit**

```bash
git add go.mod main.go
git commit -m "chore: initialize go module"
```

---

### Task 1.2: Add Charm Dependencies

**Files:**
- Modify: `go.mod`

**Step 1: Add all Charm dependencies**

Run:
```bash
go get github.com/charmbracelet/bubbletea@latest
go get github.com/charmbracelet/lipgloss@latest
go get github.com/charmbracelet/bubbles@latest
go get github.com/charmbracelet/huh@latest
go get github.com/charmbracelet/log@latest
go get github.com/gen2brain/beeep@latest
go get gopkg.in/yaml.v3@latest
```

**Step 2: Tidy dependencies**

Run: `go mod tidy`

**Step 3: Verify import works**

Update `main.go`:
```go
package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Silence unused import warnings during setup
var _ = tea.Quit
var _ = lipgloss.Color("")

func main() {
	fmt.Println("acidBurn - dependencies loaded")
}
```

**Step 4: Build to verify**

Run: `go build -o acidburn .`
Expected: Compiles without errors

**Step 5: Commit**

```bash
git add go.mod go.sum main.go
git commit -m "chore: add charm and notification dependencies"
```

---

### Task 1.3: Create Project Directory Structure

**Files:**
- Create: `internal/config/config.go`
- Create: `internal/registry/registry.go`
- Create: `internal/ui/ui.go`
- Create: `internal/compose/client.go`
- Create: `internal/scanner/scanner.go`

**Step 1: Create directory structure with placeholder files**

```bash
mkdir -p internal/config internal/registry internal/ui internal/compose internal/scanner
```

**Step 2: Create placeholder files**

Create `internal/config/config.go`:
```go
// Package config handles acidBurn configuration loading and persistence.
package config
```

Create `internal/registry/registry.go`:
```go
// Package registry manages the project registry and discovery state.
package registry
```

Create `internal/ui/ui.go`:
```go
// Package ui contains all Bubble Tea components for the TUI.
package ui
```

Create `internal/compose/client.go`:
```go
// Package compose provides a client for the process-compose REST API.
package compose
```

Create `internal/scanner/scanner.go`:
```go
// Package scanner discovers devenv.nix projects on the filesystem.
package scanner
```

**Step 3: Verify structure**

Run: `find internal -name "*.go"`
Expected: Lists all 5 files

**Step 4: Commit**

```bash
git add internal/
git commit -m "chore: create project directory structure"
```

---

### Task 1.4: Basic Bubble Tea App Shell

**Files:**
- Modify: `main.go`
- Create: `internal/ui/model.go`
- Test: `internal/ui/model_test.go`

**Step 1: Write the failing test**

Create `internal/ui/model_test.go`:
```go
package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestModelImplementsTeaModel(t *testing.T) {
	var m tea.Model = New()
	if m == nil {
		t.Fatal("New() returned nil")
	}
}

func TestModelInit(t *testing.T) {
	m := New()
	cmd := m.Init()
	// Init should return nil or a valid command
	_ = cmd
}

func TestModelView(t *testing.T) {
	m := New()
	view := m.View()
	if view == "" {
		t.Fatal("View() returned empty string")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/ui/... -v`
Expected: FAIL - `New` undefined

**Step 3: Write minimal implementation**

Create `internal/ui/model.go`:
```go
package ui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Model is the main application model for acidBurn.
type Model struct {
	width  int
	height int
}

// New creates a new acidBurn model.
func New() *Model {
	return &Model{}
}

// Init initializes the model.
func (m *Model) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

// View renders the model.
func (m *Model) View() string {
	return "acidBurn - Press q to quit"
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/ui/... -v`
Expected: PASS

**Step 5: Wire up main.go**

Update `main.go`:
```go
package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/infktd/acidburn/internal/ui"
)

func main() {
	p := tea.NewProgram(ui.New(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
```

**Step 6: Build and run manually**

Run: `go build -o acidburn . && ./acidburn`
Expected: Shows "acidBurn - Press q to quit", exits on 'q'

**Step 7: Commit**

```bash
git add main.go internal/ui/
git commit -m "feat: add basic bubble tea app shell"
```

---

## Phase 2: Configuration System

### Task 2.1: Config Struct Definition

**Files:**
- Create: `internal/config/types.go`
- Test: `internal/config/types_test.go`

**Step 1: Write the failing test**

Create `internal/config/types_test.go`:
```go
package config

import (
	"testing"
)

func TestDefaultConfigHasScanPaths(t *testing.T) {
	cfg := Default()
	if len(cfg.Projects.ScanPaths) == 0 {
		t.Fatal("Default config should have at least one scan path")
	}
}

func TestDefaultConfigHasTheme(t *testing.T) {
	cfg := Default()
	if cfg.UI.Theme == "" {
		t.Fatal("Default config should have a theme")
	}
}

func TestDefaultConfigHasPollingIntervals(t *testing.T) {
	cfg := Default()
	if cfg.Polling.FocusedProject == 0 {
		t.Fatal("Default config should have focused project polling interval")
	}
	if cfg.Polling.BackgroundProject == 0 {
		t.Fatal("Default config should have background project polling interval")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/config/... -v`
Expected: FAIL - `Default` undefined

**Step 3: Write minimal implementation**

Create `internal/config/types.go`:
```go
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
	SystemEnabled bool                       `yaml:"system_enabled"`
	TUIAlerts     bool                       `yaml:"tui_alerts"`
	CriticalOnly  bool                       `yaml:"critical_only"`
	Overrides     []NotificationOverride     `yaml:"overrides,omitempty"`
}

// NotificationOverride allows per-service notification settings.
type NotificationOverride struct {
	Service      string `yaml:"service"`
	System       bool   `yaml:"system"`
	CriticalOnly bool   `yaml:"critical_only,omitempty"`
}

// UIConfig configures the user interface.
type UIConfig struct {
	Theme         string `yaml:"theme"`
	DefaultLogView string `yaml:"default_log_view"`
	LogFollow     bool   `yaml:"log_follow"`
	ShowTimestamps bool   `yaml:"show_timestamps"`
	DimTimestamps bool   `yaml:"dim_timestamps"`
	SidebarWidth  int    `yaml:"sidebar_width"`
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
			Theme:         "acid-green",
			DefaultLogView: "focused",
			LogFollow:     true,
			ShowTimestamps: true,
			DimTimestamps: true,
			SidebarWidth:  25,
		},
		Polling: PollingConfig{
			FocusedProject:    2,
			BackgroundProject: 10,
		},
	}
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/config/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/config/
git commit -m "feat: add config struct with defaults"
```

---

### Task 2.2: Config Loading and Saving

**Files:**
- Modify: `internal/config/config.go`
- Test: `internal/config/config_test.go`

**Step 1: Write the failing test**

Create `internal/config/config_test.go`:
```go
package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadCreatesDefaultIfMissing(t *testing.T) {
	// Use temp dir
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg == nil {
		t.Fatal("Load() returned nil config")
	}
	if cfg.UI.Theme != "acid-green" {
		t.Errorf("Expected theme 'acid-green', got %q", cfg.UI.Theme)
	}
}

func TestSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Create and modify config
	cfg := Default()
	cfg.UI.Theme = "nord"

	// Save it
	if err := Save(configPath, cfg); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	// Load it back
	loaded, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if loaded.UI.Theme != "nord" {
		t.Errorf("Expected theme 'nord', got %q", loaded.UI.Theme)
	}
}

func TestConfigPath(t *testing.T) {
	path := Path()
	if path == "" {
		t.Fatal("Path() returned empty string")
	}
	if !filepath.IsAbs(path) {
		t.Errorf("Path() should return absolute path, got %q", path)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/config/... -v`
Expected: FAIL - `Load`, `Save`, `Path` undefined

**Step 3: Write minimal implementation**

Update `internal/config/config.go`:
```go
// Package config handles acidBurn configuration loading and persistence.
package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	configDir  = "acidburn"
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
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/config/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/config/
git commit -m "feat: add config load/save with YAML persistence"
```

---

## Phase 3: Project Discovery & Registry

### Task 3.1: Project Scanner

**Files:**
- Modify: `internal/scanner/scanner.go`
- Test: `internal/scanner/scanner_test.go`

**Step 1: Write the failing test**

Create `internal/scanner/scanner_test.go`:
```go
package scanner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanFindsDevenvNix(t *testing.T) {
	// Create temp directory structure
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "my-project")
	os.MkdirAll(projectDir, 0755)
	os.WriteFile(filepath.Join(projectDir, "devenv.nix"), []byte("{}"), 0644)

	// Scan
	projects, err := Scan([]string{tmpDir}, 3)
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}
	if len(projects) != 1 {
		t.Fatalf("Expected 1 project, got %d", len(projects))
	}
	if projects[0] != projectDir {
		t.Errorf("Expected %q, got %q", projectDir, projects[0])
	}
}

func TestScanRespectsDepthLimit(t *testing.T) {
	tmpDir := t.TempDir()
	// Create project at depth 4
	deepProject := filepath.Join(tmpDir, "a", "b", "c", "d", "project")
	os.MkdirAll(deepProject, 0755)
	os.WriteFile(filepath.Join(deepProject, "devenv.nix"), []byte("{}"), 0644)

	// Scan with depth 3 should not find it
	projects, _ := Scan([]string{tmpDir}, 3)
	if len(projects) != 0 {
		t.Errorf("Expected 0 projects at depth 3, got %d", len(projects))
	}

	// Scan with depth 5 should find it
	projects, _ = Scan([]string{tmpDir}, 5)
	if len(projects) != 1 {
		t.Errorf("Expected 1 project at depth 5, got %d", len(projects))
	}
}

func TestScanSkipsExcludedDirs(t *testing.T) {
	tmpDir := t.TempDir()
	// Create project inside node_modules (should be skipped)
	excluded := filepath.Join(tmpDir, "node_modules", "some-pkg")
	os.MkdirAll(excluded, 0755)
	os.WriteFile(filepath.Join(excluded, "devenv.nix"), []byte("{}"), 0644)

	projects, _ := Scan([]string{tmpDir}, 3)
	if len(projects) != 0 {
		t.Errorf("Expected 0 projects (node_modules excluded), got %d", len(projects))
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/scanner/... -v`
Expected: FAIL - `Scan` undefined

**Step 3: Write minimal implementation**

Update `internal/scanner/scanner.go`:
```go
// Package scanner discovers devenv.nix projects on the filesystem.
package scanner

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Directories to skip during scanning.
var excludedDirs = map[string]bool{
	"node_modules": true,
	".git":         true,
	".direnv":      true,
	"dist":         true,
	"target":       true,
	"vendor":       true,
	".venv":        true,
	"__pycache__":  true,
}

// Scan searches paths for directories containing devenv.nix.
// maxDepth limits how deep to recurse (1 = immediate children only).
func Scan(paths []string, maxDepth int) ([]string, error) {
	var projects []string
	seen := make(map[string]bool)

	for _, root := range paths {
		// Expand ~ if present
		if strings.HasPrefix(root, "~/") {
			home, _ := os.UserHomeDir()
			root = filepath.Join(home, root[2:])
		}

		err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil // Skip inaccessible paths
			}

			// Calculate depth relative to root
			rel, _ := filepath.Rel(root, path)
			depth := len(strings.Split(rel, string(os.PathSeparator)))
			if rel == "." {
				depth = 0
			}

			// Skip if too deep
			if depth > maxDepth {
				if d.IsDir() {
					return fs.SkipDir
				}
				return nil
			}

			// Skip excluded directories
			if d.IsDir() && excludedDirs[d.Name()] {
				return fs.SkipDir
			}

			// Check for devenv.nix
			if !d.IsDir() && d.Name() == "devenv.nix" {
				projectPath := filepath.Dir(path)
				if !seen[projectPath] {
					seen[projectPath] = true
					projects = append(projects, projectPath)
				}
			}

			return nil
		})
		if err != nil {
			return projects, err
		}
	}

	return projects, nil
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/scanner/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/scanner/
git commit -m "feat: add project scanner for devenv.nix discovery"
```

---

### Task 3.2: Registry Types and State Detection

**Files:**
- Modify: `internal/registry/registry.go`
- Create: `internal/registry/types.go`
- Test: `internal/registry/registry_test.go`

**Step 1: Write the failing test**

Create `internal/registry/registry_test.go`:
```go
package registry

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewProjectGeneratesID(t *testing.T) {
	p := NewProject("/some/path")
	if p.ID == "" {
		t.Fatal("NewProject should generate an ID")
	}
	if p.Path != "/some/path" {
		t.Errorf("Expected path '/some/path', got %q", p.Path)
	}
	if p.Name != "path" {
		t.Errorf("Expected name 'path', got %q", p.Name)
	}
}

func TestProjectStateDetection(t *testing.T) {
	tmpDir := t.TempDir()

	// No socket = Idle
	p := NewProject(tmpDir)
	if p.DetectState() != StateIdle {
		t.Errorf("Expected StateIdle, got %v", p.DetectState())
	}

	// Create socket dir but no socket = still Idle
	socketDir := filepath.Join(tmpDir, ".devenv", "state", "process-compose")
	os.MkdirAll(socketDir, 0755)
	if p.DetectState() != StateIdle {
		t.Errorf("Expected StateIdle, got %v", p.DetectState())
	}
}

func TestProjectStateMissing(t *testing.T) {
	p := NewProject("/nonexistent/path/that/does/not/exist")
	if p.DetectState() != StateMissing {
		t.Errorf("Expected StateMissing, got %v", p.DetectState())
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/registry/... -v`
Expected: FAIL - types undefined

**Step 3: Write minimal implementation**

Create `internal/registry/types.go`:
```go
package registry

import (
	"crypto/sha256"
	"encoding/hex"
	"net"
	"os"
	"path/filepath"
	"time"
)

// ProjectState represents the current state of a project.
type ProjectState int

const (
	StateIdle ProjectState = iota
	StateRunning
	StateDegraded
	StateStale
	StateMissing
)

func (s ProjectState) String() string {
	switch s {
	case StateIdle:
		return "idle"
	case StateRunning:
		return "running"
	case StateDegraded:
		return "degraded"
	case StateStale:
		return "stale"
	case StateMissing:
		return "missing"
	default:
		return "unknown"
	}
}

// Project represents a devenv project in the registry.
type Project struct {
	ID         string    `yaml:"id"`
	Path       string    `yaml:"path"`
	Name       string    `yaml:"name"`
	Hidden     bool      `yaml:"hidden"`
	LastActive time.Time `yaml:"last_active"`
}

// NewProject creates a new Project from a path.
func NewProject(path string) *Project {
	// Generate ID from path hash
	hash := sha256.Sum256([]byte(path))
	id := hex.EncodeToString(hash[:8])

	return &Project{
		ID:         id,
		Path:       path,
		Name:       filepath.Base(path),
		Hidden:     false,
		LastActive: time.Now(),
	}
}

// SocketPath returns the path to the process-compose socket.
func (p *Project) SocketPath() string {
	return filepath.Join(p.Path, ".devenv", "state", "process-compose", "pc.sock")
}

// DetectState checks the project's current state.
func (p *Project) DetectState() ProjectState {
	// Check if path exists
	if _, err := os.Stat(p.Path); os.IsNotExist(err) {
		return StateMissing
	}

	socketPath := p.SocketPath()

	// Try to connect to socket
	conn, err := net.Dial("unix", socketPath)
	if err == nil {
		conn.Close()
		return StateRunning // TODO: Check if degraded via API
	}

	// Check if socket file exists (stale)
	if _, err := os.Stat(socketPath); err == nil {
		return StateStale
	}

	return StateIdle
}
```

Update `internal/registry/registry.go`:
```go
// Package registry manages the project registry and discovery state.
package registry

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	registryDir  = "acidburn"
	registryFile = "projects.yaml"
)

// Registry holds discovered projects.
type Registry struct {
	Projects []*Project `yaml:"projects"`
}

// Path returns the default registry file path.
func Path() string {
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		home, _ := os.UserHomeDir()
		configHome = filepath.Join(home, ".config")
	}
	return filepath.Join(configHome, registryDir, registryFile)
}

// Load reads the registry from path.
func Load(path string) (*Registry, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &Registry{Projects: []*Project{}}, nil
	}
	if err != nil {
		return nil, err
	}

	var reg Registry
	if err := yaml.Unmarshal(data, &reg); err != nil {
		return nil, err
	}
	return &reg, nil
}

// Save writes the registry to path.
func Save(path string, reg *Registry) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := yaml.Marshal(reg)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// AddProject adds a project if not already present.
func (r *Registry) AddProject(path string) *Project {
	for _, p := range r.Projects {
		if p.Path == path {
			return p
		}
	}
	p := NewProject(path)
	r.Projects = append(r.Projects, p)
	return p
}

// FindByPath returns a project by its path.
func (r *Registry) FindByPath(path string) *Project {
	for _, p := range r.Projects {
		if p.Path == path {
			return p
		}
	}
	return nil
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/registry/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/registry/
git commit -m "feat: add project registry with state detection"
```

---

## Phase 4: UI - Basic Layout

### Task 4.1: Theme System

**Files:**
- Create: `internal/ui/theme.go`
- Test: `internal/ui/theme_test.go`

**Step 1: Write the failing test**

Create `internal/ui/theme_test.go`:
```go
package ui

import (
	"testing"
)

func TestGetThemeReturnsDefault(t *testing.T) {
	theme := GetTheme("acid-green")
	if theme.Primary == "" {
		t.Fatal("Theme should have a primary color")
	}
}

func TestGetThemeFallsBackToDefault(t *testing.T) {
	theme := GetTheme("nonexistent-theme")
	if theme.Primary == "" {
		t.Fatal("Unknown theme should fall back to default")
	}
}

func TestAllThemesHaveRequiredColors(t *testing.T) {
	for name, theme := range Themes {
		if theme.Primary == "" {
			t.Errorf("Theme %q missing Primary", name)
		}
		if theme.Secondary == "" {
			t.Errorf("Theme %q missing Secondary", name)
		}
		if theme.Background == "" {
			t.Errorf("Theme %q missing Background", name)
		}
		if theme.Muted == "" {
			t.Errorf("Theme %q missing Muted", name)
		}
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/ui/... -v -run Theme`
Expected: FAIL - `GetTheme`, `Themes` undefined

**Step 3: Write minimal implementation**

Create `internal/ui/theme.go`:
```go
package ui

import "github.com/charmbracelet/lipgloss"

// Theme defines the color palette for the UI.
type Theme struct {
	Name       string
	Primary    lipgloss.Color
	Secondary  lipgloss.Color
	Background lipgloss.Color
	Muted      lipgloss.Color
	Success    lipgloss.Color
	Warning    lipgloss.Color
	Error      lipgloss.Color
}

// Themes contains all available themes.
var Themes = map[string]Theme{
	"acid-green": {
		Name:       "acid-green",
		Primary:    lipgloss.Color("#39FF14"),
		Secondary:  lipgloss.Color("#00FF41"),
		Background: lipgloss.Color("#0D0D0D"),
		Muted:      lipgloss.Color("#4A4A4A"),
		Success:    lipgloss.Color("#00FF41"),
		Warning:    lipgloss.Color("#FFD700"),
		Error:      lipgloss.Color("#FF4136"),
	},
	"nord": {
		Name:       "nord",
		Primary:    lipgloss.Color("#88C0D0"),
		Secondary:  lipgloss.Color("#81A1C1"),
		Background: lipgloss.Color("#2E3440"),
		Muted:      lipgloss.Color("#4C566A"),
		Success:    lipgloss.Color("#A3BE8C"),
		Warning:    lipgloss.Color("#EBCB8B"),
		Error:      lipgloss.Color("#BF616A"),
	},
	"dracula": {
		Name:       "dracula",
		Primary:    lipgloss.Color("#BD93F9"),
		Secondary:  lipgloss.Color("#FF79C6"),
		Background: lipgloss.Color("#282A36"),
		Muted:      lipgloss.Color("#6272A4"),
		Success:    lipgloss.Color("#50FA7B"),
		Warning:    lipgloss.Color("#F1FA8C"),
		Error:      lipgloss.Color("#FF5555"),
	},
}

// GetTheme returns a theme by name, defaulting to acid-green.
func GetTheme(name string) Theme {
	if theme, ok := Themes[name]; ok {
		return theme
	}
	return Themes["acid-green"]
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/ui/... -v -run Theme`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/ui/theme.go internal/ui/theme_test.go
git commit -m "feat: add theme system with acid-green, nord, dracula"
```

---

### Task 4.2: Styles Definition

**Files:**
- Create: `internal/ui/styles.go`

**Step 1: Create styles**

Create `internal/ui/styles.go`:
```go
package ui

import "github.com/charmbracelet/lipgloss"

// Styles holds all the styled components.
type Styles struct {
	// Layout
	Header  lipgloss.Style
	Footer  lipgloss.Style
	Sidebar lipgloss.Style
	Main    lipgloss.Style

	// Components
	Title         lipgloss.Style
	Breadcrumb    lipgloss.Style
	StatusBar     lipgloss.Style
	ProjectItem   lipgloss.Style
	SelectedItem  lipgloss.Style
	ServiceRow    lipgloss.Style
	LogLine       lipgloss.Style
	LogTimestamp  lipgloss.Style
	LogLevelInfo  lipgloss.Style
	LogLevelWarn  lipgloss.Style
	LogLevelError lipgloss.Style

	// Status indicators
	StatusRunning  lipgloss.Style
	StatusIdle     lipgloss.Style
	StatusDegraded lipgloss.Style
	StatusStale    lipgloss.Style
	StatusMissing  lipgloss.Style

	// Borders
	FocusedBorder lipgloss.Style
	BlurredBorder lipgloss.Style
}

// NewStyles creates styles from a theme.
func NewStyles(theme Theme) *Styles {
	return &Styles{
		// Layout
		Header: lipgloss.NewStyle().
			Foreground(theme.Primary).
			Bold(true).
			Padding(0, 1),

		Footer: lipgloss.NewStyle().
			Foreground(theme.Muted).
			Padding(0, 1),

		Sidebar: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(theme.Muted).
			Padding(0, 1),

		Main: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(theme.Muted).
			Padding(0, 1),

		// Components
		Title: lipgloss.NewStyle().
			Foreground(theme.Primary).
			Bold(true),

		Breadcrumb: lipgloss.NewStyle().
			Foreground(theme.Muted),

		StatusBar: lipgloss.NewStyle().
			Foreground(theme.Secondary),

		ProjectItem: lipgloss.NewStyle().
			Foreground(theme.Secondary),

		SelectedItem: lipgloss.NewStyle().
			Foreground(theme.Primary).
			Bold(true),

		ServiceRow: lipgloss.NewStyle().
			Foreground(theme.Secondary),

		LogLine: lipgloss.NewStyle().
			Foreground(theme.Secondary),

		LogTimestamp: lipgloss.NewStyle().
			Foreground(theme.Muted),

		LogLevelInfo: lipgloss.NewStyle().
			Foreground(theme.Muted),

		LogLevelWarn: lipgloss.NewStyle().
			Foreground(theme.Warning),

		LogLevelError: lipgloss.NewStyle().
			Foreground(theme.Error).
			Bold(true),

		// Status indicators
		StatusRunning: lipgloss.NewStyle().
			Foreground(theme.Success),

		StatusIdle: lipgloss.NewStyle().
			Foreground(theme.Muted),

		StatusDegraded: lipgloss.NewStyle().
			Foreground(theme.Warning),

		StatusStale: lipgloss.NewStyle().
			Foreground(theme.Error),

		StatusMissing: lipgloss.NewStyle().
			Foreground(theme.Error),

		// Borders
		FocusedBorder: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(theme.Primary),

		BlurredBorder: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(theme.Muted),
	}
}
```

**Step 2: Verify it compiles**

Run: `go build ./internal/ui/...`
Expected: No errors

**Step 3: Commit**

```bash
git add internal/ui/styles.go
git commit -m "feat: add comprehensive style definitions"
```

---

### Task 4.3: KeyMap Definition

**Files:**
- Create: `internal/ui/keys.go`
- Test: `internal/ui/keys_test.go`

**Step 1: Write the failing test**

Create `internal/ui/keys_test.go`:
```go
package ui

import (
	"testing"
)

func TestDefaultKeyMapHasQuit(t *testing.T) {
	km := DefaultKeyMap()
	if len(km.Quit.Keys()) == 0 {
		t.Fatal("KeyMap should have quit key")
	}
}

func TestDefaultKeyMapHasNavigation(t *testing.T) {
	km := DefaultKeyMap()
	if len(km.Up.Keys()) == 0 {
		t.Fatal("KeyMap should have up key")
	}
	if len(km.Down.Keys()) == 0 {
		t.Fatal("KeyMap should have down key")
	}
}

func TestKeyMapShortHelp(t *testing.T) {
	km := DefaultKeyMap()
	help := km.ShortHelp()
	if len(help) == 0 {
		t.Fatal("ShortHelp should return bindings")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/ui/... -v -run KeyMap`
Expected: FAIL - `DefaultKeyMap` undefined

**Step 3: Write minimal implementation**

Create `internal/ui/keys.go`:
```go
package ui

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines all keybindings.
type KeyMap struct {
	// Global
	Quit     key.Binding
	Shutdown key.Binding
	Settings key.Binding
	Help     key.Binding
	Refresh  key.Binding
	History  key.Binding

	// Navigation
	Up       key.Binding
	Down     key.Binding
	Left     key.Binding
	Right    key.Binding
	Tab      key.Binding
	Select   key.Binding
	Back     key.Binding

	// Actions
	Start   key.Binding
	Stop    key.Binding
	Restart key.Binding
	Search  key.Binding

	// Logs
	Follow  key.Binding
	Filter  key.Binding
	Wrap    key.Binding
	Yank    key.Binding
	Top     key.Binding
	Bottom  key.Binding
}

// DefaultKeyMap returns the default keybindings.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		// Global
		Quit: key.NewBinding(
			key.WithKeys("q"),
			key.WithHelp("q", "quit"),
		),
		Shutdown: key.NewBinding(
			key.WithKeys("ctrl+x"),
			key.WithHelp("ctrl+x", "shutdown all"),
		),
		Settings: key.NewBinding(
			key.WithKeys("S"),
			key.WithHelp("S", "settings"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("R"),
			key.WithHelp("R", "refresh"),
		),
		History: key.NewBinding(
			key.WithKeys("H"),
			key.WithHelp("H", "alert history"),
		),

		// Navigation
		Up: key.NewBinding(
			key.WithKeys("k", "up"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("j", "down"),
			key.WithHelp("↓/j", "down"),
		),
		Left: key.NewBinding(
			key.WithKeys("h", "left"),
			key.WithHelp("←/h", "left"),
		),
		Right: key.NewBinding(
			key.WithKeys("l", "right"),
			key.WithHelp("→/l", "right"),
		),
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "switch pane"),
		),
		Select: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
		),

		// Actions
		Start: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "start"),
		),
		Stop: key.NewBinding(
			key.WithKeys("x"),
			key.WithHelp("x", "stop"),
		),
		Restart: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "restart"),
		),
		Search: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "search"),
		),

		// Logs
		Follow: key.NewBinding(
			key.WithKeys("f"),
			key.WithHelp("f", "follow"),
		),
		Filter: key.NewBinding(
			key.WithKeys("ctrl+f"),
			key.WithHelp("ctrl+f", "filter"),
		),
		Wrap: key.NewBinding(
			key.WithKeys("w"),
			key.WithHelp("w", "wrap"),
		),
		Yank: key.NewBinding(
			key.WithKeys("y"),
			key.WithHelp("y", "yank"),
		),
		Top: key.NewBinding(
			key.WithKeys("g"),
			key.WithHelp("g", "top"),
		),
		Bottom: key.NewBinding(
			key.WithKeys("G"),
			key.WithHelp("G", "bottom"),
		),
	}
}

// ShortHelp returns the short help bindings.
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Select, k.Start, k.Stop, k.Help, k.Quit}
}

// FullHelp returns the full help bindings grouped by category.
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Tab, k.Select, k.Back},
		{k.Start, k.Stop, k.Restart, k.Search},
		{k.Follow, k.Top, k.Bottom, k.Wrap},
		{k.Settings, k.History, k.Help, k.Quit},
	}
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/ui/... -v -run KeyMap`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/ui/keys.go internal/ui/keys_test.go
git commit -m "feat: add comprehensive keymap with vim/arrow hybrid"
```

---

### Task 4.4: Enhanced Model with Config and Registry

**Files:**
- Modify: `internal/ui/model.go`
- Modify: `internal/ui/model_test.go`

**Step 1: Update tests**

Update `internal/ui/model_test.go`:
```go
package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/infktd/acidburn/internal/config"
	"github.com/infktd/acidburn/internal/registry"
)

func TestModelImplementsTeaModel(t *testing.T) {
	cfg := config.Default()
	reg := &registry.Registry{}
	var m tea.Model = New(cfg, reg)
	if m == nil {
		t.Fatal("New() returned nil")
	}
}

func TestModelInit(t *testing.T) {
	cfg := config.Default()
	reg := &registry.Registry{}
	m := New(cfg, reg)
	cmd := m.Init()
	_ = cmd
}

func TestModelView(t *testing.T) {
	cfg := config.Default()
	reg := &registry.Registry{}
	m := New(cfg, reg)
	view := m.View()
	if view == "" {
		t.Fatal("View() returned empty string")
	}
}

func TestModelHandlesWindowSize(t *testing.T) {
	cfg := config.Default()
	reg := &registry.Registry{}
	m := New(cfg, reg)

	msg := tea.WindowSizeMsg{Width: 100, Height: 50}
	newModel, _ := m.Update(msg)
	model := newModel.(*Model)

	if model.width != 100 {
		t.Errorf("Expected width 100, got %d", model.width)
	}
	if model.height != 50 {
		t.Errorf("Expected height 50, got %d", model.height)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/ui/... -v`
Expected: FAIL - signature mismatch

**Step 3: Update implementation**

Update `internal/ui/model.go`:
```go
package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/infktd/acidburn/internal/config"
	"github.com/infktd/acidburn/internal/registry"
)

// FocusedPane tracks which pane has focus.
type FocusedPane int

const (
	PaneSidebar FocusedPane = iota
	PaneServices
	PaneLogs
)

// Model is the main application model for acidBurn.
type Model struct {
	config   *config.Config
	registry *registry.Registry
	styles   *Styles
	keys     KeyMap

	width  int
	height int

	focused       FocusedPane
	selectedProject int
	selectedService int

	showHelp     bool
	showSettings bool
}

// New creates a new acidBurn model.
func New(cfg *config.Config, reg *registry.Registry) *Model {
	theme := GetTheme(cfg.UI.Theme)
	return &Model{
		config:   cfg,
		registry: reg,
		styles:   NewStyles(theme),
		keys:     DefaultKeyMap(),
		focused:  PaneSidebar,
	}
}

// Init initializes the model.
func (m *Model) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

func (m *Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit
	case key.Matches(msg, m.keys.Help):
		m.showHelp = !m.showHelp
	case key.Matches(msg, m.keys.Tab):
		m.cycleFocus()
	case key.Matches(msg, m.keys.Up):
		m.moveUp()
	case key.Matches(msg, m.keys.Down):
		m.moveDown()
	}
	return m, nil
}

func (m *Model) cycleFocus() {
	m.focused = (m.focused + 1) % 3
}

func (m *Model) moveUp() {
	switch m.focused {
	case PaneSidebar:
		if m.selectedProject > 0 {
			m.selectedProject--
		}
	case PaneServices:
		if m.selectedService > 0 {
			m.selectedService--
		}
	}
}

func (m *Model) moveDown() {
	switch m.focused {
	case PaneSidebar:
		if m.selectedProject < len(m.registry.Projects)-1 {
			m.selectedProject++
		}
	case PaneServices:
		m.selectedService++
	}
}

// View renders the model.
func (m *Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	if m.showHelp {
		return m.renderHelp()
	}

	header := m.renderHeader()
	body := m.renderBody()
	footer := m.renderFooter()

	return lipgloss.JoinVertical(lipgloss.Left, header, body, footer)
}

func (m *Model) renderHeader() string {
	title := m.styles.Title.Render("acidBurn")
	breadcrumb := m.styles.Breadcrumb.Render(" ─── FLEET")

	stats := fmt.Sprintf("PIDs: %d ── MEM: -- ── Nix: OK", len(m.registry.Projects))
	statsView := m.styles.StatusBar.Render(stats)

	left := title + breadcrumb
	right := statsView

	gap := m.width - lipgloss.Width(left) - lipgloss.Width(right) - 2
	if gap < 0 {
		gap = 0
	}

	return m.styles.Header.Width(m.width).Render(
		left + fmt.Sprintf("%*s", gap, "") + right,
	)
}

func (m *Model) renderBody() string {
	sidebarWidth := m.config.UI.SidebarWidth
	mainWidth := m.width - sidebarWidth - 4
	bodyHeight := m.height - 4 // header + footer

	sidebar := m.renderSidebar(sidebarWidth, bodyHeight)
	main := m.renderMain(mainWidth, bodyHeight)

	return lipgloss.JoinHorizontal(lipgloss.Top, sidebar, main)
}

func (m *Model) renderSidebar(width, height int) string {
	var content string

	content += m.styles.Title.Render("PROJECTS") + "\n"
	content += m.styles.Muted().Render("/ search...") + "\n\n"

	// Group by state
	content += m.styles.Muted().Render("── ACTIVE ──") + "\n"
	for i, p := range m.registry.Projects {
		state := p.DetectState()
		if state == registry.StateRunning || state == registry.StateDegraded {
			content += m.renderProjectItem(i, p) + "\n"
		}
	}

	content += "\n" + m.styles.Muted().Render("── IDLE ──") + "\n"
	for i, p := range m.registry.Projects {
		state := p.DetectState()
		if state == registry.StateIdle || state == registry.StateStale || state == registry.StateMissing {
			content += m.renderProjectItem(i, p) + "\n"
		}
	}

	content += "\n" + m.styles.Muted().Render("── GLOBAL ──") + "\n"
	content += m.styles.ProjectItem.Render("▤ ALL LOGS") + "\n"

	style := m.styles.BlurredBorder
	if m.focused == PaneSidebar {
		style = m.styles.FocusedBorder
	}

	return style.Width(width).Height(height).Render(content)
}

func (m *Model) renderProjectItem(idx int, p *registry.Project) string {
	state := p.DetectState()
	var glyph string
	switch state {
	case registry.StateRunning:
		glyph = m.styles.StatusRunning.Render("●")
	case registry.StateDegraded:
		glyph = m.styles.StatusDegraded.Render("◐")
	case registry.StateIdle:
		glyph = m.styles.StatusIdle.Render("○")
	case registry.StateStale:
		glyph = m.styles.StatusStale.Render("✗")
	case registry.StateMissing:
		glyph = m.styles.StatusMissing.Render("✗")
	}

	name := p.Name
	if idx == m.selectedProject && m.focused == PaneSidebar {
		name = m.styles.SelectedItem.Render(name)
	} else {
		name = m.styles.ProjectItem.Render(name)
	}

	return fmt.Sprintf("%s %s", glyph, name)
}

func (m *Model) renderMain(width, height int) string {
	servicesHeight := height / 3
	logsHeight := height - servicesHeight - 2

	services := m.renderServices(width, servicesHeight)
	logs := m.renderLogs(width, logsHeight)

	return lipgloss.JoinVertical(lipgloss.Left, services, logs)
}

func (m *Model) renderServices(width, height int) string {
	content := m.styles.Title.Render("SERVICES") + "\n"
	content += m.styles.Muted().Render("No project selected") + "\n"

	style := m.styles.BlurredBorder
	if m.focused == PaneServices {
		style = m.styles.FocusedBorder
	}

	return style.Width(width).Height(height).Render(content)
}

func (m *Model) renderLogs(width, height int) string {
	content := m.styles.Title.Render("LOGS") + "\n"
	content += m.styles.Muted().Render("No logs to display") + "\n"

	style := m.styles.BlurredBorder
	if m.focused == PaneLogs {
		style = m.styles.FocusedBorder
	}

	return style.Width(width).Height(height).Render(content)
}

func (m *Model) renderFooter() string {
	var help string
	switch m.focused {
	case PaneSidebar:
		help = "[↑/↓] Navigate  [Enter] Select  [s] Start  [x] Stop  [/] Search  [?] Help"
	case PaneServices:
		help = "[↑/↓] Navigate  [r] Restart  [x] Stop  [s] Start  [Enter] View Logs  [?] Help"
	case PaneLogs:
		help = "[↑/↓] Scroll  [f] Follow  [/] Search  [Ctrl+f] Filter  [g/G] Top/Bottom  [?] Help"
	}
	return m.styles.Footer.Width(m.width).Render(help)
}

func (m *Model) renderHelp() string {
	help := `
KEYBINDINGS

  GLOBAL                     NAVIGATION
  q       Quit (detach)      ↑/k     Up
  Ctrl+X  Shutdown all       ↓/j     Down
  S       Settings           Tab     Switch pane
  H       Alert history      Enter   Select/Confirm
  R       Refresh            Esc     Back/Cancel
  ?       This help

  SIDEBAR                    SERVICES
  s       Start project      s       Start service
  x       Stop project       x       Stop service
  /       Search projects    r       Restart service

  LOGS
  f       Toggle follow      g/G     Top/Bottom
  /       Search logs        PgUp    Page up
  Ctrl+f  Filter mode        PgDn    Page down
  w       Toggle wrap        y       Yank line

                                        [Esc] Close
`
	return m.styles.Main.Width(m.width).Height(m.height).Render(help)
}

// Muted returns the muted style (helper for rendering).
func (s *Styles) Muted() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#4A4A4A"))
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/ui/... -v`
Expected: PASS

**Step 5: Update main.go**

Update `main.go`:
```go
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
		registry.Save(registry.Path(), reg)
	}

	// Run TUI
	p := tea.NewProgram(ui.New(cfg, reg), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
```

**Step 6: Build and test manually**

Run: `go build -o acidburn . && ./acidburn`
Expected: Shows the cockpit UI with sidebar, services, logs panes

**Step 7: Commit**

```bash
git add main.go internal/ui/
git commit -m "feat: implement 3-pane cockpit layout with themes and keybindings"
```

---

## Checkpoint: Phase 1-4 Complete

At this point you have:
- Go project structure
- Config system with YAML persistence
- Project scanner and registry
- Basic 3-pane TUI layout with themes
- Keybindings and help overlay

**Run full test suite:**
```bash
go test ./... -v
```

**Manual verification:**
```bash
go build -o acidburn . && ./acidburn
```

---

## Phase 5: process-compose Integration

### Task 5.1: process-compose Client

**Files:**
- Modify: `internal/compose/client.go`
- Create: `internal/compose/types.go`
- Test: `internal/compose/client_test.go`

**Step 1: Write the failing test**

Create `internal/compose/client_test.go`:
```go
package compose

import (
	"testing"
)

func TestNewClient(t *testing.T) {
	client := NewClient("/tmp/test.sock")
	if client == nil {
		t.Fatal("NewClient returned nil")
	}
	if client.socketPath != "/tmp/test.sock" {
		t.Errorf("Expected socket path '/tmp/test.sock', got %q", client.socketPath)
	}
}

func TestClientIsConnectedReturnsFalseWhenNotConnected(t *testing.T) {
	client := NewClient("/nonexistent/socket.sock")
	if client.IsConnected() {
		t.Fatal("IsConnected should return false for nonexistent socket")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/compose/... -v`
Expected: FAIL - `NewClient` undefined

**Step 3: Write implementation**

Create `internal/compose/types.go`:
```go
package compose

// ProcessStatus represents the status of a process in process-compose.
type ProcessStatus struct {
	Name       string `json:"name"`
	Status     string `json:"status"`
	IsRunning  bool   `json:"is_running"`
	Pid        int    `json:"pid"`
	ExitCode   int    `json:"exit_code"`
	SystemTime string `json:"system_time"`
}

// ProjectStatus represents the overall project status.
type ProjectStatus struct {
	Processes []ProcessStatus `json:"processes"`
}
```

Update `internal/compose/client.go`:
```go
// Package compose provides a client for the process-compose REST API.
package compose

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"
)

// Client communicates with process-compose via Unix socket.
type Client struct {
	socketPath string
	httpClient *http.Client
	connected  bool
}

// NewClient creates a new process-compose client.
func NewClient(socketPath string) *Client {
	return &Client{
		socketPath: socketPath,
		httpClient: &http.Client{
			Transport: &http.Transport{
				DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
					return net.Dial("unix", socketPath)
				},
			},
			Timeout: 5 * time.Second,
		},
	}
}

// Connect attempts to connect to the process-compose socket.
func (c *Client) Connect() error {
	conn, err := net.Dial("unix", c.socketPath)
	if err != nil {
		c.connected = false
		return err
	}
	conn.Close()
	c.connected = true
	return nil
}

// IsConnected returns true if connected to process-compose.
func (c *Client) IsConnected() bool {
	return c.connected
}

// GetStatus fetches the current process status.
func (c *Client) GetStatus() (*ProjectStatus, error) {
	resp, err := c.httpClient.Get("http://unix/processes")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var status ProjectStatus
	if err := json.NewDecoder(resp.Body).Decode(&status.Processes); err != nil {
		return nil, err
	}

	return &status, nil
}

// StartProcess starts a specific process.
func (c *Client) StartProcess(name string) error {
	url := fmt.Sprintf("http://unix/process/%s/start", name)
	resp, err := c.httpClient.Post(url, "application/json", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to start process: %d", resp.StatusCode)
	}
	return nil
}

// StopProcess stops a specific process.
func (c *Client) StopProcess(name string) error {
	url := fmt.Sprintf("http://unix/process/%s/stop", name)
	resp, err := c.httpClient.Post(url, "application/json", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to stop process: %d", resp.StatusCode)
	}
	return nil
}

// RestartProcess restarts a specific process.
func (c *Client) RestartProcess(name string) error {
	url := fmt.Sprintf("http://unix/process/%s/restart", name)
	resp, err := c.httpClient.Post(url, "application/json", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to restart process: %d", resp.StatusCode)
	}
	return nil
}

// ShutdownProject stops all processes and shuts down.
func (c *Client) ShutdownProject() error {
	resp, err := c.httpClient.Post("http://unix/project/stop", "application/json", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to shutdown: %d", resp.StatusCode)
	}
	return nil
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/compose/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/compose/
git commit -m "feat: add process-compose REST API client"
```

---

## Remaining Phases (Summary)

The plan continues with:

### Phase 6: Log Viewing
- Task 6.1: Log buffer implementation
- Task 6.2: Log viewport component
- Task 6.3: Unified log interleaver
- Task 6.4: Log search and filter

### Phase 7: Alerts & Notifications
- Task 7.1: Toast notification component
- Task 7.2: System notifications (beeep)
- Task 7.3: Alert history buffer
- Task 7.4: Health monitor goroutine

### Phase 8: Settings Panel
- Task 8.1: huh form integration
- Task 8.2: Settings view
- Task 8.3: Config save/reload

### Phase 9: Polish
- Task 9.1: Splash screen
- Task 9.2: Progress bar component
- Task 9.3: ASCII art customization

---

## Summary

This plan covers the core implementation of acidBurn. Each task follows TDD with:
- Exact file paths
- Failing test first
- Minimal implementation
- Verification
- Commit

Execute with `superpowers:executing-plans` or `superpowers:subagent-driven-development`.
