package server

import (
	"github.com/ByteTheCookies/cookieserver/internal/config"
	"github.com/ByteTheCookies/cookieserver/internal/database"
	"github.com/ByteTheCookies/cookieserver/internal/utils"
	"github.com/ByteTheCookies/cookieserver/internal/websockets"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	jwtware "github.com/gofiber/jwt/v3"
)

// RegisterRoutes configures all routes and middlewares of the Fiber app,
// including CORS policies, public and protected API endpoints, and view rendering routes.
func RegisterRoutes(app *fiber.App) {
	// Enable CORS with dynamic origins from environment variable.
	// Useful for allowing access from web dashboards or dev clients.
	app.Use(cors.New(cors.Config{
		AllowOrigins:     utils.GetEnv("ALLOW_ORIGINS", "http://localhost:8080"),
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
	privateAPI.Post("/config", HandlePostConfig)

	websocketsAPI := app.Group("/ws")
	manager := websockets.NewManager()
	websocketsAPI.All("", manager.ServeWS)
}

// GetStatus is a simple public endpoint used to check if the server is online.
func GetStatus(c *fiber.Ctx) error {
	resp := fiber.Map{
		"message": "The cookie is up!!",
	}
	return c.JSON(resp)
}

// HealthHandler returns a basic health check result from the database layer.
func HealthHandler(c *fiber.Ctx) error {
	return c.JSON(database.Health())
}
