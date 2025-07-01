package sqlite

import (
	"context"
	"errors"
	"fmt"

	"github.com/ByteTheCookies/CookieFarm/pkg/models"
)

const (
	baseFlagQuery         = `SELECT flag_code,service_name,port_service,submit_time,response_time,status, team_id, msg, username, exploit_name FROM flags`
	queryAllFlags         = baseFlagQuery + " ORDER BY submit_time DESC"
	queryFirstNFlags      = baseFlagQuery + " ORDER BY submit_time DESC LIMIT ?"
	queryUnsubmittedFlags = baseFlagQuery + " WHERE status = 'UNSUBMITTED' ORDER BY submit_time ASC LIMIT ?"
	queryPagedFlags       = baseFlagQuery + " ORDER BY submit_time DESC LIMIT ? OFFSET ?"

	baseFlagCodeQuery         = `SELECT flag_code FROM flags`
	queryAllFlagCodes         = baseFlagCodeQuery
	queryFirstNFlagCodes      = baseFlagCodeQuery + " LIMIT ?"
	queryUnsubmittedFlagCodes = baseFlagCodeQuery + " WHERE status = 'UNSUBMITTED' LIMIT ?"
	queryPagedFlagCodes       = baseFlagCodeQuery + " LIMIT ? OFFSET ?"
)

// --------- Flag Structs ---------

// GetAllFlags retrieves all flags from the database.
func GetAllFlags() ([]models.ClientData, error) {
	return queryFlags(queryAllFlags)
}

// GetUnsubmittedFlags retrieves the first n unsubmitted flags from the database.
func GetUnsubmittedFlags(limit uint) ([]models.ClientData, error) {
	return queryFlags(queryUnsubmittedFlags, limit)
}

// GetFirstNFlags retrieves the first n flags from the database.
func GetFirstNFlags(limit uint) ([]models.ClientData, error) {
	return queryFlags(queryFirstNFlags, limit)
}

// GetPagedFlags retrieves the flags from the database starting at the given offset.
func GetPagedFlags(limit, offset uint) ([]models.ClientData, error) {
	return queryFlags(queryPagedFlags, limit, offset)
}

// --------- Flag Code Only ---------

// GetAllFlagCodeList retrieves all flag codes from the database.
func GetAllFlagCodeList() ([]string, error) {
	return queryFlagCodes(queryAllFlagCodes)
}

// GetUnsubmittedFlagCodeList retrieves the first n unsubmitted flag codes from the database.
func GetUnsubmittedFlagCodeList(limit uint) ([]string, error) {
	return queryFlagCodes(queryUnsubmittedFlagCodes, limit)
}

// GetFirstNFlagCodeList retrieves the first n flag codes from the database.
func GetFirstNFlagCodeList(limit uint) ([]string, error) {
	return queryFlagCodes(queryFirstNFlagCodes, limit)
}

// GetPagedFlagCodeList retrieves the flag codes from the database starting at the given offset.
func GetPagedFlagCodeList(limit, offset uint) ([]string, error) {
	return queryFlagCodes(queryPagedFlagCodes, limit, offset)
}

// --------- Shared query logic ---------

// queryFlags executes a query to retrieve flags from the database.
// It prepares and executes a query with the provided arguments and returns a list of flags.
func queryFlags(query string, args ...any) ([]models.ClientData, error) {
	conn := DBPool.Get(context.Background())
	if conn == nil {
		return nil, errors.New("no available SQLite connection")
	}
	defer DBPool.Put(conn)

	stmt, err := conn.Prepare(query)
	if err != nil {
		return nil, fmt.Errorf("prepare queryFlags: %w", err)
	}
	defer stmt.Finalize()

	// Bind args (1-based index)
	stmt, err = BindParams(stmt, args...)
	if err != nil {
		return nil, fmt.Errorf("bind params: %w", err)
	}

	var flags []models.ClientData
	for {
		hasRow, err := stmt.Step()
		if err != nil {
			return nil, fmt.Errorf("step: %w", err)
		}
		if !hasRow {
			break
		}

		f := models.ClientData{
			FlagCode:     stmt.ColumnText(0),
			ServiceName:  stmt.ColumnText(1),
			PortService:  uint16(stmt.ColumnInt64(2)),
			SubmitTime:   uint64(stmt.ColumnInt64(3)),
			ResponseTime: uint64(stmt.ColumnInt64(4)),
			Status:       stmt.ColumnText(5),
			TeamID:       uint16(stmt.ColumnInt(6)),
			Msg:          stmt.ColumnText(7),
			Username:     stmt.ColumnText(8),
			ExploitName:  stmt.ColumnText(9),
		}
		flags = append(flags, f)
	}

	return flags, nil
}

// queryFlagCodes executes a query to retrieve flag codes from the database.
// It prepares and executes a query with the provided arguments and returns a list of flag codes.
func queryFlagCodes(query string, args ...any) ([]string, error) {
	conn := DBPool.Get(context.Background())
	if conn == nil {
		return nil, errors.New("no available SQLite connection")
	}
	defer DBPool.Put(conn)

	stmt, err := conn.Prepare(query)
	if err != nil {
		return nil, fmt.Errorf("prepare queryFlagCodes: %w", err)
	}
	defer stmt.Finalize()

	// Bind args
	stmt, err = BindParams(stmt, args...)
	if err != nil {
		return nil, fmt.Errorf("bind params: %w", err)
	}

	var codes []string
	for {
		hasRow, err := stmt.Step()
		if err != nil {
			return nil, fmt.Errorf("step: %w", err)
		}
		if !hasRow {
			break
		}
		codes = append(codes, stmt.ColumnText(0))
	}

	return codes, nil
}
