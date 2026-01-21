package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/lucasb-eyer/go-colorful"
)

// ASCII art options for devdash
var asciiArtOptions = map[string]string{
	"default": `
 ██████╗ ███████╗██╗   ██╗██████╗  █████╗ ███████╗██╗  ██╗
 ██╔══██╗██╔════╝██║   ██║██╔══██╗██╔══██╗██╔════╝██║  ██║
 ██║  ██║█████╗  ██║   ██║██║  ██║███████║███████╗███████║
 ██║  ██║██╔══╝  ╚██╗ ██╔╝██║  ██║██╔══██║╚════██║██╔══██║
 ██████╔╝███████╗ ╚████╔╝ ██████╔╝██║  ██║███████║██║  ██║
 ╚═════╝ ╚══════╝  ╚═══╝  ╚═════╝ ╚═╝  ╚═╝╚══════╝╚═╝  ╚═╝`,
	"slant": `
    ____  _______    ____  ___   _____ __  __
   / __ \/ ____/ |  / / / / /   / ___// / / /
  / / / / __/  | | / / / / /    \__ \/ /_/ /
 / /_/ / /___  | |/ / /_/ /    ___/ / __  /
/_____/_____/  |___/\____/    /____/_/ /_/
`,
	"small": `
     _                _           _
  __| | _____   ____| | __ _ ___| |__
 / _' |/ _ \ \ / / _' |/ _' / __| '_ \
| (_| |  __/\ V / (_| | (_| \__ \ | | |
 \__,_|\___| \_/ \__,_|\__,_|___/_| |_|
`,
	"minimal": `
┌─────────────────────────────┐
│       d e v D A S H         │
│     devenv control plane    │
└─────────────────────────────┘
`,
	"cyber": `
   ██████  ███████ ██    ██ ██████   █████  ███████ ██   ██
   ██   ██ ██      ██    ██ ██   ██ ██   ██ ██      ██   ██
   ██   ██ █████   ██    ██ ██   ██ ███████ ███████ ███████
   ██   ██ ██       ██  ██  ██   ██ ██   ██      ██ ██   ██
   ██████  ███████   ████   ██████  ██   ██ ███████ ██   ██
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
	frame    int // Animation frame counter
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

	// Get theme colors
	primaryColor := lipgloss.Color("#00FF00") // Default
	mutedColor := lipgloss.Color("#555555")
	if s.styles != nil {
		primaryColor = s.styles.theme.Primary
		mutedColor = s.styles.theme.Muted
	}

	// Style the ASCII art with bold primary color
	artStyle := lipgloss.NewStyle().
		Foreground(primaryColor).
		Bold(true)
	art := artStyle.Render(s.asciiArt)

	// Tagline
	taglineStyle := lipgloss.NewStyle().
		Foreground(mutedColor).
		Italic(true)
	tagline := taglineStyle.Render("── devenv fleet control ──")

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
		tagline,
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
	barWidth := 40
	filled := int(s.progress * float64(barWidth))

	// Get theme colors for gradient
	primaryHex := "#00FF00"
	secondaryHex := "#00FFFF" // Cyan as secondary gradient color
	mutedHex := "#333333"
	if s.styles != nil {
		primaryHex = string(s.styles.theme.Primary)
		mutedHex = string(s.styles.theme.Muted)
		// Use secondary color for the gradient end
		secondaryHex = string(s.styles.theme.Secondary)
	}

	// Parse colors for gradient blending
	primaryCol, _ := colorful.Hex(primaryHex)
	secondaryCol, _ := colorful.Hex(secondaryHex)

	// Wave characters for leading edge animation
	waveChars := []rune{'▓', '▒', '░'}

	// Build the progress bar with per-character gradient coloring
	var result strings.Builder
	result.WriteString("[")

	for i := 0; i < barWidth; i++ {
		if i < filled {
			// Calculate gradient position (0.0 to 1.0 across filled portion)
			var gradientPos float64
			if filled > 1 {
				gradientPos = float64(i) / float64(filled-1)
			}

			// Blend colors for gradient effect
			blendedColor := primaryCol.BlendLuv(secondaryCol, gradientPos)

			// Determine character (wave animation at leading edge)
			char := '█'
			if i >= filled-3 && filled < barWidth {
				waveIdx := (filled - 1 - i + s.frame) % 3
				if waveIdx < 0 {
					waveIdx = 0
				}
				char = waveChars[waveIdx]
			}

			// Apply gradient color to this character
			charStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(blendedColor.Hex()))
			result.WriteString(charStyle.Render(string(char)))
		} else {
			// Empty portion
			emptyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(mutedHex))
			result.WriteString(emptyStyle.Render("░"))
		}
	}

	result.WriteString("] ")

	// Percentage
	percent := fmt.Sprintf("%3.0f%%", s.progress*100)
	percentStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Bold(true)
	result.WriteString(percentStyle.Render(percent))

	return result.String()
}

// Tick advances the animation frame.
func (s *SplashScreen) Tick() {
	s.frame++
}

// Progress returns the current progress value.
func (s *SplashScreen) Progress() float64 {
	return s.progress
}

// Message returns the current message.
func (s *SplashScreen) Message() string {
	return s.message
}
