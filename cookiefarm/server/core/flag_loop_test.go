package core

import (
	"context"
	"models"
	"protocols"
	"sync"
	"testing"
	"time"

	"server/config"
	"server/database"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- UpdateFlags --------------------------------------------------------------

func TestUpdateFlags_NilSlice_DoesNotPanic(t *testing.T) {
	t.Parallel()
	r := newTestRunner(t)
	assert.NotPanics(t, func() {
		r.UpdateFlags(nil)
	})
}

func TestUpdateFlags_EmptySlice_DoesNotPanic(t *testing.T) {
	t.Parallel()
	r := newTestRunner(t)
	assert.NotPanics(t, func() {
		r.UpdateFlags([]protocols.ResponseProtocol{})
	})
}

func TestUpdateFlags_EmptySlice_NoDBWrites(t *testing.T) {
	t.Parallel()
	store := newTestStore(t)
	r := NewRunner(store)

	f := sampleFlag("FLAG{no_write}")
	insertFlag(t, store, f)

	r.UpdateFlags([]protocols.ResponseProtocol{})

	got := mustGetFlag(t, store, f.FlagCode)
	assert.Equal(t, "UNSUBMITTED", got.Status, "status must be unchanged after empty UpdateFlags")
}

func TestUpdateFlags_AllAccepted_AllUpdatedInDB(t *testing.T) {
	t.Parallel()
	store := newTestStore(t)
	r := NewRunner(store)

	codes := []string{"FLAG{acc_001}", "FLAG{acc_002}", "FLAG{acc_003}"}
	for _, c := range codes {
		insertFlag(t, store, sampleFlag(c))
	}

	responses := make([]protocols.ResponseProtocol, len(codes))
	for i, c := range codes {
		responses[i] = protocols.ResponseProtocol{Flag: c, Status: models.StatusAccepted, Msg: "accepted"}
	}

	r.UpdateFlags(responses)

	for _, c := range codes {
		got := mustGetFlag(t, store, c)
		assert.Equal(t, models.StatusAccepted, got.Status, "flag %q should be ACCEPTED", c)
	}
}

func TestUpdateFlags_AllDenied_AllUpdatedInDB(t *testing.T) {
	t.Parallel()
	store := newTestStore(t)
	r := NewRunner(store)

	codes := []string{"FLAG{den_001}", "FLAG{den_002}"}
	for _, c := range codes {
		insertFlag(t, store, sampleFlag(c))
	}

	responses := []protocols.ResponseProtocol{
		{Flag: "FLAG{den_001}", Status: models.StatusDenied, Msg: "already submitted"},
		{Flag: "FLAG{den_002}", Status: models.StatusDenied, Msg: "already submitted"},
	}

	r.UpdateFlags(responses)

	for _, c := range codes {
		got := mustGetFlag(t, store, c)
		assert.Equal(t, models.StatusDenied, got.Status)
	}
}

func TestUpdateFlags_AllError_AllUpdatedInDB(t *testing.T) {
	t.Parallel()
	store := newTestStore(t)
	r := NewRunner(store)

	codes := []string{"FLAG{err_001}", "FLAG{err_002}"}
	for _, c := range codes {
		insertFlag(t, store, sampleFlag(c))
	}

	responses := []protocols.ResponseProtocol{
		{Flag: "FLAG{err_001}", Status: models.StatusError, Msg: "checker error"},
		{Flag: "FLAG{err_002}", Status: models.StatusError, Msg: "checker error"},
	}

	r.UpdateFlags(responses)

	for _, c := range codes {
		got := mustGetFlag(t, store, c)
		assert.Equal(t, models.StatusError, got.Status)
	}
}

func TestUpdateFlags_MixedStatuses_EachUpdatedCorrectly(t *testing.T) {
	t.Parallel()
	store := newTestStore(t)
	r := NewRunner(store)

	flags := []database.Flag{
		sampleFlag("FLAG{mix_acc}"),
		sampleFlag("FLAG{mix_den}"),
		sampleFlag("FLAG{mix_err}"),
	}
	insertFlags(t, store, flags)

	responses := []protocols.ResponseProtocol{
		{Flag: "FLAG{mix_acc}", Status: models.StatusAccepted, Msg: "ok"},
		{Flag: "FLAG{mix_den}", Status: models.StatusDenied, Msg: "dup"},
		{Flag: "FLAG{mix_err}", Status: models.StatusError, Msg: "fail"},
	}

	r.UpdateFlags(responses)

	assert.Equal(t, models.StatusAccepted, mustGetFlag(t, store, "FLAG{mix_acc}").Status)
	assert.Equal(t, models.StatusDenied, mustGetFlag(t, store, "FLAG{mix_den}").Status)
	assert.Equal(t, models.StatusError, mustGetFlag(t, store, "FLAG{mix_err}").Status)
}

func TestUpdateFlags_UnknownStatus_FilteredOut_DBUnchanged(t *testing.T) {
	t.Parallel()
	store := newTestStore(t)
	r := NewRunner(store)

	f := sampleFlag("FLAG{unknown_status}")
	f.Status = "UNSUBMITTED"
	insertFlag(t, store, f)

	r.UpdateFlags([]protocols.ResponseProtocol{
		{Flag: "FLAG{unknown_status}", Status: "TOTALLY_MADE_UP", Msg: "?"},
	})

	got := mustGetFlag(t, store, f.FlagCode)
	assert.Equal(t, "UNSUBMITTED", got.Status, "unknown status must not mutate the DB row")
}

func TestUpdateFlags_UnknownStatus_MixedWithValid_ValidAreStillUpdated(t *testing.T) {
	t.Parallel()
	store := newTestStore(t)
	r := NewRunner(store)

	insertFlag(t, store, sampleFlag("FLAG{valid_one}"))
	insertFlag(t, store, sampleFlag("FLAG{bad_status}"))

	r.UpdateFlags([]protocols.ResponseProtocol{
		{Flag: "FLAG{valid_one}", Status: models.StatusAccepted, Msg: "ok"},
		{Flag: "FLAG{bad_status}", Status: "GARBAGE", Msg: "?"},
	})

	assert.Equal(t, models.StatusAccepted, mustGetFlag(t, store, "FLAG{valid_one}").Status)
	assert.Equal(t, "UNSUBMITTED", mustGetFlag(t, store, "FLAG{bad_status}").Status)
}

func TestUpdateFlags_ResponseMsgIsPersisted(t *testing.T) {
	t.Parallel()
	store := newTestStore(t)
	r := NewRunner(store)

	insertFlag(t, store, sampleFlag("FLAG{msg_check}"))

	r.UpdateFlags([]protocols.ResponseProtocol{
		{Flag: "FLAG{msg_check}", Status: models.StatusAccepted, Msg: "well done!"},
	})

	got := mustGetFlag(t, store, "FLAG{msg_check}")
	assert.Equal(t, "well done!", got.Msg)
}

func TestUpdateFlags_ResponseTimeIsUpdated(t *testing.T) {
	t.Parallel()
	store := newTestStore(t)
	r := NewRunner(store)

	f := sampleFlag("FLAG{rt_update}")
	f.ResponseTime = 0
	insertFlag(t, store, f)

	before := uint64(time.Now().Unix())
	r.UpdateFlags([]protocols.ResponseProtocol{
		{Flag: "FLAG{rt_update}", Status: models.StatusAccepted, Msg: "ok"},
	})
	after := uint64(time.Now().Unix())

	got := mustGetFlag(t, store, "FLAG{rt_update}")
	assert.GreaterOrEqual(t, got.ResponseTime, before)
	assert.LessOrEqual(t, got.ResponseTime, after+1)
}

func TestUpdateFlags_FlagNotInDB_DoesNotPanic(t *testing.T) {
	t.Parallel()
	store := newTestStore(t)
	r := NewRunner(store)

	// No flag inserted — the update is a no-op (UPDATE WHERE flag_code = ?
	// matches 0 rows) and must not return an error or panic.
	assert.NotPanics(t, func() {
		r.UpdateFlags([]protocols.ResponseProtocol{
			{Flag: "FLAG{ghost}", Status: models.StatusAccepted, Msg: "ok"},
		})
	})
}

func TestUpdateFlags_LargeBatch_AllUpdated(t *testing.T) {
	t.Parallel()
	const n = 500
	store := newTestStore(t)
	r := NewRunner(store)

	responses := make([]protocols.ResponseProtocol, n)
	for i := range n {
		code := "FLAG{bulk_" + itoa(i) + "}"
		insertFlag(t, store, sampleFlag(code))
		responses[i] = protocols.ResponseProtocol{Flag: code, Status: models.StatusAccepted, Msg: "ok"}
	}

	r.UpdateFlags(responses)

	assert.Equal(t, n, countFlagsWithStatus(t, store, models.StatusAccepted))
}

func TestUpdateFlags_AllThreeStatuses_CountsAreCorrect(t *testing.T) {
	t.Parallel()
	store := newTestStore(t)
	r := NewRunner(store)

	insertFlag(t, store, sampleFlag("FLAG{cnt_acc}"))
	insertFlag(t, store, sampleFlag("FLAG{cnt_den}"))
	insertFlag(t, store, sampleFlag("FLAG{cnt_err}"))

	r.UpdateFlags([]protocols.ResponseProtocol{
		{Flag: "FLAG{cnt_acc}", Status: models.StatusAccepted},
		{Flag: "FLAG{cnt_den}", Status: models.StatusDenied},
		{Flag: "FLAG{cnt_err}", Status: models.StatusError},
	})

	assert.Equal(t, 1, countFlagsWithStatus(t, store, models.StatusAccepted))
	assert.Equal(t, 1, countFlagsWithStatus(t, store, models.StatusDenied))
	assert.Equal(t, 1, countFlagsWithStatus(t, store, models.StatusError))
}

func TestUpdateFlags_OnlyUnknownStatuses_NothingUpdated(t *testing.T) {
	t.Parallel()
	store := newTestStore(t)
	r := NewRunner(store)

	codes := []string{"FLAG{unk_a}", "FLAG{unk_b}"}
	for _, c := range codes {
		insertFlag(t, store, sampleFlag(c))
	}

	r.UpdateFlags([]protocols.ResponseProtocol{
		{Flag: "FLAG{unk_a}", Status: "INVALID_A"},
		{Flag: "FLAG{unk_b}", Status: "INVALID_B"},
	})

	for _, c := range codes {
		got := mustGetFlag(t, store, c)
		assert.Equal(t, "UNSUBMITTED", got.Status)
	}
}

func TestUpdateFlags_ConcurrentCalls_NoPanic(t *testing.T) {
	t.Parallel()
	store := newTestStore(t)
	r := NewRunner(store)

	const goroutines = 10
	const flagsPerGoroutine = 20

	// Pre-insert all flags.
	for g := range goroutines {
		for i := range flagsPerGoroutine {
			code := "FLAG{conc_" + itoa(g*flagsPerGoroutine+i) + "}"
			insertFlag(t, store, sampleFlag(code))
		}
	}

	var wg sync.WaitGroup
	for g := range goroutines {
		wg.Add(1)
		go func(g int) {
			defer wg.Done()
			responses := make([]protocols.ResponseProtocol, flagsPerGoroutine)
			for i := range flagsPerGoroutine {
				code := "FLAG{conc_" + itoa(g*flagsPerGoroutine+i) + "}"
				responses[i] = protocols.ResponseProtocol{
					Flag:   code,
					Status: models.StatusAccepted,
					Msg:    "ok",
				}
			}
			r.UpdateFlags(responses)
		}(g)
	}
	assert.NotPanics(t, func() { wg.Wait() })
}

// --- StartFlagProcessingLoop --------------------------------------------------

// setupProcessingLoop wires config.Submit to the provided stub and returns a
// cancel func for the context passed to StartFlagProcessingLoop.  The caller is
// responsible for calling cancel when the test is done.
func setupProcessingLoop(t *testing.T, submitFn func(string, string, []string) ([]protocols.ResponseProtocol, error)) (store *database.Store, r *Runner, cancel context.CancelFunc) {
	t.Helper()
	store = newTestStore(t)
	r = NewRunner(store)

	// Wire the fake submit function into global config.
	config.Submit = submitFn
	config.SharedConfig.ConfigServer.MaxFlagBatchSize = 10

	// Very short tick so the loop fires quickly in tests.
	origInterval := config.SharedConfig.ConfigServer.SubmitFlagCheckerTime
	config.SharedConfig.ConfigServer.SubmitFlagCheckerTime = 1
	t.Cleanup(func() {
		config.SharedConfig.ConfigServer.SubmitFlagCheckerTime = origInterval
		config.Submit = nil
	})

	_, cancelFn := context.WithCancel(context.Background())
	t.Cleanup(cancelFn)
	return store, r, cancelFn
}

func TestStartFlagProcessingLoop_ContextCancelled_LoopExits(t *testing.T) {
	_, r, cancel := setupProcessingLoop(t, fakeSubmit(models.StatusAccepted, nil, nil))
	defer cancel()

	ctx, cancelLoop := context.WithCancel(context.Background())

	loopDone := make(chan struct{})
	go func() {
		defer close(loopDone)
		r.StartFlagProcessingLoop(ctx)
	}()

	// Give the loop a moment to start, then cancel.
	time.Sleep(30 * time.Millisecond)
	cancelLoop()

	select {
	case <-loopDone:
		// expected: loop exited cleanly after context cancellation
	case <-time.After(2 * time.Second):
		t.Fatal("StartFlagProcessingLoop did not exit after context cancellation")
	}
}

func TestStartFlagProcessingLoop_NoUnsubmittedFlags_SubmitNotCalled(t *testing.T) {
	var mu sync.Mutex
	var calls [][]string

	store, r, cancel := setupProcessingLoop(t, fakeSubmit(models.StatusAccepted, &mu, &calls))
	_ = store

	ctx, cancelLoop := context.WithCancel(context.Background())
	go r.StartFlagProcessingLoop(ctx)

	// Let the loop tick at least once then cancel.
	time.Sleep(150 * time.Millisecond)
	cancelLoop()
	cancel()

	mu.Lock()
	defer mu.Unlock()
	assert.Empty(t, calls, "submit must not be called when there are no UNSUBMITTED flags")
}

func TestStartFlagProcessingLoop_UnsubmittedFlags_SubmitIsCalled(t *testing.T) {
	var mu sync.Mutex
	var calls [][]string

	store, r, cancel := setupProcessingLoop(t, fakeSubmit(models.StatusAccepted, &mu, &calls))
	defer cancel()
	// Insert a flag in UNSUBMITTED state.
	f := sampleFlag("FLAG{loop_submit_001}")
	f.Status = "UNSUBMITTED"
	insertFlag(t, store, f)

	ctx, cancelLoop := context.WithCancel(context.Background())
	go r.StartFlagProcessingLoop(ctx)

	// Wait until submit has been called at least once.
	waitFor(t, 2*time.Second, 20*time.Millisecond, "submit was never called", func() bool {
		mu.Lock()
		defer mu.Unlock()
		return len(calls) > 0
	})

	cancelLoop()
}

func TestStartFlagProcessingLoop_UnsubmittedFlag_UpdatedToAccepted(t *testing.T) {
	store, r, cancel := setupProcessingLoop(t, fakeSubmit(models.StatusAccepted, nil, nil))
	defer cancel()

	f := sampleFlag("FLAG{loop_accepted}")
	f.Status = "UNSUBMITTED"
	insertFlag(t, store, f)

	ctx, cancelLoop := context.WithCancel(context.Background())
	go r.StartFlagProcessingLoop(ctx)
	defer cancelLoop()

	waitForFlagStatus(t, store, "FLAG{loop_accepted}", models.StatusAccepted, 2*time.Second)
}

func TestStartFlagProcessingLoop_UnsubmittedFlag_UpdatedToDenied(t *testing.T) {
	store, r, cancel := setupProcessingLoop(t, fakeSubmit(models.StatusDenied, nil, nil))
	defer cancel()

	f := sampleFlag("FLAG{loop_denied}")
	f.Status = "UNSUBMITTED"
	insertFlag(t, store, f)

	ctx, cancelLoop := context.WithCancel(context.Background())
	go r.StartFlagProcessingLoop(ctx)
	defer cancelLoop()

	waitForFlagStatus(t, store, "FLAG{loop_denied}", models.StatusDenied, 2*time.Second)
}

func TestStartFlagProcessingLoop_MultipleBatches_AllFlagsEventuallySubmitted(t *testing.T) {
	// Limit batch size to 2 so we exercise multi-batch behaviour.
	origBatch := config.SharedConfig.ConfigServer.MaxFlagBatchSize
	config.SharedConfig.ConfigServer.MaxFlagBatchSize = 2
	t.Cleanup(func() { config.SharedConfig.ConfigServer.MaxFlagBatchSize = origBatch })

	store, r, cancel := setupProcessingLoop(t, fakeSubmit(models.StatusAccepted, nil, nil))
	defer cancel()

	flags := []database.Flag{
		sampleFlag("FLAG{batch_001}"),
		sampleFlag("FLAG{batch_002}"),
		sampleFlag("FLAG{batch_003}"),
		sampleFlag("FLAG{batch_004}"),
	}
	for i := range flags {
		flags[i].Status = "UNSUBMITTED"
	}
	insertFlags(t, store, flags)

	ctx, cancelLoop := context.WithCancel(context.Background())
	go r.StartFlagProcessingLoop(ctx)
	defer cancelLoop()

	for _, f := range flags {
		waitForFlagStatus(t, store, f.FlagCode, models.StatusAccepted, 3*time.Second)
	}
}

func TestStartFlagProcessingLoop_SubmitError_LoopContinues(t *testing.T) {
	// submit always errors.
	store, r, cancel := setupProcessingLoop(t, errorSubmit(errSubmit))
	defer cancel()

	f := sampleFlag("FLAG{submit_err}")
	f.Status = "UNSUBMITTED"
	insertFlag(t, store, f)

	ctx, cancelLoop := context.WithCancel(context.Background())
	go r.StartFlagProcessingLoop(ctx)

	// Give the loop two ticks to attempt submission.
	time.Sleep(200 * time.Millisecond)
	cancelLoop()

	// The flag must still be UNSUBMITTED (submit failed, no update).
	got := mustGetFlag(t, store, "FLAG{submit_err}")
	assert.Equal(t, "UNSUBMITTED", got.Status)
}

func TestStartFlagProcessingLoop_SubmitFuncNil_LoopDoesNotPanic(t *testing.T) {
	// config.Submit is nil — this simulates no protocol loaded.
	origInterval := config.SharedConfig.ConfigServer.SubmitFlagCheckerTime
	config.SharedConfig.ConfigServer.SubmitFlagCheckerTime = 9999
	t.Cleanup(func() { config.SharedConfig.ConfigServer.SubmitFlagCheckerTime = origInterval })

	config.Submit = nil

	store := newTestStore(t)
	r := NewRunner(store)

	ctx, cancelLoop := context.WithCancel(context.Background())
	t.Cleanup(cancelLoop)

	// The loop should return immediately after failing to load the protocol.
	done := make(chan struct{})
	go func() {
		defer close(done)
		r.StartFlagProcessingLoop(ctx)
	}()

	select {
	case <-done:
		// expected: loop exited because protocol loading failed
	case <-time.After(2 * time.Second):
		cancelLoop()
		t.Fatal("StartFlagProcessingLoop hung when config.Submit is nil")
	}
}

func TestStartFlagProcessingLoop_MaxBatchSizeRespected(t *testing.T) {
	const batchSize = 3

	var mu sync.Mutex
	var calls [][]string

	store, r, cancel := setupProcessingLoop(t, fakeSubmit(models.StatusAccepted, &mu, &calls))
	defer cancel()

	origBatch := config.SharedConfig.ConfigServer.MaxFlagBatchSize
	config.SharedConfig.ConfigServer.MaxFlagBatchSize = batchSize
	t.Cleanup(func() { config.SharedConfig.ConfigServer.MaxFlagBatchSize = origBatch })

	// Insert more flags than the batch size.
	for i := range batchSize + 5 {
		f := sampleFlag("FLAG{maxbatch_" + itoa(i) + "}")
		f.Status = "UNSUBMITTED"
		insertFlag(t, store, f)
	}

	ctx, cancelLoop := context.WithCancel(context.Background())
	go r.StartFlagProcessingLoop(ctx)

	// Wait until the first batch has been submitted.
	waitFor(t, 2*time.Second, 20*time.Millisecond, "first batch never submitted", func() bool {
		mu.Lock()
		defer mu.Unlock()
		return len(calls) > 0
	})

	cancelLoop()

	mu.Lock()
	defer mu.Unlock()
	require.NotEmpty(t, calls)
	assert.LessOrEqual(t, len(calls[0]), batchSize, "first call must not exceed MaxFlagBatchSize")
}

func TestStartFlagProcessingLoop_AlreadyAcceptedFlags_NotResubmitted(t *testing.T) {
	var mu sync.Mutex
	var calls [][]string

	store, r, cancel := setupProcessingLoop(t, fakeSubmit(models.StatusAccepted, &mu, &calls))
	defer cancel()

	// Insert a flag that is already ACCEPTED — should never be picked up.
	f := sampleFlag("FLAG{already_accepted}")
	f.Status = models.StatusAccepted
	insertFlag(t, store, f)

	ctx, cancelLoop := context.WithCancel(context.Background())
	go r.StartFlagProcessingLoop(ctx)

	time.Sleep(150 * time.Millisecond)
	cancelLoop()

	mu.Lock()
	defer mu.Unlock()
	assert.Empty(t, calls, "ACCEPTED flags must not be re-submitted")
}

// --- helpers used only in this file ------------------------------------------

// itoa converts an int to its decimal string representation without importing
// strconv (keeps test dependencies minimal).
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := false
	if n < 0 {
		neg = true
		n = -n
	}
	buf := [20]byte{}
	pos := len(buf)
	for n > 0 {
		pos--
		buf[pos] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		pos--
		buf[pos] = '-'
	}
	return string(buf[pos:])
}
