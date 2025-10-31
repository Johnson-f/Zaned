package routes

import (
	"screener/backend/service"
	"screener/backend/supabase"

	"github.com/gofiber/fiber/v2"
)

// SetupRoutes configures all routes for the application
func SetupRoutes(app *fiber.App) {
	// Initialize services
	exampleService := service.NewExampleService()

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
		// Example protected route that uses user ID from context
		protected.Get("/user-data", func(c *fiber.Ctx) error {
			// Get user ID from context (set by JWT middleware)
			userID := c.Locals("userID").(string)

			// Use service to get user data
			data, err := exampleService.GetUserData(userID)
			if err != nil {
				// Check if it's a "record not found" error
				if err.Error() == "record not found" {
					return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
						"success": false,
						"error":   "Not Found",
						"message": "No data found for this user",
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
				"data":    data,
			})
		})

		// Get all examples for the authenticated user
		protected.Get("/examples", func(c *fiber.Ctx) error {
			userID := c.Locals("userID").(string)

			examples, err := exampleService.GetUserExamples(userID)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error":   "Internal Server Error",
					"message": err.Error(),
				})
			}

			return c.JSON(fiber.Map{
				"success": true,
				"data":    examples,
			})
		})

		// Create a new example for the authenticated user
		protected.Post("/examples", func(c *fiber.Ctx) error {
			userID := c.Locals("userID").(string)

			var body map[string]string
			if err := c.BodyParser(&body); err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error":   "Bad Request",
					"message": "Invalid request body",
				})
			}

			content := body["content"]
			if content == "" {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error":   "Bad Request",
					"message": "content field is required",
				})
			}

			example, err := exampleService.CreateUserData(userID, content)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error":   "Internal Server Error",
					"message": err.Error(),
				})
			}

			return c.Status(fiber.StatusCreated).JSON(fiber.Map{
				"success": true,
				"data":    example,
			})
		})

		// Example POST route
		protected.Post("/process", func(c *fiber.Ctx) error {
			// Get user ID from context
			userID := c.Locals("userID").(string)

			// Parse request body
			var body map[string]string
			if err := c.BodyParser(&body); err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error":   "Bad Request",
					"message": "Invalid request body",
				})
			}

			// Process the request
			data := body["data"]
			if err := exampleService.ProcessUserRequest(userID, data); err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error":   "Internal Server Error",
					"message": err.Error(),
				})
			}

			return c.JSON(fiber.Map{
				"success": true,
				"message": "Request processed successfully",
			})
		})
	}
}

