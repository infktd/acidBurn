package packages

import (
	"testing"
	"time"
)

func TestPackageStructure(t *testing.T) {
	tests := []struct {
		name    string
		pkg     Package
		wantPkg Package
	}{
		{
			name: "Go package",
			pkg: Package{
				Name:    "go",
				Version: "1.21.5",
				Type:    "Go",
				Binary:  "/nix/store/abc-go-1.21.5/bin/go",
			},
			wantPkg: Package{
				Name:    "go",
				Version: "1.21.5",
				Type:    "Go",
				Binary:  "/nix/store/abc-go-1.21.5/bin/go",
			},
		},
		{
			name: "Python package",
			pkg: Package{
				Name:    "python3",
				Version: "3.11.7",
				Type:    "Python",
				Binary:  "/usr/bin/python3",
			},
			wantPkg: Package{
				Name:    "python3",
				Version: "3.11.7",
				Type:    "Python",
				Binary:  "/usr/bin/python3",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.pkg.Name != tt.wantPkg.Name {
				t.Errorf("Name: got %q, want %q", tt.pkg.Name, tt.wantPkg.Name)
			}
			if tt.pkg.Version != tt.wantPkg.Version {
				t.Errorf("Version: got %q, want %q", tt.pkg.Version, tt.wantPkg.Version)
			}
			if tt.pkg.Type != tt.wantPkg.Type {
				t.Errorf("Type: got %q, want %q", tt.pkg.Type, tt.wantPkg.Type)
			}
			if tt.pkg.Binary != tt.wantPkg.Binary {
				t.Errorf("Binary: got %q, want %q", tt.pkg.Binary, tt.wantPkg.Binary)
			}
		})
	}
}

func TestPackageInfoStructure(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name     string
		info     PackageInfo
		wantInfo PackageInfo
	}{
		{
			name: "single package",
			info: PackageInfo{
				ProjectPath: "/home/user/project",
				Packages:    []Package{{Name: "go", Version: "1.21.5", Type: "Go", Binary: "/usr/bin/go"}},
				LastScanned: now,
			},
			wantInfo: PackageInfo{
				ProjectPath: "/home/user/project",
				Packages:    []Package{{Name: "go", Version: "1.21.5", Type: "Go", Binary: "/usr/bin/go"}},
				LastScanned: now,
			},
		},
		{
			name: "multiple packages",
			info: PackageInfo{
				ProjectPath: "/home/user/multi-project",
				Packages: []Package{
					{Name: "go", Version: "1.21.5", Type: "Go", Binary: "/usr/bin/go"},
					{Name: "python3", Version: "3.11.7", Type: "Python", Binary: "/usr/bin/python3"},
				},
				LastScanned: now,
			},
			wantInfo: PackageInfo{
				ProjectPath: "/home/user/multi-project",
				Packages: []Package{
					{Name: "go", Version: "1.21.5", Type: "Go", Binary: "/usr/bin/go"},
					{Name: "python3", Version: "3.11.7", Type: "Python", Binary: "/usr/bin/python3"},
				},
				LastScanned: now,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.info.ProjectPath != tt.wantInfo.ProjectPath {
				t.Errorf("ProjectPath: got %q, want %q", tt.info.ProjectPath, tt.wantInfo.ProjectPath)
			}
			if len(tt.info.Packages) != len(tt.wantInfo.Packages) {
				t.Errorf("Packages length: got %d, want %d", len(tt.info.Packages), len(tt.wantInfo.Packages))
			}
			for i, pkg := range tt.info.Packages {
				wantPkg := tt.wantInfo.Packages[i]
				if pkg.Name != wantPkg.Name {
					t.Errorf("Package[%d].Name: got %q, want %q", i, pkg.Name, wantPkg.Name)
				}
				if pkg.Version != wantPkg.Version {
					t.Errorf("Package[%d].Version: got %q, want %q", i, pkg.Version, wantPkg.Version)
				}
				if pkg.Type != wantPkg.Type {
					t.Errorf("Package[%d].Type: got %q, want %q", i, pkg.Type, wantPkg.Type)
				}
				if pkg.Binary != wantPkg.Binary {
					t.Errorf("Package[%d].Binary: got %q, want %q", i, pkg.Binary, wantPkg.Binary)
				}
			}
			if !tt.info.LastScanned.Equal(tt.wantInfo.LastScanned) {
				t.Errorf("LastScanned: got %v, want %v", tt.info.LastScanned, tt.wantInfo.LastScanned)
			}
		})
	}
}

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
		{
			name:            "hyphenated package name with version",
			path:            "/nix/store/abc123-gcc-arm-embedded-13.2.1/bin/arm-none-eabi-gcc",
			expectedPkg:     "gcc-arm-embedded",
			expectedVersion: "13.2.1",
			expectedHash:    "abc123",
		},
		{
			name:            "hyphenated package name without version",
			path:            "/nix/store/def456-some-cool-tool/bin/tool",
			expectedPkg:     "some-cool-tool",
			expectedVersion: "",
			expectedHash:    "def456",
		},
		{
			name:            "multiple hyphens in name and version",
			path:            "/nix/store/xyz789-node-packages-1.2.3-alpha/bin/npm",
			expectedPkg:     "node-packages",
			expectedVersion: "1.2.3-alpha",
			expectedHash:    "xyz789",
		},
		{
			name:            "empty string",
			path:            "",
			expectedPkg:     "",
			expectedVersion: "",
			expectedHash:    "",
		},
		{
			name:            "path without store directory",
			path:            "/usr/bin/go",
			expectedPkg:     "",
			expectedVersion: "",
			expectedHash:    "",
		},
		{
			name:            "path with store but no following component",
			path:            "/nix/store/",
			expectedPkg:     "",
			expectedVersion: "",
			expectedHash:    "",
		},
		{
			name:            "malformed store component - only hash",
			path:            "/nix/store/abc123/bin/tool",
			expectedPkg:     "abc123",
			expectedVersion: "",
			expectedHash:    "",
		},
		{
			name:            "malformed store component - no dashes",
			path:            "/nix/store/invalidformat/bin/tool",
			expectedPkg:     "invalidformat",
			expectedVersion: "",
			expectedHash:    "",
		},
		{
			name:            "version starting with zero",
			path:            "/nix/store/abc123-package-0.1.2/bin/pkg",
			expectedPkg:     "package",
			expectedVersion: "0.1.2",
			expectedHash:    "abc123",
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
