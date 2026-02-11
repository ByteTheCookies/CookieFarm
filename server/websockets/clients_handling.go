package websockets

import (
	"encoding/json"
	"logger"

	"github.com/gofiber/contrib/websocket"
)

func NewClient(conn *websocket.Conn, manager *Manager) *Client {
	return &Client{
		Connection: conn,
		Manager:    manager,
		Egress:     make(chan Event),
		Number:     manager.GetNextClientNumber(),
	}
}

func (c *Client) ReadMessages() {
	defer func() {
		c.Manager.RemoveClient(c)
	}()

	for {
		_, payload, err := c.Connection.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Log.Error().Err(err).Int("client", c.Number).Msg("Error reading message")
			}
			break
		}

		var request Event
		if err := json.Unmarshal(payload, &request); err != nil {
			logger.Log.Error().Err(err).Int("client", c.Number).Msg("Error unmarshalling event")
			continue
		}

		if err := c.Manager.RouteEvent(request, c); err != nil {
			logger.Log.Error().Err(err).Int("client", c.Number).Msg("Error routing event")
		}
	}
}

func (c *Client) WriteMessages() {
	defer func() {
		c.Manager.RemoveClient(c)
	}()

	for message := range c.Egress {
		data, err := json.Marshal(message)
		if err != nil {
			logger.Log.Error().Err(err).Int("client", c.Number).Msg("Failed to marshal message")
			continue
		}

		if err := c.Connection.WriteMessage(websocket.TextMessage, data); err != nil {
			logger.Log.Error().Err(err).Int("client", c.Number).Msg("Failed to send message")
			return
		}
	}

	// Channel closed; try to send a close message
	if err := c.Connection.WriteMessage(websocket.CloseMessage, []byte{}); err != nil {
		logger.Log.Error().Err(err).Int("client", c.Number).Msg("Connection closed")
	}
}

func (c *Client) CloseConnection(reason string) {
	logger.Log.Info().Int("client", c.Number).Str("reason", reason).Msg("Closing connection")
	c.Manager.RemoveClient(c)
}
