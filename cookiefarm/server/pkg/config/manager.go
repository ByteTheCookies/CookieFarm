package config

import (
	"logger"
	"os"
	"sharedconfig"
	"sync"

	"gopkg.in/yaml.v3"
)

var (
	instance *ConfigManager
	once     sync.Once
)

func GetInstance() *ConfigManager {
	once.Do(func() {
		instance = &ConfigManager{}

		instance.state.Store(&FullConfig{
			Server: Config{},
			Shared: sharedconfig.Shared{
				Services: make(map[string]uint16),
			},
		})
	})
	return instance
}

func (cm *ConfigManager) Get() *FullConfig {
	return cm.state.Load().(*FullConfig)
}

func (cm *ConfigManager) GetURLFlagChecker() string {
	return cm.Get().Server.URLFlagChecker
}

func (cm *ConfigManager) SetURLFlagChecker(v string) {
	cm.Get().Server.URLFlagChecker = v
}

func (cm *ConfigManager) GetTeamToken() string {
	return cm.Get().Server.TeamToken
}

func (cm *ConfigManager) SetTeamToken(v string) {
	cm.Get().Server.TeamToken = v
}

func (cm *ConfigManager) GetSubmitFlagCheckerTime() uint {
	return cm.Get().Server.SubmitFlagCheckerTime
}

func (cm *ConfigManager) SetSubmitFlagCheckerTime(v uint) {
	cm.Get().Server.SubmitFlagCheckerTime = v
}

func (cm *ConfigManager) GetMaxFlagBatchSize() uint {
	return cm.Get().Server.MaxFlagBatchSize
}

func (cm *ConfigManager) SetMaxFlagBatchSize(v uint) {
	cm.Get().Server.MaxFlagBatchSize = v
}

func (cm *ConfigManager) GetProtocol() string {
	return cm.Get().Server.Protocol
}

func (cm *ConfigManager) SetProtocol(v string) {
	cm.Get().Server.Protocol = v
}

func (cm *ConfigManager) GetTickTime() uint {
	return cm.Get().Server.TickTime
}

func (cm *ConfigManager) SetTickTime(v uint) {
	cm.Get().Server.TickTime = v
}

func (cm *ConfigManager) GetFlagTTL() uint64 {
	return cm.Get().Server.FlagTTL
}

func (cm *ConfigManager) SetFlagTTL(v uint64) {
	cm.Get().Server.FlagTTL = v
}

func (cm *ConfigManager) GetStartTime() string {
	return cm.Get().Server.StartTime
}

func (cm *ConfigManager) SetStartTime(v string) {
	cm.Get().Server.StartTime = v
}

func (cm *ConfigManager) GetEndTime() string {
	return cm.Get().Server.EndTime
}

func (cm *ConfigManager) SetEndTime(v string) {
	cm.Get().Server.EndTime = v
}

func (cm *ConfigManager) SetConfigured(value bool) {
	cm.Get().Configured = value
}

func (cm *ConfigManager) GetConfigured() bool {
	return cm.Get().Configured
}

func (cm *ConfigManager) GetToken() string {
	return cm.token
}

func (cm *ConfigManager) SetToken(token string) {
	cm.token = token
}

func (cm *ConfigManager) GetShared() sharedconfig.Shared {
	return cm.Get().Shared
}

func (cm *ConfigManager) SetShared(v sharedconfig.Shared) {
	cm.Get().Shared = v
}

func (cm *ConfigManager) SetConfig(c Config) {
	cm.Get().Server = c
}

func (cm *ConfigManager) GetConfig() Config {
	return cm.Get().Server
}

func (cm *ConfigManager) GetFullConfig() FullConfig {
	return *cm.Get()
}

func (cm *ConfigManager) SetFullConfig(f FullConfig) {
	cm.Get().Server = f.Server
	cm.Get().Shared = f.Shared
}

// Write the config file
func (cm *ConfigManager) Write() {
	fullConfig := cm.Get()
	marshalConfig, err := yaml.Marshal(fullConfig)
	if err != nil {
		logger.Log.Err(err).Msg("error during the unmarshal of the config")
	}
	if err := os.WriteFile(ConfigFile, marshalConfig, 0o644); err != nil {
		logger.Log.Err(err).Msg("error writing config file")
	}
}

func (cm *ConfigManager) MapServiceToPort(service string) uint16 {
	return cm.Get().Shared.Services[service]
}

func (cm *ConfigManager) MapPortToService(port uint16) string {
	for k, v := range cm.Get().Shared.Services {
		if v == port {
			return k
		}
	}
	return ""
}
