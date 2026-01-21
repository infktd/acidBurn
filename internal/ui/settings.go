package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/infktd/acidburn/internal/config"
)

// FieldType represents the type of setting field.
type FieldType int

const (
	FieldToggle FieldType = iota // Boolean on/off
	FieldSelect                   // String/int with options
	FieldButton                   // Action buttons (Save/Cancel)
)

// SelectOption represents an option in a Select field.
type SelectOption struct {
	Label string      // Display name
	Value interface{} // Actual value (string or int)
}

// SettingField represents a single setting field.
type SettingField struct {
	Label    string
	Type     FieldType
	Options  []SelectOption // For Select fields
	GetValue func() interface{}
	SetValue func(interface{})
}

// SettingsPanel manages the settings modal.
type SettingsPanel struct {
	// Configuration
	config *config.Config
	styles *Styles

	// Display state
	visible bool
	width   int
	height  int

	// Navigation state
	selectedField int  // Current field index (0-7)
	editMode      bool // True when editing a Select field

	// Working copy of values (not committed until Save)
	workingCopy struct {
		autoDiscover   bool
		scanDepth      int
		defaultLogView string
		theme          string
		systemNotifs   bool
		tuiAlerts      bool
	}

	// Field definitions
	fields []SettingField
}

// settingsSavedMsg is sent when settings are successfully saved.
type settingsSavedMsg struct{}

// settingsSaveErrorMsg is sent when settings fail to save.
type settingsSaveErrorMsg struct {
	err error
}

// NewSettingsPanel creates a settings panel from the current config.
func NewSettingsPanel(cfg *config.Config, styles *Styles, width, height int) *SettingsPanel {
	sp := &SettingsPanel{
		config: cfg,
		styles: styles,
		width:  width,
		height: height,
	}
	sp.loadWorkingCopy()
	sp.buildFields()
	return sp
}

// loadWorkingCopy copies config values to working copy.
func (sp *SettingsPanel) loadWorkingCopy() {
	sp.workingCopy.autoDiscover = sp.config.Projects.AutoDiscover
	sp.workingCopy.scanDepth = sp.config.Projects.ScanDepth
	sp.workingCopy.theme = sp.config.UI.Theme
	sp.workingCopy.defaultLogView = sp.config.UI.DefaultLogView
	sp.workingCopy.systemNotifs = sp.config.Notifications.SystemEnabled
	sp.workingCopy.tuiAlerts = sp.config.Notifications.TUIAlerts
}

// applyWorkingCopy applies working copy to config.
func (sp *SettingsPanel) applyWorkingCopy() {
	sp.config.Projects.AutoDiscover = sp.workingCopy.autoDiscover
	sp.config.Projects.ScanDepth = sp.workingCopy.scanDepth
	sp.config.UI.Theme = sp.workingCopy.theme
	sp.config.UI.DefaultLogView = sp.workingCopy.defaultLogView
	sp.config.Notifications.SystemEnabled = sp.workingCopy.systemNotifs
	sp.config.Notifications.TUIAlerts = sp.workingCopy.tuiAlerts
}

// buildFields creates the field definitions.
func (sp *SettingsPanel) buildFields() {
	sp.fields = []SettingField{
		{
			Label: "Auto-discover projects",
			Type:  FieldToggle,
			GetValue: func() interface{} {
				return sp.workingCopy.autoDiscover
			},
			SetValue: func(v interface{}) {
				sp.workingCopy.autoDiscover = v.(bool)
			},
		},
		{
			Label: "Scan depth",
			Type:  FieldSelect,
			Options: []SelectOption{
				{Label: "1", Value: 1},
				{Label: "2", Value: 2},
				{Label: "3", Value: 3},
				{Label: "4", Value: 4},
				{Label: "5", Value: 5},
			},
			GetValue: func() interface{} {
				return sp.workingCopy.scanDepth
			},
			SetValue: func(v interface{}) {
				sp.workingCopy.scanDepth = v.(int)
			},
		},
		{
			Label: "Default log view",
			Type:  FieldSelect,
			Options: []SelectOption{
				{Label: "Focused", Value: "focused"},
				{Label: "Unified", Value: "unified"},
			},
			GetValue: func() interface{} {
				return sp.workingCopy.defaultLogView
			},
			SetValue: func(v interface{}) {
				sp.workingCopy.defaultLogView = v.(string)
			},
		},
		{
			Label: "Theme",
			Type:  FieldSelect,
			Options: []SelectOption{
				{Label: "Acid Green", Value: "acid-green"},
				{Label: "Gruvbox", Value: "gruvbox"},
				{Label: "Dracula", Value: "dracula"},
				{Label: "Nord", Value: "nord"},
				{Label: "Tokyo Night", Value: "tokyo-night"},
				{Label: "Ayu Dark", Value: "ayu-dark"},
				{Label: "Solarized Dark", Value: "solarized-dark"},
				{Label: "Monokai", Value: "monokai"},
			},
			GetValue: func() interface{} {
				return sp.workingCopy.theme
			},
			SetValue: func(v interface{}) {
				sp.workingCopy.theme = v.(string)
			},
		},
		{
			Label: "System notifications",
			Type:  FieldToggle,
			GetValue: func() interface{} {
				return sp.workingCopy.systemNotifs
			},
			SetValue: func(v interface{}) {
				sp.workingCopy.systemNotifs = v.(bool)
			},
		},
		{
			Label: "TUI alerts",
			Type:  FieldToggle,
			GetValue: func() interface{} {
				return sp.workingCopy.tuiAlerts
			},
			SetValue: func(v interface{}) {
				sp.workingCopy.tuiAlerts = v.(bool)
			},
		},
		{
			Label: "Save",
			Type:  FieldButton,
		},
		{
			Label: "Cancel",
			Type:  FieldButton,
		},
	}
}

// Show makes the settings panel visible.
func (sp *SettingsPanel) Show() tea.Cmd {
	sp.visible = true
	sp.selectedField = 0
	sp.editMode = false
	sp.loadWorkingCopy()
	return nil
}

// Hide closes the settings panel.
func (sp *SettingsPanel) Hide() {
	sp.visible = false
}

// IsVisible returns whether the panel is shown.
func (sp *SettingsPanel) IsVisible() bool {
	return sp.visible
}

// SetSize updates the panel dimensions.
func (sp *SettingsPanel) SetSize(width, height int) {
	sp.width = width
	sp.height = height
}

// Cancel discards changes and closes the panel.
func (sp *SettingsPanel) Cancel() {
	sp.loadWorkingCopy()
	sp.visible = false
}

// Update handles messages for the settings panel.
func (sp *SettingsPanel) Update(msg tea.Msg) (*SettingsPanel, tea.Cmd) {
	if !sp.visible {
		return sp, nil
	}

	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return sp, nil
	}

	// Handle edit mode first (for Select fields)
	if sp.editMode {
		return sp.handleEditMode(keyMsg)
	}

	// Normal navigation mode
	return sp.handleNavigationMode(keyMsg)
}

// handleNavigationMode handles keys in navigation mode.
func (sp *SettingsPanel) handleNavigationMode(msg tea.KeyMsg) (*SettingsPanel, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if sp.selectedField > 0 {
			sp.selectedField--
		}

	case "down", "j":
		if sp.selectedField < len(sp.fields)-1 {
			sp.selectedField++
		}

	case "enter":
		return sp.handleFieldActivation()

	case "esc":
		sp.Cancel()
		return sp, nil
	}

	return sp, nil
}

// handleEditMode handles keys in edit mode (for Select fields).
func (sp *SettingsPanel) handleEditMode(msg tea.KeyMsg) (*SettingsPanel, tea.Cmd) {
	field := sp.fields[sp.selectedField]

	switch msg.String() {
	case "left", "h":
		sp.cyclePrevOption(field)

	case "right", "l":
		sp.cycleNextOption(field)

	case "enter":
		// Confirm selection and exit edit mode
		sp.editMode = false

	case "esc":
		// Cancel edit mode (value remains unchanged)
		sp.editMode = false
	}

	return sp, nil
}

// handleFieldActivation activates the current field.
func (sp *SettingsPanel) handleFieldActivation() (*SettingsPanel, tea.Cmd) {
	field := sp.fields[sp.selectedField]

	switch field.Type {
	case FieldToggle:
		// Flip boolean immediately
		currentValue := field.GetValue().(bool)
		field.SetValue(!currentValue)

	case FieldSelect:
		// Enter edit mode to use Left/Right arrows
		sp.editMode = true

	case FieldButton:
		if field.Label == "Save" {
			return sp.handleSave()
		} else if field.Label == "Cancel" {
			sp.Cancel()
		}
	}

	return sp, nil
}

// handleSave saves settings to disk.
func (sp *SettingsPanel) handleSave() (*SettingsPanel, tea.Cmd) {
	// Apply working copy to config
	sp.applyWorkingCopy()

	// Save to disk
	path := config.Path()
	err := config.Save(path, sp.config)

	if err != nil {
		// Return error as a message for toast notification
		return sp, func() tea.Msg {
			return settingsSaveErrorMsg{err: err}
		}
	}

	// Success - close modal
	sp.visible = false
	return sp, func() tea.Msg {
		return settingsSavedMsg{}
	}
}

// cyclePrevOption cycles to previous option in a Select field.
func (sp *SettingsPanel) cyclePrevOption(field SettingField) {
	currentValue := field.GetValue()
	currentIdx := -1

	for i, opt := range field.Options {
		if opt.Value == currentValue {
			currentIdx = i
			break
		}
	}

	if currentIdx > 0 {
		field.SetValue(field.Options[currentIdx-1].Value)
	} else {
		// Wrap around to last option
		field.SetValue(field.Options[len(field.Options)-1].Value)
	}
}

// cycleNextOption cycles to next option in a Select field.
func (sp *SettingsPanel) cycleNextOption(field SettingField) {
	currentValue := field.GetValue()
	currentIdx := -1

	for i, opt := range field.Options {
		if opt.Value == currentValue {
			currentIdx = i
			break
		}
	}

	if currentIdx < len(field.Options)-1 {
		field.SetValue(field.Options[currentIdx+1].Value)
	} else {
		// Wrap around to first option
		field.SetValue(field.Options[0].Value)
	}
}

// View renders the settings panel.
func (sp *SettingsPanel) View() string {
	if !sp.visible {
		return ""
	}

	// Build modal content
	content := sp.renderContent()

	// Fixed size modal box (60 cols x 20 rows)
	modalStyle := sp.styles.FocusedBorder.
		Width(60).
		Height(20).
		Padding(1, 2)

	return modalStyle.Render(content)
}

// renderContent builds the modal content.
func (sp *SettingsPanel) renderContent() string {
	var lines []string

	// Title (centered)
	titleStyle := lipgloss.NewStyle().Width(56).Align(lipgloss.Center)
	lines = append(lines, titleStyle.Render(sp.styles.Title.Render("Settings")))
	lines = append(lines, "")

	// Field list (centered)
	fieldStyle := lipgloss.NewStyle().Width(56).Align(lipgloss.Center)
	for i, field := range sp.fields {
		line := sp.renderField(i, field)
		lines = append(lines, fieldStyle.Render(line))
	}

	// Help text at bottom (centered)
	lines = append(lines, "")
	helpStyle := lipgloss.NewStyle().Width(56).Align(lipgloss.Center)
	lines = append(lines, helpStyle.Render(sp.styles.Breadcrumb.Render(sp.getHelpText())))

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

// renderField renders a single field.
func (sp *SettingsPanel) renderField(idx int, field SettingField) string {
	cursor := "  "
	if idx == sp.selectedField {
		cursor = "> "
	}

	var valueDisplay string
	switch field.Type {
	case FieldToggle:
		value := field.GetValue().(bool)
		if value {
			valueDisplay = sp.styles.StatusRunning.Render("[ON]")
		} else {
			valueDisplay = sp.styles.StatusIdle.Render("[OFF]")
		}

	case FieldSelect:
		value := field.GetValue()
		// Find matching option label
		for _, opt := range field.Options {
			if opt.Value == value {
				if sp.editMode && idx == sp.selectedField {
					// Show all options with current highlighted
					valueDisplay = sp.renderSelectOptions(field, value)
				} else {
					valueDisplay = sp.styles.Breadcrumb.Render(fmt.Sprintf("<%s>", opt.Label))
				}
				break
			}
		}

	case FieldButton:
		// For buttons, center the label
		buttonLabel := fmt.Sprintf("[%s]", field.Label)
		if idx == sp.selectedField {
			buttonLabel = sp.styles.SelectedItem.Render(buttonLabel)
		}
		return buttonLabel
	}

	// Calculate spacing for alignment
	// Total width: 52 chars (56 - 4 for cursor/padding)
	// Label gets left side, value gets right side

	// Calculate padding based on plain text lengths
	valueWidth := lipgloss.Width(valueDisplay)
	padding := 52 - len(cursor) - len(field.Label) - valueWidth
	if padding < 1 {
		padding = 1
	}

	// Apply styling if selected
	label := field.Label
	if idx == sp.selectedField && field.Type != FieldButton {
		label = sp.styles.SelectedItem.Render(label)
	}

	// Build the line: cursor + label + padding + value
	line := fmt.Sprintf("%s%s%s%s", cursor, label, strings.Repeat(" ", padding), valueDisplay)

	return line
}

// renderSelectOptions renders inline options for Select field in edit mode.
func (sp *SettingsPanel) renderSelectOptions(field SettingField, currentValue interface{}) string {
	// Display inline options: "1 [2] 3 4 5" with current in brackets
	var parts []string
	for _, opt := range field.Options {
		if opt.Value == currentValue {
			parts = append(parts, sp.styles.SelectedItem.Render(fmt.Sprintf("[%s]", opt.Label)))
		} else {
			parts = append(parts, opt.Label)
		}
	}
	return strings.Join(parts, " ")
}

// getHelpText returns context-sensitive help text.
func (sp *SettingsPanel) getHelpText() string {
	if sp.editMode {
		return "←/→ Change  [Enter] Confirm  [Esc] Cancel"
	}
	return "↑/↓ Navigate  [Enter] Select  [Esc] Cancel"
}
