package server

import (
	"context"

	json "github.com/bytedance/sonic"

	"github.com/ByteTheCookies/CookieFarm/internal/server/config"
	"github.com/ByteTheCookies/CookieFarm/internal/server/core"
	"github.com/ByteTheCookies/CookieFarm/internal/server/sqlite"
	"github.com/ByteTheCookies/CookieFarm/internal/server/websockets"
	"github.com/ByteTheCookies/CookieFarm/pkg/logger"
	"github.com/ByteTheCookies/CookieFarm/pkg/models"
	"github.com/gofiber/fiber/v2"
)

// ---------- GET ----------------

// HandleGetConfig returns the current configuration of the server.
func HandleGetConfig(c *fiber.Ctx) error {
	return c.JSON(config.SharedConfig)
}

// HandleGetAllFlags retrieves and returns all the stored flags.
func HandleGetAllFlags(c *fiber.Ctx) error {
	flags, err := sqlite.GetAllFlags()
	if err != nil {
		logger.Log.Error().Err(err).Msg("Failed to fetch all flags")
		return c.Status(fiber.StatusInternalServerError).JSON(ResponseError{Error: err.Error()})
	}
	if flags == nil {
		flags = []models.ClientData{}
	}
	data := ResponseFlags{
		Nflags: len(flags),
		Flags:  flags,
	}
	return c.JSON(data)
}

// HandleGetStats returns statistics about the server state.
// Currently returns placeholders for flags and users.
func HandleGetStats(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"stats": map[string]any{
			"total_flags": 0,
			"total_users": 0,
		},
	})
}

func HandleGetPaginatedFlags(c *fiber.Ctx) error {
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

	flags, err := sqlite.GetPagedFlags(uint(limit), uint(offset))
	if err != nil {
		logger.Log.Error().Err(err).Msg("Failed to fetch paginated flags")
		return c.Status(fiber.StatusInternalServerError).JSON(ResponseError{Error: err.Error()})
	}
	if flags == nil {
		flags = []models.ClientData{}
	}
	return c.JSON(ResponseFlags{
		Nflags: len(flags),
		Flags:  flags,
	})
}

// ---------- POST ----------------

// HandlePostFlags processes a batch of flags submitted in the request.
func HandlePostFlags(c *fiber.Ctx) error {
	var payload SubmitFlagsRequest
	if err := c.BodyParser(&payload); err != nil {
		logger.Log.Error().Err(err).Msg("Invalid SubmitFlags payload")
		return c.Status(fiber.StatusUnprocessableEntity).JSON(ResponseError{Error: err.Error()})
	}

	if err := sqlite.AddFlags(payload.Flags); err != nil {
		logger.Log.Error().Err(err).Msg("Failed to insert flags")
		return c.Status(fiber.StatusInternalServerError).JSON(ResponseError{Error: err.Error()})
	}

	return c.JSON(ResponseSuccess{Message: "Flags submitted successfully"})
}

// HandlePostFlag processes a single flag and optionally submits it to an external checker.
func HandlePostFlag(c *fiber.Ctx) error {
	var payload SubmitFlagRequest
	if err := c.BodyParser(&payload); err != nil {
		logger.Log.Error().Err(err).Msg("Invalid SubmitFlag payload")
		return c.Status(fiber.StatusUnprocessableEntity).JSON(ResponseError{Error: err.Error()})
	}

	if err := sqlite.AddFlag(payload.Flag); err != nil {
		logger.Log.Error().Err(err).Msg("Failed to insert single flag")
		return c.Status(fiber.StatusInternalServerError).JSON(ResponseError{Error: err.Error()})
	}

	if config.SharedConfig.ConfigServer.HostFlagchecker == "" {
		logger.Log.Warn().Msg("Flagchecker host not configured")
		return c.Status(fiber.StatusServiceUnavailable).JSON(ResponseError{
			Error: "Flagchecker host not configured",
		})
	}

	flags := []string{payload.Flag.FlagCode}
	response, err := config.Submit(config.SharedConfig.ConfigServer.HostFlagchecker, config.SharedConfig.ConfigServer.TeamToken, flags)
	if err != nil {
		logger.Log.Error().Err(err).Msg("Flagchecker submission failed")
		return c.Status(fiber.StatusInternalServerError).JSON(ResponseError{
			Error:   "Failed to submit flag",
			Details: err.Error(),
		})
	}

	core.UpdateFlags(response)

	return c.JSON(ResponseSuccess{Message: "Flag submitted successfully"})
}

// HandlePostConfig updates the server configuration and restarts the flag processing loop.
func HandlePostConfig(c *fiber.Ctx) error {
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

	if shutdownCancel != nil {
		shutdownCancel()
	}
	ctx, cancel := context.WithCancel(context.Background())
	shutdownCancel = cancel

	go core.StartFlagProcessingLoop(ctx)

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
	msg, err := json.Marshal(event)
	if err != nil {
		logger.Log.Error().Err(err).Msg("Failed to marshal websocket event")
		return c.Status(fiber.StatusInternalServerError).JSON(ResponseError{
			Error: "Failed to marshal config event",
		})
	}

	for client := range websockets.GlobalManager.Clients {
		client.Egress <- msg
	}

	return c.JSON(ResponseSuccess{Message: "Configuration updated successfully"})
}

func HandleDeleteFlag(c *fiber.Ctx) error {
	flagID := c.Query("flag")
	if flagID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(ResponseError{
			Error: "Missing flag ID",
		})
	}

	if err := sqlite.DeleteFlag(flagID); err != nil {
		logger.Log.Error().Err(err).Msg("Failed to delete flag")
		return c.Status(fiber.StatusInternalServerError).JSON(ResponseError{
			Error: "Failed to delete flag",
		})
	}

	return c.JSON(ResponseSuccess{Message: "Flag deleted successfully"})
}
