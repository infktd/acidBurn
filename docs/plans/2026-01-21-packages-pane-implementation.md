# Packages Pane Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add packages pane to display installed development packages (Go, Python, Node.js, etc.) with adaptive layout.

**Architecture:** Dynamic package discovery by parsing Nix store paths in `.devenv/profile/bin/`. Adaptive UI shows both Services and Packages panes when terminal ≥140 columns, otherwise toggles with 'p' key.

**Tech Stack:** Go, Bubble Tea TUI framework, Nix store path parsing, file system operations

---

## Task 1: Package Types

Create data structures for package information.

**Files:**
- Create: `internal/packages/types.go`
- Test: `internal/packages/types_test.go`

**Step 1: Write the failing test**

```go
package packages

import (
	"testing"
)

func TestPackageStructure(t *testing.T) {
	pkg := Package{
		Name:    "go",
		Version: "1.21.5",
		Type:    "Go",
		Binary:  "/nix/store/abc-go-1.21.5/bin/go",
	}

	if pkg.Name != "go" {
		t.Errorf("expected Name 'go', got %q", pkg.Name)
	}
	if pkg.Version != "1.21.5" {
		t.Errorf("expected Version '1.21.5', got %q", pkg.Version)
	}
}

func TestPackageInfoStructure(t *testing.T) {
	info := PackageInfo{
		ProjectPath: "/home/user/project",
		Packages:    []Package{{Name: "go"}},
	}

	if len(info.Packages) != 1 {
		t.Errorf("expected 1 package, got %d", len(info.Packages))
	}
}
```

**Step 2: Run test to verify it fails**

Run: `CGO_ENABLED=0 go test ./internal/packages/ -v`
Expected: FAIL with "undefined: Package" or package not found

**Step 3: Write types**

Create `internal/packages/types.go`:

```go
package packages

import "time"

// Package represents a single installed package.
type Package struct {
	Name    string // e.g., "go", "python3", "node"
	Version string // e.g., "1.21.5", "3.11.7"
	Type    string // e.g., "Go", "Python", "Node.js"
	Binary  string // Full path to binary
}

// PackageInfo holds package information for a project.
type PackageInfo struct {
	ProjectPath string
	Packages    []Package
	LastScanned time.Time
}
```

**Step 4: Run test to verify it passes**

Run: `CGO_ENABLED=0 go test ./internal/packages/ -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/packages/types.go internal/packages/types_test.go
git commit -m "feat(packages): add Package and PackageInfo types

Add data structures for package discovery:
- Package: name, version, type, binary path
- PackageInfo: project path, packages list, scan timestamp

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 2: Nix Store Path Parser

Parse Nix store paths to extract package name and version.

**Files:**
- Create: `internal/packages/scanner.go`
- Modify: `internal/packages/types_test.go` (add parser tests)

**Step 1: Write the failing test**

Add to `internal/packages/types_test.go`:

```go
func TestParseNixStorePath(t *testing.T) {
	tests := []struct {
		name            string
		path            string
		expectedPkg     string
		expectedVersion string
		expectedHash    string
	}{
		{
			name:            "standard format",
			path:            "/nix/store/abc123-go-1.21.5/bin/go",
			expectedPkg:     "go",
			expectedVersion: "1.21.5",
			expectedHash:    "abc123",
		},
		{
			name:            "python with dots",
			path:            "/nix/store/xyz789-python3-3.11.7/bin/python3",
			expectedPkg:     "python3",
			expectedVersion: "3.11.7",
			expectedHash:    "xyz789",
		},
		{
			name:            "package with complex name",
			path:            "/nix/store/def456-gopls-0.14.0-unstable/bin/gopls",
			expectedPkg:     "gopls",
			expectedVersion: "0.14.0-unstable",
			expectedHash:    "def456",
		},
		{
			name:            "no version",
			path:            "/nix/store/ghi789-bash/bin/bash",
			expectedPkg:     "bash",
			expectedVersion: "",
			expectedHash:    "ghi789",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pkg, version, hash := parseNixStorePath(tt.path)
			if pkg != tt.expectedPkg {
				t.Errorf("expected package %q, got %q", tt.expectedPkg, pkg)
			}
			if version != tt.expectedVersion {
				t.Errorf("expected version %q, got %q", tt.expectedVersion, version)
			}
			if hash != tt.expectedHash {
				t.Errorf("expected hash %q, got %q", tt.expectedHash, hash)
			}
		})
	}
}
```

**Step 2: Run test to verify it fails**

Run: `CGO_ENABLED=0 go test ./internal/packages/ -v -run TestParseNixStorePath`
Expected: FAIL with "undefined: parseNixStorePath"

**Step 3: Write minimal parser**

Create `internal/packages/scanner.go`:

```go
package packages

import (
	"path/filepath"
	"regexp"
	"strings"
)

// parseNixStorePath extracts package name, version, and hash from a Nix store path.
// Format: /nix/store/<hash>-<name>-<version>/...
// Returns: (name, version, hash)
func parseNixStorePath(path string) (string, string, string) {
	// Get the store directory component
	// e.g., /nix/store/abc123-go-1.21.5/bin/go -> abc123-go-1.21.5
	parts := strings.Split(filepath.Clean(path), string(filepath.Separator))

	var storeComponent string
	for i, part := range parts {
		if part == "store" && i+1 < len(parts) {
			storeComponent = parts[i+1]
			break
		}
	}

	if storeComponent == "" {
		return "", "", ""
	}

	// Parse: <hash>-<name>-<version>
	// Regex: (<hash>)-(<name>)-(<version>) where version is optional
	re := regexp.MustCompile(`^([a-z0-9]+)-(.+?)(?:-([0-9].*))?$`)
	matches := re.FindStringSubmatch(storeComponent)

	if len(matches) < 3 {
		// Fallback: just hash-name format
		dashIdx := strings.Index(storeComponent, "-")
		if dashIdx > 0 {
			return storeComponent[dashIdx+1:], "", storeComponent[:dashIdx]
		}
		return storeComponent, "", ""
	}

	hash := matches[1]
	name := matches[2]
	version := ""
	if len(matches) > 3 {
		version = matches[3]
	}

	return name, version, hash
}
```

**Step 4: Run test to verify it passes**

Run: `CGO_ENABLED=0 go test ./internal/packages/ -v -run TestParseNixStorePath`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/packages/scanner.go internal/packages/types_test.go
git commit -m "feat(packages): add Nix store path parser

Parse Nix store paths to extract:
- Package name
- Version (if present)
- Store hash for grouping

Handles various formats including versioned and unversioned packages.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 3: Package Type Categorization

Categorize packages by type (Go, Python, Node.js, etc.) based on name patterns.

**Files:**
- Modify: `internal/packages/scanner.go`
- Modify: `internal/packages/types_test.go`

**Step 1: Write the failing test**

Add to `internal/packages/types_test.go`:

```go
func TestCategorizePackage(t *testing.T) {
	tests := []struct {
		name         string
		packageName  string
		expectedType string
	}{
		{"go", "go", "Go"},
		{"gofmt", "gofmt", "Go"},
		{"gopls", "gopls", "Go"},
		{"python3", "python3", "Python"},
		{"python", "python", "Python"},
		{"pip", "pip", "Python"},
		{"pytest", "pytest", "Python"},
		{"node", "node", "Node.js"},
		{"npm", "npm", "Node.js"},
		{"nodejs", "nodejs", "Node.js"},
		{"cargo", "cargo", "Rust"},
		{"rustc", "rustc", "Rust"},
		{"gcc", "gcc", "C/C++"},
		{"clang", "clang", "C/C++"},
		{"unknown", "foobar", "Other"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typ := categorizePackage(tt.packageName)
			if typ != tt.expectedType {
				t.Errorf("expected type %q for %q, got %q", tt.expectedType, tt.packageName, typ)
			}
		})
	}
}
```

**Step 2: Run test to verify it fails**

Run: `CGO_ENABLED=0 go test ./internal/packages/ -v -run TestCategorizePackage`
Expected: FAIL with "undefined: categorizePackage"

**Step 3: Write categorization function**

Add to `internal/packages/scanner.go`:

```go
// categorizePackage determines the package type based on name patterns.
func categorizePackage(name string) string {
	lowerName := strings.ToLower(name)

	// Go packages
	if strings.HasPrefix(lowerName, "go") ||
	   lowerName == "gopls" ||
	   lowerName == "gofmt" ||
	   lowerName == "godoc" {
		return "Go"
	}

	// Python packages
	if strings.Contains(lowerName, "python") ||
	   lowerName == "pip" ||
	   lowerName == "pytest" ||
	   lowerName == "poetry" {
		return "Python"
	}

	// Node.js packages
	if lowerName == "node" ||
	   lowerName == "nodejs" ||
	   lowerName == "npm" ||
	   lowerName == "npx" ||
	   lowerName == "yarn" {
		return "Node.js"
	}

	// Rust packages
	if lowerName == "cargo" ||
	   lowerName == "rustc" ||
	   lowerName == "rustup" {
		return "Rust"
	}

	// C/C++ compilers
	if lowerName == "gcc" ||
	   lowerName == "g++" ||
	   lowerName == "clang" ||
	   lowerName == "clang++" {
		return "C/C++"
	}

	// Ruby
	if lowerName == "ruby" ||
	   lowerName == "gem" ||
	   lowerName == "bundle" {
		return "Ruby"
	}

	// Java
	if lowerName == "java" ||
	   lowerName == "javac" ||
	   lowerName == "maven" ||
	   lowerName == "gradle" {
		return "Java"
	}

	return "Other"
}
```

**Step 4: Run test to verify it passes**

Run: `CGO_ENABLED=0 go test ./internal/packages/ -v -run TestCategorizePackage`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/packages/scanner.go internal/packages/types_test.go
git commit -m "feat(packages): add package type categorization

Categorize packages by name patterns:
- Go (go, gopls, gofmt)
- Python (python, pip, pytest)
- Node.js (node, npm, yarn)
- Rust (cargo, rustc)
- C/C++ (gcc, clang)
- Ruby, Java, and Other

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 4: Package Scanner - Binary Discovery

Scan `.devenv/profile/bin/` to discover all binaries and resolve symlinks.

**Files:**
- Modify: `internal/packages/scanner.go`
- Modify: `internal/packages/types_test.go`

**Step 1: Write the failing test**

Add to `internal/packages/types_test.go`:

```go
import (
	"os"
	"path/filepath"
)

func TestScan(t *testing.T) {
	// Create temporary test directory structure
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "testproject")
	binDir := filepath.Join(projectDir, ".devenv", "profile", "bin")

	err := os.MkdirAll(binDir, 0755)
	if err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	// Create test binary files (regular files for testing, not actual symlinks)
	testBinaries := []string{"go", "python3", "node"}
	for _, bin := range testBinaries {
		binPath := filepath.Join(binDir, bin)
		err := os.WriteFile(binPath, []byte("#!/bin/sh\necho test"), 0755)
		if err != nil {
			t.Fatalf("failed to create test binary %s: %v", bin, err)
		}
	}

	// Run scan
	packages, err := Scan(projectDir)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	// Should find at least some packages
	if len(packages) == 0 {
		t.Error("expected to find packages, got none")
	}
}

func TestScanMissingDirectory(t *testing.T) {
	packages, err := Scan("/nonexistent/path")

	// Should not error, just return empty list
	if err != nil {
		t.Errorf("expected no error for missing directory, got %v", err)
	}
	if len(packages) != 0 {
		t.Errorf("expected 0 packages for missing directory, got %d", len(packages))
	}
}
```

**Step 2: Run test to verify it fails**

Run: `CGO_ENABLED=0 go test ./internal/packages/ -v -run TestScan`
Expected: FAIL with "undefined: Scan"

**Step 3: Write Scan function**

Add to `internal/packages/scanner.go`:

```go
import (
	"io/fs"
	"os"
	"path/filepath"
)

// Scan discovers packages in a devenv project.
// Returns list of packages found, or empty list if directory doesn't exist.
func Scan(projectPath string) ([]Package, error) {
	binDir := filepath.Join(projectPath, ".devenv", "profile", "bin")

	// Check if bin directory exists
	if _, err := os.Stat(binDir); os.IsNotExist(err) {
		return []Package{}, nil
	}

	var binaries []binaryInfo

	// Read all files in bin directory
	err := filepath.WalkDir(binDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		// Skip directories and the bin directory itself
		if d.IsDir() {
			return nil
		}

		// Resolve symlink to find Nix store path
		target, err := os.Readlink(path)
		if err != nil {
			// Not a symlink, use the file path itself
			target = path
		}

		// Parse Nix store path
		name, version, hash := parseNixStorePath(target)
		if name == "" {
			// Couldn't parse, skip
			return nil
		}

		binaries = append(binaries, binaryInfo{
			name:    name,
			version: version,
			hash:    hash,
			path:    path,
		})

		return nil
	})

	if err != nil {
		return []Package{}, err
	}

	// Group binaries by package
	packages := groupByPackage(binaries)

	return packages, nil
}

// binaryInfo holds parsed information about a binary.
type binaryInfo struct {
	name    string
	version string
	hash    string
	path    string
}
```

**Step 4: Run test to verify it passes**

Run: `CGO_ENABLED=0 go test ./internal/packages/ -v -run TestScan`
Expected: PASS (may need to implement groupByPackage stub first)

**Step 5: Commit**

```bash
git add internal/packages/scanner.go internal/packages/types_test.go
git commit -m "feat(packages): add binary discovery scanner

Scan .devenv/profile/bin/ to discover installed packages:
- Walk bin directory
- Resolve symlinks to Nix store paths
- Parse package information from paths
- Handle missing directories gracefully

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 5: Package Grouping

Group binaries from the same Nix package into single Package entries.

**Files:**
- Modify: `internal/packages/scanner.go`
- Modify: `internal/packages/types_test.go`

**Step 1: Write the failing test**

Add to `internal/packages/types_test.go`:

```go
func TestGroupByPackage(t *testing.T) {
	binaries := []binaryInfo{
		{name: "go", version: "1.21.5", hash: "abc123", path: "/path/to/go"},
		{name: "gofmt", version: "1.21.5", hash: "abc123", path: "/path/to/gofmt"},
		{name: "gopls", version: "0.14.0", hash: "xyz789", path: "/path/to/gopls"},
		{name: "python3", version: "3.11.7", hash: "def456", path: "/path/to/python3"},
	}

	packages := groupByPackage(binaries)

	// Should have 3 packages (go+gofmt grouped, gopls separate, python separate)
	if len(packages) != 3 {
		t.Errorf("expected 3 packages, got %d", len(packages))
	}

	// Find the Go package
	var goPackage *Package
	for i := range packages {
		if packages[i].Name == "go" {
			goPackage = &packages[i]
			break
		}
	}

	if goPackage == nil {
		t.Fatal("expected to find 'go' package")
	}

	if goPackage.Version != "1.21.5" {
		t.Errorf("expected go version 1.21.5, got %q", goPackage.Version)
	}

	if goPackage.Type != "Go" {
		t.Errorf("expected go type 'Go', got %q", goPackage.Type)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `CGO_ENABLED=0 go test ./internal/packages/ -v -run TestGroupByPackage`
Expected: FAIL or incorrect grouping

**Step 3: Write grouping function**

Add to `internal/packages/scanner.go`:

```go
// groupByPackage groups binaries from the same Nix package.
// Binaries with the same hash+name+version belong to the same package.
func groupByPackage(binaries []binaryInfo) []Package {
	// Use map to group by package identifier (hash-name-version)
	packageMap := make(map[string]*Package)

	for _, bin := range binaries {
		// Create unique package identifier
		pkgID := bin.hash + "-" + bin.name + "-" + bin.version

		if pkg, exists := packageMap[pkgID]; exists {
			// Package already seen, skip this binary
			continue
		}

		// New package
		packageMap[pkgID] = &Package{
			Name:    bin.name,
			Version: bin.version,
			Type:    categorizePackage(bin.name),
			Binary:  bin.path,
		}
	}

	// Convert map to slice
	packages := make([]Package, 0, len(packageMap))
	for _, pkg := range packageMap {
		packages = append(packages, *pkg)
	}

	return packages
}
```

**Step 4: Run test to verify it passes**

Run: `CGO_ENABLED=0 go test ./internal/packages/ -v -run TestGroupByPackage`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/packages/scanner.go internal/packages/types_test.go
git commit -m "feat(packages): add package grouping logic

Group binaries from same Nix package:
- Use hash+name+version as unique identifier
- Merge multiple binaries into single Package entry
- Prevents duplicates (go, gofmt, godoc → one 'go' package)

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 6: PackagesView Component Structure

Create the PackagesView component for rendering package lists.

**Files:**
- Create: `internal/ui/packages.go`
- Create: `internal/ui/packages_test.go`

**Step 1: Write the failing test**

Create `internal/ui/packages_test.go`:

```go
package ui

import (
	"testing"

	"github.com/infktd/devdash/internal/packages"
)

func TestPackagesViewCreate(t *testing.T) {
	styles := NewStyles(GetTheme("matrix"))
	pv := NewPackagesView(styles)

	if pv == nil {
		t.Fatal("NewPackagesView returned nil")
	}

	if pv.styles != styles {
		t.Error("styles not set correctly")
	}
}

func TestPackagesViewSetPackages(t *testing.T) {
	styles := NewStyles(GetTheme("matrix"))
	pv := NewPackagesView(styles)

	pkgs := []packages.Package{
		{Name: "go", Version: "1.21.5", Type: "Go"},
		{Name: "python3", Version: "3.11.7", Type: "Python"},
	}

	pv.SetPackages(pkgs)

	if len(pv.packages) != 2 {
		t.Errorf("expected 2 packages, got %d", len(pv.packages))
	}
}

func TestPackagesViewEmptyState(t *testing.T) {
	styles := NewStyles(GetTheme("matrix"))
	pv := NewPackagesView(styles)
	pv.SetSize(80, 20)

	view := pv.View()

	if view == "" {
		t.Error("View should not be empty")
	}
}

func TestPackagesViewWithPackages(t *testing.T) {
	styles := NewStyles(GetTheme("matrix"))
	pv := NewPackagesView(styles)
	pv.SetSize(80, 20)

	pkgs := []packages.Package{
		{Name: "go", Version: "1.21.5", Type: "Go"},
	}
	pv.SetPackages(pkgs)

	view := pv.View()

	if view == "" {
		t.Error("View should not be empty with packages")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `CGO_ENABLED=0 go test ./internal/ui/ -v -run TestPackagesView`
Expected: FAIL with "undefined: NewPackagesView"

**Step 3: Write PackagesView structure**

Create `internal/ui/packages.go`:

```go
package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/infktd/devdash/internal/packages"
)

// PackagesView displays installed packages for a project.
type PackagesView struct {
	styles   *Styles
	packages []packages.Package
	width    int
	height   int
	focused  bool
}

// NewPackagesView creates a new packages view.
func NewPackagesView(styles *Styles) *PackagesView {
	return &PackagesView{
		styles:   styles,
		packages: []packages.Package{},
		width:    0,
		height:   0,
	}
}

// SetPackages updates the package list.
func (pv *PackagesView) SetPackages(pkgs []packages.Package) {
	pv.packages = pkgs
}

// SetSize updates dimensions.
func (pv *PackagesView) SetSize(width, height int) {
	pv.width = width
	pv.height = height
}

// SetFocused sets focus state.
func (pv *PackagesView) SetFocused(focused bool) {
	pv.focused = focused
}

// View renders the packages view.
func (pv *PackagesView) View() string {
	if pv.width == 0 || pv.height == 0 {
		return ""
	}

	// Header
	title := fmt.Sprintf("PACKAGES (%d)", len(pv.packages))
	headerStyle := pv.styles.PaneTitle
	if pv.focused {
		headerStyle = pv.styles.PaneTitleFocused
	}
	header := headerStyle.Render(title)

	// Empty state
	if len(pv.packages) == 0 {
		emptyMsg := pv.styles.Muted.Render("No packages detected")
		content := lipgloss.NewStyle().
			Width(pv.width).
			Height(pv.height - 1).
			Align(lipgloss.Center, lipgloss.Center).
			Render(emptyMsg)

		return lipgloss.JoinVertical(lipgloss.Left, header, content)
	}

	// Column headers
	colHeaders := pv.renderColumnHeaders()

	// Package rows
	rows := pv.renderPackageRows()

	// Combine
	var lines []string
	lines = append(lines, header)
	lines = append(lines, colHeaders)
	lines = append(lines, rows...)

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

// renderColumnHeaders renders the column header row.
func (pv *PackagesView) renderColumnHeaders() string {
	nameCol := pv.styles.ColumnHeader.Width(20).Render("NAME")
	versionCol := pv.styles.ColumnHeader.Width(15).Render("VERSION")
	typeCol := pv.styles.ColumnHeader.Width(15).Render("TYPE")

	return lipgloss.JoinHorizontal(lipgloss.Top, nameCol, versionCol, typeCol)
}

// renderPackageRows renders package data rows.
func (pv *PackagesView) renderPackageRows() []string {
	var rows []string

	// Limit to visible height (header + column headers take 2 lines)
	maxRows := pv.height - 2
	if maxRows < 0 {
		maxRows = 0
	}

	for i := 0; i < len(pv.packages) && i < maxRows; i++ {
		pkg := pv.packages[i]

		// Truncate long names
		name := pkg.Name
		if len(name) > 18 {
			name = name[:15] + "..."
		}

		version := pkg.Version
		if version == "" {
			version = "unknown"
		}
		if len(version) > 13 {
			version = version[:10] + "..."
		}

		typ := pkg.Type
		if len(typ) > 13 {
			typ = typ[:10] + "..."
		}

		nameCell := lipgloss.NewStyle().Width(20).Render(name)
		versionCell := lipgloss.NewStyle().Width(15).Render(version)
		typeCell := lipgloss.NewStyle().Width(15).Render(typ)

		row := lipgloss.JoinHorizontal(lipgloss.Top, nameCell, versionCell, typeCell)
		rows = append(rows, row)
	}

	return rows
}
```

**Step 4: Run test to verify it passes**

Run: `CGO_ENABLED=0 go test ./internal/ui/ -v -run TestPackagesView`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/ui/packages.go internal/ui/packages_test.go
git commit -m "feat(ui): add PackagesView component

New component to display package list:
- Header with package count
- Column headers: NAME, VERSION, TYPE
- Package rows with truncation
- Empty state message
- Focus highlighting support

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 7: Model Integration - Add PackagesView

Integrate PackagesView into the main Model.

**Files:**
- Modify: `internal/ui/model.go:40-80` (add packagesView field and initialization)
- Modify: `internal/ui/model_test.go`

**Step 1: Write the failing test**

Add to `internal/ui/model_test.go`:

```go
func TestModelPackagesViewInitialized(t *testing.T) {
	cfg := config.Default()
	reg := &registry.Registry{}
	m := New(cfg, reg)

	if m.packagesView == nil {
		t.Error("packagesView should be initialized")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `CGO_ENABLED=0 go test ./internal/ui/ -v -run TestModelPackagesViewInitialized`
Expected: FAIL with "m.packagesView undefined"

**Step 3: Add packagesView to Model**

In `internal/ui/model.go`, find the Model struct definition and add:

```go
// Add to Model struct (around line 40-50)
packagesView *PackagesView
```

Then in the `New` function, add initialization:

```go
// In New function (around line 80-100), add:
packagesView: NewPackagesView(styles),
```

**Step 4: Run test to verify it passes**

Run: `CGO_ENABLED=0 go test ./internal/ui/ -v -run TestModelPackagesViewInitialized`
Expected: PASS

**Step 5: Run all tests**

Run: `CGO_ENABLED=0 go test ./... -v`
Expected: All tests PASS

**Step 6: Commit**

```bash
git add internal/ui/model.go internal/ui/model_test.go
git commit -m "feat(ui): integrate PackagesView into Model

Add PackagesView to main Model:
- Initialize in New()
- Add test for initialization

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 8: Adaptive Layout Logic

Add logic to determine when to show both panes vs single pane based on terminal width.

**Files:**
- Modify: `internal/ui/model.go` (add shouldShowBothPanes method)
- Modify: `internal/ui/model_test.go`

**Step 1: Write the failing test**

Add to `internal/ui/model_test.go`:

```go
func TestShouldShowBothPanes(t *testing.T) {
	cfg := config.Default()
	reg := &registry.Registry{}
	m := New(cfg, reg)

	// Narrow terminal - should show single pane
	m.width = 139
	if m.shouldShowBothPanes() {
		t.Error("expected false for width 139")
	}

	// Wide terminal - should show both panes
	m.width = 140
	if !m.shouldShowBothPanes() {
		t.Error("expected true for width 140")
	}

	m.width = 200
	if !m.shouldShowBothPanes() {
		t.Error("expected true for width 200")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `CGO_ENABLED=0 go test ./internal/ui/ -v -run TestShouldShowBothPanes`
Expected: FAIL with "m.shouldShowBothPanes undefined"

**Step 3: Add shouldShowBothPanes method**

Add to `internal/ui/model.go`:

```go
const (
	minWidthForBothPanes = 140
)

// shouldShowBothPanes determines if there's enough space to show both Services and Packages panes.
func (m *Model) shouldShowBothPanes() bool {
	return m.width >= minWidthForBothPanes
}
```

**Step 4: Run test to verify it passes**

Run: `CGO_ENABLED=0 go test ./internal/ui/ -v -run TestShouldShowBothPanes`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/ui/model.go internal/ui/model_test.go
git commit -m "feat(ui): add adaptive layout threshold logic

Determine when to show both Services and Packages panes:
- Threshold: 140 columns
- Below threshold: show single pane with toggle
- At/above threshold: show both panes

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 9: Toggle State Management

Add state for tracking which pane to show when in single-pane mode.

**Files:**
- Modify: `internal/ui/model.go` (add showPackages field)
- Modify: `internal/ui/model_test.go`

**Step 1: Write the failing test**

Add to `internal/ui/model_test.go`:

```go
func TestTogglePackagesView(t *testing.T) {
	cfg := config.Default()
	reg := &registry.Registry{}
	m := New(cfg, reg)
	m.width = 100 // Narrow terminal

	// Default should show services (showPackages = false)
	if m.showPackages {
		t.Error("expected showPackages to be false by default")
	}

	// Toggle to packages
	m.togglePackagesView()
	if !m.showPackages {
		t.Error("expected showPackages to be true after toggle")
	}

	// Toggle back to services
	m.togglePackagesView()
	if m.showPackages {
		t.Error("expected showPackages to be false after second toggle")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `CGO_ENABLED=0 go test ./internal/ui/ -v -run TestTogglePackagesView`
Expected: FAIL with "m.showPackages undefined" or "m.togglePackagesView undefined"

**Step 3: Add toggle state**

In `internal/ui/model.go`, add to Model struct:

```go
// Add to Model struct
showPackages bool // When true, show packages pane instead of services (narrow terminals)
```

Add toggle method:

```go
// togglePackagesView switches between Services and Packages pane (for narrow terminals).
func (m *Model) togglePackagesView() tea.Cmd {
	m.showPackages = !m.showPackages
	return nil
}
```

**Step 4: Run test to verify it passes**

Run: `CGO_ENABLED=0 go test ./internal/ui/ -v -run TestTogglePackagesView`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/ui/model.go internal/ui/model_test.go
git commit -m "feat(ui): add packages view toggle state

Add state management for single-pane mode:
- showPackages flag to track current pane
- togglePackagesView() to switch between panes
- Defaults to showing services pane

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 10: Key Binding for 'p' Toggle

Add 'p' key binding to toggle between Services and Packages view.

**Files:**
- Modify: `internal/ui/model.go` (Update method to handle 'p' key)
- Modify: `internal/ui/model_test.go`

**Step 1: Write the failing test**

Add to `internal/ui/model_test.go`:

```go
import (
	tea "github.com/charmbracelet/bubbletea"
)

func TestPressP_TogglesPackagesView(t *testing.T) {
	cfg := config.Default()
	reg := &registry.Registry{}
	m := New(cfg, reg)
	m.width = 100 // Narrow terminal
	m.showSplash = false

	initialState := m.showPackages

	// Press 'p' key
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}}
	newModel, _ := m.Update(msg)
	model := newModel.(*Model)

	if model.showPackages == initialState {
		t.Error("pressing 'p' should toggle showPackages")
	}
}

func TestPressP_NoEffectOnWideTerminal(t *testing.T) {
	cfg := config.Default()
	reg := &registry.Registry{}
	m := New(cfg, reg)
	m.width = 150 // Wide terminal - both panes visible
	m.showSplash = false

	// Press 'p' key shouldn't have effect
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}}
	newModel, _ := m.Update(msg)
	model := newModel.(*Model)

	// On wide terminals, 'p' could still toggle but doesn't affect display
	// Just verify it doesn't crash
	_ = model
}
```

**Step 2: Run test to verify it fails**

Run: `CGO_ENABLED=0 go test ./internal/ui/ -v -run "TestPressP_"`
Expected: FAIL or no toggle behavior

**Step 3: Add 'p' key handling**

In `internal/ui/model.go`, find the `Update` method's key handling section and add:

```go
// In Update method, in the tea.KeyMsg case, add:
case "p":
	// Toggle between packages and services view
	return m, m.togglePackagesView()
```

**Step 4: Run test to verify it passes**

Run: `CGO_ENABLED=0 go test ./internal/ui/ -v -run "TestPressP_"`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/ui/model.go internal/ui/model_test.go
git commit -m "feat(ui): add 'p' key binding to toggle panes

Handle 'p' key press to toggle between Services and Packages:
- Calls togglePackagesView()
- Works in both narrow and wide terminals
- Updates display immediately

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 11: Package Scanning on Project Selection

Trigger package scanning when a project is selected and update PackagesView.

**Files:**
- Modify: `internal/ui/model.go` (update switchToCurrentProject method)
- Modify: `internal/ui/model_test.go`

**Step 1: Write the failing test**

Add to `internal/ui/model_test.go`:

```go
func TestProjectSwitch_ScansPackages(t *testing.T) {
	cfg := config.Default()
	reg := &registry.Registry{
		Projects: []*registry.Project{
			{Path: "/tmp/testproject", Name: "test"},
		},
	}
	m := New(cfg, reg)
	m.selectedProject = 0

	// Switch to project (this will try to scan packages)
	m.switchToCurrentProject()

	// PackagesView should have been updated (even if empty)
	// This is a basic test - in real usage, packages would be populated
	// if .devenv/profile/bin/ exists
	if m.packagesView == nil {
		t.Error("packagesView should not be nil after project switch")
	}
}
```

**Step 2: Run test to verify it fails or passes**

Run: `CGO_ENABLED=0 go test ./internal/ui/ -v -run TestProjectSwitch_ScansPackages`
Expected: May pass but needs implementation verification

**Step 3: Add package scanning to switchToCurrentProject**

In `internal/ui/model.go`, find `switchToCurrentProject` method and add package scanning:

```go
import (
	"github.com/infktd/devdash/internal/packages"
)

// In switchToCurrentProject method, add after clearing services:
func (m *Model) switchToCurrentProject() tea.Cmd {
	// ... existing code ...

	// Scan packages for new project
	project := m.currentProject()
	if project != nil {
		pkgs, err := packages.Scan(project.Path)
		if err != nil {
			// Log error but don't block
			pkgs = []packages.Package{}
		}
		m.packagesView.SetPackages(pkgs)
	} else {
		m.packagesView.SetPackages([]packages.Package{})
	}

	// ... existing code ...
}
```

**Step 4: Run test to verify it passes**

Run: `CGO_ENABLED=0 go test ./internal/ui/ -v -run TestProjectSwitch_ScansPackages`
Expected: PASS

**Step 5: Run all tests**

Run: `CGO_ENABLED=0 go test ./... -v`
Expected: All tests PASS

**Step 6: Commit**

```bash
git add internal/ui/model.go internal/ui/model_test.go
git commit -m "feat(ui): scan packages on project selection

Trigger package scanning when project is switched:
- Call packages.Scan() in switchToCurrentProject()
- Update PackagesView with results
- Handle errors gracefully (empty list on error)

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 12: Render Logic - Wide Terminal Layout

Update View() method to render both Services and Packages panes when terminal is wide enough.

**Files:**
- Modify: `internal/ui/model.go` (update View method)
- Modify: `internal/ui/model_test.go`

**Step 1: Write the failing test**

Add to `internal/ui/model_test.go`:

```go
func TestView_WideTerminal_ShowsBothPanes(t *testing.T) {
	cfg := config.Default()
	reg := &registry.Registry{}
	m := New(cfg, reg)
	m.width = 150
	m.height = 40
	m.showSplash = false

	view := m.View()

	// Both "SERVICES" and "PACKAGES" should appear in view
	if !strings.Contains(view, "SERVICES") {
		t.Error("wide terminal view should contain SERVICES")
	}
	if !strings.Contains(view, "PACKAGES") {
		t.Error("wide terminal view should contain PACKAGES")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `CGO_ENABLED=0 go test ./internal/ui/ -v -run TestView_WideTerminal_ShowsBothPanes`
Expected: FAIL - packages not shown yet

**Step 3: Update View method**

In `internal/ui/model.go`, find the `View` method and update the layout rendering logic:

```go
// In View method, in the main layout section:
// Find where servicesView is rendered and update to:

	// Determine layout based on width
	var mainContent string
	if m.shouldShowBothPanes() {
		// Wide terminal: show both Services and Packages
		// Split vertical space: Services top, Packages bottom
		servicesHeight := (m.height - 4) / 2 // Split roughly in half
		packagesHeight := (m.height - 4) - servicesHeight

		m.servicesView.SetSize(contentWidth, servicesHeight)
		m.packagesView.SetSize(contentWidth, packagesHeight)

		servicesContent := m.servicesView.View()
		packagesContent := m.packagesView.View()

		mainContent = lipgloss.JoinVertical(
			lipgloss.Left,
			servicesContent,
			packagesContent,
		)
	} else {
		// Narrow terminal: show either Services or Packages based on toggle
		if m.showPackages {
			// Show packages pane
			m.packagesView.SetSize(contentWidth, m.height-4)
			mainContent = m.packagesView.View()
		} else {
			// Show services pane
			m.servicesView.SetSize(contentWidth, m.height-4)
			mainContent = m.servicesView.View()
		}
	}
```

Note: This is pseudocode - the actual implementation should integrate with existing View logic. Look for where `m.servicesView.View()` is called and adapt accordingly.

**Step 4: Run test to verify it passes**

Run: `CGO_ENABLED=0 go test ./internal/ui/ -v -run TestView_WideTerminal_ShowsBothPanes`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/ui/model.go internal/ui/model_test.go
git commit -m "feat(ui): render both panes on wide terminals

Update View() to show Services and Packages together:
- Split vertical space when width >= 140
- Services pane on top
- Packages pane on bottom
- Both visible simultaneously

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 13: Render Logic - Narrow Terminal with Indicator

Update View() to show indicator when a pane is hidden on narrow terminals.

**Files:**
- Modify: `internal/ui/model.go` (update View method to add indicator)
- Modify: `internal/ui/model_test.go`

**Step 1: Write the failing test**

Add to `internal/ui/model_test.go`:

```go
func TestView_NarrowTerminal_ShowsIndicator(t *testing.T) {
	cfg := config.Default()
	reg := &registry.Registry{}
	m := New(cfg, reg)
	m.width = 100
	m.height = 40
	m.showSplash = false

	// Showing services - should have packages indicator
	m.showPackages = false
	view := m.View()
	if !strings.Contains(view, "[p:packages]") {
		t.Error("narrow terminal showing services should contain '[p:packages]' indicator")
	}

	// Showing packages - should have services indicator
	m.showPackages = true
	view = m.View()
	if !strings.Contains(view, "[p:services]") {
		t.Error("narrow terminal showing packages should contain '[p:services]' indicator")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `CGO_ENABLED=0 go test ./internal/ui/ -v -run TestView_NarrowTerminal_ShowsIndicator`
Expected: FAIL - indicator not shown

**Step 3: Add indicator to pane headers**

In `internal/ui/model.go`, update the narrow terminal rendering to append indicator to title:

```go
// In the narrow terminal section of View():

if m.showPackages {
	// Showing packages, add services indicator
	title := "PACKAGES"
	if !m.shouldShowBothPanes() {
		title += " [p:services]"
	}
	// Update packagesView header rendering with title
	// (May need to add method to PackagesView to accept custom title)
	m.packagesView.SetSize(contentWidth, m.height-4)
	mainContent = m.packagesView.View()
	// Append indicator to header
} else {
	// Showing services, add packages indicator
	title := "SERVICES"
	if !m.shouldShowBothPanes() {
		title += " [p:packages]"
	}
	// Similar for servicesView
}
```

Note: This may require updating ServicesView and PackagesView to accept custom title overrides, or post-process the View() output to inject the indicator.

**Simpler approach:** Add indicator text to the rendered header using string replacement or custom rendering.

**Step 4: Run test to verify it passes**

Run: `CGO_ENABLED=0 go test ./internal/ui/ -v -run TestView_NarrowTerminal_ShowsIndicator`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/ui/model.go internal/ui/model_test.go
git commit -m "feat(ui): add pane indicators for narrow terminals

Show which pane is hidden on narrow terminals:
- '[p:packages]' when showing services
- '[p:services]' when showing packages
- Appears in header for visibility

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 14: Integration Test - Full Workflow

Add comprehensive integration test covering project selection, package scanning, and rendering.

**Files:**
- Modify: `internal/ui/model_test.go`

**Step 1: Write the integration test**

Add to `internal/ui/model_test.go`:

```go
func TestPackagesPaneIntegration(t *testing.T) {
	// Create test project directory with packages
	tmpDir := t.TempDir()
	projectPath := filepath.Join(tmpDir, "testproject")
	binDir := filepath.Join(projectPath, ".devenv", "profile", "bin")

	err := os.MkdirAll(binDir, 0755)
	if err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	// Create test binaries
	testBins := map[string]string{
		"go":      "#!/bin/sh\necho go1.21",
		"python3": "#!/bin/sh\necho python3.11",
		"node":    "#!/bin/sh\necho node20",
	}

	for name, content := range testBins {
		binPath := filepath.Join(binDir, name)
		err := os.WriteFile(binPath, []byte(content), 0755)
		if err != nil {
			t.Fatalf("failed to create binary %s: %v", name, err)
		}
	}

	// Create model with test project
	cfg := config.Default()
	reg := &registry.Registry{
		Projects: []*registry.Project{
			{Path: projectPath, Name: "testproject"},
		},
	}
	m := New(cfg, reg)
	m.width = 150 // Wide terminal
	m.height = 50
	m.showSplash = false
	m.selectedProject = 0

	// Trigger project switch (should scan packages)
	m.switchToCurrentProject()

	// Verify packages were scanned
	view := m.View()
	if !strings.Contains(view, "PACKAGES") {
		t.Error("view should contain PACKAGES pane")
	}

	// Switch to narrow terminal
	m.width = 100
	msg := tea.WindowSizeMsg{Width: 100, Height: 50}
	newModel, _ := m.Update(msg)
	m = newModel.(*Model)

	// Should show services by default
	view = m.View()
	if !strings.Contains(view, "[p:packages]") {
		t.Error("narrow terminal should show packages indicator")
	}

	// Toggle to packages
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}}
	newModel, _ = m.Update(keyMsg)
	m = newModel.(*Model)

	view = m.View()
	if !strings.Contains(view, "[p:services]") {
		t.Error("narrow terminal showing packages should show services indicator")
	}
}
```

**Step 2: Run test**

Run: `CGO_ENABLED=0 go test ./internal/ui/ -v -run TestPackagesPaneIntegration`
Expected: PASS

**Step 3: Run all tests**

Run: `CGO_ENABLED=0 go test ./... -v`
Expected: All tests PASS (should be 80+ tests now)

**Step 4: Commit**

```bash
git add internal/ui/model_test.go
git commit -m "test(ui): add packages pane integration test

Comprehensive test covering full workflow:
- Project with test packages
- Package scanning on selection
- Wide terminal layout (both panes)
- Narrow terminal toggle behavior
- Indicator display

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 15: Focus Cycling with Packages Pane

Update focus cycling (Tab key) to include packages pane when visible.

**Files:**
- Modify: `internal/ui/model.go` (update cycleFocus method)
- Modify: `internal/ui/model_test.go`

**Step 1: Write the failing test**

Add to `internal/ui/model_test.go`:

```go
func TestCycleFocus_IncludesPackagesPane(t *testing.T) {
	cfg := config.Default()
	reg := &registry.Registry{}
	m := New(cfg, reg)
	m.width = 150 // Wide terminal - both panes visible
	m.showSplash = false

	// Cycle: Sidebar -> Services -> Packages -> Logs -> Sidebar
	// Note: May need to add PanePackages constant

	if m.focused != PaneSidebar {
		t.Errorf("expected PaneSidebar, got %v", m.focused)
	}

	m.cycleFocus()
	if m.focused != PaneServices {
		t.Errorf("expected PaneServices after first cycle, got %v", m.focused)
	}

	// On wide terminals, packages should be in cycle
	if m.shouldShowBothPanes() {
		m.cycleFocus()
		// Should cycle to packages if both panes visible
		// This may require adding PanePackages constant
	}
}
```

**Step 2: Run test to verify it fails**

Run: `CGO_ENABLED=0 go test ./internal/ui/ -v -run TestCycleFocus_IncludesPackagesPane`
Expected: FAIL or incorrect behavior

**Step 3: Update cycleFocus**

In `internal/ui/model.go`:

Add PanePackages constant if not exists:

```go
const (
	PaneSidebar Pane = iota
	PaneServices
	PanePackages  // Add this
	PaneLogs
)
```

Update cycleFocus method:

```go
func (m *Model) cycleFocus() {
	switch m.focused {
	case PaneSidebar:
		m.focused = PaneServices
	case PaneServices:
		// If both panes visible, go to packages next
		// Otherwise skip to logs
		if m.shouldShowBothPanes() {
			m.focused = PanePackages
		} else {
			m.focused = PaneLogs
		}
	case PanePackages:
		m.focused = PaneLogs
	case PaneLogs:
		m.focused = PaneSidebar
	default:
		m.focused = PaneSidebar
	}
}
```

**Step 4: Update SetFocused for packages**

Make sure packagesView focus state is updated when focused:

```go
// In Update or View method where focus is handled:
m.packagesView.SetFocused(m.focused == PanePackages)
```

**Step 5: Run test to verify it passes**

Run: `CGO_ENABLED=0 go test ./internal/ui/ -v -run TestCycleFocus_IncludesPackagesPane`
Expected: PASS

**Step 6: Commit**

```bash
git add internal/ui/model.go internal/ui/model_test.go
git commit -m "feat(ui): include packages pane in focus cycle

Update Tab key focus cycling:
- Add PanePackages constant
- Include in cycleFocus() when both panes visible
- Skip packages on narrow terminals
- Update packagesView focus state

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 16: Mouse Hover Support for Packages Pane

Update mouse hover logic to focus packages pane when mouse is in that region.

**Files:**
- Modify: `internal/ui/model.go` (update mouse handling in Update)
- Modify: `internal/ui/model_test.go`

**Step 1: Write the failing test**

Add to `internal/ui/model_test.go`:

```go
func TestMouse_HoverPackagesPane(t *testing.T) {
	cfg := config.Default()
	cfg.UI.SidebarWidth = 30
	reg := &registry.Registry{}
	m := New(cfg, reg)
	m.width = 150
	m.height = 60
	m.showSplash = false

	// Mouse in packages area (x >= 30, y in packages region)
	// Services roughly at top half, packages at bottom half
	packagesY := 35 // Bottom half

	msg := tea.MouseMsg{
		Type: tea.MouseMotion,
		X:    50,
		Y:    packagesY,
	}

	newModel, _ := m.Update(msg)
	model := newModel.(*Model)

	if model.focused != PanePackages {
		t.Errorf("expected focus on packages, got %v", model.focused)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `CGO_ENABLED=0 go test ./internal/ui/ -v -run TestMouse_HoverPackagesPane`
Expected: FAIL

**Step 3: Update mouse handling**

In `internal/ui/model.go`, find the mouse handling in Update method and add packages region:

```go
// In Update, tea.MouseMsg case, tea.MouseMotion:
case tea.MouseMotion:
	if m.showHelp || m.settingsPanel.IsVisible() || m.alertsPanel.IsVisible() {
		return m, nil
	}

	x, y := msg.X, msg.Y
	sidebarWidth := m.cfg.UI.SidebarWidth

	if x < sidebarWidth {
		m.focused = PaneSidebar
	} else if m.shouldShowBothPanes() {
		// Both panes visible - determine based on Y position
		servicesHeight := (m.height - 4) / 2
		if y < servicesHeight {
			m.focused = PaneServices
		} else if y < servicesHeight*2 {
			m.focused = PanePackages
		} else {
			m.focused = PaneLogs
		}
	} else {
		// Single pane mode
		if m.showPackages {
			// Packages pane is shown
			if y < m.height/2 {
				m.focused = PanePackages
			} else {
				m.focused = PaneLogs
			}
		} else {
			// Services pane is shown
			if y < m.height/2 {
				m.focused = PaneServices
			} else {
				m.focused = PaneLogs
			}
		}
	}
```

**Step 4: Run test to verify it passes**

Run: `CGO_ENABLED=0 go test ./internal/ui/ -v -run TestMouse_HoverPackagesPane`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/ui/model.go internal/ui/model_test.go
git commit -m "feat(ui): add mouse hover support for packages pane

Update mouse motion handling:
- Detect mouse position in packages region
- Update focus to PanePackages when hovering
- Handle both wide and narrow terminal layouts

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 17: Final Testing and Cleanup

Run all tests, verify coverage, and ensure no regressions.

**Step 1: Run complete test suite**

```bash
CGO_ENABLED=0 go test ./... -v -race -coverprofile=coverage.out
```

Expected: All tests PASS, no race conditions, good coverage

**Step 2: Check coverage**

```bash
go tool cover -func=coverage.out | grep total
```

Expected: Coverage similar to or better than baseline (should be >50%)

**Step 3: Run tests multiple times to catch flaky tests**

```bash
for i in {1..5}; do
  echo "Run $i"
  CGO_ENABLED=0 go test ./... -v || break
done
```

Expected: All runs PASS consistently

**Step 4: Build binary**

```bash
CGO_ENABLED=0 go build -v -o devdash ./...
```

Expected: Build succeeds, no errors

**Step 5: Manual smoke test (optional)**

```bash
./devdash --version
```

Expected: Version displayed without crash

**Step 6: Final commit**

```bash
git add -A
git commit -m "chore: final cleanup and testing for packages pane

All tests passing:
- 80+ test cases
- No race conditions
- Coverage maintained
- Binary builds successfully

Feature complete:
- Package scanning from Nix store
- Adaptive layout (140 column threshold)
- Toggle with 'p' key
- Focus cycling and mouse support
- Comprehensive test coverage

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 18: Update Documentation

Update help text and documentation to mention packages pane and 'p' key.

**Files:**
- Modify: `internal/ui/help.go` (add 'p' key to help text)
- Create: `docs/PACKAGES_PANE.md` (user-facing documentation)

**Step 1: Update help text**

Edit `internal/ui/help.go` to add:

```go
// In the help text content, add:
"p           toggle packages/services view (narrow terminals)"
```

**Step 2: Create user documentation**

Create `docs/PACKAGES_PANE.md`:

```markdown
# Packages Pane

devdash displays installed development packages (Go, Python, Node.js, etc.) alongside services.

## Features

- **Automatic Discovery**: Scans `.devenv/profile/bin/` to find all installed packages
- **Version Information**: Shows package versions from Nix store paths
- **Adaptive Layout**: Automatically adjusts based on terminal size
- **Toggle View**: Press `p` to switch between services and packages on narrow terminals

## Layout Modes

### Wide Terminals (≥140 columns)

Both Services and Packages panes are visible simultaneously:
- Services pane in top half
- Packages pane in bottom half
- Full information always visible

### Narrow Terminals (<140 columns)

One pane at a time with toggle:
- Default: Services pane with `[p:packages]` indicator
- Press `p`: Packages pane with `[p:services]` indicator
- Press `p` again: Back to Services pane

## Package Information

Each package shows:
- **NAME**: Package binary name (e.g., "go", "python3")
- **VERSION**: Version extracted from Nix store path
- **TYPE**: Categorized type (Go, Python, Node.js, Rust, etc.)

## Keyboard Shortcuts

- `p`: Toggle between Services and Packages view (narrow terminals)
- `Tab`: Cycle focus through panes (includes Packages when visible)
- Mouse hover: Move focus to pane under cursor

## Project Types

- **Service-based projects**: Have both services and packages
- **Package-only projects**: Development environments without long-running services
  - Example: CLI tools, libraries, static sites
  - Only use `devenv shell`, never run `devenv up`
  - Still show useful package information

## Technical Details

Package discovery:
1. Scans `.devenv/profile/bin/` directory
2. Resolves symlinks to Nix store paths
3. Parses package name and version from path format: `/nix/store/<hash>-<name>-<version>/...`
4. Groups multiple binaries from same package (e.g., go, gofmt, godoc)
5. Categorizes by detected language/tool type
```

**Step 3: Commit documentation**

```bash
git add internal/ui/help.go docs/PACKAGES_PANE.md
git commit -m "docs: add packages pane documentation

Update help text with 'p' key:
- Toggle between packages and services view
- Shows in narrow terminal mode

Add comprehensive user documentation:
- Features and layout modes
- Package information displayed
- Keyboard shortcuts
- Project type detection

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Completion

**Summary of changes:**

- Created `internal/packages/` package for package discovery
- Added `PackagesView` UI component
- Integrated adaptive layout based on terminal width (140 column threshold)
- Added 'p' key toggle for narrow terminals
- Updated focus cycling and mouse support
- Comprehensive test coverage (80+ tests)
- Documentation for users

**Testing:**

Run: `CGO_ENABLED=0 go test ./... -v`
Expected: 80+ tests, all passing

**Build:**

Run: `CGO_ENABLED=0 go build -o devdash ./...`
Expected: Successful build

**Next Steps:**

1. Merge feature branch to main
2. Manual testing with real devenv projects
3. Gather user feedback
4. Consider future enhancements from design doc
