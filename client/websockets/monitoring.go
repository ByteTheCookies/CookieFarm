// Package websockets used for communicating with the server via WebSocket protocol
package websockets

import (
	"logger"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

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

// RecordPong is called by the pong handler set in GetConnection whenever a
// pong is received. If a ping was previously sent by MeasureLatency, it
// computes and stores the round-trip latency.
func (m *ConnectionMonitor) RecordPong() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	now := time.Now()
	m.stats.LastPongTime = now

	if m.stats.LastPingTime.IsZero() {
		return
	}

	latency := now.Sub(m.stats.LastPingTime)
	m.stats.CurrentLatency = latency
	m.stats.totalLatencySum += latency
	m.stats.latencyDataCount++
	m.stats.AverageLatency = m.stats.totalLatencySum / time.Duration(m.stats.latencyDataCount)
}

// MeasureLatency sends a ping and measures the time until pong is received.
// The pong handler is set once in GetConnection and calls RecordPong, so we
// must NOT override it here – doing so would lose the read-deadline reset.
func (m *ConnectionMonitor) MeasureLatency() {
	if m.conn == nil {
		return
	}

	m.mutex.Lock()
	m.stats.LastPingTime = time.Now()
	m.mutex.Unlock()

	// WriteControl is safe to call concurrently with WriteMessage.
	if err := m.conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(10*time.Second)); err != nil {
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
