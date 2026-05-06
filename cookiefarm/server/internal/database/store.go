package database

import (
	"context"
	"database/sql"
	"strings"
)

const (
	CountFlagsQuery = `SELECT COUNT(*) FROM flags WHERE deleted_at IS NULL`
	BaseFlagsQuery  = `SELECT id, flag_code, service_name, port_service, submit_time, response_time, status, team_id, msg, username, exploit_name, deleted_at  FROM flags WHERE deleted_at IS NULL`
)

type Store struct {
	db      *sql.DB
	Queries *Queries
}

type FlagsQuery struct {
	CursorTime  sql.NullInt64  `json:"cursor_time"`
	CursorID    sql.NullInt64  `json:"cursor_id"`
	Limit       sql.NullInt64  `json:"limit"`
	TeamID      sql.NullInt64  `json:"team_id"`
	Status      sql.NullInt64  `json:"status"`
	ServiceName sql.NullString `json:"service_name"`
	Search      sql.NullString `json:"search"`
	SearchField sql.NullString `json:"search_field"`
}

func NewStore(db *sql.DB) *Store {
	return &Store{
		db:      db,
		Queries: New(db),
	}
}

func (s *Store) WithTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	q := New(tx)
	if err := fn(q); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

// INSERT OR IGNORE INTO flags(
// flag_code, service_name, port_service,
// submit_time, response_time, status,
// team_id, msg, username, exploit_name
// ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

func (s *Store) BulkInsertFlags(ctx context.Context, rows []Flag) error {
	if len(rows) == 0 {
		return nil
	}

	const maxParams = 900 // SQLite has a default limit of 999 parameters per statement, but we reserve some for safety
	paramsPerRow := 2
	maxRowsPerBatch := maxParams / paramsPerRow

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	for start := 0; start < len(rows); start += maxRowsPerBatch {
		end := min(start+maxRowsPerBatch, len(rows))

		batch := rows[start:end]

		var (
			sb   strings.Builder
			args []any
		)

		sb.WriteString("INSERT OR IGNORE INTO flags(flag_code, service_name, port_service,submit_time, response_time, status, team_id, msg, username, exploit_name) VALUES ")

		for i, r := range batch {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString("(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
			args = append(args,
				r.FlagCode, r.ServiceName, r.PortService,
				r.SubmitTime, r.ResponseTime, r.Status,
				r.TeamID, r.Msg, r.Username, r.ExploitName,
			)
		}
		sb.WriteString(";")

		stmt, err := tx.PrepareContext(ctx, sb.String())
		if err != nil {
			return err
		}

		defer stmt.Close()
		if _, err = stmt.ExecContext(ctx, args...); err != nil {
			return err
		}
	}

	return tx.Commit()
}

var allowedSearchFields = map[string]bool{
	"flag_code":    true,
	"service_name": true,
	"username":     true,
	"exploit_name": true,
	"msg":          true,
}

func buildQuery(base string, q FlagsQuery) (*strings.Builder, []any) {
	var (
		sb   strings.Builder
		args []any
	)

	sb.WriteString(base)

	if q.TeamID.Valid {
		sb.WriteString(" AND team_id = ?")
		args = append(args, q.TeamID.Int64)
	}
	if q.Status.Valid {
		sb.WriteString(" AND status = ?")
		args = append(args, q.Status.Int64)
	}
	if q.ServiceName.Valid {
		sb.WriteString(" AND service_name = ?")
		args = append(args, q.ServiceName.String)
	}
	if q.Search.Valid && q.SearchField.Valid {
		if !allowedSearchFields[q.SearchField.String] {
			return nil, args
		}
		sb.WriteString(" AND " + q.SearchField.String + " LIKE ?")
		safe := strings.NewReplacer(`%`, `\%`, `_`, `\_`).Replace(q.Search.String)
		args = append(args, "%"+safe+"%")
	}

	return &sb, args
}

func (s *Store) QueryFlagsParams(ctx context.Context, q FlagsQuery) ([]Flag, error) {
	if !q.CursorTime.Valid {
		q.CursorTime.Int64 = 1<<63 - 1
	}
	if !q.CursorID.Valid {
		q.CursorID.Int64 = 1<<63 - 1
	}
	if !q.Limit.Valid {
		q.Limit.Int64 = 40
	}

	sb, args := buildQuery(BaseFlagsQuery, q)

	sb.WriteString(" AND (submit_time < ? OR (submit_time = ? AND id < ?))")
	args = append(args, q.CursorTime.Int64, q.CursorTime.Int64, q.CursorID.Int64)

	sb.WriteString(" ORDER BY submit_time DESC, id DESC")
	sb.WriteString(" LIMIT ?")
	args = append(args, q.Limit.Int64)

	stmt, err := s.db.PrepareContext(ctx, sb.String())
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var flags []Flag
	for rows.Next() {
		var f Flag
		if err := rows.Scan(
			&f.ID,
			&f.FlagCode,
			&f.ServiceName,
			&f.PortService,
			&f.SubmitTime,
			&f.ResponseTime,
			&f.Status,
			&f.TeamID,
			&f.Msg,
			&f.Username,
			&f.ExploitName,
			&f.DeletedAt,
		); err != nil {
			return nil, err
		}
		flags = append(flags, f)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return flags, nil
}

func (s *Store) CountFlags(ctx context.Context, q FlagsQuery) (int64, error) {
	sb, args := buildQuery(CountFlagsQuery, q)

	var count int64
	if sb == nil {
		return 0, sql.ErrNoRows
	}
	stmt, err := s.db.PrepareContext(ctx, sb.String())
	if err != nil {
		return 0, err
	}
	defer stmt.Close()
	err = stmt.QueryRowContext(ctx, args...).Scan(&count)
	return count, err
}
