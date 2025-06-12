# CRUD API Development Guide

This document provides development guidelines and setup instructions for the CRUD API service.

## Database Setup

### PostgreSQL Setup

1. **Using Docker**:
```bash
# Run PostgreSQL container (already available in docker-compose)
docker-compose up -d postgres
```

2. **Database Information**:
The PostgreSQL container is already configured with a database called `nexoan`. This database is ready for testing and development.

3. **Environment Variables**:
```bash
# Add to your .env file
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DB=nexoan
POSTGRES_TEST_DB_URI="postgresql://postgres:postgres@localhost:5432/nexoan?sslmode=disable"
```

### Running PostgreSQL Tests

1. **Run All Tests**:
```bash
POSTGRES_TEST_DB_URI="postgresql://postgres:postgres@localhost:5432/nexoan?sslmode=disable" go test -v ./db/repository/postgres/...
```

2. **Run Specific Tests**:
```bash
# Run client tests
POSTGRES_TEST_DB_URI="postgresql://postgres:postgres@localhost:5432/nexoan?sslmode=disable" go test -v -run TestNewClient ./db/repository/postgres/...

# Run data insertion tests
POSTGRES_TEST_DB_URI="postgresql://postgres:postgres@localhost:5432/nexoan?sslmode=disable" go test -v -run TestInsertSampleData ./db/repository/postgres/...
```

3. **Run Tests with Coverage**:
```bash
POSTGRES_TEST_DB_URI="postgresql://postgres:postgres@localhost:5432/nexoan?sslmode=disable" go test -v -cover ./db/repository/postgres/...
```

4. **Run Tests with Race Detection**:
```bash
POSTGRES_TEST_DB_URI="postgresql://postgres:postgres@localhost:5432/nexoan?sslmode=disable" go test -v -race ./db/repository/postgres/...
```

## Using the Docker Compose Environment

The project includes a Docker Compose configuration for development and testing:

```bash
# Start all services including PostgreSQL
docker-compose up -d

# Start only PostgreSQL
docker-compose up -d postgres

# Run tests against the Docker Compose environment
POSTGRES_TEST_DB_URI="postgresql://postgres:postgres@localhost:5432/nexoan?sslmode=disable" go test -v ./...
```

## Best Practices

1. **Environment Variables**: Always use environment variables for configuration rather than hardcoded values.
2. **Testing**: Write tests for all new functionality.
3. **Error Handling**: Always check and handle errors appropriately.
4. **Documentation**: Document any non-obvious functionality or design decisions.
5. **Code Style**: Follow Go style guidelines and project conventions. 