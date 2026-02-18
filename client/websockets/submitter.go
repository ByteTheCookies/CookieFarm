// Package websockets provides functions to manage the CookieFarm client submission.
package websockets

import (
	"logger"
	"models"

	json "github.com/bytedance/sonic"
	gorilla "github.com/gorilla/websocket"
)

// startReader launches a WSReader goroutine for the given connection and
// records a failure in the circuit breaker when the reader exits, so that
// the health-check / reconnect logic is aware the connection is gone.
func startReader(conn *gorilla.Conn) {
	go WSReader(conn)
}

// Start initializes the submission loop to the cookiefarm server.
func Start(flagsChan <-chan models.ClientData) error {
	logger.Log.Info().Msg("Starting submission loop to the cookiefarm server with websockets (websockets) ...")

	conn, err := GetConnection()
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("Error connecting to WebSocket")
	}
	defer conn.Close()

	// Start the reader for the initial connection.
	startReader(conn)

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

			// Close the old (broken) connection before reconnecting.
			_ = conn.Close()

			newConn, reconnectErr := GetConnection()
			if reconnectErr != nil {
				logger.Log.Fatal().Err(reconnectErr).Msg("Failed to reconnect to WebSocket")
			}

			conn = newConn
			// The new connection needs its own reader goroutine.
			startReader(conn)
			logger.Log.Info().Msg("Successfully reconnected to WebSocket")

			// Retry sending the flag on the fresh connection.
			if err := conn.WriteMessage(gorilla.TextMessage, marshalFlag); err != nil {
				logger.Log.Error().Err(err).Msg("Failed to send flag after reconnection, dropping flag")
			}
			continue
		}

		GetMonitor().RecordMessageSent()
	}

	logger.Log.Info().Msg("Submission loop finished")
	return nil
}
