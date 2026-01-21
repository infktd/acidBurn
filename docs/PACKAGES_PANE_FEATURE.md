# Packages Pane Feature

## Overview

Add a new pane to display package information for the currently selected project's devenv environment.

## Motivation

- Projects without services feel empty (only shows empty services table)
- Users want visibility into their devenv environment packages
- Helpful for debugging dependency versions and understanding the environment

## Proposed Design

### UI Layout

Add a 4th pane alongside PROJECTS, SERVICES, and LOGS:

```
┌─────────────┬──────────────────────────┐
│  PROJECTS   │  SERVICES / PACKAGES     │
│             │                          │
│  • devdash  │  When services exist:    │
│  • app1     │    Show SERVICES pane    │
│  • app2     │                          │
│             │  When no services:       │
│             │    Show PACKAGES pane    │
└─────────────┴──────────────────────────┘
│           LOGS                         │
└────────────────────────────────────────┘
```

**OR** separate dedicated pane:

```
┌─────────────┬──────────────┬──────────┐
│  PROJECTS   │  SERVICES    │ PACKAGES │
│             │              │          │
│  • devdash  │  nginx ●     │ go 1.21  │
│  • app1     │  postgres ●  │ python   │
│  • app2     │              │ nodejs   │
└─────────────┴──────────────┴──────────┘
│           LOGS                         │
└────────────────────────────────────────┘
```

### Package Information Display

Display format options:

**Option 1: Simple Table**
```
PACKAGE          VERSION    SOURCE
go               1.21.5     nixpkgs
nodejs           20.10.0    nixpkgs
python           3.11.7     nixpkgs
postgresql       16.1       nixpkgs
```

**Option 2: Categorized**
```
LANGUAGES
  go              1.21.5
  python          3.11.7
  nodejs          20.10.0

SERVICES
  postgresql      16.1
  redis           7.2.3

TOOLS
  git             2.43.0
  gh              2.40.1
```

**Option 3: Compact List**
```
• go 1.21.5
• python 3.11.7
• nodejs 20.10.0
• postgresql 16.1
• redis 7.2.3
• git 2.43.0
```

### Data Source

Parse package information from `devenv.nix` or query the Nix store:

**Method 1: Parse devenv.nix**
```nix
{ pkgs, ... }: {
  packages = [
    pkgs.go
    pkgs.nodejs
    pkgs.python3
  ];

  languages.go.enable = true;
  services.postgres.enable = true;
}
```

**Method 2: Query Nix Store**
```bash
# Inside the devenv shell
nix-store --query --requisites ./.devenv/profile-* | grep -E '^/nix/store/.*-(.*)-(.*)$'
```

**Method 3: devenv info command**
```bash
cd project-path
devenv info packages  # If such command exists or could be added
```

### Enhanced Information

Beyond just package names and versions:

- **Derivation path**: `/nix/store/abc123...`
- **Source URL**: github:nixos/nixpkgs
- **Update status**: "✓ Latest" / "⚠ Update available"
- **Size**: On-disk size of the package
- **Dependencies**: Show package dependency tree

### Keyboard Shortcuts

- `Tab` - Cycle through PROJECTS → SERVICES → PACKAGES → LOGS
- `p` - Toggle packages pane visibility
- `↑/↓` - Navigate package list
- `Enter` - Show package details (version, derivation, dependencies)
- `u` - Check for package updates (if feasible)

### Implementation Considerations

#### Challenges

1. **Parsing devenv.nix**: Complex, needs Nix expression parser
2. **Nix store query**: May be slow, requires shell execution
3. **Version detection**: Nix package versions are in derivation names
4. **Update checking**: Would require comparing against nixpkgs latest
5. **Cross-platform**: Must work on Linux, macOS, NixOS

#### Simpler MVP Approach

Start with basic implementation:

1. Parse `devenv.nix` for simple `packages = [ ... ]` lists
2. Extract package names only (no versions initially)
3. Show static list (no version checking or updates)
4. Add version info later via Nix store queries

### Alternative: Integration with existing tools

Instead of building from scratch, integrate with:

- `nix-tree` - Interactive Nix store explorer
- `nix profile list` - Show installed packages
- `devenv info` - If devenv adds package introspection

### Example Implementation Phases

**Phase 1: Basic Package List**
- Parse `devenv.nix` for `packages = [ pkgs.* ]`
- Display package names only
- Static list, no updates

**Phase 2: Version Display**
- Query Nix store for version numbers
- Show "package-version" format

**Phase 3: Enhanced Metadata**
- Show derivation paths
- Display package sizes
- Add source information

**Phase 4: Interactive Features**
- Package detail view
- Dependency tree visualization
- Update checking

## Benefits

- Better visibility into devenv environment
- Useful for projects without services
- Helps debug dependency issues
- Educational for understanding Nix/devenv setups

## Alternatives Considered

1. **No packages pane**: Keep it simple, focus on services only
2. **External tool**: Use `nix-tree` or similar instead
3. **Project info overlay**: Show packages in a modal instead of dedicated pane
4. **Status bar info**: Show package count in status bar, details on demand

## Open Questions

1. Should packages pane replace or complement services pane?
2. How to handle projects with many packages (100+)?
3. Should we show all packages or only "primary" ones?
4. How to categorize packages (languages, services, tools)?
5. Performance impact of Nix store queries?

## Related Work

- `nix-tree` - Visual Nix store explorer
- `devenv info` - devenv environment information
- `nix profile list` - List installed packages
- Nix flake show - Show flake outputs

## Future Enhancements

- Package search/filter
- Show package documentation links
- Highlight packages with security updates
- Compare environments between projects
- Export package list to various formats
- Package dependency graph visualization
