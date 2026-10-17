package core

import (
	"context"
	"database/sql"
	"fmt"
	"protocols"
	"sync"
	"testing"
	"time"

	"server/config"
	"server/database"

	_ "github.com/mattn/go-sqlite3"
)

// --- DB / Store helpers -------------------------------------------------------

// newTestDB opens a unique, isolated in-memory SQLite database, applies the
// schema, and registers a t.Cleanup that closes it.
func newTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	cfg := database.Config{
		DSN:             dsn,
		MaxOpenConns:    1,
		MaxIdleConns:    1,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 5 * time.Minute,
	}
	db, err := database.NewDB(cfg)
	if err != nil {
		t.Fatalf("newTestDB: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("newTestDB cleanup close: %v", err)
		}
	})
	return db
}

// newTestStore wraps newTestDB in a *database.Store.
func newTestStore(t *testing.T) *database.Store {
	t.Helper()
	return database.NewStore(newTestDB(t))
}

func newTestConfig(t *testing.T) *config.ConfigManager {
	t.Helper()
	return config.GetInstance()
}

// newTestRunner creates a Runner backed by a fresh in-memory store.
func newTestRunner(t *testing.T) *Runner {
	t.Helper()
	return NewRunner(newTestStore(t), newTestConfig(t))
}

// --- Flag fixtures ------------------------------------------------------------

// sampleFlag returns a deterministic, fully-populated database.Flag.
// The caller may override individual fields after the call.
func sampleFlag(code string) database.Flag {
	return database.Flag{
		FlagCode:     code,
		ServiceName:  "test-service",
		PortService:  8080,
		SubmitTime:   uint64(time.Now().Unix()),
		ResponseTime: uint64(time.Now().Unix()),
		Msg:          "ok",
		Status:       0,
		TeamID:       1,
		Username:     "tester",
		ExploitName:  "exploit_test",
	}
}

// insertFlag inserts a single flag into the store and fails the test on error.
func insertFlag(t *testing.T, store *database.Store, f database.Flag) {
	t.Helper()
	err := store.Queries.AddFlag(context.Background(), database.MapFromFlagToDBParams(f))
	if err != nil {
		t.Fatalf("insertFlag(%q): %v", f.FlagCode, err)
	}
}

// --- Submit func stubs --------------------------------------------------------

// TestSubmitFunc is the same signature as protocols.SubmitFunc / config.Submit.
type TestSubmitFunc func(string, string, []string) ([]protocols.ResponseProtocol, error)

// --- shutdownCancel helpers ---------------------------------------------------

// resetShutdownCancel cancels any running goroutines from a previous test and
// zeroes the package-level shutdownCancel. Must be called at the top of any
// test that exercises Runner.Run(). Also registers a Cleanup that tears it
// down after the test completes.
func resetShutdownCancel(t *testing.T, r *Runner) {
	t.Helper()
	if r.shutdownCancel != nil {
		r.shutdownCancel()
	}
	r.shutdownCancel = nil
	t.Cleanup(func() {
		if r.shutdownCancel != nil {
			r.shutdownCancel()
			r.shutdownCancel = nil
		}
	})
}

// --- CallCounter --------------------------------------------------------------

// CallCounter is a thread-safe invocation counter used by submit stubs.
type CallCounter struct {
	mu    sync.Mutex
	count int
}

// Inc increments the counter by one.
func (c *CallCounter) Inc() {
	c.mu.Lock()
	c.count++
	c.mu.Unlock()
}

// Value returns the current counter value.
func (c *CallCounter) Value() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.count
}
