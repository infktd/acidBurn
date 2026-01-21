package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/infktd/acidburn/internal/config"
	"github.com/infktd/acidburn/internal/registry"
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
	m.showAlerts = true

	view := m.View()
	if view == "" {
		t.Error("View() returned empty string for alert history")
	}
}
