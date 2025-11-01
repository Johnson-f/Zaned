package routes

import (
	"screener/backend/service"
	"screener/backend/supabase"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// SetupRoutes configures all routes for the application
func SetupRoutes(app *fiber.App) {
	// Initialize services
	screenerService := service.NewScreenerService()

	// Public routes
	public := app.Group("/api")
	{
		// Health check endpoint
		public.Get("/health", func(c *fiber.Ctx) error {
			return c.JSON(fiber.Map{
				"status":  "ok",
				"message": "Server is running",
			})
		})
	}

	// Protected routes (require JWT authentication)
	protected := app.Group("/api/protected")
	// Apply JWT middleware to all protected routes
	protected.Use(supabase.JWTAuthMiddleware())
	{
		// Get all screener data (read-only)
		protected.Get("/screener", func(c *fiber.Ctx) error {
			screeners, err := screenerService.GetAllScreeners()
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"success": false,
					"error":   "Internal Server Error",
					"message": err.Error(),
				})
			}

			return c.JSON(fiber.Map{
				"success": true,
				"data":    screeners,
			})
		})

		// Get screeners with advanced filtering, sorting, and pagination
		protected.Get("/screener/filter", func(c *fiber.Ctx) error {
			// Parse filter options from query parameters
			var filters *service.FilterOptions
			if c.Query("min_price") != "" || c.Query("max_price") != "" ||
				c.Query("min_volume") != "" || c.Query("max_volume") != "" ||
				c.Query("min_open") != "" || c.Query("max_open") != "" ||
				c.Query("min_high") != "" || c.Query("max_high") != "" ||
				c.Query("min_low") != "" || c.Query("max_low") != "" ||
				c.Query("min_close") != "" || c.Query("max_close") != "" {
				filters = &service.FilterOptions{}
				if val := c.Query("min_price"); val != "" {
					if price, err := strconv.ParseFloat(val, 64); err == nil {
						filters.MinPrice = &price
					}
				}
				if val := c.Query("max_price"); val != "" {
					if price, err := strconv.ParseFloat(val, 64); err == nil {
						filters.MaxPrice = &price
					}
				}
				if val := c.Query("min_volume"); val != "" {
					if volume, err := strconv.ParseInt(val, 10, 64); err == nil {
						filters.MinVolume = &volume
					}
				}
				if val := c.Query("max_volume"); val != "" {
					if volume, err := strconv.ParseInt(val, 10, 64); err == nil {
						filters.MaxVolume = &volume
					}
				}
				if val := c.Query("min_open"); val != "" {
					if open, err := strconv.ParseFloat(val, 64); err == nil {
						filters.MinOpen = &open
					}
				}
				if val := c.Query("max_open"); val != "" {
					if open, err := strconv.ParseFloat(val, 64); err == nil {
						filters.MaxOpen = &open
					}
				}
				if val := c.Query("min_high"); val != "" {
					if high, err := strconv.ParseFloat(val, 64); err == nil {
						filters.MinHigh = &high
					}
				}
				if val := c.Query("max_high"); val != "" {
					if high, err := strconv.ParseFloat(val, 64); err == nil {
						filters.MaxHigh = &high
					}
				}
				if val := c.Query("min_low"); val != "" {
					if low, err := strconv.ParseFloat(val, 64); err == nil {
						filters.MinLow = &low
					}
				}
				if val := c.Query("max_low"); val != "" {
					if low, err := strconv.ParseFloat(val, 64); err == nil {
						filters.MaxLow = &low
					}
				}
				if val := c.Query("min_close"); val != "" {
					if close, err := strconv.ParseFloat(val, 64); err == nil {
						filters.MinClose = &close
					}
				}
				if val := c.Query("max_close"); val != "" {
					if close, err := strconv.ParseFloat(val, 64); err == nil {
						filters.MaxClose = &close
					}
				}
			}

			// Parse sort options
			var sort *service.SortOptions
			if c.Query("sort_field") != "" {
				sort = &service.SortOptions{
					Field:     c.Query("sort_field"),
					Direction: c.Query("sort_direction", "asc"),
				}
			}

			// Parse pagination options
			var pagination *service.PaginationOptions
			page, _ := strconv.Atoi(c.Query("page", "1"))
			limit, _ := strconv.Atoi(c.Query("limit", "10"))
			if page > 0 || limit > 0 {
				pagination = &service.PaginationOptions{
					Page:  page,
					Limit: limit,
				}
			}

			result, err := screenerService.GetScreenersWithFilters(filters, sort, pagination)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"success": false,
					"error":   "Internal Server Error",
					"message": err.Error(),
				})
			}

			return c.JSON(fiber.Map{
				"success": true,
				"data":    result,
			})
		})

		// Get top gainers (must come before /:id route)
		protected.Get("/screener/top-gainers", func(c *fiber.Ctx) error {
			limit, _ := strconv.Atoi(c.Query("limit", "10"))
			screeners, err := screenerService.GetTopGainers(limit)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"success": false,
					"error":   "Internal Server Error",
					"message": err.Error(),
				})
			}

			return c.JSON(fiber.Map{
				"success": true,
				"data":    screeners,
			})
		})

		// Get most active stocks (must come before /:id route)
		protected.Get("/screener/most-active", func(c *fiber.Ctx) error {
			limit, _ := strconv.Atoi(c.Query("limit", "10"))
			screeners, err := screenerService.GetMostActive(limit)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"success": false,
					"error":   "Internal Server Error",
					"message": err.Error(),
				})
			}

			return c.JSON(fiber.Map{
				"success": true,
				"data":    screeners,
			})
		})

		// Get total count of screeners (must come before /:id route)
		protected.Get("/screener/count", func(c *fiber.Ctx) error {
			count, err := screenerService.GetCount()
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
					"count": count,
				},
			})
		})

		// Search screeners by symbol (must come before /:id route)
		protected.Get("/screener/search", func(c *fiber.Ctx) error {
			searchTerm := c.Query("q")
			if searchTerm == "" {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"success": false,
					"error":   "Bad Request",
					"message": "Search term (q) is required",
				})
			}

			limit, _ := strconv.Atoi(c.Query("limit", "10"))
			screeners, err := screenerService.SearchScreenersBySymbol(searchTerm, limit)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"success": false,
					"error":   "Internal Server Error",
					"message": err.Error(),
				})
			}

			return c.JSON(fiber.Map{
				"success": true,
				"data":    screeners,
			})
		})

		// Get screeners by price range (must come before /:id route)
		protected.Get("/screener/price-range", func(c *fiber.Ctx) error {
			minPrice, err := strconv.ParseFloat(c.Query("min"), 64)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"success": false,
					"error":   "Bad Request",
					"message": "Invalid min price",
				})
			}

			maxPrice, err := strconv.ParseFloat(c.Query("max"), 64)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"success": false,
					"error":   "Bad Request",
					"message": "Invalid max price",
				})
			}

			screeners, err := screenerService.GetScreenersByPriceRange(minPrice, maxPrice)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"success": false,
					"error":   "Internal Server Error",
					"message": err.Error(),
				})
			}

			return c.JSON(fiber.Map{
				"success": true,
				"data":    screeners,
			})
		})

		// Get screeners by volume range (must come before /:id route)
		protected.Get("/screener/volume-range", func(c *fiber.Ctx) error {
			minVolume, err := strconv.ParseInt(c.Query("min"), 10, 64)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"success": false,
					"error":   "Bad Request",
					"message": "Invalid min volume",
				})
			}

			maxVolume, err := strconv.ParseInt(c.Query("max"), 10, 64)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"success": false,
					"error":   "Bad Request",
					"message": "Invalid max volume",
				})
			}

			screeners, err := screenerService.GetScreenersByVolumeRange(minVolume, maxVolume)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"success": false,
					"error":   "Internal Server Error",
					"message": err.Error(),
				})
			}

			return c.JSON(fiber.Map{
				"success": true,
				"data":    screeners,
			})
		})

		// Get screener by symbol (must come before /:id route)
		protected.Get("/screener/symbol/:symbol", func(c *fiber.Ctx) error {
			symbol := c.Params("symbol")

			screener, err := screenerService.GetScreenerBySymbol(symbol)
			if err != nil {
				if err.Error() == "record not found" {
					return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
						"success": false,
						"error":   "Not Found",
						"message": "Screener record not found for symbol",
					})
				}
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"success": false,
					"error":   "Internal Server Error",
					"message": err.Error(),
				})
			}

			return c.JSON(fiber.Map{
				"success": true,
				"data":    screener,
			})
		})

		// Get screener by ID (must be last to avoid matching specific routes)
		protected.Get("/screener/:id", func(c *fiber.Ctx) error {
			id := c.Params("id")

			screener, err := screenerService.GetScreenerByID(id)
			if err != nil {
				if err.Error() == "record not found" {
					return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
						"success": false,
						"error":   "Not Found",
						"message": "Screener record not found",
					})
				}
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"success": false,
					"error":   "Internal Server Error",
					"message": err.Error(),
				})
			}

			return c.JSON(fiber.Map{
				"success": true,
				"data":    screener,
			})
		})

		// Get screeners by multiple symbols (POST with JSON body)
		protected.Post("/screener/symbols", func(c *fiber.Ctx) error {
			var request struct {
				Symbols []string `json:"symbols"`
			}

			if err := c.BodyParser(&request); err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"success": false,
					"error":   "Bad Request",
					"message": "Invalid request body",
				})
			}

			if len(request.Symbols) == 0 {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"success": false,
					"error":   "Bad Request",
					"message": "Symbols array is required",
				})
			}

			screeners, err := screenerService.GetScreenersBySymbols(request.Symbols)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"success": false,
					"error":   "Internal Server Error",
					"message": err.Error(),
				})
			}

			return c.JSON(fiber.Map{
				"success": true,
				"data":    screeners,
			})
		})
	}
}
