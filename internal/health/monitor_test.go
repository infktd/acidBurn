package health

import (
	"testing"
	"time"
)

func TestMonitorUpdateServiceFirstSeen(t *testing.T) {
	m := NewMonitor(time.Second)
	defer m.Close()

	// First time seeing a running service - should emit started
	event := m.UpdateService("proj", "svc", true, 0)

	if event == nil {
		t.Fatal("expected event")
	}
	if event.Type != EventServiceStarted {
		t.Errorf("expected started, got %v", event.Type)
	}
}

func TestMonitorServiceCrash(t *testing.T) {
	m := NewMonitor(time.Second)
	defer m.Close()

	// Start service
	m.UpdateService("proj", "svc", true, 0)

	// Crash with exit code
	event := m.UpdateService("proj", "svc", false, 1)

	if event == nil {
		t.Fatal("expected crash event")
	}
	if event.Type != EventServiceCrashed {
		t.Errorf("expected crashed, got %v", event.Type)
	}
	if event.ExitCode != 1 {
		t.Errorf("expected exit code 1, got %d", event.ExitCode)
	}
}

func TestMonitorServiceStopped(t *testing.T) {
	m := NewMonitor(time.Second)
	defer m.Close()

	// Start service
	m.UpdateService("proj", "svc", true, 0)

	// Stop cleanly (exit 0)
	event := m.UpdateService("proj", "svc", false, 0)

	if event == nil {
		t.Fatal("expected stopped event")
	}
	if event.Type != EventServiceStopped {
		t.Errorf("expected stopped, got %v", event.Type)
	}
}

func TestMonitorServiceRecovered(t *testing.T) {
	m := NewMonitor(time.Second)
	defer m.Close()

	// Start then crash
	m.UpdateService("proj", "svc", true, 0)
	m.UpdateService("proj", "svc", false, 1)

	// Recover
	event := m.UpdateService("proj", "svc", true, 0)

	if event == nil {
		t.Fatal("expected recovered event")
	}
	if event.Type != EventServiceRecovered {
		t.Errorf("expected recovered, got %v", event.Type)
	}
}

func TestMonitorNoEventOnSameState(t *testing.T) {
	m := NewMonitor(time.Second)
	defer m.Close()

	// Start
	m.UpdateService("proj", "svc", true, 0)

	// Update with same state
	event := m.UpdateService("proj", "svc", true, 0)

	if event != nil {
		t.Error("expected no event on same state")
	}
}

func TestMonitorGetState(t *testing.T) {
	m := NewMonitor(time.Second)
	defer m.Close()

	m.UpdateService("proj", "svc", true, 0)

	state, exists := m.GetState("proj", "svc")
	if !exists {
		t.Fatal("state should exist")
	}
	if !state.Running {
		t.Error("state should be running")
	}

	_, exists = m.GetState("proj", "other")
	if exists {
		t.Error("unknown service should not exist")
	}
}

func TestMonitorClearStates(t *testing.T) {
	m := NewMonitor(time.Second)
	defer m.Close()

	m.UpdateService("proj", "svc", true, 0)
	m.ClearStates()

	_, exists := m.GetState("proj", "svc")
	if exists {
		t.Error("state should be cleared")
	}
}

func TestEventTypeString(t *testing.T) {
	tests := []struct {
		t    EventType
		want string
	}{
		{EventServiceCrashed, "crashed"},
		{EventServiceRecovered, "recovered"},
		{EventServiceStarted, "started"},
		{EventServiceStopped, "stopped"},
	}

	for _, tt := range tests {
		if got := tt.t.String(); got != tt.want {
			t.Errorf("EventType(%d).String() = %q, want %q", tt.t, got, tt.want)
		}
	}
}

func TestMonitorEventsChannel(t *testing.T) {
	m := NewMonitor(time.Second)

	// Trigger an event
	m.UpdateService("proj", "svc", true, 0)

	// Read from channel
	select {
	case event := <-m.Events():
		if event.Type != EventServiceStarted {
			t.Errorf("expected started, got %v", event.Type)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("expected event on channel")
	}

	m.Close()
}
