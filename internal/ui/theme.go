package ui

import "github.com/charmbracelet/lipgloss"

// Theme defines the color palette for the UI.
type Theme struct {
	Name       string
	Primary    lipgloss.Color
	Secondary  lipgloss.Color
	Background lipgloss.Color
	Muted      lipgloss.Color
	Success    lipgloss.Color
	Warning    lipgloss.Color
	Error      lipgloss.Color
}

// Themes contains all available themes.
var Themes = map[string]Theme{
	"acid-green": {
		Name:       "acid-green",
		Primary:    lipgloss.Color("#39FF14"),
		Secondary:  lipgloss.Color("#00FF41"),
		Background: lipgloss.Color("#0D0D0D"),
		Muted:      lipgloss.Color("#4A4A4A"),
		Success:    lipgloss.Color("#00FF41"),
		Warning:    lipgloss.Color("#FFD700"),
		Error:      lipgloss.Color("#FF4136"),
	},
	"nord": {
		Name:       "nord",
		Primary:    lipgloss.Color("#88C0D0"),
		Secondary:  lipgloss.Color("#81A1C1"),
		Background: lipgloss.Color("#2E3440"),
		Muted:      lipgloss.Color("#4C566A"),
		Success:    lipgloss.Color("#A3BE8C"),
		Warning:    lipgloss.Color("#EBCB8B"),
		Error:      lipgloss.Color("#BF616A"),
	},
	"dracula": {
		Name:       "dracula",
		Primary:    lipgloss.Color("#BD93F9"),
		Secondary:  lipgloss.Color("#FF79C6"),
		Background: lipgloss.Color("#282A36"),
		Muted:      lipgloss.Color("#6272A4"),
		Success:    lipgloss.Color("#50FA7B"),
		Warning:    lipgloss.Color("#F1FA8C"),
		Error:      lipgloss.Color("#FF5555"),
	},
}

// GetTheme returns a theme by name, defaulting to acid-green.
func GetTheme(name string) Theme {
	if theme, ok := Themes[name]; ok {
		return theme
	}
	return Themes["acid-green"]
}
