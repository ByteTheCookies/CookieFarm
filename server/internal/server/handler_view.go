package server

import (
	"math"
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

func ValidateSortQuery(value bool, threshold string, field string) database.SortQuery {
	if field == threshold {
		return database.SortQuery{}
	}
	return database.SortQuery{
		SortName:  field,
		SortValue: value,
	}
}

func parseFilter(field string) string {
	return strings.Split(field, "-")[1]
}

func parseSort(field string) string {
	return strings.Split(field, "-")[1]
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
	sorts := []database.SortQuery{}
	searchQuery := []database.SearchQuery{}
	// Get all query params list
	queryParams := c.Queries()

	for key, value := range queryParams {
		if strings.HasPrefix(key, "filter-") {
			filter := ValidateFiltersQuery(
				value,
				"-1",
				parseFilter(key),
			)
			if !filter.IsEmpy() {
				filters = append(filters, filter)
			}
		}

		if strings.HasPrefix(key, "sort-") {
			sort := ValidateSortQuery(
				value == "true",
				"-1",
				parseFilter(key),
			)
			if !sort.IsEmpy() {
				sorts = append(sorts, sort)
			}
		}

	}

	search := ValidateSearchQuery(
		c.Query("search", ""),
		"",
		"msg",
	)
	if !search.IsEmpy() {
		searchQuery = append(searchQuery, search)
	}

	logger.Log.Debug().Msg("GG")

	finalQuery := database.CustomQuery{
		FilterQuery: filters,
		SortQuery:   sorts,
		SearchQuery: searchQuery,
	}

	logger.Log.Debug().Msgf("Final Query: %+v", finalQuery)
	// Loggami tutto in modo approfondito
	logger.Log.Debug().Msgf("Filters: %+v", filters)
	logger.Log.Debug().Msgf("Sorts: %+v", sorts)
	logger.Log.Debug().Msgf("Search: %+v", searchQuery)
	logger.Log.Debug().Msgf("QueryParams: %+v", queryParams)
	logger.Log.Debug().Msgf("Limit: %d", limit)
	logger.Log.Debug().Msgf("Offset: %d", offset)
	// Le len

	logger.Log.Debug().Msgf("Filters len: %d", len(finalQuery.FilterQuery))
	logger.Log.Debug().Msgf("Sorts len: %d", len(finalQuery.SortQuery))
	logger.Log.Debug().Msgf("Search len: %d", len(finalQuery.SearchQuery))

	var flags []models.Flag
	if len(finalQuery.FilterQuery) == 0 && len(finalQuery.SearchQuery) == 0 && len(finalQuery.SortQuery) == 0 {
		logger.Log.Debug().Msg("No filters, search or sort query provided")
		flags, err = database.GetPagedFlags(uint(limit), uint(offset))
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Error retrieving flags")
		}
	} else {
		logger.Log.Debug().Msg("Filters, search or sort query provided")
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
