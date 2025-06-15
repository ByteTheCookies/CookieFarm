// Package websockets provides functions to manage the CookieFarm client submission.
package websockets

import (
	"github.com/ByteTheCookies/CookieFarm/pkg/logger"
	"github.com/ByteTheCookies/CookieFarm/pkg/models"
	json "github.com/bytedance/sonic"
	gorilla "github.com/gorilla/websocket"
)

type EventWS struct {
	Type    string `json:"type"`
	Payload []byte `json:"payload"`
}

type EventWSFlag struct {
	Type    string            `json:"type"`
	Payload models.ClientData `json:"payload"`
}

// Start initializes the submission loop to the cookiefarm server.
func Start(flagsChan <-chan models.ClientData) error {
	logger.Log.Info().Msg("Starting submission loop to the cookiefarm server...")
	conn, err := GetConnection()
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("Error connecting to WebSocket")
	}
	defer conn.Close()

	go WSReader(conn)

	for flag := range flagsChan {
		flagObj := EventWSFlag{
			Type:    FlagEvent,
			Payload: flag,
		}
		marshalFlag, err := json.Marshal(flagObj)
		if err != nil {
			logger.Log.Error().Err(err).Msg("Error marshalling flag")
			continue
		}
		if err := conn.WriteMessage(gorilla.TextMessage, marshalFlag); err != nil {
			logger.Log.Error().Err(err).Msg("Error sending flag, attempting reconnection")
			newConn, reconnectErr := GetConnection()
			if reconnectErr != nil {
				logger.Log.Fatal().Err(reconnectErr).Msg("Failed to reconnect to WebSocket")
			} else {
				conn = newConn
				logger.Log.Info().Msg("Successfully reconnected to WebSocket")
			}
			continue
		}
	}
	logger.Log.Info().Msg("Submission loop finished")
	return nil
}

// TODO: Sistema di accumulo in caso di non connesione fino a 2 minuti
