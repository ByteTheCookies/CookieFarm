package sqlite

import (
	"context"
	"errors"
	"fmt"
	"time"

	"crawshaw.io/sqlite"
	"crawshaw.io/sqlite/sqlitex"
	"github.com/ByteTheCookies/CookieFarm/pkg/models"
)

const MaxBatchSizeSQLite = 999

func InsertBatch(ctx context.Context, conn *sqlite.Conn, flags []models.ClientData, stmt *sqlite.Stmt) error {
	for _, f := range flags {
		stmt.Reset()
		stmt.ClearBindings()
		stmt.BindText(1, f.FlagCode)
		stmt.BindText(2, f.ServiceName)
		stmt.BindInt64(3, int64(f.PortService))
		stmt.BindInt64(4, int64(f.SubmitTime))
		stmt.BindInt64(5, int64(f.ResponseTime))
		stmt.BindText(6, f.Status)
		stmt.BindInt64(7, int64(f.TeamID))
		stmt.BindText(8, f.Msg)
		stmt.BindText(9, f.Username)
		stmt.BindText(10, f.ExploitName)

		if _, err := stmt.Step(); err != nil {
			return fmt.Errorf("insert step: %w", err)
		}
	}
	return nil
}

// AddFlagsWithContext adds a batch of flags to the database with a custom context.
// It divides the flags into batches to insert in chunks, helping avoid hitting query parameter limits.
func AddFlagsWithContext(ctx context.Context, flags []models.ClientData) error {
	if len(flags) == 0 {
		return nil
	}

	conn := DBPool.Get(ctx)
	if conn == nil {
		return errors.New("no available SQLite connection")
	}
	defer DBPool.Put(conn)

	if err := sqlitex.Exec(conn, "SAVEPOINT addflags", nil); err != nil {
		return fmt.Errorf("savepoint: %w", err)
	}
	defer func() {
		_ = sqlitex.Exec(conn, "ROLLBACK TO addflags", nil)
	}()

	stmt, err := conn.Prepare(`
		INSERT OR IGNORE INTO flags(
			flag_code, service_name, port_service,
			submit_time, response_time, status,
			team_id, msg, username, exploit_name
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("prepare: %w", err)
	}
	defer stmt.Finalize()

	const batchSize = 1000 / 10
	for i := 0; i < len(flags); i += batchSize {
		end := i + batchSize
		if end > len(flags) {
			end = len(flags)
		}
		err := InsertBatch(ctx, conn, flags[i:end], stmt)
		if err != nil {
			return err
		}
	}

	committed := false
	defer func() {
		if !committed {
			_ = sqlitex.Exec(conn, "ROLLBACK TO addflags", nil)
		}
	}()

	if err := sqlitex.Exec(conn, "RELEASE addflags", nil); err != nil {
		return fmt.Errorf("release: %w", err)
	}
	committed = true

	return nil
}

// AddFlags adds a batch of flags to the database with a default timeout.
// It calls AddFlagsWithContext with a timeout context.
func AddFlags(flags []models.ClientData) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return AddFlagsWithContext(ctx, flags)
}

// AddFlag adds a single flag to the database.
// It calls the AddFlags function to add the flag as a batch of size 1.
func AddFlag(flag models.ClientData) error {
	return AddFlags([]models.ClientData{flag})
}

// AddFlagWithContext adds a single flag to the database with a custom context.
func AddFlagWithContext(ctx context.Context, flag models.ClientData) error {
	return AddFlagsWithContext(ctx, []models.ClientData{flag})
}
