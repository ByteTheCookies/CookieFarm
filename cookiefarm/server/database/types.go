package database

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
