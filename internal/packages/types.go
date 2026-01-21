package packages

import "time"

// Package represents a single installed package.
type Package struct {
	Name    string // e.g., "go", "python3", "node"
	Version string // e.g., "1.21.5", "3.11.7"
	Type    string // e.g., "Go", "Python", "Node.js"
	Binary  string // Full path to binary
}

// PackageInfo holds package information for a project.
type PackageInfo struct {
	ProjectPath string    // Absolute path to the project
	Packages    []Package // List of detected packages
	LastScanned time.Time // When the project was last scanned
}
