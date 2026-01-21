# devdash Pre-Release TODO - v0.1.3

**Target Release:** This Weekend (2026-01-25/26)
**Status:** In Progress

---

## Critical Tasks (Must Complete Before Release)

### Testing & Quality Assurance
- [ ] **Run Full QA Test Plan**
  - File: `docs/QA_TEST_PLAN.md`
  - All 9 test suites must pass
  - Document any issues found
  - Verify all reactive features work correctly
  - Test on both macOS and Linux if possible

- [ ] **Performance Testing**
  - [ ] Run with 10+ services for 30+ minutes (check for memory leaks)
  - [ ] Test with high-frequency logging services (>100 lines/sec)
  - [ ] Verify UI remains responsive under load
  - [ ] Check CPU usage is reasonable (should be low when idle)

- [ ] **Theme Testing**
  - [ ] Test all 8 themes for readability
  - [ ] Verify activity indicators visible in all themes
  - [ ] Check sparklines have sufficient contrast
  - [ ] Ensure progress bars styled correctly per theme

### Documentation
- [ ] **Update README.md**
  - [ ] Add screenshots/GIFs of new reactive features:
    - Activity indicators in action
    - State transition flashes
    - CPU/Memory sparklines
    - Progress bars during operations
  - [ ] Update feature list with v0.1.3 additions
  - [ ] Verify installation instructions still accurate
  - [ ] Update keybindings section (add Tab/Shift+Tab)

- [ ] **Finalize CHANGELOG.md**
  - [ ] Move [Unreleased] section to [0.1.3] with release date
  - [ ] Verify all technical details are accurate
  - [ ] Add any additional fixes/changes discovered during QA
  - [ ] Update version history table at bottom

- [ ] **Create Release Notes**
  - [ ] Write user-friendly release notes (less technical than CHANGELOG)
  - [ ] Highlight the 4 major reactive features
  - [ ] Include before/after comparisons if possible
  - [ ] Note any breaking changes (none expected)

### Code Quality
- [ ] **Run Tests**
  ```bash
  go test ./...
  go vet ./...
  go fmt ./...
  ```

- [ ] **Verify Build on Multiple Platforms**
  - [ ] macOS (Intel)
  - [ ] macOS (Apple Silicon)
  - [ ] Linux (x86_64)
  - [ ] Verify no platform-specific issues with Unicode rendering

- [ ] **Code Review**
  - [ ] Review all changes in v0.1.3
  - [ ] Check for any TODO comments left in code
  - [ ] Verify no debug logging left enabled
  - [ ] Ensure error handling is complete

### Git & Release Prep
- [ ] **Version Bump**
  - [ ] Update version in relevant files (if version constant exists)
  - [ ] Ensure CHANGELOG shows v0.1.3 with release date

- [ ] **Git State**
  - [ ] All changes committed
  - [ ] No uncommitted work in tree
  - [ ] Clean working directory

- [ ] **Create Git Tag**
  ```bash
  git tag -a v0.1.3 -m "Release v0.1.3 - Reactive UI Features"
  ```

---

## High Priority (Should Complete Before Release)

### User Experience
- [ ] **Test Edge Cases**
  - [ ] What happens with 0 projects?
  - [ ] What happens with 0 services?
  - [ ] What if project has no logs?
  - [ ] What if socket connection fails?
  - [ ] What if config file is corrupted?

- [ ] **Error Messages**
  - [ ] Review all error messages for clarity
  - [ ] Ensure errors are actionable (tell user what to do)
  - [ ] Check toast notifications aren't too verbose
  - [ ] Verify critical errors are visible

- [ ] **Keybindings**
  - [ ] Verify help modal (`?`) shows all new keybindings
  - [ ] Check for any keybinding conflicts
  - [ ] Ensure Tab/Shift+Tab documented in help
  - [ ] Test all keybindings work as expected

### Polish
- [ ] **Terminal Compatibility**
  - [ ] Test in iTerm2
  - [ ] Test in Terminal.app
  - [ ] Test in Alacritty
  - [ ] Test in kitty
  - [ ] Verify Unicode characters render correctly everywhere

- [ ] **Window Resizing**
  - [ ] Test resize during normal operation
  - [ ] Test resize during progress bar display
  - [ ] Test with very small terminal sizes
  - [ ] Test with very large terminal sizes

---

## Medium Priority (Nice to Have Before Release)

### Documentation
- [ ] **Architecture Documentation**
  - [ ] Document how reactive features work
  - [ ] Add developer guide for adding new features
  - [ ] Document the Model structure and state management
  - [ ] Add comments to complex functions

- [ ] **User Guide**
  - [ ] Create quickstart tutorial
  - [ ] Document common workflows
  - [ ] Add troubleshooting section
  - [ ] Include tips for optimal usage

### Features
- [ ] **Configuration Validation**
  - [ ] Add validation for config file values
  - [ ] Provide helpful error messages for invalid configs
  - [ ] Add config migration if format changed

- [ ] **Accessibility**
  - [ ] Test with screen readers if possible
  - [ ] Ensure color isn't the only indicator (use symbols too)
  - [ ] Add option to disable animations if needed
  - [ ] Test with different terminal color settings

---

## Low Priority (Post-Release)

### Future Enhancements
- [ ] **Additional Reactive Features**
  - [ ] Network I/O indicators
  - [ ] Disk usage sparklines
  - [ ] Service dependency graph visualization
  - [ ] Log search highlighting improvements

- [ ] **Configuration**
  - [ ] Allow customization of animation speeds
  - [ ] Allow customization of sparkline history depth
  - [ ] Allow disabling individual reactive features
  - [ ] Add more theme customization options

- [ ] **Performance**
  - [ ] Optimize sparkline rendering
  - [ ] Reduce memory usage for log buffers
  - [ ] Consider lazy loading for large service lists

### Community
- [ ] **GitHub**
  - [ ] Add issue templates
  - [ ] Add PR templates
  - [ ] Add contributing guidelines
  - [ ] Set up GitHub Actions for CI/CD

- [ ] **Promotion**
  - [ ] Post to relevant forums/communities
  - [ ] Create demo video/GIF
  - [ ] Write blog post about development
  - [ ] Share on social media

---

## Known Issues

### Non-Blocking Issues
_(List any known issues that don't block release but should be tracked)_

1. **Progress bars cap at 95%**
   - Expected behavior, not a bug
   - Ensures bar doesn't reach 100% before operation completes
   - Document in user guide

2. **Sparklines require 3+ readings**
   - Expected behavior, not a bug
   - Need minimum data points for meaningful visualization
   - Document in user guide

3. **Activity indicators have 2s timeout**
   - Expected behavior, not a bug
   - Prevents stale indicators from lingering
   - Document in user guide

### Potential Improvements
_(Ideas for future versions, not blocking)_

1. Configurable animation speeds
2. Ability to pin/favorite projects
3. Export logs to file
4. Custom color schemes beyond themes
5. Service grouping/filtering

---

## Pre-Release Checklist

Complete this checklist before creating the release:

### Testing
- [ ] All QA tests passed
- [ ] Performance tests passed
- [ ] No critical bugs found
- [ ] Edge cases handled gracefully

### Documentation
- [ ] README updated with new features
- [ ] CHANGELOG finalized
- [ ] Release notes written
- [ ] Screenshots/GIFs captured

### Code
- [ ] All code committed
- [ ] Tests passing
- [ ] Build successful on all platforms
- [ ] No debug code left in

### Release
- [ ] Version bumped in relevant files
- [ ] Git tag created
- [ ] Ready to push and publish

---

## Release Process

When ready to release:

1. **Final Verification**
   ```bash
   # Clean build
   go clean
   go build -o devdash .

   # Run tests
   go test ./...

   # Verify version
   ./devdash --version  # if version flag exists
   ```

2. **Update CHANGELOG**
   - Move [Unreleased] to [0.1.3] with date
   - Create new empty [Unreleased] section

3. **Commit and Tag**
   ```bash
   git add CHANGELOG.md
   git commit -m "Release v0.1.3"
   git tag -a v0.1.3 -m "Release v0.1.3 - Reactive UI Features"
   ```

4. **Push to Remote**
   ```bash
   git push origin main
   git push origin v0.1.3
   ```

5. **Create GitHub Release**
   - Go to GitHub repository
   - Create new release from tag v0.1.3
   - Copy release notes
   - Attach compiled binaries if applicable
   - Publish release

6. **Announce**
   - Update any project websites/docs
   - Post to relevant communities
   - Update social media

---

## Post-Release Tasks

After successful release:

- [ ] Monitor for bug reports
- [ ] Respond to user feedback
- [ ] Start planning v0.2.0 features
- [ ] Update project roadmap
- [ ] Thank contributors

---

## Notes

### Target Date
Weekend of January 25-26, 2026

### Version Scheme
Using [0ver](https://0ver.org/) versioning as noted in CHANGELOG.md

### Release Philosophy
- Ship working software
- Iterate based on feedback
- Don't let perfect be the enemy of good
- Better to release and improve than wait forever

### Success Criteria
Release is successful if:
1. All critical tests pass
2. No known critical bugs
3. Documentation is complete enough for users to get started
4. Reactive features work as designed on target platforms

---

**Last Updated:** 2026-01-21
**Status:** Ready for QA Testing
