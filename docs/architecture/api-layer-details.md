# API Layer - Detailed Architecture

This document provides comprehensive details about the Update API and Query API layers of the Nexoan system.

---

## Overview

The API Layer consists of two Ballerina-based REST services that provide external access to the Nexoan system:
- **Update API**: Handles entity mutations (CREATE, UPDATE, DELETE)
- **Query API**: Handles entity queries and retrieval

Both APIs act as translation layers between external HTTP/JSON clients and the internal gRPC/Protobuf CRUD service.

---

## Update API

### Overview

**Location**: `nexoan/update-api/`  
**Language**: Ballerina  
**Protocol**: HTTP/REST + JSON  
**Port**: 8080  
**Contract**: `nexoan/contracts/rest/update_api.yaml`

### Directory Structure

```
nexoan/update-api/
├── update_api_service.bal      # Main service implementation
├── types_v1_pb.bal              # Generated protobuf types for Ballerina
├── Ballerina.toml               # Package configuration
├── Dependencies.toml            # Dependency versions
├── env.template                 # Environment variable template
├── utils/                       # Utility functions
├── tests/
│   └── service_test.bal         # Service tests
└── target/                      # Build outputs
```

### Service Implementation

**File**: `update_api_service.bal`

The service exposes REST endpoints following OpenAPI specification:

```ballerina
service /entities on new http:Listener(8080) {
    
    // CREATE: POST /entities
    resource function post .(@http:Payload json payload) 
        returns json|error {
        // Implementation
    }
    
    // READ: GET /entities/{id}
    resource function get [string id]() 
        returns json|error {
        // Implementation
    }
    
    // UPDATE: PUT /entities/{id}
    resource function put [string id](@http:Payload json payload) 
        returns json|error {
        // Implementation
    }
    
    // DELETE: DELETE /entities/{id}
    resource function delete [string id]() 
        returns http:Ok|error {
        // Implementation
    }
}
```

### Request/Response Flow

#### CREATE Entity

**Request**:
```bash
POST /entities
Content-Type: application/json

{
  "id": "entity123",
  "kind": {
    "major": "Person",
    "minor": "Employee"
  },
  "created": "2024-01-01T00:00:00Z",
  "name": {
    "startTime": "2024-01-01T00:00:00Z",
    "endTime": "",
    "value": "John Doe"
  },
  "metadata": [
    {"key": "department", "value": "Engineering"},
    {"key": "role", "value": "Engineer"}
  ],
  "attributes": [
    {
      "key": "salary",
      "value": {
        "values": [
          {
            "startTime": "2024-01-01T00:00:00Z",
            "endTime": "",
            "value": 100000
          }
        ]
      }
    }
  ],
  "relationships": [
    {
      "key": "reports_to",
      "value": {
        "id": "rel123",
        "relatedEntityId": "manager456",
        "name": "reports_to",
        "startTime": "2024-01-01T00:00:00Z",
        "endTime": "",
        "direction": "outgoing"
      }
    }
  ]
}
```

**Response**:
```json
{
  "id": "entity123",
  "kind": {
    "major": "Person",
    "minor": "Employee"
  },
  "created": "2024-01-01T00:00:00Z",
  "name": {
    "startTime": "2024-01-01T00:00:00Z",
    "value": "John Doe"
  }
}
```

**Status Codes**:
- `201 Created`: Entity successfully created
- `400 Bad Request`: Invalid JSON or missing required fields
- `409 Conflict`: Entity with ID already exists
- `500 Internal Server Error`: CRUD service error

#### READ Entity

**Request**:
```bash
GET /entities/entity123
```

**Response**:
```json
{
  "id": "entity123",
  "kind": {
    "major": "Person",
    "minor": "Employee"
  },
  "created": "2024-01-01T00:00:00Z",
  "name": {
    "startTime": "2024-01-01T00:00:00Z",
    "value": "John Doe"
  },
  "metadata": [
    {"key": "department", "value": "Engineering"}
  ],
  "attributes": [],
  "relationships": []
}
```

**Status Codes**:
- `200 OK`: Entity found and returned
- `404 Not Found`: Entity doesn't exist
- `500 Internal Server Error`: CRUD service error

#### UPDATE Entity

**Request**:
```bash
PUT /entities/entity123
Content-Type: application/json

{
  "id": "entity123",
  "kind": {
    "major": "Person",
    "minor": "Employee"
  },
  "metadata": [
    {"key": "department", "value": "Sales"}
  ]
}
```

**Response**:
```json
{
  "id": "entity123",
  "kind": {
    "major": "Person",
    "minor": "Employee"
  },
  "metadata": [
    {"key": "department", "value": "Sales"}
  ]
}
```

**Status Codes**:
- `200 OK`: Entity successfully updated
- `404 Not Found`: Entity doesn't exist
- `400 Bad Request`: Invalid update data
- `500 Internal Server Error`: CRUD service error

#### DELETE Entity

**Request**:
```bash
DELETE /entities/entity123
```

**Response**:
```
204 No Content
```

**Status Codes**:
- `204 No Content`: Entity successfully deleted
- `404 Not Found`: Entity doesn't exist
- `500 Internal Server Error`: CRUD service error

### JSON to Protobuf Conversion

The Update API performs complex conversion between JSON and Protobuf formats:

**Key Conversions**:

1. **Metadata Conversion**:
```ballerina
// JSON array to Protobuf map
record {|string key; string value;|}[] metadata = jsonPayload.metadata;
map<pbAny:Any> metadataMap = {};

foreach var item in metadata {
    pbAny:Any packedValue = check pbAny:pack(item.value);
    metadataMap[item.key] = packedValue;
}
```

2. **Attributes Conversion**:
```ballerina
// JSON to Protobuf TimeBasedValueList
record {|string key; AttributeValue value;|}[] attributes = jsonPayload.attributes;
map<TimeBasedValueList> attributesMap = {};

foreach var attr in attributes {
    TimeBasedValue[] values = [];
    foreach var val in attr.value.values {
        pbAny:Any packedValue = check pbAny:pack(val.value);
        values.push({
            startTime: val.startTime,
            endTime: val.endTime,
            value: packedValue
        });
    }
    attributesMap[attr.key] = {values: values};
}
```

3. **Relationships Conversion**:
```ballerina
// JSON to Protobuf Relationship map
record {|string key; RelationshipValue value;|}[] relationships = jsonPayload.relationships;
map<Relationship> relationshipsMap = {};

foreach var rel in relationships {
    relationshipsMap[rel.key] = {
        id: rel.value.id,
        relatedEntityId: rel.value.relatedEntityId,
        name: rel.value.name,
        startTime: rel.value.startTime,
        endTime: rel.value.endTime,
        direction: rel.value.direction
    };
}
```

### gRPC Client Configuration

**Connection Setup**:
```ballerina
final grpc:Client crudClient = check new (
    string `${CRUD_SERVICE_URL}`,
    {
        timeout: 30,  // 30 second timeout
        retryConfiguration: {
            maxCount: 3,
            interval: 1,
            backoff: 2
        }
    }
);
```

**Environment Variables**:
```bash
CRUD_SERVICE_HOST=localhost
CRUD_SERVICE_PORT=50051
UPDATE_SERVICE_HOST=0.0.0.0
UPDATE_SERVICE_PORT=8080
```

### Error Handling

**Error Types**:
1. **Validation Errors**: Invalid JSON structure
2. **Conversion Errors**: Failed JSON ↔ Protobuf conversion
3. **gRPC Errors**: CRUD service communication failures
4. **Business Logic Errors**: Entity already exists, not found, etc.

**Error Response Format**:
```json
{
  "error": {
    "code": "ENTITY_NOT_FOUND",
    "message": "Entity with ID 'entity123' not found",
    "details": {
      "entityId": "entity123"
    }
  }
}
```

### Testing

**Test File**: `tests/service_test.bal`

**Test Coverage**:
- Entity creation with all fields
- Entity creation with minimal fields
- Entity read by ID
- Entity update
- Entity deletion
- Error scenarios (invalid JSON, entity not found, etc.)

**Running Tests**:
```bash
cd nexoan/update-api
bal test
```

---

## Query API

### Overview

**Location**: `nexoan/query-api/`  
**Language**: Ballerina  
**Protocol**: HTTP/REST + JSON  
**Port**: 8081  
**Contract**: `nexoan/contracts/rest/query_api.yaml`

### Directory Structure

```
nexoan/query-api/
├── query_api_service.bal        # Main service implementation
├── types_v1_pb.bal               # Generated protobuf types for Ballerina
├── types.bal                     # Type definitions
├── Ballerina.toml                # Package configuration
├── Dependencies.toml             # Dependency versions
├── env.template                  # Environment variable template
├── tests/
│   └── query_api_service_test.bal  # Service tests
└── target/                       # Build outputs
```

### Service Implementation

**File**: `query_api_service.bal`

The Query API provides specialized endpoints for retrieving entity data:

```ballerina
service /v1/entities on new http:Listener(8081) {
    
    // Get entity metadata
    resource function get [string id]/metadata() 
        returns json|error {
        // Implementation
    }
    
    // Get entity relationships
    resource function get [string id]/relationships(
        string? name = (),
        string? direction = (),
        string? relatedEntityId = (),
        string? activeAt = ()
    ) returns json|error {
        // Implementation
    }
    
    // Get entity attributes
    resource function get [string id]/attributes(
        string? name = (),
        string? activeAt = ()
    ) returns json|error {
        // Implementation
    }
    
    // Get complete entity with selective fields
    resource function get [string id](
        string[]? output = ()
    ) returns json|error {
        // Implementation
    }
}
```

### Query Operations

#### Get Metadata

**Request**:
```bash
GET /v1/entities/entity123/metadata
```

**Response**:
```json
{
  "metadata": {
    "department": "Engineering",
    "role": "Engineer",
    "employeeId": "EMP-123"
  }
}
```

**Use Case**: Retrieve flexible key-value metadata for an entity

#### Get Relationships

**Request**:
```bash
GET /v1/entities/entity123/relationships?name=reports_to&direction=outgoing
```

**Query Parameters**:
- `name`: Filter by relationship type
- `direction`: `outgoing` or `incoming`
- `relatedEntityId`: Filter by related entity
- `activeAt`: Temporal query (ISO 8601 timestamp)

**Response**:
```json
{
  "relationships": [
    {
      "id": "rel123",
      "name": "reports_to",
      "relatedEntityId": "manager456",
      "startTime": "2024-01-01T00:00:00Z",
      "endTime": null,
      "direction": "outgoing"
    }
  ]
}
```

**Use Case**: Graph traversal, finding related entities

#### Get Attributes

**Request**:
```bash
GET /v1/entities/entity123/attributes?name=salary&activeAt=2024-06-01T00:00:00Z
```

**Query Parameters**:
- `name`: Filter by attribute name
- `activeAt`: Get attribute value at specific time

**Response**:
```json
{
  "attributes": {
    "salary": {
      "values": [
        {
          "startTime": "2024-01-01T00:00:00Z",
          "endTime": "2024-06-30T23:59:59Z",
          "value": 100000
        },
        {
          "startTime": "2024-07-01T00:00:00Z",
          "endTime": null,
          "value": 110000
        }
      ]
    }
  }
}
```

**Use Case**: Time-series data, historical attribute values

#### Get Entity with Selective Fields

**Request**:
```bash
GET /v1/entities/entity123?output=metadata,relationships
```

**Query Parameters**:
- `output`: Comma-separated list of fields to retrieve
  - Options: `metadata`, `relationships`, `attributes`
  - If omitted: returns only basic entity info

**Response**:
```json
{
  "id": "entity123",
  "kind": {
    "major": "Person",
    "minor": "Employee"
  },
  "name": {
    "value": "John Doe",
    "startTime": "2024-01-01T00:00:00Z"
  },
  "created": "2024-01-01T00:00:00Z",
  "metadata": {
    "department": "Engineering"
  },
  "relationships": [
    {
      "name": "reports_to",
      "relatedEntityId": "manager456"
    }
  ]
}
```

**Use Case**: Optimized queries, reduce payload size

### Temporal Queries

The Query API supports temporal queries using the `activeAt` parameter:

**Example**: Get employee's salary on specific date
```bash
GET /v1/entities/entity123/attributes?name=salary&activeAt=2024-03-15T00:00:00Z
```

**Backend Filter**:
```sql
WHERE start_time <= '2024-03-15T00:00:00Z'
  AND (end_time IS NULL OR end_time >= '2024-03-15T00:00:00Z')
```

This returns only the attribute value that was active on March 15, 2024.

### gRPC Client Configuration

**Connection Setup**:
```ballerina
final grpc:Client crudClient = check new (
    string `${CRUD_SERVICE_URL}`,
    {
        timeout: 30,
        retryConfiguration: {
            maxCount: 3,
            interval: 1,
            backoff: 2
        }
    }
);
```

**Environment Variables**:
```bash
CRUD_SERVICE_HOST=localhost
CRUD_SERVICE_PORT=50051
QUERY_SERVICE_HOST=0.0.0.0
QUERY_SERVICE_PORT=8081
```

### Performance Optimization

**Selective Field Retrieval**:
- Only fetch requested fields from CRUD service
- Reduces database load
- Reduces network bandwidth
- Faster response times

**Example**:
```
Query: GET /v1/entities/entity123?output=metadata

Instead of:
  ├─ MongoDB (metadata)         ✓ Retrieved
  ├─ Neo4j (entity info)        ✓ Retrieved
  ├─ Neo4j (relationships)      ✗ Skipped
  └─ PostgreSQL (attributes)    ✗ Skipped
```

### Testing

**Test File**: `tests/query_api_service_test.bal`

**Test Coverage**:
- Get metadata
- Get relationships with filters
- Get attributes with temporal queries
- Get entity with selective fields
- Error scenarios

**Running Tests**:
```bash
cd nexoan/query-api
bal test
```

---

## OpenAPI Contracts

### Update API Contract

**File**: `nexoan/contracts/rest/update_api.yaml`

Defines:
- Entity schema
- Request/response structures
- HTTP methods and paths
- Status codes
- Error responses

**Code Generation**:
```bash
bal openapi -i ../contracts/rest/update_api.yaml --mode service
```

### Query API Contract

**File**: `nexoan/contracts/rest/query_api.yaml`

Defines:
- Query parameters
- Filter options
- Response structures
- Pagination (if implemented)

**Code Generation**:
```bash
bal openapi -i ../contracts/rest/query_api.yaml --mode service
```

---

## Swagger UI

### Overview

**Location**: `nexoan/swagger-ui/`  
**Purpose**: Interactive API documentation

### Features

- View all API endpoints
- Test API calls directly from browser
- See request/response examples
- Understand data models

### Access

```bash
# Start Swagger UI
cd nexoan/swagger-ui
python3 serve.py

# Open browser
http://localhost:8082
```

### Configuration

The Swagger UI serves the OpenAPI specifications from:
- `nexoan/contracts/rest/update_api.yaml`
- `nexoan/contracts/rest/query_api.yaml`

---

## Common Patterns

### 1. Request Validation

Both APIs validate incoming requests:
```ballerina
// Validate required fields
if (payload.id is ()) {
    return error("Entity ID is required");
}

if (payload.kind is () || payload.kind.major is ()) {
    return error("Entity kind.major is required");
}
```

### 2. Error Response Formatting

Consistent error response structure:
```ballerina
function handleError(error err) returns http:InternalServerError {
    return {
        body: {
            error: {
                message: err.message(),
                details: err.detail()
            }
        }
    };
}
```

### 3. gRPC Communication

Standard pattern for calling CRUD service:
```ballerina
// Create gRPC request
ReadEntityRequest request = {
    entity: {
        id: entityId,
        kind: {},
        // ... minimal entity
    },
    output: ["metadata", "relationships"]
};

// Call CRUD service
Entity|grpc:Error result = crudClient->ReadEntity(request);

// Handle response
if (result is grpc:Error) {
    return handleGrpcError(result);
}
return convertToJson(result);
```

---

## Security Considerations

### Current State

- No authentication/authorization (development mode)
- All endpoints publicly accessible
- No rate limiting

### Future Enhancements

1. **Authentication**:
   - JWT token-based authentication
   - OAuth 2.0 integration
   - API key support

2. **Authorization**:
   - Role-based access control (RBAC)
   - Entity-level permissions
   - Field-level access control

3. **Security Headers**:
   - CORS configuration
   - HTTPS enforcement
   - Security headers (CSP, HSTS, etc.)

4. **Rate Limiting**:
   - Request throttling
   - IP-based limits
   - User-based quotas

---

## Monitoring and Observability

### Logging

Both APIs log:
- Incoming requests (method, path, params)
- gRPC calls to CRUD service
- Response status codes
- Errors with stack traces

### Metrics (Planned)

- Request count by endpoint
- Response time percentiles
- Error rate
- Active connections
- gRPC call latency

### Tracing (Planned)

- Distributed tracing with OpenTelemetry
- Trace request flow: Client → API → CRUD → Database
- Identify bottlenecks

---

## Best Practices

### 1. Use Selective Retrieval

❌ **Don't**:
```bash
GET /entities/entity123
# Returns everything (slow, large payload)
```

✅ **Do**:
```bash
GET /v1/entities/entity123?output=metadata
# Returns only what you need (fast, small payload)
```

### 2. Use Specific Query Endpoints

❌ **Don't**:
```bash
GET /entities/entity123  # Get full entity, then filter relationships in client
```

✅ **Do**:
```bash
GET /v1/entities/entity123/relationships?name=reports_to
# Server-side filtering, better performance
```

### 3. Use Temporal Queries

❌ **Don't**:
```bash
GET /v1/entities/entity123/attributes?name=salary
# Returns all values, filter in client
```

✅ **Do**:
```bash
GET /v1/entities/entity123/attributes?name=salary&activeAt=2024-03-15T00:00:00Z
# Returns only relevant value
```

---

## Related Documentation

- [Main Architecture Overview](./overview.md)
- [CRUD Service Details](./crud-service-details.md)
- [Architecture Diagrams](./diagrams.md)
- [Update API README](../../nexoan/update-api/README.md)
- [Query API README](../../nexoan/query-api/README.md)

---

**Document Version**: 1.0  
**Last Updated**: October 2024  
**Component**: API Layer

