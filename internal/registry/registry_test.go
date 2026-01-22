package registry

import (
	"os"
	"path/filepath"
	"strings"
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

func TestPath(t *testing.T) {
	tests := []struct {
		name       string
		xdgConfig  string
		wantSuffix string
	}{
		{
			name:       "with XDG_CONFIG_HOME set",
			xdgConfig:  "/custom/config",
			wantSuffix: "/custom/config/devdash/projects.yaml",
		},
		{
			name:       "without XDG_CONFIG_HOME",
			xdgConfig:  "",
			wantSuffix: ".config/devdash/projects.yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore env
			oldXDG := os.Getenv("XDG_CONFIG_HOME")
			defer os.Setenv("XDG_CONFIG_HOME", oldXDG)

			if tt.xdgConfig != "" {
				os.Setenv("XDG_CONFIG_HOME", tt.xdgConfig)
			} else {
				os.Unsetenv("XDG_CONFIG_HOME")
			}

			path := Path()
			if !strings.HasSuffix(path, tt.wantSuffix) {
				t.Errorf("Path() = %q, want suffix %q", path, tt.wantSuffix)
			}
		})
	}
}

func TestLoad(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(t *testing.T) string
		wantProjs int
		wantErr   bool
	}{
		{
			name: "nonexistent file returns empty registry",
			setup: func(t *testing.T) string {
				return filepath.Join(t.TempDir(), "nonexistent.yaml")
			},
			wantProjs: 0,
			wantErr:   false,
		},
		{
			name: "valid yaml file",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				path := filepath.Join(dir, "projects.yaml")
				content := `projects:
  - id: "abc123"
    path: "/test/path"
    name: "test"
    hidden: false
    last_active: 2026-01-21T00:00:00Z
`
				os.WriteFile(path, []byte(content), 0644)
				return path
			},
			wantProjs: 1,
			wantErr:   false,
		},
		{
			name: "invalid yaml returns error",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				path := filepath.Join(dir, "bad.yaml")
				os.WriteFile(path, []byte("invalid: [yaml: bad"), 0644)
				return path
			},
			wantProjs: 0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setup(t)
			reg, err := Load(path)

			if tt.wantErr {
				if err == nil {
					t.Fatal("Load() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Load() unexpected error: %v", err)
			}
			if len(reg.Projects) != tt.wantProjs {
				t.Errorf("Load() got %d projects, want %d", len(reg.Projects), tt.wantProjs)
			}
		})
	}
}

func TestSave(t *testing.T) {
	t.Run("saves registry to file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "subdir", "projects.yaml")

		reg := &Registry{
			Projects: []*Project{
				{
					ID:   "test123",
					Path: "/test/path",
					Name: "test",
				},
			},
		}

		err := Save(path, reg)
		if err != nil {
			t.Fatalf("Save() error: %v", err)
		}

		// Verify file exists
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Fatal("Save() did not create file")
		}

		// Load it back
		loaded, err := Load(path)
		if err != nil {
			t.Fatalf("Load() error: %v", err)
		}

		if len(loaded.Projects) != 1 {
			t.Errorf("loaded %d projects, want 1", len(loaded.Projects))
		}
		if loaded.Projects[0].ID != "test123" {
			t.Errorf("loaded ID %q, want %q", loaded.Projects[0].ID, "test123")
		}
	})
}

func TestRegistryAddProject(t *testing.T) {
	reg := &Registry{Projects: []*Project{}}

	p1 := reg.AddProject("/test/path1")
	if p1 == nil {
		t.Fatal("AddProject returned nil")
	}
	if len(reg.Projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(reg.Projects))
	}

	// Adding same path again should return existing
	p2 := reg.AddProject("/test/path1")
	if len(reg.Projects) != 1 {
		t.Fatalf("expected 1 project after duplicate add, got %d", len(reg.Projects))
	}
	if p1 != p2 {
		t.Error("AddProject should return same instance for duplicate path")
	}

	// Adding different path should add new
	p3 := reg.AddProject("/test/path2")
	if len(reg.Projects) != 2 {
		t.Fatalf("expected 2 projects, got %d", len(reg.Projects))
	}
	if p3 == p1 {
		t.Error("AddProject should create new instance for different path")
	}
}

func TestRegistryFindByPath(t *testing.T) {
	reg := &Registry{
		Projects: []*Project{
			{Path: "/test/path1", Name: "test1"},
			{Path: "/test/path2", Name: "test2"},
		},
	}

	found := reg.FindByPath("/test/path1")
	if found == nil {
		t.Fatal("FindByPath returned nil for existing path")
	}
	if found.Name != "test1" {
		t.Errorf("found project name %q, want %q", found.Name, "test1")
	}

	notFound := reg.FindByPath("/nonexistent")
	if notFound != nil {
		t.Error("FindByPath should return nil for nonexistent path")
	}
}

func TestRegistryRemoveProject(t *testing.T) {
	reg := &Registry{
		Projects: []*Project{
			{Path: "/test/path1"},
			{Path: "/test/path2"},
			{Path: "/test/path3"},
		},
	}

	// Remove existing
	removed := reg.RemoveProject("/test/path2")
	if !removed {
		t.Error("RemoveProject should return true for existing project")
	}
	if len(reg.Projects) != 2 {
		t.Fatalf("expected 2 projects after removal, got %d", len(reg.Projects))
	}
	if reg.FindByPath("/test/path2") != nil {
		t.Error("removed project still exists")
	}

	// Remove nonexistent
	removed = reg.RemoveProject("/nonexistent")
	if removed {
		t.Error("RemoveProject should return false for nonexistent project")
	}
	if len(reg.Projects) != 2 {
		t.Fatalf("expected 2 projects after failed removal, got %d", len(reg.Projects))
	}
}

func TestRegistryToggleHidden(t *testing.T) {
	reg := &Registry{
		Projects: []*Project{
			{Path: "/test/path1", Hidden: false},
		},
	}

	// Toggle to true
	toggled := reg.ToggleHidden("/test/path1")
	if !toggled {
		t.Error("ToggleHidden should return true for existing project")
	}
	p := reg.FindByPath("/test/path1")
	if !p.Hidden {
		t.Error("project should be hidden after toggle")
	}

	// Toggle back to false
	reg.ToggleHidden("/test/path1")
	if p.Hidden {
		t.Error("project should not be hidden after second toggle")
	}

	// Toggle nonexistent
	toggled = reg.ToggleHidden("/nonexistent")
	if toggled {
		t.Error("ToggleHidden should return false for nonexistent project")
	}
}
