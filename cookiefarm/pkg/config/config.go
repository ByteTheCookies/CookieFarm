package sharedconfig

type Shared struct {
	Services      map[string]uint16 `json:"services" yaml:"services"`               // List of services to exploit
	RegexFlag     string            `json:"regex_flag" yaml:"regex_flag"`           // Regex used to identify flags in output
	FormatIPTeams string            `json:"format_ip_teams" yaml:"format_ip_teams"` // Format string for generating team IPs
	MyTeamID      int               `json:"my_team_id" yaml:"my_team_id"`           // ID of the current team
	URLFlagIds    string            `json:"url_flag_ids" yaml:"url_flag_ids"`       // URLFlagIds is the where the flagsId server is running
	NOPTeam       int               `json:"nop_team" yaml:"nop_team"`               // The id of the nop team in the ctf
	RangeIPTeams  uint8             `json:"range_ip_teams" yaml:"range_ip_teams"`   // Number of teams / IP range
	Configured    bool              `json:"configured" yaml:"configured"`           // True if configuration has been loaded and validated
	FlagIdsFormat string            `json:"flagids_format" yaml:"flagids_format"`   // Format string for generating flag IDs
}

func (cfg *Shared) Print() {
	for service, port := range cfg.Services {
		println("Service:", service, "Port:", port) //nolint:forbidigo
	}
	println("RegexFlag:", cfg.RegexFlag)         //nolint:forbidigo
	println("FormatIPTeams:", cfg.FormatIPTeams) //nolint:forbidigo
	println("MyTeamID:", cfg.MyTeamID)           //nolint:forbidigo
	println("URLFlagIds:", cfg.URLFlagIds)       //nolint:forbidigo
	println("NOPTeam:", cfg.NOPTeam)             //nolint:forbidigo
	println("RangeIPTeams:", cfg.RangeIPTeams)   //nolint:forbidigo
	println("Configured:", cfg.Configured)       //nolint:forbidigo
	println("FlagIdsFormat:", cfg.FlagIdsFormat) //nolint:forbidigo
}

func (cfg *Shared) Set(newcfg Shared) {
	*cfg = newcfg
}
