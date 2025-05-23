package database

import (
	"context"
	"time"

	"github.com/ByteTheCookies/cookieserver/internal/logger"
	"github.com/ByteTheCookies/cookieserver/internal/models"
)

// CI SEI? ok guardIa

type FilterQuery struct {
	FilterName  string
	FilterValue string
}

func (f FilterQuery) IsEmpy() bool {
	if f.FilterName == "" && f.FilterValue == "" {
		return true
	}
	return false
}

type SortQuery struct {
	SortName  string
	SortValue bool // 0 asc 1 desc
}

func (s SortQuery) IsEmpy() bool {
	if s.SortName == "" {
		return true
	}
	return false
}

type SearchQuery struct {
	SearchName  string
	SearchValue string
}

func (s SearchQuery) IsEmpy() bool {
	if s.SearchName == "" && s.SearchValue == "" {
		return true
	}
	return false
}

type CustomQuery struct {
	FilterQuery []FilterQuery
	SortQuery   []SortQuery
	SearchQuery []SearchQuery
}

const (
	baseFlagQuery         = `SELECT flag_code, service_name, port_service, submit_time, response_time, status, team_id, msg FROM flags`
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
func GetAllFlags() ([]models.Flag, error) {
	return queryFlags(queryAllFlags)
}

// GetUnsubmittedFlags retrieves the first n unsubmitted flags from the database.
func GetUnsubmittedFlags(limit uint) ([]models.Flag, error) {
	return queryFlags(queryUnsubmittedFlags, limit)
}

// GetFirstNFlags retrieves the first n flags from the database.
func GetFirstNFlags(limit uint) ([]models.Flag, error) {
	return queryFlags(queryFirstNFlags, limit)
}

// GetPagedFlags retrieves the flags from the database starting at the given offset.
func GetPagedFlags(limit uint, offset uint) ([]models.Flag, error) {
	return queryFlags(queryPagedFlags, limit, offset)
}

// GetCustomFlags retrieves flags based on a custom query.
func GetCustomFlags(queryParams CustomQuery, limit uint, offset uint) ([]models.Flag, error) {
	query, err := BuildCustomQuery(queryParams)
	if err != nil {
		return nil, err
	}

	args := make([]any, 0)

	// Add filter arguments
	for _, filter := range queryParams.FilterQuery {
		if filter.FilterName != "" && filter.FilterValue != "" {
			args = append(args, filter.FilterValue)
		}
	}

	// Add search arguments
	for _, search := range queryParams.SearchQuery {
		if search.SearchName != "" && search.SearchValue != "" {
			args = append(args, "%"+search.SearchValue+"%")
		}
	}

	// Add limit and offset
	args = append(args, limit, offset)

	return queryFlags(query, args...)
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
func queryFlags(query string, args ...any) ([]models.Flag, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	logger.Log.Debug().Str("query", query).Msg("Querying flags")
	logger.Log.Debug().Msgf("args %+v", args)

	stmt, err := DB.PrepareContext(ctx, query)
	if err != nil {
		logger.Log.Error().Err(err).Str("query", query).Msg("Failed to prepare queryFlags")
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx, args...)
	if err != nil {
		logger.Log.Error().Err(err).Str("query", query).Msg("Failed to execute queryFlags")
		return nil, err
	}
	defer rows.Close()

	flags := make([]models.Flag, 0)
	for rows.Next() {
		var flag models.Flag
		if err := rows.Scan(
			&flag.FlagCode, &flag.ServiceName, &flag.PortService,
			&flag.SubmitTime, &flag.ResponseTime, &flag.Status,
			&flag.TeamID, &flag.Msg,
		); err != nil {
			logger.Log.Error().Err(err).Msg("Failed to scan row in queryFlags")
			return nil, err
		}
		flags = append(flags, flag)
	}

	return flags, nil
}

// queryFlagCodes executes a query to retrieve flag codes from the database.
// It prepares and executes a query with the provided arguments and returns a list of flag codes.
func queryFlagCodes(query string, args ...any) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	stmt, err := DB.PrepareContext(ctx, query)
	if err != nil {
		logger.Log.Error().Err(err).Str("query", query).Msg("Failed to prepare queryFlagCodes")
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx, args...)
	if err != nil {
		logger.Log.Error().Err(err).Str("query", query).Msg("Failed to execute queryFlagCodes")
		return nil, err
	}
	defer rows.Close()

	var codes []string
	codePtr := new(string)
	for rows.Next() {
		if err := rows.Scan(codePtr); err != nil {
			logger.Log.Error().Err(err).Msg("Failed to scan row in queryFlagCodes")
			return nil, err
		}
		codes = append(codes, *codePtr)
	}

	return codes, nil
}

// BuildCustomQuery builds a custom SQL query based on the provided filter, sort, and search criteria.
// It constructs a SQL query string with the specified conditions and returns it.
func BuildCustomQuery(customQuery CustomQuery) (string, error) {
	var finalQuery = baseFlagQuery

	if len(customQuery.FilterQuery) > 0 {
		finalQuery += " WHERE"
		for i, filter := range customQuery.FilterQuery {
			if i > 0 {
				finalQuery += " AND"
			}
			if filter.FilterName == "submit_time" {
				finalQuery += " unixepoch('now') - response_time <= ?*60"
				continue
			}
			if filter.FilterName == "response_time" {
				finalQuery += " unixepoch('now') - submit_time <= ?*60"
				continue
			}
			finalQuery += " " + filter.FilterName + " = ?"
		}
	}

	if len(customQuery.SearchQuery) > 0 {
		if len(customQuery.FilterQuery) > 0 {
			finalQuery += " AND"
		} else {
			finalQuery += " WHERE"
		}
		logger.Log.Debug().Str("search", customQuery.SearchQuery[0].SearchValue).Msg("Search query")
		logger.Log.Debug().Str("search", customQuery.SearchQuery[0].SearchName).Msg("Search query")

		finalQuery += " " + customQuery.SearchQuery[0].SearchName + " LIKE ?"
	}

	if len(customQuery.SortQuery) > 0 {
		finalQuery += " ORDER BY"
		for i, sort := range customQuery.SortQuery {
			if i > 0 {
				finalQuery += ","
			}
			finalQuery += " " + sort.SortName
			if sort.SortValue {
				finalQuery += " DESC"
			} else {
				finalQuery += " ASC"
			}
		}
	} else {
		finalQuery += " ORDER BY submit_time DESC"
	}

	finalQuery += " LIMIT ? OFFSET ?"

	logger.Log.Debug().Str("query", finalQuery).Msg("Final custom query")

	return finalQuery, nil
}
