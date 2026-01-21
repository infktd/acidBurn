# Rebranding Guide for acidBurn

This document outlines all the places where the "acidBurn" name appears and needs to be changed if you decide to rebrand the application.

## Summary

**Rebranding Difficulty:** ⭐⭐☆☆☆ (Easy - About 30 minutes of find/replace work)

The name appears in ~23 files, but most changes are straightforward find/replace operations. The most critical changes are:
1. Go module name (requires `go mod` update)
2. Config directory path (breaks existing user configs - needs migration)
3. ASCII art in splash screen
4. Documentation

---

## Critical Changes (Required)

### 1. Go Module Name
**File:** `go.mod:1`

```go
// Current:
module github.com/infktd/acidburn

// Change to:
module github.com/infktd/<new-name>
```

**Impact:** This requires updating all import statements.

**Steps:**
```bash
# 1. Update go.mod
sed -i '' 's/github.com\/infktd\/acidburn/github.com\/infktd\/<new-name>/g' go.mod

# 2. Update all import statements
find . -name "*.go" -type f -exec sed -i '' 's/github.com\/infktd\/acidburn/github.com\/infktd\/<new-name>/g' {} +

# 3. Run go mod tidy
go mod tidy
```

### 2. Config Directory Path
**File:** `internal/config/config.go:12`

```go
// Current:
configDir  = "acidburn"

// Change to:
configDir  = "<new-name>"
```

**Impact:** ⚠️ **BREAKING CHANGE** - Existing users' configs at `~/.config/acidburn/` will no longer be found.

**Migration Strategy:**
Add migration code to check for old config and copy to new location:
```go
// In config.Path() or main.go initialization
oldPath := filepath.Join(configHome, "acidburn")
newPath := filepath.Join(configHome, "<new-name>")

if _, err := os.Stat(oldPath); err == nil {
    // Old config exists, migrate it
    if _, err := os.Stat(newPath); os.IsNotExist(err) {
        log.Printf("Migrating config from %s to %s", oldPath, newPath)
        if err := os.Rename(oldPath, newPath); err != nil {
            log.Printf("Warning: Could not migrate config: %v", err)
        }
    }
}
```

### 3. ASCII Art / Splash Screen
**File:** `internal/ui/splash.go:10-47`

All 5 ASCII art variants spell out "ACIDBURN". You'll need to:
1. Generate new ASCII art for the new name
2. Update all variants (default, block, small, minimal, hacker)

**Tools for generating ASCII art:**
- http://patorjk.com/software/taag/
- https://www.ascii-art-generator.org/

### 4. Binary Name in Build Commands
**Files:**
- `docs/QA_TEST_PLAN.md` - Multiple references to `./acidburn`
- `docs/PRE_RELEASE_TODO.md` - Multiple references to `./acidburn`
- `README.md` - Build and run examples

**Find/Replace:**
```bash
# Update build commands
find docs/ README.md -type f -exec sed -i '' 's/acidburn/<new-name>/g' {} +
```

---

## Documentation Changes (Recommended)

### 5. Config Path References
**Files:**
- `README.md` - 3 references to `~/.config/acidburn/`
- `CHANGELOG.md` - 2 references to `~/.config/acidburn/`
- `docs/plans/*.md` - Multiple references

**Find/Replace:**
```bash
find . -name "*.md" -type f -exec sed -i '' 's/acidburn/<new-name>/g' {} +
```

### 6. Package Comments
**Files:**
- `internal/config/config.go:1` - "Package config handles acidBurn configuration..."
- `internal/config/types.go:8` - "Config represents the acidBurn configuration."

**Find/Replace:**
```bash
find internal/ -name "*.go" -type f -exec sed -i '' 's/acidBurn/<NewName>/g' {} +
find internal/ -name "*.go" -type f -exec sed -i '' 's/acidburn/<new-name>/g' {} +
```

### 7. Project References in Code
**File:** `README.md:183-185`

References to "acidBurn" in prose descriptions of how the tool works.

---

## Optional Changes (Nice to Have)

### 8. Theme Name
**File:** `internal/ui/theme.go:20-29`

The default theme is called "acid-green" which ties to the acidBurn brand:

```go
// Current:
"acid-green": {
    Name:       "acid-green",
    Primary:    lipgloss.Color("#39FF14"),
    // ...
}

// Consider renaming to something neutral or brand-aligned
```

### 9. Git History
**Files:** All commit messages with "acidBurn" or "acidburn"

Git history is permanent, but you could:
1. Keep history as-is (recommended)
2. Add a note in README about the rename
3. Create a git tag marking the rebrand point

---

## Files Requiring Changes

Here's the complete list of 23 files with the name:

### Core Code (9 files)
1. `go.mod` - Module name
2. `main.go` - Import statements
3. `internal/config/config.go` - Config dir name, package comment
4. `internal/config/types.go` - Struct comment
5. `internal/ui/model.go` - Import statements
6. `internal/ui/settings.go` - Import statements
7. `internal/ui/splash.go` - ASCII art (5 variants)
8. `internal/ui/model_test.go` - Import statements
9. `internal/ui/settings_test.go` - Import statements

### Documentation (8 files)
10. `README.md` - Build commands, config paths, prose
11. `CHANGELOG.md` - Config paths, release notes
12. `docs/QA_TEST_PLAN.md` - Build/run commands
13. `docs/PRE_RELEASE_TODO.md` - Build/run commands
14. `docs/plans/2026-01-20-acidburn-design.md` - Design doc
15. `docs/plans/2026-01-20-acidburn-implementation.md` - Implementation doc
16. `STYLING_IDEAS.md` - References in styling notes

### Supporting Files (6 files)
17. `internal/ui/theme.go` - "acid-green" theme name
18. `internal/registry/types.go` - Import statements
19. `internal/registry/registry.go` - Import statements
20. `internal/notify/notify.go` - Import statements
21. `.devenv.flake.nix` - Auto-generated (will update on next `devenv up`)
22. `.gitignore` - May have references

---

## Automated Rebranding Script

Save this as `scripts/rebrand.sh`:

```bash
#!/bin/bash

# Rebranding script for acidBurn
# Usage: ./scripts/rebrand.sh <new-name>

set -e

if [ -z "$1" ]; then
    echo "Usage: $0 <new-name>"
    echo "Example: $0 devstack"
    exit 1
fi

NEW_NAME=$1
NEW_NAME_CAPS=$(echo "$NEW_NAME" | tr '[:lower:]' '[:upper:]')
NEW_NAME_TITLE=$(echo "$NEW_NAME" | sed 's/.*/\u&/')

echo "==> Rebranding acidBurn to $NEW_NAME"
echo ""

# 1. Go module and imports
echo "1. Updating Go module and imports..."
sed -i '' "s/github.com\/infktd\/acidburn/github.com\/infktd\/$NEW_NAME/g" go.mod
find . -name "*.go" -type f -exec sed -i '' "s/github.com\/infktd\/acidburn/github.com\/infktd\/$NEW_NAME/g" {} +

# 2. Config directory
echo "2. Updating config directory..."
sed -i '' "s/configDir  = \"acidburn\"/configDir  = \"$NEW_NAME\"/g" internal/config/config.go

# 3. Documentation
echo "3. Updating documentation..."
find . -name "*.md" -type f -exec sed -i '' "s/acidburn/$NEW_NAME/g" {} +
find . -name "*.md" -type f -exec sed -i '' "s/acidBurn/$NEW_NAME_TITLE/g" {} +
find . -name "*.md" -type f -exec sed -i '' "s/ACIDBURN/$NEW_NAME_CAPS/g" {} +

# 4. Package comments in Go files
echo "4. Updating Go package comments..."
find internal/ -name "*.go" -type f -exec sed -i '' "s/acidBurn/$NEW_NAME_TITLE/g" {} +

# 5. Build and test
echo "5. Running go mod tidy..."
go mod tidy

echo ""
echo "==> Rebranding complete!"
echo ""
echo "⚠️  MANUAL STEPS REQUIRED:"
echo "1. Update ASCII art in internal/ui/splash.go (5 variants)"
echo "2. Consider renaming 'acid-green' theme in internal/ui/theme.go"
echo "3. Test build: go build -o $NEW_NAME ."
echo "4. Update any CI/CD configs (if they exist)"
echo "5. Update GitHub repository name"
echo "6. Add migration code for old config path (see REBRANDING_GUIDE.md)"
echo ""
```

Make it executable:
```bash
chmod +x scripts/rebrand.sh
```

Run it:
```bash
./scripts/rebrand.sh devstack
```

---

## Post-Rebrand Checklist

After rebranding, verify:

- [ ] `go build` succeeds
- [ ] `go test ./...` passes
- [ ] Config loads from new path
- [ ] ASCII art looks correct
- [ ] Documentation updated
- [ ] README build commands work
- [ ] Import paths all updated
- [ ] No broken references in comments

---

## Breaking Changes for Users

**Config Migration:**
Users upgrading from acidBurn to the new name will need to either:
1. Manually move `~/.config/acidburn/` to `~/.config/<new-name>/`
2. Or: Add migration code (see Critical Change #2 above)

**Recommendation:** Add migration code + release note explaining the change.

---

## Estimated Time

- Automated changes (script): **5 minutes**
- ASCII art update: **10 minutes**
- Manual verification: **10 minutes**
- Migration code: **5 minutes**
- Testing: **10 minutes**

**Total:** ~40 minutes

---

## Conclusion

Rebranding is straightforward because:
✅ The name is well-isolated (mostly in config paths and imports)
✅ No hardcoded strings in business logic
✅ Automated script handles 80% of the work
✅ Only ASCII art requires manual creative work

The main consideration is the **breaking change** for existing users' config paths, which can be handled gracefully with migration code.
