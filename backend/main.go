package main

import (
	"log"
	"os"
	"os/signal"
	"screener/backend/database"
	"screener/backend/model"
	"screener/backend/routes"
	"screener/backend/service/caching"
	"screener/backend/supabase"
	"strings"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	// Initialize database connection
	if err := database.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	log.Println("Database connection established")

	// Initialize Redis cache connection
	if err := caching.InitRedis(); err != nil {
		log.Printf("Warning: Failed to initialize Redis cache: %v. Continuing without cache.", err)
	} else {
		log.Println("Redis cache connection established")
	}

	// Run database migrations
	if err := database.Migrate(&model.Screener{}, &model.Historical{}, &model.Watchlist{}, &model.WatchlistItem{}, &model.CompanyInfo{}, &model.FundamentalData{}, &model.MarketStatistics{}, &model.ScreenerResult{}); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	log.Println("Database migrations completed")

	// Initialize Supabase client (optional, for reference)
	if err := supabase.InitClient(); err != nil {
		log.Printf("Warning: Failed to initialize Supabase client: %v", err)
	}

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName: "Screener Backend",
	})

	// Middleware
	app.Use(logger.New())

	// CORS configuration - use environment variable for allowed origins
	allowedOriginsEnv := os.Getenv("ALLOWED_ORIGINS")

	// Default origins: allow localhost for development and zaned.space for production
	var allowedOrigins string
	if allowedOriginsEnv == "" {
		// Default: allow localhost and zaned.space
		allowedOrigins = "http://localhost:3000,http://localhost:2000,https://zaned.space,https://www.zaned.space"
	} else {
		allowedOrigins = allowedOriginsEnv
	}

	// Parse comma-separated origins
	originsList := []string{}
	if allowedOrigins != "*" {
		// Split by comma and trim whitespace
		origins := strings.Split(allowedOrigins, ",")
		for _, origin := range origins {
			trimmed := strings.TrimSpace(origin)
			if trimmed != "" {
				originsList = append(originsList, trimmed)
			}
		}
	}

	// When using wildcard (*), credentials cannot be allowed (browser security restriction)
	allowCredentials := allowedOrigins != "*"

	corsConfig := cors.Config{
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET, POST, PUT, DELETE, OPTIONS",
		AllowCredentials: allowCredentials,
	}

	if allowedOrigins == "*" {
		corsConfig.AllowOrigins = "*"
	} else {
		corsConfig.AllowOrigins = strings.Join(originsList, ",")
	}

	app.Use(cors.New(corsConfig))

	// Setup routes
	routes.SetupRoutes(app)

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8090"
	}

	// Channel to listen for interrupt signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on port %s", port)
		if err := app.Listen(":" + port); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	<-quit
	log.Println("Shutting down server gracefully...")

	// Gracefully shutdown the server
	if err := app.Shutdown(); err != nil {
		log.Printf("Error during shutdown: %v", err)
	}

	// Close Redis connection
	if err := caching.CloseRedis(); err != nil {
		log.Printf("Error closing Redis connection: %v", err)
	}

	log.Println("Server stopped")
}
