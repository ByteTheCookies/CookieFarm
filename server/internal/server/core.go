package server

import (
	"context"
	"os"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/ByteTheCookies/cookieserver/internal/config"
	"github.com/ByteTheCookies/cookieserver/internal/database"
	"github.com/ByteTheCookies/cookieserver/internal/logger"
	"github.com/ByteTheCookies/cookieserver/internal/models"
	"github.com/ByteTheCookies/cookieserver/protocols"
)

// ----------- END FLAG GROUPS ------------

// StartFlagProcessingLoop starts the flag processing loop.
func StartFlagProcessingLoop(ctx context.Context) {
	interval := time.Duration(config.Current.ConfigServer.SubmitFlagCheckerTime) * time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	logger.Log.Info().Msg("Starting flag processing loop...")

	var err error
	config.Submit, err = protocols.LoadProtocol(config.Current.ConfigServer.Protocol)
	if err != nil {
		logger.Log.Error().Err(err).Msg("Failed to load protocol")
		return
	}

	// Main loop for flag processing.
	for {
		logger.Log.Debug().Msg("Waiting for flags to process...")
		select {
		case <-ctx.Done():
			logger.Log.Info().Msg("Flag processing loop terminated")
			return
		case <-ticker.C:
			flags, err := database.GetUnsubmittedFlagCodeList(config.Current.ConfigServer.MaxFlagBatchSize)
			if err != nil {
				logger.Log.Error().Err(err).Msg("Failed to get unsubmitted flags")
				continue
			}
			if len(flags) == 0 {
				logger.Log.Debug().Msg("No flags to submit")
				continue
			}

			logger.Log.Info().Int("count", len(flags)).Msg("Submitting flags to checker")

			responses, err := config.Submit(
				config.Current.ConfigServer.HostFlagchecker,
				config.Current.ConfigServer.TeamToken,
				flags,
			)
			if err != nil {
				logger.Log.Error().Err(err).Msg("Error submitting flags to checker")
				continue
			}

			UpdateFlags(responses)
		}
	}
}

// UpdateFlags updates the status of flags in the database.
func UpdateFlags(flags []models.ResponseProtocol) {
	valid := make([]models.ResponseProtocol, 0, len(flags))

	accepted, denied, errored := 0, 0, 0
	for _, f := range flags {
		switch f.Status {
		case models.StatusAccepted:
			accepted++
			valid = append(valid, f)

		case models.StatusDenied:
			denied++
			valid = append(valid, f)

		case models.StatusError:
			errored++
			valid = append(valid, f)

		default:
			continue
		}
	}

	if err := database.UpdateFlagsStatus(valid); err != nil {
		logger.Log.Error().
			Err(err).
			Msg("Failed to update flags")
	}

	total := accepted + denied + errored
	logger.Log.Info().
		Int("accepted", accepted).
		Int("denied", denied).
		Int("errored", errored).
		Int("total", total).
		Msg("Flags update summary")
}

// LoadConfig loads the configuration from the given path.
func LoadConfig(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		logger.Log.Error().Err(err).Msg("Configuration file does not exist")
		return err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		logger.Log.Error().Err(err).Msg("Failed to read configuration file")
		return err
	}

	err = yaml.Unmarshal(data, &config.Current)
	if err != nil {
		logger.Log.Error().Err(err).Msg("Failed to parse configuration file")
		return err
	}

	if !config.Current.Configured {
		config.Current.Configured = true
	}

	ctx, cancel := context.WithCancel(context.Background())
	if shutdownCancel != nil {
		shutdownCancel()
	}
	shutdownCancel = cancel

	go StartFlagProcessingLoop(ctx)

	return nil
}
