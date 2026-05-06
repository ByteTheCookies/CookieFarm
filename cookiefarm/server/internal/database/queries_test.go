package database

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"
)

// --- AddFlag ------------------------------------------------------------------

func TestAddFlag_NewFlag_Inserted(t *testing.T) {
	q := newTestQueries(t)
	flag := sampleFlag("FLAG{unique_001}")

	err := q.AddFlag(context.Background(), MapFromFlagToDBParams(flag))
	assertNoError(t, err, "AddFlag new flag")

	got := mustGetFlag(t, q, flag.FlagCode)
	assertFlagsEqual(t, flag, got)
}

func TestAddFlag_DuplicateFlagCode_SilentlyIgnored(t *testing.T) {
	q := newTestQueries(t)
	flag := sampleFlag("FLAG{duplicate_001}")

	insertFlag(t, q, flag)

	// Second insert of the same PK must not return an error (INSERT OR IGNORE).
	err := q.AddFlag(context.Background(), MapFromFlagToDBParams(flag))
	assertNoError(t, err, "AddFlag duplicate flag code")

	// The row count must still be exactly 1.
	count, err := q.CountFlags(context.Background())
	assertNoError(t, err, "CountFlags after duplicate insert")
	assertInt64Equal(t, 1, count, "flag count after duplicate insert")
}

func TestAddFlag_EmptyFlagCode_Inserted(t *testing.T) {
	q := newTestQueries(t)
	flag := sampleFlag("")

	err := q.AddFlag(context.Background(), MapFromFlagToDBParams(flag))
	assertNoError(t, err, "AddFlag with empty flag code")
}

func TestAddFlag_ZeroNumericFields_Inserted(t *testing.T) {
	q := newTestQueries(t)
	flag := Flag{
		FlagCode:     "FLAG{zero_fields}",
		ServiceName:  "svc",
		PortService:  0,
		SubmitTime:   0,
		ResponseTime: 0,
		Msg:          "",
		Status:       0,
		TeamID:       0,
		Username:     "",
		ExploitName:  "",
	}

	err := q.AddFlag(context.Background(), MapFromFlagToDBParams(flag))
	assertNoError(t, err, "AddFlag with zero numeric fields")

	got := mustGetFlag(t, q, flag.FlagCode)
	assertFlagsEqual(t, flag, got)
}

func TestAddFlag_MultipleFlags_AllInserted(t *testing.T) {
	q := newTestQueries(t)
	flags := []Flag{
		sampleFlag("FLAG{multi_001}"),
		sampleFlag("FLAG{multi_002}"),
		sampleFlag("FLAG{multi_003}"),
	}

	insertFlags(t, q, flags)

	count, err := q.CountFlags(context.Background())
	assertNoError(t, err, "CountFlags after multiple inserts")
	assertInt64Equal(t, int64(len(flags)), count, "flag count after multiple inserts")
}

// --- GetFlagByCode ------------------------------------------------------------

func TestGetFlagByCode_Exists_ReturnsFlag(t *testing.T) {
	q := newTestQueries(t)
	flag := sampleFlag("FLAG{get_001}")
	insertFlag(t, q, flag)

	got, err := q.GetFlagByCode(context.Background(), flag.FlagCode)
	assertNoError(t, err, "GetFlagByCode existing")
	assertFlagsEqual(t, flag, got)
}

func TestGetFlagByCode_NotExists_ReturnsErrNoRows(t *testing.T) {
	q := newTestQueries(t)

	_, err := q.GetFlagByCode(context.Background(), "FLAG{nonexistent}")
	if !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("expected sql.ErrNoRows, got: %v", err)
	}
}

func TestGetFlagByCode_EmptyCode_ReturnsErrNoRows(t *testing.T) {
	q := newTestQueries(t)

	_, err := q.GetFlagByCode(context.Background(), "")
	if !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("expected sql.ErrNoRows for empty code, got: %v", err)
	}
}

// --- GetAllFlags --------------------------------------------------------------

func TestGetAllFlags_EmptyDB_ReturnsEmptySlice(t *testing.T) {
	q := newTestQueries(t)

	flags, err := q.GetAllFlags(context.Background())
	assertNoError(t, err, "GetAllFlags on empty DB")
	if len(flags) != 0 {
		t.Errorf("expected empty/nil slice, got len=%d", len(flags))
	}
}

func TestGetAllFlags_ReturnsAllRows(t *testing.T) {
	q := newTestQueries(t)
	inserted := []Flag{
		sampleFlag("FLAG{all_001}"),
		sampleFlag("FLAG{all_002}"),
		sampleFlag("FLAG{all_003}"),
	}
	insertFlags(t, q, inserted)

	got, err := q.GetAllFlags(context.Background())
	assertNoError(t, err, "GetAllFlags")
	assertFlagSliceLen(t, len(inserted), got, "GetAllFlags result")
}

func TestGetAllFlags_OrderedBySubmitTimeDesc(t *testing.T) {
	q := newTestQueries(t)

	now := uint64(time.Now().Unix())
	old := sampleFlag("FLAG{order_old}")
	old.SubmitTime = now - 100
	newer := sampleFlag("FLAG{order_new}")
	newer.SubmitTime = now

	insertFlags(t, q, []Flag{old, newer})

	got, err := q.GetAllFlags(context.Background())
	assertNoError(t, err, "GetAllFlags ordering")
	if len(got) < 2 {
		t.Fatalf("expected at least 2 flags, got %d", len(got))
	}
	if got[0].SubmitTime < got[1].SubmitTime {
		t.Errorf("expected descending SubmitTime order; got[0]=%d, got[1]=%d",
			got[0].SubmitTime, got[1].SubmitTime)
	}
}

// --- GetAllFlagCodes ----------------------------------------------------------

func TestGetAllFlagCodes_EmptyDB_ReturnsEmptySlice(t *testing.T) {
	q := newTestQueries(t)

	codes, err := q.GetAllFlagCodes(context.Background())
	assertNoError(t, err, "GetAllFlagCodes on empty DB")
	if len(codes) != 0 {
		t.Errorf("expected empty slice, got len=%d", len(codes))
	}
}

func TestGetAllFlagCodes_ReturnsOnlyCodes(t *testing.T) {
	q := newTestQueries(t)
	flags := []Flag{
		sampleFlag("FLAG{code_001}"),
		sampleFlag("FLAG{code_002}"),
	}
	insertFlags(t, q, flags)

	codes, err := q.GetAllFlagCodes(context.Background())
	assertNoError(t, err, "GetAllFlagCodes")
	assertStringSliceLen(t, len(flags), codes, "GetAllFlagCodes result")

	codeSet := map[string]bool{}
	for _, c := range codes {
		codeSet[c] = true
	}
	for _, f := range flags {
		if !codeSet[f.FlagCode] {
			t.Errorf("expected code %q in result, not found", f.FlagCode)
		}
	}
}

// --- GetFirstNFlags -----------------------------------------------------------

func TestGetFirstNFlags_LimitLessThanTotal_ReturnsLimitRows(t *testing.T) {
	q := newTestQueries(t)
	insertFlags(t, q, []Flag{
		sampleFlag("FLAG{n_001}"),
		sampleFlag("FLAG{n_002}"),
		sampleFlag("FLAG{n_003}"),
	})

	got, err := q.GetFirstNFlags(context.Background(), 2)
	assertNoError(t, err, "GetFirstNFlags limit=2")
	assertFlagSliceLen(t, 2, got, "GetFirstNFlags result")
}

func TestGetFirstNFlags_LimitZero_ReturnsEmpty(t *testing.T) {
	q := newTestQueries(t)
	insertFlags(t, q, []Flag{sampleFlag("FLAG{n_limit0}")})

	got, err := q.GetFirstNFlags(context.Background(), 0)
	assertNoError(t, err, "GetFirstNFlags limit=0")
	assertFlagSliceLen(t, 0, got, "GetFirstNFlags limit=0 result")
}

func TestGetFirstNFlags_LimitGreaterThanTotal_ReturnsAll(t *testing.T) {
	q := newTestQueries(t)
	insertFlags(t, q, []Flag{
		sampleFlag("FLAG{n_all_001}"),
		sampleFlag("FLAG{n_all_002}"),
	})

	got, err := q.GetFirstNFlags(context.Background(), 100)
	assertNoError(t, err, "GetFirstNFlags limit>total")
	assertFlagSliceLen(t, 2, got, "GetFirstNFlags limit>total result")
}

// --- GetFirstNFlagCodes -------------------------------------------------------

func TestGetFirstNFlagCodes_LimitLessThanTotal_ReturnsLimitCodes(t *testing.T) {
	q := newTestQueries(t)
	insertFlags(t, q, []Flag{
		sampleFlag("FLAG{nc_001}"),
		sampleFlag("FLAG{nc_002}"),
		sampleFlag("FLAG{nc_003}"),
	})

	codes, err := q.GetFirstNFlagCodes(context.Background(), 2)
	assertNoError(t, err, "GetFirstNFlagCodes limit=2")
	assertStringSliceLen(t, 2, codes, "GetFirstNFlagCodes result")
}

func TestGetFirstNFlagCodes_EmptyDB_ReturnsEmpty(t *testing.T) {
	q := newTestQueries(t)

	codes, err := q.GetFirstNFlagCodes(context.Background(), 10)
	assertNoError(t, err, "GetFirstNFlagCodes on empty DB")
	assertStringSliceLen(t, 0, codes, "GetFirstNFlagCodes empty DB result")
}

// --- GetPagedFlags ------------------------------------------------------------

// --- GetPagedFlagCodes --------------------------------------------------------

// --- GetUnsubmittedFlags ------------------------------------------------------

func TestGetUnsubmittedFlags_ReturnsOnlyUnsubmitted(t *testing.T) {
	q := newTestQueries(t)

	unsubmitted := sampleFlag("FLAG{unsub_001}")
	unsubmitted.Status = 0

	accepted := sampleFlag("FLAG{unsub_accepted}")
	accepted.Status = 1

	insertFlags(t, q, []Flag{unsubmitted, accepted})

	got, err := q.GetUnsubmittedFlags(context.Background(), 10)
	assertNoError(t, err, "GetUnsubmittedFlags")
	assertFlagSliceLen(t, 1, got, "GetUnsubmittedFlags result")
	if got[0].Status != 0 {
		t.Errorf("expected Status=UNSUBMITTED, got %d", got[0].Status)
	}
}

func TestGetUnsubmittedFlags_EmptyDB_ReturnsEmpty(t *testing.T) {
	q := newTestQueries(t)

	got, err := q.GetUnsubmittedFlags(context.Background(), 10)
	assertNoError(t, err, "GetUnsubmittedFlags empty DB")
	assertFlagSliceLen(t, 0, got, "GetUnsubmittedFlags empty DB result")
}

func TestGetUnsubmittedFlags_RespectsLimit(t *testing.T) {
	q := newTestQueries(t)
	for i := range 5 {
		f := sampleFlag("FLAG{unsub_lim_" + string(rune('A'+i)) + "}")
		f.Status = 0
		insertFlag(t, q, f)
	}

	got, err := q.GetUnsubmittedFlags(context.Background(), 3)
	assertNoError(t, err, "GetUnsubmittedFlags limit=3")
	assertFlagSliceLen(t, 3, got, "GetUnsubmittedFlags limit=3 result")
}

func TestGetUnsubmittedFlags_OrderedBySubmitTimeAsc(t *testing.T) {
	q := newTestQueries(t)
	now := uint64(time.Now().Unix())

	older := sampleFlag("FLAG{unsub_older}")
	older.Status = 0
	older.SubmitTime = now - 200

	newer := sampleFlag("FLAG{unsub_newer}")
	newer.Status = 0
	newer.SubmitTime = now

	insertFlags(t, q, []Flag{newer, older})

	got, err := q.GetUnsubmittedFlags(context.Background(), 10)
	assertNoError(t, err, "GetUnsubmittedFlags order")
	if len(got) < 2 {
		t.Fatalf("expected 2 results, got %d", len(got))
	}
	if got[0].SubmitTime > got[1].SubmitTime {
		t.Errorf("expected ascending SubmitTime; got[0]=%d, got[1]=%d",
			got[0].SubmitTime, got[1].SubmitTime)
	}
}

// --- GetUnsubmittedFlagCodes --------------------------------------------------

func TestGetUnsubmittedFlagCodes_ReturnsOnlyUnsubmittedCodes(t *testing.T) {
	q := newTestQueries(t)

	unsubmitted := sampleFlag("FLAG{unsubcode_001}")
	unsubmitted.Status = 0

	accepted := sampleFlag("FLAG{unsubcode_acc}")
	accepted.Status = 1

	insertFlags(t, q, []Flag{unsubmitted, accepted})

	codes, err := q.GetUnsubmittedFlagCodes(context.Background(), 10)
	assertNoError(t, err, "GetUnsubmittedFlagCodes")
	assertStringSliceLen(t, 1, codes, "GetUnsubmittedFlagCodes result")
	if codes[0] != unsubmitted.FlagCode {
		t.Errorf("expected code %q, got %q", unsubmitted.FlagCode, codes[0])
	}
}

// --- UpdateFlagStatusByCode ---------------------------------------------------

func TestUpdateFlagStatusByCode_ExistingFlag_UpdatesFields(t *testing.T) {
	q := newTestQueries(t)
	flag := sampleFlag("FLAG{update_001}")
	flag.Status = 0
	insertFlag(t, q, flag)

	newResponseTime := uint64(time.Now().Unix())
	err := q.UpdateFlagStatusByCode(context.Background(), UpdateFlagStatusByCodeParams{
		FlagCode:     flag.FlagCode,
		Status:       1,
		Msg:          "well done",
		ResponseTime: newResponseTime,
	})
	assertNoError(t, err, "UpdateFlagStatusByCode")

	got := mustGetFlag(t, q, flag.FlagCode)
	if got.Status != 1 {
		t.Errorf("Status: want ACCEPTED, got %d", got.Status)
	}
	if got.Msg != "well done" {
		t.Errorf("Msg: want %q, got %q", "well done", got.Msg)
	}
	if got.ResponseTime != newResponseTime {
		t.Errorf("ResponseTime: want %d, got %d", newResponseTime, got.ResponseTime)
	}
}

func TestUpdateFlagStatusByCode_NonExistentFlag_NoError(t *testing.T) {
	q := newTestQueries(t)

	// UPDATE on a non-existent PK is valid SQL and should not return an error.
	err := q.UpdateFlagStatusByCode(context.Background(), UpdateFlagStatusByCodeParams{
		FlagCode:     "FLAG{ghost}",
		Status:       1,
		Msg:          "ok",
		ResponseTime: 0,
	})
	assertNoError(t, err, "UpdateFlagStatusByCode non-existent flag")
}

func TestUpdateFlagStatusByCode_ZeroResponseTime_Stored(t *testing.T) {
	q := newTestQueries(t)
	flag := sampleFlag("FLAG{update_zero_rt}")
	insertFlag(t, q, flag)

	err := q.UpdateFlagStatusByCode(context.Background(), UpdateFlagStatusByCodeParams{
		FlagCode:     flag.FlagCode,
		Status:       2,
		Msg:          "nope",
		ResponseTime: 0,
	})
	assertNoError(t, err, "UpdateFlagStatusByCode zero ResponseTime")

	got := mustGetFlag(t, q, flag.FlagCode)
	if got.ResponseTime != 0 {
		t.Errorf("expected ResponseTime=0, got %d", got.ResponseTime)
	}
}

func TestUpdateFlagStatusByCode_ImmutableFieldsUnchanged(t *testing.T) {
	q := newTestQueries(t)
	flag := sampleFlag("FLAG{update_immutable}")
	insertFlag(t, q, flag)

	_ = q.UpdateFlagStatusByCode(context.Background(), UpdateFlagStatusByCodeParams{
		FlagCode:     flag.FlagCode,
		Status:       3,
		Msg:          "error msg",
		ResponseTime: 999,
	})

	got := mustGetFlag(t, q, flag.FlagCode)
	// Fields not touched by the UPDATE must remain unchanged.
	if got.ServiceName != flag.ServiceName {
		t.Errorf("ServiceName changed: want %q, got %q", flag.ServiceName, got.ServiceName)
	}
	if got.TeamID != flag.TeamID {
		t.Errorf("TeamID changed: want %d, got %d", flag.TeamID, got.TeamID)
	}
	if got.Username != flag.Username {
		t.Errorf("Username changed: want %q, got %q", flag.Username, got.Username)
	}
}

// --- DeleteFlagByCode ---------------------------------------------------------

func TestDeleteFlagByCode_ExistingFlag_Removed(t *testing.T) {
	q := newTestQueries(t)
	flag := sampleFlag("FLAG{delete_001}")
	insertFlag(t, q, flag)

	err := q.DeleteFlagByCode(context.Background(), flag.FlagCode)
	assertNoError(t, err, "DeleteFlagByCode existing")

	_, err = q.GetFlagByCode(context.Background(), flag.FlagCode)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("expected sql.ErrNoRows after deletion, got: %v", err)
	}
}

func TestDeleteFlagByCode_NonExistentFlag_NoError(t *testing.T) {
	q := newTestQueries(t)

	err := q.DeleteFlagByCode(context.Background(), "FLAG{ghost_delete}")
	assertNoError(t, err, "DeleteFlagByCode non-existent flag")
}

func TestDeleteFlagByCode_DoesNotAffectOtherRows(t *testing.T) {
	q := newTestQueries(t)
	keep := sampleFlag("FLAG{delete_keep}")
	remove := sampleFlag("FLAG{delete_remove}")
	insertFlags(t, q, []Flag{keep, remove})

	err := q.DeleteFlagByCode(context.Background(), remove.FlagCode)
	assertNoError(t, err, "DeleteFlagByCode selective")

	count, err := q.CountFlags(context.Background())
	assertNoError(t, err, "CountFlags after selective delete")
	assertInt64Equal(t, 1, count, "remaining flag count")

	_ = mustGetFlag(t, q, keep.FlagCode) // must still exist
}

// --- DeleteFlagByTTL ----------------------------------------------------------

func TestDeleteFlagByTTL_VeryOldFlags_Deleted(t *testing.T) {
	q := newTestQueries(t)

	old := sampleFlag("FLAG{ttl_old}")
	// Set response_time far in the past (1970) so it is older than any
	// negative strftime modifier we pass.
	old.ResponseTime = 1

	insertFlag(t, q, old)

	// "-0 seconds" means: delete where response_time < now, which covers our
	// epoch-1 timestamp.
	rowsAffected, err := q.DeleteFlagByTTL(context.Background(), "-0 seconds")
	assertNoError(t, err, "DeleteFlagByTTL old flag")
	if rowsAffected < 1 {
		t.Errorf("expected at least 1 row deleted, got %d", rowsAffected)
	}

	count, err := q.CountFlags(context.Background())
	assertNoError(t, err, "CountFlags after TTL delete")
	assertInt64Equal(t, 0, count, "flag count after TTL delete")
}

func TestDeleteFlagByTTL_FutureFlags_NotDeleted(t *testing.T) {
	q := newTestQueries(t)

	future := sampleFlag("FLAG{ttl_future}")
	// response_time set far in the future so it is never "older than now".
	future.ResponseTime = uint64(time.Now().Unix()) + 999_999_999
	insertFlag(t, q, future)

	// A very short TTL window (0 seconds): only things strictly before now are deleted.
	_, err := q.DeleteFlagByTTL(context.Background(), "-0 seconds")
	assertNoError(t, err, "DeleteFlagByTTL future flag")

	count, err := q.CountFlags(context.Background())
	assertNoError(t, err, "CountFlags after TTL (future flag should survive)")
	assertInt64Equal(t, 1, count, "future flag should still exist")
}

func TestDeleteFlagByTTL_EmptyTable_ReturnsZero(t *testing.T) {
	q := newTestQueries(t)

	rowsAffected, err := q.DeleteFlagByTTL(context.Background(), "-0 seconds")
	assertNoError(t, err, "DeleteFlagByTTL empty table")
	assertInt64Equal(t, 0, rowsAffected, "rows affected on empty table")
}

// --- CountFlags ---------------------------------------------------------------

func TestCountFlags_EmptyDB_ReturnsZero(t *testing.T) {
	q := newTestQueries(t)

	count, err := q.CountFlags(context.Background())
	assertNoError(t, err, "CountFlags empty DB")
	assertInt64Equal(t, 0, count, "CountFlags empty DB")
}

func TestCountFlags_AfterInserts_ReturnsCorrectCount(t *testing.T) {
	q := newTestQueries(t)
	insertFlags(t, q, []Flag{
		sampleFlag("FLAG{cnt_001}"),
		sampleFlag("FLAG{cnt_002}"),
		sampleFlag("FLAG{cnt_003}"),
	})

	count, err := q.CountFlags(context.Background())
	assertNoError(t, err, "CountFlags after 3 inserts")
	assertInt64Equal(t, 3, count, "CountFlags after 3 inserts")
}

// --- Store.CountFlags (filtered) ---------------------------------------------

func TestCountFlagsFiltered_EmptyDB_ReturnsZero(t *testing.T) {
	store := newTestStore(t)

	count, err := store.CountFlags(context.Background(),
		buildStoreQuery(0, 0, "", "", "", 0))
	assertNoError(t, err, "CountFlags filtered empty DB")
	assertInt64Equal(t, 0, count, "CountFlags filtered empty DB")
}

func TestCountFlagsFiltered_NoFilters_CountsAllRows(t *testing.T) {
	store := newTestStore(t)
	insertFlags(t, store.Queries, []Flag{
		sampleFlag("FLAG{cff_001}"),
		sampleFlag("FLAG{cff_002}"),
		sampleFlag("FLAG{cff_003}"),
	})

	count, err := store.CountFlags(context.Background(),
		buildStoreQuery(0, 0, "", "", "", 0))
	assertNoError(t, err, "CountFlags filtered no filters")
	assertInt64Equal(t, 3, count, "CountFlags filtered no filters — should count all rows")
}

func TestCountFlagsFiltered_ByStatus_CountsOnlyMatching(t *testing.T) {
	store := newTestStore(t)

	for i := range 3 {
		f := sampleFlag("FLAG{cffs_acc_" + string(rune('A'+i)) + "}")
		f.Status = 1
		insertFlag(t, store.Queries, f)
	}
	denied := sampleFlag("FLAG{cffs_den}")
	denied.Status = 2
	insertFlag(t, store.Queries, denied)

	count, err := store.CountFlags(context.Background(),
		buildStoreQuery(0, 1, "", "", "", 0))
	assertNoError(t, err, "CountFlags filtered by status=ACCEPTED")
	assertInt64Equal(t, 3, count, "CountFlags filtered by status=ACCEPTED — should count 3")
}

func TestCountFlagsFiltered_ByStatus_ExcludesNonMatching(t *testing.T) {
	store := newTestStore(t)

	accepted := sampleFlag("FLAG{cffs_excl_acc}")
	accepted.Status = 1
	insertFlag(t, store.Queries, accepted)

	denied := sampleFlag("FLAG{cffs_excl_den}")
	denied.Status = 2
	insertFlag(t, store.Queries, denied)

	count, err := store.CountFlags(context.Background(),
		buildStoreQuery(0, 2, "", "", "", 0))
	assertNoError(t, err, "CountFlags filtered by status=DENIED")
	assertInt64Equal(t, 1, count, "CountFlags filtered by status=DENIED — should count 1")
}

func TestCountFlagsFiltered_ByTeam_CountsOnlyThatTeam(t *testing.T) {
	store := newTestStore(t)

	for i := range 4 {
		f := sampleFlag("FLAG{cfft_t1_" + string(rune('A'+i)) + "}")
		f.TeamID = 10
		insertFlag(t, store.Queries, f)
	}
	other := sampleFlag("FLAG{cfft_t2}")
	other.TeamID = 20
	insertFlag(t, store.Queries, other)

	count, err := store.CountFlags(context.Background(),
		buildStoreQuery(10, 0, "", "", "", 0))
	assertNoError(t, err, "CountFlags filtered by team=10")
	assertInt64Equal(t, 4, count, "CountFlags filtered by team=10 — should count 4")
}

func TestCountFlagsFiltered_SearchByFlagCode_CountsMatching(t *testing.T) {
	store := newTestStore(t)

	match1 := sampleFlag("FLAG{cffs_search_xyz_001}")
	match2 := sampleFlag("FLAG{cffs_search_xyz_002}")
	nomatch := sampleFlag("FLAG{cffs_search_abc_003}")
	insertFlags(t, store.Queries, []Flag{match1, match2, nomatch})

	count, err := store.CountFlags(context.Background(),
		buildStoreQuery(0, 0, "", "search_xyz", "flag_code", 0))
	assertNoError(t, err, "CountFlags filtered search by flag_code")
	assertInt64Equal(t, 2, count, "CountFlags filtered search by flag_code — should count 2")
}

// --- BulkInsertFlags ---------------------------------------------------------

// buildStoreQuery constructs a FlagsQuery for Store method tests.
// Pass 0 / "" for any field you do not want to filter on.
// limit=0 leaves Limit.Valid=false (QueryFlagsParams then uses its default of 40).
func buildStoreQuery(teamID, status int64, serviceName, search, searchField string, limit int64) FlagsQuery {
	q := FlagsQuery{
		Limit: sql.NullInt64{Int64: limit, Valid: limit > 0},
	}
	if teamID != 0 {
		q.TeamID = sql.NullInt64{Int64: teamID, Valid: true}
	}
	if status != 0 {
		q.Status = sql.NullInt64{Int64: status, Valid: true}
	}
	if serviceName != "" {
		q.ServiceName = sql.NullString{String: serviceName, Valid: true}
	}
	if search != "" {
		q.Search = sql.NullString{String: search, Valid: true}
		if searchField != "" {
			q.SearchField = sql.NullString{String: searchField, Valid: true}
		}
	}
	return q
}

func TestBulkInsertFlags_EmptySlice_NoError(t *testing.T) {
	store := newTestStore(t)

	err := store.BulkInsertFlags(context.Background(), []Flag{})
	assertNoError(t, err, "BulkInsertFlags empty slice")

	count, err := store.Queries.CountFlags(context.Background())
	assertNoError(t, err, "CountFlags after empty BulkInsertFlags")
	assertInt64Equal(t, 0, count, "count after empty BulkInsertFlags")
}

func TestBulkInsertFlags_SingleFlag_Inserted(t *testing.T) {
	store := newTestStore(t)
	flag := sampleFlag("FLAG{bulk_single_001}")

	err := store.BulkInsertFlags(context.Background(), []Flag{flag})
	assertNoError(t, err, "BulkInsertFlags single flag")

	got := mustGetFlag(t, store.Queries, flag.FlagCode)
	assertFlagsEqual(t, flag, got)
}

func TestBulkInsertFlags_MultipleFlags_AllInserted(t *testing.T) {
	store := newTestStore(t)
	flags := []Flag{
		sampleFlag("FLAG{bulk_multi_001}"),
		sampleFlag("FLAG{bulk_multi_002}"),
		sampleFlag("FLAG{bulk_multi_003}"),
	}

	err := store.BulkInsertFlags(context.Background(), flags)
	assertNoError(t, err, "BulkInsertFlags multiple flags")

	count, err := store.Queries.CountFlags(context.Background())
	assertNoError(t, err, "CountFlags after BulkInsertFlags multiple")
	assertInt64Equal(t, int64(len(flags)), count, "count after BulkInsertFlags multiple")
}

func TestBulkInsertFlags_DuplicateInBatch_SilentlyIgnored(t *testing.T) {
	store := newTestStore(t)
	original := sampleFlag("FLAG{bulk_dup_001}")

	// Both rows share the same primary key — INSERT OR IGNORE should keep only one.
	err := store.BulkInsertFlags(context.Background(), []Flag{original, original})
	assertNoError(t, err, "BulkInsertFlags duplicate in same batch")

	count, err := store.Queries.CountFlags(context.Background())
	assertNoError(t, err, "CountFlags after BulkInsertFlags duplicate")
	assertInt64Equal(t, 1, count, "duplicate in same batch should be silently ignored")
}

func TestBulkInsertFlags_AlreadyExistingFlag_SilentlyIgnored(t *testing.T) {
	store := newTestStore(t)
	flag := sampleFlag("FLAG{bulk_exist_001}")
	insertFlag(t, store.Queries, flag) // pre-insert via Queries

	err := store.BulkInsertFlags(context.Background(), []Flag{flag})
	assertNoError(t, err, "BulkInsertFlags flag already in DB")

	count, err := store.Queries.CountFlags(context.Background())
	assertNoError(t, err, "CountFlags after BulkInsertFlags existing")
	assertInt64Equal(t, 1, count, "pre-existing flag should be silently ignored")
}

// --- QueryFlagsParams --------------------------------------------------------

func TestQueryFlagsParams_EmptyDB_ReturnsEmpty(t *testing.T) {
	store := newTestStore(t)

	got, err := store.QueryFlagsParams(context.Background(), buildStoreQuery(0, 0, "", "", "", 10))
	assertNoError(t, err, "QueryFlagsParams empty DB")
	assertFlagSliceLen(t, 0, got, "QueryFlagsParams empty DB result")
}

func TestQueryFlagsParams_NoFilters_ReturnsAll(t *testing.T) {
	store := newTestStore(t)
	insertFlags(t, store.Queries, []Flag{
		sampleFlag("FLAG{qfp_all_001}"),
		sampleFlag("FLAG{qfp_all_002}"),
		sampleFlag("FLAG{qfp_all_003}"),
	})

	got, err := store.QueryFlagsParams(context.Background(), buildStoreQuery(0, 0, "", "", "", 10))
	assertNoError(t, err, "QueryFlagsParams no filters")
	assertFlagSliceLen(t, 3, got, "QueryFlagsParams no filters result")
}

func TestQueryFlagsParams_SearchByFlagCode_ReturnsMatching(t *testing.T) {
	store := newTestStore(t)

	target := sampleFlag("FLAG{qfp_search_target_xyz}")
	other := sampleFlag("FLAG{qfp_search_other_abc}")
	insertFlags(t, store.Queries, []Flag{target, other})

	got, err := store.QueryFlagsParams(context.Background(), buildStoreQuery(0, 0, "", "search_target", "flag_code", 10))
	assertNoError(t, err, "QueryFlagsParams search by flag_code")
	assertFlagSliceLen(t, 1, got, "QueryFlagsParams search by flag_code result")
	if len(got) > 0 && got[0].FlagCode != target.FlagCode {
		t.Errorf("expected FlagCode %q, got %q", target.FlagCode, got[0].FlagCode)
	}
}

func TestQueryFlagsParams_LimitRespected(t *testing.T) {
	store := newTestStore(t)
	for i := range 5 {
		insertFlag(t, store.Queries, sampleFlag("FLAG{qfp_lim_"+string(rune('A'+i))+"}"))
	}

	got, err := store.QueryFlagsParams(context.Background(), buildStoreQuery(0, 0, "", "", "", 3))
	assertNoError(t, err, "QueryFlagsParams limit")
	assertFlagSliceLen(t, 3, got, "QueryFlagsParams limit=3 with 5 flags inserted")
}

func TestQueryFlagsParams_CursorPagination(t *testing.T) {
	store := newTestStore(t)
	for i := range 5 {
		insertFlag(t, store.Queries, sampleFlag("FLAG{qfp_pag_"+string(rune('A'+i))+"}"))
	}

	// Page 1: first 2 items (no cursor — defaults to max int64 sentinel).
	page1, err := store.QueryFlagsParams(context.Background(), buildStoreQuery(0, 0, "", "", "", 2))
	assertNoError(t, err, "QueryFlagsParams cursor page1")
	assertFlagSliceLen(t, 2, page1, "QueryFlagsParams cursor page1")

	// Build cursor from the last item on page 1.
	last1 := page1[len(page1)-1]
	p2 := buildStoreQuery(0, 0, "", "", "", 2)
	p2.CursorTime = sql.NullInt64{Int64: int64(last1.SubmitTime), Valid: true}
	p2.CursorID = sql.NullInt64{Int64: last1.ID, Valid: true}

	// Page 2: next 2 items.
	page2, err := store.QueryFlagsParams(context.Background(), p2)
	assertNoError(t, err, "QueryFlagsParams cursor page2")
	assertFlagSliceLen(t, 2, page2, "QueryFlagsParams cursor page2")

	// Build cursor from the last item on page 2.
	last2 := page2[len(page2)-1]
	p3 := buildStoreQuery(0, 0, "", "", "", 2)
	p3.CursorTime = sql.NullInt64{Int64: int64(last2.SubmitTime), Valid: true}
	p3.CursorID = sql.NullInt64{Int64: last2.ID, Valid: true}

	// Page 3: remaining 1 item.
	page3, err := store.QueryFlagsParams(context.Background(), p3)
	assertNoError(t, err, "QueryFlagsParams cursor page3")
	assertFlagSliceLen(t, 1, page3, "QueryFlagsParams cursor page3 (last partial)")
}

func TestQueryFlagsParams_OrderedBySubmitTimeDesc(t *testing.T) {
	store := newTestStore(t)

	baseTime := uint64(time.Now().Unix())
	flags := []Flag{
		sampleFlag("FLAG{qfp_ord_001}"),
		sampleFlag("FLAG{qfp_ord_002}"),
		sampleFlag("FLAG{qfp_ord_003}"),
	}
	for i := range flags {
		flags[i].SubmitTime = baseTime + uint64(i)
	}
	insertFlags(t, store.Queries, flags)

	got, err := store.QueryFlagsParams(context.Background(), buildStoreQuery(0, 0, "", "", "", 10))
	assertNoError(t, err, "QueryFlagsParams ordering")
	if len(got) < 2 {
		t.Fatalf("expected at least 2 flags, got %d", len(got))
	}
	for i := 1; i < len(got); i++ {
		if got[i-1].SubmitTime < got[i].SubmitTime {
			t.Errorf("flags not ordered DESC: got[%d].SubmitTime=%d < got[%d].SubmitTime=%d",
				i-1, got[i-1].SubmitTime, i, got[i].SubmitTime)
		}
	}
}

// --- Store.CountFlags --------------------------------------------------------

func TestStoreCountFlags_EmptyDB_ReturnsZero(t *testing.T) {
	store := newTestStore(t)

	count, err := store.CountFlags(context.Background(), FlagsQuery{})
	assertNoError(t, err, "Store.CountFlags empty DB")
	assertInt64Equal(t, 0, count, "Store.CountFlags empty DB")
}

func TestStoreCountFlags_NoFilters_CountsAllRows(t *testing.T) {
	store := newTestStore(t)
	insertFlags(t, store.Queries, []Flag{
		sampleFlag("FLAG{scf_all_001}"),
		sampleFlag("FLAG{scf_all_002}"),
		sampleFlag("FLAG{scf_all_003}"),
	})

	count, err := store.CountFlags(context.Background(), FlagsQuery{})
	assertNoError(t, err, "Store.CountFlags no filters")
	assertInt64Equal(t, 3, count, "Store.CountFlags no filters — should count all rows")
}

func TestStoreCountFlags_ByStatus_CountsOnlyMatching(t *testing.T) {
	store := newTestStore(t)

	for i := range 3 {
		f := sampleFlag("FLAG{scf_status_acc_" + string(rune('A'+i)) + "}")
		f.Status = 1
		insertFlag(t, store.Queries, f)
	}
	denied := sampleFlag("FLAG{scf_status_den}")
	denied.Status = 2
	insertFlag(t, store.Queries, denied)

	count, err := store.CountFlags(context.Background(), buildStoreQuery(0, 1, "", "", "", 0))
	assertNoError(t, err, "Store.CountFlags by status=1")
	assertInt64Equal(t, 3, count, "Store.CountFlags by status=1 — should count 3")
}

func TestStoreCountFlags_ByTeam_CountsOnlyThatTeam(t *testing.T) {
	store := newTestStore(t)

	for i := range 4 {
		f := sampleFlag("FLAG{scf_team_t10_" + string(rune('A'+i)) + "}")
		f.TeamID = 10
		insertFlag(t, store.Queries, f)
	}
	other := sampleFlag("FLAG{scf_team_t20}")
	other.TeamID = 20
	insertFlag(t, store.Queries, other)

	count, err := store.CountFlags(context.Background(), buildStoreQuery(10, 0, "", "", "", 0))
	assertNoError(t, err, "Store.CountFlags by team=10")
	assertInt64Equal(t, 4, count, "Store.CountFlags by team=10 — should count 4")
}

func TestStoreCountFlags_ByService_CountsOnlyMatching(t *testing.T) {
	store := newTestStore(t)

	for i := range 2 {
		f := sampleFlag("FLAG{scf_svc_target_" + string(rune('A'+i)) + "}")
		f.ServiceName = "target-service"
		insertFlag(t, store.Queries, f)
	}
	other := sampleFlag("FLAG{scf_svc_other}")
	other.ServiceName = "other-service"
	insertFlag(t, store.Queries, other)

	count, err := store.CountFlags(context.Background(), buildStoreQuery(0, 0, "target-service", "", "", 0))
	assertNoError(t, err, "Store.CountFlags by service")
	assertInt64Equal(t, 2, count, "Store.CountFlags by service — should count 2")
}

func TestStoreCountFlags_SearchByFlagCode_CountsMatching(t *testing.T) {
	store := newTestStore(t)

	match1 := sampleFlag("FLAG{scf_search_xyz_001}")
	match2 := sampleFlag("FLAG{scf_search_xyz_002}")
	nomatch := sampleFlag("FLAG{scf_search_abc_003}")
	insertFlags(t, store.Queries, []Flag{match1, match2, nomatch})

	count, err := store.CountFlags(context.Background(), buildStoreQuery(0, 0, "", "search_xyz", "flag_code", 0))
	assertNoError(t, err, "Store.CountFlags search by flag_code")
	assertInt64Equal(t, 2, count, "Store.CountFlags search by flag_code — should count 2")
}
