package websockets

import (
	"errors"
	"net/http"

	"github.com/ByteTheCookies/cookieserver/internal/logger"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/gorilla/websocket"
)

var ErrEventNotSupported = errors.New("this event type is not supported")

var websocketUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

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

func CookieAuthMiddleware(c *fiber.Ctx) error {
	token := c.Cookies("token")
	if token == "" || VerifyToken(token) != nil {
		return fiber.ErrUnauthorized
	}
	return nil
}

func (m *Manager) ServeWS(c *fiber.Ctx) error {
	if err := CookieAuthMiddleware(c); err != nil {
		logger.Log.Error().Err(err).Msg("Failed to authenticate user")
		return err
	}

	handler := adaptor.HTTPHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocketUpgrader.Upgrade(w, r, nil)
		if err != nil {
			logger.Log.Error().Err(err).Msg("Failed to upgrade connection")
			return
		}
		logger.Log.Debug().Msg("New WebSocket connection")

		client := NewClient(conn, m)
		m.AddClient(client)

		go client.ReadMessages()
		go client.WriteMessages()
		logger.Log.Debug().Msg("Started reading messages")
	}))

	handler(c)
	return nil
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
		client.Connection.Close()
		delete(m.Clients, client)
	}
}
