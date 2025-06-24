package sqlite

import (
	"fmt"

	"github.com/ByteTheCookies/CookieFarm/pkg/logger"
)

const (
	queryDeleteFlag    = `DELETE FROM flags WHERE flag_code = ?`
	queryDeleteTTLFlag = `DELETE FROM flags WHERE response_time < strftime('%s', 'now', ?)`
)

func DeleteFlag(flag string) error {
	_, err := DB.Exec(queryDeleteFlag, flag)
	if err != nil {
		logger.Log.Error().Err(err).Msg("Failed to delete flag from DB")
		return fmt.Errorf("delete flag error: %w", err)
	}
	logger.Log.Info().Str("flag", flag).Msg("Flag deleted successfully")
	return nil
}

func DeleteTTLFlag(ttl uint64) (int64, error) {
	result, err := DB.Exec(queryDeleteTTLFlag, fmt.Sprintf("-%d seconds", ttl))
	if err != nil {
		logger.Log.Error().Err(err).Msg("Failed to delete TTL flag from DB")
		return 0, fmt.Errorf("delete TTL flag error: %w", err)
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return 0, nil
	}
	return rowsAffected, nil
}
