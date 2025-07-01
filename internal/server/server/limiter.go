package server

import (
	"time"

	"github.com/ByteTheCookies/CookieFarm/internal/server/config"
	"github.com/ByteTheCookies/CookieFarm/pkg/logger"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

var (
	MaxRequests = config.GetEnvAsInt("RATE_LIMIT_MAX", 50)
	Window      = config.GetEnvAsInt("RATE_LIMIT_WINDOW", 1)
	whitelist   = map[string]bool{
		"127.0.0.1": true,
		"::1":       true,
	}
)

// NewLimiter returns a rate limiter middleware for Fiber.
// When in debug mode, rate limiting is disabled to ease development.
// In production, it limits to 5 requests per minute per IP to prevent abuse (e.g., brute-force on login).
func NewLimiter() fiber.Handler {
	if config.Debug {
		return func(c *fiber.Ctx) error {
			return c.Next()
		}
	}
	return limiter.New(limiter.Config{
		Max:        MaxRequests,
		Expiration: time.Duration(Window) * time.Minute,
		LimitReached: func(c *fiber.Ctx) error {
			if whitelist[c.IP()] {
				return c.Next()
			}

			logger.Log.Warn().
				Str("ip", c.IP()).
				Str("path", c.Path()).
				Msg("Rate limit exceeded on login")
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "Too many login attempts. Please try again later.",
			})
		},
	})
}
