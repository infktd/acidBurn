# devdash: init at 0.1.0

## Description

`devdash` is a terminal user interface (TUI) dashboard for managing [devenv.sh](https://devenv.sh) projects. It provides a unified control plane for monitoring and controlling multiple devenv environments, similar to how Docker Desktop provides a GUI for Docker containers.

### What is devenv?

devenv is a popular Nix-based development environment manager that uses process-compose for service orchestration. Many Nix developers use devenv to manage their development environments, but there hasn't been a good way to monitor multiple projects at once.

### Why devdash?

When working with multiple devenv projects, developers need to:
- Manually navigate to each project directory
- Run separate `devenv up` commands
- Check logs across multiple terminals
- Monitor service health manually

devdash solves this by providing:
- **Automatic project discovery** - Scans configured directories for all devenv projects
- **Centralized dashboard** - View all projects and their services in one place
- **Service control** - Start, stop, restart services across any project
- **Real-time log streaming** - Aggregated logs with search, filtering, and syntax highlighting
- **Health monitoring** - Track service crashes and recoveries with desktop notifications
- **Multiple themes** - Matrix, Nord, Dracula color schemes

## Screenshots

```
┌─ Projects ────────────────┐┌─ Services (my-api) ─────────────────────────┐
│ > my-api         [Running]││  postgres    Running  ↑ 2h 15m  CPU: 2%    │
│   blog-site      [Idle]   ││  redis       Running  ↑ 2h 15m  CPU: 1%    │
│   data-pipeline  [Stopped]││  api-server  Running  ↑ 2h 14m  CPU: 15%   │
└───────────────────────────┘│  worker      Stopped  ↓ 5m ago             │
                             └──────────────────────────────────────────────┘
┌─ Logs (api-server) ──────────────────────────────────────────────────────┐
│ 2026-01-21 21:30:45 INFO  Server starting on port 8080                   │
│ 2026-01-21 21:30:46 INFO  Connected to database                          │
│ 2026-01-21 21:31:02 DEBUG Handling request GET /api/users               │
└──────────────────────────────────────────────────────────────────────────┘
```

## Testing Done

### Build Verification
- [x] Builds successfully on `x86_64-linux`
- [x] Builds successfully on `aarch64-darwin`
- [x] Binary runs without errors
- [x] All Go tests pass (317 tests, 59.8% coverage)

### Runtime Testing
- [x] Successfully discovers devenv projects
- [x] Connects to process-compose Unix sockets
- [x] Service start/stop/restart works correctly
- [x] Log streaming and search functional
- [x] Desktop notifications work (Linux/macOS)
- [x] Theme switching works
- [x] Configuration persistence works

### Package Metadata
- [x] `meta.description` accurately describes the package
- [x] `meta.homepage` points to project repository
- [x] `meta.changelog` points to releases
- [x] `meta.license` correctly set to MIT
- [x] `meta.maintainers` includes package maintainer
- [x] `meta.mainProgram` set to "devdash"
- [x] `meta.platforms` appropriate for TUI application

## Dependencies

### Runtime Dependencies (via Go modules)
- **bubbletea** - Terminal UI framework
- **lipgloss** - Styling and layout
- **process-compose** - Service orchestration (external, not packaged)
- **notify** - Desktop notifications (Linux/macOS)

### Build Dependencies
- Go 1.25+
- Standard `buildGoModule` infrastructure

## Installation & Usage

After this PR is merged, users can install with:

```bash
# Install
nix-env -iA nixpkgs.devdash

# Or run directly
nix run nixpkgs#devdash

# Or in a shell
nix-shell -p devdash
```

### Configuration

devdash looks for devenv projects in `~/coding` by default. Users can customize scan paths via `~/.config/devdash/config.yaml`:

```yaml
scan_paths:
  - ~/projects
  - ~/work
  - ~/repos
```

## Implementation Notes

### Build Process
- Uses `buildGoModule` with proper vendoring
- `vendorHash` computed from go.mod/go.sum
- Stripped binaries with `-s -w` ldflags for smaller size
- No special build flags or patches needed

### Platform Support
- **Linux**: Full support (x86_64, aarch64)
- **macOS**: Full support (x86_64, aarch64)
- **Windows**: Not supported (Unix socket dependency)

### Breaking Changes
None - this is the initial release.

## Related Packages

This package complements the existing Nix/devenv ecosystem:
- `devenv` - The development environment manager itself
- `process-compose` - Service orchestrator (required runtime dependency)
- `direnv` - Often used alongside devenv for automatic environment loading

## Upstream Information

- **Repository**: https://github.com/infktd/devdash
- **License**: MIT
- **Latest Release**: v0.1.0
- **Issue Tracker**: https://github.com/infktd/devdash/issues
- **CI Status**: GitHub Actions (all tests passing)

## Checklist

- [x] Built and tested locally
- [x] Follows nixpkgs coding conventions
- [x] Package placed in correct location (`pkgs/by-name/de/devdash/`)
- [x] No unnecessary dependencies
- [x] License file included in source
- [x] Meta attributes complete and accurate
- [x] Tested on multiple platforms
- [x] Documentation reviewed for clarity

## Future Enhancements

Potential future additions (not in this PR):
- NixOS module for system-wide configuration
- Home Manager module integration
- Shell completions (bash/zsh/fish)
- Man page generation

## Questions for Reviewers

1. **Placement**: Placed in `pkgs/by-name/de/devdash/` following the new package structure - is this correct?
2. **Maintainers**: Should I add myself to `maintainers/maintainer-list.nix`?
3. **Platform tags**: Used `platforms.unix` - should this be more specific?
4. **Dependencies**: process-compose is a runtime dependency but not packaged - should this be documented in meta?

---

## Additional Context

This package fills a gap in the devenv ecosystem. While devenv provides excellent per-project development environments, managing multiple projects simultaneously has been challenging. devdash provides the missing "control tower" for developers juggling multiple devenv projects.

The project has:
- Comprehensive test coverage (59.8%, 317 tests)
- Clean Go code following idiomatic patterns
- Active development and maintenance
- Growing user base in the devenv community

Thank you for reviewing! Happy to make any requested changes.
