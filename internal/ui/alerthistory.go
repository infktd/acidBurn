package ui

import (
	"fmt"
	"sync"
	"time"
)

// AlertType represents the type of alert.
type AlertType int

const (
	AlertServiceCrashed AlertType = iota
	AlertServiceRecovered
	AlertProjectStarted
	AlertProjectStopped
	AlertCritical
	AlertInfo
)

func (t AlertType) String() string {
	switch t {
	case AlertServiceCrashed:
		return "crashed"
	case AlertServiceRecovered:
		return "recovered"
	case AlertProjectStarted:
		return "started"
	case AlertProjectStopped:
		return "stopped"
	case AlertCritical:
		return "critical"
	case AlertInfo:
		return "info"
	default:
		return "unknown"
	}
}

// Alert represents a stored alert.
type Alert struct {
	Type      AlertType
	Project   string
	Service   string // May be empty for project-level alerts
	Message   string
	Timestamp time.Time
}

// AlertHistory stores past alerts.
type AlertHistory struct {
	alerts   []Alert
	capacity int
	mu       sync.RWMutex
}

// NewAlertHistory creates a new alert history with given capacity.
func NewAlertHistory(capacity int) *AlertHistory {
	return &AlertHistory{
		alerts:   make([]Alert, 0, capacity),
		capacity: capacity,
	}
}

// Add appends an alert, removing oldest if at capacity.
func (h *AlertHistory) Add(alert Alert) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if len(h.alerts) >= h.capacity {
		// Remove oldest (first element)
		h.alerts = h.alerts[1:]
	}
	h.alerts = append(h.alerts, alert)
}

// AddServiceCrashed is a convenience method.
func (h *AlertHistory) AddServiceCrashed(project, service string, exitCode int) {
	h.Add(Alert{
		Type:      AlertServiceCrashed,
		Project:   project,
		Service:   service,
		Message:   fmt.Sprintf("exited with code %d", exitCode),
		Timestamp: time.Now(),
	})
}

// AddServiceRecovered is a convenience method.
func (h *AlertHistory) AddServiceRecovered(project, service string) {
	h.Add(Alert{
		Type:      AlertServiceRecovered,
		Project:   project,
		Service:   service,
		Message:   "recovered",
		Timestamp: time.Now(),
	})
}

// All returns all alerts (newest last).
func (h *AlertHistory) All() []Alert {
	h.mu.RLock()
	defer h.mu.RUnlock()

	result := make([]Alert, len(h.alerts))
	copy(result, h.alerts)
	return result
}

// Recent returns the most recent n alerts (newest first).
func (h *AlertHistory) Recent(n int) []Alert {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if n > len(h.alerts) {
		n = len(h.alerts)
	}
	if n == 0 {
		return []Alert{}
	}

	result := make([]Alert, n)
	// Copy from end to beginning (reverse order)
	for i := 0; i < n; i++ {
		result[i] = h.alerts[len(h.alerts)-1-i]
	}
	return result
}

// Len returns the number of alerts.
func (h *AlertHistory) Len() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.alerts)
}

// Clear removes all alerts.
func (h *AlertHistory) Clear() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.alerts = make([]Alert, 0, h.capacity)
}
