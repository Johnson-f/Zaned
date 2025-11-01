package routes

import (
	"screener/backend/model"
	"screener/backend/service"
	"screener/backend/supabase"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// SetupRoutes configures all routes for the application
func SetupRoutes(app *fiber.App) {
	// Initialize services
	screenerService := service.NewScreenerService()
	historicalService := service.NewHistoricalService()

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

		// Historical data routes
		// Get all historical records
		protected.Get("/historical", func(c *fiber.Ctx) error {
			historical, err := historicalService.GetAllHistorical()
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"success": false,
					"error":   "Internal Server Error",
					"message": err.Error(),
				})
			}

			return c.JSON(fiber.Map{
				"success": true,
				"data":    historical,
			})
		})

		// Get historical records with filters, sorting, and pagination
		protected.Get("/historical/filter", func(c *fiber.Ctx) error {
			var filters *service.HistoricalFilterOptions
			if c.Query("symbol") != "" || c.Query("min_epoch") != "" || c.Query("max_epoch") != "" ||
				c.Query("range") != "" || c.Query("interval") != "" ||
				c.Query("min_open") != "" || c.Query("max_open") != "" ||
				c.Query("min_high") != "" || c.Query("max_high") != "" ||
				c.Query("min_low") != "" || c.Query("max_low") != "" ||
				c.Query("min_close") != "" || c.Query("max_close") != "" ||
				c.Query("min_volume") != "" || c.Query("max_volume") != "" {
				filters = &service.HistoricalFilterOptions{}
				if val := c.Query("symbol"); val != "" {
					filters.Symbol = &val
				}
				if val := c.Query("min_epoch"); val != "" {
					if epoch, err := strconv.ParseInt(val, 10, 64); err == nil {
						filters.MinEpoch = &epoch
					}
				}
				if val := c.Query("max_epoch"); val != "" {
					if epoch, err := strconv.ParseInt(val, 10, 64); err == nil {
						filters.MaxEpoch = &epoch
					}
				}
				if val := c.Query("range"); val != "" {
					filters.Range = &val
				}
				if val := c.Query("interval"); val != "" {
					filters.Interval = &val
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
			}

			// Parse sort options
			var sort *service.HistoricalSortOptions
			if c.Query("sort_field") != "" {
				sort = &service.HistoricalSortOptions{
					Field:     c.Query("sort_field"),
					Direction: c.Query("sort_direction", "asc"),
				}
			}

			// Parse pagination options
			var pagination *service.HistoricalPaginationOptions
			page, _ := strconv.Atoi(c.Query("page", "1"))
			limit, _ := strconv.Atoi(c.Query("limit", "10"))
			if page > 0 || limit > 0 {
				pagination = &service.HistoricalPaginationOptions{
					Page:  page,
					Limit: limit,
				}
			}

			result, err := historicalService.GetHistoricalWithFilters(filters, sort, pagination)
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

		// Get historical count
		protected.Get("/historical/count", func(c *fiber.Ctx) error {
			count, err := historicalService.GetCount()
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

		// Get historical count by symbol
		protected.Get("/historical/count/:symbol", func(c *fiber.Ctx) error {
			symbol := c.Params("symbol")
			count, err := historicalService.GetCountBySymbol(symbol)
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
					"symbol": symbol,
					"count":  count,
				},
			})
		})

		// Get historical by symbol and params (range & interval)
		protected.Get("/historical/symbol/:symbol/params", func(c *fiber.Ctx) error {
			symbol := c.Params("symbol")
			rangeParam := c.Query("range")
			interval := c.Query("interval")

			if rangeParam == "" || interval == "" {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"success": false,
					"error":   "Bad Request",
					"message": "range and interval query parameters are required",
				})
			}

			historical, err := historicalService.GetHistoricalBySymbolAndParams(symbol, rangeParam, interval)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"success": false,
					"error":   "Internal Server Error",
					"message": err.Error(),
				})
			}

			return c.JSON(fiber.Map{
				"success": true,
				"data":    historical,
			})
		})

		// Get historical by symbol
		protected.Get("/historical/symbol/:symbol", func(c *fiber.Ctx) error {
			symbol := c.Params("symbol")
			historical, err := historicalService.GetHistoricalBySymbol(symbol)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"success": false,
					"error":   "Internal Server Error",
					"message": err.Error(),
				})
			}

			return c.JSON(fiber.Map{
				"success": true,
				"data":    historical,
			})
		})

		// Get stocks volume metrics
		protected.Get("/historical/volume-metrics", func(c *fiber.Ctx) error {
			metrics, err := historicalService.GetStocksVolumeMetrics()
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"success": false,
					"error":   "Internal Server Error",
					"message": err.Error(),
				})
			}

			return c.JSON(fiber.Map{
				"success": true,
				"data":    metrics,
			})
		})

		// Create historical record
		protected.Post("/historical", func(c *fiber.Ctx) error {
			var historical model.Historical
			if err := c.BodyParser(&historical); err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"success": false,
					"error":   "Bad Request",
					"message": "Invalid request body",
				})
			}

			if err := historicalService.CreateHistorical(&historical); err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"success": false,
					"error":   "Internal Server Error",
					"message": err.Error(),
				})
			}

			return c.Status(fiber.StatusCreated).JSON(fiber.Map{
				"success": true,
				"data":    historical,
			})
		})

		// Create historical records in batch
		protected.Post("/historical/batch", func(c *fiber.Ctx) error {
			var historical []model.Historical
			if err := c.BodyParser(&historical); err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"success": false,
					"error":   "Bad Request",
					"message": "Invalid request body",
				})
			}

			if err := historicalService.CreateHistoricalBatch(historical); err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"success": false,
					"error":   "Internal Server Error",
					"message": err.Error(),
				})
			}

			return c.Status(fiber.StatusCreated).JSON(fiber.Map{
				"success": true,
				"message": "Historical records created successfully",
				"count":   len(historical),
			})
		})

		// Upsert historical record
		protected.Put("/historical", func(c *fiber.Ctx) error {
			var historical model.Historical
			if err := c.BodyParser(&historical); err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"success": false,
					"error":   "Bad Request",
					"message": "Invalid request body",
				})
			}

			if err := historicalService.UpsertHistorical(&historical); err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"success": false,
					"error":   "Internal Server Error",
					"message": err.Error(),
				})
			}

			return c.JSON(fiber.Map{
				"success": true,
				"data":    historical,
			})
		})

		// Upsert historical records in batch
		protected.Put("/historical/batch", func(c *fiber.Ctx) error {
			var historical []model.Historical
			if err := c.BodyParser(&historical); err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"success": false,
					"error":   "Bad Request",
					"message": "Invalid request body",
				})
			}

			if err := historicalService.UpsertHistoricalBatch(historical); err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"success": false,
					"error":   "Internal Server Error",
					"message": err.Error(),
				})
			}

			return c.JSON(fiber.Map{
				"success": true,
				"message": "Historical records upserted successfully",
				"count":   len(historical),
			})
		})

		// Update historical record by ID
		protected.Put("/historical/:id", func(c *fiber.Ctx) error {
			id := c.Params("id")
			var historical model.Historical
			if err := c.BodyParser(&historical); err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"success": false,
					"error":   "Bad Request",
					"message": "Invalid request body",
				})
			}

			if err := historicalService.UpdateHistorical(id, &historical); err != nil {
				if err.Error() == "record not found" {
					return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
						"success": false,
						"error":   "Not Found",
						"message": "Historical record not found",
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
				"message": "Historical record updated successfully",
			})
		})

		// Delete historical record by ID
		protected.Delete("/historical/:id", func(c *fiber.Ctx) error {
			id := c.Params("id")
			if err := historicalService.DeleteHistorical(id); err != nil {
				if err.Error() == "record not found" {
					return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
						"success": false,
						"error":   "Not Found",
						"message": "Historical record not found",
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
				"message": "Historical record deleted successfully",
			})
		})

		// Delete historical records by symbol and params
		protected.Delete("/historical/symbol/:symbol/params", func(c *fiber.Ctx) error {
			symbol := c.Params("symbol")
			rangeParam := c.Query("range")
			interval := c.Query("interval")

			if rangeParam == "" || interval == "" {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"success": false,
					"error":   "Bad Request",
					"message": "range and interval query parameters are required",
				})
			}

			if err := historicalService.DeleteHistoricalBySymbolAndParams(symbol, rangeParam, interval); err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"success": false,
					"error":   "Internal Server Error",
					"message": err.Error(),
				})
			}

			return c.JSON(fiber.Map{
				"success": true,
				"message": "Historical records deleted successfully",
			})
		})

		// Delete historical records by symbol
		protected.Delete("/historical/symbol/:symbol", func(c *fiber.Ctx) error {
			symbol := c.Params("symbol")
			if err := historicalService.DeleteHistoricalBySymbol(symbol); err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"success": false,
					"error":   "Internal Server Error",
					"message": err.Error(),
				})
			}

			return c.JSON(fiber.Map{
				"success": true,
				"message": "Historical records deleted successfully",
			})
		})

		// Get historical record by ID (must be last to avoid matching specific routes)
		protected.Get("/historical/:id", func(c *fiber.Ctx) error {
			id := c.Params("id")
			historical, err := historicalService.GetHistoricalByID(id)
			if err != nil {
				if err.Error() == "record not found" {
					return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
						"success": false,
						"error":   "Not Found",
						"message": "Historical record not found",
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
				"data":    historical,
			})
		})
	}
}
