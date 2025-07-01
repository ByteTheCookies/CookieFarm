// Package main is the entry point for the API server.
package main

import (
	_ "embed"
	"fmt"

	"github.com/ByteTheCookies/CookieFarm/cmd/server/cmd"
	"github.com/ByteTheCookies/CookieFarm/pkg/logger"
	_ "go.uber.org/automaxprocs"
)

func main() {
	if !logger.IsCompletionCommand() {
		fmt.Println(logger.GetBanner("server"))
	}
	cmd.Execute()
}
