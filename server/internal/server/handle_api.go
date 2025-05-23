package server

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/ByteTheCookies/cookieserver/internal/config"
	"github.com/ByteTheCookies/cookieserver/internal/database"
	"github.com/ByteTheCookies/cookieserver/internal/logger"
	"github.com/ByteTheCookies/cookieserver/internal/models"
	"github.com/ByteTheCookies/cookieserver/internal/websockets"
	"github.com/gofiber/fiber/v2"
)

// ---------- GET ----------------

// HandleGetConfig returns the current configuration of the server.
func HandleGetConfig(c *fiber.Ctx) error {
	return c.JSON(config.Current)
}

// HandleGetAllFlags retrieves and returns all the stored flags.
func HandleGetAllFlags(c *fiber.Ctx) error {
	flags, err := database.GetAllFlags()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ResponseError{
			Error: err.Error(),
		})
	}
	if flags == nil {
		flags = []models.Flag{}
	}
	data := models.ResponseFlags{
		Nflags: len(flags),
		Flags:  flags,
	}
	return c.JSON(data)
}

// HandleGetStats returns statistics about the server state.
// Currently returns placeholders for flags and users.
func HandleGetStats(c *fiber.Ctx) error {
	logger.Log.Debug().Msg("Stats endpoint hit")
	return c.JSON(fiber.Map{
		"stats": map[string]any{
			"total_flags": 0,
			"total_users": 0,
		},
	})
}

// HandleGetPaginatedFlags returns a paginated list of flags based on the limit and offset.
func HandleGetPaginatedFlags(c *fiber.Ctx) error {
	limit, err := c.ParamsInt("limit", config.DefaultLimit)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ResponseError{
			Error: "Invalid limit parameter",
		})
	}
	offset := c.QueryInt("offset", config.DefaultOffeset)

	flags, err := database.GetPagedFlags(uint(limit), uint(offset))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ResponseError{
			Error: err.Error(),
		})
	}

	if flags == nil {
		flags = []models.Flag{}
	}
	data := models.ResponseFlags{
		Nflags: len(flags),
		Flags:  flags,
	}

	return c.JSON(data)
}

// CAPIRE LA BEST PRACTICE PER LE QUERY
// Filtro per porta e team_id
// http://.../api/v1/flags?filter-teamid=team_id&filter-port=port
// + = DESC , - = ASC
// http://.../api/v1/flags?filter-teamid=65&filter-service=8080&filter-status=DENIED&filter-time=90&search=ciao&order=+response_time,-team_id, +submit_time
func HandleCustomQuery(c *fiber.Ctx) error {
	filterTeamID := c.QueryInt("filter-teamid", -1)
	filterService := c.QueryInt("filter-service", 0)
	filterStatus := c.Query("filter-status", "")
	filterTime := c.QueryInt("filter-time", 0)
	search := c.Query("search", "")
	order := c.Query("order", "")
	limit := c.QueryInt("limit", config.DefaultLimit)
	offset := c.QueryInt("offset", config.DefaultOffeset)

	filters := []database.FilterQuery{
		{
			FilterName:  "team_id",
			FilterValue: strconv.Itoa(filterTeamID),
		},
		{
			FilterName:  "port_service",
			FilterValue: strconv.Itoa(filterService),
		},
		{
			FilterName:  "status",
			FilterValue: filterStatus,
		},
		{
			FilterName:  "time",
			FilterValue: strconv.Itoa(filterTime),
		},
	}

	searchQuery := []database.SearchQuery{
		{
			SearchName:  "flag_code",
			SearchValue: search,
		},
	}

	sortQuery := []database.SortQuery{}
	for order := range strings.SplitSeq(order, ",") {
		mode := true
		if order[0] == '-' {
			mode = false
		}

		sortQuery = append(sortQuery, database.SortQuery{
			SortName:  order[1:],
			SortValue: mode,
		})
	}

	finalQuery := database.CustomQuery{
		FilterQuery: filters,
		SortQuery:   sortQuery,
		SearchQuery: searchQuery,
	}

	flags, err := database.GetCustomFlags(finalQuery, uint(limit), uint(offset))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ResponseError{
			Error: err.Error(),
		})
	}

	if flags == nil {
		flags = []models.Flag{}
	}
	data := models.ResponseFlags{
		Nflags: len(flags),
		Flags:  flags,
	}

	return c.JSON(data)
}

// ---------- POST ----------------

// HandlePostFlags processes a batch of flags submitted in the request.
func HandlePostFlags(c *fiber.Ctx) error {
	payload := new(models.SubmitFlagsRequest)

	if err := c.BodyParser(payload); err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).
			JSON(models.ResponseError{Error: err.Error()})
	}

	flags := payload.Flags

	if err := database.AddFlags(flags); err != nil {
		logger.Log.Error().
			Err(err).
			Msg("Failed to insert flags into DB")
		return c.Status(fiber.StatusInternalServerError).
			JSON(models.ResponseError{Error: "DB error: " + err.Error()})
	}

	payload.Flags = nil

	return c.JSON(models.ResponseSuccess{
		Message: "Flags submitted successfully",
	})
}

// HandlePostFlag processes a single flag and optionally submits it to an external checker.
func HandlePostFlag(c *fiber.Ctx) error {
	payload := new(models.SubmitFlagRequest)

	if err := c.BodyParser(payload); err != nil {
		logger.Log.Error().Err(err).Msg("Invalid SubmitFlag payload")
		return c.Status(fiber.StatusUnprocessableEntity).
			JSON(models.ResponseError{Error: err.Error()})
	}
	f := payload.Flag
	if err := database.AddFlag(f); err != nil {
		logger.Log.Error().Err(err).Msg("DB insert failed in SubmitFlag")
		return c.Status(fiber.StatusInternalServerError).
			JSON(models.ResponseError{Error: "Failed to add flag: " + err.Error()})
	}

	flags := []string{f.FlagCode}

	if config.Current.ConfigServer.HostFlagchecker == "" {
		logger.Log.Warn().Msg("Flagchecker host not configured")
		return c.Status(fiber.StatusServiceUnavailable).JSON(models.ResponseError{
			Error: "Flagchecker host not configured",
		})
	}

	response, err := config.Submit(config.Current.ConfigServer.HostFlagchecker, config.Current.ConfigServer.TeamToken, flags)
	if err != nil {
		logger.Log.Error().Err(err).Msg("Failed to submit flag to external checker")
		return c.Status(fiber.StatusInternalServerError).JSON(models.ResponseError{
			Error:   "Failed to submit flag",
			Details: err.Error(),
		})
	}

	UpdateFlags(response)

	return c.JSON(models.ResponseSuccess{
		Message: "Flag submitted successfully",
	})
}

// HandlePostConfig updates the server configuration and restarts the flag processing loop.
func HandlePostConfig(c *fiber.Ctx) error {
	var configPayload struct {
		Config models.Config `json:"config"`
	}
	if err := c.BodyParser(&configPayload); err != nil {
		logger.Log.Error().Err(err).Msg("Failed to parse config payload")
		return c.Status(fiber.StatusUnprocessableEntity).JSON(models.ResponseError{
			Error:   "Failed to parse config payload",
			Details: err.Error(),
		})
	}

	config.Current = configPayload.Config

	if shutdownCancel != nil {
		shutdownCancel()
	}
	ctx, cancel := context.WithCancel(context.Background())
	shutdownCancel = cancel

	go StartFlagProcessingLoop(ctx)

	configData, err := json.Marshal(config.Current)
	if err != nil {
		logger.Log.Error().Err(err).Msg("Failed to marshal config data")
		return c.Status(fiber.StatusInternalServerError).JSON(models.ResponseError{
			Error:   "Failed to marshal config data",
			Details: err.Error(),
		})
	}

	for client := range websockets.GlobalManager.Clients {
		event := websockets.Event{
			Type:    websockets.ConfigMessage,
			Payload: configData,
		}
		msg, err := json.Marshal(event)
		if err != nil {
			logger.Log.Error().Err(err).Msg("Failed to marshal config event")
			continue
		}
		client.Egress <- msg
	}

	return c.JSON(models.ResponseSuccess{
		Message: "Configuration updated successfully",
	})
}

func HandleDeleteFlag(c *fiber.Ctx) error {
	flagID := c.Query("flag")
	if flagID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ResponseError{
			Error: "Flag ID is required",
		})
	}

	if err := database.DeleteFlag(flagID); err != nil {
		logger.Log.Error().Err(err).Msg("Failed to delete flag")
		return c.Status(fiber.StatusInternalServerError).JSON(models.ResponseError{
			Error: "Failed to delete flag",
		})
	}

	return c.JSON(models.ResponseSuccess{
		Message: "Flag deleted successfully",
	})
}
