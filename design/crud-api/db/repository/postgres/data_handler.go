package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	pb "lk/datafoundation/crud-api/lk/datafoundation/crud-api"
	"lk/datafoundation/crud-api/pkg/schema"
	"lk/datafoundation/crud-api/pkg/typeinference"

	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// UnmarshalAnyToString attempts to unmarshal an Any protobuf message to a string value
func UnmarshalAnyToString(anyValue *anypb.Any) (string, error) {
	if anyValue == nil {
		return "", nil
	}

	var stringValue wrapperspb.StringValue
	if err := anyValue.UnmarshalTo(&stringValue); err != nil {
		return "", fmt.Errorf("error unmarshaling to string value: %v", err)
	}
	return stringValue.Value, nil
}

// UnmarshalTimeBasedValueList attempts to unmarshal a TimeBasedValueList from an Any protobuf message
func UnmarshalTimeBasedValueList(anyValue *anypb.Any) ([]interface{}, error) {
	if anyValue == nil {
		return nil, nil
	}

	var timeBasedValueList pb.TimeBasedValueList
	if err := anyValue.UnmarshalTo(&timeBasedValueList); err != nil {
		return nil, fmt.Errorf("error unmarshaling to TimeBasedValueList: %v", err)
	}

	// Convert TimeBasedValueList to []interface{}
	result := make([]interface{}, len(timeBasedValueList.Values))
	for i, v := range timeBasedValueList.Values {
		result[i] = v
	}
	return result, nil
}

// UnmarshalEntityAttributes unmarshals the attributes map from a protobuf Entity
func UnmarshalEntityAttributes(attributes map[string]*anypb.Any) (map[string]interface{}, error) {
	if attributes == nil {
		return nil, nil
	}

	result := make(map[string]interface{})
	for key, value := range attributes {
		if value == nil {
			continue
		}

		// Try to unmarshal as string first
		if strValue, err := UnmarshalAnyToString(value); err == nil {
			result[key] = strValue
			continue
		}

		// Try to unmarshal as TimeBasedValueList
		if timeBasedValues, err := UnmarshalTimeBasedValueList(value); err == nil {
			result[key] = timeBasedValues
			continue
		}

		log.Printf("Warning: Could not unmarshal attribute %s with type %s", key, value.TypeUrl)
	}

	return result, nil
}

// generateSchema generates schema information for a value
func generateSchema(value *anypb.Any) (*schema.SchemaInfo, error) {
	// Generate schema directly from the Any value
	schemaGenerator := schema.NewSchemaGenerator()
	return schemaGenerator.GenerateSchema(value)
}

// logSchemaInfo logs schema information in a readable format
func logSchemaInfo(schemaInfo *schema.SchemaInfo) {
	if schemaInfo == nil {
		log.Printf("Schema is nil")
		return
	}

	// Log the schema information
	log.Printf("Schema: StorageType=%v, TypeInfo=%v",
		schemaInfo.StorageType,
		schemaInfo.TypeInfo)

	// Convert schema to JSON for logging
	schemaJSON, err := schema.SchemaInfoToJSON(schemaInfo)
	if err != nil {
		log.Printf("Failed to convert schema to JSON: %v", err)
		return
	}

	// Marshal to pretty JSON for better readability
	prettyJSON, err := json.MarshalIndent(schemaJSON, "", "  ")
	if err != nil {
		log.Printf("Failed to marshal schema to JSON: %v", err)
		return
	}

	log.Printf("Schema JSON:\n%s", string(prettyJSON))
}

// isTabularData checks if the data has a valid tabular structure
func isTabularData(value *anypb.Any) (bool, *structpb.Struct, error) {
	// Try to unmarshal as struct
	var dataStruct structpb.Struct
	if err := value.UnmarshalTo(&dataStruct); err != nil {
		return false, nil, fmt.Errorf("failed to unmarshal as struct: %v", err)
	}

	// Check for required fields
	columnsField, hasColumns := dataStruct.Fields["columns"]
	rowsField, hasRows := dataStruct.Fields["rows"]
	if !hasColumns || !hasRows {
		return false, nil, nil
	}

	// Verify columns is a list
	columnsList := columnsField.GetListValue()
	if columnsList == nil {
		return false, nil, nil
	}

	// Verify rows is a list
	rowsList := rowsField.GetListValue()
	if rowsList == nil {
		return false, nil, nil
	}

	// Verify all columns are strings
	for i, col := range columnsList.Values {
		if col.GetStringValue() == "" {
			return false, nil, fmt.Errorf("column %d is not a string", i)
		}
	}

	// Verify all rows have the same number of columns
	columnCount := len(columnsList.Values)
	for i, row := range rowsList.Values {
		rowData := row.GetListValue()
		if rowData == nil {
			return false, nil, fmt.Errorf("row %d is not a list", i)
		}
		if len(rowData.Values) != columnCount {
			return false, nil, fmt.Errorf("row %d has incorrect number of columns", i)
		}
	}

	return true, &dataStruct, nil
}

// validateAndReturnTabularDataTypes validates that all values in each column have consistent types
// and returns a map of column names to their inferred TypeInfo
func validateAndReturnTabularDataTypes(data *structpb.Struct) (map[string]typeinference.TypeInfo, error) {
	columnsList := data.Fields["columns"].GetListValue()
	rowsList := data.Fields["rows"].GetListValue()

	columnTypes := make(map[string]typeinference.TypeInfo)
	
	// If there are no rows, return empty map
	if len(rowsList.Values) == 0 {
		return columnTypes, nil
	}

	// Initialize column types
	for _, col := range columnsList.Values {
		colName := col.GetStringValue()
		columnTypes[colName] = typeinference.TypeInfo{
			Type: typeinference.StringType, // Default to string
			IsNullable: false,
		}
	}

	// Process all rows to determine types
	for _, row := range rowsList.Values {
		rowData := row.GetListValue()
		for i, value := range rowData.Values {
			colName := columnsList.Values[i].GetStringValue()
			currentType := columnTypes[colName]

			switch v := value.Kind.(type) {
			case *structpb.Value_NumberValue:
				num := v.NumberValue
				switch currentType.Type {
				case typeinference.StringType:
					// First number we've seen
					if num == float64(int64(num)) {
						columnTypes[colName] = typeinference.TypeInfo{Type: typeinference.IntType}
					} else {
						columnTypes[colName] = typeinference.TypeInfo{Type: typeinference.FloatType}
					}
				case typeinference.IntType:
					// If we see a float, promote to float
					if num != float64(int64(num)) {
						columnTypes[colName] = typeinference.TypeInfo{Type: typeinference.FloatType}
					}
				case typeinference.FloatType:
					// Already float, no change needed
				default:
					// Mixed types, convert to string
					columnTypes[colName] = typeinference.TypeInfo{
						Type: typeinference.StringType,
						IsNullable: true,
					}
				}
			case *structpb.Value_StringValue:
				str := v.StringValue
				switch currentType.Type {
				case typeinference.StringType:
					// Check if it's a datetime
					if isDateTime(str) {
						columnTypes[colName] = typeinference.TypeInfo{Type: typeinference.DateTimeType}
					}
				case typeinference.DateTimeType:
					// If current string is not a datetime, convert to string
					if !isDateTime(str) {
						columnTypes[colName] = typeinference.TypeInfo{
							Type: typeinference.StringType,
							IsNullable: true,
						}
					}
				default:
					// Mixed types, convert to string
					columnTypes[colName] = typeinference.TypeInfo{
						Type: typeinference.StringType,
						IsNullable: true,
					}
				}
			case *structpb.Value_BoolValue:
				if currentType.Type != typeinference.BoolType && currentType.Type != typeinference.StringType {
					// Mixed types, convert to string
					columnTypes[colName] = typeinference.TypeInfo{
						Type: typeinference.StringType,
						IsNullable: true,
					}
				} else if currentType.Type == typeinference.StringType {
					columnTypes[colName] = typeinference.TypeInfo{Type: typeinference.BoolType}
				}
			default:
				// Unknown type, convert to string
				columnTypes[colName] = typeinference.TypeInfo{
					Type: typeinference.StringType,
					IsNullable: true,
				}
			}
		}
	}

	return columnTypes, nil
}

// isDateTime checks if a string is a valid datetime
func isDateTime(val string) bool {
	_, err := time.Parse(time.RFC3339, val)
	if err == nil {
		return true
	}
	
	// IMPROVEME: https://github.com/LDFLK/nexoan/issues/159
	// Try other common formats
	formats := []string{
		"2006-01-02",
		"2006-01-02 15:04:05",
		"2006/01/02",
		"02/01/2006",
	}
	
	for _, format := range formats {
		if _, err := time.Parse(format, val); err == nil {
			return true
		}
	}
	
	return false
}

// HandleAttributes handles the attributes for an entity
func HandleAttributes(ctx context.Context, repo *PostgresRepository, entityID string, attributes map[string]*pb.TimeBasedValueList) error {
	if attributes == nil {
		return nil
	}

	// Process each attribute
	for attrName, timeBasedValueList := range attributes {
		if timeBasedValueList == nil {
			continue
		}

		// Process each time-based value
		for _, value := range timeBasedValueList.Values {
			if value == nil || value.Value == nil {
				continue
			}

			// Generate schema for the value
			schemaInfo, err := generateSchema(value.Value)
			if err != nil {
				return fmt.Errorf("error generating schema for attribute %s: %v", attrName, err)
			}

			// Log schema information for debugging
			logSchemaInfo(schemaInfo)

			// Handle tabular data
			if err := handleTabularData(ctx, repo, entityID, attrName, value, schemaInfo); err != nil {
				return fmt.Errorf("error handling tabular data for attribute %s: %v", attrName, err)
			}
		}
	}

	return nil
}

// validateDataAgainstSchema validates that the data matches the schema
func validateDataAgainstSchema(data *structpb.Struct, schemaInfo *schema.SchemaInfo) error {
	columnsList := data.Fields["columns"].GetListValue()
	rowsList := data.Fields["rows"].GetListValue()

	// Validate column names match schema
	schemaColumns := make(map[string]bool)
	for fieldName := range schemaInfo.Fields {
		schemaColumns[fieldName] = true
	}

	for _, col := range columnsList.Values {
		colName := col.GetStringValue()
		if !schemaColumns[colName] {
			return fmt.Errorf("column %s not found in schema", colName)
		}
	}

	// Validate data types for each row
	for i, row := range rowsList.Values {
		rowData := row.GetListValue()
		for j, value := range rowData.Values {
			colName := columnsList.Values[j].GetStringValue()
			fieldSchema := schemaInfo.Fields[colName]

			// Validate type
			switch fieldSchema.TypeInfo.Type {
			case typeinference.IntType:
				if v, ok := value.Kind.(*structpb.Value_NumberValue); !ok || v.NumberValue != float64(int64(v.NumberValue)) {
					return fmt.Errorf("row %d, column %s: expected integer, got %v", i, colName, value)
				}
			case typeinference.FloatType:
				if _, ok := value.Kind.(*structpb.Value_NumberValue); !ok {
					return fmt.Errorf("row %d, column %s: expected float, got %v", i, colName, value)
				}
			case typeinference.BoolType:
				if _, ok := value.Kind.(*structpb.Value_BoolValue); !ok {
					return fmt.Errorf("row %d, column %s: expected boolean, got %v", i, colName, value)
				}
			case typeinference.DateTimeType:
				if v, ok := value.Kind.(*structpb.Value_StringValue); !ok || !isDateTime(v.StringValue) {
					return fmt.Errorf("row %d, column %s: expected datetime, got %v", i, colName, value)
				}
			}
		}
	}

	return nil
}

// compareSchemas compares two schemas and returns true if they are compatible
func compareSchemas(existing, newSchema *schema.SchemaInfo) (bool, error) {
	if existing.StorageType != newSchema.StorageType {
		return false, fmt.Errorf("storage type mismatch: existing=%s, newSchema=%s", 
			existing.StorageType, newSchema.StorageType)
	}

	// Check all existing columns are present in newSchema
	for fieldName, existingField := range existing.Fields {
		newField, exists := newSchema.Fields[fieldName]
		if !exists {
			// Missing column in newSchema
			return false, fmt.Errorf("column %s missing in newSchema", fieldName)
		}

		// Check type compatibility
		if !isTypeCompatible(existingField.TypeInfo.Type, newField.TypeInfo.Type) {
			return false, fmt.Errorf("incompatible type for column %s: existing=%s, newSchema=%s",
				fieldName, existingField.TypeInfo.Type, newField.TypeInfo.Type)
		}

		// Check nullability
		if !existingField.TypeInfo.IsNullable && newField.TypeInfo.IsNullable {
			return false, fmt.Errorf("column %s cannot be changed from NOT NULL to NULL", fieldName)
		}
	}

	return true, nil
}

// isTypeCompatible checks if two types are compatible for schema evolution
func isTypeCompatible(existingType, newType typeinference.DataType) bool {
	// Same type is always compatible
	if existingType == newType {
		return true
	}

	// Type promotion rules
	switch existingType {
	case typeinference.IntType:
		// Int can be promoted to float
		return newType == typeinference.FloatType
	case typeinference.StringType:
		// String can accept any type
		return true
	case typeinference.DateTimeType:
		// DateTime can only be datetime or string
		return newType == typeinference.StringType
	}

	return false
}

// handleTabularData processes tabular data attributes
func handleTabularData(ctx context.Context, repo *PostgresRepository, entityID, attrName string, value *pb.TimeBasedValue, schemaInfo *schema.SchemaInfo) error {
	// Generate table name
	tableName := fmt.Sprintf("attr_%s_%s", sanitizeIdentifier(entityID), sanitizeIdentifier(attrName))

	// Convert schema to columns
	columns := schemaToColumns(schemaInfo)

	// Check if table exists
	exists, err := repo.TableExists(ctx, tableName)
	if err != nil {
		return fmt.Errorf("error checking table existence: %v", err)
	}

	if exists {
		// Get existing schema
		var schemaJSON []byte
		err = repo.DB().QueryRowContext(ctx,
			`SELECT schema_definition FROM attribute_schemas WHERE table_name = $1 ORDER BY schema_version DESC LIMIT 1`,
			tableName).Scan(&schemaJSON)
		if err != nil {
			return fmt.Errorf("error getting existing schema: %v", err)
		}

		var existingSchema schema.SchemaInfo
		if err := json.Unmarshal(schemaJSON, &existingSchema); err != nil {
			return fmt.Errorf("error unmarshaling existing schema: %v", err)
		}

		// Compare schemas
		compatible, err := compareSchemas(&existingSchema, schemaInfo)
		if err != nil {
			return fmt.Errorf("schema compatibility check failed: %v", err)
		}

		if !compatible {
			return fmt.Errorf("incompatible schema changes detected")
		}

		// Validate data against existing schema
		var tabularStruct structpb.Struct
		if err := value.Value.UnmarshalTo(&tabularStruct); err != nil {
			return fmt.Errorf("error unmarshaling tabular data: %v", err)
		}

		if err := validateDataAgainstSchema(&tabularStruct, &existingSchema); err != nil {
			return fmt.Errorf("data validation failed: %v", err)
		}
	} else {
		// Create new table
		if err := repo.CreateDynamicTable(ctx, tableName, columns); err != nil {
			return fmt.Errorf("error creating table: %v", err)
		}

		// Store schema information
		schemaJSON, err := json.Marshal(schemaInfo)
		if err != nil {
			return fmt.Errorf("error marshaling schema: %v", err)
		}

		// Insert schema record
		_, err = repo.DB().ExecContext(ctx,
			`INSERT INTO attribute_schemas (table_name, schema_version, schema_definition)
			VALUES ($1, $2, $3)`,
			tableName, 1, schemaJSON)
		if err != nil {
			return fmt.Errorf("error storing schema: %v", err)
		}
	}

	// Create entity attribute record if it doesn't exist
	var attributeID int
	err = repo.DB().QueryRowContext(ctx,
		`INSERT INTO entity_attributes (entity_id, attribute_name, table_name)
		VALUES ($1, $2, $3)
		ON CONFLICT (entity_id, attribute_name) DO UPDATE
		SET table_name = EXCLUDED.table_name
		RETURNING id`,
		entityID, attrName, tableName).Scan(&attributeID)
	if err != nil {
		return fmt.Errorf("error creating entity attribute record: %v", err)
	}

	// Extract data from the TimeBasedValue
	var tabularStruct structpb.Struct
	if err := value.Value.UnmarshalTo(&tabularStruct); err != nil {
		return fmt.Errorf("error unmarshaling tabular data: %v", err)
	}

	// Extract columns and rows
	columnsValue := tabularStruct.Fields["columns"].GetListValue()
	rowsValue := tabularStruct.Fields["rows"].GetListValue()

	if columnsValue == nil || rowsValue == nil {
		return fmt.Errorf("invalid tabular data format")
	}

	// Convert columns to string slice
	columnNames := make([]string, len(columnsValue.Values))
	for i, col := range columnsValue.Values {
		columnNames[i] = sanitizeIdentifier(col.GetStringValue())
	}

	// Convert rows to [][]interface{}
	rows := make([][]interface{}, len(rowsValue.Values))
	for i, row := range rowsValue.Values {
		rowList := row.GetListValue()
		if rowList == nil {
			return fmt.Errorf("invalid row format at index %d", i)
		}

		rows[i] = make([]interface{}, len(rowList.Values))
		for j, cell := range rowList.Values {
			switch cell.Kind.(type) {
			case *structpb.Value_StringValue:
				rows[i][j] = cell.GetStringValue()
			case *structpb.Value_NumberValue:
				rows[i][j] = cell.GetNumberValue()
			case *structpb.Value_BoolValue:
				rows[i][j] = cell.GetBoolValue()
			default:
				rows[i][j] = cell.GetStringValue()
			}
		}
	}

	// Insert the data
	if err := repo.InsertTabularData(ctx, tableName, attributeID, columnNames, rows); err != nil {
		return fmt.Errorf("error inserting tabular data: %v", err)
	}

	return nil
}

// schemaToColumns converts a schema to database columns
func schemaToColumns(schemaInfo *schema.SchemaInfo) []Column {
	var columns []Column

	for fieldName, field := range schemaInfo.Fields {
		var colType string
		switch field.TypeInfo.Type {
		case typeinference.IntType:
			colType = "INTEGER"
		case typeinference.FloatType:
			colType = "DOUBLE PRECISION"
		case typeinference.StringType:
			colType = "TEXT"
		case typeinference.BoolType:
			colType = "BOOLEAN"
		case typeinference.DateType:
			colType = "DATE"
		case typeinference.DateTimeType:
			colType = "TIMESTAMP WITH TIME ZONE"
		default:
			colType = "TEXT"
		}

		if field.TypeInfo.IsNullable {
			colType += " NULL"
		} else {
			colType += " NOT NULL"
		}

		columns = append(columns, Column{
			Name: sanitizeIdentifier(fieldName),
			Type: colType,
		})
	}

	return columns
}

// sanitizeIdentifier makes a string safe for use as a PostgreSQL identifier
// IMPROVEME: https://github.com/LDFLK/nexoan/issues/160
func sanitizeIdentifier(s string) string {
	// Replace invalid characters with underscores
	safe := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			return r
		}
		return '_'
	}, strings.ToLower(s))

	// Ensure it doesn't start with a number
	if len(safe) > 0 && safe[0] >= '0' && safe[0] <= '9' {
		safe = "_" + safe
	}

	return safe
}

// Column represents a database column definition
type Column struct {
	Name string
	Type string
}

// GetTableList retrieves a list of attribute tables for a given entity ID.
func GetTableList(ctx context.Context, repo *PostgresRepository, entityID string) ([]string, error) {
	query := `
		SELECT table_name
		FROM entity_attributes
		WHERE entity_id = $1
	`
	rows, err := repo.DB().QueryContext(ctx, query, entityID)
	if err != nil {
		return nil, fmt.Errorf("error querying for table list: %v", err)
	}
	defer rows.Close()

	var tableList []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, fmt.Errorf("error scanning table name: %v", err)
		}
		tableList = append(tableList, tableName)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over table list rows: %v", err)
	}

	return tableList, nil
}

// GetSchemaOfTable retrieves the schema for a given attribute table.
func GetSchemaOfTable(ctx context.Context, repo *PostgresRepository, tableName string) (*schema.SchemaInfo, error) {
	query := `
		SELECT schema_definition
		FROM attribute_schemas
		WHERE table_name = $1
		ORDER BY schema_version DESC
		LIMIT 1
	`
	var schemaJSON []byte
	err := repo.DB().QueryRowContext(ctx, query, tableName).Scan(&schemaJSON)
	if err != nil {
		return nil, fmt.Errorf("error getting schema for table %s: %v", tableName, err)
	}

	var schemaInfo schema.SchemaInfo
	if err := json.Unmarshal(schemaJSON, &schemaInfo); err != nil {
		return nil, fmt.Errorf("error unmarshaling schema for table %s: %v", tableName, err)
	}

	return &schemaInfo, nil
}

// GetData retrieves data from a table with optional filters.
func GetData(ctx context.Context, repo *PostgresRepository, tableName string, filters map[string]interface{}) ([]map[string]interface{}, error) {
	// Base query
	query := fmt.Sprintf("SELECT * FROM %s", sanitizeIdentifier(tableName))

	var args []interface{}
	var whereClauses []string
	argCount := 1

	// Add filters to the query
	for key, value := range filters {
		whereClauses = append(whereClauses, fmt.Sprintf("%s = $%d", sanitizeIdentifier(key), argCount))
		args = append(args, value)
		argCount++
	}

	if len(whereClauses) > 0 {
		query += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	// Execute the query
	rows, err := repo.DB().QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error querying data from %s: %v", tableName, err)
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("error getting columns from %s: %v", tableName, err)
	}

	var results []map[string]interface{}
	for rows.Next() {
		rowValues := make([]interface{}, len(columns))
		rowPointers := make([]interface{}, len(columns))
		for i := range rowValues {
			rowPointers[i] = &rowValues[i]
		}

		if err := rows.Scan(rowPointers...); err != nil {
			return nil, fmt.Errorf("error scanning row: %v", err)
		}

		rowData := make(map[string]interface{})
		for i, colName := range columns {
			val := rowValues[i]
			// Handle byte slices (common for text, json, etc.)
			if b, ok := val.([]byte); ok {
				rowData[colName] = string(b)
			} else {
				rowData[colName] = val
			}
		}
		results = append(results, rowData)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %v", err)
	}

	return results, nil
}
