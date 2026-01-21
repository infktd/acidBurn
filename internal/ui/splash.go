package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ASCII art options for acidBurn
var asciiArtOptions = map[string]string{
	"default": `
    ___    __________  ____  __  ______  _   __
   /   |  / ____/  _/ / __ \/ / / / __ \/ | / /
  / /| | / /    / /  / / / / /_/ / /_/ /  |/ /
 / ___ |/ /____/ /  / /_/ / __  / _, _/ /|  /
/_/  |_|\____/___/ /_____/_/ /_/_/ |_/_/ |_/
`,
	"block": `
 █████╗  ██████╗██╗██████╗ ██████╗ ██╗   ██╗██████╗ ███╗   ██╗
██╔══██╗██╔════╝██║██╔══██╗██╔══██╗██║   ██║██╔══██╗████╗  ██║
███████║██║     ██║██║  ██║██████╔╝██║   ██║██████╔╝██╔██╗ ██║
██╔══██║██║     ██║██║  ██║██╔══██╗██║   ██║██╔══██╗██║╚██╗██║
██║  ██║╚██████╗██║██████╔╝██████╔╝╚██████╔╝██║  ██║██║ ╚████║
╚═╝  ╚═╝ ╚═════╝╚═╝╚═════╝ ╚═════╝  ╚═════╝ ╚═╝  ╚═╝╚═╝  ╚═══╝
`,
	"small": `
          _     _ ____
  __ _ __(_) __| | __ ) _   _ _ __ _ __
 / _' / __| |/ _' |  _ \| | | | '__| '_ \
| (_| \__ \ | (_| | |_) | |_| | |  | | | |
 \__,_|___/_|\__,_|____/ \__,_|_|  |_| |_|
`,
	"minimal": `
┌─────────────────────────────┐
│     a c i d B U R N         │
│     devenv control plane    │
└─────────────────────────────┘
`,
	"hacker": `
    █████  ██████ ██ ██████  ██████  ██    ██ ██████  ███    ██
   ██   ██ ██     ██ ██   ██ ██   ██ ██    ██ ██   ██ ████   ██
   ███████ ██     ██ ██   ██ ██████  ██    ██ ██████  ██ ██  ██
   ██   ██ ██     ██ ██   ██ ██   ██ ██    ██ ██   ██ ██  ██ ██
   ██   ██  █████ ██ ██████  ██████   ██████  ██   ██ ██   ████
`,
}

var defaultAsciiArt = asciiArtOptions["default"]

// SplashScreen displays the startup splash with progress.
type SplashScreen struct {
	styles   *Styles
	width    int
	height   int
	visible  bool
	progress float64 // 0.0 to 1.0
	message  string
	asciiArt string
}

// NewSplashScreen creates a new splash screen.
func NewSplashScreen(styles *Styles, width, height int) *SplashScreen {
	return &SplashScreen{
		styles:   styles,
		width:    width,
		height:   height,
		visible:  true,
		progress: 0,
		message:  "Initializing...",
		asciiArt: defaultAsciiArt,
	}
}

// SetProgress updates the progress (0.0 to 1.0).
func (s *SplashScreen) SetProgress(p float64) {
	if p < 0 {
		p = 0
	}
	if p > 1 {
		p = 1
	}
	s.progress = p
}

// SetMessage updates the status message.
func (s *SplashScreen) SetMessage(msg string) {
	s.message = msg
}

// SetAsciiArt allows customizing the ASCII art.
func (s *SplashScreen) SetAsciiArt(art string) {
	s.asciiArt = art
}

// SetAsciiArtByName sets the ASCII art by preset name.
// Available options: "default", "block", "small", "minimal", "hacker"
func (s *SplashScreen) SetAsciiArtByName(name string) {
	if art, ok := asciiArtOptions[name]; ok {
		s.asciiArt = art
	}
}

// GetAsciiArtNames returns available ASCII art preset names.
func GetAsciiArtNames() []string {
	return []string{"default", "block", "small", "minimal", "hacker"}
}

// Show makes the splash screen visible.
func (s *SplashScreen) Show() {
	s.visible = true
}

// Hide dismisses the splash screen.
func (s *SplashScreen) Hide() {
	s.visible = false
}

// IsVisible returns whether the splash is shown.
func (s *SplashScreen) IsVisible() bool {
	return s.visible
}

// SetSize updates dimensions.
func (s *SplashScreen) SetSize(width, height int) {
	s.width = width
	s.height = height
}

// View renders the splash screen.
func (s *SplashScreen) View() string {
	if !s.visible {
		return ""
	}

	// Get theme color for ASCII art
	artColor := lipgloss.Color("#00FF00") // Default acid green
	if s.styles != nil {
		artColor = s.styles.theme.Primary
	}

	// Style the ASCII art
	artStyle := lipgloss.NewStyle().
		Foreground(artColor).
		Bold(true)

	// Center the ASCII art
	art := artStyle.Render(s.asciiArt)

	// Title
	titleStyle := lipgloss.NewStyle().
		Foreground(artColor).
		Bold(true)
	title := titleStyle.Render("acidBurn")

	// Message
	msgStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888"))
	message := msgStyle.Render(s.message)

	// Progress bar
	progressBar := s.renderProgressBar()

	// Combine all elements
	content := lipgloss.JoinVertical(
		lipgloss.Center,
		art,
		"",
		title,
		"",
		message,
		progressBar,
	)

	// Center in the screen
	containerStyle := lipgloss.NewStyle().
		Width(s.width).
		Height(s.height).
		Align(lipgloss.Center, lipgloss.Center)

	return containerStyle.Render(content)
}

func (s *SplashScreen) renderProgressBar() string {
	barWidth := 30
	filled := int(s.progress * float64(barWidth))
	empty := barWidth - filled

	// Progress bar characters
	filledChar := "\u2588" // Full block
	emptyChar := "\u2591"  // Light shade

	bar := strings.Repeat(filledChar, filled) + strings.Repeat(emptyChar, empty)
	percent := fmt.Sprintf("%3.0f%%", s.progress*100)

	barStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00"))
	percentStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))

	return barStyle.Render(bar) + " " + percentStyle.Render(percent)
}

// Progress returns the current progress value.
func (s *SplashScreen) Progress() float64 {
	return s.progress
}

// Message returns the current message.
func (s *SplashScreen) Message() string {
	return s.message
}
