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

	if len(os.Args) != 1 {
		cm.SetUseTUI(false)
		logger.SetTUI(false)
	}

	for _, arg := range os.Args {
		switch arg {
		case "--no-banner", "-B":
			cm.SetUseBanner(false)
		}
	}

	if cm.GetUseTUI() {
		if err := tui.StartTUI(logger.GetBanner("client")); err != nil {
			fmt.Printf("Error starting TUI: %v\nFalling back to CLI mode\n", err)
			if cm.GetUseBanner() {
				if !logger.IsCompletionCommand() {
					fmt.Println(logger.GetBanner("client"))
				}
			}
			cmd.Execute()
		}
	} else {
		if cm.GetUseBanner() {
			if !logger.IsCompletionCommand() {
				fmt.Println(logger.GetBanner("client"))
			}
		}
		cmd.Execute()
	}
}
