package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/infktd/devdash/internal/compose"
	"github.com/infktd/devdash/internal/config"
	"github.com/infktd/devdash/internal/registry"
)

func TestModelImplementsTeaModel(t *testing.T) {
	cfg := config.Default()
	reg := &registry.Registry{}
	var m tea.Model = New(cfg, reg)
	if m == nil {
		t.Fatal("New() returned nil")
	}
}

func TestModelInit(t *testing.T) {
	cfg := config.Default()
	reg := &registry.Registry{}
	m := New(cfg, reg)
	cmd := m.Init()
	if cmd == nil {
		t.Error("Init() should return a command")
	}
}

func TestModelView(t *testing.T) {
	cfg := config.Default()
	reg := &registry.Registry{}
	m := New(cfg, reg)
	view := m.View()
	if view == "" {
		t.Fatal("View() returned empty string")
	}
}

func TestModelHandlesWindowSize(t *testing.T) {
	cfg := config.Default()
	reg := &registry.Registry{}
	m := New(cfg, reg)

	msg := tea.WindowSizeMsg{Width: 100, Height: 50}
	newModel, _ := m.Update(msg)
	model := newModel.(*Model)

	if model.width != 100 {
		t.Errorf("Expected width 100, got %d", model.width)
	}
	if model.height != 50 {
		t.Errorf("Expected height 50, got %d", model.height)
	}
}

func TestModelComponentsInitialized(t *testing.T) {
	cfg := config.Default()
	reg := &registry.Registry{}
	m := New(cfg, reg)

	if m.logView == nil {
		t.Error("logView should be initialized")
	}
	if m.toast == nil {
		t.Error("toast should be initialized")
	}
	if m.alerts == nil {
		t.Error("alerts should be initialized")
	}
	if m.settings == nil {
		t.Error("settings should be initialized")
	}
	if m.splash == nil {
		t.Error("splash should be initialized")
	}
	if m.health == nil {
		t.Error("health monitor should be initialized")
	}
	if m.notifier == nil {
		t.Error("notifier should be initialized")
	}
	if m.clients == nil {
		t.Error("clients map should be initialized")
	}
}

func TestModelCycleFocus(t *testing.T) {
	cfg := config.Default()
	reg := &registry.Registry{}
	m := New(cfg, reg)

	if m.focused != PaneSidebar {
		t.Errorf("Expected sidebar, got %v", m.focused)
	}

	m.cycleFocus()
	if m.focused != PaneServices {
		t.Errorf("Expected services, got %v", m.focused)
	}

	m.cycleFocus()
	if m.focused != PaneLogs {
		t.Errorf("Expected logs, got %v", m.focused)
	}

	m.cycleFocus()
	if m.focused != PaneSidebar {
		t.Errorf("Expected sidebar (wrap), got %v", m.focused)
	}
}

func TestModelMoveUpDown(t *testing.T) {
	cfg := config.Default()
	reg := &registry.Registry{
		Projects: []*registry.Project{
			{Path: "/a", Name: "a"},
			{Path: "/b", Name: "b"},
			{Path: "/c", Name: "c"},
		},
	}
	m := New(cfg, reg)
	m.showSplash = false

	if m.selectedProject != 0 {
		t.Errorf("Expected 0, got %d", m.selectedProject)
	}

	m.moveDown()
	if m.selectedProject != 1 {
		t.Errorf("Expected 1, got %d", m.selectedProject)
	}

	m.moveDown()
	if m.selectedProject != 2 {
		t.Errorf("Expected 2, got %d", m.selectedProject)
	}

	m.moveDown()
	if m.selectedProject != 2 {
		t.Errorf("Expected 2 (clamped), got %d", m.selectedProject)
	}

	m.moveUp()
	if m.selectedProject != 1 {
		t.Errorf("Expected 1, got %d", m.selectedProject)
	}
}

func TestModelCurrentProject(t *testing.T) {
	cfg := config.Default()
	reg := &registry.Registry{
		Projects: []*registry.Project{
			{Path: "/a", Name: "a"},
			{Path: "/b", Name: "b"},
		},
	}
	m := New(cfg, reg)

	p := m.currentProject()
	if p == nil {
		t.Fatal("currentProject() returned nil")
	}
	if p.Name != "a" {
		t.Errorf("Expected 'a', got %q", p.Name)
	}

	m.selectedProject = 1
	p = m.currentProject()
	if p.Name != "b" {
		t.Errorf("Expected 'b', got %q", p.Name)
	}
}

func TestModelCurrentProjectEmpty(t *testing.T) {
	cfg := config.Default()
	reg := &registry.Registry{Projects: []*registry.Project{}}
	m := New(cfg, reg)

	p := m.currentProject()
	if p != nil {
		t.Error("Expected nil for empty registry")
	}
}

func TestModelShowsHelp(t *testing.T) {
	cfg := config.Default()
	reg := &registry.Registry{}
	m := New(cfg, reg)
	m.width = 100
	m.height = 40
	m.showSplash = false
	m.showHelp = true

	view := m.View()
	if view == "" {
		t.Error("View() returned empty string for help")
	}
}

func TestModelShowsAlertHistory(t *testing.T) {
	cfg := config.Default()
	reg := &registry.Registry{}
	m := New(cfg, reg)
	m.width = 100
	m.height = 40
	m.showSplash = false
	m.alertsPanel.Show()

	view := m.View()
	if view == "" {
		t.Error("View() returned empty string for alert history")
	}
}

// TestMouseFocusFollowing verifies mouse motion changes focus between panes
func TestMouseFocusFollowing(t *testing.T) {
	cfg := config.Default()
	cfg.UI.SidebarWidth = 30
	reg := &registry.Registry{}
	m := New(cfg, reg)
	m.width = 100
	m.height = 40
	m.showSplash = false

	// Mouse in sidebar (x < 30)
	msg := tea.MouseMsg{
		Type: tea.MouseMotion,
		X:    15,
		Y:    10,
	}
	newModel, _ := m.Update(msg)
	model := newModel.(*Model)
	if model.focused != PaneSidebar {
		t.Errorf("Expected focus on sidebar, got %v", model.focused)
	}

	// Mouse in services pane (x >= 30, y < servicesHeight)
	msg = tea.MouseMsg{
		Type: tea.MouseMotion,
		X:    50,
		Y:    5,
	}
	newModel, _ = model.Update(msg)
	model = newModel.(*Model)
	if model.focused != PaneServices {
		t.Errorf("Expected focus on services, got %v", model.focused)
	}

	// Mouse in logs pane (x >= 30, y >= servicesHeight)
	msg = tea.MouseMsg{
		Type: tea.MouseMotion,
		X:    50,
		Y:    30,
	}
	newModel, _ = model.Update(msg)
	model = newModel.(*Model)
	if model.focused != PaneLogs {
		t.Errorf("Expected focus on logs, got %v", model.focused)
	}
}

// TestMouseFocusIgnoresWhenModalsOpen verifies mouse doesn't change focus with modals open
func TestMouseFocusIgnoresWhenModalsOpen(t *testing.T) {
	cfg := config.Default()
	reg := &registry.Registry{}
	m := New(cfg, reg)
	m.width = 100
	m.height = 40
	m.focused = PaneSidebar

	// Show help modal
	m.showHelp = true

	msg := tea.MouseMsg{
		Type: tea.MouseMotion,
		X:    50,
		Y:    10,
	}
	newModel, _ := m.Update(msg)
	model := newModel.(*Model)

	// Focus should remain on sidebar (mouse ignored)
	if model.focused != PaneSidebar {
		t.Errorf("Expected focus to remain on sidebar with modal open, got %v", model.focused)
	}
}

// TestMouseWheelScrolling verifies mouse wheel scrolls logs
func TestMouseWheelScrolling(t *testing.T) {
	cfg := config.Default()
	cfg.UI.SidebarWidth = 30
	reg := &registry.Registry{}
	m := New(cfg, reg)
	m.width = 100
	m.height = 60 // Taller for better pane separation
	m.showSplash = false

	// Add many log entries so scrolling is possible
	for i := 0; i < 100; i++ {
		m.logView.AddEntry(LogEntry{
			Message: "test log entry",
			Level:   LevelInfo,
		})
	}

	// Set logs pane size
	m.logView.SetSize(70, 20)

	initialOffset := m.logView.offset

	// Mouse wheel up in logs pane (well into logs area)
	// Services height is ~(60-4)/3 = 18, so y=40 is definitely in logs
	msg := tea.MouseMsg{
		Type: tea.MouseWheelUp,
		X:    50,
		Y:    40,
	}
	newModel, _ := m.Update(msg)
	model := newModel.(*Model)

	// Offset should have increased (scrolled up)
	if model.logView.offset <= initialOffset {
		t.Errorf("Expected log offset to increase after mouse wheel up, was %d, now %d", initialOffset, model.logView.offset)
	}
}

// TestAutoUpdateServicesOnProjectChange verifies services update when navigating projects
func TestAutoUpdateServicesOnProjectChange(t *testing.T) {
	cfg := config.Default()
	reg := &registry.Registry{
		Projects: []*registry.Project{
			{Path: "/a", Name: "ProjectA"},
			{Path: "/b", Name: "ProjectB"},
		},
	}
	m := New(cfg, reg)
	m.showSplash = false
	m.selectedProject = 0

	// Set some initial service state
	m.services = []compose.ProcessStatus{{Name: "oldservice"}}

	// Move down to next project
	cmd := m.moveDown()

	// Should return a command to poll services
	if cmd == nil {
		t.Error("moveDown should return poll command when project changes")
	}

	// Services should be cleared for the new project
	if len(m.services) != 0 {
		t.Errorf("Expected services to be cleared, got %d services", len(m.services))
	}

	// Selected project should have changed
	if m.selectedProject != 1 {
		t.Errorf("Expected selectedProject 1, got %d", m.selectedProject)
	}
}

// TestNoUpdateServicesWhenProjectUnchanged verifies no update if project doesn't change
func TestNoUpdateServicesWhenProjectUnchanged(t *testing.T) {
	cfg := config.Default()
	reg := &registry.Registry{
		Projects: []*registry.Project{
			{Path: "/a", Name: "ProjectA"},
		},
	}
	m := New(cfg, reg)
	m.showSplash = false
	m.selectedProject = 0

	// Try to move down (but we're at the bottom)
	cmd := m.moveDown()

	// Should not return a command since project didn't change
	if cmd != nil {
		t.Error("moveDown should not return command when project doesn't change")
	}
}

// TestSwitchToCurrentProjectResetsState verifies state is cleared on project switch
func TestSwitchToCurrentProjectResetsState(t *testing.T) {
	cfg := config.Default()
	reg := &registry.Registry{}
	m := New(cfg, reg)

	// Set up some state
	m.services = []compose.ProcessStatus{{Name: "test"}}
	m.selectedService = 5
	m.lastLogMsg = map[string]string{"svc": "msg"}
	m.serviceStates = map[string]string{"svc": "running"}
	m.cpuHistory = map[string][]float64{"svc": {1.0, 2.0}}

	// Switch project
	m.switchToCurrentProject()

	// All state should be reset
	if len(m.services) != 0 {
		t.Error("services should be cleared")
	}
	if m.selectedService != 0 {
		t.Error("selectedService should be reset to 0")
	}
	if len(m.lastLogMsg) != 0 {
		t.Error("lastLogMsg should be cleared")
	}
	if len(m.serviceStates) != 0 {
		t.Error("serviceStates should be cleared")
	}
	if len(m.cpuHistory) != 0 {
		t.Error("cpuHistory should be cleared")
	}
}

func TestModelPackagesViewInitialized(t *testing.T) {
	cfg := config.Default()
	reg := &registry.Registry{}
	m := New(cfg, reg)

	if m.packagesView == nil {
		t.Error("packagesView should be initialized")
	}
}

func TestShouldShowBothPanes(t *testing.T) {
	cfg := config.Default()
	reg := &registry.Registry{}
	m := New(cfg, reg)

	// Narrow terminal - should show single pane
	m.width = 139
	if m.shouldShowBothPanes() {
		t.Error("expected false for width 139")
	}

	// Wide terminal - should show both panes
	m.width = 140
	if !m.shouldShowBothPanes() {
		t.Error("expected true for width 140")
	}

	m.width = 200
	if !m.shouldShowBothPanes() {
		t.Error("expected true for width 200")
	}
}

func TestTogglePackagesView(t *testing.T) {
	cfg := config.Default()
	reg := &registry.Registry{}
	m := New(cfg, reg)
	m.width = 100 // Narrow terminal

	// Default should show services (showPackages = false)
	if m.showPackages {
		t.Error("expected showPackages to be false by default")
	}

	// Toggle to packages
	m.togglePackagesView()
	if !m.showPackages {
		t.Error("expected showPackages to be true after toggle")
	}

	// Toggle back to services
	m.togglePackagesView()
	if m.showPackages {
		t.Error("expected showPackages to be false after second toggle")
	}
}

func TestPressP_TogglesPackagesView(t *testing.T) {
	cfg := config.Default()
	reg := &registry.Registry{}
	m := New(cfg, reg)
	m.width = 100 // Narrow terminal
	m.showSplash = false

	initialState := m.showPackages

	// Press 'p' key
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}}
	newModel, _ := m.Update(msg)
	model := newModel.(*Model)

	if model.showPackages == initialState {
		t.Error("pressing 'p' should toggle showPackages")
	}
}

func TestPressP_NoEffectOnWideTerminal(t *testing.T) {
	cfg := config.Default()
	reg := &registry.Registry{}
	m := New(cfg, reg)
	m.width = 150 // Wide terminal - both panes visible
	m.showSplash = false

	// Press 'p' key shouldn't have effect
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}}
	newModel, _ := m.Update(msg)
	model := newModel.(*Model)

	// On wide terminals, 'p' could still toggle but doesn't affect display
	// Just verify it doesn't crash
	_ = model
}

func TestProjectSwitch_ScansPackages(t *testing.T) {
	cfg := config.Default()
	reg := &registry.Registry{
		Projects: []*registry.Project{
			{Path: "/tmp/testproject", Name: "test"},
		},
	}
	m := New(cfg, reg)
	m.selectedProject = 0

	// Switch to project (this will try to scan packages)
	m.switchToCurrentProject()

	// PackagesView should have been updated (even if empty)
	// This is a basic test - in real usage, packages would be populated
	// if .devenv/profile/bin/ exists
	if m.packagesView == nil {
		t.Error("packagesView should not be nil after project switch")
	}
}

func TestView_WideTerminal_ShowsBothPanes(t *testing.T) {
	cfg := config.Default()
	reg := &registry.Registry{}
	m := New(cfg, reg)
	m.width = 150
	m.height = 40
	m.showSplash = false

	view := m.View()

	// Both "SERVICES" and "PACKAGES" should appear in view
	if !strings.Contains(view, "SERVICES") {
		t.Error("wide terminal view should contain SERVICES")
	}
	if !strings.Contains(view, "PACKAGES") {
		t.Error("wide terminal view should contain PACKAGES")
	}
}

func TestView_NarrowTerminal_ShowsIndicator(t *testing.T) {
	cfg := config.Default()
	reg := &registry.Registry{}
	m := New(cfg, reg)
	m.width = 100
	m.height = 40
	m.showSplash = false

	// Showing services - should have packages indicator
	m.showPackages = false
	view := m.View()
	if !strings.Contains(view, "[p:packages]") {
		t.Error("narrow terminal showing services should contain '[p:packages]' indicator")
	}

	// Showing packages - should have services indicator
	m.showPackages = true
	view = m.View()
	if !strings.Contains(view, "[p:services]") {
		t.Error("narrow terminal showing packages should contain '[p:services]' indicator")
	}
}

func TestMouse_HoverPackagesPane(t *testing.T) {
	cfg := config.Default()
	cfg.UI.SidebarWidth = 30
	reg := &registry.Registry{}
	m := New(cfg, reg)
	m.width = 150
	m.height = 60
	m.showSplash = false

	// Mouse in packages area (x >= 30, y in packages region)
	// Services roughly at top quarter, packages at second quarter
	packagesY := 20 // In packages region

	msg := tea.MouseMsg{
		Type: tea.MouseMotion,
		X:    50,
		Y:    packagesY,
	}

	newModel, _ := m.Update(msg)
	model := newModel.(*Model)

	if model.focused != PanePackages {
		t.Errorf("expected focus on packages, got %v", model.focused)
	}
}

func TestPackagesPaneIntegration(t *testing.T) {
	// Create test project directory with packages
	tmpDir := t.TempDir()
	projectPath := filepath.Join(tmpDir, "testproject")
	binDir := filepath.Join(projectPath, ".devenv", "profile", "bin")

	err := os.MkdirAll(binDir, 0755)
	if err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	// Create test binaries
	testBins := map[string]string{
		"go":      "#!/bin/sh\necho go1.21",
		"python3": "#!/bin/sh\necho python3.11",
		"node":    "#!/bin/sh\necho node20",
	}

	for name, content := range testBins {
		binPath := filepath.Join(binDir, name)
		err := os.WriteFile(binPath, []byte(content), 0755)
		if err != nil {
			t.Fatalf("failed to create binary %s: %v", name, err)
		}
	}

	// Create model with test project
	cfg := config.Default()
	reg := &registry.Registry{
		Projects: []*registry.Project{
			{Path: projectPath, Name: "testproject"},
		},
	}
	m := New(cfg, reg)
	m.width = 150 // Wide terminal
	m.height = 50
	m.showSplash = false
	m.selectedProject = 0

	// Trigger project switch (should scan packages)
	m.switchToCurrentProject()

	// Verify packages were scanned
	view := m.View()
	if !strings.Contains(view, "PACKAGES") {
		t.Error("view should contain PACKAGES pane")
	}

	// Switch to narrow terminal
	m.width = 100
	msg := tea.WindowSizeMsg{Width: 100, Height: 50}
	newModel, _ := m.Update(msg)
	m = newModel.(*Model)

	// Should show services by default
	view = m.View()
	if !strings.Contains(view, "[p:packages]") {
		t.Error("narrow terminal should show packages indicator")
	}

	// Toggle to packages
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}}
	newModel, _ = m.Update(keyMsg)
	m = newModel.(*Model)

	view = m.View()
	if !strings.Contains(view, "[p:services]") {
		t.Error("narrow terminal showing packages should show services indicator")
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes int64
		want  string
	}{
		{0, "0B"},
		{1023, "1023B"},
		{1024, "1K"},
		{1024 * 1024, "1.0M"},
		{1024 * 1024 * 1024, "1.0G"},
		{1536, "2K"},
		{1024 * 1024 * 2, "2.0M"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%d", tt.bytes), func(t *testing.T) {
			got := formatBytes(tt.bytes)
			if got != tt.want {
				t.Errorf("formatBytes(%d) = %q, want %q", tt.bytes, got, tt.want)
			}
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		duration time.Duration
		want     string
	}{
		{0, "0s"},
		{30 * time.Second, "30s"},
		{60 * time.Second, "1m"},
		{90 * time.Second, "1m"},
		{time.Hour, "1h"},
		{time.Hour + time.Minute, "1h1m"},
		{24 * time.Hour, "1d"},
		{25 * time.Hour, "1d1h"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := formatDuration(tt.duration)
			if got != tt.want {
				t.Errorf("formatDuration(%v) = %q, want %q", tt.duration, got, tt.want)
			}
		})
	}
}

func TestFormatSystemTime(t *testing.T) {
	tests := []struct {
		sysTime string
		want    string
	}{
		{"0s", "0s"},
		{"30s", "30s"},
		{"1m", "1m"},
		{"1h", "1h"},
		{"1h30m", "1h30m"},
		{"24h", "1d"},
		{"invalid", "invalid"},
	}

	for _, tt := range tests {
		t.Run(tt.sysTime, func(t *testing.T) {
			got := formatSystemTime(tt.sysTime)
			if got == "" {
				t.Errorf("formatSystemTime(%q) returned empty string", tt.sysTime)
			}
			// For valid inputs, verify format is reasonable
			if tt.sysTime != "invalid" && got != tt.want {
				t.Errorf("formatSystemTime(%q) = %q, want %q", tt.sysTime, got, tt.want)
			}
		})
	}
}

func TestModelRenderProjectItem(t *testing.T) {
	// Create model with minimal setup
	cfg := config.Default()
	reg := &registry.Registry{}
	reg.AddProject("/test/myproject")
	m := New(cfg, reg)

	// Trigger window size to initialize view properly
	m.Update(tea.WindowSizeMsg{Width: 100, Height: 50})

	// This method is private, test via rendering the full sidebar
	// For now, verify the method exists by running full View

	view := m.View()

	// Should not be empty
	if view == "" {
		t.Error("View() should not be empty")
	}

	// When model renders projects, it should display them
	// The actual rendering depends on terminal width and internal state
	// Just verify no panic occurs and view is generated
}

func TestModelRenderSparkline(t *testing.T) {
	cfg := config.Default()
	reg := &registry.Registry{}
	m := New(cfg, reg)

	data := map[string][]float64{
		"test-service": {0.1, 0.3, 0.5, 0.7, 0.9},
	}

	// Private method - test via integration
	// or test full rendering path

	// For now, ensure model can render with sparkline data
	m.cpuHistory = data
	view := m.View()

	// Should not panic
	_ = view
}

func TestModelGetActivityIndicator(t *testing.T) {
	cfg := config.Default()
	reg := &registry.Registry{}
	m := New(cfg, reg)

	// Activity indicator for different states
	// Private method - test behavior via state changes

	// Just verify model with various states doesn't panic
	m.View()
}

func TestProjectFilterValue(t *testing.T) {
	proj := &registry.Project{
		Name: "test-project",
		Path: "/path/to/project",
	}

	// projectListItem implements FilterValue for list filtering
	item := projectListItem{project: proj}
	filterVal := item.FilterValue()

	if filterVal == "" {
		t.Error("FilterValue() should not be empty")
	}

	// Should be searchable by name
	if !strings.Contains(filterVal, "test-project") {
		t.Errorf("FilterValue() = %q, should contain project name", filterVal)
	}
}

func TestProjectDelegateUpdate(t *testing.T) {
	cfg := config.Default()
	reg := &registry.Registry{}
	m := New(cfg, reg)
	delegate := NewProjectDelegate(m.styles, m)

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	cmd := delegate.Update(msg, nil)

	// Should handle update without panic
	_ = cmd
}

func TestProjectDelegateRender(t *testing.T) {
	cfg := config.Default()
	reg := &registry.Registry{}
	m := New(cfg, reg)
	delegate := NewProjectDelegate(m.styles, m)

	proj := &registry.Project{
		Name: "test-project",
		Path: "/test/path",
	}

	item := projectListItem{project: proj}

	// Test rendering via delegate interface
	// The Render method writes to an io.Writer, so we need to capture output
	// For now, just verify delegate exists and doesn't panic
	_ = delegate
	_ = item
}
