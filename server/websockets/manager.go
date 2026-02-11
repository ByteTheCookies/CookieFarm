package websockets

import (
	"errors"
	"fmt"
	"logger"
	"server/config"
	"time"

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

// CookieAuthMiddleware verifica il token JWT dal cookie
func CookieAuthMiddleware(c *fiber.Ctx) error {
	token := c.Cookies("token")
	if token == "" || VerifyToken(token) != nil {
		return fiber.ErrUnauthorized
	}
	return c.Next()
}

// WebSocketUpgrade middleware per l'upgrade della connessione
func WebSocketUpgrade(c *fiber.Ctx) error {
	// IsWebSocketUpgrade controlla se è una richiesta di upgrade WebSocket
	if websocket.IsWebSocketUpgrade(c) {
		return c.Next()
	}
	return fiber.ErrUpgradeRequired
}

// ServeWS gestisce le connessioni WebSocket con il nuovo middleware di Fiber
func (m *Manager) ServeWS() fiber.Handler {
	return websocket.New(func(conn *websocket.Conn) {
		logger.Log.Debug().Msg("New WebSocket connection")

		client := NewClient(conn, m)
		m.AddClient(client)

		// Imposta il timer per la durata della connessione
		connectionTimer := time.AfterFunc(ConnectionLifetime, func() {
			logger.Log.Info().Int("client", client.Number).Msg("Connection lifetime exceeded, closing")
			client.CloseConnection("Connection lifetime exceeded")
		})
		client.ConnectionTimer = connectionTimer

		// Avvia le goroutine per lettura e scrittura
		go client.ReadMessages()
		go client.WriteMessages()

		logger.Log.Debug().Msg("Started reading messages")
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

func (m *Manager) RemoveClient(client *Client) {
	m.Lock()
	defer m.Unlock()

	if _, ok := m.Clients[client]; ok {
		if client.ConnectionTimer != nil {
			client.ConnectionTimer.Stop()
		}

		// Chiudi la connessione
		err := client.Connection.WriteMessage(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, "Server closing connection"),
		)
		if err != nil {
			logger.Log.Debug().Err(err).Int("client", client.Number).Msg("Failed to send close message")
		}

		client.Connection.Close()
		delete(m.Clients, client)
		logger.Log.Debug().Int("client", client.Number).Int("active_clients", len(m.Clients)).Msg("Client removed")
	}
}
