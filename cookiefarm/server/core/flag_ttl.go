package core

import (
	"context"
	"fmt"
	"logger"
	"time"
)

func (r *Runner) ValidateFlagTTL(ctx context.Context, flagTTL uint64, tickTime int) {
	interval := time.Duration(tickTime) * time.Second
	if interval <= 0 {
		logger.Log.Warn().Msgf("Invalid tickTime %d, defaulting to 60 seconds", tickTime)
		interval = 60 * time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	totalSecond := flagTTL * uint64(tickTime)
	for {
		logger.Log.Debug().Msg("Waiting for clean flags process...")
		select {
		case <-ctx.Done():
			logger.Log.Info().Msg("Flag TTL validation loop terminated")
			return
		case <-ticker.C:
			logger.Log.Info().Msgf("Checking for flags older than %d seconds", totalSecond)
			affectedRows, err := r.store.Queries.DeleteFlagByTTL(ctx, totalSecond)
			fmt.Printf("DeleteFlagByTTL affected %d rows\n", affectedRows)
			if err != nil {
				logger.Log.Error().Err(err).Msg("Failed to delete TTL flags")
			}

			if affectedRows > 0 {
				logger.Log.Info().Int64("affected_rows", affectedRows).Msg("Deleted TTL flags successfully")
			} else {
				logger.Log.Debug().Msg("No TTL flags to delete")
			}
		}
	}
}
