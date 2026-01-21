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
	// Matrix - Classic green on black hacker theme
	"matrix": {
		Name:       "matrix",
		Primary:    lipgloss.Color("#39FF14"),
		Secondary:  lipgloss.Color("#00FF41"),
		Background: lipgloss.Color("#0D0D0D"),
		Muted:      lipgloss.Color("#4A4A4A"),
		Success:    lipgloss.Color("#00FF41"),
		Warning:    lipgloss.Color("#FFD700"),
		Error:      lipgloss.Color("#FF4136"),
	},

	// Gruvbox - Retro groove colors
	"gruvbox": {
		Name:       "gruvbox",
		Primary:    lipgloss.Color("#FE8019"), // Orange
		Secondary:  lipgloss.Color("#FABD2F"), // Yellow
		Background: lipgloss.Color("#282828"),
		Muted:      lipgloss.Color("#928374"),
		Success:    lipgloss.Color("#B8BB26"), // Green
		Warning:    lipgloss.Color("#FABD2F"), // Yellow
		Error:      lipgloss.Color("#FB4934"), // Red
	},

	// Dracula - Dark purple theme
	"dracula": {
		Name:       "dracula",
		Primary:    lipgloss.Color("#BD93F9"), // Purple
		Secondary:  lipgloss.Color("#FF79C6"), // Pink
		Background: lipgloss.Color("#282A36"),
		Muted:      lipgloss.Color("#6272A4"),
		Success:    lipgloss.Color("#50FA7B"), // Green
		Warning:    lipgloss.Color("#F1FA8C"), // Yellow
		Error:      lipgloss.Color("#FF5555"), // Red
	},

	// Nord - Cool arctic theme
	"nord": {
		Name:       "nord",
		Primary:    lipgloss.Color("#88C0D0"), // Frost blue
		Secondary:  lipgloss.Color("#81A1C1"), // Light blue
		Background: lipgloss.Color("#2E3440"),
		Muted:      lipgloss.Color("#4C566A"),
		Success:    lipgloss.Color("#A3BE8C"), // Green
		Warning:    lipgloss.Color("#EBCB8B"), // Yellow
		Error:      lipgloss.Color("#BF616A"), // Red
	},

	// Tokyo Night - Clean modern theme
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

	// Ayu Dark - Golden/orange accents
	"ayu-dark": {
		Name:       "ayu-dark",
		Primary:    lipgloss.Color("#FFB454"), // Orange/golden
		Secondary:  lipgloss.Color("#F07178"), // Coral
		Background: lipgloss.Color("#0A0E14"),
		Muted:      lipgloss.Color("#4D5566"),
		Success:    lipgloss.Color("#C2D94C"), // Lime green
		Warning:    lipgloss.Color("#FFB454"), // Orange
		Error:      lipgloss.Color("#F07178"), // Red/coral
	},

	// Solarized Dark - Scientific muted palette
	"solarized-dark": {
		Name:       "solarized-dark",
		Primary:    lipgloss.Color("#268BD2"), // Blue
		Secondary:  lipgloss.Color("#2AA198"), // Cyan
		Background: lipgloss.Color("#002B36"),
		Muted:      lipgloss.Color("#586E75"),
		Success:    lipgloss.Color("#859900"), // Green
		Warning:    lipgloss.Color("#B58900"), // Yellow
		Error:      lipgloss.Color("#DC322F"), // Red
	},

	// Monokai - Classic editor theme
	"monokai": {
		Name:       "monokai",
		Primary:    lipgloss.Color("#F92672"), // Pink/magenta
		Secondary:  lipgloss.Color("#FD971F"), // Orange
		Background: lipgloss.Color("#272822"),
		Muted:      lipgloss.Color("#75715E"),
		Success:    lipgloss.Color("#A6E22E"), // Green
		Warning:    lipgloss.Color("#E6DB74"), // Yellow
		Error:      lipgloss.Color("#F92672"), // Pink/red
	},
}

// GetTheme returns a theme by name, defaulting to matrix.
func GetTheme(name string) Theme {
	if theme, ok := Themes[name]; ok {
		return theme
	}
	return Themes["matrix"]
}
