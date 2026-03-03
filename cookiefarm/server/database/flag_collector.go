package database

import (
	"context"
	"errors"
	"logger"
	"sync"
	"time"
)

const (
	maxBufferSize = 100
	flushInterval = 10 * time.Second
	flushTimeout  = 10 * time.Second
)

// FlagCollector manages the collection and flushing of flags to the database
type FlagCollector struct {
	buffer     []Flag         // Buffer for storing flags
	mutex      sync.Mutex     // Mutex for thread-safe access
	flushTimer *time.Timer    // Timer for periodic flushing
	stopChan   chan struct{}  // Channel to signal stop
	running    bool           // Indicates if the collector is running
	flushCond  *sync.Cond     // Condition variable for flushing
	stats      CollectorStats // Statistics about the collector
	store      *Store         // Reference to the database store
}

// CollectorStats holds statistics about the flag collector
type CollectorStats struct {
	TotalFlagsReceived  int
	TotalFlushes        int
	SuccessfulFlushes   int
	FailedFlushes       int
	LastFlushTime       time.Time
	LastSuccessfulFlush time.Time
	TotalFlagsFlushed   int
	LastError           error
}

// Singleton instance of the FlagCollector
var (
	collector *FlagCollector
	once      sync.Once
)

// GetCollector return the instance of FlagCollector singleton
func GetCollector() *FlagCollector {
	once.Do(func() {
		c := &FlagCollector{
			buffer:   make([]Flag, 0, maxBufferSize),
			stopChan: make(chan struct{}),
		}
		c.flushCond = sync.NewCond(&c.mutex)
		collector = c
	})
	return collector
}

// SetStore sets the database store on the collector.
// This must be called before the collector flushes any flags.
func (fc *FlagCollector) SetStore(s *Store) {
	fc.mutex.Lock()
	defer fc.mutex.Unlock()
	fc.store = s
}

// Start the collector to begin collecting flags
func (fc *FlagCollector) Start() {
	fc.mutex.Lock()
	defer fc.mutex.Unlock()

	if fc.running {
		return
	}

	fc.running = true
	fc.flushTimer = time.NewTimer(flushInterval)

	go func() {
		for {
			select {
			case <-fc.flushTimer.C:
				logger.Log.Debug().Msg("Flushing flags on timer")
				ctx, cancel := context.WithTimeout(context.Background(), flushTimeout)
				err := fc.FlushWithContext(ctx)
				cancel()

				if err != nil {
					logger.Log.Error().Err(err).Msg("Error flushing flag buffer on timer")
				}

				fc.mutex.Lock()
				if fc.running {
					fc.flushTimer.Reset(flushInterval)
				}
				fc.mutex.Unlock()
			case <-fc.stopChan:
				fc.mutex.Lock()
				if fc.flushTimer != nil {
					fc.flushTimer.Stop()
				}
				fc.running = false
				fc.flushCond.Broadcast()
				fc.mutex.Unlock()
				return
			}
		}
	}()

	logger.Log.Info().
		Int("max_buffer", maxBufferSize).
		Dur("flush_interval", flushInterval).
		Msg("Flag collector started")
}

// Stop the collector and flush any remaining flags
func (fc *FlagCollector) Stop() error {
	fc.mutex.Lock()

	if !fc.running {
		fc.mutex.Unlock()
		return nil
	}

	close(fc.stopChan)

	ctx, cancel := context.WithTimeout(context.Background(), flushTimeout)
	defer cancel()

	fc.mutex.Unlock()

	err := fc.FlushWithContext(ctx)
	fc.running = false

	// Take a locked snapshot of the stats so we don't race with the
	// background goroutine that may still be writing them (Issue 4.8).
	stats := fc.GetStats()
	bufferSize := fc.GetBufferSize()
	logger.Log.Info().
		Int("flags_processed", stats.TotalFlagsFlushed).
		Int("buffer_remaining", bufferSize).
		Int("successful_flushes", stats.SuccessfulFlushes).
		Int("failed_flushes", stats.FailedFlushes).
		Msg("Flag collector stopped")

	return err
}

// AddFlag adds a flag to the collector's buffer
func (fc *FlagCollector) AddFlag(flag Flag) error {
	fc.mutex.Lock()

	if !fc.running && fc.stopChan == nil {
		fc.mutex.Unlock()
		return nil
	}

	if !fc.running {
		fc.mutex.Unlock()
		fc.Start()
		fc.mutex.Lock()
	}

	fc.stats.TotalFlagsReceived++
	fc.buffer = append(fc.buffer, flag)

	if len(fc.buffer) >= maxBufferSize {
		logger.Log.Debug().Msg("Flushing flag buffer due to size limit")
		flagsToInsert := make([]Flag, len(fc.buffer))
		copy(flagsToInsert, fc.buffer)
		fc.buffer = fc.buffer[:0]

		fc.mutex.Unlock()
		ctx := context.Background()
		var err error
		for _, flag := range flagsToInsert {
			err = fc.store.Queries.AddFlag(ctx, MapFromFlagToDBParams(flag))
		}
		fc.mutex.Lock()

		fc.updateFlushStats(err, len(flagsToInsert))

		if err != nil {
			if len(fc.buffer)+len(flagsToInsert) <= maxBufferSize {
				fc.buffer = append(fc.buffer, flagsToInsert...)
			} else {
				logger.Log.Error().
					Int("dropped_flags", len(flagsToInsert)).
					Msg("Buffer overflow, dropped flags due to database error")
			}
			fc.mutex.Unlock()
			return err
		}
	}

	fc.mutex.Unlock()
	return nil
}

// Flush sends all accumulated flags to the database
func (fc *FlagCollector) Flush() error {
	return fc.FlushWithContext(context.Background())
}

// FlushWithContext sends all accumulated flags to the database using the provided context
func (fc *FlagCollector) FlushWithContext(ctx context.Context) error {
	fc.mutex.Lock()

	if len(fc.buffer) == 0 {
		fc.mutex.Unlock()
		return nil
	}

	flagsToInsert := make([]Flag, len(fc.buffer))
	copy(flagsToInsert, fc.buffer)
	fc.buffer = fc.buffer[:0]

	fc.stats.TotalFlushes++

	if fc.store == nil {
		fc.buffer = append(flagsToInsert, fc.buffer...)
		nilErr := errors.New("flag collector has no store set: call SetStore before flushing")
		fc.updateFlushStats(nilErr, 0)
		fc.mutex.Unlock()
		return nilErr
	}

	fc.mutex.Unlock()
	var err error
	for _, flag := range flagsToInsert {
		err = fc.store.Queries.AddFlag(ctx, MapFromFlagToDBParams(flag))
	}
	fc.mutex.Lock()
	defer fc.mutex.Unlock()

	fc.updateFlushStats(err, len(flagsToInsert))

	if err != nil {
		if len(fc.buffer)+len(flagsToInsert) <= maxBufferSize {
			fc.buffer = append(fc.buffer, flagsToInsert...)
		} else {
			logger.Log.Error().
				Int("dropped_flags", len(flagsToInsert)).
				Msg("Buffer overflow, dropped flags due to database error")
		}
		return err
	}

	fc.flushCond.Broadcast()

	return nil
}

// updateFlushStats updates the statistics after a flush operation
func (fc *FlagCollector) updateFlushStats(err error, flagCount int) {
	fc.stats.LastFlushTime = time.Now()

	if err != nil {
		fc.stats.FailedFlushes++
		fc.stats.LastError = err
	} else {
		fc.stats.SuccessfulFlushes++
		fc.stats.LastSuccessfulFlush = fc.stats.LastFlushTime
		fc.stats.TotalFlagsFlushed += flagCount

		logger.Log.Debug().
			Int("flag_count", flagCount).
			Time("flush_time", fc.stats.LastFlushTime).
			Int("total_flushed", fc.stats.TotalFlagsFlushed).
			Msg("Successfully flushed flags to database")
	}
}

// GetBufferSize returns the current size of the buffer
func (fc *FlagCollector) GetBufferSize() int {
	fc.mutex.Lock()
	defer fc.mutex.Unlock()

	return len(fc.buffer)
}

// GetStats returns the current statistics of the collector
func (fc *FlagCollector) GetStats() CollectorStats {
	fc.mutex.Lock()
	defer fc.mutex.Unlock()

	return fc.stats
}

// IsRunning checks if the collector is currently running
func (fc *FlagCollector) IsRunning() bool {
	fc.mutex.Lock()
	defer fc.mutex.Unlock()

	return fc.running
}
