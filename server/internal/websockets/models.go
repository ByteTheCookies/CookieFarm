package websockets

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/ByteTheCookies/cookieserver/internal/config"
	"github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/websocket"
)

type Event struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// EventHandler is a function type for handling WebSocket events
type EventHandler func(Event, *Client) error

// NewMessageEvent represents a new message event
type NewMessageEvent struct {
	Sent time.Time `json:"sent"`
}

// ClientList maps clients to a boolean value indicating their status
type ClientList map[*Client]bool

// Client represents a WebSocket client connection
type Client struct {
	Connection *websocket.Conn
	Manager    *Manager
	Egress     chan []byte
	Number     int
}

// Manager handles WebSocket clients and event routing
type Manager struct {
	Clients ClientList
	sync.RWMutex
	Handlers map[string]EventHandler
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
