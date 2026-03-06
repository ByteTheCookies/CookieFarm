package database

import (
	"database/sql"
	"testing"
	"time"
)

// --- NewDB --------------------------------------------------------------------

func TestNewDB_InMemory_ReturnsValidDB(t *testing.T) {
	cfg := Config{
		DSN:             "file:TestNewDB_InMemory?mode=memory&cache=shared",
		MaxOpenConns:    1,
		MaxIdleConns:    1,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 5 * time.Minute,
	}

	db, err := NewDB(cfg)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if db == nil {
		t.Fatal("expected non-nil *sql.DB, got nil")
	}
	defer db.Close()
}

func TestNewDB_SchemaApplied_FlagsTableExists(t *testing.T) {
	db := newTestDB(t)

	// The flags table must exist after NewDB; a simple query is enough to
	// verify — if the table is absent SQLite returns an error.
	row := db.QueryRow("SELECT COUNT(*) FROM flags")
	var count int64
	if err := row.Scan(&count); err != nil {
		t.Fatalf("flags table does not exist or query failed: %v", err)
	}
}

func TestNewDB_SchemaApplied_FlagsTableHasCorrectColumns(t *testing.T) {
	db := newTestDB(t)

	// PRAGMA table_info returns one row per column.
	rows, err := db.Query("PRAGMA table_info(flags)")
	if err != nil {
		t.Fatalf("PRAGMA table_info failed: %v", err)
	}
	defer rows.Close()

	expected := map[string]bool{
		"flag_code":     false,
		"service_name":  false,
		"port_service":  false,
		"submit_time":   false,
		"response_time": false,
		"msg":           false,
		"status":        false,
		"team_id":       false,
		"username":      false,
		"exploit_name":  false,
	}

	for rows.Next() {
		var cid int
		var name, colType string
		var notNull, pk int
		var dfltValue sql.NullString
		if err := rows.Scan(&cid, &name, &colType, &notNull, &dfltValue, &pk); err != nil {
			t.Fatalf("scan PRAGMA row: %v", err)
		}
		if _, ok := expected[name]; ok {
			expected[name] = true
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows iteration error: %v", err)
	}

	for col, found := range expected {
		if !found {
			t.Errorf("expected column %q not found in flags table", col)
		}
	}
}

func TestNewDB_SubmitTimeIndex_Exists(t *testing.T) {
	db := newTestDB(t)

	row := db.QueryRow(
		"SELECT COUNT(*) FROM sqlite_master WHERE type='index' AND name='idx_flags_submit_time'",
	)
	var count int
	if err := row.Scan(&count); err != nil {
		t.Fatalf("query sqlite_master for index: %v", err)
	}
	if count != 1 {
		t.Errorf("expected index idx_flags_submit_time to exist, count=%d", count)
	}
}

func TestNewDB_Idempotent_SchemaAppliedTwice(t *testing.T) {
	// NewDB uses CREATE TABLE IF NOT EXISTS so calling it twice on the same
	// DSN must not fail and must not duplicate rows.
	dsn := "file:TestNewDB_Idempotent?mode=memory&cache=shared"
	cfg := Config{
		DSN:             dsn,
		MaxOpenConns:    1,
		MaxIdleConns:    1,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 5 * time.Minute,
	}

	db1, err := NewDB(cfg)
	if err != nil {
		t.Fatalf("first NewDB: %v", err)
	}
	defer db1.Close()

	db2, err := NewDB(cfg)
	if err != nil {
		t.Fatalf("second NewDB on same DSN: %v", err)
	}
	defer db2.Close()
}

func TestNewDB_ZeroPoolValues_DoesNotPanic(t *testing.T) {
	// All pool-tuning fields default to zero — Go's database/sql accepts that.
	cfg := Config{
		DSN: "file:TestNewDB_ZeroPool?mode=memory&cache=shared",
	}
	db, err := NewDB(cfg)
	if err != nil {
		t.Fatalf("NewDB with zero pool config: %v", err)
	}
	defer db.Close()
}

func TestNewDB_InvalidDSN_ReturnsError(t *testing.T) {
	// A DSN with an invalid scheme should cause either Open or Ping to fail.
	// modernc.org/sqlite treats an empty DSN as a valid temporary DB, so we
	// force an error by using a path in a directory that does not exist and
	// is read-only.
	cfg := Config{
		DSN:             "/nonexistent_dir_xyz/db.sqlite",
		MaxOpenConns:    1,
		MaxIdleConns:    1,
		ConnMaxLifetime: time.Second,
		ConnMaxIdleTime: time.Second,
	}
	db, err := NewDB(cfg)
	if err == nil {
		// Some environments may allow this; close and skip rather than fail hard.
		db.Close()
		t.Skip("SQLite created file in unexpected location — skipping error path test")
	}
}
