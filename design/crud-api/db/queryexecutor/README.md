# Query Executor Layer

## Overview
The Query Executor layer serves as an intermediary between the service layer and the repository layer in the CRUD API. It provides a clean abstraction for executing database operations while maintaining separation of concerns.

## Query Plan
All executor operations now accept a QueryPlan as input, which provides a standardized way to specify operation details:

```go
type QueryPlan struct {
    // Operation type (e.g., "get_metadata", "create_entity", "get_relationships")
    Operation string

    // Parameters for the operation
    Parameters map[string]interface{}

    // Filtering criteria
    Filters map[string]interface{}

    // Time-based constraints
    TimeRange struct {
        StartTime string
        EndTime   string
        ActiveAt  string
    }

    // Pagination
    Pagination struct {
        Limit  int32
        Offset int32
    }
}
```

### Example Query Plans

1. **MongoDB Metadata Operation**:
```go
plan := &QueryPlan{
    Operation: "get_metadata",
    Parameters: map[string]interface{}{
        "entityId": "entity123"
    }
}
```

2. **Neo4j Relationship Query**:
```go
plan := &QueryPlan{
    Operation: "get_filtered_relationships",
    Parameters: map[string]interface{}{
        "entityId": "entity123",
        "relationshipId": "rel456",
        "relationship": "MANAGES",
        "relatedEntityId": "entity789",
        "direction": "outgoing"
    },
    TimeRange: struct {
        StartTime: "2024-01-01",
        EndTime: "2024-12-31",
        ActiveAt: "2024-06-15"
    }
}
```

## Architecture

### Directory Structure
```
queryexecutor/
├── interfaces/
│   └── executor.go       # Defines interfaces for different database executors
├── mongo/
│   └── executor.go       # MongoDB-specific query executor implementation
└── neo4j/
    └── executor.go       # Neo4j-specific query executor implementation
```

### Interface Definitions

#### MongoExecutor Interface
```go
type MongoExecutor interface {
    // Metadata Operations with query plan
    HandleMetadata(ctx context.Context, plan *QueryPlan) error
    GetMetadata(ctx context.Context, plan *QueryPlan) (map[string]*anypb.Any, error)
}
```

#### Neo4jExecutor Interface
```go
type Neo4jExecutor interface {
    // Entity Operations with query plan
    GetGraphEntity(ctx context.Context, plan *QueryPlan) (*pb.Kind, *pb.TimeBasedValue, string, string, error)
    HandleGraphEntityCreation(ctx context.Context, plan *QueryPlan) (bool, error)
    HandleGraphEntityUpdate(ctx context.Context, plan *QueryPlan) (bool, error)
    HandleGraphEntityFilter(ctx context.Context, plan *QueryPlan) ([]map[string]interface{}, error)

    // Relationship Operations with query plan
    GetGraphRelationships(ctx context.Context, plan *QueryPlan) (map[string]*pb.Relationship, error)
    GetFilteredRelationships(ctx context.Context, plan *QueryPlan) (map[string]*pb.Relationship, error)
    HandleGraphRelationshipsCreate(ctx context.Context, plan *QueryPlan) error
    HandleGraphRelationshipsUpdate(ctx context.Context, plan *QueryPlan) error
}
```

## Query Execution

The Query Executor accepts a query plan as input, which specifies:
- The type of operation to perform (metadata, graph, or relationship)
- Required parameters for the operation
- Any filtering or sorting requirements
- Time-based constraints (if applicable)

Based on this query plan, the executor:
1. Routes the query to the appropriate database executor (MongoDB or Neo4j)
2. Extracts required parameters from the plan
3. Validates parameters and performs type assertions
4. Executes the operation using the corresponding repository
5. Returns the results in the expected format

## Implementation Details

### MongoDB Query Executor

The MongoDB Query Executor (`MongoQueryExecutor`) handles metadata-related operations:

1. **Structure**:
   ```go
   type MongoQueryExecutor struct {
       repository *mongo.MongoRepository
   }
   ```

2. **Key Operations**:
   - `HandleMetadata`: Manages entity metadata storage and updates
   - `GetMetadata`: Retrieves metadata for a specific entity

### Neo4j Query Executor

The Neo4j Query Executor (`Neo4jQueryExecutor`) manages graph operations:

1. **Structure**:
   ```go
   type Neo4jQueryExecutor struct {
       repository *neo4j.Neo4jRepository
   }
   ```

2. **Key Operations**:
   - Entity Operations:
     - `GetGraphEntity`: Retrieves entity details
     - `HandleGraphEntityCreation`: Creates new entities
     - `HandleGraphEntityUpdate`: Updates existing entities
     - `HandleGraphEntityFilter`: Filters entities based on criteria
   
   - Relationship Operations:
     - `GetGraphRelationships`: Retrieves entity relationships
     - `GetFilteredRelationships`: Gets filtered relationships
     - `HandleGraphRelationshipsCreate`: Creates relationships
     - `HandleGraphRelationshipsUpdate`: Updates relationships

