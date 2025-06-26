package websockets

import (
	"errors"
	"sync"
	"time"

	"github.com/ByteTheCookies/CookieFarm/pkg/logger"
)

var (
	circuitBreaker = &CircuitBreaker{ // circuitBreaker is the instance of the circuit breaker
		state:        StateClosed,
		failureCount: 0,
	}
	ErrCircuitOpen = errors.New("circuit breaker is open")
)

const (
	// Circuit breaker constants
	failureThreshold = 2                // Number of failed attempts before opening the circuit
	resetTimeout     = 30 * time.Second // Time to wait before switching to half-open state
	halfOpenMaxRetry = 1                // Try one more time in half-open state
)

// CircuitState state of the circuit breaker
type CircuitState int

// CircuitBreaker is a struct that implements the circuit breaker pattern
type CircuitBreaker struct {
	state           CircuitState
	failureCount    int
	lastFailureTime time.Time
	mutex           sync.Mutex
}

// CircuitBreaker methods

// RecordSuccess registers a successful connection and resets the failure count
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	cb.failureCount = 0
	cb.state = StateClosed
	logger.Log.Debug().Msg("Circuit breaker: Connection successful, circuit closed")
}

// RecordFailure registers a failed connection attempt and updates the state
func (cb *CircuitBreaker) RecordFailure() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	cb.failureCount++
	cb.lastFailureTime = time.Now()

	if cb.state == StateHalfOpen || cb.failureCount >= failureThreshold {
		cb.state = StateOpen
		logger.Log.Warn().Int("failures", cb.failureCount).Msg("Circuit breaker opened")
	}
}

// IsAllowed checks if a connection attempt is allowed based on the circuit breaker state
func (cb *CircuitBreaker) IsAllowed() bool {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	if cb.state == StateClosed {
		return true
	}

	if cb.state == StateOpen {
		if time.Since(cb.lastFailureTime) > resetTimeout {
			cb.state = StateHalfOpen
			logger.Log.Info().Msg("Circuit breaker: Switched to half-open state")
			return true
		}
		return false
	}

	return true
}
