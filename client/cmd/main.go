// Package main is the entry point for the CookieFarm client,
// responsible for initializing configuration, validating input,
// and executing exploits in a loop.
package main

import (
	_ "embed"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/ByteTheCookies/cookieclient/internal/api"
	"github.com/ByteTheCookies/cookieclient/internal/config"
	"github.com/ByteTheCookies/cookieclient/internal/executor"
	"github.com/ByteTheCookies/cookieclient/internal/logger"
	"github.com/ByteTheCookies/cookieclient/internal/submitter"
	"github.com/ByteTheCookies/cookieclient/internal/utils"
	"github.com/rs/zerolog"
	"github.com/spf13/pflag"
)

var logPath string // Path to the generated log file

//go:embed banner.txt
var banner string

// init initializes all command-line flags and binds them to the args struct.
func init() {
	fmt.Println(banner)

	config.Args.ExploitPath = pflag.StringP("exploit", "e", "", "Path to the exploit file to execute")
	config.Args.Debug = pflag.BoolP("debug", "D", false, "Enable debug logging")
	config.Args.Password = pflag.StringP("password", "p", "", "Password for authenticating to the server")
	config.HostServer = pflag.StringP("host", "h", "", "Host of the cookieserver")
	config.Args.Detach = pflag.BoolP("detach", "d", false, "Run the exploit in the background (detached mode)")
	config.Args.TickTime = pflag.IntP("tick", "t", 120, "Interval in seconds between exploit executions")
	config.Args.ThreadCount = pflag.IntP("thread", "T", 5, "Number of concurrent threads to run the exploit with")
}

// SetupClient handles the full initialization process:
// - Parse flags
// - Setup logging
// - Validate arguments
// - Authenticate with the server
// - Sync client configuration
func setupClient() error {
	pflag.Parse()

	if *config.Args.Detach {
		fmt.Println(utils.Blue + "[INFO]" + utils.Reset + " | Detaching from terminal")
		utils.Detach()
	}

	if *config.Args.Debug {
		logPath = logger.Setup("debug")
	} else {
		logPath = logger.Setup("info")
	}

	err := utils.ValidateArgs(config.Args)
	if err != nil {
		return fmt.Errorf("invalid arguments: %w", err)
	}

	logger.Log.Debug().Int("ThreadCount", *config.Args.ThreadCount).Int("Tick time", *config.Args.TickTime)
	logger.Log.Debug().Str("ExploitPath", *config.Args.ExploitPath).Str("HostServer", *config.HostServer).Msg("Arguments validated")

	config.Token, err = api.Login(*config.Args.Password)
	if err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	config.Current, err = api.GetConfig()
	if err != nil {
		return fmt.Errorf("failed to get config: %w", err)
	}

	logger.Log.Debug().Msgf("Current configuration: %+v", config.Current)

	if !config.Current.Configured {
		logger.Log.Fatal().Msg("Server not configured. Please run the configurator before using the client.")
	}

	return nil
}

// Main is the main execution flow of the CookieFarm client.
// It handles setup, starts the exploit, and manages the flag submission process.
func main() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		logger.Log.Info().Msg("Shutting down gracefully...")
		os.Exit(0)
	}()

	if err := setupClient(); err != nil {
		if logger.LogLevel != zerolog.Disabled {
			logger.Log.Fatal().Err(err).Msg("Initialization error")
			logger.Close()
		} else {
			fmt.Println("Error initializing:", err)
		}
		os.Exit(1)
	}
	defer logger.Close()

	// if err := websockets.ConnectToWebSocket(); err != nil {
	// 	logger.Log.Fatal().Err(err).Msg("Failed to connect to WebSocket")
	// }

	logger.Log.Info().Msg("Client initialized successfully")

<<<<<<< HEAD
	result, err := executor.Start(*args.ExploitPath, *args.Password, *args.TickTime, *args.ThreadCount, logPath, *args.Port)
=======
	result, err := executor.Start(*config.Args.ExploitPath, *config.Args.Password, *config.Args.TickTime, *config.Args.ThreadCount, logPath)
>>>>>>> dev
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("Failed to execute exploit")
	}
	logger.Log.Info().Msg("Exploit started successfully")

	go submitter.Start(result.FlagsChan)

	if err := result.Cmd.Wait(); err != nil {
		logger.Log.Error().Err(err).Msg("Exploit process exited with error")
	}
}
