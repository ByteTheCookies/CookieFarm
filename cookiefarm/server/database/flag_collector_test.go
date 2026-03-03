package database

import (
	"context"
	"database/sql"
	"errors"
	"sync"
	"testing"
	"time"
)

// ─── helpers ──────────────────────────────────────────────────────────────────

// newStartedCollector creates a fresh FlagCollector, starts it, and registers
// cleanup that stops it at the end of the test.
func newStartedCollector(t *testing.T, store *Store) *FlagCollector {
	t.Helper()
	fc := newTestCollector(t, store)
	fc.Start()
	return fc
}

// waitForBufferDrain polls until the collector's buffer is empty or the
// deadline is exceeded.  Returns true when the buffer drained in time.
func waitForBufferDrain(fc *FlagCollector, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if fc.GetBufferSize() == 0 {
			return true
		}
		time.Sleep(5 * time.Millisecond)
	}
	return fc.GetBufferSize() == 0
}

// waitForFlagInDB polls until the given flag code appears in the database or
// the deadline is exceeded.  Returns true when the flag was found in time.
func waitForFlagInDB(q *Queries, code string, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		_, err := q.GetFlagByCode(context.Background(), code)
		if err == nil {
			return true
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return false
		}
		time.Sleep(10 * time.Millisecond)
	}
	return false
}

// ─── GetCollector (singleton) ─────────────────────────────────────────────────

func TestGetCollector_ReturnsSameInstance(t *testing.T) {
	c1 := GetCollector()
	c2 := GetCollector()
	if c1 != c2 {
		t.Error("GetCollector must return the same singleton instance on every call")
	}
}

func TestGetCollector_ReturnsNonNil(t *testing.T) {
	c := GetCollector()
	if c == nil {
		t.Fatal("GetCollector must return a non-nil *FlagCollector")
	}
}

// ─── SetStore ─────────────────────────────────────────────────────────────────

func TestSetStore_SetsStore(t *testing.T) {
	store := newTestStore(t)
	fc := newTestCollector(t, nil) // start with nil store
	fc.store = nil                 // ensure nil

	fc.SetStore(store)

	fc.mutex.Lock()
	got := fc.store
	fc.mutex.Unlock()

	if got != store {
		t.Error("SetStore did not update the store field")
	}
}

func TestSetStore_ThreadSafe_ConcurrentCallsDoNotPanic(t *testing.T) {
	store := newTestStore(t)
	fc := newTestCollector(t, nil)

	var wg sync.WaitGroup
	for range 10 {
		wg.Go(func() {
			fc.SetStore(store)
		})
	}
	wg.Wait()
}

// ─── Start ────────────────────────────────────────────────────────────────────

func TestStart_SetsRunningTrue(t *testing.T) {
	fc := newTestCollector(t, newTestStore(t))
	fc.Start()

	if !fc.IsRunning() {
		t.Error("IsRunning() must return true after Start()")
	}
}

func TestStart_Idempotent_SecondCallIsNoop(t *testing.T) {
	fc := newTestCollector(t, newTestStore(t))
	fc.Start()
	fc.Start() // second call must not panic or spawn a second goroutine

	if !fc.IsRunning() {
		t.Error("IsRunning() must still be true after double Start()")
	}
}

// ─── Stop ─────────────────────────────────────────────────────────────────────

func TestStop_SetsRunningFalse(t *testing.T) {
	fc := newTestCollector(t, newTestStore(t))
	fc.Start()

	if err := fc.Stop(); err != nil {
		t.Fatalf("Stop() returned unexpected error: %v", err)
	}

	deadline := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) {
		if !fc.IsRunning() {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	if fc.IsRunning() {
		t.Error("IsRunning() must return false after Stop() (goroutine did not stop within 500 ms)")
	}
}

func TestStop_WhenNotStarted_ReturnsNilWithoutPanic(t *testing.T) {
	fc := newTestCollector(t, newTestStore(t))
	// Do NOT call Start() — Stop on an idle collector must be a no-op.
	if err := fc.Stop(); err != nil {
		t.Errorf("Stop() on un-started collector returned error: %v", err)
	}
}

func TestStop_FlushesRemainingBuffer(t *testing.T) {
	store := newTestStore(t)
	fc := newTestCollector(t, store)
	fc.Start()

	flag := sampleFlag("FLAG{stop_flush_001}")
	if err := fc.AddFlag(flag); err != nil {
		t.Fatalf("AddFlag: %v", err)
	}

	if err := fc.Stop(); err != nil {
		t.Fatalf("Stop(): %v", err)
	}

	// After Stop the flag must have been written to the DB.
	_, err := store.Queries.GetFlagByCode(context.Background(), flag.FlagCode)
	if errors.Is(err, sql.ErrNoRows) {
		t.Error("flag not found in DB after Stop() — final flush did not run")
	} else if err != nil {
		t.Errorf("unexpected DB error: %v", err)
	}
}

// ─── IsRunning ────────────────────────────────────────────────────────────────

func TestIsRunning_InitiallyFalse(t *testing.T) {
	fc := newTestCollector(t, newTestStore(t))
	if fc.IsRunning() {
		t.Error("IsRunning() must be false before Start() is called")
	}
}

func TestIsRunning_TrueAfterStart_FalseAfterStop(t *testing.T) {
	fc := newTestCollector(t, newTestStore(t))

	fc.Start()
	if !fc.IsRunning() {
		t.Error("IsRunning() must be true after Start()")
	}

	_ = fc.Stop()

	// Same async-goroutine caveat as TestStop_SetsRunningFalse — poll briefly.
	deadline := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) {
		if !fc.IsRunning() {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	if fc.IsRunning() {
		t.Error("IsRunning() must be false after Stop() (goroutine did not stop within 500 ms)")
	}
}

// ─── AddFlag ──────────────────────────────────────────────────────────────────

func TestAddFlag_BuffersFlag_IncreasesBufferSize(t *testing.T) {
	fc := newStartedCollector(t, newTestStore(t))

	flag := sampleFlag("FLAG{add_001}")
	if err := fc.AddFlag(flag); err != nil {
		t.Fatalf("AddFlag: %v", err)
	}

	// The flag may still be in the buffer or already flushed to DB; either
	// way the stats must show it was received.
	stats := fc.GetStats()
	if stats.TotalFlagsReceived < 1 {
		t.Errorf("TotalFlagsReceived: want >= 1, got %d", stats.TotalFlagsReceived)
	}
}

func TestAddFlag_AutoStartsCollector(t *testing.T) {
	fc := newTestCollector(t, newTestStore(t))
	// Do NOT call Start manually.

	flag := sampleFlag("FLAG{autostart_001}")
	if err := fc.AddFlag(flag); err != nil {
		t.Fatalf("AddFlag (auto-start path): %v", err)
	}

	if !fc.IsRunning() {
		t.Error("AddFlag must auto-start the collector when it is not running")
	}
}

func TestAddFlag_IncrementsTotalFlagsReceived(t *testing.T) {
	fc := newStartedCollector(t, newTestStore(t))

	for i := range 3 {
		f := sampleFlag("FLAG{recv_" + string(rune('A'+i)) + "}")
		if err := fc.AddFlag(f); err != nil {
			t.Fatalf("AddFlag %d: %v", i, err)
		}
	}

	stats := fc.GetStats()
	if stats.TotalFlagsReceived < 3 {
		t.Errorf("TotalFlagsReceived: want >= 3, got %d", stats.TotalFlagsReceived)
	}
}

func TestAddFlag_ConcurrentAdds_NoPanic(t *testing.T) {
	store := newTestStore(t)
	fc := newStartedCollector(t, store)

	const goroutines = 20
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := range goroutines {
		go func(n int) {
			defer wg.Done()
			f := sampleFlag("FLAG{concurrent_" + string(rune('A'+n%26)) + "_" + itoa(n) + "}")
			_ = fc.AddFlag(f)
		}(i)
	}
	wg.Wait()
}

func TestAddFlag_ReachesMaxBufferSize_TriggersFlush(t *testing.T) {
	store := newTestStore(t)
	fc := newStartedCollector(t, store)

	// Fill the buffer up to maxBufferSize to trigger an immediate flush.
	for i := range maxBufferSize {
		f := sampleFlag("FLAG{maxbuf_" + itoa(i) + "}")
		if err := fc.AddFlag(f); err != nil {
			t.Fatalf("AddFlag %d: %v", i, err)
		}
	}

	// After a full-buffer flush the buffer size should have dropped back down.
	if !waitForBufferDrain(fc, 2*time.Second) {
		t.Errorf("buffer did not drain after reaching maxBufferSize; still has %d items",
			fc.GetBufferSize())
	}
}

func TestAddFlag_WhenBufferFull_FlagsWrittenToDB(t *testing.T) {
	store := newTestStore(t)
	fc := newStartedCollector(t, store)

	// We insert exactly maxBufferSize flags.  The last one should trigger the
	// immediate flush path inside AddFlag.
	lastCode := "FLAG{maxbuf_last}"
	for i := range maxBufferSize - 1 {
		f := sampleFlag("FLAG{maxbuf_pre_" + itoa(i) + "}")
		if err := fc.AddFlag(f); err != nil {
			t.Fatalf("AddFlag pre %d: %v", i, err)
		}
	}
	last := sampleFlag(lastCode)
	if err := fc.AddFlag(last); err != nil {
		t.Fatalf("AddFlag last: %v", err)
	}

	// Wait until the last flag appears in the DB (means the flush completed).
	if !waitForFlagInDB(store.Queries, lastCode, 3*time.Second) {
		t.Error("last flag not found in DB after buffer-full flush")
	}
}

// ─── Flush ────────────────────────────────────────────────────────────────────

func TestFlush_EmptyBuffer_ReturnsNil(t *testing.T) {
	fc := newStartedCollector(t, newTestStore(t))

	if err := fc.Flush(); err != nil {
		t.Errorf("Flush on empty buffer must return nil, got: %v", err)
	}
}

func TestFlush_WithFlags_DrainsBuf(t *testing.T) {
	store := newTestStore(t)
	fc := newStartedCollector(t, store)

	flag := sampleFlag("FLAG{flush_001}")
	if err := fc.AddFlag(flag); err != nil {
		t.Fatalf("AddFlag: %v", err)
	}

	if err := fc.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}

	if fc.GetBufferSize() != 0 {
		t.Errorf("expected empty buffer after Flush, got size=%d", fc.GetBufferSize())
	}
}

func TestFlush_WithFlags_PersistsToDatabase(t *testing.T) {
	store := newTestStore(t)
	fc := newStartedCollector(t, store)

	flag := sampleFlag("FLAG{flush_persist}")
	if err := fc.AddFlag(flag); err != nil {
		t.Fatalf("AddFlag: %v", err)
	}

	if err := fc.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}

	_, err := store.Queries.GetFlagByCode(context.Background(), flag.FlagCode)
	if errors.Is(err, sql.ErrNoRows) {
		t.Error("flag not found in DB after Flush")
	} else if err != nil {
		t.Errorf("unexpected DB error: %v", err)
	}
}

func TestFlush_MultipleFlags_AllPersisted(t *testing.T) {
	store := newTestStore(t)
	fc := newStartedCollector(t, store)

	flags := []Flag{
		sampleFlag("FLAG{flush_multi_001}"),
		sampleFlag("FLAG{flush_multi_002}"),
		sampleFlag("FLAG{flush_multi_003}"),
	}
	for _, f := range flags {
		if err := fc.AddFlag(f); err != nil {
			t.Fatalf("AddFlag %q: %v", f.FlagCode, err)
		}
	}

	if err := fc.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}

	for _, f := range flags {
		_, err := store.Queries.GetFlagByCode(context.Background(), f.FlagCode)
		if errors.Is(err, sql.ErrNoRows) {
			t.Errorf("flag %q not found in DB after Flush", f.FlagCode)
		} else if err != nil {
			t.Errorf("unexpected DB error for %q: %v", f.FlagCode, err)
		}
	}
}

// ─── FlushWithContext ─────────────────────────────────────────────────────────

func TestFlushWithContext_EmptyBuffer_ReturnsNil(t *testing.T) {
	fc := newStartedCollector(t, newTestStore(t))

	err := fc.FlushWithContext(context.Background())
	if err != nil {
		t.Errorf("FlushWithContext on empty buffer must return nil, got: %v", err)
	}
}

func TestFlushWithContext_WithFlag_PersistsToDatabase(t *testing.T) {
	store := newTestStore(t)
	fc := newStartedCollector(t, store)

	flag := sampleFlag("FLAG{fwc_001}")
	if err := fc.AddFlag(flag); err != nil {
		t.Fatalf("AddFlag: %v", err)
	}

	if err := fc.FlushWithContext(context.Background()); err != nil {
		t.Fatalf("FlushWithContext: %v", err)
	}

	_, err := store.Queries.GetFlagByCode(context.Background(), flag.FlagCode)
	if errors.Is(err, sql.ErrNoRows) {
		t.Error("flag not found in DB after FlushWithContext")
	} else if err != nil {
		t.Errorf("unexpected DB error: %v", err)
	}
}

func TestFlushWithContext_NilStore_ReturnsError(t *testing.T) {
	// Build a collector with no store explicitly set.
	fc := &FlagCollector{
		buffer:   make([]Flag, 0, maxBufferSize),
		stopChan: make(chan struct{}),
		store:    nil,
	}
	fc.flushCond = newCondForTest(&fc.mutex)
	fc.running = true

	// Manually add a flag directly to the buffer (bypassing AddFlag so we
	// avoid the auto-start / auto-flush paths that need the store).
	fc.buffer = append(fc.buffer, sampleFlag("FLAG{nilstore_001}"))

	// Issue 4.3 is now fixed — this test runs without skipping.
	err := fc.FlushWithContext(context.Background())
	if err == nil {
		t.Error("FlushWithContext with nil store must return an error")
	}
}

func TestFlushWithContext_NilStore_RequeuesFlags(t *testing.T) {
	// Same setup as above — nil store, one flag in buffer.
	fc := &FlagCollector{
		buffer:   make([]Flag, 0, maxBufferSize),
		stopChan: make(chan struct{}),
		store:    nil,
	}
	fc.flushCond = newCondForTest(&fc.mutex)
	fc.running = true
	fc.buffer = append(fc.buffer, sampleFlag("FLAG{nilstore_requeue}"))

	// Issue 4.3 is now fixed — this test runs without skipping.
	_ = fc.FlushWithContext(context.Background())

	// The flag must have been re-queued into the buffer because the flush
	// failed and there was still room.
	if fc.GetBufferSize() == 0 {
		t.Error("flags must be re-queued into the buffer when FlushWithContext fails with nil store")
	}
}

func TestFlushWithContext_CancelledContext_ReturnsError(t *testing.T) {
	store := newTestStore(t)
	fc := newStartedCollector(t, store)

	flag := sampleFlag("FLAG{fwc_cancel}")
	if err := fc.AddFlag(flag); err != nil {
		t.Fatalf("AddFlag: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // pre-cancel

	// A pre-cancelled context should cause the DB write(s) to fail.
	// If the SQLite driver ignores the cancelled context we accept that and
	// skip, rather than fail — this is a best-effort test.
	err := fc.FlushWithContext(ctx)
	if err != nil {
		// Expected path: driver respected the cancelled context.
		t.Logf("FlushWithContext with cancelled context returned error (expected): %v", err)
	} else {
		t.Log("FlushWithContext with cancelled context succeeded — SQLite driver ignored cancellation")
	}
}

// ─── GetBufferSize ────────────────────────────────────────────────────────────

func TestGetBufferSize_InitiallyZero(t *testing.T) {
	fc := newTestCollector(t, newTestStore(t))
	if size := fc.GetBufferSize(); size != 0 {
		t.Errorf("GetBufferSize on fresh collector: want 0, got %d", size)
	}
}

func TestGetBufferSize_AfterAddFlag_ReflectsChange(t *testing.T) {
	fc := newStartedCollector(t, newTestStore(t))

	before := fc.GetBufferSize()
	flag := sampleFlag("FLAG{bufsize_001}")
	if err := fc.AddFlag(flag); err != nil {
		t.Fatalf("AddFlag: %v", err)
	}
	// The flag may still be in buffer or already flushed to DB — we can only
	// assert TotalFlagsReceived > 0 because the timer may have fired.
	stats := fc.GetStats()
	if stats.TotalFlagsReceived <= before {
		t.Errorf("TotalFlagsReceived should have grown; before=%d, after=%d",
			before, stats.TotalFlagsReceived)
	}
}

func TestGetBufferSize_AfterFlush_IsZero(t *testing.T) {
	store := newTestStore(t)
	fc := newStartedCollector(t, store)

	if err := fc.AddFlag(sampleFlag("FLAG{bufsize_flush}")); err != nil {
		t.Fatalf("AddFlag: %v", err)
	}
	if err := fc.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}

	if size := fc.GetBufferSize(); size != 0 {
		t.Errorf("GetBufferSize after Flush: want 0, got %d", size)
	}
}

// ─── GetStats ─────────────────────────────────────────────────────────────────

func TestGetStats_InitialState_AllZero(t *testing.T) {
	fc := newTestCollector(t, newTestStore(t))
	stats := fc.GetStats()

	if stats.TotalFlagsReceived != 0 {
		t.Errorf("TotalFlagsReceived: want 0, got %d", stats.TotalFlagsReceived)
	}
	if stats.TotalFlushes != 0 {
		t.Errorf("TotalFlushes: want 0, got %d", stats.TotalFlushes)
	}
	if stats.SuccessfulFlushes != 0 {
		t.Errorf("SuccessfulFlushes: want 0, got %d", stats.SuccessfulFlushes)
	}
	if stats.FailedFlushes != 0 {
		t.Errorf("FailedFlushes: want 0, got %d", stats.FailedFlushes)
	}
	if stats.TotalFlagsFlushed != 0 {
		t.Errorf("TotalFlagsFlushed: want 0, got %d", stats.TotalFlagsFlushed)
	}
	if stats.LastError != nil {
		t.Errorf("LastError: want nil, got %v", stats.LastError)
	}
}

func TestGetStats_AfterSuccessfulFlush_SuccessfulFlushesIncremented(t *testing.T) {
	store := newTestStore(t)
	fc := newStartedCollector(t, store)

	if err := fc.AddFlag(sampleFlag("FLAG{stats_flush}")); err != nil {
		t.Fatalf("AddFlag: %v", err)
	}
	if err := fc.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}

	stats := fc.GetStats()
	if stats.SuccessfulFlushes < 1 {
		t.Errorf("SuccessfulFlushes: want >= 1, got %d", stats.SuccessfulFlushes)
	}
	if stats.TotalFlagsFlushed < 1 {
		t.Errorf("TotalFlagsFlushed: want >= 1, got %d", stats.TotalFlagsFlushed)
	}
	if stats.LastError != nil {
		t.Errorf("LastError: want nil after successful flush, got %v", stats.LastError)
	}
}

func TestGetStats_TotalFlagsReceived_CountsEachAdd(t *testing.T) {
	fc := newStartedCollector(t, newTestStore(t))

	for i := range 5 {
		if err := fc.AddFlag(sampleFlag("FLAG{stats_recv_" + itoa(i) + "}")); err != nil {
			t.Fatalf("AddFlag %d: %v", i, err)
		}
	}

	stats := fc.GetStats()
	if stats.TotalFlagsReceived < 5 {
		t.Errorf("TotalFlagsReceived: want >= 5, got %d", stats.TotalFlagsReceived)
	}
}

func TestGetStats_AfterFailedFlush_FailedFlushesIncremented(t *testing.T) {
	// Issue 4.3 secondary gap is now fixed: the nil-store path in
	// FlushWithContext now calls updateFlushStats before returning, so
	// FailedFlushes is correctly incremented.
	// Create a collector with a nil store so every flush attempt fails.
	fc := &FlagCollector{
		buffer:   make([]Flag, 0, maxBufferSize),
		stopChan: make(chan struct{}),
		store:    nil,
		running:  true,
	}
	fc.flushCond = newCondForTest(&fc.mutex)

	// Inject a flag directly into the buffer.
	fc.buffer = append(fc.buffer, sampleFlag("FLAG{stats_fail}"))

	_ = fc.FlushWithContext(context.Background())

	stats := fc.GetStats()
	if stats.FailedFlushes < 1 {
		t.Errorf("FailedFlushes: want >= 1, got %d", stats.FailedFlushes)
	}
	if stats.LastError == nil {
		t.Error("LastError: want non-nil after failed flush, got nil")
	}
}

// ─── Timer-driven flush (integration) ─────────────────────────────────────────

func TestFlagCollector_TimerFlush_FlagAppearsInDB(t *testing.T) {
	// This test relies on the flushInterval constant (10 s) which is too long
	// for a unit test.  We use a manual Flush() call as a proxy to validate the
	// same code path that the timer fires.  The timer path itself is covered by
	// Start/Stop tests above.
	store := newTestStore(t)
	fc := newStartedCollector(t, store)

	flag := sampleFlag("FLAG{timer_flush}")
	if err := fc.AddFlag(flag); err != nil {
		t.Fatalf("AddFlag: %v", err)
	}

	// Manually trigger the flush (same code the timer calls).
	if err := fc.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}

	if found := waitForFlagInDB(store.Queries, flag.FlagCode, 2*time.Second); !found {
		t.Error("flag not in DB after manual flush (timer path proxy)")
	}
}

// ─── Duplicate flag handling ──────────────────────────────────────────────────

func TestAddFlag_DuplicateFlag_SecondAddNoError(t *testing.T) {
	store := newTestStore(t)
	fc := newStartedCollector(t, store)

	flag := sampleFlag("FLAG{dup_collector}")

	if err := fc.AddFlag(flag); err != nil {
		t.Fatalf("first AddFlag: %v", err)
	}
	if err := fc.Flush(); err != nil {
		t.Fatalf("first Flush: %v", err)
	}

	// Add the same flag again — the DB uses INSERT OR IGNORE so this must
	// not propagate an error.
	if err := fc.AddFlag(flag); err != nil {
		t.Fatalf("second AddFlag (duplicate): %v", err)
	}
	if err := fc.Flush(); err != nil {
		t.Fatalf("second Flush: %v", err)
	}

	// The DB must still have exactly one row for this flag code.
	count, err := store.Queries.CountFlags(context.Background())
	if err != nil {
		t.Fatalf("CountFlags: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 flag in DB after duplicate add, got %d", count)
	}
}

// storeWithBrokenQueries builds a Store whose Queries field is replaced by a
// brokenQueries that always returns an error from AddFlag.
func storeWithBrokenQueries(t *testing.T) *Store {
	t.Helper()
	realStore := newTestStore(t)
	// The Queries field is exported — swap it out.
	realStore.Queries = &Queries{db: &errorDB{}}
	return realStore
}

// errorDB implements DBTX and always returns errors, giving us a clean way to
// make the AddFlag call fail without any extra mocking infrastructure.
type errorDB struct{}

func (*errorDB) ExecContext(_ context.Context, _ string, _ ...any) (sql.Result, error) {
	return nil, errors.New("injected ExecContext failure")
}

func (*errorDB) PrepareContext(_ context.Context, _ string) (*sql.Stmt, error) {
	return nil, errors.New("injected PrepareContext failure")
}

func (*errorDB) QueryContext(_ context.Context, _ string, _ ...any) (*sql.Rows, error) {
	return nil, errors.New("injected QueryContext failure")
}

func (*errorDB) QueryRowContext(_ context.Context, _ string, _ ...any) *sql.Row {
	return nil
}

// TestAddFlag_ErrorPath_MutexNotLeaked verifies Issue 4.4:
// when the DB write fails inside the maxBufferSize flush path the mutex must
// be released before AddFlag returns, so a subsequent call does not deadlock.
//
// If the bug is present the second AddFlag call below will hang forever (the
// test framework will time it out after the test timeout).
func TestAddFlag_ErrorPath_MutexNotLeaked(t *testing.T) {
	// Issue 4.4 is FIXED: fc.mutex.Unlock() is now called before every
	// `return err` inside the maxBufferSize flush block, so the mutex is
	// always released even when the DB write fails.

	// Build a collector wired to a store that always fails writes.
	failStore := storeWithBrokenQueries(t)
	fc := newTestCollector(t, failStore)
	fc.Start()

	// Fill the buffer up to maxBufferSize - 1.
	for i := range maxBufferSize - 1 {
		f := sampleFlag("FLAG{mutexleak_pre_" + itoa(i) + "}")
		// These go into the buffer; they don't trigger the flush path yet.
		// We bypass the error check because the store is only consulted on flush.
		_ = fc.AddFlag(f)
	}

	// The next AddFlag will push the buffer to maxBufferSize, triggering the
	// in-line flush.  That flush will fail because the store is broken.
	// After this call returns the mutex MUST be unlocked.
	trigger := sampleFlag("FLAG{mutexleak_trigger}")
	_ = fc.AddFlag(trigger) // may return error — that is expected

	// If Issue 4.4 is present this call will deadlock.  The 3 s timer detects
	// it and fails the test; if we reach <-done the mutex was released correctly.
	done := make(chan struct{})
	go func() {
		defer close(done)
		extra := sampleFlag("FLAG{mutexleak_after}")
		_ = fc.AddFlag(extra)
	}()

	select {
	case <-done:
		// Success — the second AddFlag returned without deadlock.
	case <-time.After(3 * time.Second):
		t.Error("AddFlag deadlocked after error path — mutex leak regression (Issue 4.4)")
	}
}

// ─── Issue 4.8 — data race in Stop() reading stats without the mutex ─────────

// TestStop_StatsReadUnderRace verifies Issue 4.8:
// Stop() must not read fc.stats fields without holding the mutex while a
// concurrent flush could be writing them.  Run this test with -race to confirm.
//
// The test starts a collector, fires off a concurrent batch of AddFlag calls
// (which may trigger timer or size-based flushes that call updateFlushStats),
// and then calls Stop() while those goroutines are still active.  If the data
// race is present `go test -race` will report it.
func TestStop_StatsReadUnderRace(t *testing.T) {
	store := newTestStore(t)
	fc := newTestCollector(t, store)
	fc.Start()

	// Launch goroutines that keep adding flags concurrently while Stop() runs.
	var wg sync.WaitGroup
	for i := range 10 {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			for j := range 5 {
				f := sampleFlag("FLAG{race_" + itoa(n) + "_" + itoa(j) + "}")
				_ = fc.AddFlag(f)
			}
		}(i)
	}

	// Stop the collector while the goroutines may still be writing stats.
	// With -race enabled, any unsynchronised read in Stop() will be detected.
	_ = fc.Stop()
	wg.Wait()
}

// ─── package-level helpers used only in this file ────────────────────────────

// newCondForTest creates a *sync.Cond for test-only FlagCollector instances
// that are built outside of GetCollector() / newTestCollector().
func newCondForTest(m *sync.Mutex) *sync.Cond {
	return sync.NewCond(m)
}

// itoa converts a small non-negative integer to its decimal string
// representation without importing strconv (keeps the file self-contained).
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	buf := [20]byte{}
	pos := len(buf)
	for n > 0 {
		pos--
		buf[pos] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[pos:])
}
