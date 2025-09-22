#!/bin/sh
set -e

# Function to clean all databases (drops table structures, not just data)
clean_databases() {
    local phase=$1
    echo "=== Cleaning Databases $phase ==="
    
    # Clean MongoDB collections
    echo "Cleaning MongoDB collections..."
    mongosh --eval "db = db.getSiblingDB(\"$MONGO_DB_NAME\"); if (db.metadata) { db.metadata.drop(); print(\"Dropped metadata collection\"); } if (db.metadata_test) { db.metadata_test.drop(); print(\"Dropped metadata_test collection\"); }" mongodb://admin:admin123@mongodb:27017/admin?authSource=admin || echo "MongoDB cleanup completed"
    
    # Clean Neo4j database
    echo "Cleaning Neo4j database..."
    cypher-shell -u neo4j -p neo4j123 -a bolt://neo4j:7687 "MATCH (n) DETACH DELETE n" || echo "Neo4j cleanup completed"
    
    # Clean specific PostgreSQL tables by dropping them completely
    echo "Cleaning specific PostgreSQL tables by dropping them completely..."
    # Commented out: Keep attribute_schemas and entity_attributes tables intact
    # PGPASSWORD=$POSTGRES_PASSWORD psql -h $POSTGRES_HOST -p $POSTGRES_PORT -U $POSTGRES_USER -d $POSTGRES_DB -c "DROP TABLE IF EXISTS attribute_schemas CASCADE; DROP TABLE IF EXISTS entity_attributes CASCADE;" || echo "PostgreSQL table cleanup completed"
    
    # Clean attr_ prefix tables by dropping them completely
    echo "Cleaning attr_ prefix tables by dropping them completely..."
    PGPASSWORD=$POSTGRES_PASSWORD psql -h $POSTGRES_HOST -p $POSTGRES_PORT -U $POSTGRES_USER -d $POSTGRES_DB -c "DO \$\$ DECLARE r RECORD; BEGIN FOR r IN (SELECT tablename FROM pg_tables WHERE schemaname = 'public' AND tablename LIKE 'attr_%') LOOP EXECUTE 'DROP TABLE IF EXISTS ' || quote_ident(r.tablename) || ' CASCADE'; RAISE NOTICE 'Dropped attr_ table: %', r.tablename; END LOOP; END \$\$;" || echo "PostgreSQL attr_ table cleanup completed"
    
    echo "Database cleanup $phase completed!"
}

echo "=== CRUD Service Startup ==="
echo "Running tests with environment:"
echo "NEO4J_URI: $NEO4J_URI"
echo "MONGO_URI: $MONGO_URI"
echo "POSTGRES_HOST: $POSTGRES_HOST"

# Test MongoDB connection
echo "Testing MongoDB connection..."
until mongosh --eval "db.adminCommand(\"ping\")" mongodb://admin:admin123@mongodb:27017/admin; do
  echo "Waiting for MongoDB to be ready..."
  sleep 2
done
echo "MongoDB connection successful!"

# Test Neo4j connection
echo "Testing Neo4j connection..."
until cypher-shell -u neo4j -p neo4j123 -a bolt://neo4j:7687 "CALL dbms.components()"; do
  echo "Waiting for Neo4j to be ready..."
  sleep 2
done
echo "Neo4j connection successful!"

# Test PostgreSQL connection
echo "Testing PostgreSQL connection..."
until PGPASSWORD=$POSTGRES_PASSWORD psql -h $POSTGRES_HOST -p $POSTGRES_PORT -U $POSTGRES_USER -d $POSTGRES_DB -c "SELECT 1;" > /dev/null 2>&1; do
  echo "Waiting for PostgreSQL to be ready..."
  sleep 2
done
echo "PostgreSQL connection successful!"

# Clean databases before running tests
clean_databases "Before Tests"

# Run Go tests
echo "=== Running Go Tests ==="
cd /app/nexoan/crud-api
if ! go test -v ./...; then
  echo "❌ Tests failed!"
  exit 1
fi
echo "✅ All tests passed!"

# Clean databases after tests complete
clean_databases "After Tests"

echo "=== Starting CRUD Service ==="
exec crud-service 2>&1 | tee /app/crud-service.log
