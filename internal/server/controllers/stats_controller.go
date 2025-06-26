package controllers

import (
	"net/http"

	"github.com/ByteTheCookies/CookieFarm/internal/server/sqlite"
	"github.com/gofiber/fiber/v2"
)

// StatsController handles the statistics of the flag collector
type StatsController struct{}

func NewStatsController() *StatsController {
	return &StatsController{}
}

// GetFlagStats return the statistics of the flag collector
func (*StatsController) GetFlagStats(ctx *fiber.Ctx) error {
	collector := sqlite.GetCollector()
	stats := collector.GetStats()
	bufferSize := collector.GetBufferSize()

	return ctx.JSON(fiber.Map{
		"buffer_size":           bufferSize,
		"total_flags_received":  stats.TotalFlagsReceived,
		"total_flags_flushed":   stats.TotalFlagsFlushed,
		"total_flushes":         stats.TotalFlushes,
		"successful_flushes":    stats.SuccessfulFlushes,
		"failed_flushes":        stats.FailedFlushes,
		"last_flush_time":       stats.LastFlushTime,
		"last_successful_flush": stats.LastSuccessfulFlush,
		"efficiency_ratio":      float64(stats.TotalFlagsFlushed) / float64(stats.TotalFlushes+1), // +1 per evitare divisione per zero
		"status": map[string]any{
			"is_running": collector.IsRunning(),
		},
	})
}

// ForceFlushFlags force a flush of the flags in the collector
func (*StatsController) ForceFlushFlags(ctx *fiber.Ctx) error {
	collector := sqlite.GetCollector()

	statsBefore := collector.GetStats()
	bufferBefore := collector.GetBufferSize()

	err := collector.Flush()

	statsAfter := collector.GetStats()
	bufferAfter := collector.GetBufferSize()

	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"success":         false,
			"error":           err.Error(),
			"buffer_before":   bufferBefore,
			"buffer_after":    bufferAfter,
			"flags_processed": statsBefore.TotalFlagsFlushed,
			"flags_flushed":   statsAfter.TotalFlagsFlushed - statsBefore.TotalFlagsFlushed,
		})
	}

	return ctx.JSON(fiber.Map{
		"success":       true,
		"buffer_before": bufferBefore,
		"buffer_after":  bufferAfter,
		"flags_flushed": statsAfter.TotalFlagsFlushed - statsBefore.TotalFlagsFlushed,
	})
}
