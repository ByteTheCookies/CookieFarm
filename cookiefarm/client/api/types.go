package api

import "server/database"

// SubmitFlagsRequest the struct for the requests from the client to server
type SubmitFlagsRequest struct {
	Flags []database.Flag `json:"flags"` // Flags to submit
}

// SubmitFlagRequest the struct for the requests from the client to server
type SubmitFlagRequest struct {
	Flag database.Flag `json:"flag"` // Flag to submit
}
