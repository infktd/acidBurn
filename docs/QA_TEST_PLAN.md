# devdash QA Test Plan - v0.1.3 Reactive Features

## Overview
This test plan covers all reactive UI features added in the unreleased v0.1.3 version, plus regression testing for previous bug fixes.

**Testing Environment:**
- OS: macOS/Linux
- Terminal: Any modern terminal with Unicode support
- Test Projects: Multiple devenv.nix projects (ideally 3-5)
- Test Services: Services with varying CPU/memory usage

## Pre-Test Setup

### 1. Ensure Test Environment
```bash
# Build latest version
cd /Users/jayne/Desktop/codingProjects/devdash
go build -o devdash .

# Verify multiple test projects exist
./devdash
# You should see multiple projects in sidebar
```

### 2. Prepare Test Services
- At least one project with multiple services
- At least one service that generates frequent logs
- At least one service with moderate CPU/memory usage

## Feature Tests

### Test Suite 1: Real-time Log Flow Indicators

**Feature:** Animated braille spinner shows in ACTIVITY column when service logs within last 2 seconds.

#### Test 1.1: Activity Indicator Appears
**Steps:**
1. Start devdash: `./devdash`
2. Start a project with active services (press `s` on a project)
3. Wait for services to start logging
4. Focus on Services pane (press Tab if needed)
5. Observe ACTIVITY column (between STATUS and SERVICE)

**Expected:**
- [ ] ACTIVITY column is visible
- [ ] Animated spinner appears next to services that are actively logging
- [ ] Spinner uses braille characters (⣾⣽⣻⢿⡿⣟⣯⣷)
- [ ] Animation is smooth (100ms cycle)

#### Test 1.2: Activity Indicator Disappears
**Steps:**
1. Continue from Test 1.1
2. Wait for a service to stop logging for >2 seconds

**Expected:**
- [ ] Spinner disappears when no logs received for 2+ seconds
- [ ] Column shows empty space (not broken characters)

#### Test 1.3: Multiple Services
**Steps:**
1. Start project with 3+ services
2. Observe ACTIVITY column for all services

**Expected:**
- [ ] Each service shows independent activity state
- [ ] Spinners animate independently (not synchronized)
- [ ] Active services show spinner, idle services show blank

**Status:** ⬜ Pass / ⬜ Fail
**Notes:**

---

### Test Suite 2: Animated State Transitions

**Feature:** Status text flashes bold when service state changes (Running ↔ Stopped).

#### Test 2.1: Start Transition Flash
**Steps:**
1. Start devdash with idle project
2. Focus Services pane
3. Press `s` to start a service (or start entire project)
4. Watch STATUS column closely

**Expected:**
- [ ] When service transitions to "Running", text flashes bright green bold
- [ ] Flash intensity fades smoothly over ~1.5 seconds
- [ ] Final state returns to normal green "Running" text

#### Test 2.2: Stop Transition Flash
**Steps:**
1. Start devdash with running services
2. Focus Services pane
3. Press `x` to stop a service
4. Watch STATUS column closely

**Expected:**
- [ ] When service transitions to "Stopped", text flashes bright yellow/orange bold
- [ ] Flash intensity fades smoothly over ~1.5 seconds
- [ ] Final state returns to normal "Stopped" text

#### Test 2.3: Multiple State Changes
**Steps:**
1. Restart a service (press `r`)
2. Observe STATUS column

**Expected:**
- [ ] Flash occurs on Stopped transition
- [ ] Flash occurs again on Running transition
- [ ] Each transition independent (not overlapping)

**Status:** ⬜ Pass / ⬜ Fail
**Notes:**

---

### Test Suite 3: CPU/Memory Sparklines

**Feature:** Inline mini-graphs show last 10 resource readings using Unicode blocks.

#### Test 3.1: Sparklines Appear
**Steps:**
1. Start project with services using moderate CPU/memory
2. Wait for 3+ readings to be collected (~10-15 seconds)
3. Focus Services pane
4. Observe CPU and MEM columns

**Expected:**
- [ ] Sparklines appear after 3+ readings collected
- [ ] Sparklines use Unicode blocks (▁▂▃▄▅▆▇█)
- [ ] Sparklines display inline with values: "45.2% ▃▄▅▇▆▅▄▃"
- [ ] CPU and MEM columns are wide enough (18 and 20 chars)

#### Test 3.2: Sparkline Updates
**Steps:**
1. Continue from Test 3.1
2. Watch sparklines for 30 seconds

**Expected:**
- [ ] Sparklines update as new readings arrive
- [ ] Graph shows trends (increasing, decreasing, stable)
- [ ] Old values scroll left, new values appear on right

#### Test 3.3: Auto-normalization
**Steps:**
1. Observe sparkline for service with varying CPU usage
2. Note if CPU spikes from 10% to 50%

**Expected:**
- [ ] Sparkline scales to show relative changes clearly
- [ ] Not all blocks at same height (showing variation)
- [ ] Min and max values auto-adjust

#### Test 3.4: Edge Cases
**Steps:**
1. Stop a service (press `x`)
2. Restart service (press `r`)

**Expected:**
- [ ] Sparkline clears/resets when service stops
- [ ] Sparkline rebuilds when service restarts
- [ ] No broken characters or crashes

**Status:** ⬜ Pass / ⬜ Fail
**Notes:**

---

### Test Suite 4: Interactive Progress Bars

**Feature:** Detailed progress bars replace spinner during project start/stop with stages and percentage.

#### Test 4.1: Start Operation Progress
**Steps:**
1. Start devdash with idle project
2. Focus Projects pane
3. Press `s` to start project
4. Watch sidebar where project name appears

**Expected:**
- [ ] Progress bar appears immediately: `[██████░░░░] 60% Starting services...`
- [ ] Progress bar uses Unicode blocks (█ and ░)
- [ ] Percentage increases from 0% to ~95%
- [ ] Stage text changes:
  - Stage 1 (~0-30%): "Initializing environment..."
  - Stage 2 (~30-70%): "Starting services..."
  - Stage 3 (~70-95%): "Services online"
- [ ] Progress updates every ~200ms
- [ ] Bar fills left to right smoothly

#### Test 4.2: Stop Operation Progress
**Steps:**
1. Start devdash with running project
2. Focus Projects pane
3. Press `x` to stop project
4. Watch sidebar

**Expected:**
- [ ] Progress bar appears immediately
- [ ] Percentage increases from 0% to ~95%
- [ ] Stage text changes:
  - Stage 1 (~0-60%): "Stopping services..."
  - Stage 2 (~60-95%): "Cleaning up..."
- [ ] Progress updates every ~200ms

#### Test 4.3: Progress Completion
**Steps:**
1. Continue from Test 4.1 or 4.2
2. Wait for operation to complete

**Expected:**
- [ ] Progress reaches 95% and holds
- [ ] When operation completes, progress bar disappears
- [ ] Project name returns to normal display
- [ ] No visual artifacts left behind

#### Test 4.4: Multiple Projects
**Steps:**
1. Start two different projects in quick succession
2. Observe sidebar

**Expected:**
- [ ] Only one progress bar shows at a time
- [ ] Progress bar shows for the active operation
- [ ] No progress bar conflicts or overlaps

**Status:** ⬜ Pass / ⬜ Fail
**Notes:**

---

### Test Suite 5: Bidirectional Panel Navigation

**Feature:** Tab cycles forward, Shift+Tab cycles backward through panels.

#### Test 5.1: Tab Navigation
**Steps:**
1. Start devdash
2. Note which pane is focused (has colored title)
3. Press Tab
4. Press Tab again
5. Press Tab third time

**Expected:**
- [ ] First Tab: Projects → Services
- [ ] Second Tab: Services → Logs
- [ ] Third Tab: Logs → Projects (wraps around)
- [ ] Focused pane title changes color (primary color)
- [ ] Footer shows "[Tab] Switch Pane" on all panes

#### Test 5.2: Shift+Tab Navigation
**Steps:**
1. Start devdash (starts on Projects pane)
2. Press Shift+Tab
3. Press Shift+Tab again
4. Press Shift+Tab third time

**Expected:**
- [ ] First Shift+Tab: Projects → Logs (backward)
- [ ] Second Shift+Tab: Logs → Services
- [ ] Third Shift+Tab: Services → Projects (wraps around)
- [ ] Focused pane title changes color correctly

#### Test 5.3: Mixed Navigation
**Steps:**
1. Press Tab twice (Projects → Services → Logs)
2. Press Shift+Tab once (Logs → Services)
3. Press Tab once (Services → Logs)

**Expected:**
- [ ] Navigation works correctly in both directions
- [ ] Focus state always accurate
- [ ] No visual glitches during navigation

**Status:** ⬜ Pass / ⬜ Fail
**Notes:**

---

## Regression Tests

### Test Suite 6: Bug Fixes

#### Test 6.1: Projects List Navigation Glitch
**Bug:** Pressing up at top of projects list caused rendering glitch and color change.

**Steps:**
1. Start devdash
2. Ensure Projects pane is focused
3. Press up arrow repeatedly (ensure cursor at top)
4. Observe text rendering and colors

**Expected:**
- [ ] No rendering glitches occur
- [ ] Text colors remain consistent
- [ ] Cursor stays at top item (no wrapping)
- [ ] No flickering or artifacts

**Status:** ⬜ Pass / ⬜ Fail
**Notes:**

#### Test 6.2: Footer Keybindings Alignment
**Steps:**
1. Start devdash
2. Observe footer at bottom of screen
3. Focus each pane (Projects, Services, Logs)

**Expected:**
- [ ] Keybindings are centered in footer
- [ ] Footer text doesn't overflow or wrap
- [ ] All keybindings visible on all panes

**Status:** ⬜ Pass / ⬜ Fail
**Notes:**

---

## Performance Tests

### Test Suite 7: Performance and Stability

#### Test 7.1: Long Running Session
**Steps:**
1. Start devdash with multiple running projects
2. Leave running for 15+ minutes
3. Observe UI responsiveness

**Expected:**
- [ ] No memory leaks (use Activity Monitor to check)
- [ ] Animations remain smooth
- [ ] No slowdown over time
- [ ] No crashes or freezes

#### Test 7.2: Heavy Logging
**Steps:**
1. Start project with service generating very frequent logs (>100 lines/sec)
2. Observe UI responsiveness

**Expected:**
- [ ] Activity indicator animates smoothly
- [ ] No UI lag or freezing
- [ ] Logs scroll correctly
- [ ] No dropped frames in animations

#### Test 7.3: Many Services
**Steps:**
1. Start project with 10+ services
2. Observe Services table

**Expected:**
- [ ] All sparklines render correctly
- [ ] All activity indicators work
- [ ] Table scrolls smoothly
- [ ] No visual artifacts with many rows

**Status:** ⬜ Pass / ⬜ Fail
**Notes:**

---

## Theme Compatibility Tests

### Test Suite 8: Theme Testing

#### Test 8.1: Test All Themes
**Steps:**
1. Start devdash
2. Press `S` for settings
3. Navigate to Theme, press Enter
4. Test each theme:
   - Acid Green
   - Gruvbox
   - Dracula
   - Nord
   - Tokyo Night
   - Ayu Dark
   - Solarized Dark
   - Monokai

**For each theme, verify:**
- [ ] Activity indicators visible (not invisible against background)
- [ ] State flash colors appropriate for theme
- [ ] Sparklines visible and legible
- [ ] Progress bars styled with theme primary color
- [ ] All text readable (sufficient contrast)

**Status:** ⬜ Pass / ⬜ Fail
**Notes:**

---

## Edge Cases

### Test Suite 9: Edge Case Testing

#### Test 9.1: No Active Services
**Steps:**
1. Start devdash with all projects stopped
2. Observe Services pane

**Expected:**
- [ ] No sparklines shown (or shown as empty)
- [ ] No activity indicators
- [ ] No crashes or errors

#### Test 9.2: Service Crashes
**Steps:**
1. Start project with service that exits immediately
2. Observe Services table

**Expected:**
- [ ] EXIT code shown correctly
- [ ] Status shows "Stopped" with flash animation
- [ ] No lingering activity indicators
- [ ] Sparklines clear correctly

#### Test 9.3: Quick Restarts
**Steps:**
1. Press `r` to restart a service
2. Immediately press `r` again before first restart completes

**Expected:**
- [ ] No visual glitches
- [ ] Progress bars don't overlap
- [ ] UI remains responsive
- [ ] No crashes

**Status:** ⬜ Pass / ⬜ Fail
**Notes:**

---

## Test Summary

### Results Overview
- **Test Suites Passed:** __ / 9
- **Total Tests Passed:** __ / __
- **Critical Issues Found:** __
- **Minor Issues Found:** __

### Critical Issues
_(List any blocking issues that prevent release)_

### Minor Issues
_(List any non-blocking issues for future improvement)_

### Recommendation
⬜ **APPROVED FOR RELEASE** - All critical tests passed
⬜ **NEEDS WORK** - Critical issues must be resolved before release

### Tester Information
- **Tester Name:** ________________
- **Date:** ________________
- **Environment:** ________________
- **devdash Version:** ________________

---

## Notes for QA Agent

### Tips for Effective Testing
1. Test in a real terminal with Unicode support (not IDE integrated terminal)
2. Use actual devenv.nix projects, not mocks
3. Take screenshots of any visual glitches
4. Note exact reproduction steps for any bugs found
5. Test with different terminal sizes (resize window during tests)

### Known Limitations
- Progress bars cap at 95% until operation completes (expected behavior)
- Sparklines require 3+ readings before displaying (expected behavior)
- Activity indicators have 2-second timeout (expected behavior)

### Reporting Issues
For each issue found:
1. Issue title (brief description)
2. Severity (Critical, High, Medium, Low)
3. Reproduction steps (detailed)
4. Expected behavior
5. Actual behavior
6. Screenshots/terminal recordings if applicable
7. Environment details (OS, terminal, devdash version)
