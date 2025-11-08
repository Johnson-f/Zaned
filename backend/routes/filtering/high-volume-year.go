package filtering

import (
	filteringservice "screener/backend/service/filtering"
	"time"

	"github.com/gofiber/fiber/v2"
)

// SetupHighVolumeYearRoutes registers all high-volume-year related routes
func SetupHighVolumeYearRoutes(router fiber.Router) {
	// Admin endpoint to save high volume year results (call via cron daily)
	router.Post("/admin/screener/save-high-volume-year", func(c *fiber.Ctx) error {
		highVolumeYearService := filteringservice.NewHighVolumeYearService()
		if err := highVolumeYearService.SaveHighVolumeYearResults(); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"error":   "Internal Server Error",
				"message": err.Error(),
			})
		}
		return c.JSON(fiber.Map{
			"success":     true,
			"message":     "High volume year results saved successfully",
			"accepted_at": time.Now().UTC().Format(time.RFC3339),
		})
	})

	// Public endpoint to get current high volume year symbols (real-time calculation)
	router.Get("/high-volume-year", func(c *fiber.Ctx) error {
		highVolumeYearService := filteringservice.NewHighVolumeYearService()
		symbols, err := highVolumeYearService.GetSymbolsWithHighestVolumeInYear()
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

