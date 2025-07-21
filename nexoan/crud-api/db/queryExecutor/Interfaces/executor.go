package interfaces

import (
	"context"
	"lk/datafoundation/crud-api/pkg/schema"
	pb "lk/datafoundation/crud-api/lk/datafoundation/crud-api"
	"google.golang.org/protobuf/types/known/anypb"
)



// Usage examples:
//
//	// Example : Handling metadata result
//	result, err := executor.Execute(ctx, &QueryParameter{
//		Operation: "get_metadata",
//		Parameters: map[string]interface{}{
//			"entityId": "123",
//		},
//	})
//	if err != nil {
//		return err
//	}
//	
//	if result.GetType() == "metadata" {
//		metadata := result.GetData().(map[string]*anypb.Any)
//		department := metadata["department"].String()
//		role := metadata["role"].String()
//	}

//	}
type ExecutorResult interface {
	// GetType returns the type of the result
	// Common types include:
	// - "metadata": For document metadata results
	// - "entity": For graph entity results
	// - "relationships": For graph relationship results
	// - "records": For tabular data results
	// - "schema": For table schema results
	GetType() string

	// GetData returns the actual result data
	// The type of the returned interface{} depends on GetType():
	// - "metadata" returns map[string]*anypb.Any
	// - "entity" returns map[string]interface{} with "kind" and "name" keys
	// - "relationships" returns map[string]*pb.Relationship
	// - "records" returns []map[string]interface{}
	// - "schema" returns *schema.SchemaInfo
	GetData() interface{}
}

// Executor is the main parent interface that defines common behavior for all database executors.
// It provides a unified way to execute queries across different database types.
//
// Example Usage:
//
//	executor := GetExecutor()
//	result, err := executor.Execute(ctx, &QueryParameter{
//		Operation: "get_metadata",
//		Parameters: map[string]interface{}{
//			"entityId": "123",
//		},
//	})
type Executor interface {
	// Execute processes a query based on the provided QueryParameter.
	// It routes the query to the appropriate database implementation.
	//
	// Parameters:
	//   - ctx: Context for the operation
	//   - queryParameter: QueryParameter containing operation details
	//
	// Returns:
	//   - ExecutorResult: Result of the operation with type information
	//   - error: Any error that occurred during execution
	//
	// Example:
	//   result, err := executor.Execute(ctx, &QueryParameter{
	//       Operation: "get_metadata",
	//       Parameters: map[string]interface{}{
	//           "entityId": "123",
	//       },
	//   })
	Execute(ctx context.Context, queryParameter *QueryParameter) (ExecutorResult, error)
}

// QueryParameter defines the structure for query execution parameters.
// It provides a standardized way to specify operation details across different database types.
//
// Example:
//
//	param := &QueryParameter{
//		Operation: "get_metadata",
//		Parameters: map[string]interface{}{
//			"entityId": "123",
//			"kind": "Person",
//		},
//		TimeParameters: struct{
//			StartTime: "2024-01-01T00:00:00Z",
//			EndTime: "2024-12-31T23:59:59Z",
//		},
//	}
type QueryParameter struct {
	// Operation specifies the type of operation to perform
	// Valid values include:
	// - "get_metadata": Retrieve entity metadata
	// - "create_metadata": Create new metadata
	// - "update_metadata": Update existing metadata
	// - "delete_metadata": Delete metadata
	// - "create_entity": Create a new entity
	// - "update_entity": Update an existing entity
	// - "delete_entity": Delete an entity
	// - "get_relationships": Get entity relationships
	Operation string

	// Parameters specify operation parameters
	// Common parameter keys:
	// - "entityId": ID of the entity
	// - "kind": Entity kind/type
	// - "name": Entity name
	// - "metadata": Metadata key-value pairs
	Parameters map[string]interface{}

	// TimeParameters specifies temporal constraints for the operation
	TimeParameters struct {
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
//		Parameters: map[string]interface{}{
//			"entityId": "123",
//		},
//	})
type DocumentExecutor interface {
	Executor

	// HandleMetadata manages entity metadata operations.
	//
	// Parameters:
	//   - ctx: Context for the operation
	//   - queryParameter: QueryParameter containing metadata operation details
	//
	// Returns:
	//   - error: Any error that occurred during the operation
	//
	// Example:
	//   err := executor.HandleMetadata(ctx, &QueryParameter{
	//       Operation: "update_metadata",
	//       Parameters: map[string]interface{}{
	//           "entityId": "123",
	//           "metadata": map[string]interface{}{
	//               "department": "Engineering",
	//           },
	//       },
	//   })
	HandleMetadata(ctx context.Context, queryParameter *QueryParameter) error

	// GetMetadata retrieves metadata for an entity.
	//
	// Parameters:
	//   - ctx: Context for the operation
	//   - queryParameter: QueryParameter containing entity identifier
	//
	// Returns:
	//   - map[string]*anypb.Any: Map of metadata key-value pairs
	//   - error: Any error that occurred during retrieval
	//
	// Example:
	//   metadata, err := executor.GetMetadata(ctx, &QueryParameter{
	//       Parameters: map[string]interface{}{
	//           "entityId": "123",
	//       },
	//   })
	GetMetadata(ctx context.Context, queryParameter *QueryParameter) (map[string]*anypb.Any, error)
}

// GraphExecutor processes data defined as Graph storage type.
// It handles entity and relationship operations in a graph structure.
//
// Example Usage:
//
//	executor := GetGraphExecutor()
//	entity, err := executor.GraphEntity(ctx, &QueryParameter{
//		Parameters: map[string]interface{}{
//			"entityId": "123",
//		},
//	})
type GraphExecutor interface {
	Executor

	// GraphEntity retrieves an entity's graph information.
	//
	// Parameters:
	//   - ctx: Context for the operation
	//   - queryParameter: QueryParameter containing entity identifier
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
	//       Parameters: map[string]interface{}{
	//           "entityId": "123",
	//       },
	//   })
	GetGraphEntity(ctx context.Context, queryParameter *QueryParameter) (*pb.Kind, *pb.TimeBasedValue, string, string, error)

	// CreateGraphEntity creates a new entity in the graph.
	//
	// Parameters:
	//   - ctx: Context for the operation
	//   - queryParameter: QueryParameter containing entity identifier
	//
	// Returns:
	//   - *pb.Kind: Entity kind information
	//   - *pb.TimeBasedValue: Entity name with temporal information
	//   - string: Creation timestamp
	//   - string: Termination timestamp
	//   - error: Any error that occurred during retrieval
	//
	// Example:
	//   kind, name, created, terminated, err := executor.CreateGraphEntity(ctx, &QueryParameter{
	//       Parameters: map[string]interface{}{
	//           "entityId": "123",
	//       },
	//   })

	GraphEntityCreate(ctx context.Context, queryParameter *QueryParameter) (bool, error)

	// GraphEntityUpdate updates an existing entity in the graph.
	//
	// Parameters:
	//   - ctx: Context for the operation
	//   - queryParameter: QueryParameter containing update details
	//
	// Returns:
	//   - bool: True if update was successful
	//   - error: Any error that occurred during update
	//
	// Example:
	//   success, err := executor.GraphEntityUpdate(ctx, &QueryParameter{
	//       Operation: "update_entity",
	//       Parameters: map[string]interface{}{
	//           "entityId": "123",
	//           "name": "Jane Doe",
	//       },
	//   })
	GraphEntityUpdate(ctx context.Context, queryParameter *QueryParameter) (bool, error)

	// GraphEntityFilter filters entities based on criteria.
	//
	// Parameters:
	//   - ctx: Context for the operation
	//   - queryParameter: QueryParameter containing filter criteria
	//
	// Returns:
	//   - []map[string]interface{}: List of matching entities
	//   - error: Any error that occurred during filtering
	//
	// Example:
	//   entities, err := executor.GraphEntityFilter(ctx, &QueryParameter{
	//       Parameters: map[string]interface{}{
	//           "kind": "Person",
	//           "department": "Engineering",
	//       },
	//   })
	GraphEntityFilter(ctx context.Context, queryParameter *QueryParameter) ([]map[string]interface{}, error)

	// GraphAllRelationshipsByNode retrieves all relationships for an entity.
	//
	// Parameters:
	//   - ctx: Context for the operation
	//   - queryParameter: QueryParameter containing entity identifier
	//
	// Returns:
	//   - map[string]*pb.Relationship: Map of relationships keyed by relationship ID
	//   - error: Any error that occurred during retrieval
	//
	// Example:
	//   relationships, err := executor.GraphAllRelationshipsByNode(ctx, &QueryParameter{
	//       Parameters: map[string]interface{}{
	//           "entityId": "123",
	//       },
	//   })
	GraphAllRelationshipsByNode(ctx context.Context, queryParameter *QueryParameter) (map[string]*pb.Relationship, error)

	// GetRelationshipsByFilter retrieves relationships based on specified criteria.
	// This function allows filtering relationships based on various parameters
	// such as type, direction, and temporal constraints.
	//
	// Parameters:
	//   - ctx: Context for the operation
	//   - queryParameter: QueryParameter containing filter criteria
	//
	// Returns:
	//   - map[string]*pb.Relationship: Map of filtered relationships
	//   - error: Any error that occurred during filtering
	//
	// Example:
	//   relationships, err := executor.GetRelationshipsByFilter(ctx, &QueryParameter{
	//       Parameters: map[string]interface{}{
	//           "entityId": "123",
	//           "relationshipType": "MANAGES",
	//           "direction": "OUTGOING",
	//       },
	//       TimeParameters: struct{
	//           ActiveAt: "2024-03-21T00:00:00Z",
	//       },
	//   })
	GetRelationshipsByFilter(ctx context.Context, queryParameter *QueryParameter) (map[string]*pb.Relationship, error)

	// RelationshipsCreate creates new relationships for an entity.
	//
	// Parameters:
	//   - ctx: Context for the operation
	//   - queryParameter: QueryParameter containing relationship details
	//
	// Returns:
	//   - error: Any error that occurred during creation
	//
	// Example:
	//   err := executor.RelationshipsCreate(ctx, &QueryParameter{
	//       Operation: "create_relationship",
	//       Parameters: map[string]interface{}{
	//           "entityId": "123",
	//           "relatedEntityId": "456",
	//           "relationshipType": "MANAGES",
	//           "startTime": "2024-01-01T00:00:00Z",
	//       },
	//   })
	RelationshipsCreate(ctx context.Context, queryParameter *QueryParameter) error

	// UpdateRelationships updates existing relationships.
	//
	// Parameters:
	//   - ctx: Context for the operation
	//   - queryParameter: QueryParameter containing update details
	//
	// Returns:
	//   - error: Any error that occurred during update
	//
	// Example:
	//   err := executor.UpdateRelationships(ctx, &QueryParameter{
	//       Operation: "update_relationship",
	//       Parameters: map[string]interface{}{
	//           "relationshipId": "rel123",
	//           "endTime": "2024-03-21T00:00:00Z",
	//       },
	//   })
	UpdateRelationships(ctx context.Context, queryParameter *QueryParameter) error
}

// TabularExecutor interface defines operations for relational databases (e.g., PostgreSQL).
// It handles table management and data operations in a tabular structure.
//
// Example Usage:
//
//	executor := GetTabularExecutor()
//	exists, err := executor.TableExists(ctx, &QueryParameter{
//		Parameters: map[string]interface{}{
//			"tableName": "employees",
//		},
//	})
type TabularExecutor interface {
	Executor

	// TableExists checks if a table exists in the database.
	//
	// Parameters:
	//   - ctx: Context for the operation
	//   - queryParameter: QueryParameter containing table name
	//
	// Returns:
	//   - bool: True if table exists
	//   - error: Any error that occurred during check
	//
	// Example:
	//   exists, err := executor.TableExists(ctx, &QueryParameter{
	//       Parameters: map[string]interface{}{
	//           "tableName": "employees",
	//       },
	//   })
	TableExists(ctx context.Context, queryParameter *QueryParameter) (bool, error)

	// CreateLookupTable creates a new lookup table.
	//
	// Parameters:
	//   - ctx: Context for the operation
	//   - queryParameter: QueryParameter containing table schema
	//
	// Returns:
	//   - error: Any error that occurred during creation
	//
	// Example:
	//   err := executor.CreateLookupTable(ctx, &QueryParameter{
	//       Operation: "create_table",
	//       Parameters: map[string]interface{}{
	//           "tableName": "departments",
	//           "columns": []map[string]string{
	//               {"name": "id", "type": "VARCHAR(36)"},
	//               {"name": "name", "type": "VARCHAR(100)"},
	//           },
	//       },
	//   })
	CreateLookupTable(ctx context.Context, queryParameter *QueryParameter) error

	// GetTableList retrieves list of tables.
	//
	// Parameters:
	//   - ctx: Context for the operation
	//   - queryParameter: QueryParameter containing filter criteria
	//
	// Returns:
	//   - []string: List of table names
	//   - error: Any error that occurred during retrieval
	//
	// Example:
	//   tables, err := executor.GetTableList(ctx, &QueryParameter{
	//       Parameters: map[string]interface{}{
	//           "schema": "public",
	//       },
	//   })
	GetTableList(ctx context.Context, queryParameter *QueryParameter) ([]string, error)

	// GetTableSchema retrieves schema of a table.
	//
	// Parameters:
	//   - ctx: Context for the operation
	//   - queryParameter: QueryParameter containing table name
	//
	// Returns:
	//   - *schema.SchemaInfo: Table schema information
	//   - error: Any error that occurred during retrieval
	//
	// Example:
	//   schemaInfo, err := executor.GetTableSchema(ctx, &QueryParameter{
	//       Parameters: map[string]interface{}{
	//           "tableName": "employees",
	//       },
	//   })
	GetTableSchema(ctx context.Context, queryParameter *QueryParameter) (*schema.SchemaInfo, error)

	// InsertRecord inserts new records into a table.
	//
	// Parameters:
	//   - ctx: Context for the operation
	//   - queryParameter: QueryParameter containing record data
	//
	// Returns:
	//   - error: Any error that occurred during insertion
	//
	// Example:
	//   err := executor.InsertRecord(ctx, &QueryParameter{
	//       Operation: "insert",
	//       Parameters: map[string]interface{}{
	//           "tableName": "employees",
	//           "data": map[string]interface{}{
	//               "id": "emp123",
	//               "name": "John Doe",
	//               "department": "Engineering",
	//           },
	//       },
	//   })
	InsertRecord(ctx context.Context, queryParameter *QueryParameter) error

	// GetRecord retrieves records from a table.
	//
	// Parameters:
	//   - ctx: Context for the operation
	//   - queryParameter: QueryParameter containing query criteria
	//
	// Returns:
	//   - []map[string]interface{}: List of records matching criteria
	//   - error: Any error that occurred during retrieval
	//
	// Example:
	//   records, err := executor.GetRecord(ctx, &QueryParameter{
	//       Parameters: map[string]interface{}{
	//           "tableName": "employees",
	//           "department": "Engineering",
	//           "limit": 10,
	//           "offset": 0,
	//       },
	//   })
	GetRecord(ctx context.Context, queryParameter *QueryParameter) ([]map[string]interface{}, error)
	
} 