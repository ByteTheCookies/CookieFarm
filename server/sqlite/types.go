package sqlite

// FilterOptions represents query filter options
type FilterOptions struct {
	Status      string
	ServiceName string
	TeamID      string
	Search      string
	SearchField string
	SortField   string
	SortDir     string
	Limit       uint
	Offset      uint
}

// QueryBuilder for building safe SQL queries with parameter binding
type QueryBuilder struct {
	query      string
	params     []any
	conditions []string
	hasWhere   bool
}
