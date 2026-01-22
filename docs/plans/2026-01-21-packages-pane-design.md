# Packages Pane Feature Design

**Date:** 2026-01-21
**Status:** Approved
**Branch:** feature/packages-pane

## Overview

Add a packages pane to devdash that displays installed development packages (Go, Python, Node.js, etc.) for each project. This complements the existing services pane and helps users understand their complete development environment.

The feature includes adaptive layout that shows both services and packages when screen space allows, or provides a toggle between them on smaller terminals.

## Motivation

Currently, devdash focuses on long-running services managed by process-compose. However, devenv projects also include development packages and language runtimes. Users want visibility into:

- What packages are installed in each project
- Package versions for debugging compatibility issues
- Complete picture of their development environment

Some projects are package-only (no services) - they use `devenv shell` for environment activation but never run `devenv up`. These projects currently show "No services running" which doesn't reflect their purpose.

## Architecture & Components

### New Components

1. **Package Scanner** (`internal/packages/`)
   - Scans `.devenv/profile/bin/` to discover installed packages
   - Parses Nix store paths to extract package names and versions
   - Implements caching based on `.devenv/profile` mtime
   - Groups binaries by their source package

2. **PackagesView Component** (`internal/ui/packages.go`)
   - Renders package list with columns: NAME, VERSION, TYPE
   - Scrollable list matching ServicesView style
   - Shows package count in header
   - Handles empty state gracefully

3. **Adaptive Layout Manager** (enhanced `internal/ui/model.go`)
   - Determines whether to show both panes or single pane
   - Threshold: 140 columns
   - Manages 'p' key toggle for switching views
   - Updates indicators to show hidden pane

### Enhanced Detection

Current project detection already works correctly:
- Socket exists at `.devenv/run/pc.sock` → Service-based project
- No socket → Package-only project

This detection aligns with devenv behavior:
- `devenv up` creates pc.sock (only for projects with services)
- `devenv shell` activates environment (all projects, doesn't create socket)

## Package Detection & Data Flow

### Discovery Strategy

**Dynamic scanning approach:**

1. **Enumerate all binaries** in `.devenv/profile/bin/`
   - No hardcoded whitelist of package names
   - Discover everything that exists

2. **Parse Nix store paths:**
   - Symlink format: `/nix/store/<hash>-<pkgname>-<version>/bin/<binary>`
   - Extract: package name, version, store hash
   - Example: `/nix/store/abc123-python3-3.11.7/bin/python3` → name: python3, version: 3.11.7

3. **Group by package:**
   - Multiple binaries from same Nix package (same hash) → single entry
   - Example: `go`, `gofmt`, `godoc` all from `abc123-go-1.21.5` → show as "Go 1.21.5"
   - Example: `gopls` from different hash → separate entry "gopls 0.14.0"

4. **Categorize for display:**
   - Detect type from package name patterns (go → Go, python → Python, node → Node.js)
   - Used for TYPE column, not for filtering

5. **Enrich with version commands** (optional):
   - Attempt `<binary> --version`, `<binary> version`, `<binary> -v`
   - Parse output if successful
   - Fallback to Nix store path version

### Data Structures

```go
type Package struct {
    Name     string  // e.g., "go", "python3", "node"
    Version  string  // e.g., "1.21.5", "3.11.7"
    Type     string  // e.g., "Go", "Python", "Node.js"
    Binary   string  // Full path to binary
}

type PackageInfo struct {
    ProjectPath string
    Packages    []Package
    LastScanned time.Time
}
```

### Data Flow

**When project is selected:**

1. Model calls `packages.Scan(projectPath)`
2. Scanner checks cache:
   - If cached and `.devenv/profile` mtime unchanged → return cached data
   - Otherwise, perform fresh scan
3. Scanner reads `.devenv/profile/bin/` directory
4. For each file:
   - Resolve symlink to Nix store path
   - Parse store path
   - Group by package identifier
5. Return grouped `[]Package`
6. Model passes data to PackagesView for rendering

**When toggling between Services/Packages:**
- 'p' key updates `showPackages` boolean flag
- Model recalculates layout based on current width and flag
- Re-renders appropriate pane

**Adaptive Layout Decision:**
```
if terminalWidth >= 140:
    show both Services and Packages panes
    split vertical space: Services top, Packages bottom
else:
    show only one pane based on showPackages flag
    display indicator: "SERVICES [p:packages]" or "PACKAGES [p:services]"
```

### Caching Strategy

- Scan packages once when project selected
- Re-scan only when `.devenv/profile` symlink changes (mtime check)
- Avoids repeated filesystem operations
- Cache stored in Model, cleared on project switch

### Error Handling

- `.devenv/profile/bin/` doesn't exist → show empty packages list
- Symlink resolution fails → skip that binary, log error
- Version parsing fails → show "unknown" version, still display package
- No packages found → show friendly message: "No packages detected"

## UI Layout & Interaction

### PackagesView Component Structure

Similar to ServicesView:
- Header with title and count: "PACKAGES (12)"
- Column headers: "NAME", "VERSION", "TYPE"
- Scrollable list of packages
- Styles matching current theme
- Focus highlighting

### Layout Configurations

**Wide terminal (≥140 columns):**
```
┌─────────────┬──────────────────────────────────┐
│  PROJECTS   │  SERVICES (3)                    │
│             │  NAME    STATUS   CPU   MEM      │
│  ● proj-a   │  web     Running  2.1%  45MB     │
│    proj-b   │  api     Running  1.5%  32MB     │
│             │  worker  Stopped  0.0%  0MB      │
│             ├──────────────────────────────────┤
│             │  PACKAGES (8)                    │
│             │  NAME      VERSION    TYPE       │
│             │  go        1.21.5     Go         │
│             │  gopls     0.14.0     Go Tools   │
│             │  python3   3.11.7     Python     │
│             ├──────────────────────────────────┤
│             │  LOGS                            │
│             │  [log entries...]                │
└─────────────┴──────────────────────────────────┘
```

**Narrow terminal (<140 columns, showing Services):**
```
┌─────────────┬──────────────────────────────────┐
│  PROJECTS   │  SERVICES (3)         [p:packages]│
│             │  NAME    STATUS   CPU   MEM      │
│  ● proj-a   │  web     Running  2.1%  45MB     │
│    proj-b   │  api     Running  1.5%  32MB     │
│             │  worker  Stopped  0.0%  0MB      │
│             ├──────────────────────────────────┤
│             │  LOGS                            │
│             │  [log entries...]                │
└─────────────┴──────────────────────────────────┘
```

**Narrow terminal (<140 columns, after pressing 'p'):**
```
┌─────────────┬──────────────────────────────────┐
│  PROJECTS   │  PACKAGES (8)        [p:services]│
│             │  NAME      VERSION    TYPE       │
│  ● proj-a   │  go        1.21.5     Go         │
│    proj-b   │  gopls     0.14.0     Go Tools   │
│             │  python3   3.11.7     Python     │
│             │  node      20.10.0    Node.js    │
│             ├──────────────────────────────────┤
│             │  LOGS                            │
│             │  [log entries...]                │
└─────────────┴──────────────────────────────────┘
```

### Key Bindings

- `p` - Toggle between Services and Packages view (when terminal < 140 columns)
- `Tab` - Cycle focus (Sidebar → Services/Packages → Logs)
- Mouse hover still updates focus appropriately
- All existing keys remain unchanged

### Visual Feedback

- Indicator updates immediately when 'p' pressed
- Smooth transition without flicker
- Focus remains on pane after toggle
- Package count in header shows total packages found
- Conservative 140-column threshold (most modern terminals exceed this)

## Implementation Plan

### File Structure

**New files:**
```
internal/packages/
├── scanner.go       # Package discovery and parsing
├── scanner_test.go  # Scanner unit tests
└── types.go         # Package data structures

internal/ui/
├── packages.go      # PackagesView component
└── packages_test.go # PackagesView unit tests
```

**Modified files:**
```
internal/ui/
├── model.go         # Add adaptive layout logic, 'p' toggle
└── model_test.go    # Add tests for layout switching
```

### Key Functions

```go
// internal/packages/scanner.go
func Scan(projectPath string) ([]Package, error)
func parseNixStorePath(symlinkTarget string) (name, version string)
func groupByPackage(binaries []binaryInfo) []Package

// internal/ui/packages.go
func NewPackagesView(styles *Styles) *PackagesView
func (pv *PackagesView) SetPackages(packages []Package)
func (pv *PackagesView) View() string

// internal/ui/model.go
func (m *Model) shouldShowBothPanes() bool
func (m *Model) togglePackagesView() tea.Cmd
```

### Testing Strategy

**1. Package Scanner Tests:**
- Create mock `.devenv/profile/bin/` with test symlinks
- Verify parsing of various Nix store path formats
- Test caching behavior (mtime checks)
- Test error handling (missing directories, broken symlinks)
- Test grouping logic (multiple binaries from same package)

**2. PackagesView Tests:**
- Test rendering with various package counts (0, 1, many)
- Verify column alignment and truncation
- Test empty state message display
- Test styling consistency with theme
- Test focus highlighting

**3. Layout Tests:**
- Test adaptive layout at 139, 140, 141 column widths
- Verify 'p' toggle switches view correctly
- Verify indicator text appears in correct format
- Test that both panes render when wide enough
- Test space allocation between Services and Packages panes

**4. Integration Tests:**
- Test project switching updates packages
- Verify focus cycling includes packages pane
- Test mouse interaction with packages pane
- Test that package-only projects display packages (no services)
- Test that service-based projects can show both

All tests should follow existing patterns and achieve similar coverage to current codebase (70+ tests, all passing).

## Future Enhancements

Potential future additions (not in initial implementation):

- Filter packages by type (show only Go packages, etc.)
- Search packages by name
- Show package dependencies
- Link to package documentation
- Show package installation date
- Compare packages across projects
- Export package list to file

## Success Criteria

- Package scanning discovers all binaries without hardcoded whitelist
- Adaptive layout smoothly shows/hides panes based on terminal width
- Toggle key ('p') works intuitively on narrow terminals
- Package-only projects display useful information
- All existing functionality remains unchanged
- Comprehensive test coverage (similar to existing 70+ tests)
- No performance degradation on project switching
