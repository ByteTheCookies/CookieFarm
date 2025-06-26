// Package websockets used for communicating with the server via WebSocket protocol
package websockets

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/ByteTheCookies/CookieFarm/internal/client/config"
	"github.com/ByteTheCookies/CookieFarm/pkg/logger"
	"github.com/ByteTheCookies/CookieFarm/pkg/models"
	"github.com/gorilla/websocket"
)

var OnNewConfig func()

const (
	FlagEvent   = "flag"
	ConfigEvent = "config"

	maxRetries = 3

	// Connection timeouts
	dialTimeout  = 10 * time.Second // Timeout for WebSocket connection
	writeTimeout = 10 * time.Second // Timeout for writing message

	StateClosed   CircuitState = iota // Allowed connections
	StateHalfOpen                     // Allowed connections with limited retries
	StateOpen                         // Blocked connections

	pongWait = 60 * time.Second // pongWait is the time to wait for a pong response
)

type Event struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// NewMessageEvent represents a new message event
type NewMessageEvent struct {
	Sent time.Time `json:"sent"`
}

// bad handshake (401)
// connection refused (503)
func GetConnection() (*websocket.Conn, error) {
	cm := config.GetConfigManager()
	var conn *websocket.Conn
	var err error

	monitor := GetMonitor()
	monitor.SetStatus(StatusConnecting)

	if !circuitBreaker.IsAllowed() {
		logger.Log.Warn().Msg("Circuit breaker is open, connection attempt blocked")
		monitor.SetStatus(StatusFailed)
		return nil, ErrCircuitOpen
	}

	maxAttempts := maxRetries
	if circuitBreaker.state == StateHalfOpen {
		maxAttempts = halfOpenMaxRetry
	}

	dialer := &websocket.Dialer{
		HandshakeTimeout: dialTimeout,
	}

	for attempts := 0; attempts < maxAttempts; attempts++ {
		conn, _, err = dialer.Dial("ws://"+cm.GetLocalConfig().Host+":"+strconv.Itoa(int(cm.GetLocalConfig().Port))+"/ws/", http.Header{
			"Cookie": []string{"token=" + cm.GetToken()},
		})

		if err == nil {
			circuitBreaker.RecordSuccess()

			monitor.SetConnection(conn)

			conn.SetPongHandler(func(appData string) error {
				monitor.mutex.Lock()
				monitor.stats.LastPongTime = time.Now()
				monitor.mutex.Unlock()
				return nil
			})

			conn.SetReadDeadline(time.Now().Add(pongWait))

			return conn, nil
		}

		if websocket.ErrBadHandshake == err {
			logger.Log.Error().Err(err).Msg("Bad handshake, retrying login...")
			token, err := cm.GetSession()
			if err != nil {
				logger.Log.Error().Err(err).Msg("Failed to refresh token")
				circuitBreaker.RecordFailure()
				monitor.SetStatus(StatusFailed)
				return nil, err
			}
			cm.SetToken(token)
			continue
		}

		logger.Log.Warn().Err(err).Int("attempt", attempts+1).Int("maxRetries", maxAttempts).Msg("Error connecting to WebSocket, retrying...")
		monitor.SetStatus(StatusReconnecting)
		time.Sleep(time.Second * time.Duration(1<<attempts))
	}

	circuitBreaker.RecordFailure()
	monitor.SetStatus(StatusFailed)
	monitor.RecordDisconnect(err)
	logger.Log.Error().Err(err).Msg("Failed to connect to WebSocket after multiple attempts")
	return nil, err
}

// WSReader reads messages from the WebSocket connection and handles them
func WSReader(conn *websocket.Conn) {
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			logger.Log.Error().Err(err).Msg("Error reading message from WebSocket")
			circuitBreaker.RecordFailure()
			break
		}

		if err := WSHandleMessage(message); err != nil {
			logger.Log.Error().Err(err).Msg("Error handling message")
			break
		}
	}
}

// WSHandleMessage processes incoming WebSocket messages based on their type
func WSHandleMessage(message []byte) error {
	var event Event
	if err := json.Unmarshal(message, &event); err != nil {
		return err
	}

	logger.Log.Debug().Str("type", event.Type).Str("Payload", string(event.Payload)).Msg("Received event")

	switch event.Type {
	case ConfigEvent:
		return ConfigHandler(event.Payload)
	default:
		//
	}

	return nil
}

// ConfigHandler processes the configuration update received from the WebSocket server
func ConfigHandler(payload json.RawMessage) error {
	cm := config.GetConfigManager()
	var configReceived models.ConfigShared
	if err := json.Unmarshal(payload, &configReceived); err != nil {
		return err
	}

	cm.SetSharedConfig(configReceived)

	if OnNewConfig != nil {
		go OnNewConfig()
	}

	return nil
}
