package ui

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lucasb-eyer/go-colorful"

	"github.com/infktd/devdash/internal/compose"
	"github.com/infktd/devdash/internal/config"
	"github.com/infktd/devdash/internal/health"
	"github.com/infktd/devdash/internal/notify"
	"github.com/infktd/devdash/internal/packages"
	"github.com/infktd/devdash/internal/registry"
)

// FocusedPane tracks which pane has focus.
type FocusedPane int

const (
	PaneSidebar FocusedPane = iota
	PaneServices
	PanePackages
	PaneLogs
)

// projectListItem implements list.Item for the project list.
type projectListItem struct {
	project *registry.Project
	state   registry.ProjectState
}

func (i projectListItem) FilterValue() string {
	return i.project.Name
}

func (i projectListItem) Title() string {
	return i.project.Name
}

func (i projectListItem) Description() string {
	return i.project.Path
}

// sectionHeaderItem represents a section header in the project list.
type sectionHeaderItem struct {
	title string
}

func (i sectionHeaderItem) FilterValue() string {
	return "" // Headers don't participate in filtering
}

func (i sectionHeaderItem) Title() string {
	return i.title
}

func (i sectionHeaderItem) Description() string {
	return ""
}

// Model is the main application model for devdash.
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
	showSplash   bool
	showPackages bool // When true, show packages pane instead of services (narrow terminals)

	// Search state (for logs)
	searchMode         bool
	searchInput        textinput.Model
	followBeforeSearch bool // Track follow mode state before entering search

	// Project filter state (custom implementation)
	projectFilterMode  bool
	projectFilterInput string

	// Components
	logView       *LogView
	packagesView  *PackagesView
	toast         *ToastManager
	alerts        *AlertHistory
	alertsPanel   *AlertsPanel
	settings      *SettingsPanel
	helpPanel     *HelpPanel
	splash        *SplashScreen
	confirm       *ConfirmDialog
	health        *health.Monitor
	notifier      *notify.Notifier
	spinner       spinner.Model
	servicesTable table.Model
	projectsList  list.Model

	// Loading state
	loadingOp       string    // Description of current operation
	loadingProject  string    // Project being operated on
	loadingProgress float64   // Progress percentage (0.0 to 1.0)
	loadingStage    string    // Current stage description
	loadingStarted  time.Time // When operation started

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

	// Track log activity timestamps per service for flow indicators
	logActivity   map[string]time.Time
	activityFrame int // Animation frame counter for log flow spinner

	// Track service states for detecting transitions
	serviceStates       map[string]string    // Last known state per service
	stateChangeTime     map[string]time.Time // When state last changed
	stateFlashIntensity map[string]float64   // Flash intensity (1.0 = bright, 0.0 = normal)

	// Track resource usage history for sparklines
	cpuHistory map[string][]float64 // Last 10 CPU readings per service
	memHistory map[string][]int64   // Last 10 memory readings per service
}

// Messages for async operations
type tickMsg time.Time
type activityTickMsg time.Time
type progressTickMsg time.Time
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
type serviceOperationMsg struct {
	service   string
	operation string // "start", "stop", "restart"
	err       error
}
type projectDeletedMsg struct {
	project string
}
type projectHiddenMsg struct {
	project string
	hidden  bool
}
type projectRepairedMsg struct {
	project string
	err     error
}
type configEditedMsg struct {
	config *config.Config
	err    error
}

// New creates a new devdash model.
func New(cfg *config.Config, reg *registry.Registry) *Model {
	theme := GetTheme(cfg.UI.Theme)
	styles := NewStyles(theme)

	alertHistory := NewAlertHistory(100)

	// Initialize spinner
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(theme.Primary)

	// Initialize search input
	ti := textinput.New()
	ti.Placeholder = ""
	ti.Prompt = "/"
	ti.PromptStyle = lipgloss.NewStyle().Foreground(theme.Primary)
	ti.TextStyle = lipgloss.NewStyle().Foreground(theme.Primary)
	ti.Cursor.Style = lipgloss.NewStyle().Foreground(theme.Primary)
	ti.CharLimit = 100

	// Initialize services table - proper widths for content
	columns := []table.Column{
		{Title: "STATUS", Width: 10},
		{Title: "SERVICE", Width: 12},  // Reduced from 18 to 12
		{Title: "PID", Width: 8},
		{Title: "CPU", Width: 7},
		{Title: "MEM", Width: 8},
		{Title: "UPTIME", Width: 10},
	}
	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(false),
		table.WithHeight(10),
	)
	tableStyle := table.DefaultStyles()
	tableStyle.Header = tableStyle.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(theme.Muted).
		BorderBottom(true).
		Bold(true).
		Foreground(theme.Primary)
	tableStyle.Selected = tableStyle.Selected.
		Foreground(theme.Primary).
		Bold(true)
	// Default cell padding (0, 1) works well
	t.SetStyles(tableStyle)

	// Initialize projects list with custom delegate (model set later to avoid circular reference)
	projectsDelegate := &projectDelegate{styles: styles}
	projectsList := list.New([]list.Item{}, projectsDelegate, 30, 20)
	projectsList.Title = ""
	projectsList.SetShowStatusBar(false)  // Hide status bar (prevents "2items" counter from section headers)
	projectsList.SetFilteringEnabled(false)  // Disable built-in filter (we'll use custom)
	projectsList.SetShowHelp(false)

	// Apply theme colors directly to list styles and remove all padding
	projectsList.Styles.Title = lipgloss.NewStyle()  // Hide default title
	projectsList.Styles.PaginationStyle = lipgloss.NewStyle()
	projectsList.Styles.HelpStyle = lipgloss.NewStyle()
	// Remove any default top/bottom padding from the list
	projectsList.Styles.NoItems = lipgloss.NewStyle()

	m := &Model{
		config:   cfg,
		registry: reg,
		styles:   styles,
		keys:     DefaultKeyMap(),
		focused:  PaneSidebar,

		// Initialize components
		logView:       NewLogView(styles, 80, 20),
		packagesView:  NewPackagesView(styles),
		toast:         NewToastManager(styles, 60),
		alerts:        alertHistory,
		alertsPanel:   NewAlertsPanel(styles, alertHistory, 80, 24),
		settings:      NewSettingsPanel(cfg, styles, 80, 24),
		helpPanel:     NewHelpPanel(styles, 80, 24),
		splash:        NewSplashScreen(styles, 80, 24),
		confirm:       NewConfirmDialog(styles),
		health:        health.NewMonitor(2 * time.Second),
		notifier:      notify.NewNotifier(cfg.Notifications.SystemEnabled),
		spinner:             s,
		searchInput:         ti,
		servicesTable:       t,
		projectsList:        projectsList,
		clients:             make(map[string]*compose.Client),
		lastLogMsg:          make(map[string]string),
		logActivity:         make(map[string]time.Time),
		projectStates:       make(map[string]registry.ProjectState),
		serviceStates:       make(map[string]string),
		stateChangeTime:     make(map[string]time.Time),
		stateFlashIntensity: make(map[string]float64),
		cpuHistory:          make(map[string][]float64),
		memHistory:          make(map[string][]int64),
	}

	// Show splash on startup
	m.showSplash = true
	m.splash.SetMessage("Starting devdash...")
	m.splash.SetProgress(0.0)

	// Set model reference on delegate (needed for spinner and loading state)
	projectsDelegate.model = m

	// Initialize displayed projects
	m.updateDisplayedProjects()

	return m
}

// handleMouseEvent handles mouse events for focus following.
func (m *Model) handleMouseEvent(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	// Only handle motion and wheel events for focus following
	if msg.Type != tea.MouseMotion && msg.Type != tea.MouseWheelUp && msg.Type != tea.MouseWheelDown {
		return m, nil
	}

	// Skip if modals are open
	if m.showSplash || m.showSettings || m.showHelp || m.alertsPanel.IsVisible() || m.confirm.IsVisible() {
		return m, nil
	}

	// Calculate pane boundaries
	sidebarWidth := m.config.UI.SidebarWidth
	servicesHeight := (m.height - 4) / 3 // Approximate services pane height

	// Determine which pane the mouse is over
	x := msg.X
	y := msg.Y

	var newFocus FocusedPane

	if x < sidebarWidth {
		// Mouse is in sidebar (projects)
		newFocus = PaneSidebar
	} else {
		// Mouse is in main area (services or logs)
		if y < servicesHeight+2 { // +2 for header
			newFocus = PaneServices
		} else {
			newFocus = PaneLogs
		}
	}

	// Update focus if changed
	if m.focused != newFocus {
		m.focused = newFocus
	}

	// Handle mouse wheel scrolling in logs pane
	if newFocus == PaneLogs {
		if msg.Type == tea.MouseWheelUp {
			// Scroll up multiple times for better mouse feel
			for i := 0; i < 3; i++ {
				m.logView.ScrollUp()
			}
		} else if msg.Type == tea.MouseWheelDown {
			// Scroll down multiple times for better mouse feel
			for i := 0; i < 3; i++ {
				m.logView.ScrollDown()
			}
		}
	}

	return m, nil
}

// Init initializes the model.
func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.tickCmd(),
		m.activityTickCmd(),
		m.pollServicesCmd(),
		m.splashTickCmd(),
		m.spinner.Tick,
	)
}

type splashDoneMsg struct{}
type splashTickMsg struct{}

func (m *Model) tickCmd() tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m *Model) activityTickCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return activityTickMsg(t)
	})
}

func (m *Model) progressTickCmd() tea.Cmd {
	return tea.Tick(200*time.Millisecond, func(t time.Time) tea.Msg {
		return progressTickMsg(t)
	})
}

func (m *Model) splashTickCmd() tea.Cmd {
	return tea.Tick(50*time.Millisecond, func(t time.Time) tea.Msg {
		return splashTickMsg{}
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
		output, err := cmd.CombinedOutput()
		if err != nil {
			// Enhance error message with command output
			if len(output) > 0 {
				err = fmt.Errorf("%v: %s", err, string(output))
			}
		}
		return projectStartedMsg{project: projectName, err: err}
	}
}

// startServiceCmd starts a specific service
func (m *Model) startServiceCmd(client *compose.Client, serviceName string) tea.Cmd {
	return func() tea.Msg {
		err := client.StartProcess(serviceName)
		return serviceOperationMsg{
			service:   serviceName,
			operation: "start",
			err:       err,
		}
	}
}

// stopServiceCmd stops a specific service
func (m *Model) stopServiceCmd(client *compose.Client, serviceName string) tea.Cmd {
	return func() tea.Msg {
		err := client.StopProcess(serviceName)
		return serviceOperationMsg{
			service:   serviceName,
			operation: "stop",
			err:       err,
		}
	}
}

// restartServiceCmd restarts a specific service
func (m *Model) restartServiceCmd(client *compose.Client, serviceName string) tea.Cmd {
	return func() tea.Msg {
		err := client.RestartProcess(serviceName)
		return serviceOperationMsg{
			service:   serviceName,
			operation: "restart",
			err:       err,
		}
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
			if err != nil {
				err = fmt.Errorf("API shutdown failed: %v", err)
			}
			return projectStoppedMsg{project: projectName, err: err}
		}
		// Fallback to devenv down
		cmd := exec.Command("devenv", "down")
		cmd.Dir = projectPath
		output, err := cmd.CombinedOutput()
		if err != nil {
			// Enhance error message with command output
			if len(output) > 0 {
				err = fmt.Errorf("%v: %s", err, string(output))
			}
		}
		return projectStoppedMsg{project: projectName, err: err}
	}
}

// editConfigCmd opens the config file in $EDITOR
func (m *Model) editConfigCmd() tea.Cmd {
	// Get editor from environment, default to vi
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}

	// Get config path
	configPath := config.Path()

	// Use tea.ExecProcess to suspend TUI and run editor
	c := exec.Command(editor, configPath)
	return tea.ExecProcess(c, func(err error) tea.Msg {
		if err != nil {
			return configEditedMsg{config: nil, err: err}
		}

		// Reload config after editing
		newConfig, loadErr := config.Load(configPath)
		if loadErr != nil {
			return configEditedMsg{config: nil, err: fmt.Errorf("failed to reload config: %v", loadErr)}
		}

		// Return the new config in the message
		return configEditedMsg{config: newConfig, err: nil}
	})
}

// Update handles messages and updates the model.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	// Settings mode - delegate to settings panel
	if m.showSettings {
		var cmd tea.Cmd
		m.settings, cmd = m.settings.Update(msg)

		// Check if modal was closed (via cancel/Esc)
		if !m.settings.IsVisible() {
			m.showSettings = false
			// Restart tickers that were paused while settings modal was open
			return m, tea.Batch(
				cmd,
				m.tickCmd(),
				m.activityTickCmd(),
				m.pollServicesCmd(),
			)
		}

		return m, cmd
	}

	// Confirm dialog - delegate to confirm dialog
	if m.confirm.IsVisible() {
		var cmd tea.Cmd
		m.confirm, cmd = m.confirm.Update(msg)
		// Check if dialog was closed
		if !m.confirm.IsVisible() {
			// Restart tickers that were paused while confirm dialog was open
			return m, tea.Batch(
				cmd,
				m.tickCmd(),
				m.activityTickCmd(),
				m.pollServicesCmd(),
			)
		}
		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case tea.MouseMsg:
		// Handle mouse events for focus following
		return m.handleMouseEvent(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.logView.SetSize(m.width-m.config.UI.SidebarWidth-8, m.height/2)
		m.splash.SetSize(m.width, m.height)
		m.settings.SetSize(m.width, m.height)
		m.helpPanel.SetSize(m.width, m.height)
		m.alertsPanel.SetSize(m.width, m.height)
		m.toast = NewToastManager(m.styles, m.width-10)

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case splashTickMsg:
		if m.showSplash {
			// Advance animation frame for wave effect
			m.splash.Tick()

			// Increment progress by ~10% each tick (50ms intervals = ~500ms total)
			newProgress := m.splash.Progress() + 0.1
			m.splash.SetProgress(newProgress)

			if newProgress >= 1.0 {
				// Hide splash when complete
				m.showSplash = false
				m.splash.Hide()
			} else {
				// Continue ticking
				cmds = append(cmds, m.splashTickCmd())
			}
		}

	case splashDoneMsg:
		m.showSplash = false
		m.splash.Hide()

	case tickMsg:
		m.updateDisplayedProjects()
		cmds = append(cmds, m.tickCmd())
		cmds = append(cmds, m.pollServicesCmd())

	case activityTickMsg:
		// Increment animation frame and update table if we have services
		if len(m.services) > 0 {
			m.activityFrame++

			// Update flash intensity decay for state transitions
			// Flash decays over ~1.5 seconds with exponential falloff
			for serviceName, changeTime := range m.stateChangeTime {
				elapsed := time.Since(changeTime).Seconds()
				if elapsed >= 1.5 {
					// Flash complete, remove from tracking
					delete(m.stateFlashIntensity, serviceName)
					delete(m.stateChangeTime, serviceName)
				} else {
					// Exponential decay: intensity goes from 1.0 to 0.0 over 1.5 seconds
					// Using e^(-3*t) for smooth fade
					m.stateFlashIntensity[serviceName] = 1.0 * (1.0 - elapsed/1.5)
				}
			}

			m.updateServicesTable()
		}
		cmds = append(cmds, m.activityTickCmd())

	case progressTickMsg:
		// Update loading progress for ongoing operations
		if m.loadingOp != "" && !m.loadingStarted.IsZero() {
			elapsed := time.Since(m.loadingStarted).Seconds()

			// Estimate stages based on operation type
			var stage string

			switch m.loadingOp {
			case "Starting":
				if elapsed < 2.5 {
					stage = "Initializing environment..."
					m.loadingProgress = elapsed / 2.5 * 0.3 // 0-30%
				} else if elapsed < 5.5 {
					stage = "Starting services..."
					m.loadingProgress = 0.3 + ((elapsed-2.5)/3.0)*0.4 // 30-70%
				} else {
					stage = "Services online"
					m.loadingProgress = 0.7 + ((elapsed-5.5)/2.5)*0.3 // 70-100%
				}

			case "Stopping":
				if elapsed < 3.0 {
					stage = "Stopping services..."
					m.loadingProgress = elapsed / 3.0 * 0.6 // 0-60%
				} else {
					stage = "Cleaning up..."
					m.loadingProgress = 0.6 + ((elapsed-3.0)/2.0)*0.4 // 60-100%
				}
			}

			// Cap at 95% until operation actually completes
			if m.loadingProgress > 0.95 {
				m.loadingProgress = 0.95
			}

			m.loadingStage = stage
			cmds = append(cmds, m.progressTickCmd())
		}

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

			// Update table rows
			m.updateServicesTable()

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

				// Detect state transitions for flash animation
				currentState := "Stopped"
				if svc.IsRunning {
					currentState = "Running"
				}
				lastState, exists := m.serviceStates[svc.Name]
				if exists && lastState != currentState {
					// State changed! Record timestamp and start flash animation
					m.stateChangeTime[svc.Name] = time.Now()
					m.stateFlashIntensity[svc.Name] = 1.0 // Start at maximum brightness
				}
				m.serviceStates[svc.Name] = currentState

				// Record CPU and memory history for sparklines (keep last 10 readings)
				if svc.IsRunning {
					// Update CPU history
					cpuHist := m.cpuHistory[svc.Name]
					cpuHist = append(cpuHist, svc.CPU)
					if len(cpuHist) > 10 {
						cpuHist = cpuHist[len(cpuHist)-10:] // Keep only last 10
					}
					m.cpuHistory[svc.Name] = cpuHist

					// Update memory history
					memHist := m.memHistory[svc.Name]
					memHist = append(memHist, svc.Mem)
					if len(memHist) > 10 {
						memHist = memHist[len(memHist)-10:] // Keep only last 10
					}
					m.memHistory[svc.Name] = memHist
				}
			}

			// Clamp selected service
			if m.selectedService >= len(m.services) {
				m.selectedService = len(m.services) - 1
				if m.selectedService < 0 {
					m.selectedService = 0
				}
			}
			// Update table cursor
			m.servicesTable.SetCursor(m.selectedService)

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
				newLogCount := 0
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
					newLogCount++
				}

				// Track last message seen
				m.lastLogMsg[service] = logs[len(logs)-1]

				// Record log activity timestamp if new logs were added
				if newLogCount > 0 {
					m.logActivity[service] = time.Now()
				}
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
		// Clear loading state
		m.loadingOp = ""
		m.loadingProject = ""
		m.loadingProgress = 0
		m.loadingStage = ""
		m.loadingStarted = time.Time{}
		if msg.err != nil {
			m.toast.Show(fmt.Sprintf("Failed to start %s: %v", msg.project, msg.err), ToastError, 5*time.Second)
		} else {
			m.toast.Show(fmt.Sprintf("%s started", msg.project), ToastSuccess, 3*time.Second)
			// Update project state cache to Running to prevent repeated start attempts
			if p := m.currentProject(); p != nil && p.Name == msg.project {
				m.projectStates[p.Path] = registry.StateRunning
			}
		}
		cmds = append(cmds, m.toast.TickCmd())
		cmds = append(cmds, m.pollServicesCmd())

	case projectStoppedMsg:
		// Clear loading state
		m.loadingOp = ""
		m.loadingProject = ""
		m.loadingProgress = 0
		m.loadingStage = ""
		m.loadingStarted = time.Time{}
		if msg.err != nil {
			m.toast.Show(fmt.Sprintf("Failed to stop %s: %v", msg.project, msg.err), ToastError, 5*time.Second)
		} else {
			m.toast.Show(fmt.Sprintf("%s stopped", msg.project), ToastSuccess, 3*time.Second)
			// Clear services and logs for stopped project
			m.services = nil
			m.logView.buffer.Clear()
			// Update project state cache to Idle to prevent state confusion
			if p := m.currentProject(); p != nil && p.Name == msg.project {
				m.projectStates[p.Path] = registry.StateIdle
			}
		}
		cmds = append(cmds, m.toast.TickCmd())

	case serviceOperationMsg:
		if msg.err != nil {
			m.toast.Show(
				fmt.Sprintf("Failed to %s %s: %v", msg.operation, msg.service, msg.err),
				ToastError,
				5*time.Second,
			)
		} else {
			m.toast.Show(
				fmt.Sprintf("%s %sed", msg.service, msg.operation),
				ToastSuccess,
				2*time.Second,
			)
		}
		cmds = append(cmds, m.toast.TickCmd())
		cmds = append(cmds, m.pollServicesCmd())

	case settingsSavedMsg:
		m.toast.Show("Settings saved", ToastSuccess, 2*time.Second)
		// Reload styles if theme changed
		m.styles = NewStyles(GetTheme(m.config.UI.Theme))

		// Update delegate with new styles
		newDelegate := &projectDelegate{styles: m.styles, model: m}
		m.projectsList.SetDelegate(newDelegate)

		// Update table styles
		tableStyle := table.DefaultStyles()
		tableStyle.Header = tableStyle.Header.
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(m.styles.theme.Muted).
			BorderBottom(true).
			Bold(true).
			Foreground(m.styles.theme.Primary)
		tableStyle.Selected = tableStyle.Selected.
			Foreground(m.styles.theme.Primary).
			Bold(true)
		// Keep default cell padding
		m.servicesTable.SetStyles(tableStyle)

		// Update spinner style
		m.spinner.Style = lipgloss.NewStyle().Foreground(m.styles.theme.Primary)

		// Restart tickers that were paused while settings modal was open
		cmds = append(cmds, m.tickCmd())
		cmds = append(cmds, m.activityTickCmd())
		cmds = append(cmds, m.pollServicesCmd())
		cmds = append(cmds, m.toast.TickCmd())

	case settingsSaveErrorMsg:
		m.toast.Show(fmt.Sprintf("Failed to save settings: %v", msg.err), ToastError, 5*time.Second)
		cmds = append(cmds, m.toast.TickCmd())

	case projectDeletedMsg:
		m.toast.Show(fmt.Sprintf("%s removed from registry", msg.project), ToastSuccess, 2*time.Second)
		m.updateDisplayedProjects()
		cmds = append(cmds, m.toast.TickCmd())

	case projectHiddenMsg:
		if msg.hidden {
			m.toast.Show(fmt.Sprintf("%s hidden", msg.project), ToastInfo, 2*time.Second)
		} else {
			m.toast.Show(fmt.Sprintf("%s shown", msg.project), ToastInfo, 2*time.Second)
		}
		m.updateDisplayedProjects()
		cmds = append(cmds, m.toast.TickCmd())

	case projectRepairedMsg:
		if msg.err != nil {
			m.toast.Show(fmt.Sprintf("Failed to repair %s: %v", msg.project, msg.err), ToastError, 5*time.Second)
		} else {
			m.toast.Show(fmt.Sprintf("%s repaired - ready to start", msg.project), ToastSuccess, 3*time.Second)
		}
		m.updateDisplayedProjects()
		cmds = append(cmds, m.toast.TickCmd())

	case configEditedMsg:
		if msg.err != nil {
			m.toast.Show(fmt.Sprintf("Failed to edit config: %v", msg.err), ToastError, 5*time.Second)
		} else {
			// Update config in model
			m.config = msg.config
			// Reload styles with new theme
			m.styles = NewStyles(GetTheme(m.config.UI.Theme))
			// Update settings panel with new config
			m.settings = NewSettingsPanel(m.config, m.styles, m.width, m.height)
			m.toast.Show("Config reloaded", ToastSuccess, 2*time.Second)
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

	// Help mode - delegate to help panel
	if m.showHelp {
		var cmd tea.Cmd
		m.helpPanel, cmd = m.helpPanel.Update(msg)

		// Check if modal was closed
		if !m.helpPanel.IsVisible() {
			m.showHelp = false
			// Restart tickers that were paused while help modal was open
			return m, tea.Batch(
				cmd,
				m.tickCmd(),
				m.activityTickCmd(),
				m.pollServicesCmd(),
			)
		}

		return m, cmd
	}

	// Alerts modal - delegate to panel
	if m.alertsPanel.IsVisible() {
		_, cmd := m.alertsPanel.Update(msg)
		// Check if modal was closed
		if !m.alertsPanel.IsVisible() {
			// Restart tickers that were paused while alerts modal was open
			return m, tea.Batch(
				cmd,
				m.tickCmd(),
				m.activityTickCmd(),
				m.pollServicesCmd(),
			)
		}
		return m, cmd
	}

	// Log search input mode
	if m.searchMode {
		switch msg.Type {
		case tea.KeyEsc:
			m.searchMode = false
			m.searchInput.Reset()
			m.searchInput.Blur()
			m.logView.ClearSearch()
			// Restore follow mode when canceling search
			m.logView.SetFollow(m.followBeforeSearch)
			return m, nil
		case tea.KeyEnter:
			m.searchMode = false
			m.searchInput.Blur()
			// Keep search active, just exit input mode
			return m, nil
		default:
			// Delegate to textinput for all other keys
			var cmd tea.Cmd
			m.searchInput, cmd = m.searchInput.Update(msg)
			m.logView.SetSearch(m.searchInput.Value())
			return m, cmd
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
		m.helpPanel.Show()
		m.showHelp = true
		return m, nil
	case key.Matches(msg, m.keys.Tab):
		m.cycleFocus()
		return m, nil
	case key.Matches(msg, m.keys.ShiftTab):
		m.cycleFocusReverse()
		return m, nil
	case key.Matches(msg, m.keys.Settings):
		cmd := m.settings.Show()
		m.showSettings = true
		return m, cmd
	case key.Matches(msg, m.keys.EditConfig):
		return m, m.editConfigCmd()
	case key.Matches(msg, m.keys.History):
		if m.alertsPanel.IsVisible() {
			m.alertsPanel.Hide()
		} else {
			m.alertsPanel.Show()
		}
		return m, nil
	case msg.String() == "p":
		// Toggle between packages and services view
		return m, m.togglePackagesView()
	case key.Matches(msg, m.keys.Back):
		// Don't handle Esc globally if sidebar is filtering or logs has active search
		if m.focused == PaneSidebar && m.projectFilterMode {
			// Let sidebar handler deal with it
			return m.handleSidebarKey(msg)
		}
		if m.focused == PaneLogs && m.logView.IsSearchActive() {
			// Let logs handler deal with it
			return m.handleLogsKey(msg)
		}
		return m, nil
	}

	// Navigation keys (skip if sidebar is filtering - those keys might be part of search)
	if m.focused != PaneSidebar || !m.projectFilterMode {
		switch {
		case key.Matches(msg, m.keys.Up):
			cmd := m.moveUp()
			return m, cmd
		case key.Matches(msg, m.keys.Down):
			cmd := m.moveDown()
			return m, cmd
		}
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
	// Handle custom project filter mode
	if m.projectFilterMode {
		switch msg.Type {
		case tea.KeyEsc:
			// Exit filter mode and clear
			m.projectFilterMode = false
			m.projectFilterInput = ""
			m.updateDisplayedProjects()
			return m, nil
		case tea.KeyEnter:
			// Keep filter applied, just exit input mode
			m.projectFilterMode = false
			return m, nil
		case tea.KeyBackspace:
			if len(m.projectFilterInput) > 0 {
				m.projectFilterInput = m.projectFilterInput[:len(m.projectFilterInput)-1]
				m.updateDisplayedProjects()
			}
			return m, nil
		case tea.KeyRunes:
			m.projectFilterInput += string(msg.Runes)
			m.updateDisplayedProjects()
			return m, nil
		}
		return m, nil
	}

	switch {
	case key.Matches(msg, m.keys.Search) || msg.String() == "/":
		// Enter custom filter mode
		m.projectFilterMode = true
		m.projectFilterInput = ""
		return m, nil
	case key.Matches(msg, m.keys.Select):
		// Enter - move focus to services pane
		// Services are already displayed from cursor movement
		m.focused = PaneServices
		return m, nil
	case key.Matches(msg, m.keys.Start):
		// s - start project
		if p := m.currentProject(); p != nil {
			// Don't allow starting if ANY loading operation is in progress
			if m.loadingOp != "" {
				return m, nil
			}

			// Use cached state if available, otherwise detect fresh
			state, hasCached := m.projectStates[p.Path]
			if !hasCached {
				state = p.DetectState()
			}

			if state == registry.StateIdle || state == registry.StateStale {
				// Start idle project with devenv up -d
				m.loadingOp = "Starting"
				m.loadingProject = p.Name
				m.loadingProgress = 0.0
				m.loadingStage = "Initializing..."
				m.loadingStarted = time.Now()

				// Immediately update cache to prevent re-entry
				m.projectStates[p.Path] = registry.StateRunning

				m.toast.Show(fmt.Sprintf("Starting %s...", p.Name), ToastInfo, 3*time.Second)
				return m, tea.Batch(m.startProjectCmd(p), m.toast.TickCmd(), m.progressTickCmd())
			} else if state == registry.StateRunning || state == registry.StateDegraded {
				// Project already running - don't show progress, just inform
				m.toast.Show("Project already running", ToastInfo, 2*time.Second)
				return m, m.toast.TickCmd()
			}
		}
		return m, nil
	case key.Matches(msg, m.keys.Stop):
		// x - stop project
		if p := m.currentProject(); p != nil {
			// Don't allow stopping if ANY loading operation is in progress
			if m.loadingOp != "" {
				return m, nil
			}

			state, hasCached := m.projectStates[p.Path]
			if !hasCached {
				state = p.DetectState()
			}

			if state == registry.StateRunning || state == registry.StateDegraded {
				m.loadingOp = "Stopping"
				m.loadingProject = p.Name
				m.loadingProgress = 0.0
				m.loadingStage = "Stopping..."
				m.loadingStarted = time.Now()

				// Immediately update cache to prevent re-entry
				m.projectStates[p.Path] = registry.StateIdle

				m.toast.Show(fmt.Sprintf("Stopping %s...", p.Name), ToastInfo, 3*time.Second)
				return m, tea.Batch(m.stopProjectCmd(p), m.toast.TickCmd(), m.progressTickCmd())
			} else {
				m.toast.Show("Project not running", ToastInfo, 2*time.Second)
				return m, m.toast.TickCmd()
			}
		}
		return m, nil
	case key.Matches(msg, m.keys.Delete):
		// d - delete project (with confirmation)
		if p := m.currentProject(); p != nil {
			projectPath := p.Path
			projectName := p.Name
			m.confirm.Show(
				fmt.Sprintf("Remove %s from registry?", projectName),
				func() tea.Msg {
					// Delete the project
					if m.registry.RemoveProject(projectPath) {
						// Save registry
						regPath := registry.Path()
						_ = registry.Save(regPath, m.registry)
						return projectDeletedMsg{project: projectName}
					}
					return nil
				},
				func() tea.Msg {
					// Cancel - do nothing
					return nil
				},
			)
		}
		return m, nil
	case key.Matches(msg, m.keys.Hide):
		// ctrl+h - toggle hidden
		if p := m.currentProject(); p != nil {
			projectPath := p.Path
			projectName := p.Name
			if m.registry.ToggleHidden(projectPath) {
				// Save registry
				regPath := registry.Path()
				_ = registry.Save(regPath, m.registry)
				return m, func() tea.Msg {
					return projectHiddenMsg{project: projectName, hidden: p.Hidden}
				}
			}
		}
		return m, nil
	case key.Matches(msg, m.keys.Repair):
		// c - repair stale project
		if p := m.currentProject(); p != nil {
			state := p.DetectState()
			if state == registry.StateStale {
				projectName := p.Name
				m.confirm.Show(
					fmt.Sprintf("Clean up stale files for %s?", projectName),
					func() tea.Msg {
						// Repair the project
						err := p.Repair()
						return projectRepairedMsg{project: projectName, err: err}
					},
					func() tea.Msg {
						// Cancel - do nothing
						return nil
					},
				)
			}
		}
		return m, nil
	}

	// Don't pass navigation keys to list component since we handle them ourselves
	// This prevents rendering glitches when the list tries to handle navigation too
	if key.Matches(msg, m.keys.Up) || key.Matches(msg, m.keys.Down) {
		return m, nil
	}

	// Pass message to list component for filtering (handles '/' key)
	var cmd tea.Cmd
	m.projectsList, cmd = m.projectsList.Update(msg)

	// Sync selection after list update (in case filter changed)
	if selectedItem := m.projectsList.SelectedItem(); selectedItem != nil {
		if _, isProject := selectedItem.(projectListItem); isProject {
			m.selectedProject = m.listIndexToProjectIndex(m.projectsList.Index())
		}
	}

	return m, cmd
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
					return m, m.startServiceCmd(client, m.services[m.selectedService].Name)
				}
			}
		}
		return m, nil
	case key.Matches(msg, m.keys.Stop):
		// x - stop service
		if p := m.currentProject(); p != nil {
			if client := m.getOrCreateClient(p); client != nil {
				if m.selectedService < len(m.services) {
					return m, m.stopServiceCmd(client, m.services[m.selectedService].Name)
				}
			}
		}
		return m, nil
	case key.Matches(msg, m.keys.Restart):
		// r - restart service
		if p := m.currentProject(); p != nil {
			if client := m.getOrCreateClient(p); client != nil {
				if m.selectedService < len(m.services) {
					return m, m.restartServiceCmd(client, m.services[m.selectedService].Name)
				}
			}
		}
		return m, nil
	}
	return m, nil
}

func (m *Model) handleLogsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Search) || msg.String() == "/":
		// / - enter search mode
		m.followBeforeSearch = m.logView.IsFollowing() // Save follow state
		m.searchMode = true
		m.searchInput.Reset()
		m.searchInput.Focus()
		return m, m.searchInput.Cursor.BlinkCmd()
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
		// Esc - if search active, clear search first. Otherwise go back to services.
		if m.logView.IsSearchActive() {
			m.logView.ClearSearch()
			m.searchInput.Reset()
			// Restore follow mode to previous state
			m.logView.SetFollow(m.followBeforeSearch)
			return m, nil
		}
		// No search active, go back to services
		m.logView.SetService("") // Show all logs
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

const (
	minWidthForBothPanes = 140
)

// shouldShowBothPanes determines if there's enough space to show both Services and Packages panes.
func (m *Model) shouldShowBothPanes() bool {
	return m.width >= minWidthForBothPanes
}

// togglePackagesView switches between Services and Packages pane (for narrow terminals).
func (m *Model) togglePackagesView() tea.Cmd {
	m.showPackages = !m.showPackages
	return nil
}

// updateDisplayedProjects rebuilds the sidebar display order (active first, then idle, alphabetical within each).
// It also caches the detected states to avoid inconsistent state during rendering.
func (m *Model) updateDisplayedProjects() {
	var active, idle []*registry.Project
	for _, p := range m.registry.Projects {
		state := p.DetectState()
		// Cache the state so we use consistent state during rendering
		m.projectStates[p.Path] = state

		// Apply filter if active
		if m.projectFilterInput != "" {
			if !strings.Contains(strings.ToLower(p.Name), strings.ToLower(m.projectFilterInput)) {
				continue // Skip projects that don't match filter
			}
		}

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

	// Update list items with section headers
	var items []list.Item

	// Add ACTIVE section
	if len(active) > 0 {
		items = append(items, sectionHeaderItem{title: "ACTIVE"})
		for _, p := range active {
			items = append(items, projectListItem{
				project: p,
				state:   m.projectStates[p.Path],
			})
		}
	}

	// Add IDLE section
	if len(idle) > 0 {
		items = append(items, sectionHeaderItem{title: "IDLE"})
		for _, p := range idle {
			items = append(items, projectListItem{
				project: p,
				state:   m.projectStates[p.Path],
			})
		}
	}

	m.projectsList.SetItems(items)

	// Convert project index to list index (accounting for headers)
	listIndex := m.projectIndexToListIndex(m.selectedProject)
	m.projectsList.Select(listIndex)

	// Clamp selection
	if m.selectedProject >= len(m.displayedProjects) {
		m.selectedProject = len(m.displayedProjects) - 1
		if m.selectedProject < 0 {
			m.selectedProject = 0
		}
	}
	listIndex = m.projectIndexToListIndex(m.selectedProject)
	m.projectsList.Select(listIndex)

	// Ensure we're not on a header (this shouldn't happen, but safety check)
	if len(m.projectsList.Items()) > 0 {
		if _, isHeader := m.projectsList.SelectedItem().(sectionHeaderItem); isHeader {
			// Move to first non-header item
			for i, item := range m.projectsList.Items() {
				if _, isHeader := item.(sectionHeaderItem); !isHeader {
					m.projectsList.Select(i)
					m.selectedProject = m.listIndexToProjectIndex(i)
					break
				}
			}
		}
	}
}

// projectIndexToListIndex converts a project index to a list index (accounting for section headers).
func (m *Model) projectIndexToListIndex(projectIndex int) int {
	if projectIndex < 0 {
		return 0
	}

	listIndex := 0
	projectsSeen := 0

	for _, item := range m.projectsList.Items() {
		if _, isHeader := item.(sectionHeaderItem); isHeader {
			listIndex++
			continue
		}
		if projectsSeen == projectIndex {
			return listIndex
		}
		projectsSeen++
		listIndex++
	}

	return listIndex
}

// listIndexToProjectIndex converts a list index to a project index (skipping section headers).
func (m *Model) listIndexToProjectIndex(listIndex int) int {
	projectIndex := 0
	for i := 0; i < listIndex && i < len(m.projectsList.Items()); i++ {
		if _, isHeader := m.projectsList.Items()[i].(sectionHeaderItem); !isHeader {
			projectIndex++
		}
	}
	return projectIndex
}

func (m *Model) cycleFocus() {
	switch m.focused {
	case PaneSidebar:
		m.focused = PaneServices
	case PaneServices:
		// If both panes visible, go to packages next
		// Otherwise skip to logs
		if m.shouldShowBothPanes() {
			m.focused = PanePackages
		} else {
			m.focused = PaneLogs
		}
	case PanePackages:
		m.focused = PaneLogs
	case PaneLogs:
		m.focused = PaneSidebar
	default:
		m.focused = PaneSidebar
	}
}

func (m *Model) cycleFocusReverse() {
	switch m.focused {
	case PaneSidebar:
		m.focused = PaneLogs
	case PaneServices:
		m.focused = PaneSidebar
	case PanePackages:
		m.focused = PaneServices
	case PaneLogs:
		// If both panes visible, go to packages previous
		// Otherwise skip to services
		if m.shouldShowBothPanes() {
			m.focused = PanePackages
		} else {
			m.focused = PaneServices
		}
	default:
		m.focused = PaneSidebar
	}
}

// switchToCurrentProject updates the services display for the currently selected project
func (m *Model) switchToCurrentProject() {
	m.selectedService = 0
	m.services = nil // Clear services, will be repopulated
	m.lastLogMsg = make(map[string]string)           // Reset log tracking for new project
	m.logActivity = make(map[string]time.Time)       // Reset log activity tracking
	m.serviceStates = make(map[string]string)        // Reset state tracking
	m.stateChangeTime = make(map[string]time.Time)   // Reset state change times
	m.stateFlashIntensity = make(map[string]float64) // Reset flash intensity
	m.cpuHistory = make(map[string][]float64)        // Reset CPU history
	m.memHistory = make(map[string][]int64)          // Reset memory history
	m.logView.SetService("")                         // Clear service filter
	m.logView.buffer.Clear()                         // Clear old logs

	// Scan packages for new project
	project := m.currentProject()
	if project != nil {
		pkgs, err := packages.Scan(project.Path)
		if err != nil {
			// Log error but don't block
			pkgs = []packages.Package{}
		}
		m.packagesView.SetPackages(pkgs)
	} else {
		m.packagesView.SetPackages([]packages.Package{})
	}
}

func (m *Model) moveUp() tea.Cmd {
	switch m.focused {
	case PaneSidebar:
		// Use list's cursor movement
		currentIdx := m.projectsList.Index()
		if currentIdx > 0 {
			m.projectsList.CursorUp()

			// Skip over headers
			for m.projectsList.Index() > 0 {
				if _, isHeader := m.projectsList.SelectedItem().(sectionHeaderItem); !isHeader {
					break
				}
				m.projectsList.CursorUp()
			}

			// Update selectedProject to match
			oldProject := m.selectedProject
			m.selectedProject = m.listIndexToProjectIndex(m.projectsList.Index())

			// If project changed, update services display immediately
			if oldProject != m.selectedProject {
				m.switchToCurrentProject()
				return m.pollServicesCmd()
			}
		}
	case PaneServices:
		if m.selectedService > 0 {
			m.selectedService--
			m.servicesTable.SetCursor(m.selectedService)
		}
	case PaneLogs:
		m.logView.ScrollUp()
	}
	return nil
}

func (m *Model) moveDown() tea.Cmd {
	switch m.focused {
	case PaneSidebar:
		// Use list's cursor movement
		maxIdx := len(m.projectsList.Items()) - 1
		if m.projectsList.Index() < maxIdx {
			m.projectsList.CursorDown()

			// Skip over headers
			for m.projectsList.Index() < maxIdx {
				if _, isHeader := m.projectsList.SelectedItem().(sectionHeaderItem); !isHeader {
					break
				}
				m.projectsList.CursorDown()
			}

			// Update selectedProject to match
			oldProject := m.selectedProject
			m.selectedProject = m.listIndexToProjectIndex(m.projectsList.Index())

			// If project changed, update services display immediately
			if oldProject != m.selectedProject {
				m.switchToCurrentProject()
				return m.pollServicesCmd()
			}
		}
	case PaneServices:
		if m.selectedService < len(m.services)-1 {
			m.selectedService++
			m.servicesTable.SetCursor(m.selectedService)
		}
	case PaneLogs:
		m.logView.ScrollDown()
	}
	return nil
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

	// Main view
	header := m.renderHeader()
	body := m.renderBody()
	footer := m.renderFooter()

	main := lipgloss.JoinVertical(lipgloss.Left, header, body, footer)

	// Settings modal overlay (centered on screen)
	if m.showSettings {
		settingsModal := m.settings.View()
		// Place modal centered on a dark background
		main = lipgloss.Place(
			m.width,
			m.height,
			lipgloss.Center,
			lipgloss.Center,
			settingsModal,
			lipgloss.WithWhitespaceChars(" "),
			lipgloss.WithWhitespaceForeground(lipgloss.Color("#1a1a1a")),
		)
	}

	// Help modal overlay (centered on screen)
	if m.showHelp {
		helpModal := m.helpPanel.View()
		// Place modal centered on a dark background
		main = lipgloss.Place(
			m.width,
			m.height,
			lipgloss.Center,
			lipgloss.Center,
			helpModal,
			lipgloss.WithWhitespaceChars(" "),
			lipgloss.WithWhitespaceForeground(lipgloss.Color("#1a1a1a")),
		)
	}

	// Alerts modal overlay (centered on screen)
	if m.alertsPanel.IsVisible() {
		alertsModal := m.alertsPanel.View()
		// Place modal centered on a dark background
		main = lipgloss.Place(
			m.width,
			m.height,
			lipgloss.Center,
			lipgloss.Center,
			alertsModal,
			lipgloss.WithWhitespaceChars(" "),
			lipgloss.WithWhitespaceForeground(lipgloss.Color("#1a1a1a")),
		)
	}

	// Confirm dialog overlay (centered, transparent background)
	if m.confirm.IsVisible() {
		confirmModal := m.confirm.View()
		// Place modal centered - background shows through
		main = lipgloss.Place(
			m.width,
			m.height,
			lipgloss.Center,
			lipgloss.Center,
			confirmModal,
		)
	}

	return main
}

func (m *Model) renderHeader() string {
	title := m.styles.Title.Render("devdash")
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

	// Add toast to header if visible
	right := statsView
	if m.toast.IsVisible() {
		toast := m.toast.View()
		// Toast replaces stats when visible
		right = toast
	}

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
	// Title line
	titleLine := m.renderSectionTitle("PROJECTS", m.focused == PaneSidebar, width-4)

	// Custom filter UI
	var filterLine string
	listHeight := height  // Use full height - let border handle spacing
	if m.projectFilterMode || m.projectFilterInput != "" {
		filterPromptStyle := lipgloss.NewStyle().
			Foreground(m.styles.theme.Primary).
			Bold(true)
		filterInputStyle := lipgloss.NewStyle().
			Foreground(m.styles.theme.Primary)

		if m.projectFilterMode {
			// Show cursor when in input mode
			filterLine = filterPromptStyle.Render("Filter: ") + filterInputStyle.Render(m.projectFilterInput+"_")
		} else {
			// Show filter without cursor when applied but not in input mode
			filterLine = filterPromptStyle.Render("Filter: ") + filterInputStyle.Render(m.projectFilterInput)
		}
		listHeight = height - 1 // Just reduce by filter line
	}

	// Update list size and render
	m.projectsList.SetSize(width-4, listHeight)
	listView := m.projectsList.View()

	// Aggressively remove leading blank lines by splitting and rejoining
	lines := strings.Split(listView, "\n")
	// Skip leading empty lines
	for len(lines) > 0 && strings.TrimSpace(lines[0]) == "" {
		lines = lines[1:]
	}
	listView = strings.Join(lines, "\n")

	// Build content - minimal gaps
	content := titleLine + "\n" + listView
	if filterLine != "" {
		content = titleLine + "\n" + filterLine + "\n" + listView
	}

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

	// Show spinner if this project is currently loading
	if m.loadingProject == p.Name && m.loadingOp != "" {
		glyph = m.spinner.View()
	} else {
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
	}

	name := p.Name

	// Add hidden icon if project is hidden
	if p.Hidden {
		name = name + " ■"
	}

	cursor := "  " // No cursor
	if idx == m.selectedProject {
		cursor = "> " // Selection cursor
		if m.focused == PaneSidebar {
			if p.Hidden {
				// Hidden project selected - dimmed even when focused
				name = m.styles.Breadcrumb.Render(name)
			} else {
				name = m.styles.SelectedItem.Render(name)
			}
		} else {
			// Dimmed selection when not focused
			name = m.styles.Breadcrumb.Render(name)
		}
	} else {
		if p.Hidden {
			// Hidden project not selected - extra dimmed
			name = m.styles.Breadcrumb.Render(name)
		} else {
			name = m.styles.ProjectItem.Render(name)
		}
	}

	return fmt.Sprintf("%s%s %s", cursor, glyph, name)
}

func (m *Model) renderMain(width, height int) string {
	// Determine layout based on width
	if m.shouldShowBothPanes() {
		// Wide terminal: show Services, Packages, and Logs stacked vertically
		servicesHeight := height / 4
		packagesHeight := height / 4
		logsHeight := height - servicesHeight - packagesHeight - 4

		services := m.renderServices(width, servicesHeight)
		packages := m.renderPackages(width, packagesHeight)
		logs := m.renderLogs(width, logsHeight)

		return lipgloss.JoinVertical(lipgloss.Left, services, packages, logs)
	} else {
		// Narrow terminal: show either Services or Packages, plus Logs
		mainPaneHeight := height / 3
		logsHeight := height - mainPaneHeight - 2

		var mainPane string
		if m.showPackages {
			mainPane = m.renderPackages(width, mainPaneHeight)
		} else {
			mainPane = m.renderServices(width, mainPaneHeight)
		}

		logs := m.renderLogs(width, logsHeight)

		return lipgloss.JoinVertical(lipgloss.Left, mainPane, logs)
	}
}

func (m *Model) renderServices(width, height int) string {
	var content string

	if p := m.currentProject(); p != nil {
		// Title with focus indicator and toggle hint on narrow terminals
		title := fmt.Sprintf("SERVICES [%s]", p.Name)
		if !m.shouldShowBothPanes() {
			title += " [p:packages]"
		}
		content = m.renderSectionTitle(title, m.focused == PaneServices, width-4) + "\n"

		if len(m.services) == 0 {
			// Enhanced empty state with better visual hierarchy
			emptyStateHeight := height - 8
			emptyContent := m.renderEmptyState(
				"No Services Running",
				[]string{
					"This project has no active services",
					"",
					"Press 's' to start the project",
					"Press '?' for more commands",
				},
				width-6,
				emptyStateHeight,
			)
			content += emptyContent
		} else {
			// Update table focus based on pane focus
			m.servicesTable.SetHeight(height - 6)
			if m.focused == PaneServices {
				m.servicesTable.Focus()
			} else {
				m.servicesTable.Blur()
			}

			// Render the table
			content += m.servicesTable.View()
		}
	} else {
		title := "SERVICES"
		if !m.shouldShowBothPanes() {
			title += " [p:packages]"
		}
		content = m.renderSectionTitle(title, m.focused == PaneServices, width-4) + "\n"
		emptyStateHeight := height - 8
		emptyContent := m.renderEmptyState(
			"No Project Selected",
			[]string{
				"Select a project from the sidebar",
				"to view its services",
				"",
				"Use ↑/↓ or Tab to navigate",
			},
			width-6,
			emptyStateHeight,
		)
		content += emptyContent
	}

	style := m.styles.BlurredBorder
	if m.focused == PaneServices {
		style = m.styles.FocusedBorder
	}

	return style.Width(width).Height(height).Render(content)
}

func (m *Model) renderPackages(width, height int) string {
	// Update packagesView size and focus state
	m.packagesView.SetSize(width-4, height-4)
	m.packagesView.SetFocused(m.focused == PanePackages)

	// Add toggle indicator on narrow terminals
	if !m.shouldShowBothPanes() {
		m.packagesView.SetTitleSuffix("[p:services]")
	} else {
		m.packagesView.SetTitleSuffix("")
	}

	// Get packages view content
	packagesContent := m.packagesView.View()

	// Wrap in border
	style := m.styles.BlurredBorder
	if m.focused == PanePackages {
		style = m.styles.FocusedBorder
	}

	return style.Width(width).Height(height).Render(packagesContent)
}

func (m *Model) renderLogs(width, height int) string {
	m.logView.SetSize(width-4, height-4)

	// Title with focus indicator
	title := "LOGS"
	content := m.renderSectionTitle(title, m.focused == PaneLogs, width-4) + "\n"

	// Build status line with filters
	var statusParts []string

	// Show service filter if active
	if svc := m.logView.GetService(); svc != "" {
		statusParts = append(statusParts, fmt.Sprintf("[%s]", svc))
	} else {
		statusParts = append(statusParts, "[ALL]")
	}

	if m.logView.IsFollowing() {
		statusParts = append(statusParts, "[FOLLOW]")
	}

	// Show scroll position
	current, total := m.logView.ScrollInfo()
	if total > 0 {
		statusParts = append(statusParts, fmt.Sprintf("%d/%d", current, total))
	}

	// Show search info
	if m.searchMode {
		// Active input mode - show textinput with cursor
		statusParts = append(statusParts, m.searchInput.View())
	} else if m.logView.IsSearchActive() {
		// Search active but not in input mode
		matchInfo := ""
		if m.logView.MatchCount() > 0 {
			matchInfo = fmt.Sprintf(" %d/%d", m.logView.CurrentMatchIndex(), m.logView.MatchCount())
		} else {
			matchInfo = " (no matches)"
		}
		searchTerm := m.searchInput.Value()
		if m.logView.IsFilterMode() {
			statusParts = append(statusParts, fmt.Sprintf("[FILTER: %s]%s", searchTerm, matchInfo))
		} else {
			statusParts = append(statusParts, fmt.Sprintf("[SEARCH: %s]%s", searchTerm, matchInfo))
		}
	}

	// Combine status parts
	if len(statusParts) > 0 {
		content += m.styles.Breadcrumb.Render(strings.Join(statusParts, " ")) + "\n"
	} else {
		content += "\n"
	}

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

// highlightKeys highlights keybinds in brackets with theme colors
func (m *Model) highlightKeys(text string) string {
	// Simple approach: highlight text within brackets
	result := ""
	inBracket := false
	bracketContent := ""

	for _, ch := range text {
		if ch == '[' {
			inBracket = true
			bracketContent = ""
			result += lipgloss.NewStyle().Foreground(m.styles.theme.Muted).Render("[")
		} else if ch == ']' {
			inBracket = false
			result += lipgloss.NewStyle().
				Foreground(m.styles.theme.Primary).
				Bold(true).
				Render(bracketContent)
			result += lipgloss.NewStyle().Foreground(m.styles.theme.Muted).Render("]")
		} else if inBracket {
			bracketContent += string(ch)
		} else {
			result += string(ch)
		}
	}

	return result
}

// updateServicesTable updates the table rows from the services list.
func (m *Model) updateServicesTable() {
	rows := make([]table.Row, len(m.services))
	for i, svc := range m.services {
		var status string
		if svc.IsRunning {
			status = "Running"
		} else {
			status = "Stopped"
		}

		// Simple values - let table component handle width and padding
		pid := "-"
		if svc.Pid > 0 {
			pid = fmt.Sprintf("%d", svc.Pid)
		}

		cpu := "-"
		if svc.IsRunning {
			cpu = fmt.Sprintf("%.1f%%", svc.CPU)
		}

		mem := "-"
		if svc.IsRunning && svc.Mem > 0 {
			mem = formatBytes(svc.Mem)
		}

		uptimeOrExit := "-"
		if svc.IsRunning && svc.SystemTime != "" {
			uptimeOrExit = formatSystemTime(svc.SystemTime)
		} else if svc.ExitCode != 0 {
			uptimeOrExit = fmt.Sprintf("exit %d", svc.ExitCode)
		}

		// Apply flash effect to status if state recently changed
		styledStatus := m.applyFlashEffect(status, svc.Name, svc.IsRunning)

		// Get activity indicator (animated spinner if logs received recently)
		activity := m.getActivityIndicator(svc.Name)

		// Combine status and activity
		statusWithActivity := styledStatus + " " + activity

		// Simple row - table handles all width management
		rows[i] = table.Row{
			statusWithActivity,
			svc.Name,
			pid,
			cpu,
			mem,
			uptimeOrExit,
		}
	}
	m.servicesTable.SetRows(rows)
}

// applyFlashEffect applies a pulse/flash animation to status text after state changes.
// Returns the status text with styling applied based on current flash intensity.
func (m *Model) applyFlashEffect(status string, serviceName string, isRunning bool) string {
	intensity, hasFlash := m.stateFlashIntensity[serviceName]
	if !hasFlash || intensity <= 0 {
		// No flash, return plain status
		return status
	}

	// Choose color based on service state
	var baseColor lipgloss.Color
	if isRunning {
		baseColor = m.styles.theme.Success // Green for running
	} else {
		baseColor = m.styles.theme.Warning // Yellow/orange for stopped
	}

	// Apply flash styling: bold and use primary color mixed with base color
	// As intensity decreases, color transitions from bright primary to normal base color
	flashStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(baseColor)

	return flashStyle.Render(status)
}

// getActivityIndicator returns the activity indicator for a service.
// Shows animated braille spinner if logs received within last 2 seconds, otherwise empty.
func (m *Model) getActivityIndicator(serviceName string) string {
	// Braille spinner frames (8 frames for smooth animation)
	frames := []string{"⣾", "⣽", "⣻", "⢿", "⡿", "⣟", "⣯", "⣷"}

	lastActivity, exists := m.logActivity[serviceName]
	if !exists {
		return " " // No activity yet
	}

	// Show spinner if activity within last 2 seconds
	if time.Since(lastActivity) < 2*time.Second {
		frameIndex := m.activityFrame % len(frames)
		return frames[frameIndex]
	}

	return " " // Activity too old, show nothing
}

// renderSparkline generates a sparkline from a slice of values using Unicode block characters.
// Returns a string like "▁▂▃▄▅▆▇█" representing the trend.
func renderSparkline(values []float64) string {
	if len(values) == 0 {
		return ""
	}

	// Unicode block characters from empty to full
	blocks := []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

	// Find min and max for normalization
	min, max := values[0], values[0]
	for _, v := range values {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}

	// Avoid division by zero
	if max == min {
		// All values the same, use mid-level block
		result := make([]rune, len(values))
		for i := range result {
			result[i] = blocks[3]
		}
		return string(result)
	}

	// Normalize and map to blocks
	result := make([]rune, len(values))
	for i, v := range values {
		normalized := (v - min) / (max - min) // 0.0 to 1.0
		blockIndex := int(normalized * float64(len(blocks)-1))
		if blockIndex >= len(blocks) {
			blockIndex = len(blocks) - 1
		}
		result[i] = blocks[blockIndex]
	}

	return string(result)
}

// renderMemorySparkline generates a sparkline from memory values (int64).
func renderMemorySparkline(values []int64) string {
	if len(values) == 0 {
		return ""
	}

	// Convert to float64 for rendering
	floatValues := make([]float64, len(values))
	for i, v := range values {
		floatValues[i] = float64(v)
	}

	return renderSparkline(floatValues)
}

// renderCompactProgress renders a compact progress indicator for the sidebar.
// Example output: "devdash [████░░] Starting..."
func (m *Model) renderCompactProgress(progress float64, stage string) string {
	// Get current project name
	projectName := m.loadingProject
	if p := m.currentProject(); p != nil && p.Name == m.loadingProject {
		projectName = p.Name
	}

	// Compact bar - just 6 chars
	barWidth := 6
	filledWidth := int(float64(barWidth) * progress)
	if filledWidth > barWidth {
		filledWidth = barWidth
	}

	// Get theme colors
	primaryHex := string(m.styles.theme.Primary)
	secondaryHex := string(m.styles.theme.Secondary)
	mutedHex := string(m.styles.theme.Muted)

	// Parse colors
	primaryCol, _ := colorful.Hex(primaryHex)
	secondaryCol, _ := colorful.Hex(secondaryHex)

	// Build mini gradient bar
	var bar strings.Builder
	bar.WriteString("[")

	for i := 0; i < barWidth; i++ {
		if i < filledWidth {
			var gradientPos float64
			if filledWidth > 1 {
				gradientPos = float64(i) / float64(filledWidth-1)
			}
			blendedColor := primaryCol.BlendLuv(secondaryCol, gradientPos)
			charStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(blendedColor.Hex()))
			bar.WriteString(charStyle.Render("█"))
		} else {
			emptyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(mutedHex))
			bar.WriteString(emptyStyle.Render("░"))
		}
	}
	bar.WriteString("]")

	// Compact stage text - just the first word or shortened version
	compactStage := stage
	if len(stage) > 12 {
		compactStage = stage[:12] + "..."
	}

	// Build: "projectName [bar] stage"
	nameStyle := m.styles.SelectedItem
	stageStyle := lipgloss.NewStyle().Foreground(m.styles.theme.Muted)

	return nameStyle.Render(projectName) + " " + bar.String() + " " + stageStyle.Render(compactStage)
}

// renderProgressBar renders a progress bar with percentage and stage description.
// Example output: "[██████░░░░] 60% Starting services..."
func (m *Model) renderProgressBar(progress float64, stage string, width int) string {
	if width < 20 {
		width = 20 // Minimum width
	}

	// Calculate bar width (reserve space for brackets, percentage, and stage)
	// Format: "[bar] xx% stage"
	barWidth := 10 // Fixed bar width for consistency
	percentage := int(progress * 100)

	// Build the bar using block characters with gradient
	filledWidth := int(float64(barWidth) * progress)
	if filledWidth > barWidth {
		filledWidth = barWidth
	}

	// Get theme colors for gradient
	primaryHex := string(m.styles.theme.Primary)
	secondaryHex := string(m.styles.theme.Secondary)
	mutedHex := string(m.styles.theme.Muted)

	// Parse colors for gradient blending
	primaryCol, _ := colorful.Hex(primaryHex)
	secondaryCol, _ := colorful.Hex(secondaryHex)

	// Build gradient bar
	var result strings.Builder
	result.WriteString("[")

	for i := 0; i < barWidth; i++ {
		if i < filledWidth {
			// Calculate gradient position
			var gradientPos float64
			if filledWidth > 1 {
				gradientPos = float64(i) / float64(filledWidth-1)
			}
			blendedColor := primaryCol.BlendLuv(secondaryCol, gradientPos)
			charStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(blendedColor.Hex()))
			result.WriteString(charStyle.Render("█"))
		} else {
			emptyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(mutedHex))
			result.WriteString(emptyStyle.Render("░"))
		}
	}

	result.WriteString("] ")
	result.WriteString(fmt.Sprintf("%2d%% ", percentage))

	// Stage text in muted color
	stageStyle := lipgloss.NewStyle().Foreground(m.styles.theme.Muted)
	result.WriteString(stageStyle.Render(stage))

	return result.String()
}

// renderEmptyState renders a centered empty state message.
func (m *Model) renderEmptyState(title string, lines []string, width, height int) string {
	var content []string

	// Title in primary color
	titleStyle := lipgloss.NewStyle().
		Foreground(m.styles.theme.Primary).
		Bold(true)
	content = append(content, titleStyle.Render(title))
	content = append(content, "")

	// Message lines in muted color
	messageStyle := lipgloss.NewStyle().
		Foreground(m.styles.theme.Muted)

	for _, line := range lines {
		if line == "" {
			content = append(content, "")
		} else {
			content = append(content, messageStyle.Render(line))
		}
	}

	// Join all lines
	joined := lipgloss.JoinVertical(lipgloss.Left, content...)

	// Center the content
	return lipgloss.Place(
		width,
		height,
		lipgloss.Center,
		lipgloss.Center,
		joined,
	)
}

// renderSectionTitle renders a section title in code comment style with separator line.
// Uses primary color if focused, muted color otherwise.
func (m *Model) renderSectionTitle(title string, focused bool, width int) string {
	// Calculate separator line length
	// Format: "// " (3) + title + " " (1) + separator
	// We want the line to extend close to the edge
	separatorLen := width - len(title) - 4 + 4 // +4 to extend closer to edge
	if separatorLen < 0 {
		separatorLen = 0
	}
	separator := strings.Repeat("─", separatorLen)

	var style lipgloss.Style
	if focused {
		style = lipgloss.NewStyle().
			Foreground(m.styles.theme.Primary).
			Bold(true)
	} else {
		style = lipgloss.NewStyle().
			Foreground(m.styles.theme.Muted)
	}

	return style.Render("// " + title + " " + separator)
}

func (m *Model) renderFooter() string {
	var help string
	switch m.focused {
	case PaneSidebar:
		// Check if current project is stale to show repair option
		isStale := false
		if p := m.currentProject(); p != nil {
			state := p.DetectState()
			isStale = (state == registry.StateStale)
		}
		if isStale {
			help = "[↑/↓] Navigate  [Tab] Switch Pane  [c] Repair  [d] Delete  [Ctrl+h] Hide  [?] Help"
		} else {
			help = "[↑/↓] Navigate  [Tab] Switch Pane  [/] Search  [Enter] Select  [s] Start  [x] Stop  [d] Delete  [Ctrl+h] Hide  [?] Help"
		}
	case PaneServices:
		help = "[↑/↓] Navigate  [Tab] Switch Pane  [Enter] Filter  [s] Start  [x] Stop  [r] Restart  [?] Help"
	case PaneLogs:
		if m.searchMode {
			help = "[Type] Search  [Enter] Confirm  [Esc] Cancel"
		} else if m.logView.IsSearchActive() {
			help = "[n/N] Next/Prev  [Ctrl+f] Filter  [/] New Search  [Esc] Clear  [?] Help"
		} else {
			help = "[↑/↓] Scroll  [Tab] Switch Pane  [f] Follow  [/] Search  [g/G] Top/Bottom  [?] Help"
		}
	}

	// Highlight the keybinds
	highlightedHelp := m.highlightKeys(help)
	return lipgloss.NewStyle().
		Width(m.width).
		Align(lipgloss.Center).
		Render(highlightedHelp)
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

// formatDuration formats a duration in a compact human-readable form.
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		hours := int(d.Hours())
		mins := int(d.Minutes()) % 60
		if mins > 0 {
			return fmt.Sprintf("%dh%dm", hours, mins)
		}
		return fmt.Sprintf("%dh", hours)
	}
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	if hours > 0 {
		return fmt.Sprintf("%dd%dh", days, hours)
	}
	return fmt.Sprintf("%dd", days)
}

// formatSystemTime formats the system_time string from process-compose API.
// The API returns time in format like "1h2m3s" or "2m30.5s".
func formatSystemTime(sysTime string) string {
	// Try to parse as Go duration
	d, err := time.ParseDuration(sysTime)
	if err != nil {
		// If parsing fails, return as-is but truncated
		if len(sysTime) > 8 {
			return sysTime[:8]
		}
		return sysTime
	}
	return formatDuration(d)
}

// projectDelegate is a custom list delegate for rendering projects.
type projectDelegate struct {
	styles *Styles
	model  *Model // Reference to parent model for spinner/loading state
}

// NewProjectDelegate creates a new project delegate.
func NewProjectDelegate(styles *Styles, model *Model) list.ItemDelegate {
	return &projectDelegate{styles: styles, model: model}
}

func (d *projectDelegate) Height() int { return 1 }
func (d *projectDelegate) Spacing() int {
	// Subtle spacing between project items for better readability
	return 0 // Keep at 0 for now - spacing makes list too sparse in sidebar
}
func (d *projectDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }

func (d *projectDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	// Check if this is a section header
	if headerItem, ok := item.(sectionHeaderItem); ok {
		headerStyle := lipgloss.NewStyle().
			Foreground(d.styles.theme.Muted).
			Bold(true)
		// Enhanced visual separator with subtle styling
		separator := lipgloss.NewStyle().Foreground(d.styles.theme.Muted).Render("─")
		fmt.Fprintf(w, " %s %s", headerStyle.Render(headerItem.title), separator)
		return
	}

	projItem, ok := item.(projectListItem)
	if !ok {
		return
	}

	// Get status glyph (or progress indicator if loading)
	var glyph string
	var loadingDisplay string
	if d.model != nil && d.model.loadingProject == projItem.project.Name && d.model.loadingOp != "" {
		// Show compact progress indicator when this project is loading
		loadingDisplay = d.model.renderCompactProgress(d.model.loadingProgress, d.model.loadingStage)
		glyph = "" // Don't show status glyph during loading
	} else {
		switch projItem.state {
		case registry.StateRunning:
			glyph = d.styles.StatusRunning.Render("●")
		case registry.StateDegraded:
			glyph = d.styles.StatusDegraded.Render("◐")
		case registry.StateIdle:
			glyph = d.styles.StatusIdle.Render("○")
		case registry.StateStale:
			glyph = d.styles.StatusStale.Render("✗")
		case registry.StateMissing:
			glyph = d.styles.StatusMissing.Render("✗")
		}
	}

	name := projItem.project.Name
	if projItem.project.Hidden {
		name = name + " ■"
	}

	// Check if selected - use our model's selection state for consistency
	// This avoids potential flicker from list component's internal state
	isSelected := false
	if d.model != nil {
		expectedListIndex := d.model.projectIndexToListIndex(d.model.selectedProject)
		isSelected = index == expectedListIndex
	} else {
		// Fallback to list's index if model not set
		isSelected = index == m.Index()
	}
	cursor := "   "  // Indent project items under section headers
	if isSelected {
		cursor = " > "
		name = d.styles.SelectedItem.Render(name)
	} else {
		if projItem.project.Hidden {
			name = d.styles.Breadcrumb.Render(name)
		} else {
			name = d.styles.ProjectItem.Render(name)
		}
	}

	// Show loading indicator during loading, otherwise show normal status + name
	if loadingDisplay != "" {
		fmt.Fprintf(w, "%s%s", cursor, loadingDisplay)
	} else {
		fmt.Fprintf(w, "%s%s %s", cursor, glyph, name)
	}
}
