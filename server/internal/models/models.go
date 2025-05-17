// Package models defines the core data structures used by both
// the CookieFarm client and server for handling flags, configuration,
// and runtime arguments.
package models

const (
	StatusUnsubmitted = "UNSUBMITTED" // Status for unsubmitted flags
	StatusAccepted    = "ACCEPTED"    // Status for accepted flags
	StatusDenied      = "DENIED"      // Status for denied flags
	StatusError       = "ERROR"       // Status for error flags
)

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

// ResponseProtocol represents a response from the flag checker service.
type ResponseProtocol struct {
	Flag   string `json:"flag"`   // Flag string received from the flag checker service
	Msg    string `json:"msg"`    // Message from the flag checker service
	Status string `json:"status"` // Status of the response (e.g., "success", "error")
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

// SigninRequest from the client to the server
type SigninRequest struct {
	Password string `json:"password"` // Password for authentication
}

// Pagination structure for manage data in the view
type Pagination struct {
	Limit    int   // Maximum number of items per page
	Pages    int   // Total number of pages
	Current  int   // Current page number (offset / limit)
	HasPrev  bool  // Indicates if there is a previous page
	HasNext  bool  // Indicates if there is a next page
	PageList []int // List of page numbers to display in the pagination
}

// ViewParamsDashboard represents the parameters for the dashboard view
type ViewParamsDashboard struct {
	Limit int `json:"limit"` // Maximum number of items per page
}

// ViewParamsPagination represents the parameters for the pagination view
type ViewParamsPagination struct {
	Pagination Pagination // Pagination parameters
}

// ViewParamsFlags represents the parameters for the flags view
type ViewParamsFlags struct {
	Flags []Flag `json:"flags"` // List of flags to display
}

// ResponseFlags represents the response for the flags api
type ResponseFlags struct {
	Nflags int    `json:"n_flags"`
	Flags  []Flag `json:"flags"`
}

// ResponseSuccess represents the response for the success api
type ResponseSuccess struct {
	Message string `json:"message"` // Message for the success response
}

// ResponseError represents the response for the error api
type ResponseError struct {
	Error   string `json:"error"`   // Error message for the error response
	Details string `json:"details"` // Details for the error response
}

// SubmitFlagRequest the struct for the requests from the client to server
type SubmitFlagRequest struct {
	Flag Flag `json:"flag"` // Flag to submit
}

// SubmitFlagsRequest the struct for the requests from the client to server
type SubmitFlagsRequest struct {
	Flags []Flag `json:"flags"` // Flags to submit
}
