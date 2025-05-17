// Package models defines the core data structures used by both
// the CookieFarm client and server for handling flags, configuration,
// and runtime arguments.
package models

// Flag represents a single flag captured during a CTF round.
// It includes metadata about the submission and the service context.
type Flag struct {
	SubmitTime   uint64 `json:"submit_time"`   // UNIX timestamp when the flag was submitted
	ResponseTime uint64 `json:"response_time"` // UNIX timestamp when a response was received
	PortService  uint16 `json:"port_service"`  // Port of the vulnerable service
	TeamID       uint16 `json:"team_id"`       // ID of the team the flag was captured from
	FlagCode     string `json:"flag_code"`     // Actual flag string
	ServiceName  string `json:"service_name"`  // Human-readable name of the service
	Status       string `json:"status"`        // Status of the submission (e.g., "unsubmitted", "accepted", "denied")
	Msg          string `json:"msg"`           // Message from the flag checker service
}

// Service represents a single vulnerable service as defined in the configuration.
type Service struct {
	Port uint16 `json:"port" yaml:"port"` // Port where the service is exposed
	Name string `json:"name" yaml:"name"` // Name identifier of the service
}

// ConfigServer holds configuration data required by the server to submit and validate flags.
type ConfigServer struct {
	SubmitFlagCheckerTime uint64 `json:"submit_flag_checker_time" yaml:"submit_flag_checker_time"` // Time interval (s) to check and submit flags
	HostFlagchecker       string `json:"host_flagchecker" yaml:"host_flagchecker"`                 // Address of the flagchecker server
	TeamToken             string `json:"team_token" yaml:"team_token"`                             // Authentication token for team identity
	MaxFlagBatchSize      uint   `json:"max_flag_batch_size" yaml:"max_flag_batch_size"`           // Max number of flags to send in a single batch
	Protocol              string `json:"protocol" yaml:"protocol"`                                 // Protocol used to communicate with the flagchecker server
}

// ConfigClient contains all client-side configuration options.
type ConfigClient struct {
	Services      []Service `json:"services" yaml:"services"`               // List of services to exploit
	RangeIPTeams  uint8     `json:"range_ip_teams" yaml:"range_ip_teams"`   // Number of teams / IP range
	FormatIPTeams string    `json:"format_ip_teams" yaml:"format_ip_teams"` // Format string for generating team IPs
	MyTeamIP      string    `json:"my_team_ip" yaml:"my_team_ip"`           // IP address of the current team
	RegexFlag     string    `json:"regex_flag" yaml:"regex_flag"`           // Regex used to identify flags in output
}

// Config aggregates both server and client configuration,
// and includes a flag indicating whether the configuration is initialized.
type Config struct {
	Configured   bool         `json:"configured" yaml:"configured"` // True if configuration has been loaded and validated
	ConfigServer ConfigServer `json:"server" yaml:"server"`         // Server-specific configuration
	ConfigClient ConfigClient `json:"client" yaml:"client"`         // Client-specific configuration
}

// Args represents the command-line arguments or configuration
// values passed at runtime to control the exploit manager behavior.
type Args struct {
	ExploitPath *string `json:"exploit_path"` // Path to the exploit to run
	Password    *string `json:"password"`     // Authentication password
	TickTime    *int    `json:"tick_time"`    // Optional custom tick time
	Debug       *bool   `json:"debug"`        // Enable debug mode if true
	Detach      *bool   `json:"detach"`       // Run in background if true
	ThreadCount *int    `json:"thread_count"` // Optional number of concurrent threads (coroutine) to use
}

// ParsedFlagOutput represents the output of a parsed flag returned
// by an exploit run in the exploit_manager, ready to be submitted.
type ParsedFlagOutput struct {
	TeamID      uint16 `json:"team_id"`      // ID of the team the flag was extracted from
	PortService uint16 `json:"port_service"` // Port of the service that produced the flag
	FlagCode    string `json:"flag_code"`    // The actual flag string
	Status      string `json:"status"`       // Status of the flag submission (eg "success", "failed", "error", "fatal")
	Message     string `json:"message"`      // Additional message or error information
}

type EventWS struct {
	Type    string `json:"type"`
	Payload []byte `json:"payload"`
}

type EventWSFlag struct {
	Type    string `json:"type"`
	Payload Flag   `json:"payload"`
}
