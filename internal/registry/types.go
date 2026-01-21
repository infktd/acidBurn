package registry

import (
	"crypto/sha256"
	"encoding/hex"
	"net"
	"os"
	"path/filepath"
	"time"
)

// ProjectState represents the current state of a project.
type ProjectState int

const (
	StateIdle ProjectState = iota
	StateRunning
	StateDegraded
	StateStale
	StateMissing
)

func (s ProjectState) String() string {
	switch s {
	case StateIdle:
		return "idle"
	case StateRunning:
		return "running"
	case StateDegraded:
		return "degraded"
	case StateStale:
		return "stale"
	case StateMissing:
		return "missing"
	default:
		return "unknown"
	}
}

// Project represents a devenv project in the registry.
type Project struct {
	ID         string    `yaml:"id"`
	Path       string    `yaml:"path"`
	Name       string    `yaml:"name"`
	Hidden     bool      `yaml:"hidden"`
	LastActive time.Time `yaml:"last_active"`
}

// NewProject creates a new Project from a path.
func NewProject(path string) *Project {
	// Generate ID from path hash
	hash := sha256.Sum256([]byte(path))
	id := hex.EncodeToString(hash[:8])

	return &Project{
		ID:         id,
		Path:       path,
		Name:       filepath.Base(path),
		Hidden:     false,
		LastActive: time.Now(),
	}
}

// SocketPath returns the path to the process-compose socket.
// devenv creates a symlink at .devenv/run pointing to /run/user/$UID/devenv-$HASH
func (p *Project) SocketPath() string {
	return filepath.Join(p.Path, ".devenv", "run", "pc.sock")
}

// DetectState checks the project's current state.
func (p *Project) DetectState() ProjectState {
	// Check if path exists
	if _, err := os.Stat(p.Path); os.IsNotExist(err) {
		return StateMissing
	}

	socketPath := p.SocketPath()

	// Try to connect to socket
	conn, err := net.Dial("unix", socketPath)
	if err == nil {
		conn.Close()
		return StateRunning // TODO: Check if degraded via API
	}

	// Check if socket file exists (stale)
	if _, err := os.Stat(socketPath); err == nil {
		return StateStale
	}

	return StateIdle
}
