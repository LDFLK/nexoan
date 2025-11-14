# Nexoan

A polyglot data platform composed of:
- CRUD gRPC service (Go)
- Public Update and Query REST APIs (Ballerina)
- Backing stores: MongoDB, PostgreSQL, Neo4j
- Docker Compose for local development, backups, E2E tests, and utilities


## Quick start (one command)

```bash
make dev
```
This will:
- Clean all databases (pre) using the cleanup profile
- Build Go CRUD and Ballerina services
- Start the full stack (databases + services)

Endpoints once up:
- CRUD (gRPC): localhost:50051
- Update API:  http://localhost:8080
- Query API:   http://localhost:8081

Tip: Run `make logs` to tail the main service logs. Press Ctrl+C to stop tailing; services keep running.


## Prerequisites
- Docker and docker-compose (or Docker Compose V2; you can override with `make COMPOSE="docker compose"`)
- Go toolchain (for local CRUD development)
- Ballerina 2201.11.0 (for Update/Query services)


## Makefile essentials
Run `make help` anytime to see all targets.

Common workflows:
- Build everything: `make build`
  - Go only: `make build-go`
  - Ballerina only: `make build-ballerina`
- Test all: `make test`
  - Go only: `make test-go`
  - Ballerina only: `make test-ballerina`
- Coverage:
  - All: `make coverage`
  - Go: `make coverage-go` (HTML: `nexoan/crud-api/coverage.html`)
  - Ballerina: `make coverage-ballerina`
- Stack lifecycle:
  - Databases up: `make infra-up`
  - Services up: `make services-up`
  - Full stack up: `make up`
  - Stop (keep volumes): `make down`
  - Stop and remove volumes: `make down-all`
- E2E tests:
  - Local (requires services running): `make e2e`
  - In Docker: `make e2e-docker`
- DB cleanup:
  - Pre: `make clean-pre`
  - Post: `make clean-post`
- Backups/restore:
  - `make backup-mongodb | backup-postgres | backup-neo4j`
  - `make restore-mongodb | restore-postgres | restore-neo4j`


## Formatting & linting

To keep the Go codebase consistent and within a 120‑character line length, use the following Make targets:

- Format Go code:
  - `make fmt` (alias for `fmt-go`) — runs `gofumpt` then `golines -m 120` over `nexoan/crud-api`
- Lint Go code:
  - `make lint` (alias for `lint-go`) — runs `golangci-lint` against `nexoan/crud-api`
- Install required tools (first time only):
  - `make tools-go` — installs `gofumpt`, `golines`, and `golangci-lint`

Notes:
- Line length is kept to 120 chars via `golines -m 120`.
- Ensure your `$GOPATH/bin` is on the `PATH` so the installed tools are available in your shell.

## Project layout
- `nexoan/crud-api` — Go gRPC CRUD service
- `nexoan/update-api` — Ballerina REST Update API
- `nexoan/query-api` — Ballerina REST Query API
- `nexoan/tests/e2e` — Python E2E tests against running services
- `deployment/development` — Dev scripts, Dockerfiles, and backup manager
- `docs/` — Architecture, data types, storage, etc.

More details in:
- CRUD service: [nexoan/crud-api/README.md](nexoan/crud-api/README.md)
- Update API: [nexoan/update-api/README.md](nexoan/update-api/README.md)
- Query API: [nexoan/query-api/README.md](nexoan/query-api/README.md)
- Swagger UI: [nexoan/swagger-ui/README.md](nexoan/swagger-ui/README.md)


## Environment configuration
Each service has an `env.template` you can copy to `.env` for local runs outside Docker.
- CRUD: `nexoan/crud-api/env.template`
- Update: `nexoan/update-api/env.template`
- Query: `nexoan/query-api/env.template`

For Docker Compose, environment is already configured via `docker-compose.yml`.


## Using the cleanup service directly
The cleanup utility resets MongoDB, PostgreSQL, and Neo4j. It runs under the `cleanup` profile and is not started by default.

```bash
docker-compose --profile cleanup run --rm cleanup /app/cleanup.sh pre   # before
# ... work / tests ...
docker-compose --profile cleanup run --rm cleanup /app/cleanup.sh post  # after
```

What it cleans:
- PostgreSQL: `attribute_schemas`, `entity_attributes`, and all `attr_*` tables
- MongoDB: `metadata` and `metadata_test` collections
- Neo4j: All nodes and relationships

You can also trigger the same via `make clean-pre` and `make clean-post`.


## Backup and restore
Local backups/restore are managed by `deployment/development/init.sh`.

Create backups:
```bash
./deployment/development/init.sh backup_mongodb
./deployment/development/init.sh backup_postgres
./deployment/development/init.sh backup_neo4j
```
Restore from local backups:
```bash
./deployment/development/init.sh restore_mongodb
./deployment/development/init.sh restore_postgres
./deployment/development/init.sh restore_neo4j
```
GitHub integration:
```bash
./deployment/development/init.sh restore_from_github 0.0.1
./deployment/development/init.sh list_github_versions
```
See: [docs/deployment/BACKUP_INTEGRATION.md](docs/deployment/BACKUP_INTEGRATION.md)


## Try the APIs quickly
With the stack running (`make up` or `make dev`):

Update API (CRUD operations):
```bash
# Create
curl -X POST http://localhost:8080/entities \
  -H "Content-Type: application/json" \
  -d '{
    "id": "12345",
    "kind": {"major": "example", "minor": "test"},
    "created": "2024-03-17T10:00:00Z",
    "terminated": "",
    "name": {
      "startTime": "2024-03-17T10:00:00Z",
      "endTime": "",
      "value": {"typeUrl": "type.googleapis.com/google.protobuf.StringValue", "value": "entity-name"}
    },
    "metadata": [
      {"key": "owner", "value": "test-user"},
      {"key": "version", "value": "1.0"},
      {"key": "developer", "value": "V8A"}
    ],
    "attributes": [],
    "relationships": []
  }'

# Read
curl -X GET http://localhost:8080/entities/12345

# Update (Note: known issue — see repository issues for status)
curl -X PUT http://localhost:8080/entities/12345 \
  -H "Content-Type: application/json" \
  -d '{
    "id": "12345",
    "kind": {"major": "example", "minor": "test"},
    "created": "2024-03-18T00:00:00Z",
    "name": {"startTime": "2024-03-18T00:00:00Z", "value": "entity-name"},
    "metadata": [{"key": "version", "value": "5.0"}]
  }'

# Delete
curl -X DELETE http://localhost:8080/entities/12345
```

Query API:
```bash
curl -X GET "http://localhost:8081/v1/entities/12345/metadata"
```


## E2E tests
Ensure the stack is up.

```bash
cd nexoan/tests/e2e
python3 basic_crud_tests.py
python3 basic_query_tests.py
```
You can also run E2E inside Docker with `make e2e-docker`.


## Documentation and status
- Architecture and background: `docs/`
- More how-tos: `docs/index.md` and `docs/getting-started.md`
- Open issues and progress: https://github.com/LDFLK/nexoan/issues/29


## License
See [LICENSE.txt](LICENSE.txt).
