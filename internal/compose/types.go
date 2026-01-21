package compose

// ProcessStatus represents the status of a process in process-compose.
type ProcessStatus struct {
	Name       string  `json:"name"`
	Namespace  string  `json:"namespace"`
	Status     string  `json:"status"`
	IsRunning  bool    `json:"is_running"`
	Pid        int     `json:"pid"`
	ExitCode   int     `json:"exit_code"`
	SystemTime string  `json:"system_time"`
	Restarts   int     `json:"restarts"`
	Mem        int64   `json:"mem"`
	CPU        float64 `json:"cpu"`
}

// processesResponse is the API response wrapper for /processes.
type processesResponse struct {
	Data []ProcessStatus `json:"data"`
}

// ProjectStatus represents the overall project status.
type ProjectStatus struct {
	Processes []ProcessStatus `json:"processes"`
}

// logsResponse is the API response wrapper for /process/logs.
type logsResponse struct {
	Logs []string `json:"logs"`
}
