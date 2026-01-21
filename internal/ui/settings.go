package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/infktd/acidburn/internal/config"
)

// SettingsPanel manages the settings form.
type SettingsPanel struct {
	form    *huh.Form
	config  *config.Config
	visible bool
	saved   bool

	// Form values (bound to form fields)
	autoDiscover   bool
	scanDepth      int
	theme          string
	defaultLogView string
	systemNotifs   bool
	tuiAlerts      bool
}

// NewSettingsPanel creates a settings panel from the current config.
func NewSettingsPanel(cfg *config.Config) *SettingsPanel {
	sp := &SettingsPanel{
		config: cfg,
	}
	sp.loadFromConfig()
	sp.buildForm()
	return sp
}

func (sp *SettingsPanel) loadFromConfig() {
	sp.autoDiscover = sp.config.Projects.AutoDiscover
	sp.scanDepth = sp.config.Projects.ScanDepth
	sp.theme = sp.config.UI.Theme
	sp.defaultLogView = sp.config.UI.DefaultLogView
	sp.systemNotifs = sp.config.Notifications.SystemEnabled
	sp.tuiAlerts = sp.config.Notifications.TUIAlerts
}

func (sp *SettingsPanel) buildForm() {
	sp.form = huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Auto-discover projects").
				Value(&sp.autoDiscover),
			huh.NewSelect[int]().
				Title("Scan depth").
				Options(
					huh.NewOption("1", 1),
					huh.NewOption("2", 2),
					huh.NewOption("3", 3),
					huh.NewOption("4", 4),
					huh.NewOption("5", 5),
				).
				Value(&sp.scanDepth),
			huh.NewSelect[string]().
				Title("Default log view").
				Options(
					huh.NewOption("Focused", "focused"),
					huh.NewOption("Unified", "unified"),
				).
				Value(&sp.defaultLogView),
			huh.NewSelect[string]().
				Title("Theme").
				Options(
					huh.NewOption("Acid Green", "acid-green"),
					huh.NewOption("Nord", "nord"),
					huh.NewOption("Dracula", "dracula"),
				).
				Value(&sp.theme),
			huh.NewConfirm().
				Title("System notifications").
				Value(&sp.systemNotifs),
			huh.NewConfirm().
				Title("TUI alerts").
				Value(&sp.tuiAlerts),
		).Title("Settings"),
	).WithShowHelp(true)
}

// Show makes the settings panel visible and returns the init command.
func (sp *SettingsPanel) Show() tea.Cmd {
	sp.visible = true
	sp.saved = false
	sp.loadFromConfig()
	sp.buildForm()
	return sp.form.Init()
}

// Hide closes the settings panel.
func (sp *SettingsPanel) Hide() {
	sp.visible = false
}

// IsVisible returns whether the panel is shown.
func (sp *SettingsPanel) IsVisible() bool {
	return sp.visible
}

// WasSaved returns true if settings were saved on last close.
func (sp *SettingsPanel) WasSaved() bool {
	return sp.saved
}

// Form returns the huh form for Update/View integration.
func (sp *SettingsPanel) Form() *huh.Form {
	return sp.form
}

// Save applies form values to the config.
func (sp *SettingsPanel) Save() {
	sp.config.Projects.AutoDiscover = sp.autoDiscover
	sp.config.Projects.ScanDepth = sp.scanDepth
	sp.config.UI.Theme = sp.theme
	sp.config.UI.DefaultLogView = sp.defaultLogView
	sp.config.Notifications.SystemEnabled = sp.systemNotifs
	sp.config.Notifications.TUIAlerts = sp.tuiAlerts
	sp.saved = true
}

// Cancel discards form changes.
func (sp *SettingsPanel) Cancel() {
	sp.loadFromConfig() // Reset to original values
	sp.saved = false
}

// Config returns the underlying config.
func (sp *SettingsPanel) Config() *config.Config {
	return sp.config
}

// Update handles messages for the settings form.
func (sp *SettingsPanel) Update(msg tea.Msg) (*SettingsPanel, tea.Cmd) {
	if !sp.visible {
		return sp, nil
	}

	// Pass all messages to the form
	model, cmd := sp.form.Update(msg)
	sp.form = model.(*huh.Form)

	// Check if form is completed
	if sp.form.State == huh.StateCompleted {
		sp.Save()
		sp.visible = false
	}

	return sp, cmd
}

// View renders the settings panel.
func (sp *SettingsPanel) View() string {
	if !sp.visible {
		return ""
	}
	return sp.form.View()
}
