package sqlite

import (
	"context"
	"errors"
	"fmt"

	"github.com/ByteTheCookies/CookieFarm/pkg/logger"
)

const (
	queryDeleteFlag    = `DELETE FROM flags WHERE flag_code = ?`
	queryDeleteTTLFlag = `DELETE FROM flags WHERE response_time < strftime('%s', 'now', ?)`
)

func DeleteFlag(ctx context.Context, flag string) error {
	conn := DBPool.Get(ctx)
	if conn == nil {
		return errors.New("no available SQLite connection")
	}
	defer DBPool.Put(conn)

	stmt, err := conn.Prepare(queryDeleteFlag)
	if err != nil {
		logger.Log.Error().Err(err).Msg("Failed to prepare DELETE flag query")
		return fmt.Errorf("prepare delete: %w", err)
	}
	defer stmt.Finalize()

	stmt.BindText(1, flag)

	_, err = stmt.Step()
	if err != nil {
		logger.Log.Error().Err(err).Msg("Failed to execute DELETE flag query")
		return fmt.Errorf("step delete: %w", err)
	}

	logger.Log.Info().Str("flag", flag).Msg("Flag deleted successfully")
	return nil
}

func DeleteTTLFlag(ctx context.Context, ttl uint64) (int64, error) {
	conn := DBPool.Get(ctx)
	if conn == nil {
		return 0, errors.New("no available SQLite connection")
	}
	defer DBPool.Put(conn)

	stmt, err := conn.Prepare(queryDeleteTTLFlag)
	if err != nil {
		logger.Log.Error().Err(err).Msg("Failed to prepare DELETE TTL query")
		return 0, fmt.Errorf("prepare delete TTL: %w", err)
	}
	defer stmt.Finalize()

	stmt.BindText(1, fmt.Sprintf("-%d seconds", ttl))

	_, err = stmt.Step()
	if err != nil {
		logger.Log.Error().Err(err).Msg("Failed to execute DELETE TTL query")
		return 0, fmt.Errorf("step delete TTL: %w", err)
	}

	rowsAffected := conn.Changes()
	return int64(rowsAffected), nil
}
