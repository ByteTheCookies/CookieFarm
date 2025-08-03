package cmd

import (
	"context"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ByteTheCookies/CookieFarm/internal/server/config"
	"github.com/ByteTheCookies/CookieFarm/internal/server/core"
	"github.com/ByteTheCookies/CookieFarm/internal/server/server"
	"github.com/ByteTheCookies/CookieFarm/internal/server/sqlite"
	"github.com/ByteTheCookies/CookieFarm/pkg/logger"
	"github.com/ByteTheCookies/CookieFarm/pkg/models"
	cnats "github.com/ByteTheCookies/CookieFarm/pkg/nats"
	"github.com/charmbracelet/fang"
	fiberLogger "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/spf13/cobra"
)

var enablePprof bool // Enable pprof for profiling

// RootCmd represents the base command when called without any subcommands
// Exported for TUI usage
var RootCmd = &cobra.Command{
	Use:     "cks",
	Short:   "Server component of the CookieFarm A/D exploitation framework",
	Long:    `CookieFarm is an automated attack/defense (A/D) exploitation framework developed by the ByteTheCookies team for the CyberChallenge competition. This is the server-side component responsible for coordinating exploit deployment, managing targets, and interfacing with CLI clients.`, //nolint:revive
	Version: models.VERSION,
	Run:     Run,
}

func Execute() {
	theme := logger.CookieCLIColorSchema
	if err := fang.Execute(context.TODO(), RootCmd, fang.WithVersion(models.VERSION), fang.WithTheme(theme)); err != nil {
		os.Exit(1)
	}
}

func RunPprof() {
	mux := http.NewServeMux()
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	go func() {
		logger.Log.Info().Msg("pprof attivo su :6060")
		server := &http.Server{
			Addr:         "localhost:6060",
			Handler:      mux,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  120 * time.Second,
		}
		logger.Log.Info().Msgf("%s", server.ListenAndServe())
	}()
}

func init() {
	RootCmd.PersistentFlags().BoolVarP(&config.Debug, "debug", "D", false, "Enable debug logging")
	RootCmd.PersistentFlags().BoolVarP(&enablePprof, "pprof", "b", false, "Enable pprof for profiling")

	RootCmd.PersistentFlags().BoolVarP(&config.UseConfigFile, "config", "c", false, "Use configuration file instead of web config")
	RootCmd.PersistentFlags().StringVarP(&config.Password, "password", "P", "password", "Password for authentication")
	RootCmd.PersistentFlags().StringVarP(&config.ServerPort, "port", "p", "8080", "Port for server")

	RootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		if enablePprof {
			RunPprof()
		}
	}
}

// The main function initializes configuration, sets up logging, connects to the database,
// configures the Fiber HTTP server, and handles graceful shutdown on system signals.
func Run(cmd *cobra.Command, args []string) {
	var level string
	var err error

	if config.Debug {
		level = "debug"
	} else {
		level = "info"
	}

	logger.Setup(level, false)
	defer logger.Close()

	if config.UseConfigFile {
		logger.Log.Info().Msg("Using file config...")
		err := core.LoadConfigAndRun(config.ConfigPath)
		if err != nil {
			logger.Log.Warn().Err(err).Msg("Config file not found or corrupted using web config")
		}
	} else {
		logger.Log.Info().Msg("Using web config...")
	}

	config.Secret, err = server.InitSecret()
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("Failed to initialize secret key")
	}

	logger.Log.Debug().Str("plain", config.Password).Msg("Plain password before hashing")
	logger.Log.Debug().Str("Secret", string(config.Secret)).Msg("Secret key for JWT")

	config.Password, err = server.HashPassword(config.Password)
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("Password hashing failed")
	}
	logger.Log.Debug().Str("hashed", config.Password).Msg("Password after hashing")

	sqlite.DBPool = sqlite.New()
	defer sqlite.Close()

	_, err = cnats.SetupNATS()
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("Failed to set up NATS server")
	}
	// defer ns.ShutdownEmbeddedServer()

	app, err := server.NewApp()
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("Failed to initialize server")
	}

	app.Use(fiberLogger.New(fiberLogger.Config{
		Format:     "[${time}] ${ip} - ${method} ${path} - ${status}\n",
		TimeFormat: time.RFC3339,
		TimeZone:   "Local",
	}))
	server.RegisterRoutes(app)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer stop()

	addr := ":" + config.ServerPort
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
