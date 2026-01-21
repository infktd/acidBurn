package packages

import (
	"path/filepath"
	"regexp"
	"strings"
)

// nixStoreRegex parses Nix store path components.
// Pattern explanation:
// - ^([a-z0-9]+) - matches the hash at the start
// - -(.+?) - matches the package name (non-greedy to allow backtracking)
// - (?:-([0-9]+[0-9.\-a-z]*))? - optionally matches version starting with a digit
var nixStoreRegex = regexp.MustCompile(`^([a-z0-9]+)-(.+?)(?:-([0-9]+[0-9.\-a-z]*))?$`)

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
	matches := nixStoreRegex.FindStringSubmatch(storeComponent)

	if len(matches) < 3 {
		// Fallback: just hash-name format
		firstDashIndex := strings.Index(storeComponent, "-")
		if firstDashIndex > 0 {
			return storeComponent[firstDashIndex+1:], "", storeComponent[:firstDashIndex]
		}
		return storeComponent, "", ""
	}

	hash := matches[1]
	name := matches[2]
	version := matches[3] // Will be empty string if group didn't match

	return name, version, hash
}

// categorizePackage determines the package type based on name patterns.
func categorizePackage(name string) string {
	lowerName := strings.ToLower(name)

	// Go packages
	if strings.HasPrefix(lowerName, "go") ||
		lowerName == "gopls" ||
		lowerName == "gofmt" ||
		lowerName == "godoc" {
		return "Go"
	}

	// Python packages
	if strings.Contains(lowerName, "python") ||
		lowerName == "pip" ||
		lowerName == "pytest" ||
		lowerName == "poetry" {
		return "Python"
	}

	// Node.js packages
	if lowerName == "node" ||
		lowerName == "nodejs" ||
		lowerName == "npm" ||
		lowerName == "npx" ||
		lowerName == "yarn" {
		return "Node.js"
	}

	// Rust packages
	if lowerName == "cargo" ||
		lowerName == "rustc" ||
		lowerName == "rustup" {
		return "Rust"
	}

	// C/C++ compilers
	if lowerName == "gcc" ||
		lowerName == "g++" ||
		lowerName == "clang" ||
		lowerName == "clang++" {
		return "C/C++"
	}

	// Ruby
	if lowerName == "ruby" ||
		lowerName == "gem" ||
		lowerName == "bundle" {
		return "Ruby"
	}

	// Java
	if lowerName == "java" ||
		lowerName == "javac" ||
		lowerName == "maven" ||
		lowerName == "gradle" {
		return "Java"
	}

	return "Other"
}
