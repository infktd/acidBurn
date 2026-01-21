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
	// Signature acidBurn theme
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

	// Catppuccin - Soothing pastel themes
	"catppuccin-mocha": {
		Name:       "catppuccin-mocha",
		Primary:    lipgloss.Color("#89B4FA"), // Blue
		Secondary:  lipgloss.Color("#F5C2E7"), // Pink
		Background: lipgloss.Color("#1E1E2E"),
		Muted:      lipgloss.Color("#6C7086"),
		Success:    lipgloss.Color("#A6E3A1"), // Green
		Warning:    lipgloss.Color("#F9E2AF"), // Yellow
		Error:      lipgloss.Color("#F38BA8"), // Red
	},
	"catppuccin-macchiato": {
		Name:       "catppuccin-macchiato",
		Primary:    lipgloss.Color("#8AADF4"), // Blue
		Secondary:  lipgloss.Color("#F5BDE6"), // Pink
		Background: lipgloss.Color("#24273A"),
		Muted:      lipgloss.Color("#6E738D"),
		Success:    lipgloss.Color("#A6DA95"), // Green
		Warning:    lipgloss.Color("#EED49F"), // Yellow
		Error:      lipgloss.Color("#ED8796"), // Red
	},
	"catppuccin-frappe": {
		Name:       "catppuccin-frappe",
		Primary:    lipgloss.Color("#8CAAEE"), // Blue
		Secondary:  lipgloss.Color("#F4B8E4"), // Pink
		Background: lipgloss.Color("#303446"),
		Muted:      lipgloss.Color("#737994"),
		Success:    lipgloss.Color("#A6D189"), // Green
		Warning:    lipgloss.Color("#E5C890"), // Yellow
		Error:      lipgloss.Color("#E78284"), // Red
	},
	"catppuccin-latte": {
		Name:       "catppuccin-latte",
		Primary:    lipgloss.Color("#1E66F5"), // Blue
		Secondary:  lipgloss.Color("#EA76CB"), // Pink
		Background: lipgloss.Color("#EFF1F5"),
		Muted:      lipgloss.Color("#9CA0B0"),
		Success:    lipgloss.Color("#40A02B"), // Green
		Warning:    lipgloss.Color("#DF8E1D"), // Yellow
		Error:      lipgloss.Color("#D20F39"), // Red
	},

	// Tokyo Night - Clean modern themes
	"tokyo-night": {
		Name:       "tokyo-night",
		Primary:    lipgloss.Color("#7AA2F7"), // Blue
		Secondary:  lipgloss.Color("#BB9AF7"), // Purple
		Background: lipgloss.Color("#1A1B26"),
		Muted:      lipgloss.Color("#565F89"),
		Success:    lipgloss.Color("#9ECE6A"), // Green
		Warning:    lipgloss.Color("#E0AF68"), // Yellow
		Error:      lipgloss.Color("#F7768E"), // Red
	},
	"tokyo-storm": {
		Name:       "tokyo-storm",
		Primary:    lipgloss.Color("#7AA2F7"), // Blue
		Secondary:  lipgloss.Color("#BB9AF7"), // Purple
		Background: lipgloss.Color("#24283B"),
		Muted:      lipgloss.Color("#565F89"),
		Success:    lipgloss.Color("#9ECE6A"), // Green
		Warning:    lipgloss.Color("#E0AF68"), // Yellow
		Error:      lipgloss.Color("#F7768E"), // Red
	},
	"tokyo-day": {
		Name:       "tokyo-day",
		Primary:    lipgloss.Color("#2E7DE9"), // Blue
		Secondary:  lipgloss.Color("#9854F1"), // Purple
		Background: lipgloss.Color("#E1E2E7"),
		Muted:      lipgloss.Color("#A8ADB7"),
		Success:    lipgloss.Color("#587539"), // Green
		Warning:    lipgloss.Color("#8C6C3E"), // Yellow
		Error:      lipgloss.Color("#F52A65"), // Red
	},

	// Gruvbox - Retro groove colors
	"gruvbox-dark": {
		Name:       "gruvbox-dark",
		Primary:    lipgloss.Color("#FE8019"), // Orange
		Secondary:  lipgloss.Color("#FABD2F"), // Yellow
		Background: lipgloss.Color("#282828"),
		Muted:      lipgloss.Color("#928374"),
		Success:    lipgloss.Color("#B8BB26"), // Green
		Warning:    lipgloss.Color("#FABD2F"), // Yellow
		Error:      lipgloss.Color("#FB4934"), // Red
	},
	"gruvbox-light": {
		Name:       "gruvbox-light",
		Primary:    lipgloss.Color("#AF3A03"), // Orange
		Secondary:  lipgloss.Color("#B57614"), // Yellow
		Background: lipgloss.Color("#FBF1C7"),
		Muted:      lipgloss.Color("#928374"),
		Success:    lipgloss.Color("#79740E"), // Green
		Warning:    lipgloss.Color("#B57614"), // Yellow
		Error:      lipgloss.Color("#CC241D"), // Red
	},
}

// GetTheme returns a theme by name, defaulting to acid-green.
func GetTheme(name string) Theme {
	if theme, ok := Themes[name]; ok {
		return theme
	}
	return Themes["acid-green"]
}
