# Packages Pane

devdash displays installed development packages (Go, Python, Node.js, etc.) alongside services.

## Features

- **Automatic Discovery**: Scans `.devenv/profile/bin/` to find all installed packages
- **Version Information**: Shows package versions from Nix store paths
- **Adaptive Layout**: Automatically adjusts based on terminal size
- **Toggle View**: Press `p` to switch between services and packages on narrow terminals

## Layout Modes

### Wide Terminals (≥140 columns)

Both Services and Packages panes are visible simultaneously:
- Services pane in top quarter
- Packages pane in second quarter
- Logs pane in bottom half
- Full information always visible

### Narrow Terminals (<140 columns)

One pane at a time with toggle:
- Default: Services pane with `[p:packages]` indicator
- Press `p`: Packages pane with `[p:services]` indicator
- Press `p` again: Back to Services pane

## Package Information

Each package shows:
- **NAME**: Package binary name (e.g., "go", "python3")
- **VERSION**: Version extracted from Nix store path
- **TYPE**: Categorized type (Go, Python, Node.js, Rust, etc.)

## Keyboard Shortcuts

- `p`: Toggle between Services and Packages view (narrow terminals)
- `Tab`: Cycle focus through panes (includes Packages when visible)
- Mouse hover: Move focus to pane under cursor

## Project Types

- **Service-based projects**: Have both services and packages
- **Package-only projects**: Development environments without long-running services
  - Example: CLI tools, libraries, static sites
  - Only use `devenv shell`, never run `devenv up`
  - Still show useful package information

## Technical Details

Package discovery:
1. Scans `.devenv/profile/bin/` directory
2. Resolves symlinks to Nix store paths
3. Parses package name and version from path format: `/nix/store/<hash>-<name>-<version>/...`
4. Groups multiple binaries from same package (e.g., go, gofmt, godoc)
5. Categorizes by detected language/tool type

## Supported Package Types

The packages pane automatically categorizes packages into the following types:

- **Go**: go, gopls, gofmt, godoc, goimports
- **Python**: python, python3, pip, pytest, poetry
- **Node.js**: node, nodejs, npm, npx, yarn
- **Rust**: cargo, rustc, rustup
- **C/C++**: gcc, g++, clang, clang++
- **Ruby**: ruby, gem, bundle
- **Java**: java, javac, maven, gradle
- **Other**: Any unrecognized packages

## Examples

### Wide Terminal (150 columns)

```
┌─ SERVICES [myproject] ─────────────────┐
│ NAME     STATUS  CPU  MEM              │
│ postgres running 2%   45MB             │
│ redis    running 1%   12MB             │
└────────────────────────────────────────┘

┌─ PACKAGES (5) ─────────────────────────┐
│ NAME      VERSION     TYPE             │
│ go        1.21.5      Go               │
│ python3   3.11.7      Python           │
│ node      20.10.0     Node.js          │
│ cargo     1.75.0      Rust             │
│ gcc       13.2.0      C/C++            │
└────────────────────────────────────────┘

┌─ LOGS ─────────────────────────────────┐
│ [postgres] Server started              │
│ [redis] Ready to accept connections    │
└────────────────────────────────────────┘
```

### Narrow Terminal (100 columns)

Showing services:
```
┌─ SERVICES [myproject] [p:packages] ───┐
│ NAME     STATUS  CPU  MEM             │
│ postgres running 2%   45MB            │
└───────────────────────────────────────┘

┌─ LOGS ────────────────────────────────┐
│ [postgres] Server started             │
└───────────────────────────────────────┘
```

After pressing `p`:
```
┌─ PACKAGES (5) [p:services] ───────────┐
│ NAME      VERSION     TYPE            │
│ go        1.21.5      Go              │
│ python3   3.11.7      Python          │
└───────────────────────────────────────┘

┌─ LOGS ────────────────────────────────┐
│ [postgres] Server started             │
└───────────────────────────────────────┘
```

## Why Packages Pane?

The packages pane serves several purposes:

1. **Environment Verification**: Quickly verify which tools are available in your devenv
2. **Version Tracking**: See exact versions without running multiple `--version` commands
3. **Debugging**: Identify version conflicts or missing tools
4. **Documentation**: Know what's installed without checking configuration files
5. **Package-Only Projects**: Useful for projects that don't run services but still use devenv for environment management
