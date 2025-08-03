package core

import (
	"time"

	"github.com/ByteTheCookies/CookieFarm/internal/server/sqlite"
	"github.com/ByteTheCookies/CookieFarm/pkg/logger"
	"github.com/ByteTheCookies/CookieFarm/pkg/models"
	"github.com/nats-io/nats.go"

	json "github.com/bytedance/sonic"
)

func ListenForFlags(nc *nats.Conn) {
	logger.Log.Info().Msg("Starting to listen for flags...")

	maxBufferSize := 200

	js, err := nc.JetStream()
	if err != nil {
		logger.Log.Error().Err(err).Msg("Failed to create JetStream context")
		return
	}
	logger.Log.Info().Msg("JetStream context created successfully")

	sub, err := js.PullSubscribe("cookiefarm.flags", "flags-processor", nats.PullMaxWaiting(128))
	if err != nil {
		logger.Log.Error().Err(err).Msg("Failed to subscribe to flags topic")
		return
	}
	logger.Log.Info().Msg("Subscribed to flags topic successfully")

	flagsBuffer := make([]models.ClientData, 0, 200)

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if len(flagsBuffer) > 0 {
				logger.Log.Info().Msgf("Periodic tick: processing %d flags", len(flagsBuffer))
				sqlite.AddFlags(flagsBuffer)
				flagsBuffer = flagsBuffer[:0] // Clear the buffer
			}
		default:
			msgs, err := sub.Fetch(50)
			if err != nil {
				logger.Log.Fatal().Err(err)
			}

			for _, msg := range msgs {
				if len(flagsBuffer) >= maxBufferSize {
					logger.Log.Info().Msg("Flags buffer is full, processing existing flags")
					sqlite.AddFlags(flagsBuffer)
					flagsBuffer = flagsBuffer[:0] // Clear the buffer
				}

				logger.Log.Debug().Msgf("Received: %s\n", string(msg.Data))
				msg.Ack() // Ack

				var flag models.ClientData
				err := json.Unmarshal(msg.Data, &flag)
				if err != nil {
					logger.Log.Error().Err(err).Msg("Failed to unmarshal flag data")
					continue
				}
				flagsBuffer = append(flagsBuffer, flag)
			}
		}
	}
}
