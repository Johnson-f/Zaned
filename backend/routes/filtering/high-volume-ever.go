package filtering

import (
	"screener/backend/service/caching"
	filteringservice "screener/backend/service/filtering"
	"time"

	"github.com/gofiber/fiber/v2"
)

// SetupHighVolumeEverRoutes registers all high-volume-ever related routes
func SetupHighVolumeEverRoutes(router fiber.Router) {
	// Admin endpoint to save high volume ever results (call via cron daily)
	router.Post("/admin/screener/save-high-volume-ever", func(c *fiber.Ctx) error {
		highVolumeEverService := filteringservice.NewHighVolumeEverService()
		if err := highVolumeEverService.SaveHighVolumeEverResults(); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"error":   "Internal Server Error",
				"message": err.Error(),
			})
		}
		// Invalidate screener results cache
		invalidator := caching.NewInvalidationService()
		_ = invalidator.InvalidateScreenerResults("high_volume_ever")
		return c.JSON(fiber.Map{
			"success":     true,
			"message":     "High volume ever results saved successfully",
			"accepted_at": time.Now().UTC().Format(time.RFC3339),
		})
	})

	// Public endpoint to get current high volume ever symbols (real-time calculation)
	router.Get("/high-volume-ever", func(c *fiber.Ctx) error {
		highVolumeEverService := filteringservice.NewHighVolumeEverService()
		symbols, err := highVolumeEverService.GetSymbolsWithHighestVolumeEver()
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

