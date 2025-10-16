# CRUD Service - Detailed Architecture

This document provides an in-depth look at the CRUD Service architecture, components, and implementation details.

---

## Overview

The CRUD Service is the core orchestration layer of Nexoan, written in Go and exposing a gRPC interface. It manages all interactions with the three databases (MongoDB, Neo4j, PostgreSQL) and implements the business logic for entity management.

**Location**: `nexoan/crud-api/`  
**Language**: Go  
**Protocol**: gRPC  
**Port**: 50051

---

## Directory Structure

```
nexoan/crud-api/
├── cmd/
│   └── server/
│       ├── service.go          # Main gRPC server implementation
│       ├── service_test.go     # Server tests
│       └── utils.go            # Server utility functions
├── db/
│   ├── config/
│   │   └── config.go           # Database configuration
│   └── repository/
│       ├── mongo/
│       │   ├── mongodb_client.go       # MongoDB connection & client
│       │   ├── mongodb_client_test.go  # MongoDB tests
│       │   └── metadata_handler.go     # Metadata operations
│       ├── neo4j/
│       │   ├── neo4j_client.go         # Neo4j connection & client
│       │   ├── neo4j_client_test.go    # Neo4j tests
│       │   └── graph_entity_handler.go # Entity & relationship ops
│       └── postgres/
│           ├── postgres_client.go      # PostgreSQL connection
│           ├── postgres_client_test.go # PostgreSQL tests
│           ├── data_handler.go         # Attribute operations
│           └── data_handler_test.go    # Attribute tests
├── engine/
│   ├── attribute_resolver.go          # Attribute processing engine
│   ├── attribute_resolver_test.go     # Attribute tests
│   ├── graph_metadata_manager.go      # Graph metadata management
│   └── graph_metadata_test.go         # Graph metadata tests
├── pkg/
│   ├── schema/
│   │   ├── schema.go                  # Schema management
│   │   ├── schema_test.go             # Schema tests
│   │   ├── utils.go                   # Schema utilities
│   │   └── utils_test.go              # Utils tests
│   ├── storageinference/
│   │   ├── inference.go               # Storage type inference
│   │   └── inference_test.go          # Storage tests
│   └── typeinference/
│       ├── inference.go               # Data type inference
│       └── inference_test.go          # Type tests
├── protos/
│   └── types_v1.proto                 # Protobuf definitions
├── lk/
│   └── datafoundation/
│       └── crud-api/
│           ├── types_v1.pb.go         # Generated protobuf Go code
│           └── types_v1_grpc.pb.go    # Generated gRPC Go code
├── go.mod                             # Go module definition
├── go.sum                             # Go dependencies
└── README.md                          # Service documentation
```

---

## gRPC Server Implementation

### Server Structure

```go
type Server struct {
    pb.UnimplementedCrudServiceServer
    mongoRepo    *mongorepository.MongoRepository
    neo4jRepo    *neo4jrepository.Neo4jRepository
    postgresRepo *postgres.PostgresRepository
}
```

The server maintains connections to all three database repositories and implements the gRPC service interface.

### Service Methods

#### 1. CreateEntity

**Signature**: `CreateEntity(ctx context.Context, req *pb.Entity) (*pb.Entity, error)`

**Purpose**: Create a new entity across all databases

**Flow**:
```
1. Save metadata to MongoDB
   ↓
2. Create entity node in Neo4j
   ↓
3. Create relationships in Neo4j
   ↓
4. Process and save attributes to PostgreSQL
   ↓
5. Return created entity
```

**Implementation Details**:
- **MongoDB**: Calls `mongoRepo.HandleMetadata(ctx, req.Id, req)`
  - Stores entity metadata as a document
  - Uses entity ID as document ID for fast lookups
  
- **Neo4j**: Calls `neo4jRepo.HandleGraphEntityCreation(ctx, req)`
  - Creates entity node with properties: id, kind_major, kind_minor, name, created
  - Validates required fields before creation
  
- **Neo4j**: Calls `neo4jRepo.HandleGraphRelationshipsCreate(ctx, req)`
  - Creates relationship edges to other entities
  - Includes relationship properties: id, name, startTime, endTime
  
- **PostgreSQL**: Uses `engine.EntityAttributeProcessor`
  - Infers data types for attribute values
  - Determines storage type (tabular, graph, list, map, scalar)
  - Creates schema if needed
  - Creates table if needed
  - Inserts attribute values

**Error Handling**:
- Logs errors at each stage
- Continues processing even if some attributes fail
- Returns error if critical operations fail (MongoDB, Neo4j entity creation)

**Location**: `nexoan/crud-api/cmd/server/service.go:33`

---

#### 2. ReadEntity

**Signature**: `ReadEntity(ctx context.Context, req *pb.ReadEntityRequest) (*pb.Entity, error)`

**Purpose**: Retrieve entity information from databases

**Flow**:
```
1. Always fetch basic entity info from Neo4j
   ↓
2. If output includes "metadata": fetch from MongoDB
   ↓
3. If output includes "relationships": fetch from Neo4j
   ↓
4. If output includes "attributes": fetch from PostgreSQL
   ↓
5. Assemble and return complete entity
```

**Request Structure**:
```protobuf
message ReadEntityRequest {
    Entity entity = 1;              // Entity with at least ID set
    repeated string output = 2;     // Fields to retrieve: ["metadata", "relationships", "attributes"]
    string activeAt = 3;            // Time for historical queries (optional)
}
```

**Implementation Details**:
- **Basic Info** (always): Calls `neo4jRepo.GetGraphEntity(ctx, req.Entity.Id)`
  - Returns: Kind, Name, Created, Terminated
  
- **Metadata** (if requested): Calls `mongoRepo.GetMetadata(ctx, req.Entity.Id)`
  - Returns: Map of metadata key-value pairs
  
- **Relationships** (if requested): 
  - If no filters: `neo4jRepo.GetFilteredRelationships(ctx, entityId, ...)`
  - If filters provided: Filters by relationship name, direction, related entity
  - Supports temporal queries with `activeAt` parameter
  
- **Attributes** (if requested): Calls `postgresRepo.GetAttributesForEntity(ctx, entityId)`
  - Returns: Map of attribute names to time-based value lists
  - Filters by `activeAt` if provided

**Selective Retrieval**:
The `output` parameter allows clients to request only needed fields, reducing:
- Network bandwidth
- Database load
- Response time

**Location**: `nexoan/crud-api/cmd/server/service.go:88`

---

#### 3. UpdateEntity

**Signature**: `UpdateEntity(ctx context.Context, req *pb.UpdateEntityRequest) (*pb.Entity, error)`

**Purpose**: Update existing entity information

**Request Structure**:
```protobuf
message UpdateEntityRequest {
    string id = 1;          // Entity ID to update
    Entity entity = 2;      // Updated entity data
}
```

**Flow**:
```
1. Update metadata in MongoDB
   ↓
2. Update entity node in Neo4j
   ↓
3. Update relationships in Neo4j
   ↓
4. Update attributes in PostgreSQL
   ↓
5. Return updated entity
```

**Update Strategy**:
- **Metadata**: Replace existing metadata document
- **Entity Node**: Update node properties
- **Relationships**: Add new relationships, don't delete existing
- **Attributes**: Append new time-based values

**Location**: `nexoan/crud-api/cmd/server/service.go` (UpdateEntity method)

---

#### 4. DeleteEntity

**Signature**: `DeleteEntity(ctx context.Context, req *pb.EntityId) (*pb.Empty, error)`

**Purpose**: Remove entity from all databases

**Request Structure**:
```protobuf
message EntityId {
    string id = 1;          // Entity ID to delete
}
```

**Flow**:
```
1. Delete metadata from MongoDB
   ↓
2. Delete entity node and relationships from Neo4j
   ↓
3. Delete attributes from PostgreSQL
   ↓
4. Return empty response
```

**Deletion Strategy**:
- **MongoDB**: Delete metadata document by ID
- **Neo4j**: Delete entity node (CASCADE deletes relationships)
- **PostgreSQL**: Delete from entity_attributes, CASCADE deletes attribute values

**Location**: `nexoan/crud-api/cmd/server/service.go` (DeleteEntity method)

---

## Repository Layer

### MongoDB Repository

**File**: `nexoan/crud-api/db/repository/mongo/mongodb_client.go`

**Structure**:
```go
type MongoRepository struct {
    client     *mongo.Client
    database   *mongo.Database
    collection *mongo.Collection
}
```

**Key Methods**:

1. **NewMongoRepository**: Initialize MongoDB connection
2. **HandleMetadata**: Save metadata for an entity
3. **GetMetadata**: Retrieve metadata by entity ID
4. **DeleteMetadata**: Remove metadata for an entity
5. **Close**: Close MongoDB connection

**Connection String**:
```
mongodb://<user>:<password>@<host>:<port>/<database>?authSource=admin
```

**Collections**:
- `metadata`: Production metadata
- `metadata_test`: Test metadata

---

### Neo4j Repository

**File**: `nexoan/crud-api/db/repository/neo4j/neo4j_client.go`

**Structure**:
```go
type Neo4jRepository struct {
    driver  neo4j.Driver
    session neo4j.Session
}
```

**Key Methods**:

1. **NewNeo4jRepository**: Initialize Neo4j connection
2. **HandleGraphEntityCreation**: Create entity node
3. **GetGraphEntity**: Retrieve entity node by ID
4. **HandleGraphRelationshipsCreate**: Create relationships
5. **GetGraphRelationships**: Get all relationships for entity
6. **GetFilteredRelationships**: Get relationships with filters
7. **DeleteGraphEntity**: Delete entity node and relationships
8. **Close**: Close Neo4j connection

**Cypher Query Examples**:

**Create Entity Node**:
```cypher
CREATE (e:Entity {
    id: $id,
    kind_major: $kind_major,
    kind_minor: $kind_minor,
    name: $name,
    created: $created,
    terminated: $terminated
})
```

**Create Relationship**:
```cypher
MATCH (source:Entity {id: $sourceId})
MATCH (target:Entity {id: $targetId})
CREATE (source)-[r:RELATIONSHIP_TYPE {
    id: $id,
    name: $name,
    startTime: $startTime,
    endTime: $endTime
}]->(target)
```

**Get Filtered Relationships**:
```cypher
MATCH (e:Entity {id: $entityId})-[r]->(related:Entity)
WHERE r.name = $relationshipName
  AND ($activeAt IS NULL OR 
       (r.startTime <= $activeAt AND 
        (r.endTime IS NULL OR r.endTime >= $activeAt)))
RETURN r, related
```

---

### PostgreSQL Repository

**File**: `nexoan/crud-api/db/repository/postgres/postgres_client.go`

**Structure**:
```go
type PostgresRepository struct {
    db *sql.DB
}
```

**Key Methods**:

1. **NewPostgresRepository**: Initialize PostgreSQL connection
2. **CreateAttributeSchema**: Create schema for new attribute type
3. **GetAttributeSchema**: Retrieve existing schema
4. **CreateAttributeTable**: Create table for attribute values
5. **InsertAttributeValue**: Insert attribute value
6. **GetAttributesForEntity**: Get all attributes for entity
7. **DeleteEntityAttributes**: Delete attributes for entity
8. **Close**: Close PostgreSQL connection

**Schema Tables**:

**attribute_schemas**:
```sql
CREATE TABLE attribute_schemas (
    id SERIAL PRIMARY KEY,
    kind_major VARCHAR(255) NOT NULL,
    kind_minor VARCHAR(255),
    attr_name VARCHAR(255) NOT NULL,
    data_type VARCHAR(50) NOT NULL,
    storage_type VARCHAR(50) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(kind_major, kind_minor, attr_name)
);
```

**entity_attributes**:
```sql
CREATE TABLE entity_attributes (
    id SERIAL PRIMARY KEY,
    entity_id VARCHAR(255) NOT NULL,
    attr_name VARCHAR(255) NOT NULL,
    schema_id INTEGER REFERENCES attribute_schemas(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(entity_id, attr_name)
);
```

**Dynamic Attribute Tables** (e.g., `attr_Person_salary`):
```sql
CREATE TABLE attr_Person_salary (
    id SERIAL PRIMARY KEY,
    entity_id VARCHAR(255) NOT NULL,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP,
    value INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (entity_id) REFERENCES entity_attributes(entity_id) ON DELETE CASCADE
);
```

---

## Engine Layer

### Attribute Processor

**File**: `nexoan/crud-api/engine/attribute_resolver.go`

**Structure**:
```go
type EntityAttributeProcessor struct {
    // Processes entity attributes
}

type AttributeResult struct {
    AttributeName string
    Success       bool
    Error         error
    StorageType   string
    DataType      string
}
```

**Process Flow**:
```
1. For each attribute in entity
   ↓
2. Infer data type (TypeInference)
   ↓
3. Infer storage type (StorageInference)
   ↓
4. Check/Create schema in attribute_schemas
   ↓
5. Check/Create table (attr_Kind_AttrName)
   ↓
6. Insert values into table
   ↓
7. Link in entity_attributes
   ↓
8. Return results
```

**Key Methods**:
1. **ProcessEntityAttributes**: Main processing function
2. **inferAttributeType**: Determine data type
3. **inferStorageType**: Determine storage type
4. **ensureSchema**: Create schema if not exists
5. **ensureTable**: Create table if not exists
6. **insertAttributeValues**: Insert time-based values

---

### Type Inference

**File**: `nexoan/crud-api/pkg/typeinference/inference.go`

**Supported Types**:
```go
const (
    DataTypeInt      = "int"
    DataTypeFloat    = "float"
    DataTypeString   = "string"
    DataTypeBool     = "bool"
    DataTypeNull     = "null"
    DataTypeDate     = "date"
    DataTypeTime     = "time"
    DataTypeDateTime = "datetime"
)
```

**Inference Rules**:
1. Check if number → int or float based on decimal
2. Check if string matches date pattern → date
3. Check if string matches time pattern → time
4. Check if string matches datetime pattern → datetime
5. Check if boolean → bool
6. Check if null → null
7. Default → string

**Date/Time Patterns**:
- Date: `YYYY-MM-DD`, `DD/MM/YYYY`, `MM/DD/YYYY`
- Time: `HH:MM:SS`, `HH:MM:SS.mmm`, `HH:MM:SSZ`
- DateTime: RFC3339, `YYYY-MM-DD HH:MM:SS`

---

### Storage Inference

**File**: `nexoan/crud-api/pkg/storageinference/inference.go`

**Storage Types**:
```go
const (
    StorageTypeTabular = "TABULAR"
    StorageTypeGraph   = "GRAPH"
    StorageTypeList    = "LIST"
    StorageTypeMap     = "MAP"
    StorageTypeScalar  = "SCALAR"
)
```

**Inference Logic**:
```go
func InferStorageType(data interface{}) string {
    // Check structure
    if hasColumnsAndRows(data) {
        return StorageTypeTabular
    }
    if hasNodesAndEdges(data) {
        return StorageTypeGraph
    }
    if hasItems(data) {
        return StorageTypeList
    }
    if isSingleValue(data) {
        return StorageTypeScalar
    }
    return StorageTypeMap  // Default
}
```

**Storage Mapping**:
- **TABULAR** → PostgreSQL table with columns
- **GRAPH** → Neo4j subgraph (future feature)
- **LIST** → PostgreSQL array or JSONB
- **MAP** → PostgreSQL JSONB
- **SCALAR** → PostgreSQL native type column

---

## Configuration

### Environment Variables

```bash
# Neo4j Configuration
NEO4J_URI=bolt://neo4j:7687
NEO4J_USER=neo4j
NEO4J_PASSWORD=neo4j123

# MongoDB Configuration
MONGO_URI=mongodb://admin:admin123@mongodb:27017/admin?authSource=admin
MONGO_DB_NAME=nexoan
MONGO_COLLECTION=metadata
MONGO_ADMIN_USER=admin
MONGO_ADMIN_PASSWORD=admin123

# PostgreSQL Configuration
POSTGRES_HOST=postgres
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DB=nexoan
POSTGRES_SSL_MODE=disable

# Service Configuration
CRUD_SERVICE_HOST=0.0.0.0
CRUD_SERVICE_PORT=50051
```

### Configuration Loading

**File**: `nexoan/crud-api/db/config/config.go`

Configuration is loaded from environment variables at startup and validated before creating database connections.

---

## Testing

### Unit Tests

Each component has comprehensive unit tests:
- **Server Tests**: `cmd/server/service_test.go`
- **MongoDB Tests**: `db/repository/mongo/mongodb_client_test.go`
- **Neo4j Tests**: `db/repository/neo4j/neo4j_client_test.go`
- **PostgreSQL Tests**: `db/repository/postgres/data_handler_test.go`
- **Engine Tests**: `engine/attribute_resolver_test.go`

### Running Tests

```bash
# Run all tests
go test -v ./...

# Run tests with coverage
go test -v -cover ./...

# Run specific test
go test -v ./cmd/server/...

# Run with race detection
go test -race -v ./...
```

### Integration Tests

Integration tests require running databases:
```bash
# Start databases
docker-compose up -d mongodb neo4j postgres

# Run tests
go test -v ./...
```

---

## Performance Considerations

### Connection Pooling

Each repository manages its own connection pool:
- **MongoDB**: Native driver connection pool
- **Neo4j**: Driver connection pool
- **PostgreSQL**: `database/sql` connection pool

### Parallel Operations

Entity creation operations are performed in parallel where possible:
- Metadata save (MongoDB)
- Entity node creation (Neo4j)
- Relationship creation (Neo4j)
- Attribute processing (PostgreSQL)

### Indexing Strategy

**MongoDB**:
- Index on `_id` (entity ID) - automatic

**Neo4j**:
- Index on `:Entity(id)` for fast entity lookups
- Index on relationship properties for filtering

**PostgreSQL**:
- Primary keys on all tables
- Foreign keys for referential integrity
- Index on `entity_id` in attribute tables
- Index on `start_time` for temporal queries

---

## Error Handling

### Error Types

1. **Connection Errors**: Database unavailable
2. **Validation Errors**: Invalid entity data
3. **Constraint Errors**: Duplicate keys, foreign key violations
4. **Query Errors**: Malformed queries, syntax errors
5. **Transaction Errors**: Rollback scenarios

### Error Responses

gRPC errors follow standard codes:
- `InvalidArgument`: Invalid entity data
- `NotFound`: Entity not found
- `AlreadyExists`: Entity already exists
- `Internal`: Database or system error
- `Unavailable`: Database connection failed

---

## Logging

### Log Levels

- **INFO**: Normal operations (entity created, read, updated, deleted)
- **WARN**: Recoverable errors (attribute processing failed)
- **ERROR**: Critical errors (database connection failed, entity creation failed)

### Log Format

```
[timestamp] [level] [component] message
```

Example:
```
2024-10-14 10:30:45 INFO [server.CreateEntity] Creating Entity: entity123
2024-10-14 10:30:45 INFO [mongoRepo] Successfully saved metadata for entity: entity123
2024-10-14 10:30:45 INFO [neo4jRepo] Successfully saved entity in Neo4j: entity123
```

---

## Future Enhancements

1. **Connection Pooling Optimization**: Tune pool sizes for performance
2. **Distributed Transactions**: Implement two-phase commit across databases
3. **Caching Layer**: Redis for frequently accessed entities
4. **Batch Operations**: Support bulk create/update/delete
5. **Streaming**: gRPC streaming for large result sets
6. **Metrics**: Prometheus metrics for monitoring
7. **Tracing**: Distributed tracing with OpenTelemetry

---

## Related Documentation

- [Main Architecture Overview](./overview.md)
- [Architecture Diagrams](./diagrams.md)
- [API Layer Details](./api-layer-details.md)
- [Database Schemas](./database-schemas.md)
- [CRUD Service README](../../nexoan/crud-api/README.md)

---

**Document Version**: 1.0  
**Last Updated**: October 2024  
**Component**: CRUD Service

