// Package compose provides a client for the process-compose REST API.
package compose

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"
)

// Client communicates with process-compose via Unix socket.
type Client struct {
	socketPath string
	httpClient *http.Client
	connected  bool
}

// NewClient creates a new process-compose client.
func NewClient(socketPath string) *Client {
	return &Client{
		socketPath: socketPath,
		httpClient: &http.Client{
			Transport: &http.Transport{
				DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
					return net.Dial("unix", socketPath)
				},
			},
			Timeout: 5 * time.Second,
		},
	}
}

// Connect attempts to connect to the process-compose socket.
func (c *Client) Connect() error {
	conn, err := net.Dial("unix", c.socketPath)
	if err != nil {
		c.connected = false
		return err
	}
	conn.Close()
	c.connected = true
	return nil
}

// IsConnected returns true if connected to process-compose.
func (c *Client) IsConnected() bool {
	return c.connected
}

// GetStatus fetches the current process status.
func (c *Client) GetStatus() (*ProjectStatus, error) {
	resp, err := c.httpClient.Get("http://unix/processes")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var apiResp processesResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}

	return &ProjectStatus{Processes: apiResp.Data}, nil
}

// StartProcess starts a specific process.
func (c *Client) StartProcess(name string) error {
	url := fmt.Sprintf("http://unix/process/%s/start", name)
	resp, err := c.httpClient.Post(url, "application/json", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to start process: %d", resp.StatusCode)
	}
	return nil
}

// StopProcess stops a specific process.
func (c *Client) StopProcess(name string) error {
	url := fmt.Sprintf("http://unix/process/%s/stop", name)
	resp, err := c.httpClient.Post(url, "application/json", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to stop process: %d", resp.StatusCode)
	}
	return nil
}

// RestartProcess restarts a specific process.
func (c *Client) RestartProcess(name string) error {
	url := fmt.Sprintf("http://unix/process/%s/restart", name)
	resp, err := c.httpClient.Post(url, "application/json", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to restart process: %d", resp.StatusCode)
	}
	return nil
}

// ShutdownProject stops all processes and shuts down.
func (c *Client) ShutdownProject() error {
	resp, err := c.httpClient.Post("http://unix/project/stop", "application/json", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to shutdown: %d", resp.StatusCode)
	}
	return nil
}

// GetLogs fetches recent logs for a process.
// endOffset is offset from end (0 = most recent), limit is max lines to return.
func (c *Client) GetLogs(processName string, endOffset, limit int) ([]string, error) {
	url := fmt.Sprintf("http://unix/process/logs/%s/%d/%d", processName, endOffset, limit)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get logs: %d", resp.StatusCode)
	}

	var logsResp logsResponse
	if err := json.NewDecoder(resp.Body).Decode(&logsResp); err != nil {
		return nil, err
	}

	return logsResp.Logs, nil
}
