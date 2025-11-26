# OpenGIN

> 💡 **Note (α)**  
> Name needs to be proposed, voted and finalized. 

## 🧰 Makefile shortcuts

Use `make help` to see all available targets. Common ones:

- `make dev` — Clean databases, build everything, and start the full stack (databases + services).
- `make up` / `make down` — Start/stop the stack. Use `make down-all` to also remove volumes.
- `make logs` — Tail logs for main services (core, ingestion, read).
- `make e2e` — Run local E2E tests (requires services running). `make e2e-docker` runs them via Docker.
- `make coverage` — Run coverage for Go (Core API) and Ballerina (Ingestion/Read APIs).

### Formatting & linting (Go)

This repository enforces Go formatting with `gofumpt` and line wrapping with `golines` (max line length 120), and linting via `golangci-lint`.

Commands:

```bash
# Install tools once (ensure $GOPATH/bin is on your PATH)
make tools-go

# Format Go code (gofumpt + golines -m 120)
make fmt

# Lint Go code
make lint
```

These targets operate on the Core API module at `opengin/core-api`.

### Git pre-commit hooks

Automatically enforce formatting and linting before every commit using `pre-commit`.

Setup (one time):

```bash
# Ensure Go tools are installed and on PATH
make tools-go

# Install pre-commit and register the hook
make hooks-install
```

What it does:
- Runs `make fmt` (gofumpt + golines with max line length 120) and then `make lint` (golangci-lint) on commit.

Useful commands:

```bash
# Run hooks against all files now
pre-commit run --all-files

# Temporarily bypass hooks for a single commit
git commit -n -m "your message"
```

Notes:
- `make hooks-install` uses `pip` to install `pre-commit` for the current user. Ensure your user base bin is on PATH, e.g. `~/.local/bin` on macOS/Linux:
  - Add to your shell profile, e.g., `export PATH="$HOME/.local/bin:$PATH"`.
- Hooks are configured in `.pre-commit-config.yaml` and rely on the Makefile targets.

## 🚀 Running Services

### 1. Run CORE API Service
-Read about running the [CORE Service](opengin/core-api/README.md)

### 2. Run Read API Service
-Read about running the [Read API](opengin/read-api/README.md)

### 3. Run Ingestion API Service
-Read about running the [Ingestion API](opengin/ingestion-api/README.md)

### 4. Run Swagger-UI  
-Read about running the [Swagger UI](opengin/swagger-ui/README.md)

### 5. Database Cleanup Service
The cleanup service provides a way to clean all databases (PostgreSQL, MongoDB, Neo4j) before and after running tests or services.

**Usage:**
```bash
# Clean databases before starting services
docker-compose --profile cleanup run --rm cleanup /app/cleanup.sh pre

# Clean databases after services complete
docker-compose --profile cleanup run --rm cleanup /app/cleanup.sh post

# Clean databases anytime you need
docker-compose --profile cleanup run --rm cleanup /app/cleanup.sh pre
```

**What it cleans:**
- **PostgreSQL**: `attribute_schemas`, `entity_attributes`, and all `attr_*` tables
- **MongoDB**: `metadata` and `metadata_test` collections  
- **Neo4j**: All nodes and relationships

**Note**: The cleanup service uses the `cleanup` profile, so it won't start automatically with `docker-compose up`.

You can also use Makefile helpers to run these via profiles:

```bash
# Clean databases before starting services
make clean-pre

# Clean databases after finishing work/tests
make clean-post
```

### 6. Database Backup and Restore
The system provides comprehensive backup and restore capabilities for all databases.

**Local Backup Management:**
```bash
# Create backups
./deployment/development/init.sh backup_mongodb
./deployment/development/init.sh backup_postgres
./deployment/development/init.sh backup_neo4j

# Restore from local backups
./deployment/development/init.sh restore_mongodb
./deployment/development/init.sh restore_postgres
./deployment/development/init.sh restore_neo4j
```

**GitHub Integration:**
```bash
# Restore from GitHub releases
./deployment/development/init.sh restore_from_github 0.0.1
./deployment/development/init.sh list_github_versions
```

For detailed backup and restore documentation, see [Backup Integration Guide](docs/deployment/BACKUP_INTEGRATION.md).

---

## Run a sample query with CURL

### Ingestion API

**Create**

```bash
curl -X POST http://localhost:8080/entities \
-H "Content-Type: application/json" \
-d '{
  "id": "12345",
  "kind": {
    "major": "example",
    "minor": "test"
  },
  "created": "2024-03-17T10:00:00Z",
  "terminated": "",
  "name": {
    "startTime": "2024-03-17T10:00:00Z",
    "endTime": "",
    "value": {
      "typeUrl": "type.googleapis.com/google.protobuf.StringValue",
      "value": "entity-name"
    }
  },
  "metadata": [
    {"key": "owner", "value": "test-user"},
    {"key": "version", "value": "1.0"},
    {"key": "developer", "value": "V8A"}
  ],
  "attributes": [],
  "relationships": []
}'
```

**Read**

```bash
curl -X GET http://localhost:8080/entities/12345
```

**Update**

> TODO: The update creates a new record and that's a bug, please fix it. 

```bash
curl -X PUT http://localhost:8080/entities/12345 \
  -H "Content-Type: application/json" \
  -d '{
    "id": "12345",
    "kind": {
      "major": "example",
      "minor": "test"
    },
    "created": "2024-03-18T00:00:00Z",
    "name": {
      "startTime": "2024-03-18T00:00:00Z",
      "value": "entity-name"
    },
    "metadata": [
      {"key": "version", "value": "5.0"}
    ]
  }'
```

**Delete**

```bash
curl -X DELETE http://localhost:8080/entities/12345
```

### Read API 

**Retrieve Metadata**

```bash
curl -X GET "http://localhost:8081/v1/entities/12345/metadata"
```

## Run E2E Tests

Make sure the CORE server and the API server are running. 

Note when making a call to ReadEntity, the ReadEntityRequest must be in the following format (output can be one or more of metadata, relationships, attributes):

ReadEntityRequest readEntityRequest = {
    entity: {
        id: entityId,
        kind: {},
        created: "",
        terminated: "",
        name: {
            startTime: "",
            endTime: "",
            value: check pbAny:pack("")
        },
        metadata: [],
        attributes: [],
        relationships: []
    },
    output: ["relationships"]
};

### Run Ingestion API Tests

```bash
cd opengin/tests/e2e
python basic_core_tests.py
```

### Run Read API Tests

```bash
cd opengin/tests/e2e
python basic_read_tests.py
```

## Implementation Progress

[Track Progress](https://github.com/LDFLK/nexoan/issues/29)
