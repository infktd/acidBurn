package ui

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/infktd/acidburn/internal/compose"
	"github.com/infktd/acidburn/internal/config"
	"github.com/infktd/acidburn/internal/health"
	"github.com/infktd/acidburn/internal/notify"
	"github.com/infktd/acidburn/internal/registry"
)

// FocusedPane tracks which pane has focus.
type FocusedPane int

const (
	PaneSidebar FocusedPane = iota
	PaneServices
	PaneLogs
)

// Model is the main application model for acidBurn.
type Model struct {
	// Core data
	config   *config.Config
	registry *registry.Registry
	styles   *Styles
	keys     KeyMap

	// Dimensions
	width  int
	height int

	// Focus state
	focused         FocusedPane
	selectedProject int
	selectedService int

	// Overlay state
	showHelp     bool
	showSettings bool
	showAlerts   bool
	showSplash   bool

	// Search state (for logs)
	searchMode  bool
	searchInput string

	// Project search state (for sidebar)
	projectSearchMode  bool
	projectSearchInput string

	// Components
	logView   *LogView
	toast     *ToastManager
	alerts    *AlertHistory
	settings  *SettingsPanel
	splash    *SplashScreen
	health    *health.Monitor
	notifier  *notify.Notifier

	// Compose clients per project (keyed by project path)
	clients map[string]*compose.Client

	// Current project services
	services []compose.ProcessStatus

	// Displayed projects in sidebar order (active first, then idle)
	displayedProjects []*registry.Project

	// Cached project states (to avoid inconsistent state during rendering)
	projectStates map[string]registry.ProjectState

	// Track last seen log message per service to fetch only new logs
	lastLogMsg map[string]string
}

// Messages for async operations
type tickMsg time.Time
type pollServicesMsg struct{}
type servicesUpdatedMsg struct {
	services []compose.ProcessStatus
	err      error
}
type healthEventMsg health.Event
type logsUpdatedMsg struct {
	logsByService map[string][]string
	err           error
}
type projectStartedMsg struct {
	project string
	err     error
}
type projectStoppedMsg struct {
	project string
	err     error
}

// New creates a new acidBurn model.
func New(cfg *config.Config, reg *registry.Registry) *Model {
	theme := GetTheme(cfg.UI.Theme)
	styles := NewStyles(theme)

	m := &Model{
		config:   cfg,
		registry: reg,
		styles:   styles,
		keys:     DefaultKeyMap(),
		focused:  PaneSidebar,

		// Initialize components
		logView:  NewLogView(styles, 80, 20),
		toast:    NewToastManager(styles, 60),
		alerts:   NewAlertHistory(100),
		settings: NewSettingsPanel(cfg),
		splash:   NewSplashScreen(styles, 80, 24),
		health:       health.NewMonitor(2 * time.Second),
		notifier:     notify.NewNotifier(cfg.Notifications.SystemEnabled),
		clients:      make(map[string]*compose.Client),
		lastLogMsg:   make(map[string]string),
		projectStates: make(map[string]registry.ProjectState),
	}

	// Show splash on startup
	m.showSplash = true
	m.splash.SetMessage("Starting acidBurn...")
	m.splash.SetProgress(0.5)

	// Initialize displayed projects
	m.updateDisplayedProjects()

	return m
}

// Init initializes the model.
func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.tickCmd(),
		m.pollServicesCmd(),
		// Hide splash after a short delay
		tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
			return splashDoneMsg{}
		}),
	)
}

type splashDoneMsg struct{}

func (m *Model) tickCmd() tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m *Model) pollServicesCmd() tea.Cmd {
	return func() tea.Msg {
		return pollServicesMsg{}
	}
}

func (m *Model) pollLogsCmd() tea.Cmd {
	// Capture current state for the closure
	p := m.currentProject()
	services := m.services

	return func() tea.Msg {
		if p == nil || len(services) == 0 {
			return nil
		}

		// Create client in the goroutine
		socketPath := filepath.Join(p.Path, ".devenv", "run", "pc.sock")
		client := compose.NewClient(socketPath)
		if err := client.Connect(); err != nil {
			return nil
		}

		// Fetch logs for all services
		logsByService := make(map[string][]string)
		for _, svc := range services {
			logs, err := client.GetLogs(svc.Name, 0, 100)
			if err == nil && len(logs) > 0 {
				logsByService[svc.Name] = logs
			}
		}

		if len(logsByService) == 0 {
			return nil
		}
		return logsUpdatedMsg{logsByService: logsByService}
	}
}

// startProjectCmd starts an idle project using devenv up -d
func (m *Model) startProjectCmd(p *registry.Project) tea.Cmd {
	projectPath := p.Path
	projectName := p.Name
	return func() tea.Msg {
		cmd := exec.Command("devenv", "up", "-d")
		cmd.Dir = projectPath
		err := cmd.Run()
		return projectStartedMsg{project: projectName, err: err}
	}
}

// stopProjectCmd stops a running project
func (m *Model) stopProjectCmd(p *registry.Project) tea.Cmd {
	projectPath := p.Path
	projectName := p.Name
	return func() tea.Msg {
		// First try to use the API if socket exists
		socketPath := filepath.Join(projectPath, ".devenv", "run", "pc.sock")
		client := compose.NewClient(socketPath)
		if err := client.Connect(); err == nil {
			err = client.ShutdownProject()
			return projectStoppedMsg{project: projectName, err: err}
		}
		// Fallback to devenv down
		cmd := exec.Command("devenv", "down")
		cmd.Dir = projectPath
		err := cmd.Run()
		return projectStoppedMsg{project: projectName, err: err}
	}
}

// Update handles messages and updates the model.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	// Settings mode - delegate ALL messages to settings panel
	if m.showSettings {
		// Check for Esc to close
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			if keyMsg.String() == "esc" {
				m.settings.Cancel()
				m.showSettings = false
				return m, nil
			}
		}

		var cmd tea.Cmd
		m.settings, cmd = m.settings.Update(msg)
		// Check if form is completed
		if !m.settings.IsVisible() {
			m.showSettings = false
			// Apply theme change if saved
			if m.settings.WasSaved() {
				m.styles = NewStyles(GetTheme(m.config.UI.Theme))
			}
		}
		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.logView.SetSize(m.width-m.config.UI.SidebarWidth-8, m.height/2)
		m.splash.SetSize(m.width, m.height)
		m.toast = NewToastManager(m.styles, m.width-10)

	case splashDoneMsg:
		m.showSplash = false
		m.splash.Hide()

	case tickMsg:
		m.updateDisplayedProjects()
		cmds = append(cmds, m.tickCmd())
		cmds = append(cmds, m.pollServicesCmd())

	case pollServicesMsg:
		// Poll the currently selected project
		if p := m.currentProject(); p != nil {
			client := m.getOrCreateClient(p)
			if client != nil {
				status, err := client.GetStatus()
				return m, func() tea.Msg {
					if err != nil {
						return servicesUpdatedMsg{nil, err}
					}
					return servicesUpdatedMsg{status.Processes, nil}
				}
			}
		}

	case servicesUpdatedMsg:
		if msg.err == nil {
			oldServices := m.services
			m.services = msg.services

			// Sort services alphabetically
			sort.Slice(m.services, func(i, j int) bool {
				return m.services[i].Name < m.services[j].Name
			})

			// Update health monitor and check for state changes
			for _, svc := range msg.services {
				projectName := ""
				if p := m.currentProject(); p != nil {
					projectName = p.Name
				}
				event := m.health.UpdateService(projectName, svc.Name, svc.IsRunning, svc.ExitCode)
				if event != nil {
					cmds = append(cmds, func() tea.Msg {
						return healthEventMsg(*event)
					})
				}
			}

			// Clamp selected service
			if m.selectedService >= len(m.services) {
				m.selectedService = len(m.services) - 1
				if m.selectedService < 0 {
					m.selectedService = 0
				}
			}

			// Poll logs after services update
			cmds = append(cmds, m.pollLogsCmd())

			_ = oldServices // Suppress unused warning
		}

	case logsUpdatedMsg:
		if msg.err == nil && len(msg.logsByService) > 0 {
			for service, logs := range msg.logsByService {
				if len(logs) == 0 {
					continue
				}

				// Find where new logs start (after last seen message)
				lastSeen := m.lastLogMsg[service]
				startIdx := 0
				if lastSeen != "" {
					for i, log := range logs {
						if log == lastSeen {
							startIdx = i + 1
							break
						}
					}
				}

				// Add new logs (logs come oldest-first from API)
				for i := startIdx; i < len(logs); i++ {
					// Try to parse timestamp from log line, fall back to now
					ts, ok := ParseLogTimestamp(logs[i])
					if !ok {
						ts = time.Now()
					}
					m.logView.AddEntry(LogEntry{
						Timestamp: ts,
						Service:   service,
						Level:     DetectLogLevel(logs[i]),
						Message:   logs[i],
					})
				}

				// Track last message seen
				m.lastLogMsg[service] = logs[len(logs)-1]
			}
		}

	case healthEventMsg:
		event := health.Event(msg)
		// Add to alert history
		m.alerts.Add(Alert{
			Type:      alertTypeFromHealthEvent(event.Type),
			Project:   event.Project,
			Service:   event.Service,
			Message:   event.Type.String(),
			Timestamp: event.Timestamp,
		})

		// Show toast for crashes
		if event.Type == health.EventServiceCrashed {
			m.toast.Show(
				fmt.Sprintf("%s crashed (exit %d)", event.Service, event.ExitCode),
				ToastError,
				5*time.Second,
			)
			cmds = append(cmds, m.toast.TickCmd())

			// System notification
			if m.notifier.IsEnabled() {
				_ = m.notifier.ServiceCrashed(event.Project, event.Service, event.ExitCode)
			}
		} else if event.Type == health.EventServiceRecovered {
			m.toast.Show(
				fmt.Sprintf("%s recovered", event.Service),
				ToastInfo,
				3*time.Second,
			)
			cmds = append(cmds, m.toast.TickCmd())
		}

	case projectStartedMsg:
		if msg.err != nil {
			m.toast.Show(fmt.Sprintf("Failed to start %s: %v", msg.project, msg.err), ToastError, 5*time.Second)
		} else {
			m.toast.Show(fmt.Sprintf("%s started", msg.project), ToastSuccess, 3*time.Second)
		}
		cmds = append(cmds, m.toast.TickCmd())
		cmds = append(cmds, m.pollServicesCmd())

	case projectStoppedMsg:
		if msg.err != nil {
			m.toast.Show(fmt.Sprintf("Failed to stop %s: %v", msg.project, msg.err), ToastError, 5*time.Second)
		} else {
			m.toast.Show(fmt.Sprintf("%s stopped", msg.project), ToastSuccess, 3*time.Second)
			// Clear services and logs for stopped project
			m.services = nil
			m.logView.buffer.Clear()
		}
		cmds = append(cmds, m.toast.TickCmd())

	case ToastTickMsg:
		var cmd tea.Cmd
		m.toast, cmd = m.toast.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	if len(cmds) > 0 {
		return m, tea.Batch(cmds...)
	}
	return m, nil
}

func alertTypeFromHealthEvent(t health.EventType) AlertType {
	switch t {
	case health.EventServiceCrashed:
		return AlertServiceCrashed
	case health.EventServiceRecovered:
		return AlertServiceRecovered
	case health.EventServiceStarted:
		return AlertProjectStarted
	case health.EventServiceStopped:
		return AlertProjectStopped
	default:
		return AlertInfo
	}
}

func (m *Model) getOrCreateClient(p *registry.Project) *compose.Client {
	if client, ok := m.clients[p.Path]; ok {
		if client.IsConnected() {
			return client
		}
		// Try to reconnect
		delete(m.clients, p.Path)
	}

	// Try to connect to process-compose socket
	// devenv creates a symlink at .devenv/run pointing to /run/user/$UID/devenv-$HASH
	socketPath := filepath.Join(p.Path, ".devenv", "run", "pc.sock")
	client := compose.NewClient(socketPath)
	if err := client.Connect(); err != nil {
		return nil
	}

	m.clients[p.Path] = client
	return client
}

func (m *Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Splash screen - any key dismisses
	if m.showSplash {
		m.showSplash = false
		m.splash.Hide()
		return m, nil
	}

	// Help mode - Esc or ? closes
	if m.showHelp {
		if key.Matches(msg, m.keys.Back) || key.Matches(msg, m.keys.Help) {
			m.showHelp = false
		}
		return m, nil
	}

	// Alert history mode - Esc closes
	if m.showAlerts {
		if key.Matches(msg, m.keys.Back) {
			m.showAlerts = false
		}
		return m, nil
	}

	// Project search input mode (sidebar)
	if m.projectSearchMode {
		switch msg.Type {
		case tea.KeyEsc:
			m.projectSearchMode = false
			m.projectSearchInput = ""
			return m, nil
		case tea.KeyEnter:
			m.projectSearchMode = false
			// Keep filter active, just exit input mode
			return m, nil
		case tea.KeyBackspace:
			if len(m.projectSearchInput) > 0 {
				m.projectSearchInput = m.projectSearchInput[:len(m.projectSearchInput)-1]
			}
			return m, nil
		default:
			if msg.Type == tea.KeyRunes {
				m.projectSearchInput += string(msg.Runes)
			}
			return m, nil
		}
	}

	// Log search input mode
	if m.searchMode {
		switch msg.Type {
		case tea.KeyEsc:
			m.searchMode = false
			m.searchInput = ""
			m.logView.ClearSearch()
			return m, nil
		case tea.KeyEnter:
			m.searchMode = false
			// Keep search active, just exit input mode
			return m, nil
		case tea.KeyBackspace:
			if len(m.searchInput) > 0 {
				m.searchInput = m.searchInput[:len(m.searchInput)-1]
				m.logView.SetSearch(m.searchInput)
			}
			return m, nil
		default:
			if msg.Type == tea.KeyRunes {
				m.searchInput += string(msg.Runes)
				m.logView.SetSearch(m.searchInput)
			}
			return m, nil
		}
	}

	// Global keys
	switch {
	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit
	case key.Matches(msg, m.keys.Shutdown):
		// Shutdown all services
		for _, client := range m.clients {
			_ = client.ShutdownProject()
		}
		return m, tea.Quit
	case key.Matches(msg, m.keys.Help):
		m.showHelp = true
		return m, nil
	case key.Matches(msg, m.keys.Tab):
		m.cycleFocus()
		return m, nil
	case key.Matches(msg, m.keys.Settings):
		cmd := m.settings.Show()
		m.showSettings = true
		return m, cmd
	case key.Matches(msg, m.keys.History):
		m.showAlerts = !m.showAlerts
		return m, nil
	case key.Matches(msg, m.keys.Back):
		return m, nil
	}

	// Navigation keys
	switch {
	case key.Matches(msg, m.keys.Up):
		m.moveUp()
		return m, nil
	case key.Matches(msg, m.keys.Down):
		m.moveDown()
		return m, nil
	}

	// Pane-specific keys
	switch m.focused {
	case PaneSidebar:
		return m.handleSidebarKey(msg)
	case PaneServices:
		return m.handleServicesKey(msg)
	case PaneLogs:
		return m.handleLogsKey(msg)
	}

	return m, nil
}

func (m *Model) handleSidebarKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Search) || msg.String() == "/":
		// / - enter project search mode
		m.projectSearchMode = true
		m.projectSearchInput = ""
		return m, nil
	case key.Matches(msg, m.keys.Back):
		// Esc - clear project search filter
		if m.projectSearchInput != "" {
			m.projectSearchInput = ""
			return m, nil
		}
		return m, nil
	case key.Matches(msg, m.keys.Select):
		// Enter - select project and move to services
		m.focused = PaneServices
		m.selectedService = 0
		m.services = nil // Clear services, will be repopulated
		m.lastLogMsg = make(map[string]string) // Reset log tracking for new project
		m.logView.SetService("")              // Clear service filter
		m.logView.buffer.Clear()              // Clear old logs
		return m, m.pollServicesCmd()
	case key.Matches(msg, m.keys.Start):
		// s - start project
		if p := m.currentProject(); p != nil {
			state := p.DetectState()
			if state == registry.StateIdle || state == registry.StateStale {
				// Start idle project with devenv up -d
				m.toast.Show(fmt.Sprintf("Starting %s...", p.Name), ToastInfo, 3*time.Second)
				return m, tea.Batch(m.startProjectCmd(p), m.toast.TickCmd())
			} else if client := m.getOrCreateClient(p); client != nil {
				// Start all services in running project
				for _, svc := range m.services {
					_ = client.StartProcess(svc.Name)
				}
				return m, m.pollServicesCmd()
			}
		}
		return m, nil
	case key.Matches(msg, m.keys.Stop):
		// x - stop project
		if p := m.currentProject(); p != nil {
			state := p.DetectState()
			if state == registry.StateRunning || state == registry.StateDegraded {
				m.toast.Show(fmt.Sprintf("Stopping %s...", p.Name), ToastInfo, 3*time.Second)
				return m, tea.Batch(m.stopProjectCmd(p), m.toast.TickCmd())
			}
		}
		return m, nil
	}
	return m, nil
}

func (m *Model) handleServicesKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Back):
		// Esc - go back to sidebar
		m.focused = PaneSidebar
		return m, nil
	case key.Matches(msg, m.keys.Select):
		// Enter - filter logs to this service (stay in services pane)
		if m.selectedService < len(m.services) {
			currentFilter := m.logView.GetService()
			selectedName := m.services[m.selectedService].Name
			// Toggle: if already filtered to this service, show all
			if currentFilter == selectedName {
				m.logView.SetService("")
			} else {
				m.logView.SetService(selectedName)
			}
		}
		return m, nil
	case key.Matches(msg, m.keys.Start):
		// s - start service
		if p := m.currentProject(); p != nil {
			if client := m.getOrCreateClient(p); client != nil {
				if m.selectedService < len(m.services) {
					_ = client.StartProcess(m.services[m.selectedService].Name)
				}
			}
		}
		return m, m.pollServicesCmd()
	case key.Matches(msg, m.keys.Stop):
		// x - stop service
		if p := m.currentProject(); p != nil {
			if client := m.getOrCreateClient(p); client != nil {
				if m.selectedService < len(m.services) {
					_ = client.StopProcess(m.services[m.selectedService].Name)
				}
			}
		}
		return m, m.pollServicesCmd()
	case key.Matches(msg, m.keys.Restart):
		// r - restart service
		if p := m.currentProject(); p != nil {
			if client := m.getOrCreateClient(p); client != nil {
				if m.selectedService < len(m.services) {
					_ = client.RestartProcess(m.services[m.selectedService].Name)
				}
			}
		}
		return m, m.pollServicesCmd()
	}
	return m, nil
}

func (m *Model) handleLogsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Search) || msg.String() == "/":
		// / - enter search mode
		m.searchMode = true
		m.searchInput = ""
		return m, nil
	case key.Matches(msg, m.keys.NextMatch):
		// n - next search match
		m.logView.NextMatch()
		return m, nil
	case key.Matches(msg, m.keys.PrevMatch):
		// N - previous search match
		m.logView.PrevMatch()
		return m, nil
	case key.Matches(msg, m.keys.Filter):
		// ctrl+f - toggle filter mode (show only matching lines)
		m.logView.ToggleFilter()
		return m, nil
	case key.Matches(msg, m.keys.Back):
		// Esc - clear service filter and go back to services
		m.logView.SetService("") // Show all logs
		m.logView.ClearSearch()  // Also clear search
		m.focused = PaneServices
		return m, nil
	case key.Matches(msg, m.keys.Follow):
		m.logView.ToggleFollow()
		return m, nil
	case key.Matches(msg, m.keys.Top):
		m.logView.ScrollToTop()
		return m, nil
	case key.Matches(msg, m.keys.Bottom):
		m.logView.ScrollToBottom()
		return m, nil
	case key.Matches(msg, m.keys.Up):
		m.logView.ScrollUp()
		return m, nil
	case key.Matches(msg, m.keys.Down):
		m.logView.ScrollDown()
		return m, nil
	}
	return m, nil
}

func (m *Model) currentProject() *registry.Project {
	if len(m.displayedProjects) == 0 || m.selectedProject >= len(m.displayedProjects) {
		return nil
	}
	return m.displayedProjects[m.selectedProject]
}

// updateDisplayedProjects rebuilds the sidebar display order (active first, then idle, alphabetical within each).
// It also caches the detected states to avoid inconsistent state during rendering.
func (m *Model) updateDisplayedProjects() {
	var active, idle []*registry.Project
	for _, p := range m.registry.Projects {
		state := p.DetectState()
		// Cache the state so we use consistent state during rendering
		m.projectStates[p.Path] = state
		if state == registry.StateRunning || state == registry.StateDegraded {
			active = append(active, p)
		} else {
			idle = append(idle, p)
		}
	}

	// Sort each group alphabetically
	sort.Slice(active, func(i, j int) bool {
		return active[i].Name < active[j].Name
	})
	sort.Slice(idle, func(i, j int) bool {
		return idle[i].Name < idle[j].Name
	})

	m.displayedProjects = append(active, idle...)

	// Clamp selection
	if m.selectedProject >= len(m.displayedProjects) {
		m.selectedProject = len(m.displayedProjects) - 1
		if m.selectedProject < 0 {
			m.selectedProject = 0
		}
	}
}

func (m *Model) cycleFocus() {
	m.focused = (m.focused + 1) % 3
}

func (m *Model) moveUp() {
	switch m.focused {
	case PaneSidebar:
		if m.selectedProject > 0 {
			m.selectedProject--
		}
	case PaneServices:
		if m.selectedService > 0 {
			m.selectedService--
		}
	case PaneLogs:
		m.logView.ScrollUp()
	}
}

func (m *Model) moveDown() {
	switch m.focused {
	case PaneSidebar:
		maxIdx := len(m.displayedProjects) - 1
		if m.selectedProject < maxIdx {
			m.selectedProject++
		}
	case PaneServices:
		if m.selectedService < len(m.services)-1 {
			m.selectedService++
		}
	case PaneLogs:
		m.logView.ScrollDown()
	}
}

// View renders the model.
func (m *Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	// Splash screen
	if m.showSplash {
		return m.splash.View()
	}

	// Help overlay
	if m.showHelp {
		return m.renderHelp()
	}

	// Alert history overlay
	if m.showAlerts {
		return m.renderAlertHistory()
	}

	// Settings overlay
	if m.showSettings {
		return m.settings.View()
	}

	// Main view
	header := m.renderHeader()
	body := m.renderBody()
	footer := m.renderFooter()

	main := lipgloss.JoinVertical(lipgloss.Left, header, body, footer)

	// Toast overlay
	if m.toast.IsVisible() {
		toast := m.toast.View()
		// Position toast at top center
		toastStyle := lipgloss.NewStyle().
			Width(m.width).
			Align(lipgloss.Center)
		main = lipgloss.JoinVertical(lipgloss.Left, toastStyle.Render(toast), main)
	}

	return main
}

func (m *Model) renderHeader() string {
	title := m.styles.Title.Render("acidBurn")
	breadcrumb := m.styles.Breadcrumb.Render(" ─── FLEET")

	if p := m.currentProject(); p != nil {
		breadcrumb = m.styles.Breadcrumb.Render(fmt.Sprintf(" ─── %s", p.Name))
	}

	// Build stats
	activeCount := 0
	for _, svc := range m.services {
		if svc.IsRunning {
			activeCount++
		}
	}
	stats := fmt.Sprintf("Services: %d/%d ── Nix: OK", activeCount, len(m.services))
	statsView := m.styles.StatusBar.Render(stats)

	left := title + breadcrumb
	right := statsView

	gap := m.width - lipgloss.Width(left) - lipgloss.Width(right) - 2
	if gap < 0 {
		gap = 0
	}

	return m.styles.Header.Width(m.width).Render(
		left + fmt.Sprintf("%*s", gap, "") + right,
	)
}

func (m *Model) renderBody() string {
	sidebarWidth := m.config.UI.SidebarWidth
	mainWidth := m.width - sidebarWidth - 4
	bodyHeight := m.height - 4 // header + footer

	sidebar := m.renderSidebar(sidebarWidth, bodyHeight)
	main := m.renderMain(mainWidth, bodyHeight)

	return lipgloss.JoinHorizontal(lipgloss.Top, sidebar, main)
}

func (m *Model) renderSidebar(width, height int) string {
	var content string

	// Filter projects by search if active
	projects := m.displayedProjects
	if m.projectSearchInput != "" {
		filtered := make([]*registry.Project, 0)
		query := strings.ToLower(m.projectSearchInput)
		for _, p := range m.displayedProjects {
			if strings.Contains(strings.ToLower(p.Name), query) {
				filtered = append(filtered, p)
			}
		}
		projects = filtered
	}

	// Count active projects using cached states
	projectCount := len(projects)
	activeCount := 0
	activeEnd := 0
	for i, p := range projects {
		state := m.projectStates[p.Path]
		if state == registry.StateRunning || state == registry.StateDegraded {
			activeCount++
			activeEnd = i + 1
		}
	}

	// Title with focus indicator
	title := fmt.Sprintf("PROJECTS [%d/%d]", activeCount, projectCount)
	if m.focused == PaneSidebar {
		title = "> " + title
	}
	content += m.styles.Title.Render(title) + "\n"

	// Search input or placeholder
	if m.projectSearchMode {
		content += m.styles.SelectedItem.Render(fmt.Sprintf("/%s_", m.projectSearchInput)) + "\n\n"
	} else if m.projectSearchInput != "" {
		content += m.styles.Breadcrumb.Render(fmt.Sprintf("/%s", m.projectSearchInput)) + "\n\n"
	} else {
		content += m.styles.Breadcrumb.Render("/ search...") + "\n\n"
	}

	// Active section
	content += m.styles.Breadcrumb.Render("── ACTIVE ──") + "\n"
	for i := 0; i < activeEnd; i++ {
		content += m.renderProjectItem(i, projects[i]) + "\n"
	}

	// Idle section
	content += "\n" + m.styles.Breadcrumb.Render("── IDLE ──") + "\n"
	for i := activeEnd; i < len(projects); i++ {
		content += m.renderProjectItem(i, projects[i]) + "\n"
	}

	// Global section
	content += "\n" + m.styles.Breadcrumb.Render("── GLOBAL ──") + "\n"
	content += m.styles.ProjectItem.Render("▤ ALL LOGS") + "\n"

	style := m.styles.BlurredBorder
	if m.focused == PaneSidebar {
		style = m.styles.FocusedBorder
	}

	return style.Width(width).Height(height).Render(content)
}

func (m *Model) renderProjectItem(idx int, p *registry.Project) string {
	// Use cached state for consistent rendering
	state := m.projectStates[p.Path]
	var glyph string
	switch state {
	case registry.StateRunning:
		glyph = m.styles.StatusRunning.Render("●")
	case registry.StateDegraded:
		glyph = m.styles.StatusDegraded.Render("◐")
	case registry.StateIdle:
		glyph = m.styles.StatusIdle.Render("○")
	case registry.StateStale:
		glyph = m.styles.StatusStale.Render("✗")
	case registry.StateMissing:
		glyph = m.styles.StatusMissing.Render("✗")
	}

	name := p.Name
	cursor := "  " // No cursor
	if idx == m.selectedProject {
		cursor = "> " // Selection cursor
		if m.focused == PaneSidebar {
			name = m.styles.SelectedItem.Render(name)
		} else {
			// Dimmed selection when not focused
			name = m.styles.Breadcrumb.Render(name)
		}
	} else {
		name = m.styles.ProjectItem.Render(name)
	}

	return fmt.Sprintf("%s%s %s", cursor, glyph, name)
}

func (m *Model) renderMain(width, height int) string {
	servicesHeight := height / 3
	logsHeight := height - servicesHeight - 2

	services := m.renderServices(width, servicesHeight)
	logs := m.renderLogs(width, logsHeight)

	return lipgloss.JoinVertical(lipgloss.Left, services, logs)
}

func (m *Model) renderServices(width, height int) string {
	var content string

	if p := m.currentProject(); p != nil {
		// Title with focus indicator
		title := fmt.Sprintf("[Project: %s]", p.Name)
		if m.focused == PaneServices {
			title = "> " + title
		}
		content = m.styles.Title.Render(title) + "\n\n"

		if len(m.services) == 0 {
			content += m.styles.Breadcrumb.Render("No services running - press 's' to start")
		} else {
			// Table header
			header := fmt.Sprintf("%-18s %-8s %-7s %-6s %-8s %-5s", "SERVICE", "STATUS", "PID", "CPU", "MEM", "EXIT")
			content += m.styles.Breadcrumb.Render(header) + "\n"
			content += m.styles.Breadcrumb.Render("─────────────────────────────────────────────────────────") + "\n"

			for i, svc := range m.services {
				var statusGlyph, status string
				if svc.IsRunning {
					statusGlyph = m.styles.StatusRunning.Render("●")
					status = "Running"
				} else {
					statusGlyph = m.styles.StatusIdle.Render("○")
					status = "Stopped"
				}

				pid := "-"
				if svc.Pid > 0 {
					pid = fmt.Sprintf("%d", svc.Pid)
				}
				exit := "-"
				if svc.ExitCode != 0 {
					exit = fmt.Sprintf("%d", svc.ExitCode)
				}

				// Format CPU (percentage)
				cpu := "-"
				if svc.IsRunning && svc.CPU > 0 {
					cpu = fmt.Sprintf("%.1f%%", svc.CPU)
				}

				// Format memory (human readable)
				mem := "-"
				if svc.IsRunning && svc.Mem > 0 {
					mem = formatBytes(svc.Mem)
				}

				cursor := "  "
				if i == m.selectedService {
					cursor = "> "
				}

				line := fmt.Sprintf("%s%s %-16s %-8s %-7s %-6s %-8s %-5s", cursor, statusGlyph, svc.Name, status, pid, cpu, mem, exit)

				if i == m.selectedService && m.focused == PaneServices {
					line = m.styles.SelectedItem.Render(line)
				} else if i == m.selectedService {
					// Dimmed when not focused
					line = m.styles.Breadcrumb.Render(line)
				}
				content += line + "\n"
			}
		}
	} else {
		content = m.styles.Title.Render("SERVICES") + "\n"
		content += m.styles.Breadcrumb.Render("Select a project from the sidebar")
	}

	style := m.styles.BlurredBorder
	if m.focused == PaneServices {
		style = m.styles.FocusedBorder
	}

	return style.Width(width).Height(height).Render(content)
}

func (m *Model) renderLogs(width, height int) string {
	m.logView.SetSize(width-4, height-4)

	// Title with focus indicator
	title := "LOGS"
	if m.focused == PaneLogs {
		title = "> LOGS"
	}
	content := m.styles.Title.Render(title)

	// Show service filter if active
	if svc := m.logView.GetService(); svc != "" {
		content += m.styles.Breadcrumb.Render(fmt.Sprintf(" [%s]", svc))
	} else {
		content += m.styles.Breadcrumb.Render(" [ALL]")
	}

	if m.logView.IsFollowing() {
		content += m.styles.Breadcrumb.Render(" [FOLLOW]")
	}

	// Show scroll position
	current, total := m.logView.ScrollInfo()
	if total > 0 {
		content += m.styles.Breadcrumb.Render(fmt.Sprintf(" %d/%d", current, total))
	}

	// Show search info
	if m.searchMode {
		// Active input mode - show cursor
		content += m.styles.SelectedItem.Render(fmt.Sprintf(" /%s_", m.searchInput))
	} else if m.logView.IsSearchActive() {
		// Search active but not in input mode
		matchInfo := ""
		if m.logView.MatchCount() > 0 {
			matchInfo = fmt.Sprintf(" %d/%d", m.logView.CurrentMatchIndex(), m.logView.MatchCount())
		} else {
			matchInfo = " (no matches)"
		}
		if m.logView.IsFilterMode() {
			content += m.styles.Breadcrumb.Render(fmt.Sprintf(" [FILTER: %s]%s", m.searchInput, matchInfo))
		} else {
			content += m.styles.Breadcrumb.Render(fmt.Sprintf(" [SEARCH: %s]%s", m.searchInput, matchInfo))
		}
	}
	content += "\n"

	if m.logView.buffer.Len() == 0 {
		content += m.styles.Breadcrumb.Render("No logs yet - logs will appear when services run")
	} else {
		content += m.logView.View()
	}

	style := m.styles.BlurredBorder
	if m.focused == PaneLogs {
		style = m.styles.FocusedBorder
	}

	return style.Width(width).Height(height).Render(content)
}

func (m *Model) renderFooter() string {
	var help string
	switch m.focused {
	case PaneSidebar:
		if m.projectSearchMode {
			help = "[Type] Search  [Enter] Confirm  [Esc] Cancel"
		} else if m.projectSearchInput != "" {
			help = "[/] Edit search  [Esc] Clear  [Enter] Select  [?] Help"
		} else {
			help = "[↑/↓] Navigate  [/] Search  [Enter] Select  [s] Start  [x] Stop  [?] Help"
		}
	case PaneServices:
		help = "[↑/↓] Navigate  [Enter] Filter  [s] Start  [x] Stop  [r] Restart  [?] Help"
	case PaneLogs:
		if m.searchMode {
			help = "[Type] Search  [Enter] Confirm  [Esc] Cancel"
		} else if m.logView.IsSearchActive() {
			help = "[n/N] Next/Prev  [Ctrl+f] Filter  [/] New Search  [Esc] Clear  [?] Help"
		} else {
			help = "[↑/↓] Scroll  [f] Follow  [/] Search  [g/G] Top/Bottom  [?] Help"
		}
	}
	return m.styles.Footer.Width(m.width).Render(help)
}

func (m *Model) renderHelp() string {
	help := `
KEYBINDINGS

  GLOBAL                     NAVIGATION
  q       Quit (detach)      ↑/k     Up
  Ctrl+X  Shutdown all       ↓/j     Down
  S       Settings           Tab     Switch pane
  H       Alert history      Enter   Select/Confirm
  ?       This help          Esc     Back/Cancel

  SIDEBAR                    SERVICES
  s       Start project      s       Start service
  x       Stop project       x       Stop service
                             r       Restart service

  LOGS
  f       Toggle follow      g/G     Top/Bottom
  ↑/↓     Scroll

  SEARCH (in Logs)
  /       Start search       Ctrl+f  Filter mode
  n       Next match         N       Prev match
  Esc     Clear search

                                        [Esc] Close
`
	return m.styles.Main.Width(m.width).Height(m.height).Render(help)
}

func (m *Model) renderAlertHistory() string {
	var content string
	content += m.styles.Title.Render("ALERT HISTORY") + "\n\n"

	alerts := m.alerts.Recent(20)
	if len(alerts) == 0 {
		content += m.styles.Breadcrumb.Render("No alerts yet")
	} else {
		for _, a := range alerts {
			ts := a.Timestamp.Format("15:04:05")
			var icon string
			switch a.Type {
			case AlertServiceCrashed:
				icon = m.styles.StatusStale.Render("✗")
			case AlertServiceRecovered:
				icon = m.styles.StatusRunning.Render("●")
			default:
				icon = m.styles.StatusIdle.Render("○")
			}
			line := fmt.Sprintf("%s %s [%s] %s: %s", ts, icon, a.Project, a.Service, a.Message)
			content += line + "\n"
		}
	}

	content += "\n" + m.styles.Breadcrumb.Render("[Esc] Close")

	return m.styles.Main.Width(m.width).Height(m.height).Render(content)
}

// formatBytes converts bytes to human-readable format (KB, MB, GB).
func formatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.1fG", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.1fM", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.0fK", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%dB", bytes)
	}
}
