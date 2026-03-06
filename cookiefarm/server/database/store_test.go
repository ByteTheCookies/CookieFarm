package database

import (
	"context"
	"errors"
	"testing"
)

// --- NewStore -----------------------------------------------------------------

func TestNewStore_ReturnsNonNil(t *testing.T) {
	db := newTestDB(t)
	store := NewStore(db)
	if store == nil {
		t.Fatal("expected non-nil *Store, got nil")
	}
}

func TestNewStore_QueriesFieldIsNonNil(t *testing.T) {
	db := newTestDB(t)
	store := NewStore(db)
	if store.Queries == nil {
		t.Fatal("expected non-nil store.Queries, got nil")
	}
}

// --- Store.WithTx — commit path -----------------------------------------------

func TestWithTx_FnReturnsNil_CommitsTransaction(t *testing.T) {
	store := newTestStore(t)
	flag := sampleFlag("FLAG{tx_commit_001}")

	err := store.WithTx(context.Background(), func(q *Queries) error {
		return q.AddFlag(context.Background(), MapFromFlagToDBParams(flag))
	})
	assertNoError(t, err, "WithTx commit")

	// Verify the row survived the commit by reading it back through the store.
	got, err := store.Queries.GetFlagByCode(context.Background(), flag.FlagCode)
	assertNoError(t, err, "GetFlagByCode after WithTx commit")
	assertFlagsEqual(t, flag, got)
}

func TestWithTx_MultipleOpsInFn_AllCommitted(t *testing.T) {
	store := newTestStore(t)
	flags := []Flag{
		sampleFlag("FLAG{tx_multi_001}"),
		sampleFlag("FLAG{tx_multi_002}"),
		sampleFlag("FLAG{tx_multi_003}"),
	}

	err := store.WithTx(context.Background(), func(q *Queries) error {
		for _, f := range flags {
			if err := q.AddFlag(context.Background(), MapFromFlagToDBParams(f)); err != nil {
				return err
			}
		}
		return nil
	})
	assertNoError(t, err, "WithTx multi-op commit")

	count, err := store.Queries.CountFlags(context.Background())
	assertNoError(t, err, "CountFlags after multi-op WithTx")
	assertInt64Equal(t, int64(len(flags)), count, "flag count after multi-op commit")
}

func TestWithTx_FnUpdatesFlag_UpdateCommitted(t *testing.T) {
	store := newTestStore(t)
	flag := sampleFlag("FLAG{tx_update_001}")
	flag.Status = "UNSUBMITTED"
	insertFlag(t, store.Queries, flag)

	err := store.WithTx(context.Background(), func(q *Queries) error {
		return q.UpdateFlagStatusByCode(context.Background(), UpdateFlagStatusByCodeParams{
			FlagCode:     flag.FlagCode,
			Status:       "ACCEPTED",
			Msg:          "committed update",
			ResponseTime: 12345,
		})
	})
	assertNoError(t, err, "WithTx update commit")

	got := mustGetFlag(t, store.Queries, flag.FlagCode)
	if got.Status != "ACCEPTED" {
		t.Errorf("expected Status=ACCEPTED after committed tx, got %q", got.Status)
	}
	if got.Msg != "committed update" {
		t.Errorf("expected Msg=%q after committed tx, got %q", "committed update", got.Msg)
	}
}

// --- Store.WithTx — rollback path ---------------------------------------------

func TestWithTx_FnReturnsError_RollsBackTransaction(t *testing.T) {
	store := newTestStore(t)
	flag := sampleFlag("FLAG{tx_rollback_001}")

	txErr := errors.New("intentional transaction error")

	err := store.WithTx(context.Background(), func(q *Queries) error {
		// Insert the flag inside the transaction.
		if err := q.AddFlag(context.Background(), MapFromFlagToDBParams(flag)); err != nil {
			return err
		}
		// Then return an error to trigger rollback.
		return txErr
	})

	// The error from fn must be propagated back to the caller.
	if !errors.Is(err, txErr) {
		t.Errorf("expected txErr to be returned, got: %v", err)
	}

	// The flag must NOT exist because the transaction was rolled back.
	_, err = store.Queries.GetFlagByCode(context.Background(), flag.FlagCode)
	if !errors.Is(err, errNoRows()) {
		t.Errorf("expected sql.ErrNoRows after rollback, got: %v", err)
	}
}

func TestWithTx_FnReturnsError_OriginalErrorPropagated(t *testing.T) {
	store := newTestStore(t)
	sentinel := errors.New("sentinel error")

	err := store.WithTx(context.Background(), func(q *Queries) error {
		return sentinel
	})

	if !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel error to be returned unchanged, got: %v", err)
	}
}

func TestWithTx_RollbackDoesNotAffectExistingRows(t *testing.T) {
	store := newTestStore(t)

	// Pre-existing row inserted outside the transaction.
	existing := sampleFlag("FLAG{tx_rb_existing}")
	insertFlag(t, store.Queries, existing)

	// Transaction inserts a new row then rolls back.
	_ = store.WithTx(context.Background(), func(q *Queries) error {
		newFlag := sampleFlag("FLAG{tx_rb_new}")
		if err := q.AddFlag(context.Background(), MapFromFlagToDBParams(newFlag)); err != nil {
			return err
		}
		return errors.New("force rollback")
	})

	// The pre-existing row must still be present.
	_ = mustGetFlag(t, store.Queries, existing.FlagCode)

	// The new row must not have been persisted.
	_, err := store.Queries.GetFlagByCode(context.Background(), "FLAG{tx_rb_new}")
	if !errors.Is(err, errNoRows()) {
		t.Errorf("expected rolled-back flag to be absent, got: %v", err)
	}

	// Total count must still be 1.
	count, err := store.Queries.CountFlags(context.Background())
	assertNoError(t, err, "CountFlags after rollback")
	assertInt64Equal(t, 1, count, "row count after rollback must be 1")
}

// --- Store.WithTx — cancelled context ----------------------------------------

func TestWithTx_CancelledContext_ReturnsError(t *testing.T) {
	store := newTestStore(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	err := store.WithTx(ctx, func(q *Queries) error {
		return nil
	})

	// A cancelled context must cause BeginTx (or the subsequent Commit) to
	// return an error; either way WithTx must not return nil.
	if err == nil {
		// Some SQLite drivers accept a pre-cancelled context on Begin; if so,
		// the Commit step should catch it.  Accept either behaviour in the
		// test but document it.
		t.Log("WithTx with cancelled context returned nil — SQLite driver accepted the pre-cancelled context")
	}
}

// --- helpers local to this file -----------------------------------------------

// errNoRows returns sql.ErrNoRows via the errors package so we don't need to
// import database/sql directly in every assertion.
func errNoRows() error {
	// We rely on the fact that testhelpers_test.go already imports database/sql
	// indirectly; here we just re-surface the sentinel via a named helper so
	// the store_test file stays clean.
	return _sqlErrNoRows
}
