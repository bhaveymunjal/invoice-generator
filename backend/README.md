# Invoice Generator Backend

A well-structured Go backend application for invoice generation and management.

## Project Structure

```
backend/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── auth/
│   │   └── jwt.go               # JWT token handling
│   ├── config/
│   │   └── config.go            # Configuration management
│   ├── constants/
│   │   └── constants.go         # Application constants and enums
│   ├── database/
│   │   └── database.go          # Database initialization and seeding
│   ├── handlers/
│   │   └── handlers.go          # HTTP request handlers
│   ├── middleware/
│   │   └── middleware.go        # HTTP middleware (auth, validation)
│   ├── models/
│   │   └── models.go            # Database models and DTOs
│   ├── routes/
│   │   └── routes.go            # Route definitions
│   └── services/
│       ├── catalog_service.go   # Categories and items business logic
│       ├── dashboard_service.go # Dashboard statistics business logic
│       ├── invoice_service.go   # Invoice business logic
│       └── user_service.go      # User management business logic
├── scripts/
│   └── create_admin.go          # Admin user creation script
├── .env                         # Environment variables
├── go.mod                       # Go module definition
├── go.sum                       # Go module checksums
└── README.md                    # This file
```

## Architecture

This application follows Clean Architecture principles with clear separation of concerns:

- **cmd/**: Application entry points
- **internal/auth/**: Authentication and JWT handling
- **internal/config/**: Configuration management
- **internal/constants/**: Application constants and enums
- **internal/database/**: Database connection and initialization
- **internal/handlers/**: HTTP request handlers (presentation layer)
- **internal/middleware/**: HTTP middleware
- **internal/models/**: Data models and DTOs
- **internal/routes/**: Route definitions and setup
- **internal/services/**: Business logic layer

## Features

- User authentication and authorization
- Invoice creation and management
- Payment tracking
- Category and item management
- Dashboard with statistics
- Admin functionality
- JWT-based authentication
- PostgreSQL database
- RESTful API design

## Environment Variables

Create a `.env` file in the backend directory:

```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=invoice_user
DB_PASSWORD=invoice_password
DB_NAME=invoice_db
DB_SSLMODE=disable
JWT_SECRET=your-secret-key-change-in-production
SERVER_PORT=8080
SERVER_MODE=debug
CORS_ORIGINS=http://localhost:3000,http://localhost:5173
```

## Setup and Installation

1. **Prerequisites**
   - Go 1.24 or higher
   - PostgreSQL database

2. **Clone and setup**
   ```bash
   cd backend
   go mod download
   ```

3. **Database setup**
   - Create a PostgreSQL database
   - Update the `.env` file with your database credentials

4. **Run the application**
   ```bash
   go run cmd/server/main.go
   ```

5. **Create admin user** (optional)
   ```bash
   go run scripts/create_admin.go
   ```

## API Endpoints

### Public Endpoints
- `POST /api/register` - User registration
- `POST /api/login` - User login
- `GET /health` - Health check

### Protected Endpoints
- `GET /api/profile` - Get user profile
- `PUT /api/profile` - Update user profile
- `GET /api/users` - Get all users
- `GET /api/categories` - Get all categories
- `GET /api/items` - Get all items
- `GET /api/invoices` - Get invoices (paginated)
- `GET /api/invoices/:id` - Get single invoice
- `POST /api/invoices` - Create invoice
- `POST /api/invoices/:id/payments` - Add payment
- `GET /api/dashboard` - Get dashboard stats

### Admin Only Endpoints
- `GET /api/admin/stats` - Get admin statistics
- `POST /api/admin/categories` - Create category
- `POST /api/admin/items` - Create item
- `DELETE /api/admin/invoices/:id` - Delete invoice

## Development

The application is structured to be easily maintainable and extensible:

1. **Adding new features**: Create new services in `internal/services/` and corresponding handlers in `internal/handlers/`
2. **Database changes**: Update models in `internal/models/` and run migrations
3. **New routes**: Add routes in `internal/routes/routes.go`
4. **Configuration**: Add new config options in `internal/config/config.go`
5. **Constants**: Add new constants in `internal/constants/constants.go`

## Testing

```bash
go test ./...
```

## Building for Production

```bash
go build -o invoice-generator cmd/server/main.go
```

## License

This project is licensed under the MIT License.
