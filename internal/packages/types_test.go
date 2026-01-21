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
