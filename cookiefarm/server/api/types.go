package api

import (
	"server/database"
)

// ResponseFlags represents the response for the flags api
type ResponseFlags struct {
	Flags  []database.Flag `json:"flags"`
	Nflags int             `json:"n_flags"`
}

// SigninRequest from the client to the server
type SigninRequest struct {
	Username string `json:"username,omitzero"` // Username for authentication
	Password string `json:"password"`          //nolint:gosec
}

// Pagination structure for manage data in the view
type Pagination struct {
	PageList []int // List of page numbers to display in the pagination
	Pages    int   // Total number of pages
	Limit    int   // Maximum number of items per page
	Current  int   // Current page number (offset / limit)
	HasPrev  bool  // Indicates if there is a previous page
	HasNext  bool  // Indicates if there is a next page
}

// ViewParamsDashboard represents the parameters for the dashboard view
type ViewParamsDashboard struct {
	Limit int `json:"limit"` // Maximum number of items per page
}

// ViewParamsPagination represents the parameters for the pagination view
type ViewParamsPagination struct {
	Pagination Pagination // Pagination parameters
}

// ViewParamsFlags represents the parameters for the flags view
type ViewParamsFlags struct {
	Flags []database.Flag `json:"flags"` // List of flags to display
}

// ResponseSuccess represents the response for the success api
type ResponseSuccess struct {
	Message string `json:"message"` // Message for the success response
}

// ResponseError represents the response for the error api
type ResponseError struct {
	Error   string `json:"error"`   // Error message for the error response
	Details string `json:"details"` // Details for the error response
}

// SubmitFlagsRequest the struct for the requests from the client to server
type SubmitFlagsRequest struct {
	Flags []database.Flag `json:"flags"` // Flags to submit
}

// SubmitFlagRequest the struct for the requests from the client to server
type SubmitFlagRequest struct {
	Flag database.Flag `json:"flag"` // Flag to submit
}
