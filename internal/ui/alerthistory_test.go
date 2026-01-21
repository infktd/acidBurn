package ui

import (
	"fmt"
	"testing"
	"time"
)

func TestAlertHistoryAdd(t *testing.T) {
	h := NewAlertHistory(10)

	h.Add(Alert{
		Type:      AlertServiceCrashed,
		Project:   "my-project",
		Service:   "postgres",
		Message:   "test",
		Timestamp: time.Now(),
	})

	if h.Len() != 1 {
		t.Errorf("expected 1 alert, got %d", h.Len())
	}
}

func TestAlertHistoryCapacity(t *testing.T) {
	h := NewAlertHistory(3)

	for i := 0; i < 5; i++ {
		h.Add(Alert{
			Type:      AlertInfo,
			Message:   fmt.Sprintf("msg %d", i),
			Timestamp: time.Now(),
		})
	}

	if h.Len() != 3 {
		t.Errorf("expected capacity 3, got %d", h.Len())
	}

	// Should have messages 2, 3, 4 (oldest dropped)
	all := h.All()
	if all[0].Message != "msg 2" {
		t.Errorf("expected 'msg 2', got %q", all[0].Message)
	}
	if all[2].Message != "msg 4" {
		t.Errorf("expected 'msg 4', got %q", all[2].Message)
	}
}

func TestAlertHistoryRecent(t *testing.T) {
	h := NewAlertHistory(10)

	for i := 0; i < 5; i++ {
		h.Add(Alert{
			Type:      AlertInfo,
			Message:   fmt.Sprintf("msg %d", i),
			Timestamp: time.Now(),
		})
	}

	recent := h.Recent(2)
	if len(recent) != 2 {
		t.Errorf("expected 2 recent, got %d", len(recent))
	}
	// Newest first
	if recent[0].Message != "msg 4" {
		t.Errorf("expected 'msg 4' first, got %q", recent[0].Message)
	}
	if recent[1].Message != "msg 3" {
		t.Errorf("expected 'msg 3' second, got %q", recent[1].Message)
	}
}

func TestAlertHistoryClear(t *testing.T) {
	h := NewAlertHistory(10)
	h.Add(Alert{Type: AlertInfo, Message: "test", Timestamp: time.Now()})
	h.Add(Alert{Type: AlertInfo, Message: "test", Timestamp: time.Now()})

	if h.Len() != 2 {
		t.Error("should have 2 alerts")
	}

	h.Clear()

	if h.Len() != 0 {
		t.Errorf("should be empty after clear, got %d", h.Len())
	}
}

func TestAlertTypeString(t *testing.T) {
	tests := []struct {
		t    AlertType
		want string
	}{
		{AlertServiceCrashed, "crashed"},
		{AlertServiceRecovered, "recovered"},
		{AlertProjectStarted, "started"},
		{AlertProjectStopped, "stopped"},
		{AlertCritical, "critical"},
		{AlertInfo, "info"},
	}

	for _, tt := range tests {
		if got := tt.t.String(); got != tt.want {
			t.Errorf("AlertType(%d).String() = %q, want %q", tt.t, got, tt.want)
		}
	}
}

func TestAlertHistoryConvenienceMethods(t *testing.T) {
	h := NewAlertHistory(10)

	h.AddServiceCrashed("proj", "svc", 1)
	h.AddServiceRecovered("proj", "svc")

	if h.Len() != 2 {
		t.Errorf("expected 2, got %d", h.Len())
	}

	all := h.All()
	if all[0].Type != AlertServiceCrashed {
		t.Error("first should be crashed")
	}
	if all[1].Type != AlertServiceRecovered {
		t.Error("second should be recovered")
	}
}
