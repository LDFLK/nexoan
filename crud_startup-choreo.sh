#!/bin/sh
set -e

# CRUD Service Startup Script - CHOREO ENVIRONMENT
# This script is specifically configured for the Choreo environment
# and includes database cleanup, testing, and service startup for Choreo.

# Function to clean all databases (drops table structures, not just data)
clean_databases() {
    local phase=$1
    echo "=== Cleaning Databases $phase (Choreo Environment) ==="
    
    # Clean MongoDB collections
    echo "Cleaning MongoDB collections (Choreo)..."
    mongosh --eval "db = db.getSiblingDB(\"$MONGO_DB_NAME\"); if (db.metadata_choreo_test) { db.metadata_choreo_test.drop(); print(\"Dropped metadata_choreo_test collection\"); }" $MONGO_URI || echo "MongoDB cleanup completed"
    
    # Clean Neo4j database
    echo "Cleaning Neo4j database (Choreo)..."
    cypher-shell -u $NEO4J_USER -p $NEO4J_PASSWORD -a $NEO4J_URI "MATCH (n) DETACH DELETE n" || echo "Neo4j cleanup completed"
    
    # Clean attr_ prefix tables by dropping them completely
    echo "Cleaning attr_ prefix tables by dropping them completely (Choreo)..."
    PGPASSWORD=$POSTGRES_PASSWORD psql -h $POSTGRES_HOST -p $POSTGRES_PORT -U $POSTGRES_USER -d $POSTGRES_DB -c "DO \$\$ DECLARE r RECORD; BEGIN FOR r IN (SELECT tablename FROM pg_tables WHERE schemaname = 'public' AND tablename LIKE 'attr_%') LOOP EXECUTE 'DROP TABLE IF EXISTS ' || quote_ident(r.tablename) || ' CASCADE'; RAISE NOTICE 'Dropped attr_ table: %', r.tablename; END LOOP; END \$\$;" || echo "PostgreSQL attr_ table cleanup completed"
    
    # Clean core tables by dropping them completely
    echo "Cleaning core tables (attribute_schemas, entity_attributes)..."
    PGPASSWORD=$POSTGRES_PASSWORD psql -h $POSTGRES_HOST -p $POSTGRES_PORT -U $POSTGRES_USER -d $POSTGRES_DB -c "DROP TABLE IF EXISTS attribute_schemas CASCADE; DROP TABLE IF EXISTS entity_attributes CASCADE;" || echo "PostgreSQL table cleanup completed"
    
    echo "Database cleanup $phase completed! (Choreo)"
}

echo "=== CRUD Service Startup (Choreo Environment) ==="
echo "Running tests with environment:"
echo "NEO4J_URI: $NEO4J_URI"
echo "MONGO_URI: $MONGO_URI"
echo "POSTGRES_HOST: $POSTGRES_HOST"

# Test MongoDB connection
echo "Testing MongoDB connection (Choreo)..."
until mongosh --eval "db.adminCommand(\"ping\")" $MONGO_URI; do
  echo "Waiting for MongoDB to be ready..."
  sleep 2
done
echo "MongoDB connection successful! (Choreo)"

# Test Neo4j connection
echo "Testing Neo4j connection (Choreo)..."
until cypher-shell -u $NEO4J_USER -p $NEO4J_PASSWORD -a $NEO4J_URI "CALL dbms.components()"; do
  echo "Waiting for Neo4j to be ready..."
  sleep 2
done
echo "Neo4j connection successful! (Choreo)"

# Test PostgreSQL connection
echo "Testing PostgreSQL connection (Choreo)..."
until PGPASSWORD=$POSTGRES_PASSWORD psql -h $POSTGRES_HOST -p $POSTGRES_PORT -U $POSTGRES_USER -d $POSTGRES_DB -c "SELECT 1;" > /dev/null 2>&1; do
  echo "Waiting for PostgreSQL to be ready..."
  sleep 2
done
echo "PostgreSQL connection successful! (Choreo)"

# Clean databases before running tests
clean_databases "Before Tests"

# Run Go tests
echo "=== Running Go Tests (Choreo Environment) ==="
cd /app/nexoan/crud-api
if ! go test -v ./...; then
  echo "❌ Tests failed! (Choreo)"
  exit 1
fi
echo "✅ All tests passed! (Choreo)"

## Clean databases after tests complete
## It is better to not clean the database, or not run the tests at all. 
# clean_databases "After Tests"

echo "=== Starting CRUD Service (Choreo Environment) ==="
exec crud-service 2>&1 | tee /app/crud-service-choreo.log
