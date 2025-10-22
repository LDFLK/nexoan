# OpenGIN Architecture Overview

## System Overview

**OpenGIN** is a multi-database, microservices-based data management system that handles entities with metadata, attributes, and relationships. The architecture follows a layered approach with REST/gRPC communication protocols.

---

## High-Level Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                        CLIENT LAYER                             │
│                   (Web/Mobile/Desktop Clients)                  │
└──────────────────────────┬──────────────────────────────────────┘
                           │ HTTP/REST + JSON
                           │
┌──────────────────────────┴─────────────────────────────────────┐
│                        API LAYER                               │
│  ┌─────────────────────┐        ┌──────────────────────┐       │
│  │   Update API        │        │    Query API         │       │
│  │   (Ballerina)       │        │    (Ballerina)       │       │
│  │   Port: 8080        │        │    Port: 8081        │       │
│  │   - CREATE          │        │    - READ/QUERY      │       │
│  │   - UPDATE          │        │    - FILTER          │       │
│  │   - DELETE          │        │    - SEARCH          │       │
│  └─────────┬───────────┘        └──────────┬───────────┘       │
└────────────┼──────────────────────────────-┼───────────────────┘
             │                               │
             │        gRPC + Protobuf        │
             │                               │
┌────────────┴───────────────────────────────┴───────────────────┐
│                    SERVICE LAYER                               │
│                                                                │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │              CRUD Service (Go)                           │  │
│  │              Port: 50051 (gRPC)                          │  │
│  │                                                          │  │
│  │  Components:                                             │  │
│  │  ┌────────────────┐  ┌─────────────────────────────┐     │  │
│  │  │  gRPC Server   │  │  Engine                     │     │  │
│  │  │  - CreateEntity│  │  - AttributeProcessor       │     │  │
│  │  │  - ReadEntity  │  │  - GraphMetadataManager     │     │  │
│  │  │  - UpdateEntity│  │  - TypeInference            │     │  │
│  │  │  - DeleteEntity│  │  - StorageInference         │     │  │
│  │  └────────────────┘  └─────────────────────────────┘     │  │
│  │                                                          │  │
│  │  ┌──────────────────────────────────────────────────┐    │  │
│  │  │         Repository Layer                         │    │  │
│  │  │  ┌────────────┐ ┌────────────┐ ┌─────────────┐   │    │  │
│  │  │  │  MongoDB   │ │   Neo4j    │ │  PostgreSQL │   │    │  │
│  │  │  │  Repo      │ │   Repo     │ │    Repo     │   │    │  │
│  │  │  └────────────┘ └────────────┘ └─────────────┘   │    │  │
│  │  └──────────────────────────────────────────────────┘    │  │
│  └──────────────────────────────────────────────────────────┘  │
└────────┬─────────────────┬─────────────────────┬───────────────┘
         │                 │                     │
         │ Native Protocol │ Bolt Protocol       │ PostgreSQL Protocol
         │                 │                     │
┌────────┴─────────────────┴─────────────────────┴───────────────┐
│                      DATABASE LAYER                            │
│                                                                │
│  ┌───────────────┐   ┌─────────────────┐   ┌────────────────┐  │
│  │   MongoDB     │   │     Neo4j       │   │  PostgreSQL    │  │
│  │  Port: 27017  │   │ Port: 7474/7687 │   │  Port: 5432    │  │
│  │               │   │                 │   │                │  │
│  │  Storage:     │   │  Storage:       │   │  Storage:      │  │
│  │  - Metadata   │   │  - Entities     │   │  - Attributes  │  │
│  │  - Key-Value  │   │  - Relationships│   │  - Schemas     │  │
│  │    Pairs      │   │  - Graph Data   │   │  - Time-Series │  │
│  └───────────────┘   └─────────────────┘   └────────────────┘  │
└────────────────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────────────────┐
│                   SUPPORTING SERVICES                            │
│  ┌──────────────┐  ┌────────────────┐  ┌──────────────────────┐  │
│  │   Cleanup    │  │  Backup/Restore│  │    Swagger UI        │  │
│  │   Service    │  │    Service     │  │    (API Docs)        │  │
│  └──────────────┘  └────────────────┘  └──────────────────────┘  │
└──────────────────────────────────────────────────────────────────┘
```

---

## Architecture Layers

### 1. API Layer (Client-Facing Services)

#### Update API (Ballerina, Port 8080)
- **Purpose**: Handle entity mutations (CREATE, UPDATE, DELETE)
- **Technology**: Ballerina REST service
- **Location**: `nexoan/update-api/`
- **Responsibilities**:
  - Accept JSON payloads from clients
  - Validate request structure
  - Convert JSON to Protobuf Entity messages
  - Communicate with CRUD Service via gRPC
  - Convert Protobuf responses back to JSON
- **Contract**: OpenAPI specification at `nexoan/contracts/rest/update_api.yaml`

#### Query API (Ballerina, Port 8081)
- **Purpose**: Handle entity queries and retrieval
- **Technology**: Ballerina REST service
- **Location**: `nexoan/query-api/`
- **Responsibilities**:
  - Accept query requests from clients
  - Support selective field retrieval (metadata, relationships, attributes)
  - Filter and search capabilities
  - Communicate with CRUD Service via gRPC
  - Return formatted JSON responses
- **Contract**: OpenAPI specification at `nexoan/contracts/rest/query_api.yaml`

#### Swagger UI
- **Purpose**: Interactive API documentation
- **Location**: `nexoan/swagger-ui/`
- **Serves**: OpenAPI specifications for Update and Query APIs

### 2. Service Layer (Business Logic)

#### CRUD Service (Go, gRPC, Port 50051)
Central orchestration service that manages all database interactions.

**Location**: `nexoan/crud-api/`

**Core Components**:

1. **gRPC Server** (`cmd/server/service.go`)
   - `CreateEntity(ctx, *pb.Entity) (*pb.Entity, error)`
     - Orchestrates entity creation across all databases
     - Saves metadata to MongoDB
     - Creates entity node in Neo4j
     - Handles relationships in Neo4j
     - Processes attributes for PostgreSQL
   
   - `ReadEntity(ctx, *pb.ReadEntityRequest) (*pb.Entity, error)`
     - Retrieves entity information from multiple databases
     - Supports selective field retrieval via output parameter
     - Assembles complete entity from distributed storage
   
   - `UpdateEntity(ctx, *pb.UpdateEntityRequest) (*pb.Entity, error)`
     - Updates entity information across databases
     - Handles partial updates
   
   - `DeleteEntity(ctx, *pb.EntityId) (*pb.Empty, error)`
     - Removes entity from all databases
     - Cascades deletion across MongoDB, Neo4j, and PostgreSQL

2. **Engine Layer** (`engine/`)
   - **AttributeProcessor** (`attribute_resolver.go`)
     - Processes entity attributes
     - Determines storage strategy
     - Handles time-based attribute values
     - Manages attribute schema evolution
   
   - **GraphMetadataManager** (`graph_metadata_manager.go`)
     - Manages graph metadata
     - Handles relationship metadata
   
   - **Type Inference** (`pkg/typeinference/inference.go`)
     - Automatically detects data types
     - Supports: int, float, string, bool, null
     - Special types: date, time, datetime
   
   - **Storage Inference** (`pkg/storageinference/inference.go`)
     - Determines optimal storage type
     - Types: tabular, graph, list, map, scalar

3. **Repository Layer** (`db/repository/`)
   - **MongoRepository** (`mongo/mongodb_client.go`, `mongo/metadata_handler.go`)
     - Handles metadata storage and retrieval
     - Connection management
     - CRUD operations for metadata
   
   - **Neo4jRepository** (`neo4j/neo4j_client.go`, `neo4j/graph_entity_handler.go`)
     - Manages entity nodes and relationships
     - Graph traversal operations
     - Cypher query execution
   
   - **PostgresRepository** (`postgres/postgres_client.go`, `postgres/data_handler.go`)
     - Handles attribute storage
     - Schema management
     - Time-series data operations

### 3. Database Layer

#### MongoDB (Port 27017)
- **Purpose**: Flexible metadata storage
- **Collections**: 
  - `metadata` - Production metadata
  - `metadata_test` - Test metadata
- **Data Model**: Document-based key-value pairs
- **Why MongoDB**: Schema-less structure ideal for dynamic metadata
- **Location**: `deployment/development/docker/mongodb/`

#### Neo4j (Port 7474 HTTP, 7687 Bolt)
- **Purpose**: Entity and relationship storage
- **Data Model**: 
  - **Nodes**: Entities with properties (id, kind_major, kind_minor, name, created, terminated)
  - **Relationships**: Directed edges with properties (id, name, startTime, endTime)
- **Why Neo4j**: Optimized for graph traversal and relationship queries
- **Location**: `deployment/development/docker/neo4j/`

#### PostgreSQL (Port 5432)
- **Purpose**: Time-based attribute storage
- **Schema**:
  - `attribute_schemas` - Attribute type definitions
  - `entity_attributes` - Entity-to-attribute mappings
  - `attr_*` - Dynamic tables for each attribute type
- **Why PostgreSQL**: ACID compliance, complex queries, time-series support
- **Location**: `deployment/development/docker/postgres/`

### 4. Supporting Services

#### Cleanup Service
- **Purpose**: Database cleanup for testing and maintenance
- **Technology**: Python
- **Trigger**: Docker Compose profile (`--profile cleanup`)
- **Operations**:
  - Clears PostgreSQL tables
  - Drops MongoDB collections
  - Removes Neo4j nodes and relationships
- **Usage**: `docker-compose --profile cleanup run --rm cleanup /app/cleanup.sh pre`

#### Backup/Restore Service
- **Purpose**: Data persistence and version management
- **Operations**:
  - Local backup creation for all databases
  - GitHub-based backup storage with versioning
  - Automated restore from GitHub releases
- **Scripts**: `deployment/development/init.sh`
- **Documentation**: See `docs/deployment/BACKUP_INTEGRATION.md`

---

## Data Model

### Entity Structure (Protobuf)

```protobuf
message Entity {
    string id = 1;                              // Unique identifier
    Kind kind = 2;                              // major/minor classification
    string created = 3;                         // Creation timestamp (ISO 8601)
    string terminated = 4;                      // Optional termination timestamp
    TimeBasedValue name = 5;                    // Entity name with temporal tracking
    map<string, google.protobuf.Any> metadata = 6;        // Flexible metadata
    map<string, TimeBasedValueList> attributes = 7;       // Time-based attributes
    map<string, Relationship> relationships = 8;          // Entity relationships
}

message Kind {
    string major = 1;                           // Primary classification
    string minor = 2;                           // Secondary classification
}

message TimeBasedValue {
    string startTime = 1;                       // Value valid from
    string endTime = 2;                         // Value valid until (empty = current)
    google.protobuf.Any value = 3;              // Actual value (any type)
}

message Relationship {
    string id = 1;                              // Relationship identifier
    string relatedEntityId = 2;                 // Target entity
    string name = 3;                            // Relationship type
    string startTime = 4;                       // Relationship valid from
    string endTime = 5;                         // Relationship valid until
    string direction = 6;                       // Relationship direction
}
```

### Storage Distribution Strategy

The entity data is strategically distributed across three databases:

**Example Entity:**
```json
{
  "id": "entity123",
  "kind": {"major": "Person", "minor": "Employee"},
  "name": "John Doe",
  "created": "2024-01-01T00:00:00Z",
  "metadata": {"department": "Engineering", "role": "Engineer"},
  "attributes": {"salary": [{"startTime": "2024-01", "value": 100000}]},
  "relationships": {"reports_to": "manager123"}
}
```

**Storage Distribution:**

```
┌──────────────────────────────────────────────────────────────┐
│ MongoDB - Metadata Storage                                   │
│ ┌──────────────────────────────────────────────────────────┐ │
│ │ Collection: metadata                                     │ │
│ │ {                                                        │ │
│ │   "_id": "entity123",                                    │ │
│ │   "metadata": {                                          │ │
│ │     "department": "Engineering",                         │ │
│ │     "role": "Engineer"                                   │ │
│ │   }                                                      │ │
│ │ }                                                        │ │
│ └──────────────────────────────────────────────────────────┘ │
└──────────────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────────────┐
│ Neo4j - Entity & Relationship Storage                        │
│ ┌──────────────────────────────────────────────────────────┐ │
│ │ Node:                                                    │ │
│ │ (entity123:Entity {                                      │ │
│ │   id: "entity123",                                       │ │
│ │   kind_major: "Person",                                  │ │
│ │   kind_minor: "Employee",                                │ │
│ │   name: "John Doe",                                      │ │
│ │   created: "2024-01-01T00:00:00Z"                        │ │
│ │ })                                                       │ │
│ │                                                          │ │
│ │ Relationship:                                            │ │
│ │ (entity123)-[:REPORTS_TO {                               │ │
│ │   id: "rel123",                                          │ │
│ │   startTime: "2024-01-01T00:00:00Z"                      │ │
│ │ }]->(manager123)                                         │ │
│ └──────────────────────────────────────────────────────────┘ │
└──────────────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────────────┐
│ PostgreSQL - Attribute Storage                               │
│ ┌──────────────────────────────────────────────────────────┐ │
│ │ Table: attribute_schemas                                 │ │
│ │   {kind_major: "Person", attr_name: "salary",            │ │
│ │    data_type: "int", storage_type: "scalar"}             │ │
│ │                                                          │ │
│ │ Table: entity_attributes                                 │ │
│ │   {entity_id: "entity123", attr_name: "salary"}          │ │
│ │                                                          │ │
│ │ Table: attr_Person_salary                                │ │
│ │   {entity_id: "entity123",                               │ │
│ │    start_time: "2024-01",                                │ │
│ │    end_time: NULL,                                       │ │
│ │    value: 100000}                                        │ │
│ └──────────────────────────────────────────────────────────┘ │
└──────────────────────────────────────────────────────────────┘
```

---

## Data Flow Sequences

### Create Entity Flow

```
┌────────┐         ┌────────────┐         ┌──────────────┐         ┌──────────┐
│ Client │         │ Update API │         │ CRUD Service │         │ Databases│
└───┬────┘         └─────┬──────┘         └──────┬───────┘         └────┬─────┘
    │                    │                       │                      │
    │ POST /entities     │                       │                      │
    │ (JSON payload)     │                       │                      │
    ├───────────────────>│                       │                      │
    │                    │                       │                      │
    │                    │ JSON to Protobuf      │                      │
    │                    │ conversion            │                      │
    │                    │                       │                      │
    │                    │ gRPC: CreateEntity    │                      │
    │                    ├──────────────────────>│                      │
    │                    │                       │                      │
    │                    │                       │ Save metadata        │
    │                    │                       ├─────────────────────>│
    │                    │                       │      (MongoDB)       │
    │                    │                       │                      │
    │                    │                       │ Create entity node   │
    │                    │                       ├─────────────────────>│
    │                    │                       │      (Neo4j)         │
    │                    │                       │                      │
    │                    │                       │ Create relationships │
    │                    │                       ├─────────────────────>│
    │                    │                       │      (Neo4j)         │
    │                    │                       │                      │
    │                    │                       │ Save attributes      │
    │                    │                       ├─────────────────────>│
    │                    │                       │    (PostgreSQL)      │
    │                    │                       │                      │
    │                    │    Entity (Protobuf)  │                      │
    │                    │<──────────────────────┤                      │
    │                    │                       │                      │
    │                    │ Protobuf to JSON      │                      │
    │                    │ conversion            │                      │
    │                    │                       │                      │
    │   201 Created      │                       │                      │
    │   (JSON response)  │                       │                      │
    │<───────────────────┤                       │                      │
    │                    │                       │                      │
```

### Read Entity Flow

```
┌────────┐         ┌───────────┐          ┌──────────────┐         ┌──────────┐
│ Client │         │ Query API │          │ CRUD Service │         │ Databases│
└───┬────┘         └─────┬─────┘          └──────┬───────┘         └────┬─────┘
    │                    │                       │                      │
    │ GET /entities/123  │                       │                      │
    │ ?output=metadata,  │                       │                      │
    │  relationships     │                       │                      │
    ├───────────────────>│                       │                      │
    │                    │                       │                      │
    │                    │ gRPC: ReadEntity      │                      │
    │                    ├──────────────────────>│                      │
    │                    │                       │                      │
    │                    │                       │ Get entity info      │
    │                    │                       ├─────────────────────>│
    │                    │                       │    (Neo4j - always)  │
    │                    │                       │                      │
    │                    │                       │ Get metadata         │
    │                    │                       ├─────────────────────>│
    │                    │                       │  (MongoDB - if req'd)│
    │                    │                       │                      │
    │                    │                       │ Get relationships    │
    │                    │                       ├─────────────────────>│
    │                    │                       │  (Neo4j - if req'd)  │
    │                    │                       │                      │
    │                    │                       │ Assemble entity      │
    │                    │                       │                      │
    │                    │    Entity (Protobuf)  │                      │
    │                    │<──────────────────────┤                      │
    │                    │                       │                      │
    │                    │ Protobuf to JSON      │                      │
    │                    │ conversion            │                      │
    │                    │                       │                      │
    │   200 OK           │                       │                      │
    │   (JSON response)  │                       │                      │
    │<───────────────────┤                       │                      │
    │                    │                       │                      │
```

---

## Type System

### Type Inference System

**Location**: `nexoan/crud-api/pkg/typeinference/`

**Primitive Types:**
- `int` - Whole numbers without decimal points
- `float` - Numbers with decimals or scientific notation
- `string` - Text data
- `bool` - True/false values
- `null` - Absence of value

**Special Types:**
- `date` - Calendar dates (YYYY-MM-DD, DD/MM/YYYY, etc.)
- `time` - Time of day (HH:MM:SS)
- `datetime` - Combined date and time with timezone (RFC3339)

**Type Inference Rules:**
1. Numbers with decimal points → `float`
2. Whole numbers → `int`
3. Text matching date patterns → `date`
4. Text matching time patterns → `time`
5. Text matching datetime patterns → `datetime`
6. Everything else → `string`

### Storage Type Inference

**Location**: `nexoan/crud-api/pkg/storageinference/`

**Storage Types:**
1. **Tabular** - Has `columns` and `rows` fields
   ```json
   {"columns": ["id", "name"], "rows": [[1, "John"], [2, "Jane"]]}
   ```

2. **Graph** - Has `nodes` and `edges` fields
   ```json
   {"nodes": [{"id": "n1"}], "edges": [{"source": "n1", "target": "n2"}]}
   ```

3. **List** - Has `items` array
   ```json
   {"items": [1, 2, 3, 4, 5]}
   ```

4. **Scalar** - Single field with primitive value
   ```json
   {"value": 42}
   ```

5. **Map** - Key-value pairs (default)
   ```json
   {"key1": "value1", "key2": "value2"}
   ```

---

## Communication Protocols

| Layer | Protocol | Format | Port |
|-------|----------|--------|------|
| Client ↔ Update API | HTTP/REST | JSON | 8080 |
| Client ↔ Query API | HTTP/REST | JSON | 8081 |
| APIs ↔ CRUD Service | gRPC | Protobuf | 50051 |
| CRUD ↔ MongoDB | MongoDB Wire Protocol | BSON | 27017 |
| CRUD ↔ Neo4j | Bolt Protocol | Cypher | 7687 |
| CRUD ↔ PostgreSQL | PostgreSQL Wire Protocol | SQL | 5432 |

---

## Network Architecture

**Docker Network**: `ldf-network` (bridge network)

All services run within the same Docker network:
- Container-based service discovery
- Internal communication via container names
- Health checks ensure proper startup sequencing
- Volume persistence for data storage

**Exposed Ports:**
- `8080` - Update API (external access)
- `8081` - Query API (external access)
- `50051` - CRUD Service (can be internal only)
- `27017` - MongoDB (development access)
- `7474/7687` - Neo4j (development access)
- `5432` - PostgreSQL (development access)

---

## Deployment

### Containerization
- **Technology**: Docker + Docker Compose
- **Orchestration File**: `docker-compose.yml`
- **Network**: `ldf-network` (bridge)
- **Volumes**: Persistent storage for all databases

### Health Checks
All services include health check configurations:
- MongoDB: `mongo --eval "db.adminCommand('ping')"`
- Neo4j: HTTP endpoint check on port 7474
- PostgreSQL: `pg_isready`
- CRUD Service: TCP check on port 50051
- Update/Query APIs: TCP checks on respective ports

### Dependency Management
Services start in proper order using Docker Compose `depends_on`:
```
Databases (MongoDB, Neo4j, PostgreSQL)
  ↓
CRUD Service (waits for all databases to be healthy)
  ↓
Update & Query APIs (wait for CRUD Service to be healthy)
```

### Docker Compose Profiles
- **Default**: Runs all core services
- **cleanup**: Database cleanup service (separate profile)

---

## Technology Stack

| Component | Technology | Language | Purpose |
|-----------|-----------|----------|---------|
| Update API | Ballerina | Ballerina | REST API for mutations |
| Query API | Ballerina | Ballerina | REST API for queries |
| CRUD Service | Go + gRPC | Go | Business logic orchestration |
| MongoDB | MongoDB 5.0+ | - | Metadata storage |
| Neo4j | Neo4j 5.x | - | Graph storage |
| PostgreSQL | PostgreSQL 14+ | - | Attribute storage |
| Protobuf | Protocol Buffers | - | Service communication |
| Docker | Docker + Compose | - | Containerization |
| Testing | Go test, Bal test, Python | Multiple | Unit & E2E tests |

---

## Key Features

### 1. Multi-Database Strategy
- **Optimized Storage**: Each database serves its best use case
- **Data Separation**: Clear boundaries between metadata, entities, and attributes
- **Scalability**: Independent scaling of each database

### 2. Time-Based Data Support
- **Temporal Attributes**: Track attribute values over time
- **Temporal Relationships**: Time-bound entity relationships
- **Historical Queries**: Query data at specific points in time (activeAt parameter)

### 3. Type Inference
- **Automatic Detection**: No manual type specification required
- **Rich Type System**: Supports primitives and special types
- **Storage Optimization**: Determines optimal storage based on data structure

### 4. Schema Evolution
- **Dynamic Schemas**: PostgreSQL tables created on-demand
- **Attribute Flexibility**: New attributes don't require migrations
- **Kind-Based Organization**: Attributes organized by entity kind

### 5. Graph Relationships
- **Native Graph Storage**: Neo4j for optimal relationship queries
- **Bi-directional Support**: Forward and reverse relationship traversal
- **Relationship Properties**: Rich metadata on relationships

### 6. Backup & Restore
- **Multi-Database Backup**: Coordinated backups across all databases
- **Version Management**: GitHub-based version control
- **One-Command Restore**: Simple restoration from any version

### 7. API Contract-First
- **OpenAPI Specifications**: APIs defined before implementation
- **Code Generation**: Service scaffolding from contracts
- **Documentation**: Swagger UI for interactive API docs

---

## Design Decisions

### Why Multiple Databases?

**MongoDB for Metadata:**
- Schema-less structure for flexible metadata
- Fast key-value lookups
- Easy to add new metadata fields

**Neo4j for Entities & Relationships:**
- Native graph traversal
- Efficient relationship queries
- Cypher query language optimized for graphs

**PostgreSQL for Attributes:**
- ACID compliance for critical data
- Complex time-based queries
- Strong typing and constraints
- Efficient indexing for time-series data

### Why gRPC for Internal Communication?

- **Performance**: Binary protocol is faster than JSON
- **Type Safety**: Protobuf provides strong contracts
- **Code Generation**: Client/server stubs auto-generated
- **Streaming**: Support for streaming if needed in future

### Why Ballerina for APIs?

- **Cloud-Native**: Built for cloud and microservices
- **Type Safety**: Strong typing with type inference
- **OpenAPI Integration**: Native OpenAPI support
- **Network-Aware**: Built-in support for REST, gRPC, etc.

---

## Future Enhancements

Based on TODOs found in the codebase:

1. **Connection Pooling** - Optimize database connections
2. **Caching Layer** - Redis for frequently accessed data
3. **Query Optimization** - Indexed queries and batch operations
4. **Error Recovery** - Rollback mechanisms and retry logic
5. **Transaction Support** - Distributed transactions across databases
6. **GraphQL API** - Alternative query interface
7. **Event Streaming** - Kafka integration for event-driven architecture
8. **Observability** - Distributed tracing and metrics

---

## Related Documentation

- [How It Works](../how_it_works.md) - Detailed data flow documentation
- [Data Types](../datatype.md) - Type inference system details
- [Storage Types](../storage.md) - Storage type inference details
- [Backup Integration](../deployment/BACKUP_INTEGRATION.md) - Backup and restore guide
- [Core API](../../nexoan/crud-api/README.md) - Core API documentation
- [Ingestion API](../../nexoan/update-api/README.md) - Ingestion API documentation
- [Read API](../../nexoan/query-api/README.md) - Read API documentation

---

## Quick Reference

### Service Endpoints

```bash
# Update API
POST   http://localhost:8080/entities          # Create entity
GET    http://localhost:8080/entities/{id}     # Read entity
PUT    http://localhost:8080/entities/{id}     # Update entity
DELETE http://localhost:8080/entities/{id}     # Delete entity

# Query API
GET    http://localhost:8081/v1/entities/{id}/metadata       # Get metadata
GET    http://localhost:8081/v1/entities/{id}/relationships  # Get relationships
GET    http://localhost:8081/v1/entities/{id}/attributes     # Get attributes
```

### Database Connections

```bash
# MongoDB
mongodb://admin:admin123@localhost:27017/nexoan?authSource=admin

# Neo4j
bolt://neo4j:neo4j123@localhost:7687
http://localhost:7474

# PostgreSQL
postgresql://postgres:postgres@localhost:5432/nexoan
```

### Docker Commands

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop all services
docker-compose down

# Clean databases
docker-compose --profile cleanup run --rm cleanup /app/cleanup.sh pre

# Rebuild services
docker-compose build
```

---

**Document Version**: 1.0  
**Last Updated**: October 2024  
**Maintained By**: Nexoan Development Team

