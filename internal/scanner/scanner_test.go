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
