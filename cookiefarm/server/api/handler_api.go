package api

import (
	"logger"
	"models"
	"os"
	"path/filepath"
	"strings"

	"server/config"
	"server/controllers"
	"server/database"
	"server/websockets"

	json "github.com/bytedance/sonic"

	"github.com/gofiber/fiber/v2"
)

func MapFromDBParamsToFlag(params database.Flag) database.AddFlagParams {
	return database.AddFlagParams{
		FlagCode:     params.FlagCode,
		ServiceName:  params.ServiceName,
		TeamID:       params.TeamID,
		Username:     params.Username,
		ExploitName:  params.ExploitName,
		Status:       params.Status,
		Msg:          params.Msg,
		ResponseTime: params.ResponseTime,
		SubmitTime:   params.SubmitTime,
		PortService:  params.PortService,
	}
}

// ---------- GET ----------------

// HandleGetConfig returns the current configuration of the server.
func (h *Handler) HandleGetConfig(c *fiber.Ctx) error {
	return c.JSON(config.SharedConfig)
}

// HandleGetAllFlags retrieves and returns all the stored flags.
func (h *Handler) HandleGetAllFlags(c *fiber.Ctx) error {
	flags, err := h.store.Queries.GetAllFlags(c.Context())
	if err != nil {
		logger.Log.Error().Err(err).Msg("Failed to fetch all flags")
		return c.Status(fiber.StatusInternalServerError).JSON(ResponseError{Error: err.Error()})
	}
	if flags == nil {
		flags = []database.Flag{}
	}
	data := ResponseFlags{
		Nflags: len(flags),
		Flags:  flags,
	}
	return c.JSON(data)
}

// HandleGetStats returns statistics about the server state.
// Currently returns placeholders for flags and users.
func (h *Handler) HandleGetStats(c *fiber.Ctx) error {
	n := controllers.NewStatsController()
	return n.GetFlagStats(c)
}

func (h *Handler) HandleGetPaginatedFlags(c *fiber.Ctx) error {
	limit, err := c.ParamsInt("limit", config.DefaultLimit)
	if err != nil || limit <= 0 {
		logger.Log.Warn().Msg("Invalid or missing limit parameter")
		limit = config.DefaultLimit
	}
	offset := c.QueryInt("offset", config.DefaultOffset)
	if offset < 0 {
		logger.Log.Warn().Msg("Invalid offset parameter, using default")
		offset = config.DefaultOffset
	}

	// Build filter options from query parameters
	opts := database.FilterOptions{
		Status:      c.Query("status", ""),
		ServiceName: c.Query("service", ""),
		TeamID:      c.Query("team", ""),
		Search:      c.Query("search", ""),
		SearchField: c.Query("search_field", "flag_code"),
		SortField:   c.Query("sort_field", "submit_time"),
		SortDir:     c.Query("sort_dir", "DESC"),
		Limit:       uint(limit),
		Offset:      uint(offset),
	}

	flags, err := h.store.Queries.GetFilteredFlags(c.Context(), opts)
	if err != nil {
		logger.Log.Error().Err(err).Msg("Failed to fetch filtered flags")
		return c.Status(fiber.StatusInternalServerError).JSON(ResponseError{Error: err.Error()})
	}

	// Get filtered count for accurate pagination
	nFlags, err := h.store.Queries.CountFilteredFlags(c.Context(), opts)
	if err != nil {
		logger.Log.Error().Err(err).Msg("Failed to count filtered flags")
		return c.Status(fiber.StatusInternalServerError).JSON(ResponseError{Error: err.Error()})
	}

	if flags == nil {
		flags = []database.Flag{}
	}

	return c.JSON(ResponseFlags{
		Nflags: nFlags,
		Flags:  flags,
	})
}

// HandleGetFlag retrieves a single flag by its ID.
func (h *Handler) HandleGetProtocols(c *fiber.Ctx) error {
	searchPaths := []string{
		"pkg/protocols",
		"protocols",
	}

	var protocolNames []string
	for _, path := range searchPaths {
		if protocols, err := os.ReadDir(path); err == nil {
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
func (h *Handler) HandlePostFlags(c *fiber.Ctx) error {
	var payload SubmitFlagsRequest
	if err := c.BodyParser(&payload); err != nil {
		logger.Log.Error().Err(err).Msg("Invalid SubmitFlags payload")
		return c.Status(fiber.StatusUnprocessableEntity).JSON(ResponseError{Error: err.Error()})
	}
	for _, flag := range payload.Flags {
		if err := h.store.Queries.AddFlag(c.Context(), MapFromDBParamsToFlag(flag)); err != nil {
			logger.Log.Error().Err(err).Msg("Failed to insert flags")
			return c.Status(fiber.StatusInternalServerError).JSON(ResponseError{Error: err.Error()})
		}
	}

	return c.JSON(ResponseSuccess{Message: "Flags submitted successfully"})
}

// HandlePostFlag processes a single flag and optionally submits it to an external checker.
func (h *Handler) HandlePostFlag(c *fiber.Ctx) error {
	var payload SubmitFlagRequest
	if err := c.BodyParser(&payload); err != nil {
		logger.Log.Error().Err(err).Msg("Invalid SubmitFlag payload")
		return c.Status(fiber.StatusUnprocessableEntity).JSON(ResponseError{Error: err.Error()})
	}

	if err := h.store.Queries.AddFlag(c.Context(), MapFromDBParamsToFlag(payload.Flag)); err != nil {
		logger.Log.Error().Err(err).Msg("Failed to insert single flag")
		return c.Status(fiber.StatusInternalServerError).JSON(ResponseError{Error: err.Error()})
	}

	if config.SharedConfig.ConfigServer.URLFlagChecker == "" {
		logger.Log.Warn().Msg("Flagchecker host not configured")
		return c.Status(fiber.StatusServiceUnavailable).JSON(ResponseError{
			Error: "Flagchecker host not configured",
		})
	}

	flags := []string{payload.Flag.FlagCode}
	response, err := config.Submit(config.SharedConfig.ConfigServer.URLFlagChecker, config.SharedConfig.ConfigServer.TeamToken, flags)
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
func (h *Handler) HandlePostFlagsStandalone(c *fiber.Ctx) error {
	var payload SubmitFlagsRequest
	if err := c.BodyParser(&payload); err != nil {
		logger.Log.Error().Err(err).Msg("Invalid SubmitFlag payload")
		return c.Status(fiber.StatusUnprocessableEntity).JSON(ResponseError{Error: err.Error()})
	}

	for _, flag := range payload.Flags {
		if err := h.store.Queries.AddFlag(c.Context(), MapFromDBParamsToFlag(flag)); err != nil {
			logger.Log.Error().Err(err).Msg("Failed to insert single flag")
			return c.Status(fiber.StatusInternalServerError).JSON(ResponseError{Error: err.Error()})
		}
	}

	if config.SharedConfig.ConfigServer.URLFlagChecker == "" {
		logger.Log.Warn().Msg("Flagchecker host not configured")
		return c.Status(fiber.StatusServiceUnavailable).JSON(ResponseError{
			Error: "Flagchecker host not configured",
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

	response, err := config.Submit(config.SharedConfig.ConfigServer.URLFlagChecker, config.SharedConfig.ConfigServer.TeamToken, flags)
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

// HandlePostConfig updates the server configuration and restarts the flag processing loop.
func (h *Handler) HandlePostConfig(c *fiber.Ctx) error {
	var payload struct {
		Config models.ConfigShared `json:"config"`
	}
	if err := c.BodyParser(&payload); err != nil {
		logger.Log.Error().Err(err).Msg("Failed to parse config payload")
		return c.Status(fiber.StatusUnprocessableEntity).JSON(ResponseError{
			Error:   "Invalid config payload",
			Details: err.Error(),
		})
	}

	config.SharedConfig = payload.Config

	h.runner.Run()

	cfgJSON, err := json.Marshal(config.SharedConfig)
	if err != nil {
		logger.Log.Error().Err(err).Msg("Failed to marshal config")
		return c.Status(fiber.StatusInternalServerError).JSON(ResponseError{
			Error:   "Failed to marshal config",
			Details: err.Error(),
		})
	}

	event := websockets.Event{
		Type:    websockets.ConfigMessage,
		Payload: cfgJSON,
	}

	for client := range websockets.GlobalManager.Clients {
		client.Egress <- event
	}

	return c.JSON(ResponseSuccess{Message: "Configuration updated successfully"})
}

// HandleDeleteFlag deletes a flag by its ID.
func (h *Handler) HandleDeleteFlag(c *fiber.Ctx) error {
	flagID := c.Query("flag")
	if flagID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(ResponseError{
			Error: "Missing flag ID",
		})
	}

	if err := h.store.Queries.DeleteFlagByCode(c.Context(), flagID); err != nil {
		logger.Log.Error().Err(err).Msg("Failed to delete flag")
		return c.Status(fiber.StatusInternalServerError).JSON(ResponseError{
			Error: "Failed to delete flag",
		})
	}

	return c.JSON(ResponseSuccess{Message: "Flag deleted successfully"})
}
