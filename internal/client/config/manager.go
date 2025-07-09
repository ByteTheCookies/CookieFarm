package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/ByteTheCookies/CookieFarm/pkg/logger"
	"github.com/ByteTheCookies/CookieFarm/pkg/models"
	"gopkg.in/yaml.v3"
)

var (
	instance *ConfigManager
	once     sync.Once
)

// GetInstance returns the singleton instance of ConfigManager
func GetInstance() *ConfigManager {
	once.Do(func() {
		instance = &ConfigManager{
			useBanner: true, // default value
			useTUI:    true, // default value
		}
	})
	return instance
}

// Token methods
func (cm *ConfigManager) GetToken() string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.token
}

func (cm *ConfigManager) SetToken(token string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.token = token
}

// ArgsAttack methods
func (cm *ConfigManager) GetArgsAttackInstance() ArgsAttack {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.argsAttackInstance
}

func (cm *ConfigManager) SetArgsAttackInstance(args ArgsAttack) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.argsAttackInstance = args
}

// LocalConfig methods
func (cm *ConfigManager) GetLocalConfig() ConfigLocal {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.localConfig
}

func (cm *ConfigManager) SetLocalConfig(config ConfigLocal) (string, error) {
	configPath, err := cm.UpdateLocalConfigToFile(config)

	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.localConfig = config

	return configPath, err
}

// UpdateLocalConfig the value empty are not setted except for https
// for the string empty is "" for number is 0
func (cm *ConfigManager) UpdateLocalConfig(host string, port uint16, username string, https bool, exploits []Exploit) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if host != "" {
		cm.localConfig.Host = host
	}
	if port != 0 {
		cm.localConfig.Port = port
	}
	if username != "" {
		cm.localConfig.Username = username
	}
	cm.localConfig.HTTPS = https
	if len(exploits) > 0 {
		cm.localConfig.Exploits = exploits
	}

	logger.Log.Debug().Msgf("%+v", cm.localConfig)
}

// SharedConfig methods
func (cm *ConfigManager) GetSharedConfig() models.ConfigShared {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.sharedConfig
}

func (cm *ConfigManager) SetSharedConfig(config models.ConfigShared) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.sharedConfig = config
}

// TUI methods
func (cm *ConfigManager) GetUseTUI() bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.useTUI
}

func (cm *ConfigManager) SetUseTUI(useTUI bool) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.useTUI = useTUI
	logger.SetTUI(useTUI)
}

// PID methods
func (cm *ConfigManager) GetPID() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.pid
}

func (cm *ConfigManager) SetPID(pid int) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.pid = pid
}

// Banner methods
func (cm *ConfigManager) GetUseBanner() bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.useBanner
}

func (cm *ConfigManager) SetUseBanner(useBanner bool) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.useBanner = useBanner
}

// LoadLocalConfigFromFile loads the local configuration from file
func (cm *ConfigManager) LoadLocalConfigFromFile() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if _, err := os.Stat(DefaultConfigPath); os.IsNotExist(err) {
		logger.Log.Warn().Msgf("Config directory does not exist at %s, creating it", DefaultConfigPath)
		err = os.MkdirAll(DefaultConfigPath, 0o755)
		if err != nil {
			return fmt.Errorf("error creating config directory: %w", err)
		}

		if _, err := os.Create(filepath.Join(DefaultConfigPath, "config.yml")); err != nil {
			return fmt.Errorf("error creating default config file: %w", err)
		}
	}

	configFileContent, err := os.ReadFile(filepath.Join(DefaultConfigPath, "config.yml"))
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("config file does not exist at %s", DefaultConfigPath)
		}
		return fmt.Errorf("error reading config file: %w", err)
	}

	err = yaml.Unmarshal(configFileContent, &cm.localConfig)
	if err != nil {
		return fmt.Errorf("error unmarshalling config: %w", err)
	}

	return nil
}

// WriteLocalConfigToFile writes the current local configuration to file
func (cm *ConfigManager) WriteLocalConfigToFile() error {
	cm.mu.RLock()
	config := cm.localConfig
	cm.mu.RUnlock()

	configFilePath := filepath.Join(DefaultConfigPath, "config.yml")
	configFileContent, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("error marshalling config: %w", err)
	}

	err = os.WriteFile(configFilePath, configFileContent, 0o644)
	if err != nil {
		return fmt.Errorf("error writing config file: %w", err)
	}

	return nil
}

// ResetLocalConfigToDefaults resets the local configuration to defaults
func (cm *ConfigManager) ResetLocalConfigToDefaults() (string, error) {
	err := cm.LoadLocalConfigFromFile()
	if err != nil {
		logger.Log.Warn().Err(err).Msg("Could not load existing config, proceeding with empty exploits")
	}
	cm.mu.Lock()
	defer cm.mu.Unlock()

	err = os.MkdirAll(DefaultConfigPath, 0o755)
	if err != nil {
		logger.Log.Error().Err(err).Msg("Error creating config directory")
		return "", err
	}

	configPath := filepath.Join(DefaultConfigPath, "config.yml")

	file, err := os.Create(configPath)
	if err != nil {
		logger.Log.Error().Err(err).Msg("Error opening configuration file")
		return "", err
	}
	defer file.Close()

	var tmpLocalconfig ConfigLocal
	err = yaml.Unmarshal(ConfigTemplate, &tmpLocalconfig)
	if err != nil {
		logger.Log.Error().Err(err).Msg("Error unmarshalling default configuration")
		return "", err
	}

	existingExploits := cm.localConfig.Exploits
	cm.localConfig = tmpLocalconfig
	cm.localConfig.Exploits = existingExploits

	err = yaml.NewEncoder(file).Encode(cm.localConfig)
	if err != nil {
		logger.Log.Error().Err(err).Msg("Error encoding configuration to YAML")
		return "", err
	}

	return "Local config reset successfully", nil
}

// UpdateLocalConfigToFile updates the local configuration and writes to file
func (cm *ConfigManager) UpdateLocalConfigToFile(configuration ConfigLocal) (string, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	err := os.MkdirAll(DefaultConfigPath, 0o755)
	if err != nil {
		logger.Log.Error().Err(err).Msg("Error creating config directory")
		return "", err
	}

	configPath := filepath.Join(DefaultConfigPath, "config.yml")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		logger.Log.Warn().Msg("Configuration file does not exist, creating a new one with default settings")
		os.WriteFile(configPath, ConfigTemplate, 0o644)
	} else if err != nil {
		logger.Log.Error().Err(err).Msg("Error checking configuration file")
		return "", err
	}

	file, err := os.Create(configPath)
	if err != nil {
		logger.Log.Error().Err(err).Msg("Error creating or opening configuration file")
		return "", err
	}
	defer file.Close()

	cm.localConfig = configuration

	err = yaml.NewEncoder(file).Encode(configuration)
	if err != nil {
		logger.Log.Error().Err(err).Msg("Error encoding configuration to YAML")
		return "", err
	}

	return configPath, nil
}

// GetSession retrieves the current stored session
func (cm *ConfigManager) GetSession() (string, error) {
	cm.mu.RLock()
	sessionPath := filepath.Join(DefaultConfigPath, "session")
	cm.mu.RUnlock()

	data, err := os.ReadFile(sessionPath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Logout handles user logout by removing session file
func (cm *ConfigManager) Logout() (string, error) {
	cm.mu.RLock()
	sessionPath := filepath.Join(DefaultConfigPath, "session")
	cm.mu.RUnlock()

	err := os.Remove(sessionPath)
	if err != nil {
		logger.Log.Error().Err(err).Msg("Error removing session file")
		return "", err
	}
	return "Logout successfully", nil
}

// ShowLocalConfigContent displays the current local configuration file content
func (cm *ConfigManager) ShowLocalConfigContent() (string, error) {
	cm.mu.RLock()
	configPath := filepath.Join(DefaultConfigPath, "config.yml")
	cm.mu.RUnlock()

	content, err := os.ReadFile(configPath)
	if err != nil {
		logger.Log.Error().Err(err).Msg("Error reading configuration file")
		return "", fmt.Errorf("error reading configuration file: %w", err)
	}

	return string(content), nil
}

// MapPortToService maps a port to a service name using the shared configuration
func (cm *ConfigManager) MapPortToService(port uint16) string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	for _, service := range cm.sharedConfig.ConfigClient.Services {
		if service.Port == port {
			return service.Name
		}
	}
	return ""
}

// MapServiceToPort maps a service name to a port using the shared configuration
func (cm *ConfigManager) MapServiceToPort(serviceName string) uint16 {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	for _, service := range cm.sharedConfig.ConfigClient.Services {
		if service.Name == serviceName {
			return service.Port
		}
	}
	return 0
}
