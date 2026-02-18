package websockets

import (
	"encoding/json"
	"logger"
	"time"

	"github.com/gofiber/contrib/websocket"
)

const (
	pongWait   = 60 * time.Second    // Time to wait for a pong response before closing
	pingPeriod = (pongWait * 9) / 10 // How often to send a ping (must be < pongWait)
	writeWait  = 10 * time.Second    // Time allowed to write a message
	maxMsgSize = 512 * 1024       
)

func NewClient(conn *websocket.Conn, manager *Manager) *Client {
	return &Client{
		Connection: conn,
		Manager:    manager,
		Egress:     make(chan Event, 256), // Buffered to avoid blocking the writer on slow sends
		done:       make(chan struct{}),
		Number:     manager.GetNextClientNumber(),
	}
}

// signalDone closes the done channel exactly once, signalling WriteMessages to stop.
func (c *Client) signalDone() {
	c.doneOnce.Do(func() {
		close(c.done)
	})
}

// ReadMessages is the blocking read loop. It must be called directly (not as a
// goroutine) from ServeWS so that the Fiber handler stays alive for the full
// lifetime of the WebSocket connection.
//
// When the loop exits for any reason it calls signalDone(), which wakes up
// WriteMessages so that it can send the close frame and remove the client.
func (c *Client) ReadMessages() {
	defer c.signalDone()

	c.Connection.SetReadLimit(maxMsgSize)
	if err := c.Connection.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		logger.Log.Error().Err(err).Int("client", c.Number).Msg("Failed to set initial read deadline")
		return
	}

	// Each time a pong arrives, push the read deadline forward so the connection
	// stays alive as long as the client keeps responding to our pings.
	c.Connection.SetPongHandler(func(string) error {
		return c.Connection.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		_, payload, err := c.Connection.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Log.Error().Err(err).Int("client", c.Number).Msg("Unexpected WebSocket close error")
			}
			return
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

// WriteMessages is the single writer goroutine. It is the ONLY place that
// writes to the connection, which avoids all concurrent-write races.
//
// It stops when either:
//   - the Egress channel is closed, or
//   - the done channel is closed (i.e. ReadMessages exited), or
//   - a write to the connection fails.
//
// On exit it sends a proper WebSocket close frame, closes the underlying
// connection (which unblocks any lingering ReadMessages call), and then
// removes the client from the manager.
func (c *Client) WriteMessages() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()

		// Best-effort close frame – ignore the error since the connection may
		// already be gone by the time we get here.
		_ = c.Connection.SetWriteDeadline(time.Now().Add(writeWait))
		_ = c.Connection.WriteMessage(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
		)

		// Close the TCP connection so that a still-running ReadMessages loop
		// immediately returns instead of waiting for the read deadline.
		c.Connection.Close()

		c.Manager.RemoveClient(c)
	}()

	for {
		select {
		// Outgoing application message
		case message, ok := <-c.Egress:
			if !ok {
				// Channel was closed externally – time to shut down.
				return
			}
			if err := c.Connection.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				logger.Log.Error().Err(err).Int("client", c.Number).Msg("Failed to set write deadline")
				return
			}
			data, err := json.Marshal(message)
			if err != nil {
				logger.Log.Error().Err(err).Int("client", c.Number).Msg("Failed to marshal message")
				continue
			}
			if err := c.Connection.WriteMessage(websocket.TextMessage, data); err != nil {
				logger.Log.Error().Err(err).Int("client", c.Number).Msg("Failed to send message")
				return
			}

		// Periodic keepalive ping
		case <-ticker.C:
			if err := c.Connection.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				logger.Log.Error().Err(err).Int("client", c.Number).Msg("Failed to set write deadline for ping")
				return
			}
			if err := c.Connection.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				logger.Log.Error().Err(err).Int("client", c.Number).Msg("Failed to send ping, closing connection")
				return
			}

		// ReadMessages exited (or CloseConnection was called) – initiate shutdown
		case <-c.done:
			return
		}
	}
}

// CloseConnection forcefully closes a client connection from outside the
// read/write goroutines (e.g. from the connection lifetime timer).
// It signals WriteMessages to stop, and closes the underlying connection so
// that ReadMessages unblocks immediately.
func (c *Client) CloseConnection(reason string) {
	logger.Log.Info().Int("client", c.Number).Str("reason", reason).Msg("Closing connection")
	// Signal WriteMessages to start its cleanup sequence.
	c.signalDone()
	// Unblock ReadMessages so the Fiber handler can return promptly.
	c.Connection.Close()
}
