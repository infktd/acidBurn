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
