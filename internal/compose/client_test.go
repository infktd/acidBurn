package compose

import (
	"net"
	"net/http"
	"os"
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

func TestClientConnect(t *testing.T) {
	t.Run("successful connection", func(t *testing.T) {
		// Create a temporary Unix socket with a short path to avoid macOS limits
		socketPath := "/tmp/devdash-test-connect.sock"

		// Clean up any existing socket
		_ = os.Remove(socketPath)
		defer func() {
			// Ignore errors during cleanup
			_ = os.Remove(socketPath)
		}()

		// Start a listener
		listener, err := net.Listen("unix", socketPath)
		if err != nil {
			t.Fatalf("failed to create listener: %v", err)
		}
		defer listener.Close()

		client := NewClient(socketPath)
		err = client.Connect()
		if err != nil {
			t.Fatalf("Connect() error: %v", err)
		}
		if !client.IsConnected() {
			t.Error("IsConnected() should return true after successful connection")
		}
	})

	t.Run("failed connection", func(t *testing.T) {
		client := NewClient("/nonexistent/socket.sock")
		err := client.Connect()
		if err == nil {
			t.Fatal("Connect() should return error for nonexistent socket")
		}
		if client.IsConnected() {
			t.Error("IsConnected() should return false after failed connection")
		}
	})
}

func TestClientGetStatus(t *testing.T) {
	socketPath := "/tmp/devdash-test-status.sock"
	_ = os.Remove(socketPath)
	defer os.Remove(socketPath)

	// Create mock HTTP server over Unix socket
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		t.Fatalf("failed to create listener: %v", err)
	}
	defer listener.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("/processes", func(w http.ResponseWriter, r *http.Request) {
		response := `{
			"data": [
				{
					"name": "test-service",
					"status": "Running",
					"pid": 1234,
					"is_ready": "true"
				}
			]
		}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	})

	server := &http.Server{Handler: mux}
	go server.Serve(listener)
	defer server.Close()

	client := NewClient(socketPath)
	status, err := client.GetStatus()
	if err != nil {
		t.Fatalf("GetStatus() error: %v", err)
	}
	if status == nil {
		t.Fatal("GetStatus() returned nil status")
	}
	if len(status.Processes) != 1 {
		t.Errorf("expected 1 process, got %d", len(status.Processes))
	}
	if status.Processes[0].Name != "test-service" {
		t.Errorf("process name = %q, want %q", status.Processes[0].Name, "test-service")
	}
}

func TestClientStartProcess(t *testing.T) {
	socketPath := "/tmp/devdash-test-start.sock"
	_ = os.Remove(socketPath)
	defer os.Remove(socketPath)

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		t.Fatalf("failed to create listener: %v", err)
	}
	defer listener.Close()

	called := false
	mux := http.NewServeMux()
	mux.HandleFunc("/process/", func(w http.ResponseWriter, r *http.Request) {
		called = true
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	})

	server := &http.Server{Handler: mux}
	go server.Serve(listener)
	defer server.Close()

	client := NewClient(socketPath)
	err = client.StartProcess("test-service")
	if err != nil {
		t.Fatalf("StartProcess() error: %v", err)
	}
	if !called {
		t.Error("StartProcess() did not call HTTP endpoint")
	}
}

func TestClientStopProcess(t *testing.T) {
	socketPath := "/tmp/devdash-test-stop.sock"
	_ = os.Remove(socketPath)
	defer os.Remove(socketPath)

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		t.Fatalf("failed to create listener: %v", err)
	}
	defer listener.Close()

	called := false
	mux := http.NewServeMux()
	mux.HandleFunc("/process/", func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	server := &http.Server{Handler: mux}
	go server.Serve(listener)
	defer server.Close()

	client := NewClient(socketPath)
	err = client.StopProcess("test-service")
	if err != nil {
		t.Fatalf("StopProcess() error: %v", err)
	}
	if !called {
		t.Error("StopProcess() did not call HTTP endpoint")
	}
}

func TestClientRestartProcess(t *testing.T) {
	socketPath := "/tmp/devdash-test-restart.sock"
	_ = os.Remove(socketPath)
	defer os.Remove(socketPath)

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		t.Fatalf("failed to create listener: %v", err)
	}
	defer listener.Close()

	called := false
	mux := http.NewServeMux()
	mux.HandleFunc("/process/", func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	server := &http.Server{Handler: mux}
	go server.Serve(listener)
	defer server.Close()

	client := NewClient(socketPath)
	err = client.RestartProcess("test-service")
	if err != nil {
		t.Fatalf("RestartProcess() error: %v", err)
	}
	if !called {
		t.Error("RestartProcess() did not call HTTP endpoint")
	}
}

func TestClientShutdownProject(t *testing.T) {
	socketPath := "/tmp/devdash-test-shutdown.sock"
	_ = os.Remove(socketPath)
	defer os.Remove(socketPath)

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		t.Fatalf("failed to create listener: %v", err)
	}
	defer listener.Close()

	called := false
	mux := http.NewServeMux()
	mux.HandleFunc("/project/stop", func(w http.ResponseWriter, r *http.Request) {
		called = true
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	})

	server := &http.Server{Handler: mux}
	go server.Serve(listener)
	defer server.Close()

	client := NewClient(socketPath)
	err = client.ShutdownProject()
	if err != nil {
		t.Fatalf("ShutdownProject() error: %v", err)
	}
	if !called {
		t.Error("ShutdownProject() did not call HTTP endpoint")
	}
}

func TestClientGetLogs(t *testing.T) {
	socketPath := "/tmp/devdash-test-logs.sock"
	_ = os.Remove(socketPath)
	defer os.Remove(socketPath)

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		t.Fatalf("failed to create listener: %v", err)
	}
	defer listener.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("/process/logs/test-service/0/100", func(w http.ResponseWriter, r *http.Request) {
		response := `{
			"logs": ["log line 1", "log line 2"]
		}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	})

	server := &http.Server{Handler: mux}
	go server.Serve(listener)
	defer server.Close()

	client := NewClient(socketPath)
	logs, err := client.GetLogs("test-service", 0, 100)
	if err != nil {
		t.Fatalf("GetLogs() error: %v", err)
	}
	if len(logs) != 2 {
		t.Errorf("expected 2 log lines, got %d", len(logs))
	}
	if logs[0] != "log line 1" {
		t.Errorf("logs[0] = %q, want %q", logs[0], "log line 1")
	}
	if logs[1] != "log line 2" {
		t.Errorf("logs[1] = %q, want %q", logs[1], "log line 2")
	}
}
