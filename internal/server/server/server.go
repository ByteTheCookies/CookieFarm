// Package server initializes and configures the HTTP server for CookieFarm,
// including routing, static file serving, and debug settings.
package server

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ByteTheCookies/CookieFarm/internal/server/config"
	"github.com/ByteTheCookies/CookieFarm/internal/server/ui"
	"github.com/ByteTheCookies/CookieFarm/pkg/logger"
	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
)

// shutdownCancel is used to gracefully shut down the server from external signals (if implemented).
var shutdownCancel context.CancelFunc

// newConfig returns a configured Fiber config struct.
// It adapts settings depending on the debug mode (e.g. logging, strict routing).
func newConfig(debug bool) fiber.Config {
	views := ui.InitTemplateEngine(!debug)

	cfg := fiber.Config{
		Views:       views,
		JSONEncoder: sonic.Marshal,
		JSONDecoder: sonic.Unmarshal,
	}

	if debug {
		cfg.AppName = "CookieFarm Server (Dev)"
		cfg.DisableStartupMessage = false
		cfg.CaseSensitive = false
		cfg.StrictRouting = false
		cfg.EnablePrintRoutes = true
	} else {
		cfg.AppName = "CookieFarm Server"
		cfg.DisableStartupMessage = false
		cfg.CaseSensitive = true
		cfg.StrictRouting = true
		cfg.EnablePrintRoutes = false
	}

	cfg.Prefork = false        // Disable prefork mode (multi-process); not needed here.
	cfg.ServerHeader = "Fiber" // Custom server header.

	return cfg
}

func PrepareStatic(app *fiber.App) error {
	// Serve static assets from public folders with compression and caching

	type staticRoute struct {
		route string
		dir   string
	}

	routes := []staticRoute{
		{"/css", "./internal/server/public/css"},
		{"/js", "./internal/server/public/js"},
		{"/images", "./internal/server/public/images"},
	}

	var staticCfg fiber.Static
	if config.Cache {
		staticCfg = fiber.Static{
			Compress:      true,
			CacheDuration: 10 * time.Second,
			MaxAge:        3600,
		}
	} else {
		staticCfg = fiber.Static{
			Compress: true,
		}
	}

	for _, r := range routes {
		app.Static(r.route, r.dir, staticCfg)
	}

	return nil
}

// NewApp initializes and returns a new Fiber app instance,
// setting up static file routes, debug middleware, and template engine.
func NewApp() (*fiber.App, error) {
	staticPrefixes := []string{"/css", "/js", "/images", "/static"}
	cfg := newConfig(config.Debug)
	app := fiber.New(cfg)

	if err := PrepareStatic(app); err != nil {
		return nil, fmt.Errorf("prepare static: %w", err)
	}

	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed,
	}))

	app.Server().ReadTimeout = 10 * time.Second
	app.Server().WriteTimeout = 10 * time.Second
	app.Server().IdleTimeout = 60 * time.Second

	// Log static file requests in debug mode
	if config.Debug {
		app.Use(func(c *fiber.Ctx) error {
			for _, prefix := range staticPrefixes {
				if strings.HasPrefix(c.Path(), prefix) {
					logger.Log.Debug().Str("path", c.Path()).Msg("Serving static asset")
					break
				}
			}
			return c.Next()
		})
	}

	return app, nil
}
