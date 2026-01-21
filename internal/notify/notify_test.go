package notify

import (
	"testing"
)

func TestNotifierEnabled(t *testing.T) {
	n := NewNotifier(true)
	if !n.IsEnabled() {
		t.Error("notifier should be enabled")
	}

	n.SetEnabled(false)
	if n.IsEnabled() {
		t.Error("notifier should be disabled")
	}
}

func TestNotifierDisabled(t *testing.T) {
	n := NewNotifier(false)

	// All methods should return nil without sending when disabled
	if err := n.ServiceCrashed("proj", "svc", 1); err != nil {
		t.Errorf("disabled notifier should return nil, got %v", err)
	}
	if err := n.ServiceRecovered("proj", "svc"); err != nil {
		t.Errorf("disabled notifier should return nil, got %v", err)
	}
	if err := n.ProjectStarted("proj"); err != nil {
		t.Errorf("disabled notifier should return nil, got %v", err)
	}
	if err := n.ProjectStopped("proj"); err != nil {
		t.Errorf("disabled notifier should return nil, got %v", err)
	}
	if err := n.Critical("title", "msg"); err != nil {
		t.Errorf("disabled notifier should return nil, got %v", err)
	}
	if err := n.Info("title", "msg"); err != nil {
		t.Errorf("disabled notifier should return nil, got %v", err)
	}
}

// Note: We don't test actual beeep calls as they'd create real notifications
// The important thing is the enabled/disabled gating works
