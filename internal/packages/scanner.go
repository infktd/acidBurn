package packages

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// nixStoreRegex parses Nix store path components.
// Pattern explanation:
// - ^([a-z0-9]+) - matches the hash at the start (group 1)
// - -(.+?) - matches the package name, non-greedy to allow backtracking (group 2)
// - (?:-([0-9]+[0-9.\-a-z]*))? - optionally matches version starting with a digit (group 3)
//   * Must start with a digit to distinguish from hyphenated package names
//   * Can contain digits, dots, hyphens, and letters (e.g., "1.21.5", "0.14.0-unstable")
//   * Optional group allows packages without versions
//
// Examples:
//   - "abc123-go-1.21.5" → hash="abc123", name="go", version="1.21.5"
//   - "xyz789-gcc-arm-embedded-13.2.1" → hash="xyz789", name="gcc-arm-embedded", version="13.2.1"
//   - "def456-bash" → hash="def456", name="bash", version=""
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
// Uses exact matches for core tools and explicit checks for known variants
// to avoid false positives (e.g., "golang" or "google-chrome" shouldn't match "go").
//
// Examples:
//   - categorizePackage("go") → "Go"
//   - categorizePackage("python3") → "Python"
//   - categorizePackage("node") → "Node.js"
//   - categorizePackage("golang") → "Other" (not the Go toolchain)
//   - categorizePackage("python-dotenv") → "Other" (Python package, not Python itself)
func categorizePackage(name string) string {
	lowerName := strings.ToLower(name)

	// Go packages - exact match for "go" + known Go toolchain binaries
	if lowerName == "go" ||
		lowerName == "gopls" ||
		lowerName == "gofmt" ||
		lowerName == "godoc" ||
		lowerName == "goimports" ||
		lowerName == "golangci-lint" {
		return "Go"
	}

	// Python packages - exact match for python/python2/python3 + known Python tools
	if lowerName == "python" ||
		lowerName == "python2" ||
		lowerName == "python3" ||
		lowerName == "pip" ||
		lowerName == "pip2" ||
		lowerName == "pip3" ||
		lowerName == "pytest" ||
		lowerName == "poetry" ||
		lowerName == "pipenv" {
		return "Python"
	}

	// Node.js packages - exact matches for Node ecosystem tools
	if lowerName == "node" ||
		lowerName == "nodejs" ||
		lowerName == "npm" ||
		lowerName == "npx" ||
		lowerName == "yarn" ||
		lowerName == "pnpm" {
		return "Node.js"
	}

	// Rust packages - exact matches for Rust toolchain
	if lowerName == "cargo" ||
		lowerName == "rustc" ||
		lowerName == "rustup" ||
		lowerName == "rustfmt" {
		return "Rust"
	}

	// C/C++ compilers - exact matches
	if lowerName == "gcc" ||
		lowerName == "g++" ||
		lowerName == "clang" ||
		lowerName == "clang++" ||
		lowerName == "make" ||
		lowerName == "cmake" {
		return "C/C++"
	}

	// Ruby - exact matches
	if lowerName == "ruby" ||
		lowerName == "gem" ||
		lowerName == "bundle" ||
		lowerName == "bundler" ||
		lowerName == "rake" {
		return "Ruby"
	}

	// Java - exact matches
	if lowerName == "java" ||
		lowerName == "javac" ||
		lowerName == "maven" ||
		lowerName == "mvn" ||
		lowerName == "gradle" {
		return "Java"
	}

	return "Other"
}

// binaryInfo holds parsed information about a binary.
type binaryInfo struct {
	name    string
	version string
	hash    string
	path    string
}

// Scan discovers packages in a devenv project.
// Returns list of packages found, or empty list if directory doesn't exist.
func Scan(projectPath string) ([]Package, error) {
	binDir := filepath.Join(projectPath, ".devenv", "profile", "bin")

	// Check if bin directory exists
	if _, err := os.Stat(binDir); os.IsNotExist(err) {
		return []Package{}, nil
	}

	var binaries []binaryInfo

	// Read all files in bin directory
	err := filepath.WalkDir(binDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		// Skip directories and the bin directory itself
		if d.IsDir() {
			return nil
		}

		// Resolve symlink to find Nix store path
		target, err := os.Readlink(path)
		if err != nil {
			// Not a symlink, use the file path itself
			target = path
		}

		// Parse Nix store path
		name, version, hash := parseNixStorePath(target)
		if name == "" {
			// Couldn't parse, skip
			return nil
		}

		binaries = append(binaries, binaryInfo{
			name:    name,
			version: version,
			hash:    hash,
			path:    path,
		})

		return nil
	})

	if err != nil {
		return []Package{}, err
	}

	// Group binaries by package
	packages := groupByPackage(binaries)

	return packages, nil
}

// groupByPackage groups binaries from the same Nix package.
// Binaries with the same hash+name+version belong to the same package.
func groupByPackage(binaries []binaryInfo) []Package {
	// Stub for now - will be implemented in Task 5
	return []Package{}
}
