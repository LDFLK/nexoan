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

# Constants for government organization test
GOVERNMENT_ID = "gov-lk-001"
MINISTER_ID_1 = "minister-tech-001"
MINISTER_ID_2 = "minister-health-001"
DEPT_ID_1 = "dept-it-001"
DEPT_ID_2 = "dept-digital-001"
DEPT_ID_3 = "dept-hospitals-001"
DEPT_ID_4 = "dept-pharma-001"

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
            "value": "Query Test Entity Child 1"
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
            "value": "Query Test Entity Child 2"
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
            "value": "Query Test Entity Child 3"
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
            "value": "Query Test Entity"
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

def create_government_entities():
    """Create government organizational hierarchy for search tests."""
    print("\n🟢 Creating government organizational hierarchy...")

    # Create Government entity
    gov_payload = {
        "id": GOVERNMENT_ID,
        "kind": {"major": "Organization", "minor": "Government"},
        "created": "2024-01-01T00:00:00Z",
        "terminated": "",
        "name": {
            "startTime": "2024-01-01T00:00:00Z",
            "endTime": "",
            "value": "Government of Sri Lanka"
        },
        "metadata": [],
        "attributes": [],
        "relationships": [
            {
                "key": "minister-rel-1",
                "value": {
                    "relatedEntityId": MINISTER_ID_1,
                    "startTime": "2024-01-01T00:00:00Z",
                    "endTime": "",
                    "id": "gov-rel-001",
                    "name": "has_minister"
                }
            },
            {
                "key": "minister-rel-2",
                "value": {
                    "relatedEntityId": MINISTER_ID_2,
                    "startTime": "2024-01-01T00:00:00Z",
                    "endTime": "",
                    "id": "gov-rel-002",
                    "name": "has_minister"
                }
            }
        ]
    }

    # Create Technology Minister entity
    tech_minister_payload = {
        "id": MINISTER_ID_1,
        "kind": {"major": "Organization", "minor": "Minister"},
        "created": "2024-01-01T00:00:00Z",
        "terminated": "",
        "name": {
            "startTime": "2024-01-01T00:00:00Z",
            "endTime": "",
            "value": "Ministry of Technology"
        },
        "metadata": [],
        "attributes": [],
        "relationships": [
            {
                "key": "dept-rel-1",
                "value": {
                    "relatedEntityId": DEPT_ID_1,
                    "startTime": "2024-01-01T00:00:00Z",
                    "endTime": "",
                    "id": "minister-rel-001",
                    "name": "has_department"
                }
            },
            {
                "key": "dept-rel-2",
                "value": {
                    "relatedEntityId": DEPT_ID_2,
                    "startTime": "2024-01-01T00:00:00Z",
                    "endTime": "",
                    "id": "minister-rel-002",
                    "name": "has_department"
                }
            }
        ]
    }

    # Create Health Minister entity
    health_minister_payload = {
        "id": MINISTER_ID_2,
        "kind": {"major": "Organization", "minor": "Minister"},
        "created": "2024-01-01T00:00:00Z",
        "terminated": "",
        "name": {
            "startTime": "2024-01-01T00:00:00Z",
            "endTime": "",
            "value": "Ministry of Health"
        },
        "metadata": [],
        "attributes": [],
        "relationships": [
            {
                "key": "dept-rel-3",
                "value": {
                    "relatedEntityId": DEPT_ID_3,
                    "startTime": "2024-01-01T00:00:00Z",
                    "endTime": "",
                    "id": "minister-rel-003",
                    "name": "has_department"
                }
            },
            {
                "key": "dept-rel-4",
                "value": {
                    "relatedEntityId": DEPT_ID_4,
                    "startTime": "2024-01-01T00:00:00Z",
                    "endTime": "",
                    "id": "minister-rel-004",
                    "name": "has_department"
                }
            }
        ]
    }

    # Create Technology Department entities
    dept1_payload = {
        "id": DEPT_ID_1,
        "kind": {"major": "Organization", "minor": "Department"},
        "created": "2024-01-01T00:00:00Z",
        "terminated": "",
        "name": {
            "startTime": "2024-01-01T00:00:00Z",
            "endTime": "",
            "value": "IT Department"
        },
        "metadata": [],
        "attributes": [],
        "relationships": []
    }

    dept2_payload = {
        "id": DEPT_ID_2,
        "kind": {"major": "Organization", "minor": "Department"},
        "created": "2024-01-01T00:00:00Z",
        "terminated": "",
        "name": {
            "startTime": "2024-01-01T00:00:00Z",
            "endTime": "",
            "value": "Digital Services Department"
        },
        "metadata": [],
        "attributes": [],
        "relationships": []
    }

    # Create Health Department entities
    dept3_payload = {
        "id": DEPT_ID_3,
        "kind": {"major": "Organization", "minor": "Department"},
        "created": "2024-01-01T00:00:00Z",
        "terminated": "",
        "name": {
            "startTime": "2024-01-01T00:00:00Z",
            "endTime": "",
            "value": "Hospitals Department"
        },
        "metadata": [],
        "attributes": [],
        "relationships": []
    }

    dept4_payload = {
        "id": DEPT_ID_4,
        "kind": {"major": "Organization", "minor": "Department"},
        "created": "2024-01-01T00:00:00Z",
        "terminated": "",
        "name": {
            "startTime": "2024-01-01T00:00:00Z",
            "endTime": "",
            "value": "Pharmaceutical Department"
        },
        "metadata": [],
        "attributes": [],
        "relationships": []
    }

    # Create all entities
    # Create departments first
    for payload in [dept1_payload, dept2_payload, dept3_payload, dept4_payload]:
        res = requests.post(UPDATE_API_URL, json=payload)
        assert res.status_code in [201, 200], f"Failed to create entity: {res.text}"
        print(f"✅ Created {payload['kind']['minor']} entity: {payload['id']}")

    # Then create ministers
    for payload in [tech_minister_payload, health_minister_payload]:
        res = requests.post(UPDATE_API_URL, json=payload)
        assert res.status_code in [201, 200], f"Failed to create entity: {res.text}"
        print(f"✅ Created {payload['kind']['minor']} entity: {payload['id']}")

    # Finally create government
    res = requests.post(UPDATE_API_URL, json=gov_payload)
    assert res.status_code in [201, 200], f"Failed to create entity: {res.text}"
    print(f"✅ Created {gov_payload['kind']['minor']} entity: {gov_payload['id']}")

def test_search_without_major_kind_or_id():
    """Test that search fails when major kind or id is not provided."""
    print("\n🔍 Testing search without major kind...")
    url = f"{QUERY_API_URL}/search"
    payload = {
        "kind": {
            "minor": "Department"  # Only providing minor kind
        }
    }
    res = requests.post(url, json=payload)
    assert res.status_code == 400, f"Search should fail without major kind: {res.text}"
    
    body = res.json()
    assert isinstance(body, dict), "Error response should be a dictionary"
    assert "error" in body, "Error response should contain error message"
    print("✅ Search correctly failed without major kind")

def test_search_by_kind_major():
    """Test searching entities by major kind."""
    print("\n🔍 Testing search by major kind...")
    url = f"{QUERY_API_URL}/search"
    payload = {
        "kind": {
            "major": "Organization"
        }
    }
    res = requests.post(url, json=payload)
    assert res.status_code == 200, f"Search failed: {res.text}"
    
    body = res.json()
    assert isinstance(body, dict), "Search response should be a dictionary"
    assert "body" in body, "Search response should have a 'body' field"
    assert isinstance(body["body"], list), "Search response body should be a list"
    assert len(body["body"]) == 7, "Expected 7 organizations in search response"
    
    # Verify all returned entities are of major kind "Organization"
    for entity in body["body"]:
        assert entity["kind"]["major"] == "Organization", f"Expected major kind 'Organization', got {entity['kind']['major']}"
    
    print("✅ Search by major kind successful")

def test_search_by_kind_minor():
    """Test searching entities by minor kind."""
    print("\n🔍 Testing search by minor kind...")
    url = f"{QUERY_API_URL}/search"
    payload = {
        "kind": {
            "major": "Organization",  # Adding compulsory major kind
            "minor": "Department"
        }
    }
    res = requests.post(url, json=payload)
    assert res.status_code == 200, f"Search failed: {res.text}"
    
    body = res.json()
    assert isinstance(body, dict), "Search response should be a dictionary"
    assert "body" in body, "Search response should have a 'body' field"
    assert isinstance(body["body"], list), "Search response body should be a list"
    assert len(body["body"]) == 4, "Expected 4 departments in search response"
    
    # Verify all returned entities are departments
    for entity in body["body"]:
        assert entity["kind"]["minor"] == "Department", f"Expected minor kind 'Department', got {entity['kind']['minor']}"
    
    print("✅ Search by minor kind successful")

def test_search_by_name():
    """Test searching entities by name."""
    print("\n🔍 Testing search by name...")
    url = f"{QUERY_API_URL}/search"
    payload = {
        "kind": {
            "major": "Organization"  # Adding compulsory major kind
        },
        "name": "Ministry of Technology"
    }
    res = requests.post(url, json=payload)
    assert res.status_code == 200, f"Search failed: {res.text}"

    print(res.text)
    
    body = res.json()
    assert isinstance(body, dict), "Search response should be a dictionary"
    assert "body" in body, "Search response should have a 'body' field"
    assert isinstance(body["body"], list), "Search response body should be a list"
    assert len(body["body"]) == 1, "Expected 1 entity in search response"
    
    # Verify the returned entity is the Technology Minister
    entity = body["body"][0]
    assert entity["id"] == MINISTER_ID_1, f"Expected minister ID {MINISTER_ID_1}, got {entity['id']}"
    assert entity["kind"]["minor"] == "Minister", f"Expected minor kind 'Minister', got {entity['kind']['minor']}"
    
    print("✅ Search by name successful")

def test_search_by_created_date():
    """Test searching entities by creation date."""
    print("\n🔍 Testing search by creation date...")
    url = f"{QUERY_API_URL}/search"
    payload = {
        "kind": {
            "major": "Organization"  # Adding compulsory major kind
        },
        "created": "2024-01-01T00:00:00Z"
    }
    res = requests.post(url, json=payload)
    assert res.status_code == 200, f"Search failed: {res.text}"
    
    body = res.json()
    assert isinstance(body, dict), "Search response should be a dictionary"
    assert "body" in body, "Search response should have a 'body' field"
    assert isinstance(body["body"], list), "Search response body should be a list"
    assert len(body["body"]) == 7, "Expected 7 entities created on the same date"
    
    # Verify all returned entities have the same creation date
    for entity in body["body"]:
        assert entity["created"] == "2024-01-01T00:00:00Z", f"Expected creation date '2024-01-01T00:00:00Z', got {entity['created']}"
    
    print("✅ Search by creation date successful")

def test_search_by_name_and_kind():
    """Test searching entities by both name and kind."""
    print("\n🔍 Testing search by name and kind...")
    url = f"{QUERY_API_URL}/search"
    payload = {
        "kind": {
            "major": "Organization",
            "minor": "Minister"
        },
        "name": "Ministry of Technology"
    }
    res = requests.post(url, json=payload)
    assert res.status_code == 200, f"Search failed: {res.text}"
    
    body = res.json()
    assert isinstance(body, dict), "Search response should be a dictionary"
    assert "body" in body, "Search response should have a 'body' field"
    assert isinstance(body["body"], list), "Search response body should be a list"
    assert len(body["body"]) == 1, "Expected 1 entity in search response"
    
    # Verify the returned entity matches both filters
    entity = body["body"][0]
    assert entity["id"] == MINISTER_ID_1, f"Expected minister ID {MINISTER_ID_1}, got {entity['id']}"
    assert entity["kind"]["minor"] == "Minister", f"Expected minor kind 'Minister', got {entity['kind']['minor']}"
    
    print("✅ Search by name and kind successful")

def test_search_by_kind_and_created_date():
    """Test searching entities by both kind and creation date."""
    print("\n🔍 Testing search by kind and creation date...")
    url = f"{QUERY_API_URL}/search"
    payload = {
        "kind": {
            "major": "Organization",
            "minor": "Department"
        },
        "created": "2024-01-01T00:00:00Z"
    }
    res = requests.post(url, json=payload)
    assert res.status_code == 200, f"Search failed: {res.text}"
    
    body = res.json()
    assert isinstance(body, dict), "Search response should be a dictionary"
    assert "body" in body, "Search response should have a 'body' field"
    assert isinstance(body["body"], list), "Search response body should be a list"
    assert len(body["body"]) == 4, "Expected 4 departments in search response"
    
    # Verify all returned entities match both filters
    for entity in body["body"]:
        assert entity["kind"]["minor"] == "Department", f"Expected minor kind 'Department', got {entity['kind']['minor']}"
        assert entity["created"] == "2024-01-01T00:00:00Z", f"Expected creation date '2024-01-01T00:00:00Z', got {entity['created']}"
    
    print("✅ Search by kind and creation date successful")

def test_search_by_name_kind_and_created_date():
    """Test searching entities by name, kind, and creation date."""
    print("\n🔍 Testing search by name, kind, and creation date...")
    url = f"{QUERY_API_URL}/search"
    payload = {
        "kind": {
            "major": "Organization",
            "minor": "Department"
        },
        "name": "IT Department",
        "created": "2024-01-01T00:00:00Z"
    }
    res = requests.post(url, json=payload)
    assert res.status_code == 200, f"Search failed: {res.text}"
    
    body = res.json()
    assert isinstance(body, dict), "Search response should be a dictionary"
    assert "body" in body, "Search response should have a 'body' field"
    assert isinstance(body["body"], list), "Search response body should be a list"
    assert len(body["body"]) == 1, "Expected 1 entity in search response"
    
    # Verify the returned entity matches all filters
    entity = body["body"][0]
    assert entity["id"] == DEPT_ID_1, f"Expected department ID {DEPT_ID_1}, got {entity['id']}"
    assert entity["kind"]["minor"] == "Department", f"Expected minor kind 'Department', got {entity['kind']['minor']}"
    assert entity["created"] == "2024-01-01T00:00:00Z", f"Expected creation date '2024-01-01T00:00:00Z', got {entity['created']}"
    
    print("✅ Search by name, kind, and creation date successful")

def test_search_by_name_partial_match():
    """Test that searching with a partial name match returns no results."""
    print("\n🔍 Testing search by partial name match...")
    url = f"{QUERY_API_URL}/search"
    payload = {
        "kind": {
            "major": "Organization"
        },
        "name": "Ministry"  # Partial name that should not match
    }
    res = requests.post(url, json=payload)
    assert res.status_code == 200, f"Search failed: {res.text}"
    
    body = res.json()
    assert isinstance(body, dict), "Search response should be a dictionary"
    assert "body" in body, "Search response should have a 'body' field"
    assert isinstance(body["body"], list), "Search response body should be a list"
    assert len(body["body"]) == 0, "Expected 0 results for partial name match"
    
    print("✅ Search correctly returned no results for partial name match")

def test_search_by_terminated_date():
    """Test searching entities by termination date."""
    print("\n🔍 Testing search by termination date...")
    url = f"{QUERY_API_URL}/search"
    payload = {
        "kind": {
            "major": "Organization"
        },
        "terminated": "2024-12-31T23:59:59Z"
    }
    res = requests.post(url, json=payload)
    assert res.status_code == 200, f"Search failed: {res.text}"
    
    body = res.json()
    assert isinstance(body, dict), "Search response should be a dictionary"
    assert "body" in body, "Search response should have a 'body' field"
    assert isinstance(body["body"], list), "Search response body should be a list"
    assert len(body["body"]) == 0, "Expected 0 terminated entities in search response"
    
    print("✅ Search by termination date successful")

def test_search_by_active_entities():
    """Test searching for active (non-terminated) entities."""
    print("\n🔍 Testing search for active entities...")
    url = f"{QUERY_API_URL}/search"
    payload = {
        "kind": {
            "major": "Organization"
        },
        "terminated": ""  # Empty terminated date means active entities
    }
    res = requests.post(url, json=payload)
    assert res.status_code == 200, f"Search failed: {res.text}"
    
    body = res.json()
    assert isinstance(body, dict), "Search response should be a dictionary"
    assert "body" in body, "Search response should have a 'body' field"
    assert isinstance(body["body"], list), "Search response body should be a list"
    assert len(body["body"]) == 7, "Expected 7 active entities in search response"
    
    # Verify all returned entities are active
    for entity in body["body"]:
        assert entity["terminated"] == "", f"Expected empty terminated date, got {entity['terminated']}"
    
    print("✅ Search for active entities successful")

def test_search_by_kind_and_terminated():
    """Test searching entities by both kind and termination status."""
    print("\n🔍 Testing search by kind and termination status...")
    url = f"{QUERY_API_URL}/search"
    payload = {
        "kind": {
            "major": "Organization",
            "minor": "Department"
        },
        "terminated": ""  # Active departments
    }
    res = requests.post(url, json=payload)
    assert res.status_code == 200, f"Search failed: {res.text}"
    
    body = res.json()
    assert isinstance(body, dict), "Search response should be a dictionary"
    assert "body" in body, "Search response should have a 'body' field"
    assert isinstance(body["body"], list), "Search response body should be a list"
    assert len(body["body"]) == 4, "Expected 4 active departments in search response"
    
    # Verify all returned entities are active departments
    for entity in body["body"]:
        assert entity["kind"]["minor"] == "Department", f"Expected minor kind 'Department', got {entity['kind']['minor']}"
        assert entity["terminated"] == "", f"Expected empty terminated date, got {entity['terminated']}"
    
    print("✅ Search by kind and termination status successful")

def test_search_by_name_kind_and_terminated():
    """Test searching entities by name, kind, and termination status."""
    print("\n🔍 Testing search by name, kind, and termination status...")
    url = f"{QUERY_API_URL}/search"
    payload = {
        "kind": {
            "major": "Organization",
            "minor": "Minister"
        },
        "name": "Ministry of Technology",
        "terminated": ""  # Active minister
    }
    res = requests.post(url, json=payload)
    assert res.status_code == 200, f"Search failed: {res.text}"
    
    body = res.json()
    assert isinstance(body, dict), "Search response should be a dictionary"
    assert "body" in body, "Search response should have a 'body' field"
    assert isinstance(body["body"], list), "Search response body should be a list"
    assert len(body["body"]) == 1, "Expected 1 active minister in search response"
    
    # Verify the returned entity matches all filters
    entity = body["body"][0]
    assert entity["id"] == MINISTER_ID_1, f"Expected minister ID {MINISTER_ID_1}, got {entity['id']}"
    assert entity["kind"]["minor"] == "Minister", f"Expected minor kind 'Minister', got {entity['kind']['minor']}"
    assert entity["terminated"] == "", f"Expected empty terminated date, got {entity['terminated']}"
    
    print("✅ Search by name, kind, and termination status successful")

def test_search_by_id():
    """Test searching entities by ID."""
    print("\n🔍 Testing search by ID...")
    url = f"{QUERY_API_URL}/search"
    payload = {
        "id": DEPT_ID_1,  # Using the IT Department ID
        # "kind": {
        #     "major": ""  # Required field
        # }
    }
    res = requests.post(url, json=payload)
    assert res.status_code == 200, f"Search failed: {res.text}"
    
    body = res.json()
    assert isinstance(body, dict), "Search response should be a dictionary"
    assert "body" in body, "Search response should have a 'body' field"
    assert isinstance(body["body"], list), "Search response body should be a list"
    assert len(body["body"]) == 1, "Expected exactly one entity in search response"
    
    # Verify the returned entity matches the ID
    entity = body["body"][0]
    assert entity["id"] == DEPT_ID_1, f"Expected department ID {DEPT_ID_1}, got {entity['id']}"
    assert entity["kind"]["minor"] == "Department", f"Expected minor kind 'Department', got {entity['kind']['minor']}"
        
    # Parse the name JSON string and decode the hex value directly
    name_obj = json.loads(entity["name"])
    hex_value = name_obj["value"]
    decoded_name = bytes.fromhex(hex_value).decode('utf-8')
    
    assert decoded_name == "IT Department", f"Expected name 'IT Department', got {decoded_name}"
    
    print("✅ Search by ID successful")

def test_search_by_id_not_found():
    """Test searching for a non-existent ID."""
    print("\n🔍 Testing search by non-existent ID...")
    url = f"{QUERY_API_URL}/search"
    payload = {
        "id": "non-existent-id",
        # "kind": {
        #     "major": ""  # Required field
        # }
    }
    res = requests.post(url, json=payload)
    assert res.status_code == 200, f"Search failed: {res.text}"
    
    body = res.json()
    assert isinstance(body, dict), "Search response should be a dictionary"
    assert "body" in body, "Search response should have a 'body' field"
    assert isinstance(body["body"], list), "Search response body should be a list"
    assert len(body["body"]) == 0, "Expected empty list for non-existent ID"
    
    print("✅ Search by non-existent ID returned empty results as expected")

def test_search_by_id_with_other_filters():
    """Test that other filters are ignored when searching by ID."""
    print("\n🔍 Testing search by ID with additional filters...")
    url = f"{QUERY_API_URL}/search"
    payload = {
        "id": DEPT_ID_1,
        "kind": {
            "major": "Organization", 
            "minor": "Minister"  # This should be ignored since we're searching by ID
        },
        "name": "Wrong Name",  # This should be ignored
        "created": "2023-01-01T00:00:00Z"  # This should be ignored
    }
    res = requests.post(url, json=payload)
    assert res.status_code == 200, f"Search failed: {res.text}"
    
    body = res.json()
    assert isinstance(body, dict), "Search response should be a dictionary"
    assert "body" in body, "Search response should have a 'body' field"
    assert isinstance(body["body"], list), "Search response body should be a list"
    assert len(body["body"]) == 1, "Expected exactly one entity in search response"
    
    # Verify the returned entity matches the ID despite other filters
    entity = body["body"][0]
    assert entity["id"] == DEPT_ID_1, f"Expected department ID {DEPT_ID_1}, got {entity['id']}"
    assert entity["kind"]["minor"] == "Department", f"Expected minor kind 'Department', got {entity['kind']['minor']}"
    
    # Decode the name value from protobuf Any type
    name_obj = json.loads(entity["name"])
    hex_value = name_obj["value"]
    decoded_name = bytes.fromhex(hex_value).decode('utf-8')
    assert decoded_name == "IT Department", f"Expected name 'IT Department', got {decoded_name}"
    
    print("✅ Search by ID ignored additional filters as expected")

def test_relations_no_filters():
    """Test /relations endpoint with no filters (should return all relationships for the entity)."""
    print("\n🔍 Testing /relations with no filters...")
    url = f"{QUERY_API_URL}/{ENTITY_ID}/relations"
    res = requests.post(url, json={})
    assert res.status_code == 200, f"Failed to get relationships: {res.text}"
    body = res.json()
    assert isinstance(body, list), "Response should be a list"
    assert len(body) >= 3, "Expected at least 3 relationships"
    print("✅ /relations with no filters:", json.dumps(body, indent=2))

def test_relations_filter_by_name():
    """Test /relations endpoint filtering by relationship name."""
    print("\n🔍 Testing /relations with filter by name...")
    url = f"{QUERY_API_URL}/{ENTITY_ID}/relations"
    payload = {"name": "linked"}
    res = requests.post(url, json=payload)
    assert res.status_code == 200, f"Failed to get relationships: {res.text}"
    body = res.json()
    assert all(rel["name"] == "linked" for rel in body), "All relationships should have name 'linked'"
    print("✅ /relations with name filter:", json.dumps(body, indent=2))

def test_relations_filter_by_related_entity_id():
    """Test /relations endpoint filtering by relatedEntityId."""
    print("\n🔍 Testing /relations with filter by relatedEntityId...")
    url = f"{QUERY_API_URL}/{ENTITY_ID}/relations"
    payload = {"relatedEntityId": RELATED_ID_1}
    res = requests.post(url, json=payload)
    assert res.status_code == 200, f"Failed to get relationships: {res.text}"
    body = res.json()
    assert all(rel["relatedEntityId"] == RELATED_ID_1 for rel in body), "All relationships should have the correct relatedEntityId"
    print("✅ /relations with relatedEntityId filter:", json.dumps(body, indent=2))

def test_relations_filter_by_start_time():
    """Test /relations endpoint filtering by startTime."""
    print("\n🔍 Testing /relations with filter by startTime...")
    url = f"{QUERY_API_URL}/{ENTITY_ID}/relations"
    payload = {"startTime": "2024-06-01T00:00:00Z"}
    res = requests.post(url, json=payload)
    assert res.status_code == 200, f"Failed to get relationships: {res.text}"
    body = res.json()
    assert all(rel["startTime"] == "2024-06-01T00:00:00Z" for rel in body), "All relationships should have the correct startTime"
    print("✅ /relations with startTime filter:", json.dumps(body, indent=2))

def test_relations_filter_by_end_time():
    """Test /relations endpoint filtering by endTime."""
    print("\n🔍 Testing /relations with filter by endTime...")
    url = f"{QUERY_API_URL}/{ENTITY_ID}/relations"
    payload = {"endTime": "2024-12-31T23:59:59Z"}
    res = requests.post(url, json=payload)
    assert res.status_code == 200, f"Failed to get relationships: {res.text}"
    body = res.json()
    assert all(rel["endTime"] == "2024-12-31T23:59:59Z" for rel in body), "All relationships should have the correct endTime"
    print("✅ /relations with endTime filter:", json.dumps(body, indent=2))

def test_relations_filter_by_multiple_fields():
    """Test /relations endpoint filtering by multiple fields."""
    print("\n🔍 Testing /relations with multiple filters...")
    url = f"{QUERY_API_URL}/{ENTITY_ID}/relations"
    payload = {
        "name": "linked",
        "relatedEntityId": RELATED_ID_2,
        "startTime": "2024-06-01T00:00:00Z"
    }
    res = requests.post(url, json=payload)
    assert res.status_code == 200, f"Failed to get relationships: {res.text}"
    body = res.json()
    assert len(body) == 1, "Expected exactly one relationship"
    rel = body[0]
    assert rel["name"] == "linked"
    assert rel["relatedEntityId"] == RELATED_ID_2
    assert rel["startTime"] == "2024-06-01T00:00:00Z"
    print("✅ /relations with multiple filters:", json.dumps(body, indent=2))

def test_relations_filter_nonexistent():
    """Test /relations endpoint with filters that match nothing."""
    print("\n🔍 Testing /relations with non-existent filter...")
    url = f"{QUERY_API_URL}/{ENTITY_ID}/relations"
    payload = {"name": "nonexistent"}
    res = requests.post(url, json=payload)
    assert res.status_code == 200, f"Failed to get relationships: {res.text}"
    body = res.json()
    assert isinstance(body, list), "Response should be a list"
    assert len(body) == 0, "Expected no relationships for non-existent filter"
    print("✅ /relations with non-existent filter returned empty list.")

def test_relations_filter_by_active_at():
    """Test /relations endpoint filtering by activeAt only."""
    print("\n🔍 Testing /relations with filter by activeAt...")
    url = f"{QUERY_API_URL}/{ENTITY_ID}/relations"
    payload = {"activeAt": "2024-07-01T00:00:00Z"}
    res = requests.post(url, json=payload)
    assert res.status_code == 200, f"Failed to get relationships: {res.text}"
    body = res.json()
    # Should return relationships active at this time
    assert isinstance(body, list), "Response should be a list"
    assert any(rel["id"] == "rel-001" for rel in body), "Expected rel-001 to be active at 2024-07-01T00:00:00Z"
    assert any(rel["id"] == "rel-002" for rel in body), "Expected rel-002 to be active at 2024-07-01T00:00:00Z"
    assert any(rel["id"] == "rel-003" for rel in body), "Expected rel-003 to be active at 2024-07-01T00:00:00Z"
    print("✅ /relations with activeAt filter:", json.dumps(body, indent=2))


def test_relations_filter_by_active_at_and_name():
    """Test /relations endpoint filtering by activeAt and name."""
    print("\n🔍 Testing /relations with filter by activeAt and name...")
    url = f"{QUERY_API_URL}/{ENTITY_ID}/relations"
    payload = {"activeAt": "2024-07-01T00:00:00Z", "name": "linked"}
    res = requests.post(url, json=payload)
    assert res.status_code == 200, f"Failed to get relationships: {res.text}"
    body = res.json()
    assert all(rel["name"] == "linked" for rel in body), "All relationships should have name 'linked'"
    assert any(rel["id"] == "rel-001" for rel in body), "Expected rel-001 to be present"
    assert any(rel["id"] == "rel-002" for rel in body), "Expected rel-002 to be present"
    print("✅ /relations with activeAt and name filter:", json.dumps(body, indent=2))


def test_relations_filter_by_active_at_and_related_entity_id():
    """Test /relations endpoint filtering by activeAt and relatedEntityId."""
    print("\n🔍 Testing /relations with filter by activeAt and relatedEntityId...")
    url = f"{QUERY_API_URL}/{ENTITY_ID}/relations"
    payload = {"activeAt": "2024-07-01T00:00:00Z", "relatedEntityId": RELATED_ID_1}
    res = requests.post(url, json=payload)
    assert res.status_code == 200, f"Failed to get relationships: {res.text}"
    body = res.json()
    assert all(rel["relatedEntityId"] == RELATED_ID_1 for rel in body), "All relationships should have the correct relatedEntityId"
    assert any(rel["id"] == "rel-001" for rel in body), "Expected rel-001 to be present"
    print("✅ /relations with activeAt and relatedEntityId filter:", json.dumps(body, indent=2))


def test_relations_filter_by_active_at_and_direction():
    """Test /relations endpoint filtering by activeAt and direction."""
    print("\n🔍 Testing /relations with filter by activeAt and direction...")
    url = f"{QUERY_API_URL}/{ENTITY_ID}/relations"
    payload = {"activeAt": "2024-07-01T00:00:00Z", "direction": "OUTGOING"}
    res = requests.post(url, json=payload)
    assert res.status_code == 200, f"Failed to get relationships: {res.text}"
    body = res.json()
    assert all(rel["direction"] == "OUTGOING" for rel in body), "All relationships should have direction 'OUTGOING'"
    print("✅ /relations with activeAt and direction filter:", json.dumps(body, indent=2))


def test_relations_filter_by_active_at_and_name_and_direction():
    """Test /relations endpoint filtering by activeAt, name, and direction."""
    print("\n🔍 Testing /relations with filter by activeAt, name, and direction...")
    url = f"{QUERY_API_URL}/{ENTITY_ID}/relations"
    payload = {"activeAt": "2024-07-01T00:00:00Z", "name": "linked", "direction": "OUTGOING"}
    res = requests.post(url, json=payload)
    assert res.status_code == 200, f"Failed to get relationships: {res.text}"
    body = res.json()
    assert all(rel["name"] == "linked" for rel in body), "All relationships should have name 'linked'"
    assert all(rel["direction"] == "OUTGOING" for rel in body), "All relationships should have direction 'OUTGOING'"
    print("✅ /relations with activeAt, name, and direction filter:", json.dumps(body, indent=2))


def test_relations_filter_by_active_at_and_time_range_invalid():
    """Test /relations endpoint with activeAt and startTime (should return 400)."""
    print("\n🔍 Testing /relations with activeAt and startTime (should fail)...")
    url = f"{QUERY_API_URL}/{ENTITY_ID}/relations"
    payload = {"activeAt": "2024-07-01T00:00:00Z", "startTime": "2024-01-01T00:00:00Z"}
    res = requests.post(url, json=payload)
    assert res.status_code == 400, f"Expected 400 Bad Request, got {res.status_code}: {res.text}"
    body = res.json()
    assert "error" in body, "Error message should be present in 400 response"
    print("✅ /relations with activeAt and startTime correctly failed:", json.dumps(body, indent=2))

def test_gov_relations_filter_by_active_at_and_direction():
    """Test /relations endpoint for government entity with activeAt and direction OUTGOING."""
    print("\n🔍 Testing /relations for government entity with activeAt and direction OUTGOING...")
    url = f"{QUERY_API_URL}/{GOVERNMENT_ID}/relations"
    payload = {"activeAt": "2024-07-01T00:00:00Z", "direction": "OUTGOING"}
    res = requests.post(url, json=payload)
    assert res.status_code == 200, f"Failed to get relationships: {res.text}"
    body = res.json()
    # Should return has_minister relationships
    assert all(rel["direction"] == "OUTGOING" for rel in body), "All relationships should be OUTGOING"
    assert all(rel["name"] == "has_minister" for rel in body), "All relationships should be has_minister"
    assert any(rel["relatedEntityId"] == MINISTER_ID_1 for rel in body), "Should include MINISTER_ID_1"
    assert any(rel["relatedEntityId"] == MINISTER_ID_2 for rel in body), "Should include MINISTER_ID_2"
    print("✅ Government OUTGOING has_minister relationships:", json.dumps(body, indent=2))

def test_minister_relations_filter_by_active_at_and_direction():
    """Test /relations endpoint for minister entity with activeAt and direction OUTGOING."""
    print("\n🔍 Testing /relations for minister entity with activeAt and direction OUTGOING...")
    url = f"{QUERY_API_URL}/{MINISTER_ID_1}/relations"
    payload = {"activeAt": "2024-07-01T00:00:00Z", "direction": "OUTGOING"}
    res = requests.post(url, json=payload)
    assert res.status_code == 200, f"Failed to get relationships: {res.text}"
    body = res.json()
    # Should return has_department relationships
    print("✅ Minister OUTGOING has_department relationships:", json.dumps(body, indent=2))
    assert all(rel["direction"] == "OUTGOING" for rel in body), "All relationships should be OUTGOING"
    assert all(rel["name"] == "has_department" for rel in body), "All relationships should be has_department"
    assert any(rel["relatedEntityId"] == DEPT_ID_1 for rel in body), "Should include DEPT_ID_1"
    assert any(rel["relatedEntityId"] == DEPT_ID_2 for rel in body), "Should include DEPT_ID_2"
    print("✅ Minister OUTGOING has_department relationships:", json.dumps(body, indent=2))

    print("\n🔍 Testing /relations for minister entity with activeAt and direction INCOMING...")
    payload = {"activeAt": "2024-07-01T00:00:00Z", "direction": "INCOMING"}
    res = requests.post(url, json=payload)
    assert res.status_code == 200, f"Failed to get relationships: {res.text}"
    body = res.json()
    # Should return has_minister relationship from government
    assert all(rel["direction"] == "INCOMING" for rel in body), "All relationships should be INCOMING"
    assert all(rel["name"] == "has_minister" for rel in body), "All relationships should be has_minister"
    assert any(rel["relatedEntityId"] == GOVERNMENT_ID for rel in body), "Should include GOVERNMENT_ID as relatedEntityId"
    print("✅ Minister INCOMING has_minister relationship:", json.dumps(body, indent=2))


def test_department_relations_filter_by_active_at_and_direction():
    """Test /relations endpoint for department entity with activeAt and direction INCOMING."""
    print("\n🔍 Testing /relations for department entity with activeAt and direction INCOMING...")
    url = f"{QUERY_API_URL}/{DEPT_ID_1}/relations"
    payload = {"activeAt": "2024-07-01T00:00:00Z", "direction": "INCOMING"}
    res = requests.post(url, json=payload)
    assert res.status_code == 200, f"Failed to get relationships: {res.text}"
    body = res.json()
    # Should return has_department relationship from minister
    assert all(rel["direction"] == "INCOMING" for rel in body), "All relationships should be INCOMING"
    assert all(rel["name"] == "has_department" for rel in body), "All relationships should be has_department"
    assert any(rel["relatedEntityId"] == MINISTER_ID_1 for rel in body), "Should include MINISTER_ID_1 as relatedEntityId"
    print("✅ Department INCOMING has_department relationship:", json.dumps(body, indent=2))

def test_minister_relations_filter_by_active_at_only():
    """Test /relations endpoint for minister entity with only activeAt (2024-01-02)."""
    print("\n🔍 Testing /relations for minister entity with only activeAt...")
    url = f"{QUERY_API_URL}/{MINISTER_ID_1}/relations"
    payload = {"activeAt": "2024-01-02T00:00:00Z"}
    res = requests.post(url, json=payload)
    assert res.status_code == 200, f"Failed to get relationships: {res.text}"
    body = res.json()
    # Should return both outgoing has_department and incoming has_minister relationships
    directions = set(rel["direction"] for rel in body)
    names = set(rel["name"] for rel in body)
    related_ids = set(rel["relatedEntityId"] for rel in body)
    print("✅ Minister relationships with only activeAt:", json.dumps(body, indent=2))
    assert "OUTGOING" in directions, "Should include OUTGOING relationships"
    assert "INCOMING" in directions, "Should include INCOMING relationships"
    assert "has_department" in names, "Should include has_department relationships"
    assert "has_minister" in names, "Should include has_minister relationships"
    assert MINISTER_ID_1 in related_ids or DEPT_ID_1 in related_ids or DEPT_ID_2 in related_ids or GOVERNMENT_ID in related_ids, "Should include expected related entity IDs"

if __name__ == "__main__":
    print("🚀 Running Query API E2E Tests...")

    try:
        create_entity_for_query()
        test_attribute_lookup()
        test_metadata_lookup()
        
        # Run government organization search tests
        create_government_entities()
        test_search_without_major_kind_or_id()
        test_search_by_kind_major()
        test_search_by_kind_minor()
        test_search_by_name()
        test_search_by_created_date()
        
        # Run ID-based search tests
        test_search_by_id()
        test_search_by_id_not_found()
        test_search_by_id_with_other_filters()
        
        # Run combined filter tests
        test_search_by_name_and_kind()
        test_search_by_kind_and_created_date()
        test_search_by_name_kind_and_created_date()
        test_search_by_name_partial_match()
        
        # Run terminated date filter tests
        test_search_by_terminated_date()
        test_search_by_active_entities()
        test_search_by_kind_and_terminated()
        test_search_by_name_kind_and_terminated()
        
        # Run relationship filter tests
        test_relations_no_filters()
        test_relations_filter_by_name()
        test_relations_filter_by_related_entity_id()
        test_relations_filter_by_start_time()
        test_relations_filter_by_end_time()
        test_relations_filter_by_multiple_fields()
        test_relations_filter_nonexistent()
        test_relations_filter_by_active_at()
        test_relations_filter_by_active_at_and_name()
        test_relations_filter_by_active_at_and_related_entity_id()
        test_relations_filter_by_active_at_and_direction()
        test_relations_filter_by_active_at_and_name_and_direction()
        test_relations_filter_by_active_at_and_time_range_invalid()
        test_gov_relations_filter_by_active_at_and_direction()
        test_minister_relations_filter_by_active_at_and_direction()
        test_department_relations_filter_by_active_at_and_direction()
        test_minister_relations_filter_by_active_at_only()
        
        print("\n🎉 All Query API tests passed!")
    except AssertionError as e:
        print(f"\n❌ Test failed: {e}")
        sys.exit(1)
