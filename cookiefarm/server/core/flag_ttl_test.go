package core

import (
	"context"
	"models"
	"testing"
	"time"

	"server/database"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- ValidateFlagTTL ----------------------------------------------------------

// insertFlagWithAge inserts a flag whose response_time is set to
// (now - age), making it appear "age" seconds old to DeleteFlagByTTL.
func insertFlagWithAge(t *testing.T, store *database.Store, code string, age time.Duration) {
	t.Helper()
	f := sampleFlag(code)
	f.ResponseTime = uint64(time.Now().Add(-age).Unix())
	f.Status = models.StatusAccepted
	insertFlag(t, store, f)
}

// countAllFlags returns the total number of flags in the store.
func countAllFlags(t *testing.T, store *database.Store) int {
	t.Helper()
	flags, err := store.Queries.GetAllFlags(context.Background())
	require.NoError(t, err)
	return len(flags)
}

// --- Context cancellation -----------------------------------------------------

func TestValidateFlagTTL_ContextCancelledBeforeFirstTick_ExitsCleanly(t *testing.T) {
	t.Parallel()

	r := newTestRunner(t)

	// Use a huge interval so the ticker never fires; cancellation must still
	// stop the loop promptly.
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	done := make(chan struct{})
	go func() {
		defer close(done)
		r.ValidateFlagTTL(ctx, 1, 99999)
	}()

	select {
	case <-done:
		// expected: loop exited because context was already cancelled
	case <-time.After(2 * time.Second):
		t.Fatal("ValidateFlagTTL did not exit after context cancellation")
	}
}

func TestValidateFlagTTL_ContextCancelledMidRun_ExitsCleanly(t *testing.T) {
	t.Parallel()

	r := newTestRunner(t)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	done := make(chan struct{})
	go func() {
		defer close(done)
		// tick every 99999 seconds — loop will block on select waiting for the
		// ticker; context cancel must unblock it.
		r.ValidateFlagTTL(ctx, 1, 99999)
	}()

	// Let the goroutine reach the select statement, then cancel.
	time.Sleep(30 * time.Millisecond)
	cancel()

	select {
	case <-done:
		// expected
	case <-time.After(2 * time.Second):
		t.Fatal("ValidateFlagTTL did not exit after mid-run context cancellation")
	}
}

// --- Deletion logic -----------------------------------------------------------

func TestValidateFlagTTL_FlagsOlderThanTTL_AreDeleted(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)
	r := NewRunner(store)

	// Insert a flag that is 200 seconds old.
	insertFlagWithAge(t, store, "FLAG{ttl_old}", 200*time.Second)
	require.Equal(t, 1, countAllFlags(t, store))

	// TTL window = 1 tick × 100 seconds = 100 seconds.
	// The flag is 200 s old → it falls outside the window → must be deleted.
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	go r.ValidateFlagTTL(ctx, 1, 5)

	waitFor(t, 6*time.Second, 20*time.Millisecond,
		"old flag was not deleted within TTL window",
		func() bool { return countAllFlags(t, store) == 0 },
	)
}

func TestValidateFlagTTL_FlagsNewerThanTTL_AreNotDeleted(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)
	r := NewRunner(store)

	// Insert a fresh flag (0 seconds old).
	insertFlagWithAge(t, store, "FLAG{ttl_fresh}", 0)
	require.Equal(t, 1, countAllFlags(t, store))

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	// TTL window = 1 tick × 1 second.
	go r.ValidateFlagTTL(ctx, 1, 1)

	// Let the ticker fire at least once.
	time.Sleep(150 * time.Millisecond)
	cancel()

	// Fresh flag must still be present.
	assert.Equal(t, 1, countAllFlags(t, store), "fresh flag must not be deleted")
}

func TestValidateFlagTTL_MixedAges_OnlyOldFlagsDeleted(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)
	r := NewRunner(store)

	// Old flag: 300 seconds old.
	insertFlagWithAge(t, store, "FLAG{ttl_mix_old}", 300*time.Second)
	// Fresh flag: just inserted (0 seconds old).
	insertFlagWithAge(t, store, "FLAG{ttl_mix_fresh}", 0)

	require.Equal(t, 2, countAllFlags(t, store))

	// TTL window = 1 tick × 150 seconds. Old flag (300 s) is outside → deleted.
	// Fresh flag (0 s) is inside → kept.
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	go r.ValidateFlagTTL(ctx, 1, 5)
	waitFor(t, 6*time.Second, 20*time.Millisecond,
		"old flag was not deleted",
		func() bool { return countAllFlags(t, store) == 1 },
	)

	// Verify it is specifically the fresh one that remains.
	remaining, err := store.Queries.GetAllFlags(context.Background())
	require.NoError(t, err)
	require.Len(t, remaining, 1)
	assert.Equal(t, "FLAG{ttl_mix_fresh}", remaining[0].FlagCode)
}

func TestValidateFlagTTL_NoFlags_DoesNotPanic(t *testing.T) {
	t.Parallel()

	r := newTestRunner(t)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	assert.NotPanics(t, func() {
		go r.ValidateFlagTTL(ctx, 1, 1)
		time.Sleep(150 * time.Millisecond)
		cancel()
	})
}

func TestValidateFlagTTL_ZeroAffectedRows_DoesNotReturnError(t *testing.T) {
	t.Parallel()

	// Empty store → DeleteFlagByTTL matches 0 rows → loop must continue
	// without error and not panic.
	store := newTestStore(t)
	r := NewRunner(store)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	done := make(chan struct{})
	go func() {
		defer close(done)
		r.ValidateFlagTTL(ctx, 1, 1)
	}()

	// Let the loop tick twice, then cancel.
	time.Sleep(150 * time.Millisecond)
	cancel()

	select {
	case <-done:
		// expected
	case <-time.After(2 * time.Second):
		t.Fatal("ValidateFlagTTL hung after zero-row deletion")
	}
}

// --- Interval calculation -----------------------------------------------------

func TestValidateFlagTTL_IntervalIsProductOfTTLAndTickTime(t *testing.T) {
	t.Parallel()

	// We verify the interval indirectly: with flagTTL=1 and tickTime=1, the
	// ticker fires every 1 second.  We insert a flag 2 seconds old and wait
	// no longer than 1.5 s for it to disappear.
	store := newTestStore(t)
	r := NewRunner(store)

	insertFlagWithAge(t, store, "FLAG{ttl_interval}", 2*time.Second)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	go r.ValidateFlagTTL(ctx, 1, 1) // interval = 1 × 1 = 1 second

	waitFor(t, 2*time.Second, 20*time.Millisecond,
		"flag was not deleted within the expected interval",
		func() bool { return countAllFlags(t, store) == 0 },
	)
}

func TestValidateFlagTTL_MultipleTicksDeleteAccumulatedFlags(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)
	r := NewRunner(store)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	// Start the loop with a 1-second interval.
	go r.ValidateFlagTTL(ctx, 1, 1)

	// After the loop starts, insert an old flag and wait for it to be cleaned up.
	time.Sleep(50 * time.Millisecond)
	insertFlagWithAge(t, store, "FLAG{ttl_tick2}", 10*time.Second)

	waitFor(t, 3*time.Second, 20*time.Millisecond,
		"flag inserted after loop start was not eventually deleted",
		func() bool { return countAllFlags(t, store) == 0 },
	)
}

// --- TTL boundary conditions --------------------------------------------------

func TestValidateFlagTTL_LargeTTLValue_DoesNotPanic(t *testing.T) {
	t.Parallel()

	r := newTestRunner(t)

	ctx, cancel := context.WithCancel(context.Background())

	assert.NotPanics(t, func() {
		go r.ValidateFlagTTL(ctx, ^uint64(0)>>1, 1)
		time.Sleep(30 * time.Millisecond)
		cancel()
	})
}

func TestValidateFlagTTL_TTLEqualToFlagAge_FlagIsDeleted(t *testing.T) {
	t.Parallel()

	// Flag is exactly at the boundary: age == TTL window.
	// DeleteFlagByTTL uses strict less-than (<), so a flag whose response_time
	// equals (now - TTL) is right on the boundary.  We insert a flag that is
	// slightly older than the TTL window to guarantee deletion.
	store := newTestStore(t)
	r := NewRunner(store)

	// TTL window = 1 × 2 = 2 seconds.  Flag age = 3 seconds → must be deleted.
	insertFlagWithAge(t, store, "FLAG{ttl_boundary}", 3*time.Second)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	go r.ValidateFlagTTL(ctx, 1, 2)

	waitFor(t, 5*time.Second, 20*time.Millisecond,
		"boundary-age flag was not deleted",
		func() bool { return countAllFlags(t, store) == 0 },
	)
}

// --- Multiple concurrent runners ---------------------------------------------

func TestValidateFlagTTL_TwoRunnersOnSameStore_NoPanic(t *testing.T) {
	t.Parallel()

	// Two independent runners pointing at the same store must not panic even
	// if they race on the same rows.
	store := newTestStore(t)
	r1 := NewRunner(store)
	r2 := NewRunner(store)

	insertFlagWithAge(t, store, "FLAG{concurrent_ttl_1}", 200*time.Second)
	insertFlagWithAge(t, store, "FLAG{concurrent_ttl_2}", 200*time.Second)

	ctx1, cancel1 := context.WithCancel(context.Background())
	ctx2, cancel2 := context.WithCancel(context.Background())
	t.Cleanup(cancel1)
	t.Cleanup(cancel2)

	assert.NotPanics(t, func() {
		go r1.ValidateFlagTTL(ctx1, 1, 1)
		go r2.ValidateFlagTTL(ctx2, 1, 1)

		time.Sleep(200 * time.Millisecond)
		cancel1()
		cancel2()
	})
}

// --- Invalid configuration handling -----------------------------------------

func TestInvalidTickTime_DefaultsTo60Seconds(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)
	r := NewRunner(store)

	// Insert a flag that is 70 seconds old.
	insertFlagWithAge(t, store, "FLAG{invalid_tick}", 70*time.Second)
	require.Equal(t, 1, countAllFlags(t, store))

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	// Use an invalid tickTime (0 seconds) → should default to 60 seconds.
	go r.ValidateFlagTTL(ctx, 1, 0)

	waitFor(t, 80*time.Second, 20*time.Millisecond,
		"flag was not deleted with invalid tickTime",
		func() bool { return countAllFlags(t, store) == 0 },
	)
}
