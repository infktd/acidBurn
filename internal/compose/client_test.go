package compose

import (
	"testing"
)

func TestNewClient(t *testing.T) {
	client := NewClient("/tmp/test.sock")
	if client == nil {
		t.Fatal("NewClient returned nil")
	}
	if client.socketPath != "/tmp/test.sock" {
		t.Errorf("Expected socket path '/tmp/test.sock', got %q", client.socketPath)
	}
}

func TestClientIsConnectedReturnsFalseWhenNotConnected(t *testing.T) {
	client := NewClient("/nonexistent/socket.sock")
	if client.IsConnected() {
		t.Fatal("IsConnected should return false for nonexistent socket")
	}
}
