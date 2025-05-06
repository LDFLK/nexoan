import requests
import json
import sys
import os

def get_service_urls():
    query_host = os.getenv('QUERY_SERVICE_HOST', 'localhost')
    query_port = os.getenv('QUERY_SERVICE_PORT', '8081')
    update_host = os.getenv('UPDATE_SERVICE_HOST', 'localhost')
    update_port = os.getenv('UPDATE_SERVICE_PORT', '8080')
    
    return {
        'query': f"http://{query_host}:{query_port}/v1/entities",
        'update': f"http://{update_host}:{update_port}/entities"
    }

# Get service URLs from environment variables
urls = get_service_urls()
QUERY_API_URL = urls['query']
UPDATE_API_URL = urls['update']

ENTITY_ID = "query-test-entity"
RELATED_ID_1 = "query-related-entity-1"
RELATED_ID_2 = "query-related-entity-2"
RELATED_ID_3 = "query-related-entity-3"


"""
The current tests only contain metadata validation.
"""

def decode_protobuf_any_value(any_value):
    """Decode a protobuf Any value to get the actual string value"""
    if isinstance(any_value, dict) and 'typeUrl' in any_value and 'value' in any_value:
        if 'StringValue' in any_value['typeUrl']:
            try:
                # If it's hex encoded (which appears to be the case)
                hex_value = any_value['value']
                binary_data = bytes.fromhex(hex_value)
                # For StringValue in hex format, typically the structure is:
                # 0A (field tag) + 03 (length) + actual string bytes
                # Skip the first 2 bytes (field tag and length)
                if len(binary_data) > 2:
                    return binary_data[2:].decode('utf-8')
            except Exception as e:
                print(f"Failed to decode protobuf value: {e}")
                return any_value['value']
    
    # If any_value is a string that looks like a JSON object
    elif isinstance(any_value, str) and any_value.startswith('{') and any_value.endswith('}'):
        try:
            # Try to parse it as JSON
            obj = json.loads(any_value)
            # Recursively decode
            return decode_protobuf_any_value(obj)
        except json.JSONDecodeError:
            pass
    
    # Return the original value if decoding fails
    return any_value

def create_entity_for_query():
    """Create a base entity with metadata, attributes, and relationships."""
    print("\n🟢 Creating entity for query tests...")

# First related entity
    payload_child_1 = {
        "id": RELATED_ID_1,
        "kind": {"major": "test", "minor": "child"},
        "created": "2024-01-01T00:00:00Z",
        "terminated": "",
        "name": {
            "startTime": "2024-01-01T00:00:00Z",
            "endTime": "",
            "value": {
                "typeUrl": "type.googleapis.com/google.protobuf.StringValue",
                "value": "Query Test Entity Child 1"
            }
        },
        "metadata": [
            {"key": "source", "value": "unit-test-1"},
            {"key": "env", "value": "test-1"}
        ],
        "attributes": [
            {
                "key": "humidity",
                "value": {
                    "values": [
                        {
                            "startTime": "2024-01-01T00:00:00Z",
                            "endTime": "2024-01-02T00:00:00Z",
                            "value": {
                                "typeUrl": "type.googleapis.com/google.protobuf.StringValue",
                                "value": "10.5"
                            }
                        }
                    ]
                }
            }
        ],
        "relationships": [
        ]
    }

    # Second related entity
    payload_child_2 = {
        "id": RELATED_ID_2,
        "kind": {"major": "test", "minor": "child"},
        "created": "2024-01-01T00:00:00Z",
        "terminated": "",
        "name": {
            "startTime": "2024-01-01T00:00:00Z",
            "endTime": "",
            "value": {
                "typeUrl": "type.googleapis.com/google.protobuf.StringValue",
                "value": "Query Test Entity Child 2"
            }
        },
        "metadata": [
            {"key": "source", "value": "unit-test-2"},
            {"key": "env", "value": "test-2"}
        ],
        "attributes": [],
        "relationships": []
    }

    # Third related entity
    
    payload_child_3 = {
        "id": RELATED_ID_3,
        "kind": {"major": "test", "minor": "child"},
        "created": "2024-01-01T00:00:00Z",
        "terminated": "",
        "name": {
            "startTime": "2024-01-01T00:00:00Z",
            "endTime": "",
            "value": {
                "typeUrl": "type.googleapis.com/google.protobuf.StringValue",
                "value": "Query Test Entity Child 3"
            }
        },
        "metadata": [
            {"key": "source", "value": "unit-test-3"},
            {"key": "env", "value": "test-3"}
        ],
        "attributes": [],
        "relationships": []
    }

    payload_source = {
        "id": ENTITY_ID,
        "kind": {"major": "test", "minor": "parent"},
        "created": "2024-01-01T00:00:00Z",
        "terminated": "",
        "name": {
            "startTime": "2024-01-01T00:00:00Z",
            "endTime": "",
            "value": {
                "typeUrl": "type.googleapis.com/google.protobuf.StringValue",
                "value": "Query Test Entity"
            }
        },
        "metadata": [
            {"key": "source", "value": "unit-test"},
            {"key": "env", "value": "test"}
        ],
        "attributes": [
            {
                "key": "temperature",
                "value": {
                    "values": [
                        {
                            "startTime": "2024-01-01T00:00:00Z",
                            "endTime": "2024-01-02T00:00:00Z",
                            "value": {
                                "typeUrl": "type.googleapis.com/google.protobuf.StringValue",
                                "value": "25.5"
                            }
                        }
                    ]
                }
            }
        ],
        "relationships": [
            {
                "key": "rel-001",
                "value": {
                    "relatedEntityId": RELATED_ID_1,
                    "startTime": "2024-01-01T00:00:00Z",
                    "endTime": "2024-12-31T23:59:59Z",
                    "id": "rel-001",
                    "name": "linked"
                }
            },
            {
                "key": "rel-002",
                "value": {
                    "relatedEntityId": RELATED_ID_2,
                    "startTime": "2024-06-01T00:00:00Z",  # Different timestamp
                    "endTime": "2024-12-31T23:59:59Z",
                    "id": "rel-002",
                    "name": "linked"  # Same type as the first relationship
                }
            },
            {
                "key": "rel-003",
                "value": {
                    "relatedEntityId": RELATED_ID_3,
                    "startTime": "2024-01-01T00:00:00Z",  # Same timestamp as the first relationship
                    "endTime": "2024-12-31T23:59:59Z",
                    "id": "rel-003",
                    "name": "associated"  # Different type
                }
            }
        ]
    }

    res = requests.post(UPDATE_API_URL, json=payload_child_1)
    assert res.status_code == 201 or res.status_code == 200, f"Failed to create entity: {res.text}"
    print("✅ Created first related entity.")

    res = requests.post(UPDATE_API_URL, json=payload_child_2)
    assert res.status_code == 201 or res.status_code == 200, f"Failed to create entity: {res.text}"
    print("✅ Created second related entity.")

    res = requests.post(UPDATE_API_URL, json=payload_child_3)
    assert res.status_code == 201 or res.status_code == 200, f"Failed to create entity: {res.text}"
    print("✅ Created third related entity.")

    res = requests.post(UPDATE_API_URL, json=payload_source)
    assert res.status_code == 201 or res.status_code == 200, f"Failed to create entity: {res.text}"
    print("✅ Created base entity for query tests.")

def test_attribute_lookup():
    """Test retrieving attributes via the query API."""
    print("\n🔍 Testing attribute retrieval...")
    url = f"{QUERY_API_URL}/{ENTITY_ID}/attributes/temperature"
    res = requests.get(url)
    assert res.status_code == 404, f"Failed to get attribute: {res.text}"
    
    # Add response body validation
    body = res.json()
    assert isinstance(body, dict), "Response should be a dictionary"
    assert "error" in body, "Error message should be present in 404 response"
    print("✅ Attribute response:", json.dumps(res.json(), indent=2))

def test_metadata_lookup():
    """Test retrieving metadata."""
    print("\n🔍 Testing metadata retrieval...")
    url = f"{QUERY_API_URL}/{ENTITY_ID}/metadata"
    res = requests.get(url)
    assert res.status_code == 200, f"Failed to get metadata: {res.text}"
    
    body = res.json()
    print("✅ Raw metadata response:", json.dumps(body, indent=2))
    
    # Enhanced metadata validation
    assert isinstance(body, dict), "Metadata response should be a dictionary"
    assert len(body) == 2, f"Expected 2 metadata entries, got {len(body)}"
    assert "source" in body, "Source metadata key missing"
    assert "env" in body, "Env metadata key missing"
    
    source_value = decode_protobuf_any_value(body["source"])
    env_value = decode_protobuf_any_value(body["env"])
    
    assert source_value == "unit-test", f"Source value mismatch: {source_value}"
    assert env_value == "test", f"Env value mismatch: {env_value}"

def test_relationship_query():
    """Test relationship query via POST /relations."""
    print("\n🔍 Testing relationship filtering...")
    url = f"{QUERY_API_URL}/{ENTITY_ID}/relations"
    payload = {
        "relatedEntityId": RELATED_ID_1,
        "startTime": "2024-01-01T00:00:00Z",
        "endTime": "2024-12-31T23:59:59Z",
        "id": "rel-001",
        "name": "linked"
    }
    res = requests.post(url, json=payload)
    assert res.status_code == 200, f"Failed to get relationships: {res.text}"
    
    body = res.json()
    # Add relationship response validation
    assert isinstance(body, list), "Relationship response should be a list"
    assert len(body) > 0, "Expected at least one relationship"
    
    relationship = body[0]
    assert "relatedEntityId" in relationship, "Relationship should have relatedEntityId"
    assert relationship["relatedEntityId"] == RELATED_ID_1, "Related entity ID mismatch"
    assert relationship["name"] == "linked", "Relationship name mismatch"
    assert relationship["id"] == "rel-001", "Relationship ID mismatch"
    print("✅ Relationship response:", json.dumps(res.json(), indent=2))

def test_relationship_query_associated():
    """Test relationship query for 'associated' relationships with a specific start time."""
    print("\n🔍 Testing relationship filtering for 'associated' relationships...")
    
    # Define the API endpoint and payload
    url = f"{QUERY_API_URL}/{ENTITY_ID}/relations"
    payload = {
        "relatedEntityId": "",
        "startTime": "2024-02-01T00:00:00Z",  # Start time filter
        "endTime": "",
        "id": "",
        "name": "associated"  # Relationship name filter
    }
    
    # Send the POST request
    res = requests.post(url, json=payload)
    assert res.status_code == 200, f"Failed to get relationships: {res.text}"
    
    # Parse the response
    body = res.json()
    assert isinstance(body, list), "Relationship response should be a list"
    assert len(body) == 1, f"Expected exactly one relationship, got {len(body)}"
    
    # Validate the returned relationship
    relationship = body[0]
    assert "relatedEntityId" in relationship, "Relationship should have relatedEntityId"
    assert relationship["relatedEntityId"] == RELATED_ID_3, "Related entity ID mismatch"
    assert relationship["name"] == "associated", "Relationship name mismatch"
    assert relationship["startTime"] == "2024-01-01T00:00:00Z", "Start time mismatch"
    assert relationship["id"] == "rel-003", "Relationship ID mismatch"
    
    print("✅ Relationship response for 'associated':", json.dumps(body, indent=2))

def test_relationship_query_linked():
    """Test relationship query for 'linked' relationships with a specific start time."""
    print("\n🔍 Testing relationship filtering for 'linked' relationships...")
    
    # Define the API endpoint and payload
    url = f"{QUERY_API_URL}/{ENTITY_ID}/relations"
    payload = {
        "relatedEntityId": "",
        "startTime": "2024-02-01T00:00:00Z",  # Start time filter
        "endTime": "",
        "id": "",
        "name": "linked"  # Relationship name filter
    }
    
    # Send the POST request
    res = requests.post(url, json=payload)
    assert res.status_code == 200, f"Failed to get relationships: {res.text}"
    
    # Parse the response
    body = res.json()
    assert isinstance(body, list), "Relationship response should be a list"
    assert len(body) == 1, f"Expected exactly one relationship, got {len(body)}"
    
    # Validate the returned relationship
    relationship = body[0]
    assert "relatedEntityId" in relationship, "Relationship should have relatedEntityId"
    assert relationship["relatedEntityId"] == RELATED_ID_1, "Related entity ID mismatch"
    assert relationship["name"] == "linked", "Relationship name mismatch"
    assert relationship["startTime"] == "2024-01-01T00:00:00Z", "Start time mismatch"
    assert relationship["id"] == "rel-001", "Relationship ID mismatch"
    
    print("✅ Relationship response for 'linked':", json.dumps(body, indent=2))

def test_allrelationships_query():
    """Test relationship query without a payload to retrieve all relationships."""
    print("\n🔍 Testing relationship retrieval without a payload...")
    
    # Define the API endpoint
    url = f"{QUERY_API_URL}/{ENTITY_ID}/allrelations"
    
    # Send the POST request without a payload
    res = requests.post(url)
    assert res.status_code == 200, f"Failed to get relationships: {res.text}"
    
    # Parse the response
    body = res.json()
    assert isinstance(body, list), "Relationship response should be a list"
    assert len(body) == 3, f"Expected exactly 3 relationships, got {len(body)}"
    
    # Expected relationships for validation
    expected_relationships = [
        {
            "relatedEntityId": RELATED_ID_1,
            "name": "linked",
            "id": "rel-001",
            "startTime": "2024-01-01T00:00:00Z",
            "endTime": "2024-12-31T23:59:59Z"
        },
        {
            "relatedEntityId": RELATED_ID_2,
            "name": "linked",
            "id": "rel-002",
            "startTime": "2024-06-01T00:00:00Z",
            "endTime": "2024-12-31T23:59:59Z"
        },
        {
            "relatedEntityId": RELATED_ID_3,
            "name": "associated",
            "id": "rel-003",
            "startTime": "2024-01-01T00:00:00Z",
            "endTime": "2024-12-31T23:59:59Z"
        }
    ]
    
    # Validate all relationships
    for expected in expected_relationships:
        matching_relationships = [
            rel for rel in body if rel["id"] == expected["id"]
        ]
        assert len(matching_relationships) == 1, f"Expected exactly one match for relationship ID {expected['id']}"
        relationship = matching_relationships[0]
        assert relationship["relatedEntityId"] == expected["relatedEntityId"], f"Related entity ID mismatch for {expected['id']}"
        assert relationship["name"] == expected["name"], f"Relationship name mismatch for {expected['id']}"
        assert relationship["id"] == expected["id"], f"Relationship ID mismatch for {expected['id']}"
        assert relationship["startTime"] == expected["startTime"], f"Start time mismatch for {expected['id']}"
        assert relationship["endTime"] == expected["endTime"], f"End time mismatch for {expected['id']}"
    
    print("✅ All relationships retrieved successfully without a payload.")

def test_entity_search():
    """Test search by entity ID."""
    print("\n🔍 Testing entity search...")
    url = f"{QUERY_API_URL}/search"
    payload = {
        "id": ENTITY_ID,
        "created": "",
        "terminated": ""
    }
    res = requests.post(url, json=payload)
    assert res.status_code == 200, f"Search failed: {res.text}"
    
    body = res.json()
    # Add search response validation
    ## FIXME: Make sure to implement the entities/search and update this test case
    assert isinstance(body, dict), "Search response should be a dictionary"
    assert "body" in body, "Search response should have a 'body' field"
    assert isinstance(body["body"], list), "Search response body should be a list"
    assert len(body["body"]) == 0, "Expected an empty list in search response"


if __name__ == "__main__":
    print("🚀 Running Query API E2E Tests...")

    try:
        create_entity_for_query()
        test_attribute_lookup()
        test_metadata_lookup()
        test_relationship_query()
        test_relationship_query_associated()
        test_relationship_query_linked()
        test_allrelationships_query()
        test_entity_search()
        print("\n🎉 All Query API tests passed!")
    except AssertionError as e:
        print(f"\n❌ Test failed: {e}")
        sys.exit(1)
