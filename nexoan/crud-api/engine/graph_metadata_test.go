package engine

import (
	"context"
	"testing"
	"time"

	"lk/datafoundation/crud-api/pkg/storageinference"

	"github.com/stretchr/testify/assert"
)

// TestGraphMetadataManager tests the graph metadata manager functionality
func TestGraphMetadataManager(t *testing.T) {
	manager := NewGraphMetadataManager()
	assert.NotNil(t, manager)

	ctx := context.Background()

	// Test creating attribute metadata
	metadata := &AttributeMetadata{
		EntityID:      "test-entity-1",
		AttributeName: "test-attribute",
		StorageType:   storageinference.TabularData,
		StoragePath:   "tables/attr_test-entity-1_test-attribute",
		Created:       time.Now(),
		Updated:       time.Now(),
		Schema: map[string]interface{}{
			"columns": []string{"id", "name"},
			"types":   []string{"int", "string"},
		},
	}

	// Test creating attribute node
	err := manager.CreateAttributeNode(ctx, metadata)
	assert.NoError(t, err)

	// Test creating IS_ATTRIBUTE relationship
	attributeID := GenerateAttributeID(metadata.EntityID, metadata.AttributeName)
	err = manager.CreateIS_ATTRIBUTE_Relationship(ctx, metadata.EntityID, attributeID)
	assert.NoError(t, err)

	// Test getting attribute metadata
	retrievedMetadata, err := manager.GetAttributeMetadata(ctx, metadata.EntityID, metadata.AttributeName)
	assert.NoError(t, err)
	assert.NotNil(t, retrievedMetadata)
	assert.Equal(t, metadata.EntityID, retrievedMetadata.EntityID)
	assert.Equal(t, metadata.AttributeName, retrievedMetadata.AttributeName)

	// Test updating attribute metadata
	metadata.Updated = time.Now()
	metadata.Schema["new_field"] = "new_value"
	err = manager.UpdateAttributeMetadata(ctx, metadata)
	assert.NoError(t, err)

	// Test listing entity attributes
	attributes, err := manager.ListEntityAttributes(ctx, metadata.EntityID)
	assert.NoError(t, err)
	assert.NotNil(t, attributes)

	// Test deleting attribute node
	err = manager.DeleteAttributeNode(ctx, metadata.EntityID, metadata.AttributeName)
	assert.NoError(t, err)
}

// TestDatasetTypeMapping tests the mapping between storage types and dataset types
func TestDatasetTypeMapping(t *testing.T) {
	testCases := map[storageinference.StorageType]string{
		storageinference.TabularData: TabularDataset,
		storageinference.GraphData:   GraphDataset,
		storageinference.MapData:     DocumentDataset,
		storageinference.ListData:    BlobDataset,
		storageinference.ScalarData:  BlobDataset,
		storageinference.UnknownData: BlobDataset,
	}

	for storageType, expectedDatasetType := range testCases {
		t.Run(string(storageType), func(t *testing.T) {
			datasetType := GetDatasetType(storageType)
			assert.Equal(t, expectedDatasetType, datasetType)
		})
	}
}

// TestAttributeIDGeneration tests the attribute ID generation
func TestAttributeIDGeneration(t *testing.T) {
	entityID := "test-entity-123"
	attributeName := "user_profile"

	attributeID := GenerateAttributeID(entityID, attributeName)
	expectedID := "test-entity-123_attr_user_profile"
	assert.Equal(t, expectedID, attributeID)
}

// TestStoragePathGeneration tests the storage path generation
func TestStoragePathGeneration(t *testing.T) {
	entityID := "test-entity-123"
	attributeName := "user_profile"

	testCases := map[storageinference.StorageType]string{
		storageinference.TabularData: "tables/attr_test-entity-123_user_profile",
		storageinference.GraphData:   "graphs/attr_test-entity-123_user_profile",
		storageinference.MapData:     "documents/attr_test-entity-123_user_profile",
		storageinference.ListData:    "blobs/attr_test-entity-123_user_profile",
		storageinference.ScalarData:  "blobs/attr_test-entity-123_user_profile",
		storageinference.UnknownData: "unknown/attr_test-entity-123_user_profile",
	}

	for storageType, expectedPath := range testCases {
		t.Run(string(storageType), func(t *testing.T) {
			path := GenerateStoragePath(entityID, attributeName, storageType)
			assert.Equal(t, expectedPath, path)
		})
	}
}

// TestGraphMetadataIntegration tests the integration of graph metadata with attribute processing
func TestGraphMetadataIntegration(t *testing.T) {
	// Create an entity with mixed data types
	entity, err := createEntityWithAttributes("integration-test-entity", map[string]string{
		"tabular_data": `{
			"columns": ["id", "name"],
			"rows": [[1, "John"], [2, "Jane"]]
		}`,
		"graph_data": `{
			"nodes": [{"id": "user1", "type": "user"}],
			"edges": [{"source": "user1", "target": "user2"}]
		}`,
		"document_data": `{
			"user_profile": {"name": "John", "age": 30}
		}`,
	})
	assert.NoError(t, err)

	processor := NewEntityAttributeProcessor()
	ctx := context.Background()

	// Test create operation - this should create graph metadata
	err = processor.ProcessEntityAttributes(ctx, entity, "create")
	assert.NoError(t, err)

	// Test read operation - this should verify graph metadata
	err = processor.ProcessEntityAttributes(ctx, entity, "read")
	assert.NoError(t, err)

	// Test update operation - this should update graph metadata
	err = processor.ProcessEntityAttributes(ctx, entity, "update")
	assert.NoError(t, err)

	// Test delete operation - this should delete graph metadata
	err = processor.ProcessEntityAttributes(ctx, entity, "delete")
	assert.NoError(t, err)
}

// TestDataDiscoveryService tests the data discovery service functionality
// TODO: Re-enable this test when the complete data discovery functionality is implemented
// Currently disabled due to placeholder implementations returning nil values
/*
func TestDataDiscoveryService(t *testing.T) {
	service := NewDataDiscoveryService()
	assert.NotNil(t, service)

	ctx := context.Background()

	// Test discovering entity attributes
	locations, err := service.DiscoverEntityAttributes(ctx, "test-entity-1")
	assert.NoError(t, err)
	assert.NotNil(t, locations) // Should be non-nil even if empty

	// Test finding attributes by type
	tabularAttributes, err := service.FindAttributeByType(ctx, storageinference.TabularData)
	assert.NoError(t, err)
	assert.NotNil(t, tabularAttributes) // Should be non-nil even if empty

	// Test finding attributes by name
	userProfileAttributes, err := service.FindAttributeByName(ctx, "user_profile")
	assert.NoError(t, err)
	assert.NotNil(t, userProfileAttributes) // Should be non-nil even if empty

	// Test getting attribute location
	location, err := service.GetAttributeLocation(ctx, "test-entity-1", "test-attribute")
	assert.NoError(t, err)
	assert.NotNil(t, location)

	// Test search attributes
	criteria := &AttributeSearchCriteria{
		StorageType: storageinference.TabularData,
		DatasetType: TabularDataset,
	}
	searchResults, err := service.SearchAttributes(ctx, criteria)
	assert.NoError(t, err)
	assert.NotNil(t, searchResults) // Should be non-nil even if empty

	// Test generating discovery report
	report, err := service.GenerateDiscoveryReport(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, report)

	// Test getting storage resolver
	resolver, err := service.GetStorageResolver(storageinference.TabularData)
	assert.NoError(t, err)
	assert.NotNil(t, resolver)

	// Test validating attribute location
	valid, err := service.ValidateAttributeLocation(ctx, location)
	assert.NoError(t, err)
	assert.True(t, valid)

	// Test getting attribute schema
	schema, err := service.GetAttributeSchema(ctx, "test-entity-1", "test-attribute")
	assert.NoError(t, err)
	assert.NotNil(t, schema)

	// Test listing entities with attributes
	entities, err := service.ListEntitiesWithAttributes(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, entities) // Should be non-nil even if empty
}
*/

// TestAttributeLocationStructure tests the AttributeLocation structure
func TestAttributeLocationStructure(t *testing.T) {
	location := &AttributeLocation{
		EntityID:      "test-entity-1",
		AttributeName: "user_profile",
		StorageType:   storageinference.TabularData,
		StoragePath:   "tables/attr_test-entity-1_user_profile",
		DatasetType:   TabularDataset,
		Created:       time.Now().Format("2006-01-02T15:04:05Z"),
		Updated:       time.Now().Format("2006-01-02T15:04:05Z"),
	}

	assert.Equal(t, "test-entity-1", location.EntityID)
	assert.Equal(t, "user_profile", location.AttributeName)
	assert.Equal(t, storageinference.TabularData, location.StorageType)
	assert.Equal(t, "tables/attr_test-entity-1_user_profile", location.StoragePath)
	assert.Equal(t, TabularDataset, location.DatasetType)
	assert.NotEmpty(t, location.Created)
	assert.NotEmpty(t, location.Updated)
}

// TestAttributeSearchCriteria tests the search criteria structure
func TestAttributeSearchCriteria(t *testing.T) {
	criteria := &AttributeSearchCriteria{
		StorageType:          storageinference.TabularData,
		DatasetType:          TabularDataset,
		EntityIDPattern:      "test-*",
		AttributeNamePattern: "user_*",
		CreatedAfter:         "2024-01-01T00:00:00Z",
		CreatedBefore:        "2024-12-31T23:59:59Z",
		UpdatedAfter:         "2024-01-01T00:00:00Z",
		UpdatedBefore:        "2024-12-31T23:59:59Z",
		SchemaProperties: map[string]interface{}{
			"has_columns":  true,
			"column_count": 5,
		},
	}

	assert.Equal(t, storageinference.TabularData, criteria.StorageType)
	assert.Equal(t, TabularDataset, criteria.DatasetType)
	assert.Equal(t, "test-*", criteria.EntityIDPattern)
	assert.Equal(t, "user_*", criteria.AttributeNamePattern)
	assert.Equal(t, "2024-01-01T00:00:00Z", criteria.CreatedAfter)
	assert.Equal(t, "2024-12-31T23:59:59Z", criteria.CreatedBefore)
	assert.Equal(t, "2024-01-01T00:00:00Z", criteria.UpdatedAfter)
	assert.Equal(t, "2024-12-31T23:59:59Z", criteria.UpdatedBefore)
	assert.Equal(t, true, criteria.SchemaProperties["has_columns"])
	assert.Equal(t, 5, criteria.SchemaProperties["column_count"])
}

// TestDiscoveryReportStructure tests the DiscoveryReport structure
func TestDiscoveryReportStructure(t *testing.T) {
	report := &DiscoveryReport{
		TotalAttributes: 100,
		ByStorageType: map[storageinference.StorageType]int{
			storageinference.TabularData: 50,
			storageinference.GraphData:   20,
			storageinference.MapData:     30,
		},
		ByDatasetType: map[string]int{
			TabularDataset:  50,
			GraphDataset:    20,
			DocumentDataset: 30,
		},
		RecentAttributes: []*AttributeLocation{
			{
				EntityID:      "entity-1",
				AttributeName: "attr-1",
				StorageType:   storageinference.TabularData,
				DatasetType:   TabularDataset,
			},
		},
		StorageBreakdown: map[string]interface{}{
			"total_size":     "1.5GB",
			"table_count":    50,
			"graph_count":    20,
			"document_count": 30,
		},
	}

	assert.Equal(t, 100, report.TotalAttributes)
	assert.Equal(t, 50, report.ByStorageType[storageinference.TabularData])
	assert.Equal(t, 20, report.ByStorageType[storageinference.GraphData])
	assert.Equal(t, 30, report.ByStorageType[storageinference.MapData])
	assert.Equal(t, 50, report.ByDatasetType[TabularDataset])
	assert.Equal(t, 20, report.ByDatasetType[GraphDataset])
	assert.Equal(t, 30, report.ByDatasetType[DocumentDataset])
	assert.Len(t, report.RecentAttributes, 1)
	assert.Equal(t, "entity-1", report.RecentAttributes[0].EntityID)
	assert.Equal(t, "1.5GB", report.StorageBreakdown["total_size"])
	assert.Equal(t, 50, report.StorageBreakdown["table_count"])
}
