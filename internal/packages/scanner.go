package packages

import (
	"path/filepath"
	"regexp"
	"strings"
)

// parseNixStorePath extracts package name, version, and hash from a Nix store path.
// Format: /nix/store/<hash>-<name>-<version>/...
// Returns: (name, version, hash)
func parseNixStorePath(path string) (string, string, string) {
	// Get the store directory component
	// e.g., /nix/store/abc123-go-1.21.5/bin/go -> abc123-go-1.21.5
	parts := strings.Split(filepath.Clean(path), string(filepath.Separator))

	var storeComponent string
	for i, part := range parts {
		if part == "store" && i+1 < len(parts) {
			storeComponent = parts[i+1]
			break
		}
	}

	if storeComponent == "" {
		return "", "", ""
	}

	// Parse: <hash>-<name>-<version>
	// Regex: (<hash>)-(<name>)-(<version>) where version is optional
	re := regexp.MustCompile(`^([a-z0-9]+)-(.+?)(?:-([0-9].*))?$`)
	matches := re.FindStringSubmatch(storeComponent)

	if len(matches) < 3 {
		// Fallback: just hash-name format
		dashIdx := strings.Index(storeComponent, "-")
		if dashIdx > 0 {
			return storeComponent[dashIdx+1:], "", storeComponent[:dashIdx]
		}
		return storeComponent, "", ""
	}

	hash := matches[1]
	name := matches[2]
	version := ""
	if len(matches) > 3 {
		version = matches[3]
	}

	return name, version, hash
}
