package config

import (
	"sharedconfig"
	"sync/atomic"
	"system"
)

var DefaultPath, _ = system.ExpandTilde("~/.config/cookiefarm")

type InfraMetaData struct {
	FlagRegex     string `json:"regex_flag" yaml:"regex_flag"`
	FormatIPTeams string `json:"format_ip_teams" yaml:"format_ip_teams"`
	MyTeamID      int    `json:"my_team_id" yaml:"my_team_id"`
	URLFlagIds    string `json:"url_flag_ids" yaml:"url_flag_ids"`
	NOPTeam       int    `json:"nop_team" yaml:"nop_team"`
	RangeIPTeams  uint8  `json:"range_ip_teams" yaml:"range_ip_teams"`
}

type LocalConfig struct {
	Host     string `json:"host" yaml:"host"`
	Username string `json:"username" yaml:"username"`
	Port     uint16 `json:"port" yaml:"port"`
	HTTPS    bool   `json:"https" yaml:"https"`
}

type RuntimeConfig struct {
	Local  LocalConfig
	Shared sharedconfig.Shared
	Token  string
}

type ConfigManager struct {
	state atomic.Value // *RuntimeConfig
}

func (l *LocalConfig) Print() {
	println("Host:", l.Host)         //nolint:forbidigo
	println("Username:", l.Username) //nolint:forbidigo
	println("Port:", l.Port)         //nolint:forbidigo
	println("HTTPS:", l.HTTPS)       //nolint:forbidigo
}
