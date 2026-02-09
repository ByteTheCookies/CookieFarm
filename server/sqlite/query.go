package sqlite

import (
	"context"
	"errors"
	"fmt"
)

// NewQueryBuilder creates a new query builder
func NewQueryBuilder(baseQuery string) *QueryBuilder {
	return &QueryBuilder{
		query:      baseQuery,
		params:     make([]any, 0),
		conditions: make([]string, 0),
		hasWhere:   false,
	}
}

// AddCondition adds a WHERE condition with parameter binding
func (qb *QueryBuilder) AddCondition(condition string, params ...any) *QueryBuilder {
	qb.conditions = append(qb.conditions, condition)
	qb.params = append(qb.params, params...)
	return qb
}

// AddSort adds ORDER BY clause with validation
func (qb *QueryBuilder) AddSort(field, direction string) *QueryBuilder {
	validFields := map[string]bool{
		"submit_time":   true,
		"response_time": true,
		"status":        true,
		"flag_code":     true,
		"service_name":  true,
		"team_id":       true,
		"port_service":  true,
	}

	validDirections := map[string]bool{
		"ASC":  true,
		"DESC": true,
	}

	if validFields[field] && validDirections[direction] {
		qb.query += fmt.Sprintf(" ORDER BY %s %s", field, direction)
	} else {
		qb.query += " ORDER BY submit_time DESC"
	}
	return qb
}

// AddPagination adds LIMIT and OFFSET
func (qb *QueryBuilder) AddPagination(limit, offset uint) *QueryBuilder {
	if limit > 0 {
		qb.query += fmt.Sprintf(" LIMIT %d", limit)
	}
	if offset > 0 {
		qb.query += fmt.Sprintf(" OFFSET %d", offset)
	}
	return qb
}

// Build finalizes the query and returns query string with parameters
func (qb *QueryBuilder) Build() (string, []any) {
	return qb.query, qb.params
}

func buildFilterConditions(opts FilterOptions) (*QueryBuilder, []any) {
	qb := NewQueryBuilder("")

	// Add status filter
	if opts.Status != "" && opts.Status != "all" {
		qb.AddCondition("status = ?", opts.Status)
	}

	// Add service filter
	if opts.ServiceName != "" {
		qb.AddCondition("service_name LIKE ?", "%"+opts.ServiceName+"%")
	}

	// Add team filter
	if opts.TeamID != "" {
		qb.AddCondition("team_id = ?", opts.TeamID)
	}

	// Add search
	if opts.Search != "" {
		switch opts.SearchField {
		case "flag_code":
			qb.AddCondition("flag_code LIKE ?", "%"+opts.Search+"%")
		case "service_name":
			qb.AddCondition("service_name LIKE ?", "%"+opts.Search+"%")
		case "exploit_name":
			qb.AddCondition("exploit_name LIKE ?", "%"+opts.Search+"%")
		case "msg":
			qb.AddCondition("msg LIKE ?", "%"+opts.Search+"%")
		case "all":
			qb.AddCondition("(flag_code LIKE ? OR service_name LIKE ? OR exploit_name LIKE ? OR msg LIKE ? OR CAST(team_id AS TEXT) LIKE ?)",
				"%"+opts.Search+"%", "%"+opts.Search+"%", "%"+opts.Search+"%", "%"+opts.Search+"%", "%"+opts.Search+"%")
		default:
			qb.AddCondition("flag_code LIKE ?", "%"+opts.Search+"%")
		}
	}

	_, params := qb.Build()
	return qb, params
}

func CountFilteredFlags(opts FilterOptions) (int, error) {
	qb := NewQueryBuilder("SELECT COUNT(*) FROM flags")
	filterQb, params := buildFilterConditions(opts)

	// Apply WHERE conditions for count query
	if len(filterQb.conditions) > 0 {
		qb.query += " WHERE " + filterQb.conditions[0]
		for i := 1; i < len(filterQb.conditions); i++ {
			qb.query += " AND " + filterQb.conditions[i]
		}
	}
	qb.params = append(qb.params, params...)

	query, finalParams := qb.Build()

	conn := DBPool.Get(context.Background())
	if conn == nil {
		return 0, errors.New("no available SQLite connection")
	}
	defer DBPool.Put(conn)

	stmt, err := conn.Prepare(query)
	if err != nil {
		return 0, fmt.Errorf("prepare CountFilteredFlags: %w", err)
	}
	defer stmt.Finalize()

	// Bind parameters
	stmt, err = BindParams(stmt, finalParams...)
	if err != nil {
		return 0, fmt.Errorf("bind params: %w", err)
	}

	hasRow, err := stmt.Step()
	if err != nil {
		return 0, fmt.Errorf("step: %w", err)
	}
	if !hasRow {
		return 0, errors.New("no row returned from COUNT query")
	}

	count := stmt.ColumnInt(0)
	return count, nil
}
