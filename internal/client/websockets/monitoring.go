// Package websockets used for communicating with the server via WebSocket protocol
package websockets

import (
	"sync"
	"time"

	"github.com/ByteTheCookies/CookieFarm/pkg/logger"
	"github.com/gorilla/websocket"
)

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

// Global instance of the connection monitor
var (
	monitor *ConnectionMonitor
	once    sync.Once
)

// GetMonitor returns the singleton instance of ConnectionMonitor
func GetMonitor() *ConnectionMonitor {
	once.Do(func() {
		monitor = &ConnectionMonitor{
			stats: ConnectionStats{
				CurrentStatus: StatusDisconnected,
			},
			stopChan: make(chan struct{}),
		}
	})
	return monitor
}

// SetConnection registers a WebSocket connection with the monitor
func (m *ConnectionMonitor) SetConnection(conn *websocket.Conn) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.conn = conn
	m.stats.ConnectionAttempts++
	m.stats.SuccessfulConnects++
	m.stats.CurrentStatus = StatusConnected
	m.stats.LastConnectTime = time.Now()
	m.stats.ConsecutiveErrs = 0

	if !m.isMonitoring {
		m.isMonitoring = true
		go m.startMonitoring()
	}

	logger.Log.Info().Msg("WebSocket connection registered with monitor")
}

// RecordDisconnect records a disconnection event
func (m *ConnectionMonitor) RecordDisconnect(err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.stats.CurrentStatus = StatusDisconnected
	m.stats.LastDisconnectTime = time.Now()
	if err != nil {
		m.stats.LastError = err
		m.stats.ConsecutiveErrs++
	}

	logger.Log.Info().
		Int("consecutive_errors", m.stats.ConsecutiveErrs).
		Time("last_connect", m.stats.LastConnectTime).
		Time("last_disconnect", m.stats.LastDisconnectTime).
		Int("messages_sent", m.stats.MessagesSent).
		Int("messages_received", m.stats.MessagesReceived).
		Dur("average_latency", m.stats.AverageLatency).
		Msg("WebSocket disconnected")
}

// RecordMessageSent records statistics about sent messages
func (m *ConnectionMonitor) RecordMessageSent() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.stats.MessagesSent++
	m.stats.LastSendTime = time.Now()
}

// RecordMessageReceived records statistics about received messages
func (m *ConnectionMonitor) RecordMessageReceived() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.stats.MessagesReceived++
	m.stats.LastReceiveTime = time.Now()
}

// MeasureLatency sends a ping and measures the time until pong is received
func (m *ConnectionMonitor) MeasureLatency() {
	if m.conn == nil {
		return
	}

	m.mutex.Lock()
	m.stats.LastPingTime = time.Now()
	m.mutex.Unlock()

	m.conn.SetPongHandler(func(string) error {
		m.mutex.Lock()
		defer m.mutex.Unlock()

		m.stats.LastPongTime = time.Now()
		latency := m.stats.LastPongTime.Sub(m.stats.LastPingTime)
		m.stats.CurrentLatency = latency

		m.stats.totalLatencySum += latency
		m.stats.latencyDataCount++
		m.stats.AverageLatency = m.stats.totalLatencySum / time.Duration(m.stats.latencyDataCount)

		return nil
	})

	// Send ping message
	err := m.conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(10*time.Second))
	if err != nil {
		logger.Log.Error().Err(err).Msg("Failed to send ping for latency measurement")
	}
}

// GetStatus returns the current connection status
func (m *ConnectionMonitor) GetStatus() ConnectionStatus {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.stats.CurrentStatus
}

// GetStats returns a copy of the current connection statistics
func (m *ConnectionMonitor) GetStats() ConnectionStats {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.stats
}

// SetStatus updates the connection status
func (m *ConnectionMonitor) SetStatus(status ConnectionStatus) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.stats.CurrentStatus = status
}

// startMonitoring begins the monitoring process
func (m *ConnectionMonitor) startMonitoring() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.performHealthCheck()
		case <-m.stopChan:
			return
		}
	}
}

// StopMonitoring stops the monitoring process
func (m *ConnectionMonitor) StopMonitoring() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.isMonitoring {
		close(m.stopChan)
		m.isMonitoring = false
		m.stopChan = make(chan struct{})
	}
}

// performHealthCheck performs periodic health checks on the connection
func (m *ConnectionMonitor) performHealthCheck() {
	m.mutex.RLock()
	conn := m.conn
	status := m.stats.CurrentStatus
	lastActivity := m.stats.LastReceiveTime
	if m.stats.LastSendTime.After(lastActivity) {
		lastActivity = m.stats.LastSendTime
	}
	m.mutex.RUnlock()

	if conn != nil && status == StatusConnected {
		m.MeasureLatency()

		if time.Since(lastActivity) > 5*time.Minute {
			logger.Log.Warn().
				Time("last_activity", lastActivity).
				Msg("WebSocket connection inactive for more than 5 minutes")
			m.RecordDisconnect(nil)
		}

		stats := m.GetStats()
		logger.Log.Debug().
			Int("messages_sent", stats.MessagesSent).
			Int("messages_received", stats.MessagesReceived).
			Dur("current_latency", stats.CurrentLatency).
			Dur("average_latency", stats.AverageLatency).
			Msg("WebSocket connection health check")
	}
}
