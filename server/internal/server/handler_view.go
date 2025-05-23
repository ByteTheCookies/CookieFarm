package server

import (
	"math"
	"strconv"
	"strings"

	"github.com/ByteTheCookies/cookieserver/internal/config"
	"github.com/ByteTheCookies/cookieserver/internal/database"
	"github.com/ByteTheCookies/cookieserver/internal/logger"
	"github.com/ByteTheCookies/cookieserver/internal/models"
	"github.com/ByteTheCookies/cookieserver/internal/utils"
	"github.com/gofiber/fiber/v2"
)

// HandleIndexPage renders the main dashboard page.
// It checks the cookie-based authentication and sets the default pagination limit.
func HandleIndexPage(c *fiber.Ctx) error {
	if err := CookieAuthMiddleware(c); err != nil {
		return err
	}

	limit := c.QueryInt("limit", config.DefaultLimit)
	if limit <= 0 {
		limit = config.DefaultLimit
	}

	logger.Log.Debug().Int("Limit", limit).Msg("Index page request")
	data := models.ViewParamsDashboard{
		Limit: limit,
	}
	return c.Render("pages/dashboard", data, "layouts/main")
}

// HandleLoginPage renders the login page.
func HandleLoginPage(c *fiber.Ctx) error {
	return c.Render("pages/login", map[string]any{}, "layouts/main")
}

// HandlePartialsPagination renders only the pagination component as a partial view.
// It computes the current page and the total number of pages based on the flags count.
func HandlePartialsPagination(c *fiber.Ctx) error {
	if err := CookieAuthMiddleware(c); err != nil {
		return err
	}

	limit, err := c.ParamsInt("limit", config.DefaultLimit)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid 'limit' parameter")
	}
	logger.Log.Debug().Int("limit", limit).Msg("Paginated flags request")

	totalFlags, err := database.FlagsNumber(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Error retrieving flag count")
	}

	offset := c.QueryInt("offset", config.DefaultOffeset)

	totalPages := int(math.Ceil(float64(totalFlags) / float64(limit)))
	current := offset / limit
	pageList := utils.MakePagination(current, totalPages)

	data := models.ViewParamsPagination{
		Pagination: models.Pagination{
			Limit:    limit,
			Pages:    totalPages,
			Current:  current,
			HasPrev:  current > 0,
			HasNext:  current < totalPages-1,
			PageList: pageList,
		},
	}

	return c.Render("partials/pagination", data, "layouts/main")
}

func ValidateFiltersQuery(value string, threshold string, field string) database.FilterQuery {
	if value == threshold {
		return database.FilterQuery{}
	}
	return database.FilterQuery{
		FilterName:  field,
		FilterValue: value,
	}
}

func ValidateSearchQuery(value string, threshold string, field string) database.SearchQuery {
	if value == threshold {
		return database.SearchQuery{}
	}
	return database.SearchQuery{
		SearchName:  field,
		SearchValue: value,
	}
}

// HandlePartialsFlags renders only the flags rows as a partial view.
// It fetches a limited and paginated list of flags from the database.
func HandlePartialsFlags(c *fiber.Ctx) error {
	if err := CookieAuthMiddleware(c); err != nil {
		return err
	}

	limit, err := c.ParamsInt("limit", config.DefaultLimit)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid 'limit' parameter")
	}
	offset := c.QueryInt("offset", config.DefaultOffeset)
	logger.Log.Debug().Int("offset", offset).Int("limit", limit).Msg("Paginated flags request")

	filters := []database.FilterQuery{}

	filterTeamID := ValidateFiltersQuery(
		strconv.Itoa(c.QueryInt("filter-teamid", -1)),
		"-1",
		"team_id",
	)
	if !filterTeamID.IsEmpy() {
		filters = append(filters, filterTeamID)
	}

	filterService := ValidateFiltersQuery(
		strconv.Itoa(c.QueryInt("filter-service", 0)),
		"-1",
		"port_service",
	)
	if !filterService.IsEmpy() {
		filters = append(filters, filterService)
	}

	filterStatus := ValidateFiltersQuery(
		c.Query("filter-status", ""),
		"",
		"status",
	)
	if !filterStatus.IsEmpy() {
		filters = append(filters, filterStatus)
	}

	filterResponseTime := ValidateFiltersQuery(
		strconv.Itoa(c.QueryInt("filter-response-time", 0)),
		"0",
		"response_time",
	)
	if !filterResponseTime.IsEmpy() {
		filters = append(filters, filterResponseTime)
	}

	filterSubmitTime := ValidateFiltersQuery(
		strconv.Itoa(c.QueryInt("filter-submit-time", 0)),
		"0",
		"submit_time",
	)
	if !filterSubmitTime.IsEmpy() {
		filters = append(filters, filterSubmitTime)
	}

	search := ValidateSearchQuery(
		c.Query("search", ""),
		"",
		"msg",
	)

	searchQuery := []database.SearchQuery{search}

	sortQuery := []database.SortQuery{}
	for order := range strings.SplitSeq(c.Query("order", ""), ",") {
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

	var flags []models.Flag
	if len(finalQuery.FilterQuery) == 0 && len(finalQuery.SearchQuery) == 0 && len(finalQuery.SortQuery) == 0 {
		flags, err = database.GetPagedFlags(uint(limit), uint(offset))
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Error retrieving flags")
		}
	} else {
		flags, err = database.GetCustomFlags(finalQuery, uint(limit), uint(offset))
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Error retrieving flags")
		}
	}

	logger.Log.Debug().Int("n_flags", len(flags)).Msg("Paginated flags response")

	data := models.ViewParamsFlags{
		Flags: flags,
	}

	return c.Render("partials/flags_rows", data)
}
