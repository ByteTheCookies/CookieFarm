package websockets

import (
	"errors"
	"fmt"
	"logger"
	"time"

	"server/config"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
)

var (
	ErrEventNotSupported = errors.New("this event type is not supported")
	ErrConnectionTimeout = errors.New("connection timeout exceeded")
	GlobalManager        *Manager // WebSocket manager
)

const (
	ConnectionLifetime = 24 * time.Hour // Lifetime of the connection
)

func NewManager() *Manager {
	m := &Manager{
		Clients:  make(ClientList),
		Handlers: make(map[string]EventHandler),
	}
	m.SetupEventHandlers()
	return m
}

func (m *Manager) GetNextClientNumber() int {
	m.Lock()
	defer m.Unlock()
	clientNumber := len(m.Clients) + 1
	return clientNumber
}

func (m *Manager) SetupEventHandlers() {
	m.Handlers[FlagMessage] = FlagHandler
}

func (m *Manager) RouteEvent(event Event, c *Client) error {
	if handler, ok := m.Handlers[event.Type]; ok {
		if err := handler(event, c); err != nil {
			return err
		}
		return nil
	} else {
		return ErrEventNotSupported
	}
}

// VerifyToken verifies the JWT token using the secret key
func VerifyToken(token string) error {
	_, err := jwt.Parse(token, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return config.Secret, nil
	})
	return err
}

// CookieAuthMiddleware verifies the JWT token from the cookie
func CookieAuthMiddleware(c *fiber.Ctx) error {
	token := c.Cookies("token")
	if token == "" || VerifyToken(token) != nil {
		return fiber.ErrUnauthorized
	}
	return c.Next()
}

// WebSocketUpgrade middleware for upgrading the connection to WebSocket
func WebSocketUpgrade(c *fiber.Ctx) error {
	if websocket.IsWebSocketUpgrade(c) {
		return c.Next()
	}
	return fiber.ErrUpgradeRequired
}

// ServeWS handles WebSocket connections.
//
// The handler passed to websocket.New MUST block for the entire lifetime of
// the connection – Fiber closes the underlying *Conn as soon as the handler
// returns. We therefore:
//
//  1. Spawn WriteMessages as a background goroutine.
//  2. Call ReadMessages directly, which blocks until the peer disconnects,
//     the read deadline fires, or CloseConnection is called.
//
// When ReadMessages returns it closes client.done, which wakes WriteMessages
// so it can send the WebSocket close frame, close the TCP connection, and
// remove the client from the manager.
func (m *Manager) ServeWS() fiber.Handler {
	return websocket.New(func(conn *websocket.Conn) {
		logger.Log.Debug().Msg("New WebSocket connection")

		client := NewClient(conn, m)
		m.AddClient(client)

		// Schedule automatic teardown after ConnectionLifetime.
		connectionTimer := time.AfterFunc(ConnectionLifetime, func() {
			logger.Log.Info().Int("client", client.Number).Msg("Connection lifetime exceeded, closing")
			client.CloseConnection("Connection lifetime exceeded")
		})
		client.ConnectionTimer = connectionTimer

		// WriteMessages is the sole writer goroutine; it exits when client.done
		// is closed and then calls RemoveClient.
		go client.WriteMessages()

		// Block here – Fiber keeps the *Conn alive while we're inside this
		// handler function.
		client.ReadMessages()
	}, websocket.Config{
		RecoverHandler: func(conn *websocket.Conn) {
			if err := recover(); err != nil {
				logger.Log.Error().Interface("error", err).Msg("WebSocket panic recovered")
				conn.Close()
			}
		},
	})
}

func (m *Manager) AddClient(client *Client) {
	m.Lock()
	defer m.Unlock()
	m.Clients[client] = true
}

// RemoveClient removes a client from the active-client map and stops its
// lifetime timer. It intentionally does NOT write to the connection –
// WriteMessages is the only goroutine allowed to write, and it already sends
// the close frame before calling RemoveClient.
func (m *Manager) RemoveClient(client *Client) {
	m.Lock()
	defer m.Unlock()

	if _, ok := m.Clients[client]; ok {
		if client.ConnectionTimer != nil {
			client.ConnectionTimer.Stop()
		}

		delete(m.Clients, client)
		logger.Log.Debug().
			Int("client", client.Number).
			Int("active_clients", len(m.Clients)).
			Msg("Client removed")
	}
}
