package registry

import (
	"crypto/sha256"
	"encoding/hex"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/infktd/devdash/internal/compose"
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
		// Socket is reachable, query API to check service states
		return p.checkServiceStates()
	}

	// Check if socket file exists (stale)
	if _, err := os.Stat(socketPath); err == nil {
		return StateStale
	}

	return StateIdle
}

// checkServiceStates queries the compose API to determine if the project is running or degraded.
func (p *Project) checkServiceStates() ProjectState {
	client := compose.NewClient(p.SocketPath())
	if err := client.Connect(); err != nil {
		// Can't connect, consider it idle
		return StateIdle
	}

	status, err := client.GetStatus()
	if err != nil {
		// API error, assume running since socket was reachable
		return StateRunning
	}

	// Count running and total processes
	runningCount := 0
	totalCount := len(status.Processes)

	for _, proc := range status.Processes {
		if proc.IsRunning {
			runningCount++
		}
	}

	// If no processes, consider idle
	if totalCount == 0 {
		return StateIdle
	}

	// If all processes running, fully operational
	if runningCount == totalCount {
		return StateRunning
	}

	// If some processes running, degraded
	if runningCount > 0 {
		return StateDegraded
	}

	// No processes running but socket exists
	return StateStale
}

// Repair cleans up stale socket files and symlinks.
func (p *Project) Repair() error {
	runDir := filepath.Join(p.Path, ".devenv", "run")

	// Remove the entire run directory (contains socket and symlink)
	if err := os.RemoveAll(runDir); err != nil {
		return err
	}

	return nil
}
