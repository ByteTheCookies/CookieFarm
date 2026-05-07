package config

import (
	"fmt"
	"logger"
	"maps"
	"os"
	"path/filepath"
	"sharedconfig"
	"sync"

	"gopkg.in/yaml.v3"
)

var (
	instance *ConfigManager
	once     sync.Once
)

const clientNameConfig = "client.yml"

const configTemplate = `
host: "localhost"
port: 8080
https: false
username: "cookieguest"
`

func GetInstance() *ConfigManager {
	once.Do(func() {
		instance = &ConfigManager{}

		instance.state.Store(&RuntimeConfig{
			Local: LocalConfig{},
			Shared: sharedconfig.Shared{
				Services: make(map[string]uint16),
			},
			Token: "",
		})
	})
	return instance
}

func (cm *ConfigManager) update(fn func(*RuntimeConfig)) {
	old := cm.state.Load().(*RuntimeConfig)

	newState := *old
	newState.Shared.Services = copyMap(old.Shared.Services)
	fn(&newState)

	cm.state.Store(&newState)
}

func copyMap(src map[string]uint16) map[string]uint16 {
	dst := make(map[string]uint16, len(src))
	maps.Copy(dst, src)
	return dst
}

func (cm *ConfigManager) Get() *RuntimeConfig {
	return cm.state.Load().(*RuntimeConfig)
}

func (cm *ConfigManager) GetHost() string {
	return cm.Get().Local.Host
}

func (cm *ConfigManager) GetPort() uint16 {
	return cm.Get().Local.Port
}

func (cm *ConfigManager) GetHTTPS() bool {
	return cm.Get().Local.HTTPS
}

func (cm *ConfigManager) GetUsername() string {
	return cm.Get().Local.Username
}

func (cm *ConfigManager) GetToken() string {
	return cm.Get().Token
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

func (cm *ConfigManager) SetToken(token string) {
	cm.update(func(s *RuntimeConfig) {
		s.Token = token
	})
}

func (cm *ConfigManager) SetLocalConfig(cfg LocalConfig) {
	cm.update(func(s *RuntimeConfig) {
		s.Local = cfg
	})
}

func (cm *ConfigManager) SetSharedConfig(sc sharedconfig.Shared) {
	cm.update(func(s *RuntimeConfig) {
		sc.Services = copyMap(sc.Services)
		s.Shared = sc
	})
}

func (cm *ConfigManager) SetHost(host string) {
	cm.update(func(s *RuntimeConfig) {
		s.Local.Host = host
	})
}

func (cm *ConfigManager) SetPort(port uint16) {
	cm.update(func(s *RuntimeConfig) {
		s.Local.Port = port
	})
}

func (cm *ConfigManager) SetHTTPS(https bool) {
	cm.update(func(s *RuntimeConfig) {
		s.Local.HTTPS = https
	})
}

func (cm *ConfigManager) SetUsername(username string) {
	cm.update(func(s *RuntimeConfig) {
		s.Local.Username = username
	})
}

func (cm *ConfigManager) Read() error {
	if err := read(&cm.Get().Local, clientNameConfig); err != nil {
		return err
	}

	if err := read(&cm.Get().Shared, "shared.yml"); err != nil {
		return err
	}

	token, err := cm.GetSession()
	if err != nil {
		return err
	}

	cm.SetToken(token)
	return nil
}

func (cm *ConfigManager) Write() error {
	err := cm.WriteLocal()
	if err != nil {
		return err
	}

	return cm.WriteShared()
}

func (cm *ConfigManager) WriteLocal() error {
	return write(&cm.Get().Local, clientNameConfig)
}

func (cm *ConfigManager) WriteShared() error {
	return write(&cm.Get().Shared, "shared.yml")
}

func (*ConfigManager) Reset() error {
	if err := os.MkdirAll(DefaultPath, 0o755); err != nil {
		return err
	}

	configPath := filepath.Join(DefaultPath, clientNameConfig)
	return os.WriteFile(configPath, []byte(configTemplate), 0o644)
}

func (*ConfigManager) GetSession() (string, error) {
	data, err := os.ReadFile(filepath.Join(DefaultPath, "session"))
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (*ConfigManager) Logout() (string, error) {
	sessionPath := filepath.Join(DefaultPath, "session")

	if err := os.Remove(sessionPath); err != nil {
		logger.Log.Error().Err(err).Msg("Error removing session")
		return "", err
	}

	return "Logout successfully", nil
}

func ReadConfig[T any](filename string) (*T, error) {
	content, err := os.ReadFile(filepath.Join(DefaultPath, filename))
	if err != nil {
		return nil, fmt.Errorf("error reading config: %w", err)
	}

	var cfg T
	if err := yaml.Unmarshal(content, &cfg); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	return &cfg, nil
}

func (*ConfigManager) ShowContent(filename string) (string, error) {
	content, err := os.ReadFile(filepath.Join(DefaultPath, filename))
	if err != nil {
		return "", fmt.Errorf("error reading config: %w", err)
	}

	return string(content), nil
}

func write[T any](cfg *T, filename string) error {
	configFilePath := filepath.Join(DefaultPath, filename)

	configFileContent, err := yaml.Marshal(*cfg)
	if err != nil {
		return fmt.Errorf("error marshalling config: %w", err)
	}

	err = os.WriteFile(configFilePath, configFileContent, 0o644)
	if err != nil {
		return fmt.Errorf("error writing config file: %w", err)
	}

	return nil
}

func read[T any](cfg *T, filename string) error {
	configFilePath := filepath.Join(DefaultPath, filename)

	configFileContent, err := os.ReadFile(configFilePath)
	if err != nil {
		return fmt.Errorf("error reading config file: %w", err)
	}

	err = yaml.Unmarshal(configFileContent, cfg)
	if err != nil {
		return fmt.Errorf("error unmarshalling config: %w", err)
	}

	return nil
}
