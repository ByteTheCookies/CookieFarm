package nats

import (
	"encoding/json"

	"github.com/ByteTheCookies/CookieFarm/internal/client/config"
	"github.com/ByteTheCookies/CookieFarm/pkg/logger"
	"github.com/ByteTheCookies/CookieFarm/pkg/models"
	cnats "github.com/ByteTheCookies/CookieFarm/pkg/nats"
	"github.com/nats-io/nats.go"
)

var OnNewConfig func()

func ListenConfig() {
	logger.Log.Info().Msg("Starting to listen for configuration changes...")

	// Create a NATS client
	nc := cnats.NewNATSClient()
	nc.Connect(config.GetConfigManager().GetNATSToken())
	defer nc.Disconnect()

	if !nc.Conn.IsConnected() {
		logger.Log.Fatal().Msg("Failed to connect to NATS server")
		return
	}

	js, err := nc.Conn.JetStream()
	if err != nil {
		logger.Log.Error().Err(err).Msg("Failed to create JetStream context")
		return
	}

	sub, err := js.PullSubscribe("cookiefarm.config", "config-processor", nats.PullMaxWaiting(128))
	if err != nil {
		logger.Log.Error().Err(err).Msg("Failed to subscribe to config topic")
		return
	}

	logger.Log.Info().Msg("Subscribed to config topic successfully")
	cm := config.GetConfigManager()

	for {
		msgs, err := sub.Fetch(1)
		if err != nil {
			logger.Log.Fatal().Err(err)
			continue
		}

		for _, msg := range msgs {
			var configData models.ConfigShared
			if err := json.Unmarshal(msg.Data, &configData); err != nil {
				logger.Log.Error().Err(err).Msg("Error unmarshalling config data")
				continue
			}

			cm.SetSharedConfig(configData)
			if OnNewConfig != nil {
				go OnNewConfig()
			}
			msg.Ack() // Acknowledge the message
			logger.Log.Info().Msgf("Configuration updated: %+v", configData)
		}
	}
}
