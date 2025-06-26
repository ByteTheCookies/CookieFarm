package ui

import (
	"path/filepath"
	"time"

	"github.com/ByteTheCookies/CookieFarm/pkg/logger"
	"github.com/gofiber/template/html/v2"
)

func InitTemplateEngine(debug bool) *html.Engine {
	path, err := filepath.Abs("internal/server/ui/views")
	logger.Log.Debug().Str("path", path)
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("Unable to resolve absolute template path")
	}
	logger.Log.Debug().Str("template_path", path).Msg("Using disk templates")
	engine := html.New(path, ".html")

	engineFuncMap := map[string]any{
		"mul":    func(a, b int) int { return a * b },
		"add":    func(a, b int) int { return a + b },
		"sub":    func(a, b int) int { return a - b },
		"subu64": func(a, b uint64) uint64 { return a - b },
		"format_timestamp": func(timestamp uint64) string {
			return time.Unix(int64(timestamp), 0).Format("15:04:05.12340")
		},
	}
	engine.AddFuncMap(engineFuncMap)

	return engine
}
