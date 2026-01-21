package ui

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines all keybindings.
type KeyMap struct {
	// Global
	Quit       key.Binding
	Shutdown   key.Binding
	Settings   key.Binding
	EditConfig key.Binding
	Help       key.Binding
	Refresh    key.Binding
	History    key.Binding

	// Navigation
	Up     key.Binding
	Down   key.Binding
	Left     key.Binding
	Right    key.Binding
	Tab      key.Binding
	ShiftTab key.Binding
	Select   key.Binding
	Back     key.Binding

	// Actions
	Start   key.Binding
	Stop    key.Binding
	Restart key.Binding
	Search  key.Binding

	// Project Management
	Hide   key.Binding
	Delete key.Binding
	Edit   key.Binding
	Move   key.Binding
	Repair key.Binding

	// Logs
	Follow    key.Binding
	Filter    key.Binding
	Wrap      key.Binding
	Yank      key.Binding
	Top       key.Binding
	Bottom    key.Binding
	NextMatch key.Binding
	PrevMatch key.Binding
}

// DefaultKeyMap returns the default keybindings.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		// Global
		Quit: key.NewBinding(
			key.WithKeys("q"),
			key.WithHelp("q", "quit"),
		),
		Shutdown: key.NewBinding(
			key.WithKeys("ctrl+x"),
			key.WithHelp("ctrl+x", "shutdown all"),
		),
		Settings: key.NewBinding(
			key.WithKeys("S"),
			key.WithHelp("S", "settings"),
		),
		EditConfig: key.NewBinding(
			key.WithKeys("E"),
			key.WithHelp("E", "edit config"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("R"),
			key.WithHelp("R", "refresh"),
		),
		History: key.NewBinding(
			key.WithKeys("H"),
			key.WithHelp("H", "alerts"),
		),

		// Navigation
		Up: key.NewBinding(
			key.WithKeys("k", "up"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("j", "down"),
			key.WithHelp("↓/j", "down"),
		),
		Left: key.NewBinding(
			key.WithKeys("h", "left"),
			key.WithHelp("←/h", "left"),
		),
		Right: key.NewBinding(
			key.WithKeys("l", "right"),
			key.WithHelp("→/l", "right"),
		),
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next pane"),
		),
		ShiftTab: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "prev pane"),
		),
		Select: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
		),

		// Actions
		Start: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "start"),
		),
		Stop: key.NewBinding(
			key.WithKeys("x"),
			key.WithHelp("x", "stop"),
		),
		Restart: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "restart"),
		),
		Search: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "search"),
		),

		// Project Management
		Hide: key.NewBinding(
			key.WithKeys("ctrl+h"),
			key.WithHelp("ctrl+h", "hide/show"),
		),
		Delete: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "delete"),
		),
		Edit: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "rename"),
		),
		Move: key.NewBinding(
			key.WithKeys("m"),
			key.WithHelp("m", "relocate"),
		),
		Repair: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "repair stale"),
		),

		// Logs
		Follow: key.NewBinding(
			key.WithKeys("f"),
			key.WithHelp("f", "follow"),
		),
		Filter: key.NewBinding(
			key.WithKeys("ctrl+f"),
			key.WithHelp("ctrl+f", "filter"),
		),
		Wrap: key.NewBinding(
			key.WithKeys("w"),
			key.WithHelp("w", "wrap"),
		),
		Yank: key.NewBinding(
			key.WithKeys("y"),
			key.WithHelp("y", "yank"),
		),
		Top: key.NewBinding(
			key.WithKeys("g"),
			key.WithHelp("g", "top"),
		),
		Bottom: key.NewBinding(
			key.WithKeys("G"),
			key.WithHelp("G", "bottom"),
		),
		NextMatch: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "next match"),
		),
		PrevMatch: key.NewBinding(
			key.WithKeys("N"),
			key.WithHelp("N", "prev match"),
		),
	}
}

// ShortHelp returns the short help bindings.
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Select, k.Start, k.Stop, k.Help, k.Quit}
}

// FullHelp returns the full help bindings grouped by category.
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Tab, k.Select, k.Back},
		{k.Start, k.Stop, k.Restart, k.Search},
		{k.Follow, k.Top, k.Bottom, k.Wrap, k.NextMatch, k.PrevMatch},
		{k.Settings, k.History, k.Help, k.Quit},
	}
}
