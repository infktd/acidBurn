package ui

import "github.com/charmbracelet/lipgloss"

// Styles holds all the styled components.
type Styles struct {
	// Theme reference
	theme Theme

	// Layout
	Header  lipgloss.Style
	Footer  lipgloss.Style
	Sidebar lipgloss.Style
	Main    lipgloss.Style

	// Components
	Title         lipgloss.Style
	Breadcrumb    lipgloss.Style
	StatusBar     lipgloss.Style
	ProjectItem   lipgloss.Style
	SelectedItem  lipgloss.Style
	ServiceRow    lipgloss.Style
	LogLine       lipgloss.Style
	LogTimestamp  lipgloss.Style
	LogLevelInfo  lipgloss.Style
	LogLevelWarn  lipgloss.Style
	LogLevelError lipgloss.Style

	// Status indicators
	StatusRunning  lipgloss.Style
	StatusIdle     lipgloss.Style
	StatusDegraded lipgloss.Style
	StatusStale    lipgloss.Style
	StatusMissing  lipgloss.Style

	// Borders
	FocusedBorder lipgloss.Style
	BlurredBorder lipgloss.Style
}

// NewStyles creates styles from a theme.
func NewStyles(theme Theme) *Styles {
	return &Styles{
		// Theme reference
		theme: theme,

		// Layout
		Header: lipgloss.NewStyle().
			Foreground(theme.Primary).
			Bold(true).
			Padding(0, 1),

		Footer: lipgloss.NewStyle().
			Foreground(theme.Muted).
			Padding(0, 1),

		Sidebar: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(theme.Muted).
			Padding(0, 1),

		Main: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(theme.Muted).
			Padding(0, 1),

		// Components
		Title: lipgloss.NewStyle().
			Foreground(theme.Primary).
			Bold(true),

		Breadcrumb: lipgloss.NewStyle().
			Foreground(theme.Muted),

		StatusBar: lipgloss.NewStyle().
			Foreground(theme.Secondary),

		ProjectItem: lipgloss.NewStyle().
			Foreground(theme.Secondary),

		SelectedItem: lipgloss.NewStyle().
			Foreground(theme.Primary).
			Bold(true),

		ServiceRow: lipgloss.NewStyle().
			Foreground(theme.Secondary),

		LogLine: lipgloss.NewStyle().
			Foreground(theme.Secondary),

		LogTimestamp: lipgloss.NewStyle().
			Foreground(theme.Muted),

		LogLevelInfo: lipgloss.NewStyle().
			Foreground(theme.Muted),

		LogLevelWarn: lipgloss.NewStyle().
			Foreground(theme.Warning),

		LogLevelError: lipgloss.NewStyle().
			Foreground(theme.Error).
			Bold(true),

		// Status indicators
		StatusRunning: lipgloss.NewStyle().
			Foreground(theme.Success),

		StatusIdle: lipgloss.NewStyle().
			Foreground(theme.Muted),

		StatusDegraded: lipgloss.NewStyle().
			Foreground(theme.Warning),

		StatusStale: lipgloss.NewStyle().
			Foreground(theme.Error),

		StatusMissing: lipgloss.NewStyle().
			Foreground(theme.Error),

		// Borders
		FocusedBorder: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(theme.Primary),

		BlurredBorder: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(theme.Muted),
	}
}
