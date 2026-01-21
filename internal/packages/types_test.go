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
