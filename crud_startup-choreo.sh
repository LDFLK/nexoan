#!/bin/sh
set -e

# CRUD Service Startup Script - CHOREO ENVIRONMENT
# This script is specifically configured for the Choreo environment
# and includes database cleanup, testing, and service startup for Choreo.

echo "=== CRUD Service Startup (Choreo Environment) ==="
echo "Running tests with environment:"
echo "NEO4J_URI: $NEO4J_URI"
echo "MONGO_URI: $MONGO_URI"
echo "POSTGRES_HOST: $POSTGRES_HOST"


echo "=== Starting CRUD Service (Choreo Environment) ==="
exec crud-service 2>&1 | tee /app/crud-service-choreo.log
