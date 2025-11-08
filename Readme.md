# Zaned

A full-stack application with Next.js frontend and Go backend, integrated with Supabase for authentication and database management.

## Project Structure

```
zaned/
├── frontend/          # Next.js application
├── backend/          # Go Fiber API server
├── database/         # Database migrations and schemas
└── docs/             # Documentation
```

## Features

### Frontend (Next.js)
- Next.js 16 with App Router
- Supabase authentication with SSR
- Tailwind CSS for styling
- shadcn/ui components
- TypeScript

### Backend (Go)
- Fiber web framework
- GORM ORM for database operations
- Supabase JWT authentication middleware
- PostgreSQL database support
- RESTful API structure (models, routes, services)

## Prerequisites
- Bun (latest version)
- Go 1.21+
- Supabase account and projects

## Getting Started

### Frontend Setup

1. Navigate to the frontend directory:
```bash
cd frontend
```

2. Install dependencies:
```bash
bun install
```

3. Copy `.env.example` to `.env.local` and fill in your Supabase credentials:
```bash
cp .env.example .env.local
```

4. Update `.env.local` with your Supabase project URL and anon key:
```
NEXT_PUBLIC_SUPABASE_URL=https://your-project.supabase.co
NEXT_PUBLIC_SUPABASE_ANON_KEY=your-anon-key
```

5. Run the development server:
```bash
bun run dev
```

The frontend will be available at `http://localhost:2000`

### Backend Setup

1. Navigate to the backend directory:
```bash
cd backend
```

2. Copy `.env.example` to `.env`:
```bash
cp .env.example .env
```

3. Update `.env` with your Supabase credentials:
```
SUPABASE_URL=https://your-project.supabase.co
SUPABASE_ANON_KEY=your-anon-key
SUPABASE_JWT_SECRET=your-jwt-secret
DATABASE_URL=postgresql://postgres:[password]@db.[project-ref].supabase.co:5432/postgres?sslmode=require
PORT=8080
```

4. Run the backend server:
```bash
go run main.go
```

The backend API will be available at `http://localhost:8080`

## API Endpoints

### Public Routes
- `GET /api/health` - Health check endpoint

### Protected Routes (Require JWT Authentication)
- `GET /api/protected/user-data` - Get user-specific data
- `GET /api/protected/examples` - Get all examples for authenticated user
- `POST /api/protected/examples` - Create a new example
- `POST /api/protected/process` - Process user request

All protected routes require an `Authorization: Bearer <token>` header with a valid Supabase JWT token.

## Database

The backend uses GORM for database operations. Models are defined in `backend/model/` and migrations run automatically on server startup.

### Models
- **User** - User accounts (UUID primary key)
- **Example** - Example data model with user relationship

## Tech Stack

### Frontend
- Next.js 16
- React 19
- TypeScript
- Tailwind CSS
- Supabase Auth (SSR)
- shadcn/ui

### Backend
- Go 1.21+
- Fiber web framework
- GORM ORM
- PostgreSQL (via Supabase)
- JWT authentication

## Development

### Running Both Services

Terminal 1 - Frontend:
```bash
cd frontend
bun run dev
```

Terminal 2 - Backend:
```bash
cd backend
go run main.go
```

## License

Apache-2.0

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

