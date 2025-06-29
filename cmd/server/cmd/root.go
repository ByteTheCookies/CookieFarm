package cmd

import (
	"context"
	"image/color"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ByteTheCookies/CookieFarm/internal/server/config"
	"github.com/ByteTheCookies/CookieFarm/internal/server/core"
	"github.com/ByteTheCookies/CookieFarm/internal/server/server"
	"github.com/ByteTheCookies/CookieFarm/internal/server/sqlite"
	"github.com/ByteTheCookies/CookieFarm/pkg/logger"
	"github.com/charmbracelet/fang"
	flogger "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/spf13/cobra"
)

var (
	debug     bool
	useBanner bool
)

const VERSION = "v1.2.0"

// RootCmd represents the base command when called without any subcommands
// Exported for TUI usage
var RootCmd = &cobra.Command{
	Use:   "cks",
	Short: "The server for CookieFarm A/D tool",
	Long: `CookieFarm is a exploiter writed by the team ByteTheCookies for CyberChallenge
competition. This is the server side for the CookieFarm server for attack the teams with exploits.`,
	Version: VERSION,
	Run:     Run,
}

func Execute() {
	theme := fang.ColorScheme{
		Base:           color.RGBA{0xe9, 0xe9, 0xe9, 0xe9},
		Title:          color.RGBA{0xCD, 0xA1, 0x57, 0xff},
		Description:    color.RGBA{0xD9, 0xD9, 0xD9, 0xff},
		Codeblock:      color.RGBA{0x0a, 0x0c, 0x0d, 0xff},
		Program:        color.RGBA{0xCD, 0xA1, 0x57, 0xff},
		DimmedArgument: color.RGBA{0x88, 0x88, 0x88, 0xff},
		Comment:        color.RGBA{0x88, 0x88, 0x88, 0xff},
		Flag:           color.RGBA{0x21, 0x96, 0xF3, 0xff},
		FlagDefault:    color.RGBA{0xD9, 0xD9, 0xD9, 0xff},
		Command:        color.RGBA{0xCD, 0xA1, 0x57, 0xff},
		QuotedString:   color.RGBA{0x21, 0x9B, 0x54, 0xff},
		Argument:       color.RGBA{0xED, 0xED, 0xED, 0xff},
		Help:           color.RGBA{0x88, 0x88, 0x88, 0xff},
		Dash:           color.RGBA{0x88, 0x88, 0x88, 0xff},
		ErrorHeader:    [2]color.Color{color.RGBA{0xED, 0xED, 0xED, 0xff}, color.RGBA{0xE7, 0x4C, 0x3C, 0xff}},
		ErrorDetails:   color.RGBA{0xE7, 0x4C, 0x3C, 0xff},
	}
	if err := fang.Execute(context.TODO(), RootCmd, fang.WithVersion(VERSION), fang.WithTheme(theme)); err != nil {
		os.Exit(1)
	}
}

func init() {
	RootCmd.PersistentFlags().BoolVarP(&debug, "debug", "D", false, "Enable debug logging")
	RootCmd.PersistentFlags().BoolVarP(&config.ConfigFile, "config", "c", false, "Use configuration file instead of web config")
	RootCmd.PersistentFlags().StringVarP(&config.Password, "password", "p", "password", "Password for authentication")
	RootCmd.PersistentFlags().StringVarP(&config.ServerPort, "port", "P", "8080", "Port for server")

	RootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		if debug {
			logger.Setup("debug", true)
		} else {
			logger.Setup("info", true)
		}
	}
}

// The main function initializes configuration, sets up logging, connects to the database,
// configures the Fiber HTTP server, and handles graceful shutdown on system signals.
func Run(cmd *cobra.Command, args []string) {
	level := "info"
	if config.Debug {
		level = "debug"
	}
	logger.Setup(level, false)
	defer logger.Close()

	if config.ConfigFile {
		logger.Log.Info().Msg("Using file config...")
		err := core.LoadConfig(config.ConfigPath)
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
	logger.Log.Debug().Str("plain", config.Password).Msg("Plain password before hashing")
	logger.Log.Debug().Str("Secret", string(config.Secret)).Msg("Secret key for JWT")

	hashed, err := server.HashPassword(config.Password)
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("Password hashing failed")
	}
	config.Password = hashed
	logger.Log.Debug().Str("hashed", config.Password).Msg("Password after hashing")

	sqlite.DBPool = sqlite.New()
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
