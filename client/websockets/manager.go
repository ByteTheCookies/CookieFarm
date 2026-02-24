// Package websockets used for communicating with the server via WebSocket protocol
package websockets

import (
	"encoding/json"
	"errors"
	"logger"
	"models"
	"net/http"
	"strconv"
	"time"

	"client/config"

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

func tryToConnect(cm *config.ConfigManager, maxAttempts int, dialer *websocket.Dialer) (*websocket.Conn, error) {
	host := cm.GetLocalConfig().Host
	port := strconv.Itoa(int(cm.GetLocalConfig().Port))
	for attempts := range maxAttempts {
		conn, resp, err := dialer.Dial("ws://"+host+":"+port+"/ws/", http.Header{
			"Cookie": []string{"token=" + cm.GetToken()},
		})

		if resp != nil && resp.Body != nil {
			_ = resp.Body.Close()
		}

		if err == nil {
			circuitBreaker.RecordSuccess()
			monitor.SetConnection(conn)

			conn.SetPongHandler(func(appData string) error {
				if err := conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
					return err
				}
				monitor.RecordPong()
				return nil
			})

			conn.SetPingHandler(func(appData string) error {
				if err := conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
					return err
				}
				return conn.WriteControl(
					websocket.PongMessage,
					[]byte(appData),
					time.Now().Add(10*time.Second),
				)
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

	return nil, errors.New("exceeded maximum connection attempts")
}

// bad handshake (401)
// connection refused (503)
func GetConnection() (*websocket.Conn, error) {
	cm := config.GetConfigManager()
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

	conn, err := tryToConnect(cm, maxAttempts, dialer)
	if err == nil {
		logger.Log.Info().Msg("Successfully connected to WebSocket")
		monitor.SetStatus(StatusConnected)
		return conn, nil
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
