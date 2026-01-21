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
