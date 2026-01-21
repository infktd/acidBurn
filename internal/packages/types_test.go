package packages

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseNixStorePath(t *testing.T) {
	tests := []struct {
		name            string
		path            string
		expectedPkg     string
		expectedVersion string
		expectedHash    string
	}{
		{
			name:            "standard format",
			path:            "/nix/store/abc123-go-1.21.5/bin/go",
			expectedPkg:     "go",
			expectedVersion: "1.21.5",
			expectedHash:    "abc123",
		},
		{
			name:            "python with dots",
			path:            "/nix/store/xyz789-python3-3.11.7/bin/python3",
			expectedPkg:     "python3",
			expectedVersion: "3.11.7",
			expectedHash:    "xyz789",
		},
		{
			name:            "package with complex name",
			path:            "/nix/store/def456-gopls-0.14.0-unstable/bin/gopls",
			expectedPkg:     "gopls",
			expectedVersion: "0.14.0-unstable",
			expectedHash:    "def456",
		},
		{
			name:            "no version",
			path:            "/nix/store/ghi789-bash/bin/bash",
			expectedPkg:     "bash",
			expectedVersion: "",
			expectedHash:    "ghi789",
		},
		{
			name:            "hyphenated package name with version",
			path:            "/nix/store/abc123-gcc-arm-embedded-13.2.1/bin/arm-none-eabi-gcc",
			expectedPkg:     "gcc-arm-embedded",
			expectedVersion: "13.2.1",
			expectedHash:    "abc123",
		},
		{
			name:            "hyphenated package name without version",
			path:            "/nix/store/def456-some-cool-tool/bin/tool",
			expectedPkg:     "some-cool-tool",
			expectedVersion: "",
			expectedHash:    "def456",
		},
		{
			name:            "multiple hyphens in name and version",
			path:            "/nix/store/xyz789-node-packages-1.2.3-alpha/bin/npm",
			expectedPkg:     "node-packages",
			expectedVersion: "1.2.3-alpha",
			expectedHash:    "xyz789",
		},
		{
			name:            "empty string",
			path:            "",
			expectedPkg:     "",
			expectedVersion: "",
			expectedHash:    "",
		},
		{
			name:            "path without store directory",
			path:            "/usr/bin/go",
			expectedPkg:     "",
			expectedVersion: "",
			expectedHash:    "",
		},
		{
			name:            "path with store but no following component",
			path:            "/nix/store/",
			expectedPkg:     "",
			expectedVersion: "",
			expectedHash:    "",
		},
		{
			name:            "malformed store component - only hash",
			path:            "/nix/store/abc123/bin/tool",
			expectedPkg:     "abc123",
			expectedVersion: "",
			expectedHash:    "",
		},
		{
			name:            "malformed store component - no dashes",
			path:            "/nix/store/invalidformat/bin/tool",
			expectedPkg:     "invalidformat",
			expectedVersion: "",
			expectedHash:    "",
		},
		{
			name:            "version starting with zero",
			path:            "/nix/store/abc123-package-0.1.2/bin/pkg",
			expectedPkg:     "package",
			expectedVersion: "0.1.2",
			expectedHash:    "abc123",
		},
		// Regex fallback tests (exercises lines 40-45)
		{
			name:            "fallback_just_hash_and_name",
			path:            "/nix/store/abc123-tool/bin/tool",
			expectedPkg:     "tool",
			expectedVersion: "",
			expectedHash:    "abc123",
		},
		{
			name:            "fallback_special_chars_in_name",
			path:            "/nix/store/xyz789-my_special.tool/bin/tool",
			expectedPkg:     "my_special.tool",
			expectedVersion: "",
			expectedHash:    "xyz789",
		},
		{
			name:            "fallback_version_like_but_not_version",
			path:            "/nix/store/abc-tool-v2/bin/tool",
			expectedPkg:     "tool-v2",
			expectedVersion: "",
			expectedHash:    "abc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pkg, version, hash := parseNixStorePath(tt.path)
			if pkg != tt.expectedPkg {
				t.Errorf("expected package %q, got %q", tt.expectedPkg, pkg)
			}
			if version != tt.expectedVersion {
				t.Errorf("expected version %q, got %q", tt.expectedVersion, version)
			}
			if hash != tt.expectedHash {
				t.Errorf("expected hash %q, got %q", tt.expectedHash, hash)
			}
		})
	}
}

func TestCategorizePackage(t *testing.T) {
	tests := []struct {
		name         string
		packageName  string
		expectedType string
	}{
		// Go toolchain
		{"go_lowercase", "go", "Go"},
		{"go_uppercase", "GO", "Go"},
		{"go_mixedcase", "Go", "Go"},
		{"gofmt", "gofmt", "Go"},
		{"gopls", "gopls", "Go"},
		{"godoc", "godoc", "Go"},
		{"goimports", "goimports", "Go"},
		{"golangci_lint", "golangci-lint", "Go"},

		// Go false positives (should be Other)
		{"golang_not_go_toolchain", "golang", "Other"},
		{"go_tools_prefix", "go-tools", "Other"},
		{"google_chrome_prefix", "google-chrome", "Other"},
		{"gopher_prefix", "gopher", "Other"},

		// Python toolchain
		{"python", "python", "Python"},
		{"python2", "python2", "Python"},
		{"python3", "python3", "Python"},
		{"python3_uppercase", "PYTHON3", "Python"},
		{"pip", "pip", "Python"},
		{"pip2", "pip2", "Python"},
		{"pip3", "pip3", "Python"},
		{"pytest", "pytest", "Python"},
		{"poetry", "poetry", "Python"},
		{"pipenv", "pipenv", "Python"},

		// Python false positives (should be Other)
		{"python_dotenv_package", "python-dotenv", "Other"},
		{"python_requests_package", "python-requests", "Other"},
		{"pythonista", "pythonista", "Other"},

		// Node.js toolchain
		{"node", "node", "Node.js"},
		{"nodejs", "nodejs", "Node.js"},
		{"node_uppercase", "NODE", "Node.js"},
		{"npm", "npm", "Node.js"},
		{"npx", "npx", "Node.js"},
		{"yarn", "yarn", "Node.js"},
		{"pnpm", "pnpm", "Node.js"},

		// Node.js false positives (should be Other)
		{"node_fetch_package", "node-fetch", "Other"},
		{"node_sass_package", "node-sass", "Other"},
		{"nodemon", "nodemon", "Other"},

		// Rust toolchain
		{"cargo", "cargo", "Rust"},
		{"rustc", "rustc", "Rust"},
		{"rustup", "rustup", "Rust"},
		{"rustfmt", "rustfmt", "Rust"},

		// C/C++ toolchain
		{"gcc", "gcc", "C/C++"},
		{"g++", "g++", "C/C++"},
		{"clang", "clang", "C/C++"},
		{"clang++", "clang++", "C/C++"},
		{"make", "make", "C/C++"},
		{"cmake", "cmake", "C/C++"},

		// Ruby toolchain
		{"ruby", "ruby", "Ruby"},
		{"gem", "gem", "Ruby"},
		{"bundle", "bundle", "Ruby"},
		{"bundler", "bundler", "Ruby"},
		{"rake", "rake", "Ruby"},

		// Java toolchain
		{"java", "java", "Java"},
		{"javac", "javac", "Java"},
		{"maven", "maven", "Java"},
		{"mvn", "mvn", "Java"},
		{"gradle", "gradle", "Java"},

		// Other/Unknown packages
		{"unknown_package", "foobar", "Other"},
		{"bash", "bash", "Other"},
		{"zsh", "zsh", "Other"},
		{"git", "git", "Other"},
		{"docker", "docker", "Other"},
		{"kubectl", "kubectl", "Other"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typ := categorizePackage(tt.packageName)
			if typ != tt.expectedType {
				t.Errorf("categorizePackage(%q) = %q, want %q", tt.packageName, typ, tt.expectedType)
			}
		})
	}
}

func TestScan(t *testing.T) {
	// Create temporary test directory structure
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "testproject")
	binDir := filepath.Join(projectDir, ".devenv", "profile", "bin")

	err := os.MkdirAll(binDir, 0755)
	if err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	// Create test binary files (regular files for testing, not actual symlinks)
	testBinaries := []string{"go", "python3", "node"}
	for _, bin := range testBinaries {
		binPath := filepath.Join(binDir, bin)
		err := os.WriteFile(binPath, []byte("#!/bin/sh\necho test"), 0755)
		if err != nil {
			t.Fatalf("failed to create test binary %s: %v", bin, err)
		}
	}

	// Run scan
	packages, err := Scan(projectDir)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	// Should not error - package count validation will be in Task 5 when grouping is implemented
	_ = packages
}

func TestScanMissingDirectory(t *testing.T) {
	packages, err := Scan("/nonexistent/path")

	// Should not error, just return empty list
	if err != nil {
		t.Errorf("expected no error for missing directory, got %v", err)
	}
	if len(packages) != 0 {
		t.Errorf("expected 0 packages for missing directory, got %d", len(packages))
	}
}
