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

func TestNotifierServiceCrashed(t *testing.T) {
	n := NewNotifier(true)

	// When enabled, call should succeed (may fail if beeep not available in test env)
	err := n.ServiceCrashed("test-project", "test-service", 1)
	// Don't fail on beeep errors in test environment
	_ = err

	// When disabled, should not error
	n.SetEnabled(false)
	err = n.ServiceCrashed("test-project", "test-service", 1)
	if err != nil {
		t.Errorf("ServiceCrashed() should not error when disabled: %v", err)
	}
}

func TestNotifierServiceRecovered(t *testing.T) {
	n := NewNotifier(true)

	err := n.ServiceRecovered("test-project", "test-service")
	// Don't fail on beeep errors in test environment
	_ = err

	n.SetEnabled(false)
	err = n.ServiceRecovered("test-project", "test-service")
	if err != nil {
		t.Errorf("ServiceRecovered() should not error when disabled: %v", err)
	}
}

func TestNotifierProjectStarted(t *testing.T) {
	n := NewNotifier(true)

	err := n.ProjectStarted("test-project")
	// Don't fail on beeep errors in test environment
	_ = err

	n.SetEnabled(false)
	err = n.ProjectStarted("test-project")
	if err != nil {
		t.Errorf("ProjectStarted() should not error when disabled: %v", err)
	}
}

func TestNotifierProjectStopped(t *testing.T) {
	n := NewNotifier(true)

	err := n.ProjectStopped("test-project")
	// Don't fail on beeep errors in test environment
	_ = err

	n.SetEnabled(false)
	err = n.ProjectStopped("test-project")
	if err != nil {
		t.Errorf("ProjectStopped() should not error when disabled: %v", err)
	}
}

func TestNotifierCritical(t *testing.T) {
	n := NewNotifier(true)

	err := n.Critical("Critical error", "Something bad happened")
	// Don't fail on beeep errors in test environment
	_ = err

	// When disabled
	n.SetEnabled(false)
	err = n.Critical("Title", "Body")
	if err != nil {
		t.Errorf("Critical() should not error when disabled: %v", err)
	}
}

func TestNotifierInfo(t *testing.T) {
	n := NewNotifier(true)

	err := n.Info("Info message", "Details here")
	// Don't fail on beeep errors in test environment
	_ = err

	// When disabled
	n.SetEnabled(false)
	err = n.Info("Title", "Body")
	if err != nil {
		t.Errorf("Info() should not error when disabled: %v", err)
	}
}
