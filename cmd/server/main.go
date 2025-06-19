// Package main is the entry point for the API server.
package main

import (
	"context"
	_ "embed"
	"os/signal"
	"syscall"
	"time"

	"github.com/ByteTheCookies/CookieFarm/internal/server/config"
	"github.com/ByteTheCookies/CookieFarm/internal/server/core"
	"github.com/ByteTheCookies/CookieFarm/internal/server/server"
	"github.com/ByteTheCookies/CookieFarm/internal/server/sqlite"
	"github.com/ByteTheCookies/CookieFarm/pkg/logger"
	flogger "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/spf13/pflag"
)

func init() {
	logger.PrintBanner("server")

	config.Debug = pflag.BoolP("debug", "d", false, "Enable debug-level logging")
	config.ConfigPath = pflag.StringP("config", "c", "", "Path to the configuration file")
	config.Password = pflag.StringP("password", "p", "password", "Password for authentication")
	config.ServerPort = pflag.StringP("port", "P", "8080", "Port for server")
}

// The main function initializes configuration, sets up logging, connects to the database,
// configures the Fiber HTTP server, and handles graceful shutdown on system signals.
func main() {
	pflag.Parse()

	level := "info"
	if *config.Debug {
		level = "debug"
	}
	logger.Setup(level, false)
	defer logger.Close()

	if *config.ConfigPath != "" {
		logger.Log.Info().Msg("Using file config...")
		err := core.LoadConfig(*config.ConfigPath)
		if err != nil {
			logger.Log.Warn().Err(err).Msg("Config file not found or corrupted using web config")
		}
	} else {
		logger.Log.Info().Msg("Using web config...")
	}

	var err error
	config.Secret, err = server.InitSecret()
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("Failed to initialize secret key")
	}
	logger.Log.Debug().Str("plain", *config.Password).Msg("Plain password before hashing")
	logger.Log.Debug().Str("Secret", string(config.Secret)).Msg("Secret key for JWT")

	hashed, err := server.HashPassword(*config.Password)
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("Password hashing failed")
	}
	config.Password = &hashed
	logger.Log.Debug().Str("hashed", *config.Password).Msg("Password after hashing")

	sqlite.DB = sqlite.New()
	if err := sqlite.DB.Ping(); err != nil {
		logger.Log.Fatal().Err(err).Msg("Failed to connect to the database")
	}
	logger.Log.Info().Msg("Database initialized")
	defer sqlite.Close()

	app, err := server.NewApp()
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("Failed to initialize server")
	}

	app.Use(flogger.New(flogger.Config{
		Format:     "[${time}] ${ip} - ${method} ${path} - ${status}\n",
		TimeFormat: time.RFC3339,
		TimeZone:   "Local",
	}))

	server.RegisterRoutes(app)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer stop()

	addr := ":" + *config.ServerPort

	errCh := make(chan error, 1)

	go func() {
		logger.Log.Info().Str("addr", addr).Msg("HTTP server starting")
		err := app.Listen(addr)
		if err != nil {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		logger.Log.Warn().Msg("Shutdown signal received, terminating...")

	case err := <-errCh:
		if err != nil {
			logger.Log.Fatal().Err(err).Msg("Server failed to start")
		}
	}

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		logger.Log.Error().Err(err).Msg("Error during shutdown, forcing exit")
	}

	logger.Log.Info().Msg("Server stopped gracefully")
}
