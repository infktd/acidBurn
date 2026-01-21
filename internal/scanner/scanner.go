// Package scanner discovers devenv.nix projects on the filesystem.
package scanner

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Directories to skip during scanning.
var excludedDirs = map[string]bool{
	"node_modules": true,
	".git":         true,
	".direnv":      true,
	"dist":         true,
	"target":       true,
	"vendor":       true,
	".venv":        true,
	"__pycache__":  true,
}

// Scan searches paths for directories containing devenv.nix.
// maxDepth limits how deep to recurse (1 = immediate children only).
func Scan(paths []string, maxDepth int) ([]string, error) {
	var projects []string
	seen := make(map[string]bool)

	for _, root := range paths {
		// Expand ~ if present
		if strings.HasPrefix(root, "~/") {
			home, _ := os.UserHomeDir()
			root = filepath.Join(home, root[2:])
		}

		err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil // Skip inaccessible paths
			}

			// Calculate depth relative to root
			rel, _ := filepath.Rel(root, path)
			depth := len(strings.Split(rel, string(os.PathSeparator)))
			if rel == "." {
				depth = 0
			}

			// For directories, skip if too deep
			if d.IsDir() {
				if depth > maxDepth {
					return fs.SkipDir
				}
				// Skip excluded directories
				if excludedDirs[d.Name()] {
					return fs.SkipDir
				}
				return nil
			}

			// For files, check depth of parent directory (depth - 1)
			parentDepth := depth - 1
			if parentDepth > maxDepth {
				return nil
			}

			// Check for devenv.nix
			if d.Name() == "devenv.nix" {
				projectPath := filepath.Dir(path)
				if !seen[projectPath] {
					seen[projectPath] = true
					projects = append(projects, projectPath)
				}
			}

			return nil
		})
		if err != nil {
			return projects, err
		}
	}

	return projects, nil
}
