package server

import (
	"time"

	"github.com/ByteTheCookies/CookieFarm/internal/server/config"
	"github.com/ByteTheCookies/CookieFarm/internal/server/sqlite"
	"github.com/ByteTheCookies/CookieFarm/internal/server/websockets"
	"github.com/ByteTheCookies/CookieFarm/pkg/logger"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	jwtware "github.com/gofiber/jwt/v3"
)

// RegisterRoutes configures all routes and middlewares of the Fiber app,
// including CORS policies, public and protected API endpoints, and view rendering routes.
func RegisterRoutes(app *fiber.App) {
	if len(config.Secret) == 0 {
		logger.Log.Fatal().Msg("JWT secret not configured")
	}

	// Enable CORS with dynamic origins from environment variable.
	// Useful for allowing access from web dashboards or dev clients.
	app.Use(cors.New(cors.Config{
		AllowOrigins:     config.GetEnv("ALLOW_ORIGINS", "http://localhost:8080"),
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS,PATCH",
		AllowHeaders:     "Accept,Authorization,Content-Type",
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// ------------------ VIEW ROUTES ------------------

	// These endpoints render HTML templates (SSR views)
	View := app.Group("/")
	View.Get("/", HandleIndexPage)
	View.Get("/dashboard", HandleIndexPage)
	View.Get("/login", HandleLoginPage)
	View.Get("/flags/:limit", HandlePartialsFlags)
	View.Get("/pagination/:limit", HandlePartialsPagination)

	// ------------------ PUBLIC API ------------------

	// Public endpoints that do not require authentication
	publicAPI := app.Group("/api/v1")
	publicAPI.Get("/", GetStatus) // Simple status check
	publicAPI.Post("/auth/login", NewLimiter(), HandleLogin)
	publicAPI.Get("/protocols", HandleGetProtocols)

	// ------------------ PRIVATE API ------------------

	// Protected routes, accessible only with a valid JWT token.
	// The token can be passed via Authorization header or cookie.
	privateAPI := app.Group("/api/v1", jwtware.New(jwtware.Config{
		SigningKey:  config.Secret,
		TokenLookup: "header:Authorization,cookie:token",
	}))
	privateAPI.Get("/stats", HandleGetStats)
	privateAPI.Get("/flags", HandleGetAllFlags)
	privateAPI.Get("/flags/:limit", HandleGetPaginatedFlags)
	privateAPI.Get("/config", HandleGetConfig)
	privateAPI.Get("/health", HealthHandler)
	privateAPI.Post("/submit-flags", HandlePostFlags)
	privateAPI.Post("/submit-flag", HandlePostFlag)
	privateAPI.Post("/submit-flags-standalone", HandlePostFlagsStandalone)
	privateAPI.Delete("/delete-flag", HandleDeleteFlag)
	privateAPI.Post("/config", HandlePostConfig)

	websocketsAPI := app.Group("/ws")
	websockets.GlobalManager = websockets.NewManager()
	websocketsAPI.Get("/", websockets.GlobalManager.ServeWS)
}

// GetStatus is a simple public endpoint used to check if the server is online.
func GetStatus(c *fiber.Ctx) error {
	resp := fiber.Map{
		"message": "The cookie is up!!",
		"time":    time.Now().Format(time.RFC3339),
	}
	return c.JSON(resp)
}

// HealthHandler returns a basic health check result from the database layer.
func HealthHandler(c *fiber.Ctx) error {
	return c.JSON(sqlite.Health())
}
