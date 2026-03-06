package database

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"testing"
	"time"
)

// _sqlErrNoRows is a package-level alias for sql.ErrNoRows so that test files
// in this package can reference it without importing database/sql themselves.
var _sqlErrNoRows = sql.ErrNoRows

// --- DB helpers --------------------------------------------------------------

// newTestDB opens a unique in-memory SQLite database, applies the schema and
// returns a ready-to-use *sql.DB.  Each call gets a completely isolated DB
// (unique cache= name prevents cross-test sharing).
func newTestDB(t *testing.T) *sql.DB {
	t.Helper()
	// Use a named in-memory database so every test gets its own isolated
	// instance even when tests run in parallel.
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	cfg := Config{
		DSN:             dsn,
		MaxOpenConns:    1,
		MaxIdleConns:    1,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 5 * time.Minute,
	}
	db, err := NewDB(cfg)
	if err != nil {
		t.Fatalf("newTestDB: failed to open database: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("newTestDB cleanup: close error: %v", err)
		}
	})
	return db
}

// newTestStore wraps newTestDB in a Store.
func newTestStore(t *testing.T) *Store {
	t.Helper()
	return NewStore(newTestDB(t))
}

// newTestQueries returns a *Queries bound to a fresh in-memory database.
func newTestQueries(t *testing.T) *Queries {
	t.Helper()
	return New(newTestDB(t))
}

// --- Fixture helpers ---------------------------------------------------------

// sampleFlag returns a fully-populated Flag with deterministic values.
// The caller may override individual fields after calling this function.
func sampleFlag(flagCode string) Flag {
	return Flag{
		FlagCode:     flagCode,
		ServiceName:  "testservice",
		PortService:  8080,
		SubmitTime:   uint64(time.Now().Unix()),
		ResponseTime: uint64(time.Now().Unix()),
		Msg:          "ok",
		Status:       "ACCEPTED",
		TeamID:       1,
		Username:     "tester",
		ExploitName:  "exploit_test",
	}
}

// insertFlag is a convenience that inserts a single Flag via Queries.AddFlag
// and fails the test on error.
func insertFlag(t *testing.T, q *Queries, f Flag) {
	t.Helper()
	if err := q.AddFlag(context.Background(), MapFromFlagToDBParams(f)); err != nil {
		t.Fatalf("insertFlag(%q): %v", f.FlagCode, err)
	}
}

// insertFlags inserts multiple flags and fails the test on the first error.
func insertFlags(t *testing.T, q *Queries, flags []Flag) {
	t.Helper()
	for _, f := range flags {
		insertFlag(t, q, f)
	}
}

// mustGetFlag fetches a flag by code and fails the test if not found or on any
// other error.
func mustGetFlag(t *testing.T, q *Queries, code string) Flag {
	t.Helper()
	f, err := q.GetFlagByCode(context.Background(), code)
	if err != nil {
		t.Fatalf("mustGetFlag(%q): %v", code, err)
	}
	return f
}

// --- Assertion helpers --------------------------------------------------------

// assertFlagsEqual compares two Flag structs field by field and reports all
// mismatches via t.Errorf (non-fatal so all differences are shown at once).
func assertFlagsEqual(t *testing.T, want, got Flag) {
	t.Helper()
	if want.FlagCode != got.FlagCode {
		t.Errorf("FlagCode: want %q, got %q", want.FlagCode, got.FlagCode)
	}
	if want.ServiceName != got.ServiceName {
		t.Errorf("ServiceName: want %q, got %q", want.ServiceName, got.ServiceName)
	}
	if want.PortService != got.PortService {
		t.Errorf("PortService: want %d, got %d", want.PortService, got.PortService)
	}
	if want.SubmitTime != got.SubmitTime {
		t.Errorf("SubmitTime: want %d, got %d", want.SubmitTime, got.SubmitTime)
	}
	if want.ResponseTime != got.ResponseTime {
		t.Errorf("ResponseTime: want %d, got %d", want.ResponseTime, got.ResponseTime)
	}
	if want.Msg != got.Msg {
		t.Errorf("Msg: want %q, got %q", want.Msg, got.Msg)
	}
	if want.Status != got.Status {
		t.Errorf("Status: want %q, got %q", want.Status, got.Status)
	}
	if want.TeamID != got.TeamID {
		t.Errorf("TeamID: want %d, got %d", want.TeamID, got.TeamID)
	}
	if want.Username != got.Username {
		t.Errorf("Username: want %q, got %q", want.Username, got.Username)
	}
	if want.ExploitName != got.ExploitName {
		t.Errorf("ExploitName: want %q, got %q", want.ExploitName, got.ExploitName)
	}
}

// assertNoError fails the test immediately if err is non-nil.
func assertNoError(t *testing.T, err error, ctx string) {
	t.Helper()
	if err != nil {
		t.Fatalf("%s: unexpected error: %v", ctx, err)
	}
}

// assertInt64Equal fails the test if want != got.
func assertInt64Equal(t *testing.T, want, got int64, label string) {
	t.Helper()
	if want != got {
		t.Errorf("%s: want %d, got %d", label, want, got)
	}
}

// assertStringSliceLen fails the test if len(got) != want.
func assertStringSliceLen(t *testing.T, want int, got []string, label string) {
	t.Helper()
	if len(got) != want {
		t.Errorf("%s: want len=%d, got len=%d", label, want, len(got))
	}
}

// assertFlagSliceLen fails the test if len(got) != want.
func assertFlagSliceLen(t *testing.T, want int, got []Flag, label string) {
	t.Helper()
	if len(got) != want {
		t.Errorf("%s: want len=%d, got len=%d", label, want, len(got))
	}
}

// --- FlagCollector helpers ----------------------------------------------------

// newTestCollector creates a brand-new FlagCollector (bypassing the singleton)
// wired to the provided store, and registers cleanup that stops it.
// It exploits the fact that the unexported fields can be initialised directly
// because the test is in the same package (package database).
func newTestCollector(t *testing.T, store *Store) *FlagCollector {
	t.Helper()
	fc := &FlagCollector{
		buffer:   make([]Flag, 0, maxBufferSize),
		stopChan: make(chan struct{}),
		store:    store,
	}
	fc.flushCond = sync.NewCond(&fc.mutex)
	t.Cleanup(func() {
		// Guard against double-stop: Stop() closes stopChan synchronously but
		// sets fc.running = false only inside the background goroutine
		// (asynchronously).  If the test already called Stop(), we must wait
		// briefly for the goroutine to finish before checking IsRunning() so
		// we don't accidentally call Stop() again — which would panic on a
		// close of an already-closed channel.
		deadline := time.Now().Add(500 * time.Millisecond)
		for time.Now().Before(deadline) {
			select {
			case <-fc.stopChan:
				// channel is already closed — Stop() was already called
				return
			default:
			}
			time.Sleep(5 * time.Millisecond)
		}
		// stopChan is still open, so Stop() has not been called yet
		_ = fc.Stop()
	})
	return fc
}
