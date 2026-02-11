package main

import (
	"client/cmd"
	"client/config"
	"client/tui"
	_ "embed"
	"fmt"
	"logger"
	"os"
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
