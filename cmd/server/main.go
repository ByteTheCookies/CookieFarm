// Package main is the entry point for the API server.
package main

import (
	_ "embed"
	"fmt"
	"os"
	"strings"

	"github.com/ByteTheCookies/CookieFarm/cmd/server/cmd"
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
	if !isCompletionCommand() {
		fmt.Println(logger.GetBanner("server"))
	}
	cmd.Execute()
}
