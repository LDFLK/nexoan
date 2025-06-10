# Nexoan

> 💡 **Note (α)**  
> Name needs to be proposed, voted and finalized. 

## 🚀 Running Services

### 1. Run CRUD API Service
-Read about running the [CRUD Service](design/crud-api/README.md)

### 2. Run Query API Serivce
-Read about running the [Query API](design/query-api/README.md)

### 3. Run Update API Service
-Read about running the [Update API](design/update-api/README.md)

### 4. Run Swagger-UI  
-Read about running the [Swagger UI](design/swagger-ui/README.md)

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
    id: entityId,
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
cd design/tests/e2e
python basic_crud_tests.py
```

### Run Query API Tests

```bash
cd design/tests/e2e
python basic_query_tests.py
```

## Implementation Progress

[Track Progress](https://github.com/LDFLK/nexoan/issues/29)