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
# Add to your .env file or export in your shell
export POSTGRES_HOST=localhost
export POSTGRES_PORT=5432
export POSTGRES_USER=postgres
export POSTGRES_PASSWORD=postgres
export POSTGRES_DB=nexoan
export POSTGRES_SSL_MODE=disable
```

### PostgreSQL Table Structure

The database uses a dynamic table structure to handle different types of attributes. Here are the core tables:

1. **entity_attributes** - Maps entities to their attributes
```sql
CREATE TABLE entity_attributes (
    id SERIAL PRIMARY KEY,
    entity_id VARCHAR(255) NOT NULL,
    attribute_name VARCHAR(255) NOT NULL,
    table_name VARCHAR(255) NOT NULL,
    schema_version INT NOT NULL DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(entity_id, attribute_name)
);
```

2. **attribute_schemas** - Stores schema definitions for each attribute
```sql
CREATE TABLE attribute_schemas (
    id SERIAL PRIMARY KEY,
    table_name VARCHAR(255) NOT NULL,
    schema_version INT NOT NULL,
    schema_definition JSONB NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(table_name, schema_version)
);
```

3. **Dynamic Attribute Tables** - Created automatically for each attribute type
```sql
-- Example: attr_emp_data_employee_records
CREATE TABLE attr_emp_data_employee_records (
    id SERIAL PRIMARY KEY,
    entity_attribute_id INTEGER REFERENCES entity_attributes(id),
    -- Dynamic columns based on the attribute schema
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

To view the tables in your database:
```bash
# Connect to PostgreSQL
docker exec -it postgres psql -U postgres -d nexoan

# List all tables
\dt

# View table structure
\d entity_attributes
\d attribute_schemas
\d+ attr_*  # Shows all attribute tables
```

### Running PostgreSQL Tests

1. **Run All Tests**:
```bash
# Make sure environment variables are set first
go test -v ./db/repository/postgres/...
```

2. **Run Specific Tests**:
```bash
# Run repository tests
go test -v -run TestNewPostgresRepository ./db/repository/postgres/...

# Run data insertion tests
go test -v -run TestInsertSampleData ./db/repository/postgres/...
```

3. **Run Tests with Coverage**:
```bash
go test -v -cover ./db/repository/postgres/...
```

4. **Run Tests with Race Detection**:
```bash
go test -v -race ./db/repository/postgres/...
```

## Using the Docker Compose Environment

The project includes a Docker Compose configuration for development and testing:

```bash
# Start all services including PostgreSQL
docker-compose up -d

# Start only PostgreSQL
docker-compose up -d postgres

# Run tests against the Docker Compose environment
# Note: The docker-compose.yml already sets the required environment variables
go test -v ./...
```
