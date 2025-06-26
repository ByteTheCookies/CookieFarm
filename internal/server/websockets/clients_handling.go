package websockets

import (
	"encoding/json"
	"log"
	"time"

	"github.com/ByteTheCookies/CookieFarm/pkg/logger"
	"github.com/gorilla/websocket"
)

const (
	// pongWait is how long we will await a pong response from
	pongWait = 60 * time.Second
	// pingInterval has to be less than pongWait, We cant multiply by 0.9 to get 90% of time
	// Because that can make decimals, so instead *9 / 10 to get 90%
	// The reason why it has to be less than PingRequency is because otherwise it will send a new Ping before getting response
	pingInterval = (pongWait * 9) / 10
	// writeWait is the time allowed to write a message to the peer
	writeWait = 10 * time.Second
	// maxMessageSize is the maximum message size allowed from peer
	maxMessageSize = 1024 // 1KB limit for security
)

func NewClient(conn *websocket.Conn, manager *Manager) *Client {
	return &Client{
		Connection: conn,
		Manager:    manager,
		Egress:     make(chan []byte),
		Number:     manager.GetNextClientNumber(),
		Closed:     make(chan struct{}),
		IsClosed:   false,
	}
}

func (client *Client) ReadMessages() {
	defer func() {
		client.CloseConnection("Read routine ended")
	}()

	client.Connection.SetReadLimit(maxMessageSize)

	if err := client.Connection.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		log.Println(err)
		return
	}
	client.Connection.SetPongHandler(client.PongHandler)

	for {
		select {
		case <-client.Closed:
			return
		default:
			_, payload, err := client.Connection.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					logger.Log.Error().Err(err).Msg("Error reading message")
				}
				return
			}

			var request Event
			if err := json.Unmarshal(payload, &request); err != nil {
				logger.Log.Error().Err(err).Msg("Error unmarshalling message")
				return
			}

			if err := client.Manager.RouteEvent(request, client); err != nil {
				logger.Log.Error().Err(err).Msg("Error routing event")
			}
		}
	}
}

func (client *Client) PongHandler(pongMsg string) error {
	logger.Log.Debug().Int("client", client.Number).Msg("Received pong")
	return client.Connection.SetReadDeadline(time.Now().Add(pongWait))
}

// CloseConnection safely closes a client connection and prevents goroutine leaks
func (client *Client) CloseConnection(reason string) {
	client.mutex.Lock()
	defer client.mutex.Unlock()

	if !client.IsClosed {
		logger.Log.Warn().Str("reason", reason).Int("client", client.Number).Msg("Closing client connection")
		client.IsClosed = true
		close(client.Closed)
		close(client.Egress)
		client.Manager.RemoveClient(client)
	}
}

func (client *Client) WriteMessages() {
	ticker := time.NewTicker(pingInterval)
	defer func() {
		ticker.Stop()
		client.CloseConnection("Write routine ended")
	}()
	for {
		select {
		case <-client.Closed:
			return
		case message, ok := <-client.Egress:
			if !ok {
				if err := client.Connection.WriteMessage(websocket.CloseMessage, []byte{}); err != nil {
					logger.Log.Error().Err(err).Msg("Failed to write close message")
				}
				return
			}

			if err := client.Connection.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(writeWait)); err != nil {
				logger.Log.Error().Err(err).Int("client", client.Number).Msg("Failed to write ping control message")
				return
			}

			if err := client.Connection.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				logger.Log.Error().Err(err).Int("client", client.Number).Msg("Failed to set write deadline")
				return
			}

			if err := client.Connection.WriteMessage(websocket.TextMessage, message); err != nil {
				logger.Log.Error().Err(err).Msg("Failed to write message")
				return
			}
			logger.Log.Debug().Msgf("Sent message: %s", string(message))

		case <-ticker.C:
			logger.Log.Debug().Int("client", client.Number).Msg("Sending ping")

			if err := client.Connection.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				logger.Log.Error().Err(err).Int("client", client.Number).Msg("Failed to set write deadline for ping")
				return
			}

			if err := client.Connection.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				logger.Log.Error().Err(err).Int("client", client.Number).Msg("Failed to write ping message")
				return
			}
		}
	}
}
