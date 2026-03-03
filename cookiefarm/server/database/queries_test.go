package database

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"
)

// ─── AddFlag ──────────────────────────────────────────────────────────────────

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
		Status:       "UNSUBMITTED",
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

// ─── GetFlagByCode ────────────────────────────────────────────────────────────

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

// ─── GetAllFlags ──────────────────────────────────────────────────────────────

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

// ─── GetAllFlagCodes ──────────────────────────────────────────────────────────

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

// ─── GetFirstNFlags ───────────────────────────────────────────────────────────

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

// ─── GetFirstNFlagCodes ───────────────────────────────────────────────────────

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

// ─── GetPagedFlags ────────────────────────────────────────────────────────────

func TestGetPagedFlags_FirstPage_ReturnsCorrectWindow(t *testing.T) {
	q := newTestQueries(t)
	insertFlags(t, q, []Flag{
		sampleFlag("FLAG{page_001}"),
		sampleFlag("FLAG{page_002}"),
		sampleFlag("FLAG{page_003}"),
	})

	got, err := q.GetPagedFlags(context.Background(), GetPagedFlagsParams{Limit: 2, Offset: 0})
	assertNoError(t, err, "GetPagedFlags page 1")
	assertFlagSliceLen(t, 2, got, "GetPagedFlags page 1")
}

func TestGetPagedFlags_SecondPage_ReturnsRemainder(t *testing.T) {
	q := newTestQueries(t)
	insertFlags(t, q, []Flag{
		sampleFlag("FLAG{page2_001}"),
		sampleFlag("FLAG{page2_002}"),
		sampleFlag("FLAG{page2_003}"),
	})

	got, err := q.GetPagedFlags(context.Background(), GetPagedFlagsParams{Limit: 2, Offset: 2})
	assertNoError(t, err, "GetPagedFlags page 2")
	assertFlagSliceLen(t, 1, got, "GetPagedFlags page 2")
}

func TestGetPagedFlags_OffsetBeyondTotal_ReturnsEmpty(t *testing.T) {
	q := newTestQueries(t)
	insertFlags(t, q, []Flag{sampleFlag("FLAG{paged_solo}")})

	got, err := q.GetPagedFlags(context.Background(), GetPagedFlagsParams{Limit: 10, Offset: 999})
	assertNoError(t, err, "GetPagedFlags offset>total")
	assertFlagSliceLen(t, 0, got, "GetPagedFlags offset beyond total")
}

// ─── GetPagedFlagCodes ────────────────────────────────────────────────────────

func TestGetPagedFlagCodes_FirstPage(t *testing.T) {
	q := newTestQueries(t)
	insertFlags(t, q, []Flag{
		sampleFlag("FLAG{pcode_001}"),
		sampleFlag("FLAG{pcode_002}"),
		sampleFlag("FLAG{pcode_003}"),
	})

	codes, err := q.GetPagedFlagCodes(context.Background(), GetPagedFlagCodesParams{Limit: 2, Offset: 0})
	assertNoError(t, err, "GetPagedFlagCodes page 1")
	assertStringSliceLen(t, 2, codes, "GetPagedFlagCodes page 1")
}

func TestGetPagedFlagCodes_OffsetBeyondTotal_ReturnsEmpty(t *testing.T) {
	q := newTestQueries(t)
	insertFlags(t, q, []Flag{sampleFlag("FLAG{pcode_solo}")})

	codes, err := q.GetPagedFlagCodes(context.Background(), GetPagedFlagCodesParams{Limit: 10, Offset: 999})
	assertNoError(t, err, "GetPagedFlagCodes offset>total")
	assertStringSliceLen(t, 0, codes, "GetPagedFlagCodes offset beyond total")
}

// ─── GetFlagsByTeam ───────────────────────────────────────────────────────────

func TestGetFlagsByTeam_ReturnsOnlyRequestedTeam(t *testing.T) {
	q := newTestQueries(t)

	team1Flag := sampleFlag("FLAG{team_t1}")
	team1Flag.TeamID = 1

	team2Flag := sampleFlag("FLAG{team_t2}")
	team2Flag.TeamID = 2

	insertFlags(t, q, []Flag{team1Flag, team2Flag})

	got, err := q.GetFlagsByTeam(context.Background(), GetFlagsByTeamParams{
		TeamID: 1,
		Limit:  10,
		Offset: 0,
	})
	assertNoError(t, err, "GetFlagsByTeam team=1")
	assertFlagSliceLen(t, 1, got, "GetFlagsByTeam team=1 result")
	if got[0].TeamID != 1 {
		t.Errorf("expected TeamID=1, got %d", got[0].TeamID)
	}
}

func TestGetFlagsByTeam_TeamWithNoFlags_ReturnsEmpty(t *testing.T) {
	q := newTestQueries(t)
	flag := sampleFlag("FLAG{team_only1}")
	flag.TeamID = 1
	insertFlag(t, q, flag)

	got, err := q.GetFlagsByTeam(context.Background(), GetFlagsByTeamParams{
		TeamID: 99,
		Limit:  10,
		Offset: 0,
	})
	assertNoError(t, err, "GetFlagsByTeam non-existent team")
	assertFlagSliceLen(t, 0, got, "GetFlagsByTeam non-existent team result")
}

func TestGetFlagsByTeam_PaginationWorks(t *testing.T) {
	q := newTestQueries(t)
	for i := range 5 {
		f := sampleFlag("FLAG{team_paged_" + string(rune('A'+i)) + "}")
		f.TeamID = 7
		insertFlag(t, q, f)
	}

	page1, err := q.GetFlagsByTeam(context.Background(), GetFlagsByTeamParams{TeamID: 7, Limit: 3, Offset: 0})
	assertNoError(t, err, "GetFlagsByTeam page1")
	assertFlagSliceLen(t, 3, page1, "GetFlagsByTeam page1")

	page2, err := q.GetFlagsByTeam(context.Background(), GetFlagsByTeamParams{TeamID: 7, Limit: 3, Offset: 3})
	assertNoError(t, err, "GetFlagsByTeam page2")
	assertFlagSliceLen(t, 2, page2, "GetFlagsByTeam page2")
}

// ─── GetUnsubmittedFlags ──────────────────────────────────────────────────────

func TestGetUnsubmittedFlags_ReturnsOnlyUnsubmitted(t *testing.T) {
	q := newTestQueries(t)

	unsubmitted := sampleFlag("FLAG{unsub_001}")
	unsubmitted.Status = "UNSUBMITTED"

	accepted := sampleFlag("FLAG{unsub_accepted}")
	accepted.Status = "ACCEPTED"

	insertFlags(t, q, []Flag{unsubmitted, accepted})

	got, err := q.GetUnsubmittedFlags(context.Background(), 10)
	assertNoError(t, err, "GetUnsubmittedFlags")
	assertFlagSliceLen(t, 1, got, "GetUnsubmittedFlags result")
	if got[0].Status != "UNSUBMITTED" {
		t.Errorf("expected Status=UNSUBMITTED, got %q", got[0].Status)
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
		f.Status = "UNSUBMITTED"
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
	older.Status = "UNSUBMITTED"
	older.SubmitTime = now - 200

	newer := sampleFlag("FLAG{unsub_newer}")
	newer.Status = "UNSUBMITTED"
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

// ─── GetUnsubmittedFlagCodes ──────────────────────────────────────────────────

func TestGetUnsubmittedFlagCodes_ReturnsOnlyUnsubmittedCodes(t *testing.T) {
	q := newTestQueries(t)

	unsubmitted := sampleFlag("FLAG{unsubcode_001}")
	unsubmitted.Status = "UNSUBMITTED"

	accepted := sampleFlag("FLAG{unsubcode_acc}")
	accepted.Status = "ACCEPTED"

	insertFlags(t, q, []Flag{unsubmitted, accepted})

	codes, err := q.GetUnsubmittedFlagCodes(context.Background(), 10)
	assertNoError(t, err, "GetUnsubmittedFlagCodes")
	assertStringSliceLen(t, 1, codes, "GetUnsubmittedFlagCodes result")
	if codes[0] != unsubmitted.FlagCode {
		t.Errorf("expected code %q, got %q", unsubmitted.FlagCode, codes[0])
	}
}

// ─── UpdateFlagStatusByCode ───────────────────────────────────────────────────

func TestUpdateFlagStatusByCode_ExistingFlag_UpdatesFields(t *testing.T) {
	q := newTestQueries(t)
	flag := sampleFlag("FLAG{update_001}")
	flag.Status = "UNSUBMITTED"
	insertFlag(t, q, flag)

	newResponseTime := uint64(time.Now().Unix())
	err := q.UpdateFlagStatusByCode(context.Background(), UpdateFlagStatusByCodeParams{
		FlagCode:     flag.FlagCode,
		Status:       "ACCEPTED",
		Msg:          "well done",
		ResponseTime: newResponseTime,
	})
	assertNoError(t, err, "UpdateFlagStatusByCode")

	got := mustGetFlag(t, q, flag.FlagCode)
	if got.Status != "ACCEPTED" {
		t.Errorf("Status: want ACCEPTED, got %q", got.Status)
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
		Status:       "ACCEPTED",
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
		Status:       "DENIED",
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
		Status:       "ERROR",
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

// ─── DeleteFlagByCode ─────────────────────────────────────────────────────────

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

// ─── DeleteFlagByTTL ──────────────────────────────────────────────────────────

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

// ─── CountFlags ───────────────────────────────────────────────────────────────

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

// ─── GetFilteredFlags ─────────────────────────────────────────────────────────

// buildFilteredParams builds a GetFilteredFlagsParams with sane defaults for
// the nullable / interface fields so callers only override what they need.
func buildFilteredParams(teamID uint16, status, search, searchField string, limit, offset int64) GetFilteredFlagsParams {
	return GetFilteredFlagsParams{
		TeamID: sql.NullInt64{Int64: int64(teamID), Valid: teamID != 0},
		Status: sql.NullString{
			String: status,
			Valid:  status != "",
		},
		Search:      nullOrValue(search),
		SearchField: nullOrValue(searchField),
		SearchLike:  sql.NullString{String: "%" + search + "%", Valid: search != ""},
		Limit:       sql.NullInt64{Int64: limit, Valid: true},
		Offset:      sql.NullInt64{Int64: offset, Valid: true},
	}
}

// nullOrValue returns nil when s is empty, otherwise the string itself.
// This matches the sqlc convention for nullable interface{} arguments.
func nullOrValue(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func TestGetFilteredFlags_ByTeam_ReturnsOnlyThatTeam(t *testing.T) {
	q := newTestQueries(t)

	f1 := sampleFlag("FLAG{filt_t1_001}")
	f1.TeamID = 10
	f2 := sampleFlag("FLAG{filt_t2_001}")
	f2.TeamID = 20
	insertFlags(t, q, []Flag{f1, f2})

	params := buildFilteredParams(10, "", "", "", 10, 0)
	got, err := q.GetFilteredFlags(context.Background(), params)
	assertNoError(t, err, "GetFilteredFlags by team")
	assertFlagSliceLen(t, 1, got, "GetFilteredFlags by team result")
	if got[0].TeamID != 10 {
		t.Errorf("expected TeamID=10, got %d", got[0].TeamID)
	}
}

func TestGetFilteredFlags_ByStatus_ReturnsOnlyMatchingStatus(t *testing.T) {
	q := newTestQueries(t)

	accepted := sampleFlag("FLAG{filt_acc}")
	accepted.Status = "ACCEPTED"
	denied := sampleFlag("FLAG{filt_den}")
	denied.Status = "DENIED"
	insertFlags(t, q, []Flag{accepted, denied})

	params := buildFilteredParams(0, "ACCEPTED", "", "", 10, 0)
	got, err := q.GetFilteredFlags(context.Background(), params)
	assertNoError(t, err, "GetFilteredFlags by status")
	for _, f := range got {
		if f.Status != "ACCEPTED" {
			t.Errorf("expected Status=ACCEPTED, got %q", f.Status)
		}
	}
}

func TestGetFilteredFlags_SearchByFlagCode_ReturnsMatching(t *testing.T) {
	q := newTestQueries(t)

	target := sampleFlag("FLAG{search_target_xyz}")
	other := sampleFlag("FLAG{search_other_abc}")
	insertFlags(t, q, []Flag{target, other})

	params := buildFilteredParams(0, "", "FLAG{search_target_xyz}", "flag_code", 10, 0)
	got, err := q.GetFilteredFlags(context.Background(), params)
	assertNoError(t, err, "GetFilteredFlags search by flag_code")
	assertFlagSliceLen(t, 1, got, "GetFilteredFlags search by flag_code result")
	if len(got) > 0 && got[0].FlagCode != target.FlagCode {
		t.Errorf("expected FlagCode %q, got %q", target.FlagCode, got[0].FlagCode)
	}
}

func TestGetFilteredFlags_SearchByServiceName_ReturnsMatching(t *testing.T) {
	q := newTestQueries(t)

	f1 := sampleFlag("FLAG{filt_svc_001}")
	f1.ServiceName = "specialservice"
	f2 := sampleFlag("FLAG{filt_svc_002}")
	f2.ServiceName = "normalservice"
	insertFlags(t, q, []Flag{f1, f2})

	params := buildFilteredParams(0, "", "special", "service_name", 10, 0)
	got, err := q.GetFilteredFlags(context.Background(), params)
	assertNoError(t, err, "GetFilteredFlags search by service_name")
	assertFlagSliceLen(t, 1, got, "GetFilteredFlags search by service_name result")
}

func TestGetFilteredFlags_SearchAll_MatchesAcrossColumns(t *testing.T) {
	q := newTestQueries(t)

	f1 := sampleFlag("FLAG{filt_all_001}")
	f1.ExploitName = "magic_exploit"
	f2 := sampleFlag("FLAG{filt_all_002}")
	f2.ExploitName = "normal_exploit"
	insertFlags(t, q, []Flag{f1, f2})

	params := buildFilteredParams(0, "", "magic", "exploit_name", 10, 0)
	got, err := q.GetFilteredFlags(context.Background(), params)
	assertNoError(t, err, "GetFilteredFlags search all")
	assertFlagSliceLen(t, 1, got, "GetFilteredFlags search all result")
}

func TestGetFilteredFlags_Pagination_LimitOffset(t *testing.T) {
	q := newTestQueries(t)
	for i := range 5 {
		f := sampleFlag("FLAG{filt_pag_" + string(rune('A'+i)) + "}")
		insertFlag(t, q, f)
	}

	page1, err := q.GetFilteredFlags(context.Background(),
		buildFilteredParams(0, "", "", "", 2, 0))
	assertNoError(t, err, "GetFilteredFlags pagination page1")
	assertFlagSliceLen(t, 2, page1, "GetFilteredFlags page1")

	page3, err := q.GetFilteredFlags(context.Background(),
		buildFilteredParams(0, "", "", "", 2, 4))
	assertNoError(t, err, "GetFilteredFlags pagination page3")
	assertFlagSliceLen(t, 1, page3, "GetFilteredFlags page3 (last partial)")
}

func TestGetFilteredFlags_EmptyDB_ReturnsEmpty(t *testing.T) {
	q := newTestQueries(t)

	params := buildFilteredParams(0, "", "", "", 10, 0)
	got, err := q.GetFilteredFlags(context.Background(), params)
	assertNoError(t, err, "GetFilteredFlags empty DB")
	assertFlagSliceLen(t, 0, got, "GetFilteredFlags empty DB result")
}

// ─── CountFilteredFlags ───────────────────────────────────────────────────────

// buildCountFilteredParams mirrors buildFilteredParams for the count variant.
// buildCountFilteredParams builds a CountFilteredFlagsParams with sane defaults.
func buildCountFilteredParams(teamID uint16, status, search, searchField string) CountFilteredFlagsParams {
	return CountFilteredFlagsParams{
		TeamID: sql.NullInt64{Int64: int64(teamID), Valid: teamID != 0},
		Status: sql.NullString{
			String: status,
			Valid:  status != "",
		},
		Search:      nullOrValue(search),
		SearchField: nullOrValue(searchField),
		SearchLike:  sql.NullString{String: "%" + search + "%", Valid: search != ""},
	}
}

func TestCountFilteredFlags_EmptyDB_ReturnsZero(t *testing.T) {
	q := newTestQueries(t)

	count, err := q.CountFilteredFlags(context.Background(),
		buildCountFilteredParams(0, "", "", ""))
	assertNoError(t, err, "CountFilteredFlags empty DB")
	assertInt64Equal(t, 0, count, "CountFilteredFlags empty DB")
}

func TestCountFilteredFlags_NoFilters_CountsAllRows(t *testing.T) {
	q := newTestQueries(t)
	insertFlags(t, q, []Flag{
		sampleFlag("FLAG{cff_001}"),
		sampleFlag("FLAG{cff_002}"),
		sampleFlag("FLAG{cff_003}"),
	})

	count, err := q.CountFilteredFlags(context.Background(),
		buildCountFilteredParams(0, "", "", ""))
	assertNoError(t, err, "CountFilteredFlags no filters")
	assertInt64Equal(t, 3, count, "CountFilteredFlags no filters — should count all rows")
}

func TestCountFilteredFlags_ByStatus_CountsOnlyMatching(t *testing.T) {
	q := newTestQueries(t)

	for i := range 3 {
		f := sampleFlag("FLAG{cffs_acc_" + string(rune('A'+i)) + "}")
		f.Status = "ACCEPTED"
		insertFlag(t, q, f)
	}
	denied := sampleFlag("FLAG{cffs_den}")
	denied.Status = "DENIED"
	insertFlag(t, q, denied)

	count, err := q.CountFilteredFlags(context.Background(),
		buildCountFilteredParams(0, "ACCEPTED", "", ""))
	assertNoError(t, err, "CountFilteredFlags by status=ACCEPTED")
	assertInt64Equal(t, 3, count, "CountFilteredFlags by status=ACCEPTED — should count 3")
}

func TestCountFilteredFlags_ByStatus_ExcludesNonMatching(t *testing.T) {
	q := newTestQueries(t)

	accepted := sampleFlag("FLAG{cffs_excl_acc}")
	accepted.Status = "ACCEPTED"
	insertFlag(t, q, accepted)

	denied := sampleFlag("FLAG{cffs_excl_den}")
	denied.Status = "DENIED"
	insertFlag(t, q, denied)

	count, err := q.CountFilteredFlags(context.Background(),
		buildCountFilteredParams(0, "DENIED", "", ""))
	assertNoError(t, err, "CountFilteredFlags by status=DENIED")
	assertInt64Equal(t, 1, count, "CountFilteredFlags by status=DENIED — should count 1")
}

func TestCountFilteredFlags_ByTeam_CountsOnlyThatTeam(t *testing.T) {
	q := newTestQueries(t)

	for i := range 4 {
		f := sampleFlag("FLAG{cfft_t1_" + string(rune('A'+i)) + "}")
		f.TeamID = 10
		insertFlag(t, q, f)
	}
	other := sampleFlag("FLAG{cfft_t2}")
	other.TeamID = 20
	insertFlag(t, q, other)

	count, err := q.CountFilteredFlags(context.Background(),
		buildCountFilteredParams(10, "", "", ""))
	assertNoError(t, err, "CountFilteredFlags by team=10")
	assertInt64Equal(t, 4, count, "CountFilteredFlags by team=10 — should count 4")
}

func TestCountFilteredFlags_SearchByFlagCode_CountsMatching(t *testing.T) {
	q := newTestQueries(t)

	match1 := sampleFlag("FLAG{cffs_search_xyz_001}")
	match2 := sampleFlag("FLAG{cffs_search_xyz_002}")
	nomatch := sampleFlag("FLAG{cffs_search_abc_003}")
	insertFlags(t, q, []Flag{match1, match2, nomatch})

	count, err := q.CountFilteredFlags(context.Background(),
		buildCountFilteredParams(0, "", "search_xyz", "flag_code"))
	assertNoError(t, err, "CountFilteredFlags search by flag_code")
	assertInt64Equal(t, 2, count, "CountFilteredFlags search by flag_code — should count 2")
}
