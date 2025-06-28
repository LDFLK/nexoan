#!/bin/bash

echo "=== Checking entity_attributes table ==="
docker exec postgres psql -U postgres -d nexoan -c "SELECT * FROM entity_attributes;"

echo -e "\n=== Checking attribute_schemas table ==="
docker exec postgres psql -U postgres -d nexoan -c "SELECT id, table_name, schema_version, created_at, schema_definition::text FROM attribute_schemas;"

echo -e "\n=== Table Descriptions ==="
docker exec postgres psql -U postgres -d nexoan -c "\d+ entity_attributes"
docker exec postgres psql -U postgres -d nexoan -c "\d+ attribute_schemas" 