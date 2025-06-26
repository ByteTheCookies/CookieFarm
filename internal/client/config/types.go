package config

// ArgsAttack represents the command-line arguments or configuration
// values passed at runtime to control the exploit manager behavior.
type ArgsAttack struct {
	ServicePort uint16 `json:"port"`         // Service Port
	TickTime    int    `json:"tick_time"`    // Optional custom tick time
	ThreadCount int    `json:"thread_count"` // Optional number of concurrent threads (coroutine) to use
	Detach      bool   `json:"detach"`       // Run in background if true
	ExploitPath string `json:"exploit_path"` // Path to the exploit to run
}

type Exploit struct {
	Name string `json:"name"` // Name of the exploit
	PID  int    `json:"pid"`  // Process ID of the exploit
}

type ConfigLocal struct {
	Host     string    `json:"host"`     // Host address of the server
	Port     uint16    `json:"port"`     // Port of the server
	HTTPS    bool      `json:"protocol"` // Protocol used to connect to the server (e.g., http, https)
	Username string    `json:"username"` // Username of the client
	Exploits []Exploit `json:"exploits"` // List of exploits available in the client
}
