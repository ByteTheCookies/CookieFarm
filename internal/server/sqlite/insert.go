package sqlite

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ByteTheCookies/CookieFarm/pkg/models"
)

// AddFlagsWithContext adds a batch of flags to the database with a custom context.
// It divides the flags into batches to insert in chunks, helping avoid hitting query parameter limits.
func AddFlagsWithContext(ctx context.Context, flags []models.ClientData) error {
	if len(flags) == 0 {
		return nil
	}

	tx, err := DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	const maxParams = 50
	perRow := 8
	maxRows := maxParams / perRow
	for i := 0; i < len(flags); i += maxRows {
		end := i + maxRows
		if end > len(flags) {
			end = len(flags)
		}
		batch := flags[i:end]

		parts := make([]string, len(batch))
		args := make([]any, 0, len(batch)*perRow)
		for j, f := range batch {
			parts[j] = "(?, ?, ?, ?, ?, ?, ?)"
			args = append(args,
				f.FlagCode, f.ServiceName, f.PortService,
				f.SubmitTime, f.ResponseTime, f.Status, f.TeamID,
			)
		}
		query := "INSERT INTO flags(flag_code,service_name,port_service,submit_time,response_time,status,team_id) VALUES " +
			strings.Join(parts, ",") +
			" ON CONFLICT(flag_code) DO NOTHING"

		if _, err := tx.ExecContext(ctx, query, args...); err != nil {
			return fmt.Errorf("batch insert: %w", err)
		}
	}

	return tx.Commit()
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
