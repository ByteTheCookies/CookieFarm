package core

import (
	"context"
	"time"

	"github.com/ByteTheCookies/CookieFarm/internal/server/config"
	"github.com/ByteTheCookies/CookieFarm/internal/server/sqlite"
	"github.com/ByteTheCookies/CookieFarm/pkg/logger"
	"github.com/ByteTheCookies/CookieFarm/pkg/models"
	"github.com/ByteTheCookies/CookieFarm/pkg/protocols"
)

var shutdownCancel context.CancelFunc

// ----------- END FLAG GROUPS ------------

// StartFlagProcessingLoop starts the flag processing loop.
func StartFlagProcessingLoop(ctx context.Context) {
	interval := time.Duration(config.SharedConfig.ConfigServer.SubmitFlagCheckerTime) * time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	logger.Log.Info().Msg("Starting flag processing loop...")

	var err error
	config.Submit, err = protocols.LoadProtocol(config.SharedConfig.ConfigServer.Protocol)
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
			flags, err := sqlite.GetUnsubmittedFlagCodeList(config.SharedConfig.ConfigServer.MaxFlagBatchSize)
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
				config.SharedConfig.ConfigServer.URLFlagChecker,
				config.SharedConfig.ConfigServer.TeamToken,
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

func UpdateFlags(flags []protocols.ResponseProtocol) {
	statusCounts := map[string]int{
		models.StatusAccepted: 0,
		models.StatusDenied:   0,
		models.StatusError:    0,
	}

	valid := flags[:0]

	for _, f := range flags {
		if _, exists := statusCounts[f.Status]; exists {
			statusCounts[f.Status]++
			valid = append(valid, f)
		}
	}

	if err := sqlite.UpdateFlagsStatus(valid); err != nil {
		logger.Log.Error().
			Err(err).
			Msg("Failed to update flags")
	}

	total := statusCounts[models.StatusAccepted] + statusCounts[models.StatusDenied] + statusCounts[models.StatusError]
	logger.Log.Info().
		Int("accepted", statusCounts[models.StatusAccepted]).
		Int("denied", statusCounts[models.StatusDenied]).
		Int("errored", statusCounts[models.StatusError]).
		Int("total", total).
		Msg("Flags update summary")
}
