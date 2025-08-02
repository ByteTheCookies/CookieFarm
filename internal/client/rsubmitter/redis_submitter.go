package rsubmitter

import (
	"context"

	"github.com/ByteTheCookies/CookieFarm/pkg/logger"
	"github.com/ByteTheCookies/CookieFarm/pkg/models"
	"github.com/ByteTheCookies/CookieFarm/pkg/redis"
)

var ctx = context.Background()

func Start(flagsChan <-chan models.ClientData) error {
	rdb := redis.NewRedisConnWithConfig(redis.DefaultRedisConfig())
	err := rdb.Connect(ctx)
	if err != nil {
		logger.Log.Error().Err(err).Msg("Failed to connect to Redis")
		return err
	}

	logger.Log.Info().Msg("Starting submission loop to the cookiefarm server with redis (redis) ...")
	for flag := range flagsChan {
		id, err := rdb.Add(ctx, flag)
		if err != nil {
			logger.Log.Error().Err(err).Msg("Failed to submit flag to Redis")
			continue // Skip this flag and continue with the next one
		}
		logger.Log.Debug().Str("id", id).Str("flag", flag.FlagCode).Msg("Flag submitted to Redis")
	}
	logger.Log.Info().Msg("Submission loop finished")
	return nil
}
