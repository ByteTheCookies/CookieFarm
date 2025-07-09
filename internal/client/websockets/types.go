package websockets

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/ByteTheCookies/CookieFarm/pkg/models"
	"github.com/gorilla/websocket"
)

// ========== Circuit Breaker States =========

// CircuitState state of the circuit breaker
type CircuitState int

// CircuitBreaker is a struct that implements the circuit breaker pattern
type CircuitBreaker struct {
	state           CircuitState
	failureCount    int
	lastFailureTime time.Time
	mutex           sync.Mutex
}

// ========== WebSocket Event Types =========

type Event struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// NewMessageEvent represents a new message event
type NewMessageEvent struct {
	Sent time.Time `json:"sent"`
}

type EventWS struct {
	Type    string `json:"type"`
	Payload []byte `json:"payload"`
}

type EventWSFlag struct {
	Type    string            `json:"type"`
	Payload models.ClientData `json:"payload"`
}

// ========= WebSocket Connection Monitoring =========

// ConnectionStatus represents the current status of the WebSocket connection
type ConnectionStatus int

const (
	StatusDisconnected ConnectionStatus = iota
	StatusConnecting
	StatusConnected
	StatusReconnecting
	StatusFailed
)

// ConnectionStats holds statistics about the WebSocket connection
type ConnectionStats struct {
	// Connection tracking
	ConnectionAttempts int
	SuccessfulConnects int
	FailedConnects     int
	LastConnectTime    time.Time
	LastDisconnectTime time.Time

	// Message tracking
	MessagesSent     int
	MessagesReceived int
	LastSendTime     time.Time
	LastReceiveTime  time.Time

	// Error tracking
	LastError       error
	ConsecutiveErrs int

	// Status
	CurrentStatus ConnectionStatus

	// Latency tracking
	LastPingTime     time.Time
	LastPongTime     time.Time
	CurrentLatency   time.Duration
	AverageLatency   time.Duration
	totalLatencySum  time.Duration
	latencyDataCount int
}

// ConnectionMonitor monitors WebSocket connections and provides statistics
type ConnectionMonitor struct {
	stats        ConnectionStats
	mutex        sync.RWMutex
	conn         *websocket.Conn
	isMonitoring bool
	stopChan     chan struct{}
}
