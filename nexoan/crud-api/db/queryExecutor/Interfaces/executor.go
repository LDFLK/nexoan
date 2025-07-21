package interfaces

import (
	"context"
	"lk/datafoundation/crud-api/pkg/schema"
	pb "lk/datafoundation/crud-api/lk/datafoundation/crud-api"
	"google.golang.org/protobuf/types/known/anypb"
)

// Executor is the main parent interface that defines common behavior for all database executors.
// It provides a unified way to execute queries across different database types.
//
// Example Usage:
//
//	executor := GetExecutor()
//	result, err := executor.Execute(ctx, &QueryParameter{
//		Operation: "get_metadata",
//		Filters: map[string]interface{}{
//			"entityId": "123",
//		},
//	})
type Executor interface {
	// Execute processes a query based on the provided QueryParameter.
	// It routes the query to the appropriate database implementation.
	//
	// Parameters:
	//   - ctx: Context for the operation
	//   - plan: QueryParameter containing operation details
	//
	// Returns:
	//   - interface{}: Result of the operation, type depends on the operation
	//   - error: Any error that occurred during execution
	//
	// Example:
	//   result, err := executor.Execute(ctx, &QueryParameter{
	//       Operation: "get_metadata",
	//       Filters: map[string]interface{}{
	//           "entityId": "123",
	//       },
	//   })
	Execute(ctx context.Context, plan *QueryParameter) (interface{}, error)
}

// QueryParameter defines the structure for query execution parameters.
// It provides a standardized way to specify operation details across different database types.
//
// Example:
//
//	param := &QueryParameter{
//		Operation: "get_metadata",
//		Filters: map[string]interface{}{
//			"entityId": "123",
//			"kind": "Person",
//		TimeRange: struct{
//			StartTime: "2024-01-01T00:00:00Z",
//			EndTime: "2024-12-31T23:59:59Z",
//		},
//	}
type QueryParameter struct {
	// Operation specifies the type of operation to perform
	// Valid values include:
	// - "get_metadata": Retrieve entity metadata
	// - "create_entity": Create a new entity
	// - "get_relationships": Get entity relationships
	Operation string

	// Filters specify criteria for the operation
	// Common filter keys:
	// - "entityId": ID of the entity
	// - "kind": Entity kind/type
	// - "name": Entity name
	Filters map[string]interface{}

	// TimeRange specifies temporal constraints for the operation
	TimeRange struct {
		// StartTime is the ISO8601 timestamp for the start of the time range
		StartTime string

		// EndTime is the ISO8601 timestamp for the end of the time range
		EndTime string

		// ActiveAt is the ISO8601 timestamp to check if entity/relationship was active
		ActiveAt string
	}
}

// DocumentExecutor interface defines operations for document-based databases (e.g., MongoDB).
// It handles metadata operations for entities.
//
// Example Usage:
//
//	executor := GetDocumentExecutor()
//	metadata, err := executor.GetMetadata(ctx, &QueryParameter{
//		Filters: map[string]interface{}{
//			"entityId": "123",
//		},
//	})
type DocumentExecutor interface {
	Executor

	// HandleMetadata manages entity metadata operations.
	//
	// Parameters:
	//   - ctx: Context for the operation
	//   - plan: QueryParameter containing metadata operation details
	//
	// Returns:
	//   - error: Any error that occurred during the operation
	//
	// Example:
	//   err := executor.HandleMetadata(ctx, &QueryParameter{
	//       Operation: "update_metadata",
	//       Filters: map[string]interface{}{
	//           "entityId": "123",
	//           "metadata": map[string]interface{}{
	//               "department": "Engineering",
	//           },
	//       },
	//   })
	HandleMetadata(ctx context.Context, plan *QueryParameter) error

	// GetMetadata retrieves metadata for an entity.
	//
	// Parameters:
	//   - ctx: Context for the operation
	//   - plan: QueryParameter containing entity identifier
	//
	// Returns:
	//   - map[string]*anypb.Any: Map of metadata key-value pairs
	//   - error: Any error that occurred during retrieval
	//
	// Example:
	//   metadata, err := executor.GetMetadata(ctx, &QueryParameter{
	//       Filters: map[string]interface{}{
	//           "entityId": "123",
	//       },
	//   })
	GetMetadata(ctx context.Context, plan *QueryParameter) (map[string]*anypb.Any, error)
}

// GraphExecutor interface defines operations for graph databases (e.g., Neo4j).
// It handles entity and relationship operations in a graph structure.
//
// Example Usage:
//
//	executor := GetGraphExecutor()
//	entity, err := executor.GraphEntity(ctx, &QueryParameter{
//		Filters: map[string]interface{}{
//			"entityId": "123",
//		},
//	})
type GraphExecutor interface {
	Executor

	// GraphEntity retrieves an entity's graph information.
	//
	// Parameters:
	//   - ctx: Context for the operation
	//   - plan: QueryParameter containing entity identifier
	//
	// Returns:
	//   - *pb.Kind: Entity kind information
	//   - *pb.TimeBasedValue: Entity name with temporal information
	//   - string: Creation timestamp
	//   - string: Termination timestamp
	//   - error: Any error that occurred during retrieval
	//
	// Example:
	//   kind, name, created, terminated, err := executor.GraphEntity(ctx, &QueryParameter{
	//       Filters: map[string]interface{}{
	//           "entityId": "123",
	//       },
	//   })
	GraphEntity(ctx context.Context, plan *QueryParameter) (*pb.Kind, *pb.TimeBasedValue, string, string, error)

	// GraphEntityCreate creates a new entity in the graph.
	//
	// Parameters:
	//   - ctx: Context for the operation
	//   - plan: QueryParameter containing entity details
	//
	// Returns:
	//   - bool: True if creation was successful
	//   - error: Any error that occurred during creation
	//
	// Example:
	//   success, err := executor.GraphEntityCreate(ctx, &QueryParameter{
	//       Operation: "create_entity",
	//       Filters: map[string]interface{}{
	//           "entityId": "123",
	//           "kind": map[string]string{
	//               "major": "Person",
	//               "minor": "Employee",
	//           },
	//           "name": "John Doe",
	//       },
	//   })
	GraphEntityCreate(ctx context.Context, plan *QueryParameter) (bool, error)

	// GraphEntityUpdate updates an existing entity in the graph.
	//
	// Parameters:
	//   - ctx: Context for the operation
	//   - plan: QueryParameter containing update details
	//
	// Returns:
	//   - bool: True if update was successful
	//   - error: Any error that occurred during update
	//
	// Example:
	//   success, err := executor.GraphEntityUpdate(ctx, &QueryParameter{
	//       Operation: "update_entity",
	//       Filters: map[string]interface{}{
	//           "entityId": "123",
	//           "name": "Jane Doe",
	//       },
	//   })
	GraphEntityUpdate(ctx context.Context, plan *QueryParameter) (bool, error)

	// GraphEntityFilter filters entities based on criteria.
	//
	// Parameters:
	//   - ctx: Context for the operation
	//   - plan: QueryParameter containing filter criteria
	//
	// Returns:
	//   - []map[string]interface{}: List of matching entities
	//   - error: Any error that occurred during filtering
	//
	// Example:
	//   entities, err := executor.GraphEntityFilter(ctx, &QueryParameter{
	//       Filters: map[string]interface{}{
	//           "kind": "Person",
	//           "department": "Engineering",
	//       },
	//   })
	GraphEntityFilter(ctx context.Context, plan *QueryParameter) ([]map[string]interface{}, error)

	// GraphRelationships retrieves all relationships for an entity.
	//
	// Parameters:
	//   - ctx: Context for the operation
	//   - plan: QueryParameter containing entity identifier
	//
	// Returns:
	//   - map[string]*pb.Relationship: Map of relationships keyed by relationship ID
	//   - error: Any error that occurred during retrieval
	//
	// Example:
	//   relationships, err := executor.GraphRelationships(ctx, &QueryParameter{
	//       Filters: map[string]interface{}{
	//           "entityId": "123",
	//       },
	//   })
	GraphRelationships(ctx context.Context, plan *QueryParameter) (map[string]*pb.Relationship, error)

	// FilteredRelationships retrieves relationships based on filter criteria.
	//
	// Parameters:
	//   - ctx: Context for the operation
	//   - plan: QueryParameter containing filter criteria
	//
	// Returns:
	//   - map[string]*pb.Relationship: Map of filtered relationships
	//   - error: Any error that occurred during filtering
	//
	// Example:
	//   relationships, err := executor.FilteredRelationships(ctx, &QueryParameter{
	//       Filters: map[string]interface{}{
	//           "entityId": "123",
	//           "relationshipType": "MANAGES",
	//           "direction": "OUTGOING",
	//       },
	//       TimeRange: struct{
	//           ActiveAt: "2024-03-21T00:00:00Z",
	//       },
	//   })
	FilteredRelationships(ctx context.Context, plan *QueryParameter) (map[string]*pb.Relationship, error)

	// RelationshipsCreate creates new relationships for an entity.
	//
	// Parameters:
	//   - ctx: Context for the operation
	//   - plan: QueryParameter containing relationship details
	//
	// Returns:
	//   - error: Any error that occurred during creation
	//
	// Example:
	//   err := executor.RelationshipsCreate(ctx, &QueryParameter{
	//       Operation: "create_relationship",
	//       Filters: map[string]interface{}{
	//           "entityId": "123",
	//           "relatedEntityId": "456",
	//           "relationshipType": "MANAGES",
	//           "startTime": "2024-01-01T00:00:00Z",
	//       },
	//   })
	RelationshipsCreate(ctx context.Context, plan *QueryParameter) error

	// RelationshipsUpdate updates existing relationships.
	//
	// Parameters:
	//   - ctx: Context for the operation
	//   - plan: QueryParameter containing update details
	//
	// Returns:
	//   - error: Any error that occurred during update
	//
	// Example:
	//   err := executor.RelationshipsUpdate(ctx, &QueryParameter{
	//       Operation: "update_relationship",
	//       Filters: map[string]interface{}{
	//           "relationshipId": "rel123",
	//           "endTime": "2024-03-21T00:00:00Z",
	//       },
	//   })
	RelationshipsUpdate(ctx context.Context, plan *QueryParameter) error
}

// TabularExecutor interface defines operations for relational databases (e.g., PostgreSQL).
// It handles table management and data operations in a tabular structure.
//
// Example Usage:
//
//	executor := GetTabularExecutor()
//	exists, err := executor.TableExists(ctx, &QueryParameter{
//		Filters: map[string]interface{}{
//			"tableName": "employees",
//		},
//	})
type TabularExecutor interface {
	Executor

	// TableExists checks if a table exists in the database.
	//
	// Parameters:
	//   - ctx: Context for the operation
	//   - plan: QueryParameter containing table name
	//
	// Returns:
	//   - bool: True if table exists
	//   - error: Any error that occurred during check
	//
	// Example:
	//   exists, err := executor.TableExists(ctx, &QueryParameter{
	//       Filters: map[string]interface{}{
	//           "tableName": "employees",
	//       },
	//   })
	TableExists(ctx context.Context, plan *QueryParameter) (bool, error)

	// CreateLookupTable creates a new lookup table.
	//
	// Parameters:
	//   - ctx: Context for the operation
	//   - plan: QueryParameter containing table schema
	//
	// Returns:
	//   - error: Any error that occurred during creation
	//
	// Example:
	//   err := executor.CreateLookupTable(ctx, &QueryParameter{
	//       Operation: "create_table",
	//       Filters: map[string]interface{}{
	//           "tableName": "departments",
	//           "columns": []map[string]string{
	//               {"name": "id", "type": "VARCHAR(36)"},
	//               {"name": "name", "type": "VARCHAR(100)"},
	//           },
	//       },
	//   })
	CreateLookupTable(ctx context.Context, plan *QueryParameter) error

	// GetTableList retrieves list of tables.
	//
	// Parameters:
	//   - ctx: Context for the operation
	//   - plan: QueryParameter containing filter criteria
	//
	// Returns:
	//   - []string: List of table names
	//   - error: Any error that occurred during retrieval
	//
	// Example:
	//   tables, err := executor.GetTableList(ctx, &QueryParameter{
	//       Filters: map[string]interface{}{
	//           "schema": "public",
	//       },
	//   })
	GetTableList(ctx context.Context, plan *QueryParameter) ([]string, error)

	// GetSchemaOfTable retrieves schema information for a table.
	//
	// Parameters:
	//   - ctx: Context for the operation
	//   - plan: QueryParameter containing table name
	//
	// Returns:
	//   - *schema.SchemaInfo: Table schema information
	//   - error: Any error that occurred during retrieval
	//
	// Example:
	//   schemaInfo, err := executor.GetSchemaOfTable(ctx, &QueryParameter{
	//       Filters: map[string]interface{}{
	//           "tableName": "employees",
	//       },
	//   })
	GetSchemaOfTable(ctx context.Context, plan *QueryParameter) (*schema.SchemaInfo, error)

	// InsertRecordData inserts new records into a table.
	//
	// Parameters:
	//   - ctx: Context for the operation
	//   - plan: QueryParameter containing record data
	//
	// Returns:
	//   - error: Any error that occurred during insertion
	//
	// Example:
	//   err := executor.InsertRecordData(ctx, &QueryParameter{
	//       Operation: "insert",
	//       Filters: map[string]interface{}{
	//           "tableName": "employees",
	//           "data": map[string]interface{}{
	//               "id": "emp123",
	//               "name": "John Doe",
	//               "department": "Engineering",
	//           },
	//       },
	//   })
	InsertRecordData(ctx context.Context, plan *QueryParameter) error

	// GetRecord retrieves records from a table.
	//
	// Parameters:
	//   - ctx: Context for the operation
	//   - plan: QueryParameter containing query criteria
	//
	// Returns:
	//   - []map[string]interface{}: List of records matching criteria
	//   - error: Any error that occurred during retrieval
	//
	// Example:
	//   records, err := executor.GetRecord(ctx, &QueryParameter{
	//       Filters: map[string]interface{}{
	//           "tableName": "employees",
	//           "department": "Engineering",
	//           "limit": 10,
	//           "offset": 0,
	//       },
	//   })
	GetRecord(ctx context.Context, plan *QueryParameter) ([]map[string]interface{}, error)
} 