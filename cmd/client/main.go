package main

import (
	_ "embed"
	"fmt"
	"os"
	"strings"

	"github.com/ByteTheCookies/CookieFarm/cmd/client/cmd"
	"github.com/ByteTheCookies/CookieFarm/internal/client/config"
	"github.com/ByteTheCookies/CookieFarm/internal/client/tui"
	"github.com/ByteTheCookies/CookieFarm/pkg/logger"
)

func isCompletionCommand() bool {
	for _, arg := range os.Args {
		if strings.Contains(arg, "__complete") || strings.Contains(arg, "completion") {
			return true
		}
	}
	return false
}

func main() {
	cm := config.GetConfigManager()
	debug := false
	cm.SetUseTUI(true)
	logger.SetTUI(true)

	if len(os.Args) == 1 {
		cm.SetUseTUI(true)
		cm.SetUseBanner(true)
	} else {
		cm.SetUseTUI(false)
		logger.SetTUI(false)
	}

	for _, arg := range os.Args {
		switch arg {
		case "-D", "--debug":
			debug = true
		case "--no-banner", "-B":
			cm.SetUseBanner(false)
		}
	}

	if cm.GetUseTUI() {
		if err := tui.StartTUI(logger.GetBanner("client"), debug); err != nil {
			fmt.Printf("Error starting TUI: %v\nFalling back to CLI mode\n", err)
			if cm.GetUseBanner() {
				if !isCompletionCommand() {
					fmt.Println(logger.GetBanner("client"))
				}
			}
			cmd.Execute()
		}
	} else {
		if cm.GetUseBanner() {
			if !isCompletionCommand() {
				fmt.Println(logger.GetBanner("client"))
			}
		}
		cmd.Execute()
	}
}
