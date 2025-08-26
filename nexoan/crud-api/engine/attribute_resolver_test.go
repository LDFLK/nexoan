package engine

import (
	"context"
	"fmt"
	"testing"
	"time"

	pb "lk/datafoundation/crud-api/lk/datafoundation/crud-api"
	"lk/datafoundation/crud-api/pkg/schema"
	"lk/datafoundation/crud-api/pkg/storageinference"

	"github.com/stretchr/testify/assert"
)

// createTimeBasedValue creates a TimeBasedValue with the given JSON data
func createTimeBasedValue(jsonStr string) (*pb.TimeBasedValue, error) {
	anyValue, err := schema.JSONToAny(jsonStr)
	if err != nil {
		return nil, err
	}

	return &pb.TimeBasedValue{
		StartTime: time.Now().Format(time.RFC3339),
		EndTime:   time.Now().Add(24 * time.Hour).Format(time.RFC3339),
		Value:     anyValue,
	}, nil
}

// createEntityWithAttributes creates an Entity with the given attributes
func createEntityWithAttributes(entityID string, attributes map[string]string) (*pb.Entity, error) {
	entity := &pb.Entity{
		Id: entityID,
		Kind: &pb.Kind{
			Major: "test",
			Minor: "v1",
		},
		Created:    time.Now().Format(time.RFC3339),
		Attributes: make(map[string]*pb.TimeBasedValueList),
	}

	for attrName, jsonStr := range attributes {
		timeBasedValue, err := createTimeBasedValue(jsonStr)
		if err != nil {
			return nil, fmt.Errorf("failed to create TimeBasedValue for %s: %v", attrName, err)
		}

		entity.Attributes[attrName] = &pb.TimeBasedValueList{
			Values: []*pb.TimeBasedValue{timeBasedValue},
		}
	}

	return entity, nil
}

// TestEntityWithGraphDataOnly tests an entity containing only graph data
func TestEntityWithGraphDataOnly(t *testing.T) {
	graphData := `{
		"nodes": [
			{"id": "user1", "type": "user", "properties": {"name": "Alice", "age": 30}},
			{"id": "user2", "type": "user", "properties": {"name": "Bob", "age": 25}},
			{"id": "post1", "type": "post", "properties": {"title": "Hello", "content": "World"}}
		],
		"edges": [
			{"source": "user1", "target": "user2", "type": "follows", "properties": {"since": "2024-01-01"}},
			{"source": "user1", "target": "post1", "type": "created", "properties": {"timestamp": "2024-03-20T10:00:00Z"}}
		]
	}`

	entity, err := createEntityWithAttributes("graph-entity-1", map[string]string{
		"social_network": graphData,
	})
	assert.NoError(t, err)

	processor := NewEntityAttributeProcessor()
	ctx := context.Background()

	// Test all CRUD operations
	operations := []string{"create", "read", "update", "delete"}
	for _, operation := range operations {
		t.Run(operation, func(t *testing.T) {
			err := processor.ProcessEntityAttributes(ctx, entity, operation)
			assert.NoError(t, err)
		})
	}
}

// TestEntityWithTabularDataOnly tests an entity containing only tabular data
func TestEntityWithTabularDataOnly(t *testing.T) {
	tabularData := `{
		"columns": ["id", "name", "age", "department"],
		"rows": [
			[1, "John Doe", 30, "Engineering"],
			[2, "Jane Smith", 25, "Marketing"],
			[3, "Bob Wilson", 35, "Sales"]
		]
	}`

	entity, err := createEntityWithAttributes("tabular-entity-1", map[string]string{
		"employees": tabularData,
	})
	assert.NoError(t, err)

	processor := NewEntityAttributeProcessor()
	ctx := context.Background()

	// Test all CRUD operations
	operations := []string{"create", "read", "update", "delete"}
	for _, operation := range operations {
		t.Run(operation, func(t *testing.T) {
			err := processor.ProcessEntityAttributes(ctx, entity, operation)
			assert.NoError(t, err)
		})
	}
}

// TestEntityWithDocumentDataOnly tests an entity containing only document data
func TestEntityWithDocumentDataOnly(t *testing.T) {
	documentData := `{
		"user_profile": {
			"name": "John Doe",
			"email": "john@example.com",
			"age": 30,
			"active": true,
			"preferences": {
				"theme": "dark",
				"notifications": true,
				"language": "en"
			},
			"address": {
				"street": "123 Main St",
				"city": "New York",
				"zip": "10001"
			}
		}
	}`

	entity, err := createEntityWithAttributes("document-entity-1", map[string]string{
		"profile": documentData,
	})
	assert.NoError(t, err)

	processor := NewEntityAttributeProcessor()
	ctx := context.Background()

	// Test all CRUD operations
	operations := []string{"create", "read", "update", "delete"}
	for _, operation := range operations {
		t.Run(operation, func(t *testing.T) {
			err := processor.ProcessEntityAttributes(ctx, entity, operation)
			assert.NoError(t, err)
		})
	}
}

// TestEntityWithMixedDataTypes tests an entity containing mixed data types
func TestEntityWithMixedDataTypes(t *testing.T) {
	graphData := `{
		"nodes": [
			{"id": "user1", "type": "user", "properties": {"name": "Alice"}},
			{"id": "user2", "type": "user", "properties": {"name": "Bob"}}
		],
		"edges": [
			{"source": "user1", "target": "user2", "type": "follows"}
		]
	}`

	tabularData := `{
		"columns": ["id", "name", "score"],
		"rows": [
			[1, "Alice", 95.5],
			[2, "Bob", 88.0]
		]
	}`

	documentData := `{
		"settings": {
			"theme": "dark",
			"notifications": true,
			"language": "en"
		}
	}`

	entity, err := createEntityWithAttributes("mixed-entity-1", map[string]string{
		"social_graph":     graphData,
		"performance_data": tabularData,
		"user_settings":    documentData,
	})
	assert.NoError(t, err)

	processor := NewEntityAttributeProcessor()
	ctx := context.Background()

	// Test all CRUD operations
	operations := []string{"create", "read", "update", "delete"}
	for _, operation := range operations {
		t.Run(operation, func(t *testing.T) {
			err := processor.ProcessEntityAttributes(ctx, entity, operation)
			assert.NoError(t, err)
		})
	}
}

// TestComplexGraphEntity tests a complex graph entity with multiple node and edge types
func TestComplexGraphEntity(t *testing.T) {
	complexGraphData := `{
		"nodes": [
			{"id": "user1", "type": "user", "properties": {"name": "Alice", "age": 30, "location": "NY"}},
			{"id": "user2", "type": "user", "properties": {"name": "Bob", "age": 25, "location": "SF"}},
			{"id": "post1", "type": "post", "properties": {"title": "Hello", "content": "World", "created": "2024-03-20"}},
			{"id": "post2", "type": "post", "properties": {"title": "Graph", "content": "DB", "created": "2024-03-21"}},
			{"id": "tag1", "type": "tag", "properties": {"name": "technology"}},
			{"id": "tag2", "type": "tag", "properties": {"name": "database"}}
		],
		"edges": [
			{"source": "user1", "target": "user2", "type": "follows", "properties": {"since": "2024-01-01"}},
			{"source": "user1", "target": "post1", "type": "created", "properties": {"timestamp": "2024-03-20T10:00:00Z"}},
			{"source": "user2", "target": "post1", "type": "likes", "properties": {"timestamp": "2024-03-20T11:00:00Z"}},
			{"source": "user2", "target": "post2", "type": "created", "properties": {"timestamp": "2024-03-21T09:00:00Z"}},
			{"source": "post1", "target": "tag1", "type": "tagged_with", "properties": {"confidence": 0.9}},
			{"source": "post2", "target": "tag1", "type": "tagged_with", "properties": {"confidence": 0.8}},
			{"source": "post2", "target": "tag2", "type": "tagged_with", "properties": {"confidence": 0.95}}
		]
	}`

	entity, err := createEntityWithAttributes("complex-graph-entity-1", map[string]string{
		"social_network": complexGraphData,
	})
	assert.NoError(t, err)

	processor := NewEntityAttributeProcessor()
	ctx := context.Background()

	// Test all CRUD operations
	operations := []string{"create", "read", "update", "delete"}
	for _, operation := range operations {
		t.Run(operation, func(t *testing.T) {
			err := processor.ProcessEntityAttributes(ctx, entity, operation)
			assert.NoError(t, err)
		})
	}
}

// TestComplexTabularEntity tests a complex tabular entity with various data types
func TestComplexTabularEntity(t *testing.T) {
	complexTabularData := `{
		"columns": ["id", "name", "age", "salary", "department", "is_active", "hire_date", "last_login"],
		"rows": [
			[1, "John Doe", 30, 75000.50, "Engineering", true, "2020-01-15", "2024-03-20T09:00:00Z"],
			[2, "Jane Smith", 25, 65000.00, "Marketing", true, "2021-03-10", "2024-03-20T08:30:00Z"],
			[3, "Bob Wilson", 35, 85000.75, "Sales", false, "2019-11-20", "2024-03-19T17:00:00Z"],
			[4, "Alice Brown", 28, 70000.25, "Engineering", true, "2022-06-05", "2024-03-20T10:15:00Z"],
			[5, "Charlie Davis", 32, 80000.00, "Finance", true, "2020-08-12", "2024-03-20T07:45:00Z"]
		]
	}`

	entity, err := createEntityWithAttributes("complex-tabular-entity-1", map[string]string{
		"employee_data": complexTabularData,
	})
	assert.NoError(t, err)

	processor := NewEntityAttributeProcessor()
	ctx := context.Background()

	// Test all CRUD operations
	operations := []string{"create", "read", "update", "delete"}
	for _, operation := range operations {
		t.Run(operation, func(t *testing.T) {
			err := processor.ProcessEntityAttributes(ctx, entity, operation)
			assert.NoError(t, err)
		})
	}
}

// TestComplexDocumentEntity tests a complex document entity with nested structures
func TestComplexDocumentEntity(t *testing.T) {
	complexDocumentData := `{
		"user_profile": {
			"personal_info": {
				"name": "John Doe",
				"email": "john@example.com",
				"phone": "+1-555-123-4567",
				"age": 30,
				"birth_date": "1994-05-15",
				"gender": "male"
			},
			"address": {
				"street": "123 Main Street",
				"city": "New York",
				"state": "NY",
				"zip_code": "10001",
				"country": "USA"
			},
			"preferences": {
				"theme": "dark",
				"language": "en",
				"timezone": "America/New_York",
				"notifications": {
					"email": true,
					"sms": false,
					"push": true,
					"frequency": "daily"
				}
			},
			"security": {
				"two_factor_enabled": true,
				"last_password_change": "2024-01-15T10:30:00Z",
				"login_history": [
					{"timestamp": "2024-03-20T09:00:00Z", "ip": "192.168.1.100", "device": "Chrome/Windows"},
					{"timestamp": "2024-03-19T17:30:00Z", "ip": "192.168.1.100", "device": "Chrome/Windows"}
				]
			}
		}
	}`

	entity, err := createEntityWithAttributes("complex-document-entity-1", map[string]string{
		"profile": complexDocumentData,
	})
	assert.NoError(t, err)

	processor := NewEntityAttributeProcessor()
	ctx := context.Background()

	// Test all CRUD operations
	operations := []string{"create", "read", "update", "delete"}
	for _, operation := range operations {
		t.Run(operation, func(t *testing.T) {
			err := processor.ProcessEntityAttributes(ctx, entity, operation)
			assert.NoError(t, err)
		})
	}
}

// TestEntityWithMultipleAttributesOfSameType tests an entity with multiple attributes of the same type
func TestEntityWithMultipleAttributesOfSameType(t *testing.T) {
	graphData1 := `{
		"nodes": [{"id": "user1", "type": "user", "properties": {"name": "Alice"}}],
		"edges": []
	}`

	graphData2 := `{
		"nodes": [{"id": "user2", "type": "user", "properties": {"name": "Bob"}}],
		"edges": []
	}`

	tabularData1 := `{
		"columns": ["id", "name"],
		"rows": [[1, "John"]]
	}`

	tabularData2 := `{
		"columns": ["id", "score"],
		"rows": [[1, 95.5]]
	}`

	documentData1 := `{
		"settings": {"theme": "dark"}
	}`

	documentData2 := `{
		"metadata": {"version": "1.0"}
	}`

	entity, err := createEntityWithAttributes("multi-attr-entity-1", map[string]string{
		"friends_graph":    graphData1,
		"family_graph":     graphData2,
		"personal_data":    tabularData1,
		"performance_data": tabularData2,
		"user_settings":    documentData1,
		"system_metadata":  documentData2,
	})
	assert.NoError(t, err)

	processor := NewEntityAttributeProcessor()
	ctx := context.Background()

	// Test all CRUD operations
	operations := []string{"create", "read", "update", "delete"}
	for _, operation := range operations {
		t.Run(operation, func(t *testing.T) {
			err := processor.ProcessEntityAttributes(ctx, entity, operation)
			assert.NoError(t, err)
		})
	}
}

// TestStorageTypeDetection tests that storage types are correctly detected
func TestStorageTypeDetection(t *testing.T) {
	testCases := map[string]struct {
		jsonData string
		expected storageinference.StorageType
	}{
		"graph_data": {
			jsonData: `{
				"nodes": [{"id": "user1", "type": "user"}],
				"edges": [{"source": "user1", "target": "user2"}]
			}`,
			expected: storageinference.GraphData,
		},
		"tabular_data": {
			jsonData: `{
				"columns": ["id", "name"],
				"rows": [[1, "John"]]
			}`,
			expected: storageinference.TabularData,
		},
		"document_data": {
			jsonData: `{
				"key1": "value1",
				"key2": "value2"
			}`,
			expected: storageinference.MapData,
		},
		"scalar_data": {
			jsonData: `42`,
			expected: storageinference.ScalarData,
		},
		"list_data": {
			jsonData: `[1, 2, 3, 4, 5]`,
			expected: storageinference.ListData,
		},
	}

	processor := NewEntityAttributeProcessor()

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			anyValue, err := schema.JSONToAny(testCase.jsonData)
			assert.NoError(t, err)

			detectedType, err := processor.determineStorageType(anyValue)
			assert.NoError(t, err)
			assert.Equal(t, testCase.expected, detectedType)
		})
	}
}

// TestEmptyEntity tests an entity with no attributes
func TestEmptyEntity(t *testing.T) {
	entity := &pb.Entity{
		Id: "empty-entity-1",
		Kind: &pb.Kind{
			Major: "test",
			Minor: "v1",
		},
		Created:    time.Now().Format(time.RFC3339),
		Attributes: make(map[string]*pb.TimeBasedValueList),
	}

	processor := NewEntityAttributeProcessor()
	ctx := context.Background()

	// Test all CRUD operations
	operations := []string{"create", "read", "update", "delete"}
	for _, operation := range operations {
		t.Run(operation, func(t *testing.T) {
			err := processor.ProcessEntityAttributes(ctx, entity, operation)
			assert.NoError(t, err)
		})
	}
}

// TestNilEntity tests handling of nil entity
func TestNilEntity(t *testing.T) {
	processor := NewEntityAttributeProcessor()
	ctx := context.Background()

	// Test all CRUD operations
	operations := []string{"create", "read", "update", "delete"}
	for _, operation := range operations {
		t.Run(operation, func(t *testing.T) {
			err := processor.ProcessEntityAttributes(ctx, nil, operation)
			assert.NoError(t, err) // Should handle nil gracefully
		})
	}
}

// TestInvalidOperation tests handling of invalid operation
func TestInvalidOperation(t *testing.T) {
	entity, err := createEntityWithAttributes("test-entity-1", map[string]string{
		"test_data": `{
			"key1": "value1",
			"key2": "value2"
		}`,
	})
	assert.NoError(t, err)

	processor := NewEntityAttributeProcessor()
	ctx := context.Background()

	err = processor.ProcessEntityAttributes(ctx, entity, "invalid_operation")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown operation")
}

// TestUnsupportedStorageType tests handling of unsupported storage types
func TestUnsupportedStorageType(t *testing.T) {
	entity, err := createEntityWithAttributes("test-entity-2", map[string]string{
		"scalar_data": `42`,
	})
	assert.NoError(t, err)

	processor := NewEntityAttributeProcessor()
	ctx := context.Background()

	// Should not error, but should log a warning and skip the attribute
	err = processor.ProcessEntityAttributes(ctx, entity, "create")
	assert.NoError(t, err) // Should handle gracefully
}

// TestBasicFunctionality tests basic functionality of the attribute resolver
func TestBasicFunctionality(t *testing.T) {
	// Test that we can create a processor
	processor := NewEntityAttributeProcessor()
	assert.NotNil(t, processor)
	assert.NotNil(t, processor.resolvers)

	// Test that we have the expected resolvers
	assert.NotNil(t, processor.resolvers[storageinference.GraphData])
	assert.NotNil(t, processor.resolvers[storageinference.TabularData])
	assert.NotNil(t, processor.resolvers[storageinference.MapData])

	// Test with a simple document entity
	entity, err := createEntityWithAttributes("test-entity", map[string]string{
		"simple_data": `{"key": "value"}`,
	})
	assert.NoError(t, err)
	assert.NotNil(t, entity)

	// Test processing
	ctx := context.Background()
	err = processor.ProcessEntityAttributes(ctx, entity, "create")
	assert.NoError(t, err)
}
