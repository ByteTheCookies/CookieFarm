package core

import (
	"context"
	"logger"
	"os"

	"server/config"
	"server/database"

	"gopkg.in/yaml.v3"
)

type Runner struct {
	store          *database.Store
	config         *config.ConfigManager
	shutdownCancel context.CancelFunc
}

func NewRunner(s *database.Store, c *config.ConfigManager) *Runner {
	return &Runner{store: s, config: c}
}

func (r *Runner) Submission() {
	ctx, cancel := context.WithCancel(context.Background())
	if r.shutdownCancel != nil {
		r.shutdownCancel()
	}
	r.shutdownCancel = cancel

	go r.StartFlagProcessingLoop(ctx)
}

// LoadConfigAndRun loads the configuration from the given path.
func (r *Runner) LoadConfig(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		logger.Log.Error().Err(err).Msg("Configuration file does not exist")
		return err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		logger.Log.Error().Err(err).Msg("Failed to read configuration file")
		return err
	}

	tmp := r.config.GetFullConfig()

	err = yaml.Unmarshal(data, &tmp)
	if err != nil {
		logger.Log.Error().Err(err).Msg("Failed to parse configuration file into tmp")
		return err
	}

	r.config.SetFullConfig(tmp)
	r.config.SetConfigured(true)

	return nil
}
