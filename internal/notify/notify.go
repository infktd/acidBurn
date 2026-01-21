package notify

import (
	"fmt"

	"github.com/gen2brain/beeep"
)

// Notifier handles system notifications.
type Notifier struct {
	enabled bool
}

// NewNotifier creates a system notifier.
func NewNotifier(enabled bool) *Notifier {
	return &Notifier{enabled: enabled}
}

// SetEnabled enables or disables notifications.
func (n *Notifier) SetEnabled(enabled bool) {
	n.enabled = enabled
}

// IsEnabled returns whether notifications are enabled.
func (n *Notifier) IsEnabled() bool {
	return n.enabled
}

// ServiceCrashed sends a notification for a crashed service.
func (n *Notifier) ServiceCrashed(project, service string, exitCode int) error {
	if !n.enabled {
		return nil
	}
	title := "acidBurn: Service Crashed"
	body := fmt.Sprintf("%s in %s exited with code %d", service, project, exitCode)
	return beeep.Alert(title, body, "")
}

// ServiceRecovered sends a notification for a recovered service.
func (n *Notifier) ServiceRecovered(project, service string) error {
	if !n.enabled {
		return nil
	}
	title := "acidBurn: Service Recovered"
	body := fmt.Sprintf("%s in %s is now running", service, project)
	return beeep.Notify(title, body, "")
}

// ProjectStarted sends a notification when a project starts.
func (n *Notifier) ProjectStarted(project string) error {
	if !n.enabled {
		return nil
	}
	title := "acidBurn: Project Started"
	body := fmt.Sprintf("%s is now running", project)
	return beeep.Notify(title, body, "")
}

// ProjectStopped sends a notification when a project stops.
func (n *Notifier) ProjectStopped(project string) error {
	if !n.enabled {
		return nil
	}
	title := "acidBurn: Project Stopped"
	body := fmt.Sprintf("%s has been stopped", project)
	return beeep.Notify(title, body, "")
}

// Critical sends a critical alert (e.g., Nix daemon down).
func (n *Notifier) Critical(title, message string) error {
	if !n.enabled {
		return nil
	}
	return beeep.Alert(title, message, "")
}

// Info sends an informational notification.
func (n *Notifier) Info(title, message string) error {
	if !n.enabled {
		return nil
	}
	return beeep.Notify(title, message, "")
}
