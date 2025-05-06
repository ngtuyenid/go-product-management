# Product Management API

A secure RESTful API for product management, featuring JWT authentication, rate limiting, and other security measures.

## Features

- **Clean Architecture**: Clear separation of concerns with layers for business logic, storage, and transport
- **Security Measures**:
  - JWT-based authentication and role-based authorization
  - Rate limiting to prevent DDoS attacks
  - Secure headers
  - CORS configuration
- **RESTful API**:
  - Products management
  - Categories management
  - Real-time statistics
- **Performance**:
  - Efficient memory management with sync.Pool
  - Concurrent operations using goroutines
  - Caching for real-time statistics
- **Database**:
  - PostgreSQL with GORM ORM
  - Connection pooling
  - Transaction management
  - SQL migrations support

## Getting Started

### Prerequisites

- Go 1.20 or higher
- PostgreSQL 12 or higher
- Docker (optional)

### Setup and Installation

1. Clone the repository:
```bash
git clone github.com/thanhnguyen/product-api
cd product-api
```

2. Install dependencies:
```bash
go mod download
```

3. Configure the application:
   - Update the environment variables in the `.env` file as needed

4. Run database migrations:
```bash
cd migrations
chmod +x run.sh
./run.sh --up
```

5. Run the application:
```bash
go run cmd/api/main.go
```

### Docker

You can also run the application using Docker:

```bash
# Build the Docker image
docker build -t product-api .

# Run the container
docker run -p 8080:8080 product-api
```

Or use Docker Compose to run the entire stack:

```bash
docker-compose up -d
```

## Database Migrations

The project includes a database migration system to manage your database schema:

```bash
# Apply all pending migrations
./migrations/run.sh --up

# Rollback the latest migration
./migrations/run.sh --down

# Apply a specific migration
./migrations/run.sh --migration=001_initial_schema

# Rollback a specific migration
./migrations/run.sh --down --migration=001_initial_schema

# Get help
./migrations/run.sh --help
```

Migration files are stored in the `migrations/sql` directory:
- Regular migrations: `NNN_name.sql`
- Rollback migrations: `NNN_name_down.sql`

## API Endpoints

### Public Endpoints

- `GET /health`: Health check

### Protected Endpoints (Require JWT token)

#### Products
- `POST /api/v1/products`: Create a product
- `GET /api/v1/products`: List products with filtering and pagination
- `GET /api/v1/products/:id`: Get a product by ID
- `PUT /api/v1/products/:id`: Update a product
- `DELETE /api/v1/products/:id`: Delete a product

#### Stats (Admin only)
- `GET /api/v1/stats`: Get all statistics
- `GET /api/v1/stats/categories`: Get product counts by category
- `GET /api/v1/stats/wishlist`: Get wishlist counts by product
- `GET /api/v1/stats/top-products`: Get top products
- `POST /api/v1/stats/refresh`: Force a refresh of the statistics

## Project Structure

- `cmd/api`: Application entry point
- `cmd/migrate`: Database migration tool
- `internal/`: Internal packages
  - `business/`: Business logic
    - `entity/`: Domain entities
    - `usecase/`: Business use cases
  - `storage/`: Data storage
    - `postgres/`: PostgreSQL implementation
    - `cache/`: Cache implementation
  - `transport/`: API transport layer
    - `http/`: HTTP handlers and middleware
    - `dto/`: Data Transfer Objects
  - `config/`: Application configuration
- `migrations/`: Database migration files
  - `sql/`: SQL migration scripts
  - `run.sh`: Migration runner script
- `pkg/`: Shared packages
  - `logger/`: Logging functionality

## Security Considerations

This API implements several security measures:

1. **Authentication**: JWT-based authentication with secure token handling
2. **Authorization**: Role-based access control for sensitive operations
3. **Rate Limiting**: Prevents abuse and DoS attacks
4. **Secure Headers**: Protection against common web vulnerabilities
5. **Input Validation**: Thorough validation of all inputs
6. **Database Security**: Parameterized queries to prevent SQL injection
7. **Error Handling**: Secure error handling that doesn't leak sensitive information

## License

This project is licensed under the MIT License - see the LICENSE file for details. 