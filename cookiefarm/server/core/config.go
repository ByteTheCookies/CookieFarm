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
	shutdownCancel context.CancelFunc
}

func NewRunner(s *database.Store) *Runner {
	return &Runner{store: s}
}

func (r *Runner) Run() {
	ctx, cancel := context.WithCancel(context.Background())
	if r.shutdownCancel != nil {
		r.shutdownCancel()
	}
	r.shutdownCancel = cancel

	go r.StartFlagProcessingLoop(ctx)

	if config.SharedConfig.ConfigServer.FlagTTL != 0 {
		logger.Log.Warn().Msgf("Flag TTL is set to %d seconds, starting validation loop", config.SharedConfig.ConfigServer.FlagTTL)
		go r.ValidateFlagTTL(ctx, config.SharedConfig.ConfigServer.FlagTTL, config.SharedConfig.ConfigServer.TickTime)
	}
}

// LoadConfigAndRun loads the configuration from the given path.
func (*Runner) LoadConfig(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		logger.Log.Error().Err(err).Msg("Configuration file does not exist")
		return err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		logger.Log.Error().Err(err).Msg("Failed to read configuration file")
		return err
	}

	err = yaml.Unmarshal(data, &config.SharedConfig)
	if err != nil {
		logger.Log.Error().Err(err).Msg("Failed to parse configuration file")
		return err
	}

	if !config.SharedConfig.Configured {
		config.SharedConfig.Configured = true
	}

	return nil
}
