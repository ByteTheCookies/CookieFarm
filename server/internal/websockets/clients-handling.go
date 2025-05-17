package websockets

import (
	"encoding/json"
	"log"
	"time"

	"github.com/ByteTheCookies/cookieserver/internal/logger"
	"github.com/gorilla/websocket"
)

const (
	// pongWait is how long we will await a pong response from
	pongWait = 10 * time.Second
	// pingInterval has to be less than pongWait, We cant multiply by 0.9 to get 90% of time
	// Because that can make decimals, so instead *9 / 10 to get 90%
	// The reason why it has to be less than PingRequency is becuase otherwise it will send a new Ping before getting response
	pingInterval = (pongWait * 9) / 10
)

func NewClient(conn *websocket.Conn, manager *Manager) *Client {
	return &Client{
		Connection: conn,
		Manager:    manager,
		Egress:     make(chan []byte),
		Number:     manager.GetNextClientNumber(),
	}
}

func (client *Client) ReadMessages() {
	defer func() {
		client.Manager.RemoveClient(client)
	}()

	client.Connection.SetReadLimit(1024) // Potrei diminuirla

	if err := client.Connection.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		log.Println(err)
		return
	}
	client.Connection.SetPongHandler(client.PongHandler)

	for {
		_, payload, err := client.Connection.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Log.Error().Err(err).Msg("Error reading message")
			}
			break
		}

		var request Event
		if err := json.Unmarshal(payload, &request); err != nil {
			logger.Log.Error().Err(err).Msg("Error unmarshalling message")
			break
		}

		if err := client.Manager.RouteEvent(request, client); err != nil {
			logger.Log.Error().Err(err).Msg("Error routing event")
		}
	}
}

func (client *Client) PongHandler(pongMsg string) error {
	logger.Log.Debug().Msg("pong")
	return client.Connection.SetReadDeadline(time.Now().Add(pongWait))
}

func (client *Client) WriteMessages() {
	ticker := time.NewTicker(pingInterval)
	defer func() {
		ticker.Stop()
		client.Manager.RemoveClient(client)
	}()
	for {
		select {
		case message, ok := <-client.Egress:
			if !ok {
				if err := client.Connection.WriteMessage(websocket.CloseMessage, []byte{}); err != nil {
					logger.Log.Error().Err(err).Msg("Failed to write close message")
				}
				return
			}

			if err := client.Connection.WriteMessage(websocket.TextMessage, message); err != nil {
				logger.Log.Error().Err(err).Msg("Failed to write message")
			}
			logger.Log.Debug().Msgf("Sent message: %s", string(message))

		case <-ticker.C:
			logger.Log.Debug().Msg("ping")
			if err := client.Connection.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				logger.Log.Error().Err(err).Msg("Failed to write ping message")
				return
			}
		}
	}
}
