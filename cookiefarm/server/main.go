// Package main is the entry point for the API server.
package main

import (
	_ "embed" // Embed the banner text file for display in the TUI.
	"logger"

	"server/cmd"

	_ "go.uber.org/automaxprocs" // Automatically set GOMAXPROCS to match Linux container CPU quota.
)

func main() {
	if !logger.IsCompletionCommand() {
		logger.PrintBanner(true, "server")
	}
	cmd.Execute()
}
