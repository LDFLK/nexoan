package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/structpb"

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

	// Get data with a filter (all columns)
	filters := map[string]interface{}{"col2": 20}
	anyData, err := repo.GetData(context.Background(), tableName, filters)
	assert.NoError(t, err)
	assert.NotNil(t, anyData)

	// Unmarshal the Any data to get the JSON string
	var structValue structpb.Struct
	err = anyData.UnmarshalTo(&structValue)
	assert.NoError(t, err)

	jsonStr := structValue.Fields["data"].GetStringValue()
	assert.NotEmpty(t, jsonStr)

	// Parse the JSON to verify the structure
	var tabularData map[string]interface{}
	err = json.Unmarshal([]byte(jsonStr), &tabularData)
	assert.NoError(t, err)

	columns := tabularData["columns"].([]interface{})
	rows := tabularData["rows"].([]interface{})

	assert.Equal(t, "id", columns[0])
	assert.Equal(t, "col1", columns[1])
	assert.Equal(t, "col2", columns[2])
	assert.Len(t, rows, 1)

	// Verify the filtered data
	row := rows[0].([]interface{})
	assert.Equal(t, "val2", row[1]) // col1 is at index 1

	// Get all data (no filter)
	allAnyData, err := repo.GetData(context.Background(), tableName, nil)
	assert.NoError(t, err)
	assert.NotNil(t, allAnyData)

	// Unmarshal the Any data for all data
	var allStructValue structpb.Struct
	err = allAnyData.UnmarshalTo(&allStructValue)
	assert.NoError(t, err)

	allJsonStr := allStructValue.Fields["data"].GetStringValue()
	assert.NotEmpty(t, allJsonStr)

	// Parse the JSON for all data
	var allTabularData map[string]interface{}
	err = json.Unmarshal([]byte(allJsonStr), &allTabularData)
	assert.NoError(t, err)

	allColumns := allTabularData["columns"].([]interface{})
	allRows := allTabularData["rows"].([]interface{})

	assert.Equal(t, "id", allColumns[0])
	assert.Equal(t, "col1", allColumns[1])
	assert.Equal(t, "col2", allColumns[2])
	assert.Len(t, allRows, 2)

	// Verify the structure matches the expected tabular format
	row1 := allRows[0].([]interface{})
	row2 := allRows[1].([]interface{})
	assert.Equal(t, "val1", row1[1])      // First row, col1
	assert.Equal(t, float64(10), row1[2]) // First row, col2 (JSON numbers are float64)
	assert.Equal(t, "val2", row2[1])      // Second row, col1
	assert.Equal(t, float64(20), row2[2]) // Second row, col2

	// Test column selection
	selectedColumnsData, err := repo.GetData(context.Background(), tableName, nil, "col1", "col2")
	assert.NoError(t, err)
	assert.NotNil(t, selectedColumnsData)

	// Unmarshal the selected columns data
	var selectedStructValue structpb.Struct
	err = selectedColumnsData.UnmarshalTo(&selectedStructValue)
	assert.NoError(t, err)

	selectedJsonStr := selectedStructValue.Fields["data"].GetStringValue()
	assert.NotEmpty(t, selectedJsonStr)

	// Parse the JSON for selected columns
	var selectedTabularData map[string]interface{}
	err = json.Unmarshal([]byte(selectedJsonStr), &selectedTabularData)
	assert.NoError(t, err)

	selectedColumns := selectedTabularData["columns"].([]interface{})
	selectedRows := selectedTabularData["rows"].([]interface{})

	// Verify only selected columns are returned
	assert.Len(t, selectedColumns, 2)
	assert.Equal(t, "col1", selectedColumns[0])
	assert.Equal(t, "col2", selectedColumns[1])
	assert.Len(t, selectedRows, 2)

	// Verify the data
	selectedRow1 := selectedRows[0].([]interface{})
	selectedRow2 := selectedRows[1].([]interface{})
	assert.Equal(t, "val1", selectedRow1[0])      // First row, col1
	assert.Equal(t, float64(10), selectedRow1[1]) // First row, col2
	assert.Equal(t, "val2", selectedRow2[0])      // Second row, col1
	assert.Equal(t, float64(20), selectedRow2[1]) // Second row, col2

	// Test column selection with filters
	filteredSelectedData, err := repo.GetData(context.Background(), tableName, filters, "col1")
	assert.NoError(t, err)
	assert.NotNil(t, filteredSelectedData)

	// Unmarshal the filtered selected data
	var filteredSelectedStructValue structpb.Struct
	err = filteredSelectedData.UnmarshalTo(&filteredSelectedStructValue)
	assert.NoError(t, err)

	filteredSelectedJsonStr := filteredSelectedStructValue.Fields["data"].GetStringValue()
	assert.NotEmpty(t, filteredSelectedJsonStr)

	// Parse the JSON for filtered selected data
	var filteredSelectedTabularData map[string]interface{}
	err = json.Unmarshal([]byte(filteredSelectedJsonStr), &filteredSelectedTabularData)
	assert.NoError(t, err)

	filteredSelectedColumns := filteredSelectedTabularData["columns"].([]interface{})
	filteredSelectedRows := filteredSelectedTabularData["rows"].([]interface{})

	// Verify only selected column is returned with filter
	assert.Len(t, filteredSelectedColumns, 1)
	assert.Equal(t, "col1", filteredSelectedColumns[0])
	assert.Len(t, filteredSelectedRows, 1)
	assert.Equal(t, "val2", filteredSelectedRows[0].([]interface{})[0]) // Filtered row, col1
}

func TestGetDataTabularFormat(t *testing.T) {
	repo := setupTestDB(t)

	// Use unique table name for this test
	tableName := fmt.Sprintf("test_tabular_table_%d", time.Now().UnixNano())

	// Create a table that matches the original tabular data structure
	_, err := repo.DB().Exec(fmt.Sprintf(`
		CREATE TABLE %s (
			id TEXT,
			name TEXT,
			email TEXT,
			department TEXT
		)
	`, tableName))
	assert.NoError(t, err)

	// Insert data that matches the original format
	_, err = repo.DB().Exec(fmt.Sprintf(`
		INSERT INTO %s (id, name, email, department) VALUES 
		('001', 'John Doe', 'john@example.com', 'Engineering'),
		('002', 'Jane Smith', 'jane@example.com', 'Marketing'),
		('003', 'Bob Wilson', 'bob@example.com', 'Sales')
	`, tableName))
	assert.NoError(t, err)

	// Get all data
	anyData, err := repo.GetData(context.Background(), tableName, nil)
	assert.NoError(t, err)
	assert.NotNil(t, anyData)

	// Unmarshal the Any data to get the JSON string
	var structValue structpb.Struct
	err = anyData.UnmarshalTo(&structValue)
	assert.NoError(t, err)

	jsonStr := structValue.Fields["data"].GetStringValue()
	assert.NotEmpty(t, jsonStr)

	// Parse the JSON to verify the structure
	var tabularData map[string]interface{}
	err = json.Unmarshal([]byte(jsonStr), &tabularData)
	assert.NoError(t, err)

	// Verify the structure matches the original tabular format
	expectedColumns := []string{"id", "name", "email", "department"}
	columns := tabularData["columns"].([]interface{})
	rows := tabularData["rows"].([]interface{})

	// Convert interface{} columns to strings for comparison
	actualColumns := make([]string, len(columns))
	for i, col := range columns {
		actualColumns[i] = col.(string)
	}
	assert.Equal(t, expectedColumns, actualColumns)
	assert.Len(t, rows, 3)

	// Verify the data matches the original input
	expectedRows := [][]interface{}{
		{"001", "John Doe", "john@example.com", "Engineering"},
		{"002", "Jane Smith", "jane@example.com", "Marketing"},
		{"003", "Bob Wilson", "bob@example.com", "Sales"},
	}

	for i, expectedRow := range expectedRows {
		actualRow := rows[i].([]interface{})
		assert.Equal(t, expectedRow, actualRow)
	}

	// Test with filters
	filters := map[string]interface{}{"department": "Engineering"}
	filteredAnyData, err := repo.GetData(context.Background(), tableName, filters)
	assert.NoError(t, err)
	assert.NotNil(t, filteredAnyData)

	// Unmarshal the filtered Any data
	var filteredStructValue structpb.Struct
	err = filteredAnyData.UnmarshalTo(&filteredStructValue)
	assert.NoError(t, err)

	filteredJsonStr := filteredStructValue.Fields["data"].GetStringValue()
	assert.NotEmpty(t, filteredJsonStr)

	// Parse the filtered JSON
	var filteredTabularData map[string]interface{}
	err = json.Unmarshal([]byte(filteredJsonStr), &filteredTabularData)
	assert.NoError(t, err)

	filteredColumns := filteredTabularData["columns"].([]interface{})
	filteredRows := filteredTabularData["rows"].([]interface{})

	// Convert interface{} columns to strings for comparison
	filteredActualColumns := make([]string, len(filteredColumns))
	for i, col := range filteredColumns {
		filteredActualColumns[i] = col.(string)
	}
	assert.Equal(t, expectedColumns, filteredActualColumns)
	assert.Len(t, filteredRows, 1)
	assert.Equal(t, expectedRows[0], filteredRows[0].([]interface{}))
}
