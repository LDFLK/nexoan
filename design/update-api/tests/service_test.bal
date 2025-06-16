import ballerina/io;
import ballerina/test;
import ballerina/protobuf.types.'any as pbAny;
import ballerina/http;
import ballerina/os;

// Get environment variables without fallback values
string testCrudHostname = os:getEnv("CRUD_SERVICE_HOST");
string testCrudPort = os:getEnv("CRUD_SERVICE_PORT");
string testUpdateHostname = os:getEnv("UPDATE_SERVICE_HOST");
string testUpdatePort = os:getEnv("UPDATE_SERVICE_PORT");

// Try to get complete CRUD service URL from environment variable first
string? crudServiceUrlEnv = os:getEnv("CRUD_SERVICE_URL");
string testCrudServiceUrl = crudServiceUrlEnv ?: "http://" + testCrudHostname + ":" + testCrudPort;

// Construct URLs using string concatenation
string testUpdateServiceUrl = "http://" + testUpdateHostname + ":" + testUpdatePort;

// Before Suite Function
@test:BeforeSuite
function beforeSuiteFunc() {
    io:println("I'm the before suite function!");
    io:println("CRUD Service URL: " + testCrudServiceUrl);
    io:println("Update Service URL: " + testUpdateServiceUrl);
}

// After Suite Function
@test:AfterSuite
function afterSuiteFunc() {
    io:println("I'm the after suite function!");
}

// Helper function to unpack Any values to strings
function unwrapAny(pbAny:Any anyValue) returns string|error {
    return pbAny:unpack(anyValue, string);
}

// Helper function to convert JSON to protobuf Any value
function jsonToAny(json data) returns pbAny:Any|error {
    if data is int {
        // For integer values
        map<json> structMap = {
            "value": data
        };
        return pbAny:pack(structMap);
    } else if data is float {
        // For float values
        map<json> structMap = {
            "value": data
        };
        return pbAny:pack(structMap);
    } else if data is string {
        // For string values
        map<json> structMap = {
            "value": data
        };
        return pbAny:pack(structMap);
    } else if data is boolean {
        // For boolean values
        map<json> structMap = {
            "value": data
        };
        return pbAny:pack(structMap);
    } else if data is () {
        // For null values
        map<json> structMap = {
            "null_value": ()
        };
        return pbAny:pack(structMap);
    } else if data is json[] {
        // For arrays, wrap in a list_value structure
        map<json> structMap = {
            "values": data
        };
        return pbAny:pack(structMap);
    } else if data is map<json> {
        // For objects, just pack them directly
        return pbAny:pack(data);
    } else {
        return error("Unsupported data type: " + data.toString());
    }
}

@test:Config {}
function testMetadataHandling() returns error? {
    // Initialize the client
    CrudServiceClient ep = check new (testCrudServiceUrl);
    
    // Test data setup
    string testId = "test-entity-1";
    string expectedValue1 = "value1";
    string expectedValue2 = "value2";
    
    // Create the metadata array
    record {| string key; pbAny:Any value; |}[] metadataArray = [];

    // Pack string values into protobuf.Any directly
    pbAny:Any packedValue1 = check pbAny:pack(expectedValue1);
    pbAny:Any packedValue2 = check pbAny:pack(expectedValue2);

    // Add packed values to the metadata array
    metadataArray.push({key: "key1", value: packedValue1});
    metadataArray.push({key: "key2", value: packedValue2});

    // Create entity request
    Entity createEntityRequest = {
        id: testId,
        kind: {
            major: "test",
            minor: "metadata"
        },
        created: "2023-01-01",
        terminated: "",
        name: {
            startTime: "2023-01-01",
            endTime: "",
            value: check pbAny:pack("test-entity")
        },
        metadata: metadataArray
    };

    // Create entity
    Entity createEntityResponse = check ep->CreateEntity(createEntityRequest);
    io:println("Entity created with ID: " + createEntityResponse.id);
    io:println("Created entity metadata: ", createEntityResponse.metadata);
    
    // Read entity
    ReadEntityRequest readEntityRequest = {
        entity: {
            id: testId,
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
        output: ["metadata"]
    };
    io:println("ReadEntityRequest: ", readEntityRequest);
    Entity readEntityResponse = check ep->ReadEntity(readEntityRequest);
    io:println("Entity retrieved, verifying data...");
    io:println("Retrieved entity: ", readEntityResponse);
    io:println("Retrieved entity metadata: ", readEntityResponse.metadata);
    
    // Verify metadata values
    map<string> actualValues = {};
    foreach var item in readEntityResponse.metadata {
        string|error unwrapped = unwrapAny(item.value);
        if unwrapped is string {
            actualValues[item.key] = unwrapped.trim();
        } else {
            test:assertFail("Failed to unpack metadata value for key: " + item.key);
        }
    }
    
    // Assert the values match
    test:assertEquals(actualValues["key1"], expectedValue1, "Metadata value for key1 doesn't match");
    test:assertEquals(actualValues["key2"], expectedValue2, "Metadata value for key2 doesn't match");
    
    // Clean up
    Empty _ = check ep->DeleteEntity({id: testId});
    Empty _ = check ep->DeleteEntity({id: testId});
    io:println("Test entity deleted");
    
    return;
}

// TODO: Re-enable once the Result type response handling is added
// See: https://github.com/LDFLK/nexoan/issues/23
@test:Config {
    enable: false
}
function testMetadataUnpackError() returns error? {
    // Test case to verify handling of non-existent entities
    CrudServiceClient ep = check new (testCrudServiceUrl);
    
    // Try to read a non-existent entity
    ReadEntityRequest readEntityRequest = {
        entity: {
            id: "non-existent-entity",
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
        output: ["metadata"]
    };
    Entity|error response = ep->ReadEntity(readEntityRequest);
    
    // Assert that we get an error for non-existent entity
    test:assertTrue(response is error, "Expected error for non-existent entity");
    
    return;
}

@test:Config {}
function testMetadataUpdating() returns error? {
    // Initialize the client
    CrudServiceClient ep = check new (testCrudServiceUrl);
    
    // Test data setup
    string testId = "test-entity-update";
    
    // Initial metadata values
    string initialValue1 = "initial-value1";
    string initialValue2 = "initial-value2";
    
    // Updated metadata values
    string updatedValue1 = "updated-value1";
    string updatedValue2 = "updated-value2";
    string newValue3 = "new-value3";
    
    // Create the initial metadata array
    record {| string key; pbAny:Any value; |}[] initialMetadataArray = [];
    pbAny:Any packedInitialValue1 = check pbAny:pack(initialValue1);
    pbAny:Any packedInitialValue2 = check pbAny:pack(initialValue2);
    initialMetadataArray.push({key: "key1", value: packedInitialValue1});
    initialMetadataArray.push({key: "key2", value: packedInitialValue2});

    // Create initial entity request
    Entity createEntityRequest = {
        id: testId,
        kind: {
            major: "test",
            minor: "update-metadata"
        },
        created: "2023-01-01",
        terminated: "",
        name: {
            startTime: "2023-01-01",
            endTime: "",
            value: check pbAny:pack("test-update-entity")
        },
        metadata: initialMetadataArray
    };

    // Create entity
    Entity createEntityResponse = check ep->CreateEntity(createEntityRequest);
    io:println("Entity created with ID: " + createEntityResponse.id);
    
    // Verify initial metadata
    ReadEntityRequest readEntityRequest = {
        entity: {
            id: testId,
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
        output: ["metadata"]
    };
    Entity initialReadResponse = check ep->ReadEntity(readEntityRequest);
    verifyMetadata(initialReadResponse.metadata, {"key1": initialValue1, "key2": initialValue2});
    io:println("Initial metadata verified");
    
    // Create updated metadata array
    record {| string key; pbAny:Any value; |}[] updatedMetadataArray = [];
    pbAny:Any packedUpdatedValue1 = check pbAny:pack(updatedValue1);
    pbAny:Any packedUpdatedValue2 = check pbAny:pack(updatedValue2);
    pbAny:Any packedNewValue3 = check pbAny:pack(newValue3);
    updatedMetadataArray.push({key: "key1", value: packedUpdatedValue1});
    updatedMetadataArray.push({key: "key2", value: packedUpdatedValue2});
    updatedMetadataArray.push({key: "key3", value: packedNewValue3});

    // Update entity request
    Entity updateEntityRequest = {
        id: testId,
        kind: {
            major: "test",
            minor: "update-metadata"
        },
        created: "2023-01-01",
        terminated: "",
        name: {
            startTime: "2023-01-01",
            endTime: "",
            value: check pbAny:pack("test-update-entity")
        },
        metadata: updatedMetadataArray
    };
    
    // Update entity
    UpdateEntityRequest updateRequest = {
        id: testId,
        entity: updateEntityRequest
    };
    Entity updateEntityResponse = check ep->UpdateEntity(updateRequest);
    io:println("Entity updated with ID: " + updateEntityResponse.id);
    
    // Verify updated metadata
    ReadEntityRequest updatedReadRequest = {
        entity: {
            id: testId,
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
        output: ["metadata"]
    };
    Entity updatedReadResponse = check ep->ReadEntity(updatedReadRequest);
    verifyMetadata(updatedReadResponse.metadata, {
        "key1": updatedValue1, 
        "key2": updatedValue2,
        "key3": newValue3
    });
    io:println("Updated metadata verified");
    
    // Clean up
    Empty _ = check ep->DeleteEntity({id: testId});
    Empty _ = check ep->DeleteEntity({id: testId});
    io:println("Test entity deleted");
    
    return;
}

// Helper function to verify metadata
function verifyMetadata(record {| string key; pbAny:Any value; |}[] metadata, map<string> expected) {
    map<string> actual = {};
    foreach var item in metadata {
        string|error unwrapped = unwrapAny(item.value);
        if unwrapped is string {
            actual[item.key] = unwrapped.trim();
        }
    }
    
    // Verify all expected key-value pairs exist in the actual metadata
    foreach var [key, expectedValue] in expected.entries() {
        test:assertTrue(actual.hasKey(key), "Metadata key not found: " + key);
        test:assertEquals(actual[key] ?: "", expectedValue, 
            string `Metadata value for ${key} doesn't match: expected ${expectedValue}, got ${actual[key] ?: ""}`);
    }
}

@test:Config {}
function testEntityReading() returns error? {
    // Initialize the client
    CrudServiceClient ep = check new (testCrudServiceUrl);
    
    // Test data setup
    string testId = "test-entity-read";
    string metadataKey = "readTest";
    string metadataValue = "read-test-value";
    
    // Create a test entity first
    record {| string key; pbAny:Any value; |}[] metadataArray = [];
    pbAny:Any packedValue = check pbAny:pack(metadataValue);
    metadataArray.push({key: metadataKey, value: packedValue});
    
    Entity createEntityRequest = {
        id: testId,
        kind: {
            major: "test",
            minor: "read-entity"
        },
        created: "2023-01-01",
        terminated: "",
        name: {
            startTime: "2023-01-01",
            endTime: "",
            value: check pbAny:pack("test-read-entity")
        },
        metadata: metadataArray,
        attributes: [],
        relationships: []
    };
    
    // Create entity
    Entity createEntityResponse = check ep->CreateEntity(createEntityRequest);
    io:println("Test entity created with ID: " + createEntityResponse.id);
    io:println("Created entity metadata: ", createEntityResponse.metadata);
    
    // Read the entity
    ReadEntityRequest readEntityRequest = {
        entity: {
            id: testId,
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
        output: ["metadata"]
    };
    io:println("ReadEntityRequest: ", readEntityRequest);
    Entity readEntityResponse = check ep->ReadEntity(readEntityRequest);
    io:println("Entity retrieved, verifying data...");
    io:println("Retrieved entity: ", readEntityResponse);
    io:println("Retrieved entity metadata: ", readEntityResponse.metadata);
    
    // Verify entity fields
    test:assertEquals(readEntityResponse.id, testId, "Entity ID mismatch");
    
    // Verify metadata
    boolean foundMetadata = false;
    foreach var item in readEntityResponse.metadata {
        if item.key == metadataKey {
            string|error unwrapped = unwrapAny(item.value);
            if unwrapped is string {
                test:assertEquals(unwrapped.trim(), metadataValue, 
                    string `Metadata value mismatch: expected ${metadataValue}, got ${unwrapped}`);
                foundMetadata = true;
            }
        }
    }
    
    test:assertTrue(foundMetadata, "Expected metadata key not found");
    
    // Test reading non-existent entity
    string nonExistentId = "non-existent-entity-" + testId;
    ReadEntityRequest nonExistentRequest = {
        entity: {
            id: nonExistentId,
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
        output: ["metadata"]
    };
    Entity nonExistentEntity = check ep->ReadEntity(nonExistentRequest);
    io:println("Non-existent entity: " + nonExistentEntity.id);
    io:println("Non-existent entity metadata: ", nonExistentEntity.metadata);
    
    // Validate that metadata for non-existent entity is empty
    test:assertEquals(nonExistentEntity.metadata.length(), 0, "Non-existent entity should have empty metadata");
    
    // Assert that we get an error for non-existent entity
    // For non-existence entities, we send a response with an empty data
    // But once the Result API is integrated this can be tested. 
    // FIXME: https://github.com/LDFLK/nexoan/issues/23
    // test:assertTrue(nonExistentResponse is error, "Expected error for non-existent entity ID");
    
    // Clean up
    Empty _ = check ep->DeleteEntity({id: testId});
    Empty _ = check ep->DeleteEntity({id: testId});
    io:println("Test entity deleted");
    
    return;
}

@test:Config {}
function testCreateMinimalGraphEntity() returns error? {
    // Initialize the client
    CrudServiceClient ep = check new (testCrudServiceUrl);
    
    // Test data setup - minimal entity with just required fields
    string testId = "test-minimal-entity";
    
    // Create entity request with only required fields - no metadata, attributes, or relationships
    Entity createEntityRequest = {
        id: testId,
        kind: {
            major: "test",
            minor: "minimal"
        },
        created: "2023-01-01",
        terminated: "",
        name: {
            startTime: "2023-01-01",
            endTime: "",
            value: check pbAny:pack("minimal-test-entity")
        },
        metadata: [],
        attributes: [],
        relationships: []
    };

    // Create entity
    Entity createEntityResponse = check ep->CreateEntity(createEntityRequest);
    io:println("Minimal entity created with ID: " + createEntityResponse.id);
    
    // Verify entity was created correctly
    ReadEntityRequest readEntityRequest = {
        entity: {
            id: testId,
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
        output: ["metadata", "attributes", "relationships"]
    };
    Entity readEntityResponse = check ep->ReadEntity(readEntityRequest);
    
    // Basic entity verification
    test:assertEquals(readEntityResponse.id, testId, "Entity ID doesn't match");
    test:assertEquals(readEntityResponse.kind.major, "test", "Entity kind.major doesn't match");
    test:assertEquals(readEntityResponse.kind.minor, "minimal", "Entity kind.minor doesn't match");
    
    // Verify empty collections
    test:assertEquals(readEntityResponse.metadata.length(), 0, "Metadata should be empty");
    test:assertEquals(readEntityResponse.attributes.length(), 0, "Attributes should be empty");
    test:assertEquals(readEntityResponse.relationships.length(), 0, "Relationships should be empty");
    
    // Clean up
    Empty _ = check ep->DeleteEntity({id: testId});
    Empty _ = check ep->DeleteEntity({id: testId});
    io:println("Test minimal entity deleted");
    
    return;
}

@test:Config {}
function testCreateMinimalGraphEntityViaRest() returns error? {
    // Initialize an HTTP client for the REST API
    http:Client restClient = check new (testUpdateServiceUrl);
    
    // Test data setup - minimal JSON entity
    string testId = "test-minimal-json-entity";
    
    // Minimal JSON payload with required fields matching the Entity structure
    json minimalEntityJson = {
        "id": testId,
        "kind": {
            "major": "test",
            "minor": "minimal-json"
        },
        "created": "2023-01-01",
        "terminated": "",
        "name": {
            "startTime": "2023-01-01",
            "endTime": "",
            "value": "minimal-json-test-entity"
        },
        "metadata": [],
        "attributes": [],
        "relationships": []
    };

    // Create entity via REST API
    http:Response|error response = restClient->post("/entities", minimalEntityJson);
    
    // Verify HTTP request was successful
    if response is error {
        test:assertFail("Failed to create entity via REST API: " + response.message());
    }
    
    http:Response httpResponse = <http:Response>response;
    test:assertEquals(httpResponse.statusCode, 201, "Expected 201 OK status code");
    
    // Parse response JSON
    json responseJson = check httpResponse.getJsonPayload();
    test:assertEquals(check responseJson.id, testId, "Entity ID in response doesn't match");
    
    // Initialize the gRPC client to verify entity was properly created
    CrudServiceClient ep = check new (testCrudServiceUrl);
    
    // Verify entity data
    ReadEntityRequest readEntityRequest = {
        entity: {
            id: testId,
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
        output: ["metadata","attributes", "relationships"]
    };
    Entity readEntityResponse = check ep->ReadEntity(readEntityRequest);
    
    // Basic entity verification
    test:assertEquals(readEntityResponse.id, testId, "Entity ID doesn't match");
    test:assertEquals(readEntityResponse.kind.major, "test", "Entity kind.major doesn't match");
    test:assertEquals(readEntityResponse.kind.minor, "minimal-json", "Entity kind.minor doesn't match");
    
    // Verify empty collections
    test:assertEquals(readEntityResponse.metadata.length(), 0, "Metadata should be empty");
    test:assertEquals(readEntityResponse.attributes.length(), 0, "Attributes should be empty");
    test:assertEquals(readEntityResponse.relationships.length(), 0, "Relationships should be empty");
    
    // Clean up
    Empty _ = check ep->DeleteEntity({id: testId});
    Empty _ = check ep->DeleteEntity({id: testId});
    io:println("Test minimal JSON entity deleted");
    
    return;
}

@test:Config {
    groups: ["entity", "relationship"]
}
function testEntityWithRelationship() returns error? {
    // Test IDs for entities
    string sourceEntityId = "test-entity-with-relationship-source";
    string targetEntityId = "test-entity-with-relationship-target";
    
    // Initialize REST client
    http:Client restClient = check new (testUpdateServiceUrl);
    
    // Create source entity
    json sourceEntityJson = {
        "id": sourceEntityId,
        "kind": {
            "major": "test",
            "minor": "relationship-source"
        },
        "created": "2023-01-01",
        "terminated": "",
        "name": {
            "startTime": "2023-01-01",
            "endTime": "",
            "value": "source-entity"
        },
        "metadata": [],
        "attributes": [],
        "relationships": []
    };
    
    // Create target entity
    json targetEntityJson = {
        "id": targetEntityId,
        "kind": {
            "major": "test",
            "minor": "relationship-target"
        },
        "created": "2023-01-01", 
        "terminated": "",
        "name": {
            "startTime": "2023-01-01",
            "endTime": "",
            "value": "target-entity"
        },
        "metadata": [],
        "attributes": [],
        "relationships": []
    };
    
    // Create both entities via REST API
    http:Response|error sourceResponse = restClient->post("/entities", sourceEntityJson);
    http:Response|error targetResponse = restClient->post("/entities", targetEntityJson);
    
    // Verify HTTP requests were successful
    if sourceResponse is error {
        test:assertFail("Failed to create source entity: " + sourceResponse.message());
    }
    if targetResponse is error {
        test:assertFail("Failed to create target entity: " + targetResponse.message());
    }
    
    http:Response sourceHttpResponse = <http:Response>sourceResponse;
    http:Response targetHttpResponse = <http:Response>targetResponse;
    test:assertEquals(sourceHttpResponse.statusCode, 201, "Expected 201 status code for source entity");
    test:assertEquals(targetHttpResponse.statusCode, 201, "Expected 201 status code for target entity");
    
    // Create relationship between entities - include full entity structure
    string relationshipId = "rel-" + sourceEntityId + "-" + targetEntityId;
    json relationshipJson = {
        "id": sourceEntityId,
        "kind": {
        },
        "created": "",
        "terminated": "",
        "name": {
        },
        "metadata": [],
        "attributes": [],
        "relationships": {
            relationshipId: {
                "relatedEntityId": targetEntityId,
                "startTime": "2023-01-01",
                "endTime": "",
                "id": relationshipId,
                "name": "CONNECTS_TO"
            }
        }
    };
    
    // Update source entity with relationship
    http:Response|error updateResponse = restClient->put("/entities/" + sourceEntityId, relationshipJson);
    
    // Verify update was successful
    if updateResponse is error {
        test:assertFail("Failed to update entity with relationship: " + updateResponse.message());
    }
    
    http:Response updateHttpResponse = <http:Response>updateResponse;
    test:assertEquals(updateHttpResponse.statusCode, 200, "Expected 200 status code for relationship update");
    
    // Initialize the gRPC client to verify relationship was properly created
    CrudServiceClient ep = check new (testCrudServiceUrl);
    
    // Read source entity to verify relationship
    ReadEntityRequest readEntityRequest = {
        entity: {
            id: sourceEntityId,
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
    Entity readEntityResponse = check ep->ReadEntity(readEntityRequest);
    
    // Verify relationship data
    test:assertEquals(readEntityResponse.relationships.length(), 1, "Entity should have one relationship");
    
    // Find the relationship by iterating through the array
    Relationship? targetRelationship = ();
    foreach var rel in readEntityResponse.relationships {
        if rel.key == relationshipId {
            targetRelationship = rel.value;
            break;
        }
    }
    
    io:println("Target relationship: " + targetRelationship.toJsonString());
    test:assertFalse(targetRelationship is (), "Relationship with key 'CONNECTS_TO' not found");
    Relationship relationship = <Relationship>targetRelationship;
    test:assertEquals(relationship.relatedEntityId, targetEntityId, "Related entity ID doesn't match");
    test:assertEquals(relationship.name, "CONNECTS_TO", "Relationship name doesn't match");
    test:assertEquals(relationship.startTime, "2023-01-01T00:00:00Z", "Relationship start time doesn't match");
    test:assertEquals(relationship.id, relationshipId, "Relationship ID doesn't match");
    
    // Clean up
    Empty _ = check ep->DeleteEntity({id: sourceEntityId});
    Empty _ = check ep->DeleteEntity({id: sourceEntityId});
    Empty _ = check ep->DeleteEntity({id: targetEntityId});
    io:println("Test entities with relationship deleted");
    
    return;
}

@test:Config {
    groups: ["entity", "attributes", "graph", "only_nodes"]
}
function testEntityWithSimpleOnlyNodesGraphAttributes() returns error? {
    // TODO: Complete Test Case https://github.com/LDFLK/nexoan/issues/143
    // Test ID for entity
    string testId = "test-entity-simple-only-nodes-graph";
    
    // Initialize the gRPC client to verify entity
    CrudServiceClient ep = check new (testCrudServiceUrl);
    
    // Create entity with tabular data in attributes
    json socialNetworkGraph = {
        "nodes": [
            {"id": "user1", "type": "user", "properties": {"name": "Alice", "age": 30, "location": "NY"}},
            {"id": "user2", "type": "user", "properties": {"name": "Bob", "age": 25, "location": "SF"}},
            {"id": "user3", "type": "user", "properties": {"name": "Charlie", "age": 35, "location": "LA"}},
            {"id": "post1", "type": "post", "properties": {"title": "Hello", "content": "World", "created": "2024-03-20"}},
            {"id": "post2", "type": "post", "properties": {"title": "Graph", "content": "DB", "created": "2024-03-21"}}
        ]
    };

    // Convert JSON to protobuf Any values
    pbAny:Any socialNetworkGraphAny = check jsonToAny(socialNetworkGraph);

    Entity createEntityRequest = {
        id: testId,
        kind: {
            major: "test",
            minor: "graph"
        },
        created: "2024-01-01T00:00:00Z",
        terminated: "",
        name: {
            startTime: "2024-01-01T00:00:00Z",
            endTime: "",
            value: check pbAny:pack("test-entity")
        },
        metadata: [
            {
                key: "test_metadata",
                value: check pbAny:pack("test_value")
            }
        ],
        attributes: [
            {
                key: "employee_salary_history",
                value: {
                    values: [
                        {
                            startTime: "2024-01-01T00:00:00Z",
                            endTime: "",
                            value: socialNetworkGraphAny
                        }
                    ]
                }
            }
        ],
        relationships: []
    };

    // Create entity via gRPC
    Entity createEntityResponse = check ep->CreateEntity(createEntityRequest);
    io:println("Entity created with ID: " + createEntityResponse.id);
    
    // Read entity to verify attributes
    ReadEntityRequest readEntityRequest = {
        id: testId,
        entity: {
            id: testId,
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
        output: ["metadata", "attributes", "relationships"]
    };
    Entity readEntityResponse = check ep->ReadEntity(readEntityRequest);
    
    // Verify the response
    test:assertTrue(readEntityResponse.id != "", "Entity should be found");
    test:assertEquals(readEntityResponse.id, testId, "Entity ID should match");
    
    // Clean up
    Empty _ = check ep->DeleteEntity({id: testId});
    Empty _ = check ep->DeleteEntity({id: testId});
    io:println("Test entity with graph attributes deleted");
    
    return;
}

@test:Config {
    groups: ["entity", "attributes", "graph", "simple"]
}
function testEntityWithSimpleGraphAttributes() returns error? {
    // TODO: Complete Test Case https://github.com/LDFLK/nexoan/issues/143
    // Test ID for entity
    string testId = "test-simple-entity-graph";
    
    // Initialize the gRPC client to verify entity
    CrudServiceClient ep = check new (testCrudServiceUrl);
    
    // Create entity with tabular data in attributes
    json socialNetworkGraph = {
        "nodes": [
            {"id": "user1", "type": "user", "properties": {"name": "Alice", "age": 30, "location": "NY"}},
            {"id": "user2", "type": "user", "properties": {"name": "Bob", "age": 25, "location": "SF"}},
            {"id": "user3", "type": "user", "properties": {"name": "Charlie", "age": 35, "location": "LA"}},
            {"id": "post1", "type": "post", "properties": {"title": "Hello", "content": "World", "created": "2024-03-20"}},
            {"id": "post2", "type": "post", "properties": {"title": "Graph", "content": "DB", "created": "2024-03-21"}}
        ],
        "edges": [
            {"source": "user1", "target": "user2", "type": "follows", "properties": {"since": "2024-01-01"}},
            {"source": "user2", "target": "user3", "type": "follows", "properties": {"since": "2024-02-01"}},
            {"source": "user1", "target": "post1", "type": "created", "properties": {"timestamp": "2024-03-20T10:00:00Z"}},
            {"source": "user2", "target": "post1", "type": "likes", "properties": {"timestamp": "2024-03-20T11:00:00Z"}},
            {"source": "user3", "target": "post2", "type": "created", "properties": {"timestamp": "2024-03-21T09:00:00Z"}}
        ]
    };

    // Convert JSON to protobuf Any values
    pbAny:Any socialNetworkGraphAny = check jsonToAny(socialNetworkGraph);

    Entity createEntityRequest = {
        id: testId,
        kind: {
            major: "test",
            minor: "graph"
        },
        created: "2024-01-01T00:00:00Z",
        terminated: "",
        name: {
            startTime: "2024-01-01T00:00:00Z",
            endTime: "",
            value: check pbAny:pack("test-entity")
        },
        metadata: [
            {
                key: "test_metadata",
                value: check pbAny:pack("test_value")
            }
        ],
        attributes: [
            {
                key: "employee_salary_history",
                value: {
                    values: [
                        {
                            startTime: "2024-01-01T00:00:00Z",
                            endTime: "",
                            value: socialNetworkGraphAny
                        }
                    ]
                }
            }
        ],
        relationships: []
    };

    // Create entity via gRPC
    Entity createEntityResponse = check ep->CreateEntity(createEntityRequest);
    io:println("Entity created with ID: " + createEntityResponse.id);
    
    // Read entity to verify attributes
    ReadEntityRequest readEntityRequest = {
        id: testId,
        entity: {
            id: testId,
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
        output: ["metadata", "attributes", "relationships"]
    };
    Entity readEntityResponse = check ep->ReadEntity(readEntityRequest);
    
    // Verify the response
    test:assertTrue(readEntityResponse.id != "", "Entity should be found");
    test:assertEquals(readEntityResponse.id, testId, "Entity ID should match");
    
    // Clean up
    Empty _ = check ep->DeleteEntity({id: testId});
    io:println("Test entity with graph attributes deleted");
    
    return;
}

@test:Config {
    groups: ["entity", "attributes", "graph", "multi"]
}
function testEntityWithMultiGraphAttributes() returns error? {
    // TODO: Complete Test Case https://github.com/LDFLK/nexoan/issues/143
    // Test ID for entity
    string testId = "test-entity-graph";
    
    // Initialize the gRPC client to verify entity
    CrudServiceClient ep = check new (testCrudServiceUrl);
    
    // Create entity with tabular data in attributes
    json salaryGraph = {
        "nodes": [
					{"id": "user1", "type": "user", "properties": {"name": "Alice", "age": 30, "location": "NY"}},
					{"id": "user2", "type": "user", "properties": {"name": "Bob", "age": 25, "location": "SF"}},
					{"id": "user3", "type": "user", "properties": {"name": "Charlie", "age": 35, "location": "LA"}},
					{"id": "post1", "type": "post", "properties": {"title": "Hello", "content": "World", "created": "2024-03-20"}},
					{"id": "post2", "type": "post", "properties": {"title": "Graph", "content": "DB", "created": "2024-03-21"}}
				],
				"edges": [
					{"source": "user1", "target": "user2", "type": "follows", "properties": {"since": "2024-01-01"}},
					{"source": "user2", "target": "user3", "type": "follows", "properties": {"since": "2024-02-01"}},
					{"source": "user1", "target": "post1", "type": "created", "properties": {"timestamp": "2024-03-20T10:00:00Z"}},
					{"source": "user2", "target": "post1", "type": "likes", "properties": {"timestamp": "2024-03-20T11:00:00Z"}},
					{"source": "user3", "target": "post2", "type": "created", "properties": {"timestamp": "2024-03-21T09:00:00Z"}}
				]
    };

    json projectGraph = {
        "nodes": [
            {"id": "proj_redesign", "type": "project", "properties": {"id": "P001", "name": "System Redesign", "status": "active"}},
            {"id": "proj_migration", "type": "project", "properties": {"id": "P002", "name": "API Migration", "status": "completed"}},
            {"id": "proj_audit", "type": "project", "properties": {"id": "P003", "name": "Security Audit", "status": "completed"}},
            {"id": "role_lead", "type": "role", "properties": {"title": "Lead Developer", "level": "senior"}},
            {"id": "role_dev", "type": "role", "properties": {"title": "Developer", "level": "mid"}}
        ],
        "edges": [
            {"source": "proj_redesign", "target": "role_lead", "type": "has_role", "properties": {"start_date": "2024-01-01", "end_date": ""}},
            {"source": "proj_migration", "target": "role_dev", "type": "has_role", "properties": {"start_date": "2023-06-01", "end_date": "2023-12-31"}},
            {"source": "proj_audit", "target": "role_dev", "type": "has_role", "properties": {"start_date": "2023-01-01", "end_date": "2023-05-31"}},
            {"source": "proj_redesign", "target": "proj_migration", "type": "follows", "properties": {"transition_date": "2023-12-31"}},
            {"source": "proj_migration", "target": "proj_audit", "type": "follows", "properties": {"transition_date": "2023-05-31"}}
        ]
    };

    // Convert JSON to protobuf Any values
    pbAny:Any salaryGraphAny = check jsonToAny(salaryGraph);
    pbAny:Any projectGraphAny = check jsonToAny(projectGraph);

    Entity createEntityRequest = {
        id: testId,
        kind: {
            major: "test",
            minor: "graph"
        },
        created: "2024-01-01T00:00:00Z",
        terminated: "",
        name: {
            startTime: "2024-01-01T00:00:00Z",
            endTime: "",
            value: check pbAny:pack("test-entity")
        },
        metadata: [
            {
                key: "test_metadata",
                value: check pbAny:pack("test_value")
            }
        ],
        attributes: [
            {
                key: "employee_salary_history",
                value: {
                    values: [
                        {
                            startTime: "2024-01-01T00:00:00Z",
                            endTime: "",
                            value: salaryGraphAny
                        }
                    ]
                }
            },
            {
                key: "project_assignments",
                value: {
                    values: [
                        {
                            startTime: "2024-01-01T00:00:00Z",
                            endTime: "",
                            value: projectGraphAny
                        }
                    ]
                }
            }
        ],
        relationships: []
    };

    // Create entity via gRPC
    Entity createEntityResponse = check ep->CreateEntity(createEntityRequest);
    io:println("Entity created with ID: " + createEntityResponse.id);
    
    // Read entity to verify attributes
    ReadEntityRequest readEntityRequest = {
        id: testId,
        entity: {
            id: testId,
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
        output: ["metadata", "attributes", "relationships"]
    };
    Entity readEntityResponse = check ep->ReadEntity(readEntityRequest);
    
    // Verify the response
    test:assertTrue(readEntityResponse.id != "", "Entity should be found");
    test:assertEquals(readEntityResponse.id, testId, "Entity ID should match");
    
    // Clean up
    
    Empty _ = check ep->DeleteEntity({id: testId});
    io:println("Test entity with graph attributes deleted");
    
    return;
}

@test:Config {
    groups: ["entity", "attributes", "list"]
}
function testEntityWithSimpleListAttributes() returns error? {
    // TODO: Complete Test Case https://github.com/LDFLK/nexoan/issues/143
    // Test ID for entity
    string testId = "test-entity-list";
    
    // Initialize the gRPC client to verify entity
    CrudServiceClient ep = check new (testCrudServiceUrl);
    
    // Create entity with list data in attributes
    json salaryList = {
        "values": [
            1,
            2,
            3
        ]
    };

    // Convert JSON to protobuf Any values
    pbAny:Any salaryListAny = check jsonToAny(salaryList);

    Entity createEntityRequest = {
        id: testId,
        kind: {
            major: "test",
            minor: "list"
        },
        created: "2024-01-01T00:00:00Z",
        terminated: "",
        name: {
            startTime: "2024-01-01T00:00:00Z",
            endTime: "",
            value: check pbAny:pack("test-entity")
        },
        metadata: [
            {
                key: "test_metadata",
                value: check pbAny:pack("test_value")
            }
        ],
        attributes: [
            {
                key: "salary_history",
                value: {
                    values: [
                        {
                            startTime: "2024-01-01T00:00:00Z",
                            endTime: "",
                            value: salaryListAny
                        }
                    ]
                }
            }
        ],
        relationships: []
    };

    // Create entity via gRPC
    Entity createEntityResponse = check ep->CreateEntity(createEntityRequest);
    io:println("Entity created with ID: " + createEntityResponse.id);
    
    // Clean up
    Empty _ = check ep->DeleteEntity({id: testId});
    io:println("Test entity with list attributes deleted");
    
    return;
}

@test:Config {
    groups: ["entity", "attributes", "list", "mixed"]
}
function testEntityWithMixedTypeListAttributes() returns error? {
    // TODO: Complete Test Case https://github.com/LDFLK/nexoan/issues/143
    // Test ID for entity
    string testId = "test-entity-mixed-list";
    
    // Initialize the gRPC client to verify entity
    CrudServiceClient ep = check new (testCrudServiceUrl);
    
    // Create entity with mixed type list data in attributes
    json mixedTypeList = {
        "values": [
            "Hello",
            42,
            true,
            "2024-03-21",
            "2024-03-21T15:30:00Z",
            null
        ]
    };

    // Convert JSON to protobuf Any values
    pbAny:Any mixedTypeListAny = check jsonToAny(mixedTypeList);

    Entity createEntityRequest = {
        id: testId,
        kind: {
            major: "test",
            minor: "mixed-list"
        },
        created: "2024-01-01T00:00:00Z",
        terminated: "",
        name: {
            startTime: "2024-01-01T00:00:00Z",
            endTime: "",
            value: check pbAny:pack("test-entity")
        },
        metadata: [
            {
                key: "test_metadata",
                value: check pbAny:pack("test_value")
            }
        ],
        attributes: [
            {
                key: "mixed_type_history",
                value: {
                    values: [
                        {
                            startTime: "2024-01-01T00:00:00Z",
                            endTime: "",
                            value: mixedTypeListAny
                        }
                    ]
                }
            }
        ],
        relationships: []
    };

    // Create entity via gRPC
    Entity createEntityResponse = check ep->CreateEntity(createEntityRequest);
    io:println("Entity created with ID: " + createEntityResponse.id);
    
    // Read entity to verify attributes
    ReadEntityRequest readEntityRequest = {
        id: testId,
        entity: {
            id: testId,
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
        output: ["metadata", "attributes", "relationships"]
    };
    Entity readEntityResponse = check ep->ReadEntity(readEntityRequest);
    
    // Verify the response
    test:assertTrue(readEntityResponse.id != "", "Entity should be found");
    test:assertEquals(readEntityResponse.id, testId, "Entity ID should match");
    
    // Clean up
    Empty _ = check ep->DeleteEntity({id: testId});
    io:println("Test entity with mixed type list attributes deleted");
    
    return;
}

@test:Config {
    groups: ["entity", "attributes", "list", "empty"]
}
function testEntityWithEmptyListAttributes() returns error? {
    // TODO: Complete Test Case https://github.com/LDFLK/nexoan/issues/143
    // Test ID for entity
    string testId = "test-entity-empty-list";
    
    // Initialize the gRPC client to verify entity
    CrudServiceClient ep = check new (testCrudServiceUrl);
    
    // Create entity with empty list data in attributes
    json emptyList = {
        "values": []
    };

    // Convert JSON to protobuf Any values
    pbAny:Any emptyListAny = check jsonToAny(emptyList);

    Entity createEntityRequest = {
        id: testId,
        kind: {
            major: "test",
            minor: "empty-list"
        },
        created: "2024-01-01T00:00:00Z",
        terminated: "",
        name: {
            startTime: "2024-01-01T00:00:00Z",
            endTime: "",
            value: check pbAny:pack("test-entity")
        },
        metadata: [
            {
                key: "test_metadata",
                value: check pbAny:pack("test_value")
            }
        ],
        attributes: [
            {
                key: "empty_list_history",
                value: {
                    values: [
                        {
                            startTime: "2024-01-01T00:00:00Z",
                            endTime: "",
                            value: emptyListAny
                        }
                    ]
                }
            }
        ],
        relationships: []
    };

    // Create entity via gRPC
    Entity createEntityResponse = check ep->CreateEntity(createEntityRequest);
    io:println("Entity created with ID: " + createEntityResponse.id);
    
    // Read entity to verify attributes
    ReadEntityRequest readEntityRequest = {
        id: testId,
        entity: {
            id: testId,
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
        output: ["metadata", "attributes", "relationships"]
    };
    Entity readEntityResponse = check ep->ReadEntity(readEntityRequest);
    
    // Verify the response
    test:assertTrue(readEntityResponse.id != "", "Entity should be found");
    test:assertEquals(readEntityResponse.id, testId, "Entity ID should match");
    
    // Clean up
    Empty _ = check ep->DeleteEntity({id: testId});
    io:println("Test entity with empty list attributes deleted");
    
    return;
}

@test:Config {
    groups: ["entity", "attributes", "map"]
}
function testEntityWithMapAttributes() returns error? {
    // TODO: Complete Test Case https://github.com/LDFLK/nexoan/issues/143
    // Test ID for entity
    string testId = "test-entity-map";
    
    // Initialize the gRPC client to verify entity
    CrudServiceClient ep = check new (testCrudServiceUrl);
    
    // Create entity with map data in attributes
    json userProfileMap = {
       "properties": {
					"name": "John",
					"age": 30,
					"active": true
				}
    };

    // Convert JSON to protobuf Any values
    pbAny:Any userProfileMapAny = check jsonToAny(userProfileMap);

    Entity createEntityRequest = {
        id: testId,
        kind: {
            major: "test",
            minor: "map"
        },
        created: "2024-01-01T00:00:00Z",
        terminated: "",
        name: {
            startTime: "2024-01-01T00:00:00Z",
            endTime: "",
            value: check pbAny:pack("test-entity")
        },
        metadata: [
            {
                key: "test_metadata",
                value: check pbAny:pack("test_value")
            }
        ],
        attributes: [
            {
                key: "user_profile",
                value: {
                    values: [
                        {
                            startTime: "2024-01-01T00:00:00Z",
                            endTime: "",
                            value: userProfileMapAny
                        }
                    ]
                }
            }
        ],
        relationships: []
    };

    // Create entity via gRPC
    Entity createEntityResponse = check ep->CreateEntity(createEntityRequest);
    io:println("Entity created with ID: " + createEntityResponse.id);
    
    // Read entity to verify attributes
    ReadEntityRequest readEntityRequest = {
        id: testId,
        entity: {
            id: testId,
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
        output: ["metadata", "attributes", "relationships"]
    };
    Entity readEntityResponse = check ep->ReadEntity(readEntityRequest);
    
    // Verify the response
    test:assertTrue(readEntityResponse.id != "", "Entity should be found");
    test:assertEquals(readEntityResponse.id, testId, "Entity ID should match");
    
    // Clean up
    Empty _ = check ep->DeleteEntity({id: testId});
    io:println("Test entity with map attributes deleted");
    
    return;
}

@test:Config {
    groups: ["entity", "attributes", "map", "nested"]
}
function testEntityWithNestedMapAttributes() returns error? {
    // TODO: Complete Test Case https://github.com/LDFLK/nexoan/issues/143
    // Test ID for entity
    string testId = "test-entity-nested-map";
    
    // Initialize the gRPC client to verify entity
    CrudServiceClient ep = check new (testCrudServiceUrl);
    
    // Create entity with nested map data in attributes
    json nestedMap = {
        "organization": {
            "name": "Tech Corp",
            "departments": {
                "engineering": {
                    "head": "Alice Smith",
                    "size": 50,
                    "projects": {
                        "project1": {
                            "name": "System Redesign",
                            "status": "active",
                            "team": ["John", "Jane", "Bob"]
                        },
                        "project2": {
                            "name": "API Migration",
                            "status": "completed",
                            "team": ["Alice", "Charlie"]
                        }
                    }
                },
                "marketing": {
                    "head": "Bob Johnson",
                    "size": 20,
                    "campaigns": {
                        "campaign1": {
                            "name": "Summer Sale",
                            "budget": 50000,
                            "channels": ["social", "email", "web"]
                        }
                    }
                }
            },
            "metadata": {
                "founded": "2020-01-01",
                "location": {
                    "country": "USA",
                    "offices": ["NY", "SF", "LA"]
                }
            }
        }
    };

    // Convert JSON to protobuf Any values
    pbAny:Any nestedMapAny = check jsonToAny(nestedMap);

    Entity createEntityRequest = {
        id: testId,
        kind: {
            major: "test",
            minor: "nested-map"
        },
        created: "2024-01-01T00:00:00Z",
        terminated: "",
        name: {
            startTime: "2024-01-01T00:00:00Z",
            endTime: "",
            value: check pbAny:pack("test-entity")
        },
        metadata: [
            {
                key: "test_metadata",
                value: check pbAny:pack("test_value")
            }
        ],
        attributes: [
            {
                key: "org_structure",
                value: {
                    values: [
                        {
                            startTime: "2024-01-01T00:00:00Z",
                            endTime: "",
                            value: nestedMapAny
                        }
                    ]
                }
            }
        ],
        relationships: []
    };

    // Create entity via gRPC
    Entity createEntityResponse = check ep->CreateEntity(createEntityRequest);
    io:println("Entity created with ID: " + createEntityResponse.id);
    
    // Read entity to verify attributes
    ReadEntityRequest readEntityRequest = {
        id: testId,
        entity: {
            id: testId,
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
        output: ["metadata", "attributes", "relationships"]
    };
    Entity readEntityResponse = check ep->ReadEntity(readEntityRequest);
    
    // Verify the response
    test:assertTrue(readEntityResponse.id != "", "Entity should be found");
    test:assertEquals(readEntityResponse.id, testId, "Entity ID should match");
    
    // Clean up
    Empty _ = check ep->DeleteEntity({id: testId});
    io:println("Test entity with nested map attributes deleted");
    
    return;
}

@test:Config {
    groups: ["entity", "attributes", "map", "empty"]
}
function testEntityWithEmptyMapValues() returns error? {
    // TODO: Complete Test Case https://github.com/LDFLK/nexoan/issues/143
    // Test ID for entity
    string testId = "test-entity-empty-map-values";
    
    // Initialize the gRPC client to verify entity
    CrudServiceClient ep = check new (testCrudServiceUrl);
    
    // Create entity with map data containing empty values
    // FIXME: https://github.com/LDFLK/nexoan/issues/137
    json emptyValuesMap = {
        "properties": {
            "empty_str": "",
            "zero": 0,
            "null_val": null
        }
    };

    // Convert JSON to protobuf Any values
    pbAny:Any emptyValuesMapAny = check jsonToAny(emptyValuesMap);

    Entity createEntityRequest = {
        id: testId,
        kind: {
            major: "test",
            minor: "empty-map-values"
        },
        created: "2024-01-01T00:00:00Z",
        terminated: "",
        name: {
            startTime: "2024-01-01T00:00:00Z",
            endTime: "",
            value: check pbAny:pack("test-entity")
        },
        metadata: [
            {
                key: "test_metadata",
                value: check pbAny:pack("test_value")
            }
        ],
        attributes: [
            {
                key: "empty_values",
                value: {
                    values: [
                        {
                            startTime: "2024-01-01T00:00:00Z",
                            endTime: "",
                            value: emptyValuesMapAny
                        }
                    ]
                }
            }
        ],
        relationships: []
    };

    // Create entity via gRPC
    Entity createEntityResponse = check ep->CreateEntity(createEntityRequest);
    io:println("Entity created with ID: " + createEntityResponse.id);
    
    // Read entity to verify attributes
    ReadEntityRequest readEntityRequest = {
        id: testId,
        entity: {
            id: testId,
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
        output: ["metadata", "attributes", "relationships"]
    };
    Entity readEntityResponse = check ep->ReadEntity(readEntityRequest);
    
    // Verify the response
    test:assertTrue(readEntityResponse.id != "", "Entity should be found");
    test:assertEquals(readEntityResponse.id, testId, "Entity ID should match");
    
    // Clean up
    Empty _ = check ep->DeleteEntity({id: testId});
    io:println("Test entity with empty map values deleted");
    
    return;
}

@test:Config {
    groups: ["entity", "attributes", "map", "nested"]
}
function testEntityWithNestedMapValues() returns error? {
    // TODO: Complete Test Case https://github.com/LDFLK/nexoan/issues/143
    // Test ID for entity
    string testId = "test-entity-nested-map-values";
    
    // Initialize the gRPC client to verify entity
    CrudServiceClient ep = check new (testCrudServiceUrl);
    
    // Create entity with deeply nested map data
    json nestedMap = {
        "properties": {
            "user": {
                "profile": {
                    "personal": {
                        "name": "John Doe",
                        "age": 30,
                        "address": {
                            "street": "123 Main St",
                            "city": "New York",
                            "zip": "10001"
                        }
                    },
                    "preferences": {
                        "theme": "dark",
                        "notifications": true,
                        "language": "en"
                    }
                },
                "settings": {
                    "account": {
                        "type": "premium",
                        "status": "active",
                        "subscription": {
                            "plan": "yearly",
                            "start_date": "2024-01-01",
                            "end_date": "2024-12-31"
                        }
                    },
                    "security": {
                        "two_factor": true,
                        "last_login": "2024-03-21T10:00:00Z",
                        "devices": {
                            "mobile": ["iPhone", "iPad"],
                            "desktop": ["MacBook", "Windows PC"]
                        }
                    }
                }
            },
            "metadata": {
                "created_at": "2024-01-01T00:00:00Z",
                "updated_at": "2024-03-21T15:30:00Z",
                "version": {
                    "major": 1,
                    "minor": 0,
                    "patch": 0
                }
            }
        }
    };

    // Convert JSON to protobuf Any values
    pbAny:Any nestedMapAny = check jsonToAny(nestedMap);

    Entity createEntityRequest = {
        id: testId,
        kind: {
            major: "test",
            minor: "nested-map-values"
        },
        created: "2024-01-01T00:00:00Z",
        terminated: "",
        name: {
            startTime: "2024-01-01T00:00:00Z",
            endTime: "",
            value: check pbAny:pack("test-entity")
        },
        metadata: [
            {
                key: "test_metadata",
                value: check pbAny:pack("test_value")
            }
        ],
        attributes: [
            {
                key: "nested_data",
                value: {
                    values: [
                        {
                            startTime: "2024-01-01T00:00:00Z",
                            endTime: "",
                            value: nestedMapAny
                        }
                    ]
                }
            }
        ],
        relationships: []
    };

    // Create entity via gRPC
    Entity createEntityResponse = check ep->CreateEntity(createEntityRequest);
    io:println("Entity created with ID: " + createEntityResponse.id);
    
    // Read entity to verify attributes
    ReadEntityRequest readEntityRequest = {
        id: testId,
        entity: {
            id: testId,
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
        output: ["metadata", "attributes", "relationships"]
    };
    Entity readEntityResponse = check ep->ReadEntity(readEntityRequest);
    
    // Verify the response
    test:assertTrue(readEntityResponse.id != "", "Entity should be found");
    test:assertEquals(readEntityResponse.id, testId, "Entity ID should match");
    
    // Clean up
    Empty _ = check ep->DeleteEntity({id: testId});
    io:println("Test entity with nested map values deleted");
    
    return;
}






