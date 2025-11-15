package routes

import (
	"context"
	"fmt"
	"screener/backend/model"
	"screener/backend/routes/filtering"
	"screener/backend/service"
	"screener/backend/service/caching"
	indicatorsscreening "screener/backend/service/filtering/indicators/screening"
	"screener/backend/supabase"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// SetupRoutes configures all routes for the application
func SetupRoutes(app *fiber.App) {
	// Initialize services
	screenerService := service.NewScreenerService()
	historicalService := service.NewHistoricalService()
	watchlistService := service.NewWatchlistService()
	companyInfoService := service.NewCompanyInfoService()
	fundamentalDataService := service.NewFundamentalDataService()

	// Root route
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"message": "Screener Backend API",
			"version": "1.0.0",
			"endpoints": fiber.Map{
				"health": "/api/health",
				"docs":   "See API documentation",
			},
		})
	})

	// Convenience health check at root level (redirects to /api/health)
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"message": "Server is running",
		})
	})

	// Public routes
	public := app.Group("/api")
	{
		// Register filtering routes (inside-day, high-volume-quarter, high-volume-year, high-volume-ever)
		filtering.SetupInsideDayRoutes(public)
		filtering.SetupHighVolumeQuarterRoutes(public)
		filtering.SetupHighVolumeYearRoutes(public)
		filtering.SetupHighVolumeEverRoutes(public)
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

			// Invalidate symbols cache since screener table may have been updated
			invalidator := caching.NewInvalidationService()
			_ = invalidator.InvalidateSymbols()

			return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
				"success":     true,
				"job_id":      jobID,
				"accepted_at": time.Now().UTC().Format(time.RFC3339),
			})
		})

		// Watchlist price update endpoint (public): trigger price updates for all watchlist items
		public.Post("/admin/watchlist/update-prices", func(c *fiber.Ctx) error {
			fetcher := service.NewFetcherService()
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
			defer cancel()

			jobID, err := fetcher.RunWatchlistPriceUpdate(ctx)
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

		// Company info ingestion endpoint (public): trigger company info fetch for all screener symbols
		public.Post("/admin/ingest/company-data", func(c *fiber.Ctx) error {
			fetcher := service.NewFetcherService()
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
			defer cancel()

			jobID, err := fetcher.RunCompanyInfoIngestion(ctx)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"success": false,
					"error":   "Internal Server Error",
					"message": err.Error(),
				})
			}

			// Invalidate company info cache after ingestion
			invalidator := caching.NewInvalidationService()
			_ = invalidator.InvalidateAllCompanyInfo()

			return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
				"success":     true,
				"job_id":      jobID,
				"accepted_at": time.Now().UTC().Format(time.RFC3339),
			})
		})

		// Fundamental data ingestion endpoint (public): trigger fundamental data fetch for all screener symbols
		public.Post("/admin/ingest/fundamental-data", func(c *fiber.Ctx) error {
			fetcher := service.NewFetcherService()
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
			defer cancel()

			jobID, err := fetcher.RunFundamentalDataIngestion(ctx)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"success": false,
					"error":   "Internal Server Error",
					"message": err.Error(),
				})
			}

			// Invalidate fundamental data cache after ingestion
			invalidator := caching.NewInvalidationService()
			_ = invalidator.InvalidateAllFundamentalData()

			return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
				"success":     true,
				"job_id":      jobID,
				"accepted_at": time.Now().UTC().Format(time.RFC3339),
			})
		})

		// Market statistics aggregation endpoint (public): trigger market aggregation (call every 5 minutes via external cron)
		public.Post("/admin/market-statistics/aggregate", func(c *fiber.Ctx) error {
			fetcher := service.NewFetcherService()
			jobID := fmt.Sprintf("market-aggregation-%d", time.Now().UnixNano())

			// Start aggregation in background to avoid timeout
			go func() {
				ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
				defer cancel()
				_, err := fetcher.RunMarketAggregation(ctx)
				if err != nil {
					// Log error but don't block the response
					fmt.Printf("Market aggregation error: %v\n", err)
				} else {
					// Invalidate market statistics cache after aggregation
					invalidator := caching.NewInvalidationService()
					_ = invalidator.InvalidateMarketStatistics()
				}
			}()

			return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
				"success":     true,
				"job_id":      jobID,
				"accepted_at": time.Now().UTC().Format(time.RFC3339),
				"message":     "Aggregation started in background",
			})
		})

		// Market statistics end-of-day storage endpoint (public): trigger end-of-day storage (call at market close via external cron)
		public.Post("/admin/market-statistics/store-eod", func(c *fiber.Ctx) error {
			statsService := service.NewMarketStatisticsService()
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
			defer cancel()

			err := statsService.StoreEndOfDayStats(ctx)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"success": false,
					"error":   "Internal Server Error",
					"message": err.Error(),
				})
			}

			// Invalidate market statistics cache after storing EOD stats
			invalidator := caching.NewInvalidationService()
			_ = invalidator.InvalidateMarketStatistics()

			return c.JSON(fiber.Map{
				"success":     true,
				"message":     "End-of-day statistics stored successfully",
				"accepted_at": time.Now().UTC().Format(time.RFC3339),
			})
		})

		// Market statistics historical data endpoint (public): get historical market statistics for charting
		public.Get("/market-statistics", func(c *fiber.Ctx) error {
			statsService := service.NewMarketStatisticsService()
			ctx := c.Context()

			// Parse query parameters
			startDateStr := c.Query("startDate")
			endDateStr := c.Query("endDate")

			// Default to last 30 days if not provided
			startDate := time.Now().AddDate(0, 0, -30)
			endDate := time.Now()

			if startDateStr != "" {
				if parsed, err := time.Parse("2006-01-02", startDateStr); err == nil {
					startDate = parsed
				}
			}
			if endDateStr != "" {
				if parsed, err := time.Parse("2006-01-02", endDateStr); err == nil {
					endDate = parsed
				}
			}

			stats, err := statsService.GetHistoricalStats(ctx, startDate, endDate)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"success": false,
					"error":   "Internal Server Error",
					"message": err.Error(),
				})
			}

			return c.JSON(fiber.Map{
				"success": true,
				"data":    stats,
			})
		})

		// Market statistics current day endpoint (public): get today's real-time aggregated stats
		public.Get("/market-statistics/current", func(c *fiber.Ctx) error {
			statsService := service.NewMarketStatisticsService()

			stats, err := statsService.GetCurrentDayStats()
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"success": false,
					"error":   "Internal Server Error",
					"message": err.Error(),
				})
			}

			return c.JSON(fiber.Map{
				"success": true,
				"data":    stats,
			})
		})

		// Market statistics for frontend polling endpoint (public): returns advances, decliners, unchanged
		// Frontend should poll this endpoint every 5 minutes to get real-time market statistics
		public.Get("/market-statistics/live", func(c *fiber.Ctx) error {
			statsService := service.NewMarketStatisticsService()

			stats, err := statsService.GetMarketStatsForFrontend()
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"success": false,
					"error":   "Internal Server Error",
					"message": err.Error(),
				})
			}

			return c.JSON(fiber.Map{
				"success": true,
				"data":    stats,
			})
		})

		// Get screener results with time period filtering (public)
		public.Get("/screener-results", func(c *fiber.Ctx) error {
			resultType := c.Query("type")      // "inside_day", "high_volume_quarter", "high_volume_year", "high_volume_ever"
			period := c.Query("period", "all") // "7d", "30d", "90d", "ytd", "all"

			if resultType == "" {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"success": false,
					"error":   "Bad Request",
					"message": "type query parameter is required",
				})
			}

			historicalService := service.NewHistoricalService()
			symbols, err := historicalService.GetScreenerResults(resultType, period)
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
					"type":    resultType,
					"period":  period,
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

			adrService := indicatorsscreening.NewADRScreeningService()
			symbols, err := adrService.GetSymbolsByADR(rangeParam, interval, lookback, minADR, maxADR)
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

			atrService := indicatorsscreening.NewATRScreeningService()
			symbols, err := atrService.GetSymbolsByATR(rangeParam, interval, lookback, minATR, maxATR)
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

			adrService := indicatorsscreening.NewADRScreeningService()
			adrPercent, err := adrService.GetADRForSymbol(symbol, rangeParam, interval, lookback)
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

			atrService := indicatorsscreening.NewATRScreeningService()
			atrPercent, err := atrService.GetATRForSymbol(symbol, rangeParam, interval, lookback)
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

			volumeService := indicatorsscreening.NewVolumeScreeningService()
			symbols, err := volumeService.GetSymbolsByAvgVolumeDollars(rangeParam, interval, lookback, minVolDollarsM, maxVolDollarsM)
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

			volumeService := indicatorsscreening.NewVolumeScreeningService()
			symbols, err := volumeService.GetSymbolsByAvgVolumePercent(rangeParam, interval, lookback, minVolPercent, maxVolPercent)
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

			volumeService := indicatorsscreening.NewVolumeScreeningService()
			avgVolDollarsM, err := volumeService.GetAvgVolumeDollarsForSymbol(symbol, rangeParam, interval, lookback)
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

			volumeService := indicatorsscreening.NewVolumeScreeningService()
			volPercent, err := volumeService.GetAvgVolumePercentForSymbol(symbol, rangeParam, interval, lookback)
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

		// Company Info routes (public, read-only)
		// Get all company info
		public.Get("/company-info", func(c *fiber.Ctx) error {
			companyInfo, err := companyInfoService.GetAllCompanyInfo()
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"success": false,
					"error":   "Internal Server Error",
					"message": err.Error(),
				})
			}

			return c.JSON(fiber.Map{
				"success": true,
				"data":    companyInfo,
			})
		})

		// Get company info by multiple symbols (POST with JSON body) - must come before /:symbol route
		public.Post("/company-info/symbols", func(c *fiber.Ctx) error {
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

			companyInfo, err := companyInfoService.GetCompanyInfoBySymbols(request.Symbols)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"success": false,
					"error":   "Internal Server Error",
					"message": err.Error(),
				})
			}

			return c.JSON(fiber.Map{
				"success": true,
				"data":    companyInfo,
			})
		})

		// Search company info by name, sector, industry, or symbol - must come before /:symbol route
		public.Get("/company-info/search", func(c *fiber.Ctx) error {
			searchTerm := c.Query("q")
			if searchTerm == "" {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"success": false,
					"error":   "Bad Request",
					"message": "Search term (q) is required",
				})
			}

			companyInfo, err := companyInfoService.SearchCompanyInfo(searchTerm)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"success": false,
					"error":   "Internal Server Error",
					"message": err.Error(),
				})
			}

			return c.JSON(fiber.Map{
				"success": true,
				"data":    companyInfo,
			})
		})

		// Get company info by sector - must come before /:symbol route
		public.Get("/company-info/sector/:sector", func(c *fiber.Ctx) error {
			sector := c.Params("sector")
			if sector == "" {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"success": false,
					"error":   "Bad Request",
					"message": "Sector is required",
				})
			}

			companyInfo, err := companyInfoService.GetCompanyInfoBySector(sector)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"success": false,
					"error":   "Internal Server Error",
					"message": err.Error(),
				})
			}

			return c.JSON(fiber.Map{
				"success": true,
				"data":    companyInfo,
			})
		})

		// Get company info by industry - must come before /:symbol route
		public.Get("/company-info/industry/:industry", func(c *fiber.Ctx) error {
			industry := c.Params("industry")
			if industry == "" {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"success": false,
					"error":   "Bad Request",
					"message": "Industry is required",
				})
			}

			companyInfo, err := companyInfoService.GetCompanyInfoByIndustry(industry)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"success": false,
					"error":   "Internal Server Error",
					"message": err.Error(),
				})
			}

			return c.JSON(fiber.Map{
				"success": true,
				"data":    companyInfo,
			})
		})

		// Get company info by symbol (must be last to avoid matching specific routes)
		public.Get("/company-info/:symbol", func(c *fiber.Ctx) error {
			symbol := c.Params("symbol")
			if symbol == "" {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"success": false,
					"error":   "Bad Request",
					"message": "Symbol is required",
				})
			}

			companyInfo, err := companyInfoService.GetCompanyInfoBySymbol(symbol)
			if err != nil {
				if err.Error() == "record not found" {
					return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
						"success": false,
						"error":   "Not Found",
						"message": "Company info not found for symbol",
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
				"data":    companyInfo,
			})
		})

		// Fundamental Data routes (public, read-only)
		// Get all fundamental data
		public.Get("/fundamental-data", func(c *fiber.Ctx) error {
			fundamentalData, err := fundamentalDataService.GetAllFundamentalData()
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"success": false,
					"error":   "Internal Server Error",
					"message": err.Error(),
				})
			}

			return c.JSON(fiber.Map{
				"success": true,
				"data":    fundamentalData,
			})
		})

		// Get fundamental data by symbol
		public.Get("/fundamental-data/symbol/:symbol", func(c *fiber.Ctx) error {
			symbol := c.Params("symbol")
			if symbol == "" {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"success": false,
					"error":   "Bad Request",
					"message": "Symbol is required",
				})
			}

			fundamentalData, err := fundamentalDataService.GetFundamentalDataBySymbol(symbol)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"success": false,
					"error":   "Internal Server Error",
					"message": err.Error(),
				})
			}

			return c.JSON(fiber.Map{
				"success": true,
				"data":    fundamentalData,
			})
		})

		// Get fundamental data by symbol and statement type
		public.Get("/fundamental-data/symbol/:symbol/type/:statementType", func(c *fiber.Ctx) error {
			symbol := c.Params("symbol")
			statementType := c.Params("statementType")
			if symbol == "" || statementType == "" {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"success": false,
					"error":   "Bad Request",
					"message": "Symbol and statement type are required",
				})
			}

			fundamentalData, err := fundamentalDataService.GetFundamentalDataBySymbolAndType(symbol, statementType)
			if err != nil {
				if err.Error() == "record not found" {
					return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
						"success": false,
						"error":   "Not Found",
						"message": "Fundamental data not found",
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
				"data":    fundamentalData,
			})
		})

		// Get fundamental data by symbol, statement type, and frequency
		public.Get("/fundamental-data/symbol/:symbol/type/:statementType/frequency/:frequency", func(c *fiber.Ctx) error {
			symbol := c.Params("symbol")
			statementType := c.Params("statementType")
			frequency := c.Params("frequency")
			if symbol == "" || statementType == "" || frequency == "" {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"success": false,
					"error":   "Bad Request",
					"message": "Symbol, statement type, and frequency are required",
				})
			}

			fundamentalData, err := fundamentalDataService.GetFundamentalDataBySymbolTypeAndFrequency(symbol, statementType, frequency)
			if err != nil {
				if err.Error() == "record not found" {
					return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
						"success": false,
						"error":   "Not Found",
						"message": "Fundamental data not found",
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
				"data":    fundamentalData,
			})
		})

		// Get fundamental data by statement type
		public.Get("/fundamental-data/type/:statementType", func(c *fiber.Ctx) error {
			statementType := c.Params("statementType")
			if statementType == "" {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"success": false,
					"error":   "Bad Request",
					"message": "Statement type is required",
				})
			}

			fundamentalData, err := fundamentalDataService.GetFundamentalDataByStatementType(statementType)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"success": false,
					"error":   "Internal Server Error",
					"message": err.Error(),
				})
			}

			return c.JSON(fiber.Map{
				"success": true,
				"data":    fundamentalData,
			})
		})

		// Get fundamental data by frequency
		public.Get("/fundamental-data/frequency/:frequency", func(c *fiber.Ctx) error {
			frequency := c.Params("frequency")
			if frequency == "" {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"success": false,
					"error":   "Bad Request",
					"message": "Frequency is required",
				})
			}

			fundamentalData, err := fundamentalDataService.GetFundamentalDataByFrequency(frequency)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"success": false,
					"error":   "Internal Server Error",
					"message": err.Error(),
				})
			}

			return c.JSON(fiber.Map{
				"success": true,
				"data":    fundamentalData,
			})
		})

		// Search fundamental data by symbol
		public.Get("/fundamental-data/search", func(c *fiber.Ctx) error {
			searchTerm := c.Query("q")
			if searchTerm == "" {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"success": false,
					"error":   "Bad Request",
					"message": "Search term (q) is required",
				})
			}

			fundamentalData, err := fundamentalDataService.SearchFundamentalData(searchTerm)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"success": false,
					"error":   "Internal Server Error",
					"message": err.Error(),
				})
			}

			return c.JSON(fiber.Map{
				"success": true,
				"data":    fundamentalData,
			})
		})

		// Get fundamental metrics for a symbol
		public.Get("/fundamental-data/metrics", func(c *fiber.Ctx) error {
			symbol := c.Query("symbol")
			statementType := c.Query("statement_type", "income")
			frequency := c.Query("frequency", "annual")

			if symbol == "" {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"success": false,
					"error":   "Bad Request",
					"message": "Symbol is required",
				})
			}

			metrics, err := fundamentalDataService.GetFundamentalMetrics(symbol, statementType, frequency)
			if err != nil {
				if err.Error() == "record not found" {
					return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
						"success": false,
						"error":   "Not Found",
						"message": "Fundamental data not found",
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
				"data":    metrics,
			})
		})

		// Filter stocks by revenue growth (QoQ/YoY)
		public.Get("/fundamental-data/revenue-growth", func(c *fiber.Ctx) error {
			statementType := c.Query("statement_type", "income")
			frequency := c.Query("frequency", "quarterly")

			var minQoQ, maxQoQ, minYoY, maxYoY *float64

			if minStr := c.Query("min_qoq_growth"); minStr != "" {
				if val, err := strconv.ParseFloat(minStr, 64); err == nil {
					minQoQ = &val
				}
			}
			if maxStr := c.Query("max_qoq_growth"); maxStr != "" {
				if val, err := strconv.ParseFloat(maxStr, 64); err == nil {
					maxQoQ = &val
				}
			}
			if minStr := c.Query("min_yoy_growth"); minStr != "" {
				if val, err := strconv.ParseFloat(minStr, 64); err == nil {
					minYoY = &val
				}
			}
			if maxStr := c.Query("max_yoy_growth"); maxStr != "" {
				if val, err := strconv.ParseFloat(maxStr, 64); err == nil {
					maxYoY = &val
				}
			}

			filter := service.RevenueGrowthFilter{
				MinQoQGrowth:  minQoQ,
				MaxQoQGrowth:  maxQoQ,
				MinYoYGrowth:  minYoY,
				MaxYoYGrowth:  maxYoY,
				StatementType: statementType,
				Frequency:     frequency,
			}

			results, err := fundamentalDataService.GetStocksWithRevenueGrowth(filter)
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
					"stocks": results,
					"count":  len(results),
					"params": filter,
				},
			})
		})

		// Filter stocks by EPS range
		public.Get("/fundamental-data/eps-filter", func(c *fiber.Ctx) error {
			statementType := c.Query("statement_type", "income")
			frequency := c.Query("frequency", "annual")
			date := c.Query("date") // Optional: specific date, or latest if empty

			var minEPS, maxEPS *float64

			if minStr := c.Query("min_eps"); minStr != "" {
				if val, err := strconv.ParseFloat(minStr, 64); err == nil {
					minEPS = &val
				}
			}
			if maxStr := c.Query("max_eps"); maxStr != "" {
				if val, err := strconv.ParseFloat(maxStr, 64); err == nil {
					maxEPS = &val
				}
			}

			filter := service.EPSFilter{
				MinEPS:        minEPS,
				MaxEPS:        maxEPS,
				Date:          date,
				StatementType: statementType,
				Frequency:     frequency,
			}

			results, err := fundamentalDataService.GetStocksWithEPSRange(filter)
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
					"stocks": results,
					"count":  len(results),
					"params": filter,
				},
			})
		})

		// Filter stocks by margin range
		public.Get("/fundamental-data/margin-filter", func(c *fiber.Ctx) error {
			marginType := c.Query("margin_type", "gross") // gross, operating, net
			statementType := c.Query("statement_type", "income")
			frequency := c.Query("frequency", "annual")
			date := c.Query("date") // Optional: specific date, or latest if empty

			var minMargin, maxMargin *float64

			if minStr := c.Query("min_margin"); minStr != "" {
				if val, err := strconv.ParseFloat(minStr, 64); err == nil {
					minMargin = &val
				}
			}
			if maxStr := c.Query("max_margin"); maxStr != "" {
				if val, err := strconv.ParseFloat(maxStr, 64); err == nil {
					maxMargin = &val
				}
			}

			filter := service.MarginFilter{
				MarginType:    marginType,
				MinMargin:     minMargin,
				MaxMargin:     maxMargin,
				Date:          date,
				StatementType: statementType,
				Frequency:     frequency,
			}

			results, err := fundamentalDataService.GetStocksWithMarginRange(filter)
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
					"stocks": results,
					"count":  len(results),
					"params": filter,
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
		// Get historical records by symbol, range, and interval (must be before /historical/:id)
		protected.Get("/historical/by-symbol", func(c *fiber.Ctx) error {
			symbol := c.Query("symbol")
			rangeParam := c.Query("range")
			interval := c.Query("interval")

			if symbol == "" || rangeParam == "" || interval == "" {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"success": false,
					"error":   "Bad Request",
					"message": "symbol, range, and interval query parameters are required",
				})
			}

			historical, err := historicalService.GetHistoricalBySymbolRangeInterval(symbol, rangeParam, interval)
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

		// Watchlist routes
		// Get all watchlists for the authenticated user
		protected.Get("/watchlist", func(c *fiber.Ctx) error {
			userIDStr, ok := c.Locals("userID").(string)
			if !ok {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"success": false,
					"error":   "Unauthorized",
					"message": "User ID not found in token",
				})
			}

			userID, err := uuid.Parse(userIDStr)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"success": false,
					"error":   "Bad Request",
					"message": "Invalid user ID format",
				})
			}

			watchlists, err := watchlistService.GetWatchlistsByUserID(userID)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"success": false,
					"error":   "Internal Server Error",
					"message": err.Error(),
				})
			}

			return c.JSON(fiber.Map{
				"success": true,
				"data":    watchlists,
			})
		})

		// Get a specific watchlist by ID
		protected.Get("/watchlist/:id", func(c *fiber.Ctx) error {
			id := c.Params("id")
			watchlist, err := watchlistService.GetWatchlistByID(id)
			if err != nil {
				if err.Error() == "record not found" {
					return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
						"success": false,
						"error":   "Not Found",
						"message": "Watchlist not found",
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
				"data":    watchlist,
			})
		})

		// Create a new watchlist
		protected.Post("/watchlist", func(c *fiber.Ctx) error {
			userIDStr, ok := c.Locals("userID").(string)
			if !ok {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"success": false,
					"error":   "Unauthorized",
					"message": "User ID not found in token",
				})
			}

			userID, err := uuid.Parse(userIDStr)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"success": false,
					"error":   "Bad Request",
					"message": "Invalid user ID format",
				})
			}

			var watchlist model.Watchlist
			if err := c.BodyParser(&watchlist); err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"success": false,
					"error":   "Bad Request",
					"message": "Invalid request body",
				})
			}

			watchlist.UserID = userID
			if err := watchlistService.CreateWatchlist(&watchlist); err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"success": false,
					"error":   "Internal Server Error",
					"message": err.Error(),
				})
			}

			return c.Status(fiber.StatusCreated).JSON(fiber.Map{
				"success": true,
				"data":    watchlist,
			})
		})

		// Update a watchlist
		protected.Put("/watchlist/:id", func(c *fiber.Ctx) error {
			id := c.Params("id")
			var watchlist model.Watchlist
			if err := c.BodyParser(&watchlist); err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"success": false,
					"error":   "Bad Request",
					"message": "Invalid request body",
				})
			}

			if err := watchlistService.UpdateWatchlist(id, &watchlist); err != nil {
				if err.Error() == "record not found" {
					return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
						"success": false,
						"error":   "Not Found",
						"message": "Watchlist not found",
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
				"data":    watchlist,
			})
		})

		// Delete a watchlist
		protected.Delete("/watchlist/:id", func(c *fiber.Ctx) error {
			id := c.Params("id")
			if err := watchlistService.DeleteWatchlist(id); err != nil {
				if err.Error() == "record not found" {
					return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
						"success": false,
						"error":   "Not Found",
						"message": "Watchlist not found",
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
				"message": "Watchlist deleted successfully",
			})
		})

		// Watchlist Items routes
		// Get all items for a watchlist
		protected.Get("/watchlist/:id/items", func(c *fiber.Ctx) error {
			watchlistIDStr := c.Params("id")
			watchlistID, err := uuid.Parse(watchlistIDStr)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"success": false,
					"error":   "Bad Request",
					"message": "Invalid watchlist ID format",
				})
			}

			items, err := watchlistService.GetWatchlistItems(watchlistID)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"success": false,
					"error":   "Internal Server Error",
					"message": err.Error(),
				})
			}

			return c.JSON(fiber.Map{
				"success": true,
				"data":    items,
			})
		})

		// Get a specific item by ID
		protected.Get("/watchlist/item/:id", func(c *fiber.Ctx) error {
			id := c.Params("id")
			item, err := watchlistService.GetWatchlistItemByID(id)
			if err != nil {
				if err.Error() == "record not found" {
					return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
						"success": false,
						"error":   "Not Found",
						"message": "Item not found",
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
				"data":    item,
			})
		})

		// Add an item to a watchlist
		protected.Post("/watchlist/:id/items", func(c *fiber.Ctx) error {
			watchlistIDStr := c.Params("id")
			watchlistID, err := uuid.Parse(watchlistIDStr)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"success": false,
					"error":   "Bad Request",
					"message": "Invalid watchlist ID format",
				})
			}

			var item model.WatchlistItem
			if err := c.BodyParser(&item); err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"success": false,
					"error":   "Bad Request",
					"message": "Invalid request body",
				})
			}

			if err := watchlistService.AddItemToWatchlist(watchlistID, &item); err != nil {
				if err.Error() == "watchlist not found" {
					return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
						"success": false,
						"error":   "Not Found",
						"message": err.Error(),
					})
				}
				if err.Error() == "item already exists in watchlist" {
					return c.Status(fiber.StatusConflict).JSON(fiber.Map{
						"success": false,
						"error":   "Conflict",
						"message": err.Error(),
					})
				}
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"success": false,
					"error":   "Internal Server Error",
					"message": err.Error(),
				})
			}

			return c.Status(fiber.StatusCreated).JSON(fiber.Map{
				"success": true,
				"data":    item,
			})
		})

		// Update a watchlist item
		protected.Put("/watchlist/item/:id", func(c *fiber.Ctx) error {
			id := c.Params("id")
			var item model.WatchlistItem
			if err := c.BodyParser(&item); err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"success": false,
					"error":   "Bad Request",
					"message": "Invalid request body",
				})
			}

			if err := watchlistService.UpdateWatchlistItem(id, &item); err != nil {
				if err.Error() == "record not found" {
					return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
						"success": false,
						"error":   "Not Found",
						"message": "Item not found",
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
				"data":    item,
			})
		})

		// Delete a watchlist item
		protected.Delete("/watchlist/item/:id", func(c *fiber.Ctx) error {
			id := c.Params("id")
			if err := watchlistService.DeleteWatchlistItem(id); err != nil {
				if err.Error() == "record not found" {
					return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
						"success": false,
						"error":   "Not Found",
						"message": "Item not found",
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
				"message": "Item deleted successfully",
			})
		})

		// Toggle starred status of an item
		protected.Patch("/watchlist/item/:id/star", func(c *fiber.Ctx) error {
			id := c.Params("id")
			item, err := watchlistService.ToggleItemStarred(id)
			if err != nil {
				if err.Error() == "record not found" {
					return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
						"success": false,
						"error":   "Not Found",
						"message": "Item not found",
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
				"data":    item,
			})
		})

		// Get all starred items for the authenticated user
		protected.Get("/watchlist/starred", func(c *fiber.Ctx) error {
			userIDStr, ok := c.Locals("userID").(string)
			if !ok {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"success": false,
					"error":   "Unauthorized",
					"message": "User ID not found in token",
				})
			}

			userID, err := uuid.Parse(userIDStr)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"success": false,
					"error":   "Bad Request",
					"message": "Invalid user ID format",
				})
			}

			items, err := watchlistService.GetStarredItems(userID)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"success": false,
					"error":   "Internal Server Error",
					"message": err.Error(),
				})
			}

			return c.JSON(fiber.Map{
				"success": true,
				"data":    items,
			})
		})

		// Batch update items (useful for price updates)
		protected.Put("/watchlist/items/batch", func(c *fiber.Ctx) error {
			var items []model.WatchlistItem
			if err := c.BodyParser(&items); err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"success": false,
					"error":   "Bad Request",
					"message": "Invalid request body",
				})
			}

			if err := watchlistService.BatchUpdateItems(items); err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"success": false,
					"error":   "Internal Server Error",
					"message": err.Error(),
				})
			}

			return c.JSON(fiber.Map{
				"success": true,
				"message": "Items updated successfully",
				"count":   len(items),
			})
		})
	}
}
