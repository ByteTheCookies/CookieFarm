package main

import (
	_ "embed"
	"fmt"
	"os"

	"github.com/ByteTheCookies/CookieFarm/cmd/client/cmd"
	"github.com/ByteTheCookies/CookieFarm/internal/client/config"
	"github.com/ByteTheCookies/CookieFarm/internal/client/tui"
	"github.com/ByteTheCookies/CookieFarm/pkg/logger"
)

func main() {
	cm := config.GetConfigManager()

	// If no arguments are provided, set TUI mode to false
	if len(os.Args) != 1 {
		cm.SetUseTUI(false)
	}

	// Check if print the banner is enabled
	for _, arg := range os.Args {
		if arg == "--no-banner" || arg == "-B" {
			cm.SetUseBanner(false)
		}
	}

	if cm.GetUseTUI() {
		if err := tui.StartTUI(logger.GetBanner("client")); err != nil {
			fmt.Printf("Error starting TUI: %v\nFalling back to CLI mode\n", err)
			if !logger.IsCompletionCommand() {
				logger.PrintBanner(cm.GetUseBanner(), "client")
			}
			cmd.Execute()
		}
	} else {
		if !logger.IsCompletionCommand() {
			logger.PrintBanner(cm.GetUseBanner(), "client")
		}
		cmd.Execute()
	}
}
