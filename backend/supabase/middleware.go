package supabase

import (
	"github.com/gofiber/fiber/v2"
)

// JWTAuthMiddleware is a Fiber middleware that verifies JWT tokens from Supabase
func JWTAuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Extract token from Authorization header
		authHeader := c.Get("Authorization")
		tokenString, err := ExtractTokenFromHeader(authHeader)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
				"message": err.Error(),
			})
		}

		// Verify JWT token
		userID, err := VerifyJWT(tokenString)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
				"message": "Invalid token",
			})
		}

		// Attach user ID to context for use in routes
		c.Locals("userID", userID)

		// Continue to next handler
		return c.Next()
	}
}

