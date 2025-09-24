# Nexoan

> 💡 **Note (α)**  
> Name needs to be proposed, voted and finalized. 

## 🚀 Running Services

### 1. Run CRUD API Service
-Read about running the [CRUD Service](nexoan/crud-api/README.md)

### 2. Run Query API Serivce
-Read about running the [Query API](nexoan/query-api/README.md)

### 3. Run Update API Service
-Read about running the [Update API](nexoan/update-api/README.md)

### 4. Run Swagger-UI  
-Read about running the [Swagger UI](nexoan/swagger-ui/README.md)

### 5. Database Cleanup Service
The cleanup service provides a way to clean all databases (PostgreSQL, MongoDB, Neo4j) before and after running tests or services.

**Usage:**
```bash
# Clean databases before starting services
docker-compose --profile cleanup run --rm cleanup /app/cleanup.sh pre

# Clean databases after services complete
docker-compose --profile cleanup run --rm cleanup /app/cleanup.sh post

# Clean databases anytime you need
docker-compose --profile cleanup run --rm cleanup /app/cleanup.sh pre
```

**What it cleans:**
- **PostgreSQL**: `attribute_schemas`, `entity_attributes`, and all `attr_*` tables
- **MongoDB**: `metadata` and `metadata_test` collections  
- **Neo4j**: All nodes and relationships

**Note**: The cleanup service uses the `cleanup` profile, so it won't start automatically with `docker-compose up`.

### 6. Database Backup and Restore
The system provides comprehensive backup and restore capabilities for all databases.

**Local Backup Management:**
```bash
# Create backups
./deployment/development/init.sh backup_mongodb
./deployment/development/init.sh backup_postgres
./deployment/development/init.sh backup_neo4j

# Restore from local backups
./deployment/development/init.sh restore_mongodb
./deployment/development/init.sh restore_postgres
./deployment/development/init.sh restore_neo4j
```

**GitHub Integration:**
```bash
# Restore from GitHub releases
./deployment/development/init.sh restore_from_github 0.0.1
./deployment/development/init.sh list_github_versions
```

For detailed backup and restore documentation, see [Backup Integration Guide](docs/deployment/BACKUP_INTEGRATION.md).

---

## Run a sample query with CURL

### Update API

**Create**

```bash
curl -X POST http://localhost:8080/entities \
-H "Content-Type: application/json" \
-d '{
  "id": "12345",
  "kind": {
    "major": "example",
    "minor": "test"
  },
  "created": "2024-03-17T10:00:00Z",
  "terminated": "",
  "name": {
    "startTime": "2024-03-17T10:00:00Z",
    "endTime": "",
    "value": {
      "typeUrl": "type.googleapis.com/google.protobuf.StringValue",
      "value": "entity-name"
    }
  },
  "metadata": [
    {"key": "owner", "value": "test-user"},
    {"key": "version", "value": "1.0"},
    {"key": "developer", "value": "V8A"}
  ],
  "attributes": [],
  "relationships": []
}'
```

**Read**

```bash
curl -X GET http://localhost:8080/entities/12345
```

**Update**

> TODO: The update creates a new record and that's a bug, please fix it. 

```bash
curl -X PUT http://localhost:8080/entities/12345 \
  -H "Content-Type: application/json" \
  -d '{
    "id": "12345",
    "kind": {
      "major": "example",
      "minor": "test"
    },
    "created": "2024-03-18T00:00:00Z",
    "name": {
      "startTime": "2024-03-18T00:00:00Z",
      "value": "entity-name"
    },
    "metadata": [
      {"key": "version", "value": "5.0"}
    ]
  }'
```

**Delete**

```bash
curl -X DELETE http://localhost:8080/entities/12345
```

### Query API 

**Retrieve Metadata**

```bash
curl -X GET "http://localhost:8081/v1/entities/12345/metadata"
```

## Run E2E Tests

Make sure the CRUD server and the API server are running. 

Note when making a call to ReadEntity, the ReadEntityRequest must be in the following format (output can be one or more of metadata, relationships, attributes):

ReadEntityRequest readEntityRequest = {
    entity: {
        id: entityId,
        kind: {},
        created: "",
        terminated: "",
        name: {
            startTime: "",
            endTime: "",
            value: check pbAny:pack("")
        },
        metadata: [],
        attributes: [],
        relationships: []
    },
    output: ["relationships"]
};

### Run Update API Tests

```bash
cd nexoan/tests/e2e
python basic_crud_tests.py
```

### Run Query API Tests

```bash
cd nexoan/tests/e2e
python basic_query_tests.py
```

## Implementation Progress

[Track Progress](https://github.com/LDFLK/nexoan/issues/29)
