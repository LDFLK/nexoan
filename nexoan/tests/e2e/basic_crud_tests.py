import requests
import json
import sys
import base64
import os
import unittest

"""
This file contains the end-to-end tests for the CRUD API.
It is used to test the API's functionality by creating, reading, updating, and deleting an entity.

The current tests only contain metadata validation.

Running the tests:

## Run CRUD Server

```bash
cd nexoan/crud-api
./crud-server
```
## Run API Server

```bash
cd nexoan/update-api
bal run
```

## Run Tests

```bash
cd nexoan/tests/e2e
python3 basic_crud_tests.py
```

"""

class CrudTestUtils:

    @staticmethod
    def decode_protobuf_any_value(any_value):
        """Decode a protobuf Any value to get the actual string value"""
        if isinstance(any_value, dict) and 'typeUrl' in any_value and 'value' in any_value:
            if 'StringValue' in any_value['typeUrl']:
                try:
                    # First try direct base64 decoding if that's how it's encoded
                    try:
                        binary_data = base64.b64decode(any_value['value'])
                        # For StringValue, skip the field tag byte and length byte
                        # and decode the remaining bytes as UTF-8
                        return binary_data[2:].decode('utf-8')
                    except:
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
        # Return the original value if decoding fails
        return any_value.strip()

class TestCRUDAPI(unittest.TestCase):
    def setUp(self):
        print("🟢 Setting up test environment...")
        update_service_url = os.getenv('UPDATE_SERVICE_URL', f"http://0.0.0.0:8080")
        print("🟢 UPDATE_SERVICE_URL: ", update_service_url)
        self.base_url = f"{update_service_url}/entities"
        print("🟢 BASE_URL: ", self.base_url)

class BasicCRUDTests:

    def __init__(self, entity_id):
        self.entity_id = entity_id
        self.base_url = get_base_url()
        self.headers = {
            'Content-Type': 'application/json'
        }
        self.payload = self.create_payload()

    def create_payload(self):
        """Returns the entity payload for create and update operations."""
        return {
            "create": {
                "id": self.entity_id,
                "kind": {"major": "example", "minor": "test"},
                "created": "2024-03-17T10:00:00Z",
                "terminated": "",
                "name": {
                    "startTime": "2024-03-17T10:00:00Z",
                    "endTime": "",
                    "value": "entity-name"
                },
                "metadata": [
                    {"key": "owner", "value": "test-user"},
                    {"key": "version", "value": "1.0"},
                    {"key": "developer", "value": "V8A"}
                ],
                "attributes": [],
                "relationships": []
            },
            "update": {
                "id": self.entity_id,
                "kind": {"major": "example", "minor": "test"},
                "created": "2024-03-18T00:00:00Z",
                "name": {
                    "startTime": "2024-03-18T00:00:00Z",
                    "value": "entity-name"
                },
                "metadata": [{"key": "version", "value": "5.0"}]
            }
        }


class MetadataValidationTests(BasicCRUDTests):

    def __init__(self, entity_id):
        super().__init__(entity_id)

    def create_entity(self):
        """Creates an entity and validates the response."""
        print("\n🟢 Creating entity...")
        response = requests.post(self.base_url, json=self.payload["create"], headers={"Content-Type": "application/json"})
        
        if response.status_code == 201:
            print("✅ Entity created:", json.dumps(response.json(), indent=2))
        else:
            print(f"❌ Create failed: {response.text}")
            sys.exit(1)

    def read_entity(self):
        """Reads and validates the created entity."""
        print("\n🟢 Reading entity...")
        response = requests.get(f"{self.base_url}/{self.entity_id}")
        
        if response.status_code == 200:
            data = response.json()
            assert data["id"] == self.entity_id, "Read entity ID mismatch"
            print("✅ Read Entity:", json.dumps(data, indent=2))
        else:
            print(f"❌ Read failed: {response.text}")
            sys.exit(1)

    def update_entity(self):
        """Updates the entity and validates the response."""
        print("\n🟢 Updating entity...")
        response = requests.put(f"{self.base_url}/{self.entity_id}", json=self.payload["update"], headers={"Content-Type": "application/json"})
        
        if response.status_code == 200:
            updated_entity = response.json()
            decoded_value = CrudTestUtils.decode_protobuf_any_value(updated_entity["metadata"][0]["value"])
            print("decoded value: ", decoded_value)
            assert decoded_value == "5.0", "Update did not modify metadata"
            print("✅ Entity updated:", json.dumps(updated_entity, indent=2))
        else:
            print(f"❌ Update failed: {response.text}")
            sys.exit(1)

    def validate_update(self):
        """Validates that the update has been applied correctly."""
        print("\n🟢 Validating update...")
        response = requests.get(f"{self.base_url}/{self.entity_id}")
        
        if response.status_code == 200:
            updated_data = response.json()
            decoded_value = CrudTestUtils.decode_protobuf_any_value(updated_data["metadata"][0]["value"])
            assert decoded_value == "5.0", "Updated entity does not reflect changes"
            print("✅ Update Validation Passed:", json.dumps(updated_data, indent=2))
        else:
            print(f"❌ Read failed after update: {response.text}")
            sys.exit(1)

    def delete_entity(self):
        """Deletes the entity."""
        print("\n🟢 Deleting entity...")
        response = requests.delete(f"{self.base_url}/{self.entity_id}")
        
        if response.status_code == 204:
            print("✅ Entity deleted successfully.")
        else:
            print(f"❌ Delete failed: {response.text}")
            sys.exit(1)

    def verify_deletion(self):
        """Verifies that the entity has been deleted."""
        print("\n🟢 Verifying deletion...")
        response = requests.get(f"{self.base_url}/{self.entity_id}")
        
        if response.status_code == 500:
            print("❌ Server error occurred:", response.text)
            sys.exit(1)
        else:
            print(f"\n🟢 Entity was not deleted properly: {response.status_code} {response.text}")


class GraphEntityTests(BasicCRUDTests):

    def __init__(self):
        super().__init__(None)
        self.MINISTER_ID = "minister_education"
        self.DEPARTMENTS = [
            {"id": "dept_exams", "name": "Department of Exams"},
            {"id": "dept_nie", "name": "National Institute of Education"},
            {"id": "dept_ed_publications", "name": "Department of Educational Publications"}
        ]
        self.START_DATE = "2015-04-11T00:00:00Z"


    def create_minister(self):
        """Create a Minister entity."""
        print("\n🟢 Creating Minister entity...")
        
        payload = {
            "id": self.MINISTER_ID,
            "kind": {"major": "Organization", "minor": "Minister"},
            "created": self.START_DATE,
            "terminated": "",
            "name": {
                "startTime": self.START_DATE,
                "endTime": "",
                "value": "Minister of Education"
            },
            "metadata": [],
            "attributes": [],
            "relationships": []
        }
        
        res = requests.post(self.base_url, json=payload)
        print(res.status_code, res.json())
        assert res.status_code in [201], f"Failed to create Minister: {res.text}"

        print(f"Response: {res.status_code} - {res.text}")
        print("✅ Created Minister entity.")


    def read_minister(self):
        """Read the Minister entity."""
        print("\n🟢 Reading Minister entity...")
        res = requests.get(f"{self.base_url}/{self.MINISTER_ID}")
        print(res.status_code, res.json())
        assert res.status_code in [200], f"Failed to read Minister: {res.text}"
        
        # Verify the response data
        response_data = res.json()
        assert response_data["id"] == self.MINISTER_ID, f"Expected ID {self.MINISTER_ID}, got {response_data['id']}"
        assert response_data["kind"]["major"] == "Organization", f"Expected major kind 'Organization', got {response_data['kind']['major']}"
        assert response_data["kind"]["minor"] == "Minister", f"Expected minor kind 'Minister', got {response_data['kind']['minor']}"
        assert response_data["created"] == self.START_DATE, f"Expected created date {self.START_DATE}, got {response_data['created']}"
        # The name value is a protobuf Any that needs to be decoded
        name_value = response_data["name"]["value"]
        decoded_name = CrudTestUtils.decode_protobuf_any_value(name_value)
        assert decoded_name == "Minister of Education", f"Expected name 'Minister of Education', got {name_value}"
        print(f"✅ Validated {decoded_name} entity.")


    def create_departments(self):
        """Create Department entities."""
        print("\n🟢 Creating Department entities...")
        
        for dept in self.DEPARTMENTS:
            payload = {
                "id": dept["id"],
                "kind": {"major": "Organization", "minor": "Department"},
                "created": self.START_DATE,
                "terminated": "",
                "name": {
                    "startTime": self.START_DATE,
                    "endTime": "",
                    "value": dept["name"]
                },
                "metadata": []
            }
            
            res = requests.post(self.base_url, json=payload)
            assert res.status_code in [200, 201], f"Failed to create {dept['name']}: {res.text}"
            print(f"Response: {res.status_code} - {res.text}")
            print(f"✅ Created {dept['name']} entity.")


    def read_departments(self):
        """Validate the Department entities in Neo4j."""
        print("\n🟢 Validating Department entities in Neo4j...")
        
        for dept in self.DEPARTMENTS:
            res = requests.get(f"{self.base_url}/{dept['id']}")
            assert res.status_code == 200, f"Failed to read {dept['name']}: {res.text}"
            
            # Verify the response data
            response_data = res.json()
            assert response_data["id"] == dept["id"], f"Expected ID {dept['id']}, got {response_data['id']}"
            assert response_data["kind"]["major"] == "Organization", f"Expected major kind 'Organization', got {response_data['kind']['major']}"
            assert response_data["kind"]["minor"] == "Department", f"Expected minor kind 'Department', got {response_data['kind']['minor']}"
            assert response_data["created"] == self.START_DATE, f"Expected created date {self.START_DATE}, got {response_data['created']}"
            
            # The name value is a protobuf Any that needs to be decoded
            name_value = response_data["name"]["value"]
            decoded_name = CrudTestUtils.decode_protobuf_any_value(name_value)
            assert decoded_name == dept["name"], f"Expected name '{dept['name']}', got {decoded_name}"
            
            print(f"✅ Validated {dept['name']} entity.")
    
    
    def create_relationships(self):
        """Create HAS_DEPARTMENT relationships from Minister to Departments."""
        print("\n🔗 Creating relationships...")
        
        for dept in self.DEPARTMENTS:
            rel_id = f"rel_{dept['id']}"
            payload = {
                "id": self.MINISTER_ID,
                "kind": {},
                "created": "",
                "terminated": "",
                "name": {
                },
                "metadata": [],
                "attributes": [],
                "relationships": [
                    {
                        "key": "HAS_DEPARTMENT",
                        "value": {
                            "relatedEntityId": dept["id"],
                            "startTime": self.START_DATE,
                            "endTime": "",
                            "id": rel_id,
                            "name": "HAS_DEPARTMENT"
                        }
                    }
                ]
            }
            
            url = f"{self.base_url}/{self.MINISTER_ID}"
            res = requests.put(url, json=payload)

            if res.status_code in [200]:
                print(f"✅ Created relationship between Minister and {dept['name']}.")
            else:
                print(f"❌ Failed to create relationship for {dept['name']}: {res.status_code} - {res.text}")
                sys.exit(1)

    def update_relationships(self):
        """Update HAS_DEPARTMENT relationships to add termination dates."""
        print("\n🔗 Updating relationships...")
        
        # Set a termination date for all relationships
        termination_date = "2024-12-31T00:00:00Z"
        
        for dept in self.DEPARTMENTS:
            rel_id = f"rel_{dept['id']}"
            payload = {
                "id": self.MINISTER_ID,
                "kind": {},
                "name": {},
                "relationships": [
                    {
                        "key": rel_id,
                        "value": {
                            "endTime": termination_date,
                            "id": rel_id,
                            "relatedEntityId": "",
                            "startTime": "",
                            "name": ""
                        }
                    }
                ]
            }
            
            url = f"{self.base_url}/{self.MINISTER_ID}"
            res = requests.put(url, json=payload)

            if res.status_code in [200]:
                print(f"✅ Updated relationship between Minister and {dept['name']} with termination date {termination_date}.")
            else:
                print(f"❌ Failed to update relationship for {dept['name']}: {res.status_code} - {res.text}")
                sys.exit(1)

        # Verify the updates
        print("\n🔍 Verifying relationship updates...")
        res = requests.get(f"{self.base_url}/{self.MINISTER_ID}")
        if res.status_code == 200:
            data = res.json()
            relationships = data.get("relationships", [])
            for rel in relationships:
                if rel.get("name") == "HAS_DEPARTMENT":
                    assert rel.get("endTime") == termination_date, f"Expected termination date {termination_date}, got {rel.get('endTime')}"
            print("✅ Successfully verified relationship updates.")
        else:
            print(f"❌ Failed to verify relationship updates: {res.status_code} - {res.text}")
            sys.exit(1)


def get_base_url():
    print("🟢 Setting up test environment...")
    update_service_url = os.getenv('UPDATE_SERVICE_URL', f"http://0.0.0.0:8080")
    print("🟢 UPDATE_SERVICE_URL: ", update_service_url)
    return f"{update_service_url}/entities"

if __name__ == "__main__":
    print("🚀 Running End-to-End API Test Suite...")
    
    try:
        print("🟢 Running Metadata Validation Tests...")
        metadata_validation_tests = MetadataValidationTests(entity_id="123")
        metadata_validation_tests.create_entity()
        metadata_validation_tests.read_entity()
        metadata_validation_tests.update_entity()
        metadata_validation_tests.validate_update()
        metadata_validation_tests.delete_entity()
        metadata_validation_tests.verify_deletion()
        print("\n🟢 Running Metadata Validation Tests... Done")

        print("\n🟢 Running Graph Entity Tests...")
        graph_entity_tests = GraphEntityTests()
        graph_entity_tests.create_minister()
        graph_entity_tests.read_minister()
        graph_entity_tests.create_departments()
        graph_entity_tests.read_departments()
        graph_entity_tests.create_relationships()
        graph_entity_tests.update_relationships()
        print("\n🟢 Running Graph Entity Tests... Done")

        print("\n🎉 All tests passed successfully!")
    
    except AssertionError as e:
        print(f"\n❌ Test failed: {e}")
        sys.exit(1)
