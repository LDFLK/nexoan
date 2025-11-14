.PHONY: help build build-go build-ballerina test test-go test-ballerina e2e e2e-docker infra-up infra-down services-up services-down up down down-all logs clean-pre clean-post backup-mongodb backup-postgres backup-neo4j restore-mongodb restore-postgres restore-neo4j dev coverage coverage-go coverage-ballerina fmt fmt-go lint lint-go tools-go

# Select docker compose command. Override with: make COMPOSE="docker compose"
COMPOSE ?= docker-compose

# Paths
CRUD_DIR := nexoan/crud-api
UPDATE_DIR := nexoan/update-api
QUERY_DIR := nexoan/query-api
E2E_DIR := nexoan/tests/e2e
DEPLOY_DEV := deployment/development

.DEFAULT_GOAL := help

help:
	@echo "Nexoan — Make targets"
	@echo "------------------------------------------------------------"
	@echo "build               Build all components (Go + Ballerina)"
	@echo "build-go            Build CRUD service (Go)"
	@echo "build-ballerina     Build Update & Query APIs (Ballerina)"
	@echo "test                Run all tests (Go + Ballerina)"
	@echo "test-go             Run Go tests for CRUD API"
	@echo "test-ballerina      Run Ballerina tests for Update & Query APIs"
	@echo "coverage            Run coverage for Go + Ballerina"
	@echo "coverage-go         Run Go coverage (CRUD API) and show summary"
	@echo "coverage-ballerina  Run Ballerina coverage (Update & Query APIs)"
	@echo "fmt                 Format Go code (gofumpt + golines -m 120)"
	@echo "fmt-go              Same as 'fmt' (CRUD API only)"
	@echo "lint                Lint Go code (golangci-lint)"
	@echo "lint-go             Same as 'lint' (CRUD API only)"
	@echo "tools-go            Install Go dev tools: gofumpt, golines, golangci-lint"
	@echo "e2e                 Run E2E tests locally (requires services running)"
	@echo "e2e-docker          Run E2E tests in docker-compose 'e2e' service"
	@echo "infra-up            Start databases (MongoDB, Neo4j, Postgres)"
	@echo "infra-down          Stop databases"
	@echo "services-up         Start services (crud, update, query)"
	@echo "services-down       Stop services"
	@echo "up                  Start full stack (infra + services)"
	@echo "down                Stop stack (keeps volumes)"
	@echo "down-all            Stop stack and remove volumes"
	@echo "logs                Tail logs for main services"
	@echo "clean-pre           Clean databases (pre) using cleanup profile"
	@echo "clean-post          Clean databases (post) using cleanup profile"
	@echo "backup-<db>         Backup mongodb | postgres | neo4j"
	@echo "restore-<db>        Restore mongodb | postgres | neo4j"
	@echo "dev                 One command: clean, build, start full stack, ready for development"
	@echo "------------------------------------------------------------"
	@echo "Tip: override compose command with COMPOSE=\"docker compose\" if needed"

build: build-go build-ballerina

build-go:
	@echo "Building CRUD service (Go)"
	@cd $(CRUD_DIR) && go build ./... && go build -o crud-service cmd/server/service.go cmd/server/utils.go

build-ballerina:
	@echo "Building Update & Query APIs (Ballerina)"
	@cd $(UPDATE_DIR) && bal build
	@cd $(QUERY_DIR) && bal build

test: test-go test-ballerina

test-go:
	@echo "Running Go tests"
	@cd $(CRUD_DIR) && go test -v ./...

test-ballerina:
	@echo "Running Ballerina tests (Update API)"
	@cd $(UPDATE_DIR) && bal test
	@echo "Running Ballerina tests (Query API)"
	@cd $(QUERY_DIR) && bal test

coverage: coverage-go coverage-ballerina

coverage-go:
	@echo "Running Go coverage"
	@cd $(CRUD_DIR) && go test -coverprofile=coverage.out ./...
	@cd $(CRUD_DIR) && go tool cover -func=coverage.out | tail -n 1 || true
	@cd $(CRUD_DIR) && go tool cover -html=coverage.out -o coverage.html
	@echo "Go coverage HTML report: $(CRUD_DIR)/coverage.html"

coverage-ballerina:
	@echo "Running Ballerina coverage (Update API)"
	@cd $(UPDATE_DIR) && bal test --code-coverage
	@echo "Running Ballerina coverage (Query API)"
	@cd $(QUERY_DIR) && bal test --code-coverage

e2e:
	@echo "Running local E2E tests (ensure services are up: make up)"
	@cd $(E2E_DIR) && python3 basic_crud_tests.py && python3 basic_query_tests.py

e2e-docker:
	@echo "Running E2E tests via docker-compose (will build and run dependent services if needed)"
	@$(COMPOSE) up --build -d mongodb neo4j postgres crud update query
	@$(COMPOSE) up --build e2e
	@$(COMPOSE) rm -f e2e || true

infra-up:
	@echo "Starting databases (MongoDB, Neo4j, Postgres)"
	@$(COMPOSE) up -d --build mongodb neo4j postgres

infra-down:
	@echo "Stopping databases"
	@$(COMPOSE) stop mongodb neo4j postgres || true

services-up:
	@echo "Starting services (crud, update, query)"
	@$(COMPOSE) up -d --build crud update query

services-down:
	@echo "Stopping services (crud, update, query)"
	@$(COMPOSE) stop crud update query || true

up: infra-up services-up
	@echo "Full stack started."
	@echo "- CRUD (gRPC): localhost:50051"
	@echo "- Update API:  http://localhost:8080"
	@echo "- Query API:   http://localhost:8081"

logs:
	@$(COMPOSE) logs -f crud update query

down:
	@echo "Stopping stack (keeping volumes)"
	@$(COMPOSE) down

down-all:
	@echo "Stopping stack and removing volumes"
	@$(COMPOSE) down -v

clean-pre:
	@echo "Cleaning databases (pre) via cleanup profile"
	@$(COMPOSE) --profile cleanup run --rm cleanup /app/cleanup.sh pre

clean-post:
	@echo "Cleaning databases (post) via cleanup profile"
	@$(COMPOSE) --profile cleanup run --rm cleanup /app/cleanup.sh post

backup-mongodb:
	@cd $(DEPLOY_DEV) && ./init.sh backup_mongodb

backup-postgres:
	@cd $(DEPLOY_DEV) && ./init.sh backup_postgres

backup-neo4j:
	@cd $(DEPLOY_DEV) && ./init.sh backup_neo4j

restore-mongodb:
	@cd $(DEPLOY_DEV) && ./init.sh restore_mongodb

restore-postgres:
	@cd $(DEPLOY_DEV) && ./init.sh restore_postgres

restore-neo4j:
	@cd $(DEPLOY_DEV) && ./init.sh restore_neo4j

# Formatting & linting
fmt: fmt-go

fmt-go:
	@echo "Formatting Go code (gofumpt + golines -m 120)"
	@cd $(CRUD_DIR) && gofumpt -w .
	@cd $(CRUD_DIR) && golines -m 120 -w .

lint: lint-go

lint-go:
	@echo "Linting Go code (golangci-lint)"
	@cd $(CRUD_DIR) && golangci-lint run ./...

# Install dev tools locally (ensure GOPATH/bin is on your PATH)
tools-go:
	@echo "Installing gofumpt, golines, golangci-lint"
	@go install mvdan.cc/gofumpt@latest
	@go install github.com/segmentio/golines@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# One-shot developer bootstrap: clean -> build -> full stack up -> tail logs hint
# Note: Feel free to interrupt logs with Ctrl+C; services keep running.
dev: clean-pre build up
	@echo "\n✅ Dev environment is up and ready!"
	@echo "- CRUD (gRPC): localhost:50051"
	@echo "- Update API:  http://localhost:8080"
	@echo "- Query API:   http://localhost:8081"
	@echo "- Tail logs:   make logs"
	@echo "- Run E2E:     make e2e or make e2e-docker"
