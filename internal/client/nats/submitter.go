package nats

import (
	"github.com/ByteTheCookies/CookieFarm/internal/client/config"
	"github.com/ByteTheCookies/CookieFarm/pkg/logger"
	"github.com/ByteTheCookies/CookieFarm/pkg/models"
	cnats "github.com/ByteTheCookies/CookieFarm/pkg/nats"
	"github.com/nats-io/nats.go"
)

func CleanUpPending(pendingFlags *[][]byte, nc *cnats.NATSClient) {
	for _, pendingFlag := range *pendingFlags {
		if err := nc.Publish("cookiefarm.flags", pendingFlag); err != nil {
			logger.Log.Error().Err(err).Msg("Error publishing pending flag data to NATS")
		}
	}
	*pendingFlags = [][]byte{} // Clear the slice after publishing
}

func Start(flagsChan <-chan models.ClientData) error {
	nc := cnats.NewNATSClient()
	nc.Connect(config.GetConfigManager().GetNATSToken())

	if !nc.Conn.IsConnected() {
		logger.Log.Fatal().Msg("Failed to connect to NATS server")
		return nats.ErrConnectionClosed
	} else {
		logger.Log.Info().Msg("Connected to NATS server successfully")
	}

	var pendingFlags [][]byte

	for flag := range flagsChan {
		flagData, err := flag.MarshalBinary()
		if err != nil {
			logger.Log.Error().Err(err).Msg("Error marshalling flag data")
			continue
		}

		if !nc.Conn.IsConnected() {
			logger.Log.Warn().Msg("NATS connection is not active, storing flag data for later")
			pendingFlags = append(pendingFlags, flagData)
		} else {
			if len(pendingFlags) > 0 {
				CleanUpPending(&pendingFlags, nc)
			}
			if err := nc.Publish("cookiefarm.flags", flagData); err != nil {
				logger.Log.Error().Err(err).Msg("Error publishing flag data to NATS")
				// Attempt to reconnect
				if err := nc.Connect(config.GetConfigManager().GetNATSToken()); err != nil {
					logger.Log.Fatal().Err(err).Msg("Failed to reconnect to NATS server")
				} else {
					logger.Log.Info().Msg("Successfully reconnected to NATS server")
				}
			}
		}
	}
	return nil
}
