# Choreo 

## Local Development and Testing

### Prerequisites

- Docker and Docker Compose installed
- Git repository cloned
- Ports available:
  - 50051 (CRUD service)
  - 8080 (Update service)
  - 27018 (MongoDB choreo)
  - 7475/7688 (Neo4j choreo)
  - 5433 (PostgreSQL choreo)

### Docker Compose Setup (Recommended)

The easiest way to run the choreo services locally is using the dedicated docker-compose file that includes all required databases.

#### Quick Start

```bash
# Clone the repository and navigate to the root directory
cd /path/to/LDFArchitecture

# Start all choreo services (includes databases)
docker-compose -f docker-compose-choreo.yml up --build

# Or start specific services
docker-compose -f docker-compose-choreo.yml up --build crud-choreo update-choreo

# Run in background
docker-compose -f docker-compose-choreo.yml up --build -d
```

#### Available Docker Compose Commands

```bash
# Build services with no cache (recommended after code changes)
docker-compose -f docker-compose-choreo.yml build --no-cache

# Start all services
docker-compose -f docker-compose-choreo.yml up

# Start with build
docker-compose -f docker-compose-choreo.yml up --build

# Start specific services
docker-compose -f docker-compose-choreo.yml up mongodb-choreo neo4j-choreo postgres-choreo
docker-compose -f docker-compose-choreo.yml up crud-choreo update-choreo

# View logs
docker-compose -f docker-compose-choreo.yml logs crud-choreo
docker-compose -f docker-compose-choreo.yml logs update-choreo

# Stop all services
docker-compose -f docker-compose-choreo.yml down

# Stop and remove volumes (clean slate)
docker-compose -f docker-compose-choreo.yml down -v

# Debug: Run interactive shell in a service
docker-compose -f docker-compose-choreo.yml run --entrypoint="" crud-choreo sh
docker-compose -f docker-compose-choreo.yml run --entrypoint="" update-choreo sh

# Run end-to-end tests
docker-compose -f docker-compose-choreo.yml up e2e-choreo
```

#### Service Architecture

The docker-compose setup includes:
- **mongodb-choreo**: MongoDB instance (port 27018)
- **neo4j-choreo**: Neo4j graph database (ports 7475, 7688)
- **postgres-choreo**: PostgreSQL database (port 5433)
- **crud-choreo**: CRUD API service (port 50051)
- **update-choreo**: Update API service (port 8080)
- **e2e-choreo**: End-to-end test runner

#### Database Access

- **MongoDB**: `mongodb://admin:admin123@localhost:27018/admin`
- **Neo4j**: `http://localhost:7475` (user: neo4j, password: neo4j123)
- **PostgreSQL**: `postgresql://postgres:postgres@localhost:5433/ldf_choreo_nexoan`

### Manual Environment Setup (Alternative)

If you prefer to run services manually or against external databases, set up your environment variables:

```bash
# MongoDB Configuration
export MONGO_URI="mongodb+srv://username:password@your-cluster.mongodb.net/?retryWrites=true&w=majority"
export MONGO_DB_NAME="your-mongo-db-name"
export MONGO_COLLECTION="your-mongo-collection-name"
export MONGO_ADMIN_USER="your-mongo-admin-username"
export MONGO_ADMIN_PASSWORD="your-mongo-admin-password"

# Neo4j Configuration
export NEO4J_URI="neo4j+s://your-neo4j-instance.databases.neo4j.io"
export NEO4J_USER="your-neo4j-username"
export NEO4J_PASSWORD="your-neo4j-password"

# PostgreSQL Configuration
export POSTGRES_HOST="your-postgres-host"
export POSTGRES_PORT=5432
export POSTGRES_USER="your-postgres-username"
export POSTGRES_PASSWORD="your-postgres-password"
export POSTGRES_DB="your-postgres-database"
export POSTGRES_SSL_MODE="require"

# Service Configuration
export CRUD_SERVICE_HOST="0.0.0.0"
export CRUD_SERVICE_PORT="50051"
export UPDATE_SERVICE_HOST="0.0.0.0"
export UPDATE_SERVICE_PORT="8080"
export QUERY_SERVICE_HOST="0.0.0.0"
export QUERY_SERVICE_PORT="8081"
```

### Running Services Locally

1. Start the CRUD Service:
```bash
# Build the CRUD service image
# For ARM64 (Apple Silicon):
docker build --platform linux/arm64 -t ldf-choreo-crud-service -f Dockerfile.crud.choreo .
# For AMD64:
docker build --platform linux/amd64 -t ldf-choreo-crud-service -f Dockerfile.crud.choreo .

# Run the CRUD service using environment variables
docker run -d \
  --name ldf-choreo-crud-service \
  -p 50051:50051 \
  -e NEO4J_URI="$NEO4J_URI" \
  -e NEO4J_USER="$NEO4J_USER" \
  -e NEO4J_PASSWORD="$NEO4J_PASSWORD" \
  -e MONGO_URI="$MONGO_URI" \
  -e MONGO_DB_NAME="$MONGO_DB_NAME" \
  -e MONGO_COLLECTION="$MONGO_COLLECTION" \
  -e MONGO_ADMIN_USER="$MONGO_ADMIN_USER" \
  -e MONGO_ADMIN_PASSWORD="$MONGO_ADMIN_PASSWORD" \
  -e POSTGRES_HOST="$POSTGRES_HOST" \
  -e POSTGRES_PORT="$POSTGRES_PORT" \
  -e POSTGRES_USER="$POSTGRES_USER" \
  -e POSTGRES_PASSWORD="$POSTGRES_PASSWORD" \
  -e POSTGRES_DB="$POSTGRES_DB" \
  -e POSTGRES_SSL_MODE="$POSTGRES_SSL_MODE" \
  -e POSTGRES_TEST_DB_URI="$POSTGRES_TEST_DB_URI" \
  -e CRUD_SERVICE_HOST="$CRUD_SERVICE_HOST" \
  -e CRUD_SERVICE_PORT="$CRUD_SERVICE_PORT" \
  ldf-choreo-crud-service
```

2. Start the Update Service:

```bash
# Build the update service image
docker build -t ldf-choreo-update-service -f Dockerfile.update.choreo .

# Run the update service using environment variables
docker run -d -p 8080:8080 \
  --name ldf-choreo-update-service \
  -e CRUD_SERVICE_URL="http://host.docker.internal:$CRUD_SERVICE_PORT" \
  -e UPDATE_SERVICE_HOST="$UPDATE_SERVICE_HOST" \
  -e UPDATE_SERVICE_PORT="$UPDATE_SERVICE_PORT" \
  ldf-choreo-update-service
```

## Testing Locally 

### Running Database services

```bash
docker-compose -f docker-compose-choreo.yml up -d mongodb-choreo neo4j-choreo postgres-choreo
```



## References

1. https://wso2.com/choreo/docs/develop-components/develop-components-with-git/
2. https://wso2.com/choreo/docs/devops-and-ci-cd/manage-configurations-and-secrets/
3. https://wso2.com/choreo/docs/choreo-concepts/ci-cd/
4. [Test Runner](https://wso2.com/choreo/docs/testing/test-components-with-test-runner/)
5. [Choreo Examples](https://github.com/wso2/choreo-samples)
6. [Expose TCP Server via a Service](https://wso2.com/choreo/docs/develop-components/develop-services/expose-a-tcp-server-via-a-service/#step-2-build-and-deploy)