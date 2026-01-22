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
