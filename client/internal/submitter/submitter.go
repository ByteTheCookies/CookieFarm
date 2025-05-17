// Package submitter provides functions to manage the CookieFarm client submission.
package submitter

import (
	"github.com/ByteTheCookies/cookieclient/internal/logger"
	"github.com/ByteTheCookies/cookieclient/internal/models"
	"github.com/ByteTheCookies/cookieclient/internal/websockets"
	json "github.com/bytedance/sonic"
	gorilla "github.com/gorilla/websocket"
)

// Start initializes the submission loop to the cookiefarm server.
func Start(flagsChan <-chan models.Flag) {
	conn, err := websockets.GetConnection()
	if err != nil {
		logger.Log.Error().Err(err).Msg("Error connecting to WebSocket")
		return
	}

	for {
		select {
		case flag := <-flagsChan:
			flagObj := models.EventWSFlag{
				Type:    websockets.FlagEvent,
				Payload: flag,
			}
			marshalFlag, err := json.Marshal(flagObj)
			if err != nil {
				logger.Log.Error().Err(err).Msg("Error marshalling flag")
				continue
			}
			conn.WriteMessage(gorilla.TextMessage, marshalFlag)
		}
	}
}
