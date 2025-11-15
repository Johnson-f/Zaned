package filtering

import (
	"screener/backend/service/caching"
	filteringservice "screener/backend/service/filtering"
	"time"

	"github.com/gofiber/fiber/v2"
)

// SetupHighVolumeQuarterRoutes registers all high-volume-quarter related routes
func SetupHighVolumeQuarterRoutes(router fiber.Router) {
	// Admin endpoint to save high volume quarter results (call via cron daily)
	router.Post("/admin/screener/save-high-volume-quarter", func(c *fiber.Ctx) error {
		highVolumeQuarterService := filteringservice.NewHighVolumeQuarterService()
		if err := highVolumeQuarterService.SaveHighVolumeQuarterResults(); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"error":   "Internal Server Error",
				"message": err.Error(),
			})
		}
		// Invalidate screener results cache
		invalidator := caching.NewInvalidationService()
		_ = invalidator.InvalidateScreenerResults("high_volume_quarter")
		return c.JSON(fiber.Map{
			"success":     true,
			"message":     "High volume quarter results saved successfully",
			"accepted_at": time.Now().UTC().Format(time.RFC3339),
		})
	})

	// Public endpoint to get current high volume quarter symbols (real-time calculation)
	router.Get("/high-volume-quarter", func(c *fiber.Ctx) error {
		highVolumeQuarterService := filteringservice.NewHighVolumeQuarterService()
		symbols, err := highVolumeQuarterService.GetSymbolsWithHighestVolumeInQuarter()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"error":   "Internal Server Error",
				"message": err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"success": true,
			"data": fiber.Map{
				"symbols": symbols,
				"count":   len(symbols),
			},
		})
	})
}

