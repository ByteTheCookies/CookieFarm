package websockets

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gofiber/contrib/websocket"
)

type Client struct {
	Connection      *websocket.Conn
	Manager         *Manager
	Egress          chan Event
	Number          int
	ConnectionTimer *time.Timer
}

type Manager struct {
	Clients  ClientList
	Handlers map[string]EventHandler
	sync.RWMutex
}

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
