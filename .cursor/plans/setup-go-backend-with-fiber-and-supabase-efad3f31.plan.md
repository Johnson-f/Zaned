<!-- efad3f31-9f73-462e-8779-cd16511df33e 12ab0cab-c7a5-4621-9a7c-058d83f8026a -->
# Setup Go Backend with Fiber and Supabase JWT Verification

## Project Structure

```
backend/
├── main.go                 # Entry point with Fiber server setup
├── go.mod                  # Go module file
├── go.sum                  # Go dependencies (auto-generated)
├── .env.example            # Environment variables template
├── model/                  # Data models/structs
├── routes/                 # HTTP route handlers
├── service/                # Business logic layer
└── supabase/              # Supabase utilities
    ├── client.go          # Supabase client initialization
    ├── jwt.go             # JWT verification and parsing
    └── middleware.go      # Fiber middleware for JWT auth
```

## Implementation Steps

### 1. Initialize Go Module

- Create `go.mod` in backend folder with module name (e.g., `screener/backend`)
- Set Go version (1.21+)

### 2. Install Dependencies

- `github.com/gofiber/fiber/v2` - Fiber web framework
- `github.com/golang-jwt/jwt/v5` - JWT token parsing
- `github.com/joho/godotenv` - Environment variable loading
- `github.com/supabase-community/supabase-go` - Supabase client (optional, for reference)

### 3. Supabase Utilities (`supabase/` folder)

- **client.go**: Initialize Supabase client with URL and anon key from env
- **jwt.go**: JWT verification functions to decode and validate Supabase JWT tokens
  - Extract user ID from JWT claims
  - Verify token signature using Supabase JWT secret
- **middleware.go**: Fiber middleware that:
  - Extracts JWT from Authorization header
  - Verifies token using Supabase JWT secret
  - Attaches user ID to Fiber context for use in routes

### 4. Main Application (`main.go`)

- Load environment variables using godotenv
- Initialize Fiber app
- Register JWT middleware
- Set up route groups (public and protected)
- Start server on port 8080 (configurable via env)

### 5. Route Structure (`routes/` folder)

- **routes.go**: Basic route setup function
- Example protected route that uses user ID from context
- Example public route for health checks

### 6. Service Layer (`service/` folder)

- **service.go**: Example service structure
- Functions that receive user ID from routes and perform business logic

### 7. Model Layer (`model/` folder)

- **model.go**: Example data models/structs
- User model and other domain models

### 8. Environment Configuration

- Create `.env.example` with:
  - `SUPABASE_URL` - Supabase project URL
  - `SUPABASE_JWT_SECRET` - Supabase JWT secret for verification
  - `PORT` - Server port (default: 8080)

## Key Implementation Details

- JWT verification will use Supabase's JWT secret to verify tokens sent from frontend
- User ID will be extracted from JWT claims and made available in Fiber context
- Middleware will return 401 Unauthorized for invalid/missing tokens
- Protected routes will have access to user ID via `c.Locals("userID")`