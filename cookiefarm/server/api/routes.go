package api

import (
	"logger"
	"time"

	"server/config"
	"server/core"
	"server/database"
	"server/websockets"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	jwtware "github.com/gofiber/jwt/v3"
)

type Handler struct {
	store  *database.Store
	runner *core.Runner
}

func NewHandler(s *database.Store, r *core.Runner) *Handler {
	return &Handler{store: s, runner: r}
}

// RegisterRoutes configures all routes and middlewares of the Fiber app,
// including CORS policies, public and protected API endpoints, and view rendering routes.
func (h *Handler) RegisterRoutes(app *fiber.App) {
	if len(config.Secret) == 0 {
		logger.Log.Fatal().Msg("JWT secret not configured")
	}

	// Enable CORS with dynamic origins from environment variable.
	// Useful for allowing access from web dashboards or dev clients.
	app.Use(cors.New(cors.Config{
		AllowOrigins:     config.GetEnv("ALLOW_ORIGINS", "http://localhost:8080,http://localhost:3000"),
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
	publicAPI.Get("/auth/verify", HandleVerify)
	publicAPI.Get("/protocols", h.HandleGetProtocols)

	// ------------------ PRIVATE API ------------------

	// Protected routes, accessible only with a valid JWT token.
	// The token can be passed via Authorization header or cookie.
	privateAPI := app.Group("/api/v1", jwtware.New(jwtware.Config{
		SigningKey:  config.Secret,
		TokenLookup: "header:Authorization,cookie:token",
	}))
	privateAPI.Get("/stats", h.HandleGetStats)
	privateAPI.Get("/flags", h.HandleGetAllFlags)
	privateAPI.Get("/flags/:limit", h.HandleGetPaginatedFlags)
	privateAPI.Get("/config", h.HandleGetConfig)
	privateAPI.Post("/config", h.HandlePostConfig)
	privateAPI.Post("/submit-flags", h.HandlePostFlags)
	privateAPI.Post("/submit-flag", h.HandlePostFlag)
	privateAPI.Post("/submit-flags-standalone", h.HandlePostFlagsStandalone)
	privateAPI.Delete("/delete-flag", h.HandleDeleteFlag)

	websockets.GlobalManager = websockets.NewManager()
	app.Use("/ws",
		websockets.CookieAuthMiddleware,
		websockets.WebSocketUpgrade,
	)
	app.Get("/ws", websockets.GlobalManager.ServeWS())
}

// GetStatus is a simple public endpoint used to check if the server is online.
func GetStatus(c *fiber.Ctx) error {
	resp := fiber.Map{
		"message": "The cookie is up!!",
		"time":    time.Now().Format(time.RFC3339),
	}
	return c.JSON(resp)
}
