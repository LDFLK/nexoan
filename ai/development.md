# Development & Usage Guide

## Prerequisites
- Docker & Docker Compose
- Go (for Core API)
- Ballerina (for Ingestion/Read APIs)
- `grpcurl` (for testing gRPC)

## Running the Stack

The easiest way to run the full stack is via Docker Compose.

```bash
# Start all services (Core, Ingestion, Read, Databases)
docker compose up --build

# Stop all services and remove volumes (clean slate)
docker compose down -v
```

### Service Ports
- **Ingestion API**: `http://localhost:8080`
- **Read API**: `http://localhost:8081`
- **Core API (gRPC)**: `localhost:50051`
- **MongoDB**: `localhost:27017`
- **Neo4j**: `http://localhost:7474` (Browser), `bolt://localhost:7687`
- **PostgreSQL**: `localhost:5432`

## Testing

### End-to-End Tests
Python scripts are available for E2E testing.

```bash
# Ingestion API Tests
cd opengin/tests/e2e
python basic_core_tests.py

# Read API Tests
cd opengin/tests/e2e
python basic_read_tests.py
```

### Manual Testing (cURL)

**Create Entity:**
```bash
curl -X POST http://localhost:8080/entities \
-H "Content-Type: application/json" \
-d '{
  "id": "test-entity-1",
  "kind": {"major": "test", "minor": "unit"},
  "name": {"value": "Test Entity", "startTime": "2024-01-01T00:00:00Z", "endTime": ""},
  "metadata": [{"key": "owner", "value": "admin"}]
}'
```

**Read Entity:**
```bash
curl -X GET http://localhost:8081/v1/entities/test-entity-1
```

## Project Structure

- `opengin/core-api`: Go source code for the Core service.
- `opengin/ingestion-api`: Ballerina source code for Ingestion service.
- `opengin/read-api`: Ballerina source code for Read service.
- `opengin/tests`: E2E test scripts.
- `deployment/`: Docker and deployment scripts.
- `docs/`: Detailed documentation.
