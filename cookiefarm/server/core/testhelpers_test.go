package core

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"protocols"
	"sync"
	"testing"
	"time"

	"server/database"

	_ "modernc.org/sqlite"
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

// newTestRunner creates a Runner backed by a fresh in-memory store.
func newTestRunner(t *testing.T) *Runner {
	t.Helper()
	return NewRunner(newTestStore(t))
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
		Status:       "UNSUBMITTED",
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

// insertFlags inserts multiple flags and fails on the first error.
func insertFlags(t *testing.T, store *database.Store, flags []database.Flag) {
	t.Helper()
	for _, f := range flags {
		insertFlag(t, store, f)
	}
}

// mustGetFlag fetches a flag by code and fails the test if it is not found.
func mustGetFlag(t *testing.T, store *database.Store, code string) database.Flag {
	t.Helper()
	f, err := store.Queries.GetFlagByCode(context.Background(), code)
	if err != nil {
		t.Fatalf("mustGetFlag(%q): %v", code, err)
	}
	return f
}

// --- Submit func stubs --------------------------------------------------------

// TestSubmitFunc is the same signature as protocols.SubmitFunc / config.Submit.
type TestSubmitFunc func(string, string, []string) ([]protocols.ResponseProtocol, error)

// fakeSubmit returns a TestSubmitFunc that echoes every flag back with a fixed
// status. If mu and calls are non-nil, each invocation appends a copy of the
// flag slice to calls in a thread-safe manner.
func fakeSubmit(status string, mu *sync.Mutex, calls *[][]string) TestSubmitFunc {
	return func(_ string, _ string, flags []string) ([]protocols.ResponseProtocol, error) {
		if calls != nil && mu != nil {
			cp := make([]string, len(flags))
			copy(cp, flags)
			mu.Lock()
			*calls = append(*calls, cp)
			mu.Unlock()
		}
		out := make([]protocols.ResponseProtocol, len(flags))
		for i, f := range flags {
			out[i] = protocols.ResponseProtocol{Flag: f, Status: status, Msg: "fake"}
		}
		return out, nil
	}
}

// errorSubmit returns a TestSubmitFunc that always returns the provided error.
func errorSubmit(err error) TestSubmitFunc {
	return func(string, string, []string) ([]protocols.ResponseProtocol, error) {
		return nil, err
	}
}

// errSubmit is the sentinel error used by errorSubmit-based tests.
var errSubmit = errors.New("submit: network error")

// --- Timing / polling helpers -------------------------------------------------

// waitFor polls cond every poll interval until it returns true or the timeout
// expires, at which point it fails the test with msg.
func waitFor(t *testing.T, timeout, poll time.Duration, msg string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(poll)
	}
	t.Fatalf("waitFor timed out after %s: %s", timeout, msg)
}

// waitForFlagStatus polls the store until the flag with the given code has the
// expected status, or fails the test after timeout.
func waitForFlagStatus(t *testing.T, store *database.Store, code, wantStatus string, timeout time.Duration) {
	t.Helper()
	waitFor(t, timeout, 20*time.Millisecond,
		fmt.Sprintf("flag %q never reached status %q", code, wantStatus),
		func() bool {
			f, err := store.Queries.GetFlagByCode(context.Background(), code)
			return err == nil && f.Status == wantStatus
		},
	)
}

// countFlagsWithStatus counts flags in the store with the given status value.
func countFlagsWithStatus(t *testing.T, store *database.Store, status string) int {
	t.Helper()
	flags, err := store.Queries.GetAllFlags(context.Background())
	if err != nil {
		t.Fatalf("countFlagsWithStatus: GetAllFlags: %v", err)
	}
	n := 0
	for _, f := range flags {
		if f.Status == status {
			n++
		}
	}
	return n
}

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
