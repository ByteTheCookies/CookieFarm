package sqlite

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ByteTheCookies/CookieFarm/pkg/logger"
	"github.com/ByteTheCookies/CookieFarm/pkg/protocols"
)

const (
	defaultBatchSize = 1000
)

// UpdateFlagsStatus aggiorna lo status e il messaggio delle flag in batch
func UpdateFlagsStatus(responses []protocols.ResponseProtocol) error {
	if len(responses) == 0 {
		return nil
	}

	batchSize := defaultBatchSize
	for start := 0; start < len(responses); start += batchSize {
		end := start + batchSize
		if end > len(responses) {
			end = len(responses)
		}
		batch := responses[start:end]
		if err := updateFlagsBatch(batch); err != nil {
			return err
		}
	}
	return nil
}

// updateFlagsBatch updates the status and message of flags in the database in batch.
func updateFlagsBatch(batch []protocols.ResponseProtocol) error {
	if len(batch) == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	conn := DBPool.Get(ctx)
	if conn == nil {
		return fmt.Errorf("no available SQLite connection")
	}
	defer DBPool.Put(conn)

	placeholders := make([]string, 0, len(batch))
	args := make([]any, 0, len(batch)*3)

	now := uint64(time.Now().Unix())

	for _, r := range batch {
		placeholders = append(placeholders, "(?, ?, ?)")
		args = append(args, r.Flag, r.Status, r.Msg)
	}

	query := fmt.Sprintf(`
		WITH batch_values (flag_code, status, msg) AS (
			VALUES %s
		)
		UPDATE flags
		SET
			status = (SELECT status FROM batch_values WHERE batch_values.flag_code = flags.flag_code),
			msg = (SELECT msg FROM batch_values WHERE batch_values.flag_code = flags.flag_code),
			response_time = ?
		WHERE flag_code IN (SELECT flag_code FROM batch_values)
	`, strings.Join(placeholders, ","))

	args = append(args, now)

	stmt, err := conn.Prepare(query)
	if err != nil {
		return fmt.Errorf("prepare updateFlagsBatch: %w", err)
	}
	defer stmt.Finalize()

	stmt, err = BindParams(stmt, args...)
	if err != nil {
		return fmt.Errorf("bind params updateFlagsBatch: %w", err)
	}

	for {
		hasRow, err := stmt.Step()
		if err != nil {
			return fmt.Errorf("step updateFlagsBatch: %w", err)
		}
		if !hasRow {
			break
		}
	}

	logger.Log.Debug().
		Int("count", len(batch)).
		Msg("Updated statuses for flag batch")

	return nil
}
