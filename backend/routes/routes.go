package routes

import (
	"context"
	"screener/backend/model"
	"screener/backend/service"
	"screener/backend/supabase"
	"strconv"
	"time"

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

		// Admin ingestion endpoint (public): trigger screener+historicals fetch for all symbols
		public.Post("/admin/ingest/historicals", func(c *fiber.Ctx) error {
			concurrency, _ := strconv.Atoi(c.Query("concurrency", "8"))
			fetcher := service.NewFetcherService()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()

			jobID, err := fetcher.RunIngestion(ctx, concurrency)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"success": false,
					"error":   "Internal Server Error",
					"message": err.Error(),
				})
			}

			return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
				"success":     true,
				"job_id":      jobID,
				"accepted_at": time.Now().UTC().Format(time.RFC3339),
			})
		})

		// Inside day screener (public) - DAILY bars only (interval=1d)
		public.Get("/inside-day", func(c *fiber.Ctx) error {
			symbols, err := historicalService.GetSymbolsWithDailyInsideDay()
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

		// Highest volume in last 90 days (public) - DAILY bars only
		public.Get("/high-volume-quarter", func(c *fiber.Ctx) error {
			symbols, err := historicalService.GetSymbolsWithHighestVolumeInQuarter()
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

		// Highest volume in last 365 days (public) - DAILY bars only
		public.Get("/high-volume-year", func(c *fiber.Ctx) error {
			symbols, err := historicalService.GetSymbolsWithHighestVolumeInYear()
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

		// Highest volume ever (public) - DAILY bars only
		public.Get("/high-volume-ever", func(c *fiber.Ctx) error {
			symbols, err := historicalService.GetSymbolsWithHighestVolumeEver()
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

		// ADR screening (public) - filter stocks by ADR% with configurable lookback
		public.Get("/adr-screen", func(c *fiber.Ctx) error {
			rangeParam := c.Query("range")
			interval := c.Query("interval")
			lookbackStr := c.Query("lookback", "14") // default 14 days

			lookback, err := strconv.Atoi(lookbackStr)
			if err != nil || lookback <= 0 {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"success": false,
					"error":   "Bad Request",
					"message": "lookback must be a positive integer",
				})
			}

			var minADR, maxADR *float64
			if minStr := c.Query("min_adr"); minStr != "" {
				if val, err := strconv.ParseFloat(minStr, 64); err == nil {
					minADR = &val
				}
			}
			if maxStr := c.Query("max_adr"); maxStr != "" {
				if val, err := strconv.ParseFloat(maxStr, 64); err == nil {
					maxADR = &val
				}
			}

			symbols, err := historicalService.GetSymbolsByADR(rangeParam, interval, lookback, minADR, maxADR)
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
					"params": fiber.Map{
						"range":    rangeParam,
						"interval": interval,
						"lookback": lookback,
						"min_adr":  minADR,
						"max_adr":  maxADR,
					},
				},
			})
		})

		// ATR screening (public) - filter stocks by ATR% with configurable lookback
		public.Get("/atr-screen", func(c *fiber.Ctx) error {
			rangeParam := c.Query("range")
			interval := c.Query("interval")
			lookbackStr := c.Query("lookback", "14") // default 14 days

			lookback, err := strconv.Atoi(lookbackStr)
			if err != nil || lookback <= 0 {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"success": false,
					"error":   "Bad Request",
					"message": "lookback must be a positive integer",
				})
			}

			var minATR, maxATR *float64
			if minStr := c.Query("min_atr"); minStr != "" {
				if val, err := strconv.ParseFloat(minStr, 64); err == nil {
					minATR = &val
				}
			}
			if maxStr := c.Query("max_atr"); maxStr != "" {
				if val, err := strconv.ParseFloat(maxStr, 64); err == nil {
					maxATR = &val
				}
			}

			symbols, err := historicalService.GetSymbolsByATR(rangeParam, interval, lookback, minATR, maxATR)
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
					"params": fiber.Map{
						"range":    rangeParam,
						"interval": interval,
						"lookback": lookback,
						"min_atr":  minATR,
						"max_atr":  maxATR,
					},
				},
			})
		})

		// Get ADR% for a specific stock (public)
		public.Get("/adr", func(c *fiber.Ctx) error {
			symbol := c.Query("symbol")
			rangeParam := c.Query("range")
			interval := c.Query("interval")
			lookbackStr := c.Query("lookback", "14")

			if symbol == "" || rangeParam == "" || interval == "" {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"success": false,
					"error":   "Bad Request",
					"message": "symbol, range, and interval are required",
				})
			}

			lookback, err := strconv.Atoi(lookbackStr)
			if err != nil || lookback <= 0 {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"success": false,
					"error":   "Bad Request",
					"message": "lookback must be a positive integer",
				})
			}

			adrPercent, err := historicalService.GetADRForSymbol(symbol, rangeParam, interval, lookback)
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
					"symbol":      symbol,
					"adr_percent": adrPercent,
					"params": fiber.Map{
						"range":    rangeParam,
						"interval": interval,
						"lookback": lookback,
					},
				},
			})
		})

		// Get ATR% for a specific stock (public)
		public.Get("/atr", func(c *fiber.Ctx) error {
			symbol := c.Query("symbol")
			rangeParam := c.Query("range")
			interval := c.Query("interval")
			lookbackStr := c.Query("lookback", "14")

			if symbol == "" || rangeParam == "" || interval == "" {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"success": false,
					"error":   "Bad Request",
					"message": "symbol, range, and interval are required",
				})
			}

			lookback, err := strconv.Atoi(lookbackStr)
			if err != nil || lookback <= 0 {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"success": false,
					"error":   "Bad Request",
					"message": "lookback must be a positive integer",
				})
			}

			atrPercent, err := historicalService.GetATRForSymbol(symbol, rangeParam, interval, lookback)
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
					"symbol":      symbol,
					"atr_percent": atrPercent,
					"params": fiber.Map{
						"range":    rangeParam,
						"interval": interval,
						"lookback": lookback,
					},
				},
			})
		})

		// Average volume in dollars screening (public)
		public.Get("/avg-volume-dollars-screen", func(c *fiber.Ctx) error {
			rangeParam := c.Query("range")
			interval := c.Query("interval")
			lookbackStr := c.Query("lookback", "50")

			lookback, err := strconv.Atoi(lookbackStr)
			if err != nil || lookback <= 0 {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"success": false,
					"error":   "Bad Request",
					"message": "lookback must be a positive integer",
				})
			}

			var minVolDollarsM, maxVolDollarsM *float64
			if minStr := c.Query("min_vol_dollars_m"); minStr != "" {
				if val, err := strconv.ParseFloat(minStr, 64); err == nil {
					minVolDollarsM = &val
				}
			}
			if maxStr := c.Query("max_vol_dollars_m"); maxStr != "" {
				if val, err := strconv.ParseFloat(maxStr, 64); err == nil {
					maxVolDollarsM = &val
				}
			}

			symbols, err := historicalService.GetSymbolsByAvgVolumeDollars(rangeParam, interval, lookback, minVolDollarsM, maxVolDollarsM)
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
					"params": fiber.Map{
						"range":             rangeParam,
						"interval":          interval,
						"lookback":          lookback,
						"min_vol_dollars_m": minVolDollarsM,
						"max_vol_dollars_m": maxVolDollarsM,
					},
				},
			})
		})

		// Average volume in percent screening (public)
		public.Get("/avg-volume-percent-screen", func(c *fiber.Ctx) error {
			rangeParam := c.Query("range")
			interval := c.Query("interval")
			lookbackStr := c.Query("lookback", "50")

			lookback, err := strconv.Atoi(lookbackStr)
			if err != nil || lookback <= 0 {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"success": false,
					"error":   "Bad Request",
					"message": "lookback must be a positive integer",
				})
			}

			var minVolPercent, maxVolPercent *float64
			if minStr := c.Query("min_vol_percent"); minStr != "" {
				if val, err := strconv.ParseFloat(minStr, 64); err == nil {
					minVolPercent = &val
				}
			}
			if maxStr := c.Query("max_vol_percent"); maxStr != "" {
				if val, err := strconv.ParseFloat(maxStr, 64); err == nil {
					maxVolPercent = &val
				}
			}

			symbols, err := historicalService.GetSymbolsByAvgVolumePercent(rangeParam, interval, lookback, minVolPercent, maxVolPercent)
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
					"params": fiber.Map{
						"range":           rangeParam,
						"interval":        interval,
						"lookback":        lookback,
						"min_vol_percent": minVolPercent,
						"max_vol_percent": maxVolPercent,
					},
				},
			})
		})

		// Get average volume in dollars for a specific stock (public)
		public.Get("/avg-volume-dollars", func(c *fiber.Ctx) error {
			symbol := c.Query("symbol")
			rangeParam := c.Query("range")
			interval := c.Query("interval")
			lookbackStr := c.Query("lookback", "50")

			if symbol == "" || rangeParam == "" || interval == "" {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"success": false,
					"error":   "Bad Request",
					"message": "symbol, range, and interval are required",
				})
			}

			lookback, err := strconv.Atoi(lookbackStr)
			if err != nil || lookback <= 0 {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"success": false,
					"error":   "Bad Request",
					"message": "lookback must be a positive integer",
				})
			}

			avgVolDollarsM, err := historicalService.GetAvgVolumeDollarsForSymbol(symbol, rangeParam, interval, lookback)
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
					"symbol":            symbol,
					"avg_vol_dollars_m": avgVolDollarsM,
					"params": fiber.Map{
						"range":    rangeParam,
						"interval": interval,
						"lookback": lookback,
					},
				},
			})
		})

		// Get average volume in percent for a specific stock (public)
		public.Get("/avg-volume-percent", func(c *fiber.Ctx) error {
			symbol := c.Query("symbol")
			rangeParam := c.Query("range")
			interval := c.Query("interval")
			lookbackStr := c.Query("lookback", "50")

			if symbol == "" || rangeParam == "" || interval == "" {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"success": false,
					"error":   "Bad Request",
					"message": "symbol, range, and interval are required",
				})
			}

			lookback, err := strconv.Atoi(lookbackStr)
			if err != nil || lookback <= 0 {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"success": false,
					"error":   "Bad Request",
					"message": "lookback must be a positive integer",
				})
			}

			volPercent, err := historicalService.GetAvgVolumePercentForSymbol(symbol, rangeParam, interval, lookback)
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
					"symbol":      symbol,
					"vol_percent": volPercent,
					"params": fiber.Map{
						"range":    rangeParam,
						"interval": interval,
						"lookback": lookback,
					},
				},
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
