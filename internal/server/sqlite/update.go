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

func updateFlagsBatch(batch []protocols.ResponseProtocol) error {
	if len(batch) == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	placeholders := make([]string, 0, len(batch))
	args := make([]any, 0, len(batch)*3)

	now := uint64(time.Now().Unix())

	for i, r := range batch {
		placeholders = append(placeholders, fmt.Sprintf("($%d,$%d,$%d)", i*3+1, i*3+2, i*3+3))
		args = append(args, r.Flag, r.Status, r.Msg)
	}

	query := fmt.Sprintf(`
		WITH batch_values (flag_code, status, msg) AS (
			VALUES %s
		)
		UPDATE flags
		SET
			status = batch_values.status,
			msg = batch_values.msg,
			response_time = $%d
		FROM batch_values
		WHERE flags.flag_code = batch_values.flag_code`,
		strings.Join(placeholders, ","),
		len(args)+1,
	)

	args = append(args, now)

	result, err := DB.ExecContext(ctx, query, args...)
	if err != nil {
		logger.Log.Error().Err(err).
			Int("count", len(batch)).
			Msg("Failed to execute batch update of flags status")
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	logger.Log.Debug().
		Int("count", len(batch)).
		Int64("rows_affected", rowsAffected).
		Msg("Updated statuses for flag batch")

	return nil
}
