package api

import (
	"database/sql"
	"logger"
	"models"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"server/ckp"
	"server/config"
	"server/database"
	"server/internal/exploit"

	json "github.com/bytedance/sonic"
	"github.com/golang-jwt/jwt/v4"

	"github.com/gofiber/fiber/v3"
)

const flagCheckerHostNotConfigureWarnMessage = "Flagchecker host not configured"

func sqlNullFloatToInt64(value sql.NullFloat64) int64 {
	if !value.Valid {
		return 0
	}

	return int64(value.Float64)
}

func percentage(part, total int64) float64 {
	if total <= 0 {
		return 0
	}

	return (float64(part) / float64(total)) * 100
}

// ---------- GET ----------------

// HandleGetConfig returns the current shared configuration of the server.
//
// @Summary Get shared config
// @Description Returns current shared configuration used by clients/exploit runtime.// @Tags config
// @Produce json
// @Security CookieAuth
// @Success 200 {object} ResponseSharedConfig
// @Router /config [get]
func (h *Handler) HandleGetConfig(c fiber.Ctx) error {
	return c.JSON(h.config.GetShared())
}

// HandleGetConfig returns the current full configuration of the server.
//
// @Summary Get full config
// @Description Returns current server and shared configuration.
// @Tags config
// @Produce json
// @Security CookieAuth
// @Success 200 {object} ResponseSharedConfig
// @Router /config [get]
func (h *Handler) HandleGetFullConfig(c fiber.Ctx) error {
	return c.JSON(h.config.GetFullConfig())
}

// HandleGetAllFlags retrieves and returns all the stored flags.
//
// @Summary List flags
// @Description Returns all stored flags.
// @Tags flags
// @Produce json
// @Security CookieAuth
// @Success 200 {object} ResponseFlags
// @Failure 500 {object} ResponseError
// @Router /flags [get]
func (h *Handler) HandleGetAllFlags(c fiber.Ctx) error {
	flags, err := h.store.Queries.GetAllFlags(c.RequestCtx())
	if err != nil {
		logger.Log.Error().Err(err).Msg("Failed to fetch all flags")
		return c.Status(fiber.StatusInternalServerError).JSON(ResponseError{Error: err.Error()})
	}
	if flags == nil {
		flags = []database.Flag{}
	}
	data := ResponseFlags{
		Nflags: int64(len(flags)),
		Flags:  flags,
	}
	return c.JSON(data)
}

// HandleGetStats returns statistics about the server state.
// Returns aggregated statistics for flags. The response body contains a JSON object
// with the key "flags_stats" that maps to an array of aggregated rows (grouped
// statistics returned by the database query). Each element typically contains
// fields such as service name, status and count depending on the underlying
// query implementation.
//
// @Summary Get stats
// @Description Returns aggregated server/flags statistics (grouped counts by service/status).
// @Tags stats
// @Produce json
// @Security CookieAuth
// @Success 200 {object} map[string]any "{"flags_stats": [{"service":"svc","status":1,"count":10}, ...] }"
// @Failure 500 {object} ResponseError
// @Router /stats [get]
func (h *Handler) HandleGetStats(c fiber.Ctx) error {
	rows, err := h.store.Queries.FlagsStats(c.RequestCtx())
	if err != nil {
		logger.Log.Error().Err(err).Msg("Failed to fetch flags stats")
		return c.Status(fiber.StatusInternalServerError).JSON(ResponseError{Error: err.Error()})
	}

	if rows == nil {
		return c.JSON(fiber.Map{"flags_stats": []any{}})
	}

	return c.JSON(fiber.Map{"flags_stats": rows})
}

// HandleGetChartStats returns aggregated chart data without returning raw flags.
//
// @Summary Get chart stats
// @Description Returns tick-bucket and exploit-share aggregates for charts.
// @Tags stats
// @Produce json
// @Security CookieAuth
// @Param tick_seconds query int false "Tick bucket size in seconds"
// @Success 200 {object} ResponseChartStats
// @Failure 500 {object} ResponseError
// @Router /stats/charts [get]
func (h *Handler) HandleGetChartStats(c fiber.Ctx) error {
	tickSeconds := fiber.Query[int](c, "tick_seconds", 60)
	if tickSeconds <= 0 {
		tickSeconds = 60
	}

	tickRows, err := h.store.Queries.FlagsTickStats(c.RequestCtx(), uint64(tickSeconds))
	if err != nil {
		logger.Log.Error().Err(err).Msg("Failed to fetch flag tick stats")
		return c.Status(fiber.StatusInternalServerError).JSON(ResponseError{Error: err.Error()})
	}

	tickSeries := make([]FlagTickPoint, 0, len(tickRows))
	for _, row := range tickRows {
		tickSeries = append(tickSeries, FlagTickPoint{
			Timestamp: row.Bucket * int64(tickSeconds),
			Total:     row.Total,
			Queued:    sqlNullFloatToInt64(row.Queued),
			Accepted:  sqlNullFloatToInt64(row.Accepted),
			Denied:    sqlNullFloatToInt64(row.Denied),
			Error:     sqlNullFloatToInt64(row.Error),
			Invalid:   sqlNullFloatToInt64(row.Invalid),
		})
	}

	exploitRows, err := h.store.Queries.FlagsExploitShare(c.RequestCtx())
	if err != nil {
		logger.Log.Error().Err(err).Msg("Failed to fetch flag exploit share")
		return c.Status(fiber.StatusInternalServerError).JSON(ResponseError{Error: err.Error()})
	}

	var totalFlags int64
	for _, row := range exploitRows {
		totalFlags += row.Value
	}

	exploitShare := make([]FlagExploitShare, 0, len(exploitRows))
	for _, row := range exploitRows {
		exploitShare = append(exploitShare, FlagExploitShare{
			Name:       filepath.Base(row.ExploitName),
			Value:      row.Value,
			Percentage: percentage(row.Value, totalFlags),
		})
	}

	exploitTickRows, err := h.store.Queries.FlagsExploitTickStats(c.RequestCtx(), uint64(tickSeconds))
	if err != nil {
		logger.Log.Error().Err(err).Msg("Failed to fetch flag exploit tick stats")
		return c.Status(fiber.StatusInternalServerError).JSON(ResponseError{Error: err.Error()})
	}

	exploitTickSeriesByName := make(map[string]*FlagExploitTickSeries)
	exploitTickOrder := make([]string, 0)
	for _, row := range exploitTickRows {
		name := filepath.Base(row.ExploitName)
		series, exists := exploitTickSeriesByName[name]
		if !exists {
			series = &FlagExploitTickSeries{Name: name, Data: []FlagExploitTickPoint{}}
			exploitTickSeriesByName[name] = series
			exploitTickOrder = append(exploitTickOrder, name)
		}

		series.Total += row.Value
		series.Data = append(series.Data, FlagExploitTickPoint{
			Timestamp: row.Bucket * int64(tickSeconds),
			Value:     row.Value,
		})
	}

	exploitTickSeries := make([]FlagExploitTickSeries, 0, len(exploitTickOrder))
	for _, name := range exploitTickOrder {
		exploitTickSeries = append(exploitTickSeries, *exploitTickSeriesByName[name])
	}

	exploitStatusRows, err := h.store.Queries.FlagsExploitStatusStats(c.RequestCtx())
	if err != nil {
		logger.Log.Error().Err(err).Msg("Failed to fetch flag exploit status stats")
		return c.Status(fiber.StatusInternalServerError).JSON(ResponseError{Error: err.Error()})
	}

	exploitStatusPercentage := make([]FlagExploitStatusBreakdown, 0, len(exploitStatusRows))
	for _, row := range exploitStatusRows {
		queued := sqlNullFloatToInt64(row.Queued)
		accepted := sqlNullFloatToInt64(row.Accepted)
		denied := sqlNullFloatToInt64(row.Denied)
		errorCount := sqlNullFloatToInt64(row.Error)
		invalid := sqlNullFloatToInt64(row.Invalid)

		exploitStatusPercentage = append(exploitStatusPercentage, FlagExploitStatusBreakdown{
			Name:          filepath.Base(row.ExploitName),
			Total:         row.Total,
			Queued:        percentage(queued, row.Total),
			Accepted:      percentage(accepted, row.Total),
			Denied:        percentage(denied, row.Total),
			Error:         percentage(errorCount, row.Total),
			Invalid:       percentage(invalid, row.Total),
			QueuedCount:   queued,
			AcceptedCount: accepted,
			DeniedCount:   denied,
			ErrorCount:    errorCount,
			InvalidCount:  invalid,
		})
	}

	return c.JSON(ResponseChartStats{
		TickSeries:              tickSeries,
		ExploitShare:            exploitShare,
		ExploitTickSeries:       exploitTickSeries,
		ExploitStatusPercentage: exploitStatusPercentage,
		TotalFlags:              totalFlags,
	})
}

// HandleGetPaginatedFlags returns paginated and filtered flags.
//
// @Summary List flags (paginated)
// @Description Returns paginated flags with optional status/team/search filters.
// @Tags flags
// @Produce json
// @Security CookieAuth
// @Param limit path int true "Page size"
// @Param offset query int false "Offset"
// @Param status query int false "Status filter"
// @Param service query string false "Service filter"
// @Param team query string false "Team filter"
// @Param search query string false "Search text"
// @Param search_field query string false "Search field"
// @Success 200 {object} ResponseFlags
// @Failure 500 {object} ResponseError
// @Router /flags/{limit} [get]
func (h *Handler) HandleGetPaginatedFlags(c fiber.Ctx) error {
	limit := fiber.Params[int](c, "limit", config.DefaultLimit)
	if limit <= 0 {
		logger.Log.Warn().Msg("Invalid or missing limit parameter")
		limit = config.DefaultLimit
	}

	// Build filter options from query parameters
	optsStatus := int64(fiber.Query[int](c, "status", 5))
	optsService := strings.TrimSpace(c.Query("service", ""))
	teamStr := strings.TrimSpace(c.Query("team", ""))
	searchStr := strings.TrimSpace(c.Query("search", ""))
	searchField := strings.TrimSpace(c.Query("search_field", ""))

	var serviceNull sql.NullString
	if optsService != "" {
		serviceNull = sql.NullString{String: optsService, Valid: true}
	} else {
		serviceNull = sql.NullString{Valid: false}
	}

	// Parse team ID if provided
	var teamID uint16
	if teamStr != "" {
		if parsed, err := strconv.ParseUint(teamStr, 10, 16); err == nil {
			teamID = uint16(parsed)
		} else {
			logger.Log.Warn().Err(err).Msg("Invalid team parameter, ignoring")
		}
	}

	cursorTime, cursorID := database.ParseCursor(c.Query("cursor", ""))
	opts := database.FlagsQuery{
		Status:      sql.NullInt64{Int64: optsStatus, Valid: optsStatus != 5},
		TeamID:      sql.NullInt64{Int64: int64(teamID), Valid: teamID != 0},
		ServiceName: serviceNull,
		Search: sql.NullString{
			String: "%" + searchStr + "%",
			Valid:  searchStr != "",
		},
		SearchField: sql.NullString{
			String: searchField,
			Valid:  searchStr != "" && searchField != "",
		},
		Limit:      sql.NullInt64{Int64: int64(limit), Valid: true},
		CursorTime: cursorTime,
		CursorID:   cursorID,
	}

	logger.Log.Debug().
		// Int64("status", opts.Status.Int64).
		// Int64("team_id", opts.TeamID.Int64).
		Int64("limit", int64(limit)).
		Int64("cursor id", cursorID.Int64).
		Int64("cursor time", cursorTime.Int64).
		Msg("Fetching paginated flags with filters")

	flags, err := h.store.QueryFlagsParams(c.RequestCtx(), opts)
	if err != nil {
		logger.Log.Error().Err(err).Msg("Failed to fetch filtered flags")
		return c.Status(fiber.StatusInternalServerError).JSON(ResponseError{Error: err.Error()})
	}

	h.store.QueryFlagsParams(c, opts)

	var nextCursor string
	if len(flags) == limit {
		last := flags[len(flags)-1]
		nextCursor = database.EncodeCursor(int64(last.SubmitTime), last.ID)
	}

	// Get filtered count for accurate pagination
	nFlags, err := h.store.CountFlags(c.RequestCtx(), opts)
	if err != nil {
		logger.Log.Error().Err(err).Msg("Failed to count filtered flags")
		return c.Status(fiber.StatusInternalServerError).JSON(ResponseError{Error: err.Error()})
	}

	if flags == nil {
		flags = []database.Flag{}
	}

	return c.JSON(ResponseFlags{
		Nflags: nFlags,
		Next:   nextCursor,
		Flags:  flags,
	})
}

// HandleGetFlag retrieves a single flag by its ID.
//
// @Summary List protocols
// @Description Returns protocol plugins available on server.
// @Tags protocols
// @Produce json
// @Success 200 {object} map[string][]string
// @Failure 500 {object} ResponseError
// @Router /protocols [get]
func (*Handler) HandleGetProtocols(c fiber.Ctx) error {
	searchPaths := []string{
		"pkg/protocols",
		"protocols",
	}

	var protocolNames []string
	for _, exploitPath := range searchPaths {
		if protocols, err := os.ReadDir(exploitPath); err == nil {
			for _, entry := range protocols {
				if entry.IsDir() {
					protocolNames = append(protocolNames, strings.Split(entry.Name(), ".")[0])
				} else if matched, _ := filepath.Match("*.so", entry.Name()); matched {
					protocolNames = append(protocolNames, strings.Split(entry.Name(), ".")[0])
				}
			}
			break
		}
	}

	if len(protocolNames) == 0 {
		logger.Log.Error().Msg("Failed to read protocols directory")
		return c.Status(fiber.StatusInternalServerError).JSON(ResponseError{Error: "No protocols found"})
	}

	return c.JSON(fiber.Map{
		"protocols": protocolNames,
	})
}

// ---------- POST ----------------

// HandlePostFlags processes a batch of flags submitted in the request.
//
// @Summary Submit flag batch (store only)
// @Description Stores a batch of flags without immediate external checker submission.
// @Tags submit
// @Accept json
// @Produce json
// @Security CookieAuth
// @Param request body models.SubmitFlagsRequest true "Flags payload"
// @Success 200 {object} ResponseSuccess
// @Failure 422 {object} ResponseError
// @Failure 500 {object} ResponseError
// @Router /submit-flags [post]
func (h *Handler) HandlePostFlags(c fiber.Ctx) error {
	var payload models.SubmitFlagsRequest
	if err := c.Bind().Body(&payload); err != nil {
		logger.Log.Error().Err(err).Msg("Invalid SubmitFlags payload")
		return c.Status(fiber.StatusUnprocessableEntity).JSON(ResponseError{Error: err.Error()})
	}

	if err := h.store.BulkInsertFlags(c, payload.Flags); err != nil {
		logger.Log.Error().Err(err).Msg("Failed to insert flags")
		return c.Status(fiber.StatusInternalServerError).JSON(ResponseError{Error: err.Error()})
	}

	return c.JSON(ResponseSuccess{Message: "Flags submitted successfully"})
}

// HandlePostFlag processes a single flag and optionally submits it to an external checker.
//
// @Summary Submit one flag
// @Description Stores one flag and attempts checker submission immediately.
// @Tags submit
// @Accept json
// @Produce json
// @Security CookieAuth
// @Param request body models.SubmitFlagRequest true "Flag payload"
// @Success 200 {object} ResponseSuccess
// @Failure 422 {object} ResponseError
// @Failure 500 {object} ResponseError
// @Failure 503 {object} ResponseError
// @Router /submit-flag [post]
func (h *Handler) HandlePostFlag(c fiber.Ctx) error {
	var payload models.SubmitFlagRequest
	if err := c.Bind().Body(&payload); err != nil {
		logger.Log.Error().Err(err).Msg("Invalid SubmitFlag payload")
		return c.Status(fiber.StatusUnprocessableEntity).JSON(ResponseError{Error: err.Error()})
	}

	if err := h.store.Queries.AddFlag(c.RequestCtx(), database.MapFromFlagToDBParams(payload.Flag)); err != nil {
		logger.Log.Error().Err(err).Msg("Failed to insert single flag")
		return c.Status(fiber.StatusInternalServerError).JSON(ResponseError{Error: err.Error()})
	}

	if h.config.GetURLFlagChecker() == "" {
		logger.Log.Warn().Msg(flagCheckerHostNotConfigureWarnMessage)
		return c.Status(fiber.StatusServiceUnavailable).JSON(ResponseError{
			Error: flagCheckerHostNotConfigureWarnMessage,
		})
	}

	flags := []string{payload.Flag.FlagCode}
	response, err := config.Submit(h.config.GetURLFlagChecker(), h.config.GetTeamToken(), flags)
	if err != nil {
		logger.Log.Error().Err(err).Msg("Flagchecker submission failed")
		return c.Status(fiber.StatusInternalServerError).JSON(ResponseError{
			Error:   "Failed to submit flag",
			Details: err.Error(),
		})
	}

	h.runner.UpdateFlags(response)

	return c.JSON(ResponseSuccess{Message: "Flag submitted successfully"})
}

// HandlePostFlag processes a single flag and optionally submits it to an external checker.
//
// @Summary Submit flag batch (standalone)
// @Description Stores a batch and attempts checker submission immediately.
// @Tags submit
// @Accept json
// @Produce json
// @Security CookieAuth
// @Param request body models.SubmitFlagsRequest true "Flags payload"
// @Success 200 {object} ResponseSuccess
// @Failure 422 {object} ResponseError
// @Failure 500 {object} ResponseError
// @Failure 503 {object} ResponseError
// @Router /submit-flags-standalone [post]
func (h *Handler) HandlePostFlagsStandalone(c fiber.Ctx) error {
	var payload models.SubmitFlagsRequest
	if err := c.Bind().Body(&payload); err != nil {
		logger.Log.Error().Err(err).Msg("Invalid SubmitFlag payload")
		return c.Status(fiber.StatusUnprocessableEntity).JSON(ResponseError{Error: err.Error()})
	}

	if err := h.store.BulkInsertFlags(c.RequestCtx(), payload.Flags); err != nil {
		logger.Log.Error().Err(err).Msg("Failed to insert single flag")
		return c.Status(fiber.StatusInternalServerError).JSON(ResponseError{Error: err.Error()})
	}

	if h.config.GetURLFlagChecker() == "" {
		logger.Log.Warn().Msg(flagCheckerHostNotConfigureWarnMessage)
		return c.Status(fiber.StatusServiceUnavailable).JSON(ResponseError{
			Error: flagCheckerHostNotConfigureWarnMessage,
		})
	}
	flags := make([]string, len(payload.Flags))

	for i, flag := range payload.Flags {
		flags[i] = flag.FlagCode
		if flag.FlagCode == "" {
			logger.Log.Warn().Msg("Empty flag code found, skipping submission")
			continue
		}
	}

	response, err := config.Submit(h.config.GetURLFlagChecker(), h.config.GetTeamToken(), flags)
	if err != nil {
		logger.Log.Error().Err(err).Msg("Flagchecker submission failed")
		return c.Status(fiber.StatusInternalServerError).JSON(ResponseError{
			Error:   "Failed to submit flag",
			Details: err.Error(),
		})
	}

	h.runner.UpdateFlags(response)

	return c.JSON(ResponseSuccess{Message: "Flag submitted successfully"})
}

// HandlePostConfig updates the full configuration and restarts the flag processing loop.
//
// @Summary Update full config
// @Description Updates server/shared configuration and restarts background runner loops.
// @Tags config
// @Accept json
// @Produce json
// @Security CookieAuth
// @Param request body UpdateConfigRequest true "Full config payload"
// @Success 200 {object} ResponseSuccess
// @Failure 422 {object} ResponseError
// @Failure 500 {object} ResponseError
// @Router /config [post]
func (h *Handler) HandlePostConfig(c fiber.Ctx) error {
	var payload UpdateConfigRequest
	if err := c.Bind().Body(&payload); err != nil {
		logger.Log.Error().Err(err).Msg("Failed to parse config payload")
		return c.Status(fiber.StatusUnprocessableEntity).JSON(ResponseError{
			Error:   "Invalid config payload",
			Details: err.Error(),
		})
	}

	nextConfig := payload.Config
	nextConfig.Configured = true
	nextConfig.Shared.Configured = true

	h.config.SetFullConfig(nextConfig)
	h.config.SetConfigured(true)

	h.runner.Submission()

	cfgJSON, err := json.Marshal(h.config.GetShared())
	if err != nil {
		logger.Log.Error().Err(err).Msg("Failed to marshal config")
		return c.Status(fiber.StatusInternalServerError).JSON(ResponseError{
			Error:   "Failed to marshal config",
			Details: err.Error(),
		})
	}

	for _, conn := range h.connections.GetAll() {
		ckp.HandlerConfig(conn, append(cfgJSON, byte('\n')))
	}

	return c.JSON(ResponseSuccess{Message: "Configuration updated successfully"})
}

// HandleDeleteFlag deletes a flag by its ID.
//
// @Summary Delete flag
// @Description Deletes one flag by `flag` query parameter (flag code).
// @Tags flags
// @Produce json
// @Security CookieAuth
// @Param flag query string true "Flag code"
// @Success 200 {object} ResponseSuccess
// @Failure 400 {object} ResponseError
// @Failure 500 {object} ResponseError
// @Router /delete-flag [delete]
func (h *Handler) HandleDeleteFlag(c fiber.Ctx) error {
	flagID := c.Query("flag")
	if flagID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(ResponseError{
			Error: "Missing flag ID",
		})
	}

	if err := h.store.Queries.DeleteFlagByCode(c.RequestCtx(), flagID); err != nil {
		logger.Log.Error().Err(err).Msg("Failed to delete flag")
		return c.Status(fiber.StatusInternalServerError).JSON(ResponseError{
			Error: "Failed to delete flag",
		})
	}

	return c.JSON(ResponseSuccess{Message: "Flag deleted successfully"})
}

// @Summary List exploits
// @Description Returns all stored exploits.
// @Tags exploits
// @Produce json
// @Security CookieAuth
// @Success 200 {object} map[string]any
// @Failure 500 {object} ResponseError
// @Router /exploits [get]
func (h *Handler) HandleGetExploits(c fiber.Ctx) error {
	exploits, err := h.store.Queries.GetAllExploits(c.RequestCtx())
	if err != nil {
		logger.Log.Error().Err(err).Msg("Failed to fetch exploits")
		return c.Status(fiber.StatusInternalServerError).JSON(ResponseError{
			Error: "Failed to fetch exploits",
		})
	}

	if exploits == nil {
		exploits = []database.Exploit{}
	}

	return c.JSON(fiber.Map{
		"exploits": exploits,
		"count":    len(exploits),
	})
}

// @Summary Get exploit by name
// @Description Returns exploit(s) with content by name.
// @Tags exploits
// @Produce json
// @Security CookieAuth
// @Param name path string true "Exploit name"
// @Success 200 {array} ExploitWithContent
// @Failure 400 {object} ResponseError
// @Failure 404 {object} ResponseError
// @Failure 500 {object} ResponseError
// @Router /exploit/{name} [get]
func (h *Handler) HandleGetExploit(c fiber.Ctx) error {
	exploitName := c.Params("name")
	if exploitName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(ResponseError{
			Error: "Missing exploit name",
		})
	}

	exploits, err := h.store.Queries.GetExploitsByName(c.RequestCtx(), exploitName)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ResponseError{
			Error: "Failed to fetch exploit",
		})
	}

	if len(exploits) == 0 {
		return c.Status(fiber.StatusNotFound).JSON(ResponseError{
			Error: "Exploit not found",
		})
	}

	exploitsWContent := exploit.BuildExploitPayload(exploits)
	return c.JSON(exploitsWContent)
}

// @Summary Upload exploit
// @Description Uploads a new exploit file.
// @Tags exploits
// @Accept multipart/form-data
// @Produce json
// @Security CookieAuth
// @Param file formData file true "Exploit file"
// @Success 200 {object} map[string]any
// @Failure 400 {object} ResponseError
// @Failure 401 {object} ResponseError
// @Failure 413 {object} ResponseError
// @Failure 500 {object} ResponseError
// @Router /exploit [post]
func (h *Handler) HandlePostExploit(c fiber.Ctx) error {
	token := c.Cookies("token", "")
	jwtParsed, err := VerifyToken(token)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(ResponseError{Error: "Invalid token"})
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		logger.Log.Warn().Err(err).Msg("No file provided in request")
		return c.Status(fiber.StatusBadRequest).JSON(ResponseError{Error: "Missing file upload (field name: 'file')"})
	}

	fileHeader, err = exploit.SanitizeExploit(c, fileHeader)
	if err != nil {
		return c.Status(exploit.GetStatusCodeByErr(err)).JSON(ResponseError{Error: err.Error()})
	}

	username := jwtParsed.Claims.(jwt.MapClaims)["username"].(string)
	exploitS, err := exploit.CreateExploit(c, h.store, fileHeader, username)
	if err != nil {
		return c.Status(exploit.GetStatusCodeByErr(err)).JSON(ResponseError{Error: err.Error()})
	}

	err = h.store.Queries.CreateExploit(c, exploitS)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "cannot save exploit metadata"})
	}

	return c.JSON(fiber.Map{
		"message":      "uploaded successfully",
		"exploit_name": fileHeader.Filename,
		"hash":         exploitS.Hash,
		"version":      exploitS.Version,
	})
}

// @Summary Delete exploit
// @Description Deletes an exploit by ID.
// @Tags exploits
// @Produce json
// @Security CookieAuth
// @Param id path int true "Exploit ID"
// @Success 200 {object} ResponseSuccess
// @Failure 400 {object} ResponseError
// @Failure 500 {object} ResponseError
// @Router /exploit/{id} [delete]
func (h *Handler) HandleDeleteExploit(c fiber.Ctx) error {
	exploitID := c.Params("id")
	if exploitID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(ResponseError{
			Error: "Missing exploit ID",
		})
	}

	id, err := strconv.ParseInt(exploitID, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ResponseError{
			Error: "Invalid exploit ID",
		})
	}
	err = h.store.Queries.DeleteExploitByID(c.RequestCtx(), id)
	if err != nil {
		logger.Log.Error().Err(err).Msg("Failed to delete exploit")
		return c.Status(fiber.StatusInternalServerError).JSON(ResponseError{
			Error: "Failed to delete exploit",
		})
	}

	return c.JSON(ResponseSuccess{Message: "Exploit deleted successfully"})
}
