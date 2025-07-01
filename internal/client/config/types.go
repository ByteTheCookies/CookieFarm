package config

import (
	"sync"

	"github.com/ByteTheCookies/CookieFarm/pkg/models"
)

// ArgsAttack represents the command-line arguments or configuration
// values passed at runtime to control the exploit manager behavior.
type ArgsAttack struct {
	ExploitPath string `json:"exploit_path"` // Path to the exploit to run
	TickTime    int    `json:"tick_time"`    // Optional custom tick time
	ThreadCount int    `json:"thread_count"` // Optional number of concurrent threads (coroutine) to use
	ServicePort uint16 `json:"port"`         // Service Port
	Detach      bool   `json:"detach"`       // Run in background if true
}

type Exploit struct {
	Name string `json:"name"` // Name of the exploit
	PID  int    `json:"pid"`  // Process ID of the exploit
}

type ConfigLocal struct {
	Exploits []Exploit `json:"exploits"` // List of exploits available in the client
	Host     string    `json:"host"`     // Host address of the server
	Username string    `json:"username"` // Username of the client
	Port     uint16    `json:"port"`     // Port of the server
	HTTPS    bool      `json:"protocol"` // Protocol used to connect to the server (e.g., http, https)
}

// ConfigManager manages all configuration state in a thread-safe manner
type ConfigManager struct {
	mu                 sync.RWMutex
	argsAttackInstance ArgsAttack
	localConfig        ConfigLocal
	sharedConfig       models.ConfigShared
	token              string
	pid                int
	useTUI             bool
	useBanner          bool
}
