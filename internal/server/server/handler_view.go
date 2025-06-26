package server

import (
	"math"

	"github.com/ByteTheCookies/CookieFarm/internal/server/config"
	"github.com/ByteTheCookies/CookieFarm/internal/server/sqlite"
	"github.com/ByteTheCookies/CookieFarm/pkg/logger"
	"github.com/gofiber/fiber/v2"
)

const windowSize = 5

func MakePagination(current, totalPages int) []int {
	pages := []int{}

	half := windowSize / 2
	start := current - half
	end := current + half

	if start < 0 {
		end += -start
		start = 0
	}

	if end > totalPages-1 {
		start -= (end - (totalPages - 1))
		end = totalPages - 1
	}
	if start < 0 {
		start = 0
	}

	for i := start; i <= end; i++ {
		pages = append(pages, i)
	}
	return pages
}

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
	data := ViewParamsDashboard{
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

	totalFlags, err := sqlite.FlagsNumber(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Error retrieving flag count")
	}

	offset := c.QueryInt("offset", config.DefaultOffset)

	totalPages := int(math.Ceil(float64(totalFlags) / float64(limit)))
	current := offset / limit
	pageList := MakePagination(current, totalPages)

	data := ViewParamsPagination{
		Pagination: Pagination{
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

	offset := c.QueryInt("offset", config.DefaultOffset)
	logger.Log.Debug().Int("offset", offset).Int("limit", limit).Msg("Paginated flags request")

	flags, err := sqlite.GetPagedFlags(uint(limit), uint(offset))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Error retrieving flags")
	}

	logger.Log.Debug().Int("n_flags", len(flags)).Msg("Paginated flags response")

	data := ViewParamsFlags{
		Flags: flags,
	}

	return c.Render("partials/flags_rows", data)
}
