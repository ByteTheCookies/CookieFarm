package cmd

import (
	"context"
	"logger"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"server/ckp"
	"server/config"
	"server/core"
	"server/database"

	_ "github.com/mattn/go-sqlite3" // SQLite driver for database interactions

	"server/api"

	"github.com/charmbracelet/fang"
	"github.com/gofiber/fiber/v3"
	fiberLogger "github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/spf13/cobra"
)

var (
	enablePprof bool // Enable pprof for profiling
	Version     string
)

const CKP_PORT = 7777

// RootCmd represents the base command when called without any subcommands
// Exported for TUI usage
var RootCmd = &cobra.Command{
	Use:     "cks",
	Short:   "Server component of the CookieFarm A/D exploitation framework",
	Long:    `CookieFarm is an automated attack/defense (A/D) exploitation framework developed by the ByteTheCookies team for the CyberChallenge competition. This is the server-side component responsible for coordinating exploit deployment, managing targets, and interfacing with CLI clients.`, //nolint:revive
	Run:     Run,
	Version: Version,
}

func Execute() {
	theme := logger.CookieCLIColorSchema
	if err := fang.Execute(context.TODO(), RootCmd, fang.WithVersion(Version), fang.WithTheme(theme)); err != nil {
		os.Exit(1)
	}
}

func RunPprof() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", pprof.Index)
	mux.HandleFunc("/cmdline", pprof.Cmdline)
	mux.HandleFunc("/profile", pprof.Profile)
	mux.HandleFunc("/symbol", pprof.Symbol)
	mux.HandleFunc("/trace", pprof.Trace)

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

	RootCmd.PersistentFlags().StringVarP(&config.ConfigFile, "config", "c", "", "Use configuration file instead of web config")
	RootCmd.PersistentFlags().StringVarP(&config.Password, "password", "P", "password", "Password for authentication")
	RootCmd.PersistentFlags().StringVarP(&config.ServerPort, "port", "p", "8080", "Port for server")

	RootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		if enablePprof {
			RunPprof()
		}
	}
}

func setupEnv(cfg *config.ConfigManager) (*database.Store, *core.Runner) {
	cfgDB := database.Config{
		DSN:             "file:cookiefarm.db?cache=shared&_journal=WAL&mode=rwc&_busy_timeout=5000",
		MaxOpenConns:    1, // SQLite does not support concurrent writes, so we limit to 1 connection for safety
		MaxIdleConns:    1,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 1 * time.Minute,
	}
	db, err := database.NewDB(cfgDB)
	if err != nil {
		panic("Failed to connect to database: " + err.Error())
	}
	store := database.NewStore(db)
	database.GetCollector().SetStore(store)
	runner := core.NewRunner(store, cfg)

	setupPassword()

	return store, runner
}

func setupPassword() {
	var err error
	config.Secret, err = api.InitSecret()
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("Failed to initialize secret key")
	}

	logger.Log.Debug().Str("plain", config.Password).Msg("Plain password before hashing")
	logger.Log.Debug().Str("Secret", string(config.Secret)).Msg("Secret key for JWT")

	config.Password, err = api.HashPassword(config.Password)
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("Password hashing failed")
	}
	logger.Log.Debug().Str("hashed", config.Password).Msg("Password after hashing")
}

func loadConfig(runner *core.Runner) {
	if config.ConfigFile != "" {
		logger.Log.Info().Msg("Using file config...")
		err := runner.LoadConfig(config.ConfigFile)
		if err != nil {
			logger.Log.Warn().Err(err).Msg("Config file not found or corrupted using web config")
		}
		runner.Submission()
	} else {
		logger.Log.Info().Msg("Using web config...")
	}
}

func listen(app *fiber.App, addr string, errCh chan<- error) {
	logger.Log.Info().Str("addr", addr).Msg("HTTP server starting")
	err := app.Listen(addr, fiber.ListenConfig{
		DisableStartupMessage: !config.Debug,
		EnablePrintRoutes:     config.Debug,
		EnablePrefork:         false,
	})
	if err != nil {
		errCh <- err
	}
}

func handleShutdown(app *fiber.App, ctx context.Context, errCh <-chan error) {
	select {
	case <-ctx.Done():
		logger.Log.Warn().Msg("Shutdown signal received, terminating...")
	case err := <-errCh:
		if err != nil {
			logger.Log.Fatal().Err(err).Msg("Server failed to start")
		}
	}

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		logger.Log.Error().Err(err).Msg("Error during shutdown, forcing exit")
	}

	logger.Log.Info().Msg("Server stopped gracefully")
}

func setupApp(store *database.Store, runner *core.Runner, cfg *config.ConfigManager, connections *ckp.Connections) *fiber.App {
	app, err := api.NewApp()
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("Failed to initialize server")
	}

	app.Use(fiberLogger.New(fiberLogger.Config{
		Format:     "[${time}] ${ip} - ${method} ${path} - ${status}\n",
		TimeFormat: time.RFC3339,
		TimeZone:   "Local",
	}))
	handler := api.NewHandler(store, runner, cfg, connections)
	handler.RegisterRoutes(app)

	return app
}

// The main function initializes configuration, sets up logging, connects to the database,
// configures the Fiber HTTP server, and handles graceful shutdown on system signals.
func Run(cmd *cobra.Command, args []string) {
	var level string

	if config.Debug {
		level = "debug"
	} else {
		level = "info"
	}

	logger.Setup(level, false)
	defer logger.Close()

	cfg := config.GetInstance()
	store, runner := setupEnv(cfg)

	loadConfig(runner)

	connections, err := ckp.StartServer(CKP_PORT)
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("Failed to start CKP server")
	}

	app := setupApp(store, runner, cfg, connections)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer stop()

	addr := ":" + config.ServerPort
	errCh := make(chan error, 1)
	go listen(app, addr, errCh)

	handleShutdown(app, ctx, errCh)
}
