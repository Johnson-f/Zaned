package filtering

import (
	"screener/backend/service/caching"
	filteringservice "screener/backend/service/filtering"
	"time"

	"github.com/gofiber/fiber/v2"
)

// SetupInsideDayRoutes registers all inside-day related routes
func SetupInsideDayRoutes(router fiber.Router) {
	// Admin endpoint to save inside day results (call via cron daily)
	router.Post("/admin/screener/save-inside-day", func(c *fiber.Ctx) error {
		insideDayService := filteringservice.NewInsideDayService()
		if err := insideDayService.SaveInsideDayResults(); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"error":   "Internal Server Error",
				"message": err.Error(),
			})
		}
		// Invalidate screener results cache
		invalidator := caching.NewInvalidationService()
		_ = invalidator.InvalidateScreenerResults("inside_day")
		return c.JSON(fiber.Map{
			"success":     true,
			"message":     "Inside day results saved successfully",
			"accepted_at": time.Now().UTC().Format(time.RFC3339),
		})
	})

	// Public endpoint to get current inside day symbols (real-time calculation)
	router.Get("/inside-day", func(c *fiber.Ctx) error {
		insideDayService := filteringservice.NewInsideDayService()
		symbols, err := insideDayService.GetSymbolsWithDailyInsideDay()
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
