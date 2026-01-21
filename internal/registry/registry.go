// Package registry manages the project registry and discovery state.
package registry

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	registryDir  = "acidburn"
	registryFile = "projects.yaml"
)

// Registry holds discovered projects.
type Registry struct {
	Projects []*Project `yaml:"projects"`
}

// Path returns the default registry file path.
func Path() string {
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			// Fallback to current directory if home can't be determined
			configHome = ".config"
		} else {
			configHome = filepath.Join(home, ".config")
		}
	}
	return filepath.Join(configHome, registryDir, registryFile)
}

// Load reads the registry from path.
func Load(path string) (*Registry, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &Registry{Projects: []*Project{}}, nil
	}
	if err != nil {
		return nil, err
	}

	var reg Registry
	if err := yaml.Unmarshal(data, &reg); err != nil {
		return nil, err
	}
	return &reg, nil
}

// Save writes the registry to path.
func Save(path string, reg *Registry) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := yaml.Marshal(reg)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// AddProject adds a project if not already present.
func (r *Registry) AddProject(path string) *Project {
	for _, p := range r.Projects {
		if p.Path == path {
			return p
		}
	}
	p := NewProject(path)
	r.Projects = append(r.Projects, p)
	return p
}

// FindByPath returns a project by its path.
func (r *Registry) FindByPath(path string) *Project {
	for _, p := range r.Projects {
		if p.Path == path {
			return p
		}
	}
	return nil
}
