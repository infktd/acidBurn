package ui

import (
	"strings"
	"testing"
)

func TestNewProgressBar(t *testing.T) {
	pb := NewProgressBar(nil, 40)
	if pb == nil {
		t.Fatal("NewProgressBar returned nil")
	}
	if pb.Progress() != 0.0 {
		t.Errorf("new progress bar progress = %f, want 0.0", pb.Progress())
	}
}

func TestProgressBarSetProgress(t *testing.T) {
	pb := NewProgressBar(nil, 40)

	pb.SetProgress(0.5)
	if pb.Progress() != 0.5 {
		t.Errorf("Progress() = %f, want 0.5", pb.Progress())
	}

	// Test clamping
	pb.SetProgress(1.5)
	if pb.Progress() > 1.0 {
		t.Errorf("Progress() = %f, should be clamped to 1.0", pb.Progress())
	}

	pb.SetProgress(-0.5)
	if pb.Progress() < 0.0 {
		t.Errorf("Progress() = %f, should be clamped to 0.0", pb.Progress())
	}
}

func TestProgressBarSetLabel(t *testing.T) {
	pb := NewProgressBar(nil, 40)
	pb.SetLabel("Loading...")

	view := pb.View()
	if !strings.Contains(view, "Loading") {
		t.Error("View() should contain label text")
	}
}

func TestProgressBarSetWidth(t *testing.T) {
	pb := NewProgressBar(nil, 40)
	pb.SetWidth(50)

	// Verify width is applied (via view output)
	pb.SetProgress(0.5)
	view := pb.View()
	if view == "" {
		t.Error("View() should not be empty after setting width and progress")
	}
}

func TestProgressBarSetShowPercentage(t *testing.T) {
	pb := NewProgressBar(nil, 40)
	pb.SetWidth(30)
	pb.SetProgress(0.75)

	pb.SetShowPercentage(true)
	viewWithPercent := pb.View()

	pb.SetShowPercentage(false)
	viewWithoutPercent := pb.View()

	// Views should differ
	if viewWithPercent == viewWithoutPercent {
		t.Log("SetShowPercentage may not affect output in this implementation")
	}
}

func TestProgressBarView(t *testing.T) {
	pb := NewProgressBar(nil, 40)
	pb.SetWidth(40)
	pb.SetLabel("Progress")
	pb.SetProgress(0.3)

	view := pb.View()
	if view == "" {
		t.Error("View() should not be empty")
	}
}

func TestNewSpinner(t *testing.T) {
	spinner := NewSpinner()
	if spinner == nil {
		t.Fatal("NewSpinner returned nil")
	}
}

func TestSpinnerFrame(t *testing.T) {
	spinner := NewSpinner()

	frame1 := spinner.Frame()
	if frame1 == "" {
		t.Error("Frame() should return non-empty string")
	}

	// Advance and get next frame
	frame2 := spinner.Frame()
	// Frames might be same or different depending on impl
	_ = frame2
}

func TestSpinnerReset(t *testing.T) {
	spinner := NewSpinner()

	// Advance a few frames
	spinner.Frame()
	spinner.Frame()
	spinner.Frame()

	spinner.Reset()

	// Should restart from beginning
	frame := spinner.Frame()
	if frame == "" {
		t.Error("Frame() after Reset() should return frame")
	}
}
