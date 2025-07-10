package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"lk/datafoundation/crud-api/pkg/schema"
	"lk/datafoundation/crud-api/pkg/storageinference"
	"lk/datafoundation/crud-api/pkg/typeinference"
)

func setupTestDB(t *testing.T) *PostgresRepository {
	// Build database URI from environment variables (same as other tests)
	dbURI := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=%s",
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_DB"),
		os.Getenv("POSTGRES_SSL_MODE"))
	
	repo, err := NewPostgresRepositoryFromDSN(dbURI)
	assert.NoError(t, err, "Failed to create repository")
	
	// Check if repo is nil to avoid panic
	if repo == nil {
		t.Fatal("Repository is nil after successful creation")
	}

	// Initialize tables
	err = repo.InitializeTables(context.Background())
	assert.NoError(t, err)

	// Clean up test tables after test and close repository
	t.Cleanup(func() {
		if repo != nil && repo.DB() != nil {
			// Clean up all test tables created during this test
			_, err := repo.DB().Exec(`
				-- Drop all test tables that start with test_
				DO $$ 
				DECLARE 
					table_name TEXT;
				BEGIN 
					FOR table_name IN 
						SELECT tablename FROM pg_tables 
						WHERE schemaname = 'public' 
						AND (tablename LIKE 'test_data_table_%' 
							OR tablename LIKE 'attr_test_%'
							OR tablename = 'test_data_table'
							OR tablename = 'attr_test_entity_test_attribute')
					LOOP 
						EXECUTE 'DROP TABLE IF EXISTS ' || quote_ident(table_name) || ' CASCADE';
					END LOOP;
				END $$;
				
				-- Clean up test entity_attributes entries
				DELETE FROM entity_attributes WHERE entity_id LIKE 'test_%' OR entity_id = 'test_entity';
				
				-- Clean up test attribute_schemas entries  
				DELETE FROM attribute_schemas WHERE table_name LIKE 'test_%' OR table_name = 'test_table';
			`)
			if err != nil {
				t.Logf("Warning: Failed to clean up test data: %v", err)
			}
			
			// Close the repository after cleanup
			repo.Close()
		}
	})

	return repo
}

func TestGetTableList(t *testing.T) {
	repo := setupTestDB(t)
	// Do not defer repo.Close() here - let cleanup handle it

	// Use unique entity and attribute names for this test
	entityID := fmt.Sprintf("test_entity_%d", time.Now().UnixNano())
	attributeName := fmt.Sprintf("test_attribute_%d", time.Now().UnixNano())
	tableName := fmt.Sprintf("attr_%s_%s", entityID, attributeName)

	// Insert a dummy entity attribute
	_, err := repo.DB().Exec(`
		INSERT INTO entity_attributes (entity_id, attribute_name, table_name)
		VALUES ($1, $2, $3)
	`, entityID, attributeName, tableName)
	assert.NoError(t, err)

	// Get the table list
	tableList, err := GetTableList(context.Background(), repo, entityID)
	assert.NoError(t, err)
	assert.Equal(t, []string{tableName}, tableList)
}

func TestGetSchemaOfTable(t *testing.T) {
	repo := setupTestDB(t)
	// Do not defer repo.Close() here - let cleanup handle it

	// Use unique table name for this test
	tableName := fmt.Sprintf("test_table_%d", time.Now().UnixNano())

	// Insert a dummy schema
	schemaInfo := &schema.SchemaInfo{
		StorageType: storageinference.TabularData,
		Fields: map[string]*schema.SchemaInfo{
			"col1": {
				StorageType: storageinference.ScalarData,
				TypeInfo:    &typeinference.TypeInfo{Type: typeinference.StringType},
			},
		},
	}
	schemaJSON, _ := json.Marshal(schemaInfo)

	_, err := repo.DB().Exec(`
		INSERT INTO attribute_schemas (table_name, schema_version, schema_definition)
		VALUES ($1, 1, $2)
	`, tableName, schemaJSON)
	assert.NoError(t, err)

	// Get the schema
	retrievedSchema, err := GetSchemaOfTable(context.Background(), repo, tableName)
	assert.NoError(t, err)
	assert.Equal(t, schemaInfo.StorageType, retrievedSchema.StorageType)
	assert.Equal(t, len(schemaInfo.Fields), len(retrievedSchema.Fields))
}

func TestGetData(t *testing.T) {
	repo := setupTestDB(t)
	// Do not defer repo.Close() here - let cleanup handle it

	// Use unique table name for this test
	tableName := fmt.Sprintf("test_data_table_%d", time.Now().UnixNano())

	// Create a dummy table and insert data
	_, err := repo.DB().Exec(fmt.Sprintf(`
		CREATE TABLE %s (
			id SERIAL PRIMARY KEY,
			col1 TEXT,
			col2 INTEGER
		)
	`, tableName))
	assert.NoError(t, err)

	_, err = repo.DB().Exec(fmt.Sprintf(`
		INSERT INTO %s (col1, col2) VALUES ('val1', 10), ('val2', 20)
	`, tableName))
	assert.NoError(t, err)

	// Get data with a filter
	filters := map[string]interface{}{"col2": 20}
	data, err := GetData(context.Background(), repo, tableName, filters)
	assert.NoError(t, err)
	assert.Len(t, data, 1)
	assert.Equal(t, "val2", data[0]["col1"])

	// Get all data (no filter)
	allData, err := GetData(context.Background(), repo, tableName, nil)
	assert.NoError(t, err)
	assert.Len(t, allData, 2)
}