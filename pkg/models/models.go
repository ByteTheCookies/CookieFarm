package models

// Service represents a single vulnerable service as defined in the configuration.
type Service struct {
	Port uint16 `json:"port" yaml:"port"` // Port where the service is exposed
	Name string `json:"name" yaml:"name"` // Name identifier of the service
}

// ConfigServer holds configuration data required by the server to submit and validate flags.
type ConfigServer struct {
	SubmitFlagCheckerTime uint64 `json:"submit_flag_checker_time" yaml:"submit_flag_checker_time"` // Time interval (s) to check and submit flags
	MaxFlagBatchSize      uint   `json:"max_flag_batch_size" yaml:"max_flag_batch_size"`           // Max number of flags to send in a single batch
	URLFlagChecker        string `json:"url_flag_checker" yaml:"url_flag_checker"`                 // Address of the flagchecker server
	TeamToken             string `json:"team_token" yaml:"team_token"`                             // Authentication token for team identity
	Protocol              string `json:"protocol" yaml:"protocol"`                                 // Protocol used to communicate with the flagchecker server
	TickTime              int    `json:"tick_time" yaml:"tick_time"`                               // Duration of one game tick in seconds
	StartTime             string `json:"start_time" yaml:"start_time"`                             // CTF competition start time (HH:MM:SS format)
	EndTime               string `json:"end_time" yaml:"end_time"`                                 // CTF competition end time (HH:MM:SS format)
	FlagTTL               uint64 `json:"flag_ttl" yaml:"flag_ttl"`                                 // Time-to-live for flags in ticks
}

// ConfigClient contains all client-side configuration options.
type ConfigClient struct {
	RegexFlag     string    `json:"regex_flag" yaml:"regex_flag"`           // Regex used to identify flags in output
	FormatIPTeams string    `json:"format_ip_teams" yaml:"format_ip_teams"` // Format string for generating team IPs
	MyTeamID      string    `json:"my_team_id" yaml:"my_team_id"`           // ID of the current team
	Services      []Service `json:"services" yaml:"services"`               // List of services to exploit
	RangeIPTeams  uint8     `json:"range_ip_teams" yaml:"range_ip_teams"`   // Number of teams / IP range
	NOPTeam       int       `json:"nop_team" yaml:"nop_team"`               // The id of the nop team in the ctf
	URLFlagIds    string    `json:"url_flag_ids" yaml:"url_flag_ids"`       // URLFlagIds is the where the flagsId server is running
}

// ConfigShared aggregates both server and client configuration,
// and includes a flag indicating whether the configuration is initialized.
type ConfigShared struct {
	Configured   bool         `json:"configured" yaml:"configured"` // True if configuration has been loaded and validated
	ConfigServer ConfigServer `json:"server" yaml:"server"`         // Server-specific configuration
	ConfigClient ConfigClient `json:"client" yaml:"client"`         // Client-specific configuration
}

const (
	StatusUnsubmitted = "UNSUBMITTED" // Status for unsubmitted flags
	StatusAccepted    = "ACCEPTED"    // Status for accepted flags
	StatusDenied      = "DENIED"      // Status for denied flags
	StatusError       = "ERROR"       // Status for error flags

	VERSION = "v1.2.0"
)

// ClientData represents a single flag captured during a CTF round.
// It includes metadata about the submission and the service context.
type ClientData struct {
	SubmitTime   uint64 `json:"submit_time"`   // UNIX timestamp when the flag was submitted
	ResponseTime uint64 `json:"response_time"` // UNIX timestamp when a response was received
	FlagCode     string `json:"flag_code"`     // Actual flag string
	ServiceName  string `json:"service_name"`  // Human-readable name of the service
	Status       string `json:"status"`        // Status of the submission (e.g., "unsubmitted", "accepted", "denied")
	Username     string `json:"username"`      // Username of client
	ExploitName  string `json:"exploit_name"`  // Name of the exploit
	Msg          string `json:"msg"`           // Message from the flag checker service
	PortService  uint16 `json:"port_service"`  // Port of the vulnerable service
	TeamID       uint16 `json:"team_id"`       // ID of the team the flag was captured from
}

// SubmitFlagsRequest the struct for the requests from the client to server
type SubmitFlagsRequest struct {
	Flags []ClientData `json:"flags"` // Flags to submit
}

// SubmitFlagRequest the struct for the requests from the client to server
type SubmitFlagRequest struct {
	Flag ClientData `json:"flag"` // Flag to submit
}
