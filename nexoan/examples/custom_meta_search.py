import requests
import json
import os
import binascii
from datetime import datetime, timezone
from google.protobuf.wrappers_pb2 import StringValue

def get_service_urls():
    """Get service URLs from environment variables."""
    query_service_url = os.getenv('QUERY_SERVICE_URL', 'http://0.0.0.0:8081')
    update_service_url = os.getenv('UPDATE_SERVICE_URL', 'http://0.0.0.0:8080')
    
    return {
        'query': f"{query_service_url}/v1/entities",
        'update': f"{update_service_url}/entities"
    }

def decode_protobuf_any_value(any_value):
    """Decode a protobuf Any value to get the actual value"""
    if isinstance(any_value, str) and any_value.startswith('{"typeUrl"'):
        # It's a JSON string, parse it first
        try:
            any_value = json.loads(any_value)
        except json.JSONDecodeError:
            return any_value
    
    if isinstance(any_value, dict) and 'typeUrl' in any_value and 'value' in any_value:
        type_url = any_value['typeUrl']
        value = any_value['value']

        if 'Struct' in type_url:
            try:
                # For Struct type, the value is hex-encoded protobuf data
                binary_data = bytes.fromhex(value)
                
                # Try to use protobuf library first
                try:
                    from google.protobuf import struct_pb2
                    from google.protobuf.json_format import MessageToDict

                    struct_msg = struct_pb2.Struct()
                    struct_msg.ParseFromString(binary_data)
                    result = MessageToDict(struct_msg)

                    # The result should contain the actual JSON data
                    if isinstance(result, dict):
                        # Look for the actual data field
                        if 'data' in result and isinstance(result['data'], str):
                            try:
                                # The data field contains the JSON string
                                data_json = json.loads(result['data'])
                                return data_json
                            except json.JSONDecodeError:
                                pass
                        # If no 'data' field, return the result as is
                        return result

                except ImportError:
                    print("protobuf library not available, trying manual extraction")
                except Exception as e:
                    print(f"Failed to decode with protobuf library: {e}")

                # Manual extraction fallback - look for JSON patterns in the binary data
                try:
                    # Convert binary data to string and look for JSON
                    text_data = binary_data.decode('utf-8', errors='ignore')
                    print(f"Debug: Extracted text from binary: {repr(text_data[:200])}...")

                    # Find JSON-like content (look for { and })
                    start_idx = text_data.find('{')
                    end_idx = text_data.rfind('}')

                    if start_idx != -1 and end_idx != -1 and end_idx > start_idx:
                        json_str = text_data[start_idx:end_idx + 1]
                        print(f"Debug: Extracted JSON string: {repr(json_str[:100])}...")
                        try:
                            return json.loads(json_str)
                        except json.JSONDecodeError as e:
                            print(f"Debug: JSON decode failed: {e}")

                except Exception as e:
                    print(f"Failed to extract JSON from binary data: {e}")

                # Try a different approach - look for specific patterns
                try:
                    # The hex data might contain the JSON in a different format
                    text_data = binary_data.decode('utf-8', errors='ignore')

                    # Look for common JSON patterns
                    patterns = ['"columns"', '"rows"', '"data"']
                    for pattern in patterns:
                        if pattern in text_data:
                            # Find the start of the JSON object containing this pattern
                            start_idx = text_data.find('{', text_data.find(pattern) - 100)
                            if start_idx != -1:
                                # Find the matching closing brace
                                brace_count = 0
                                end_idx = start_idx
                                for i, char in enumerate(text_data[start_idx:], start_idx):
                                    if char == '{':
                                        brace_count += 1
                                    elif char == '}':
                                        brace_count -= 1
                                        if brace_count == 0:
                                            end_idx = i
                                            break

                                if end_idx > start_idx:
                                    json_str = text_data[start_idx:end_idx + 1]
                                    print(f"Debug: Found JSON with pattern {pattern}: {repr(json_str[:100])}...")
                                    try:
                                        return json.loads(json_str)
                                    except json.JSONDecodeError as e:
                                        print(f"Debug: JSON decode failed: {e}")
                                        continue

                except Exception as e:
                    print(f"Failed to extract JSON with pattern matching: {e}")

                # Final fallback: return the hex value
                return value

            except Exception as e:
                print(f"Failed to decode Struct: {e}")
                return value

    # Return the original value if decoding fails
    return any_value

def get_attribute_display_name(entity_metadata, attribute_key):
    """
    Get the human-readable display name for an attribute from entity metadata.
    
    Args:
        entity_metadata: List of metadata key-value pairs from the entity
        attribute_key: The technical attribute key (e.g., 'personal_information')
    
    Returns:
        str: Human-readable display name or fallback to formatted key
    """
    if entity_metadata:
        for meta in entity_metadata:
            if meta.get('key') == attribute_key:
                return meta.get('value', attribute_key.replace('_', ' ').title())
    
    # Fallback to formatted key name
    return attribute_key.replace('_', ' ').title()

def display_tabular_data(data, attribute_name, entity_metadata=None):
    """
    Display tabular data in a nice format.
    
    Args:
        data: The decoded tabular data (dict with 'columns' and 'rows')
        attribute_name: Name of the attribute for display purposes
        entity_metadata: Optional metadata to get human-readable names
    """
    if not isinstance(data, dict) or 'columns' not in data or 'rows' not in data:
        print(f"    ⚠️ Unexpected data structure for {attribute_name}: {type(data)}")
        print(f"    Raw data: {data}")
        return
    
    try:
        # Get the human-readable display name from metadata if available
        if entity_metadata:
            title = get_attribute_display_name(entity_metadata, attribute_name)
        else:
            title = attribute_name.replace('_', ' ').title()
        
        # Display the title with nice formatting
        print(f"\n    📊 {title}")
        print("    " + "=" * (len(title) + 4))
        
        # Display basic info
        print(f"    📋 Shape: {len(data['rows'])} rows × {len(data['columns'])} columns")
        print(f"    📋 Columns: {data['columns']}")
        
        # Display the table
        print(f"\n    📊 Data Table:")
        for i, row in enumerate(data['rows']):
            print(f"      Row {i+1}: {row}")
        
        # Show intelligent analysis based on column structure
        if 'field' in data['columns'] and 'value' in data['columns']:
            print(f"\n    🔍 Key Information:")
            for row in data['rows']:
                if len(row) >= 2:
                    field = row[1] if len(row) > 1 else 'Unknown'
                    value = row[2] if len(row) > 2 else 'N/A'
                    print(f"      - {field}: {value}")
                    
        elif 'metric' in data['columns'] and 'target' in data['columns'] and 'actual' in data['columns']:
            print(f"\n    📈 Performance Analysis:")
            for row in data['rows']:
                if len(row) >= 5:
                    metric = row[3] if len(row) > 3 else 'Unknown'
                    target = row[4] if len(row) > 4 else 'N/A'
                    actual = row[0] if len(row) > 0 else 'N/A'
                    status = row[2] if len(row) > 2 else 'N/A'
                    print(f"      - {metric}: {actual}/{target} ({status})")
                    
        elif 'category' in data['columns'] and 'allocated_amount' in data['columns']:
            print(f"\n    💰 Budget Analysis:")
            total_allocated = 0
            total_spent = 0
            for row in data['rows']:
                if len(row) >= 5:
                    category = row[4] if len(row) > 4 else 'Unknown'
                    allocated = row[1] if len(row) > 1 and isinstance(row[1], (int, float)) else 0
                    spent = row[2] if len(row) > 2 and isinstance(row[2], (int, float)) else 0
                    if isinstance(allocated, (int, float)):
                        total_allocated += allocated
                    if isinstance(spent, (int, float)):
                        total_spent += spent
                    print(f"      - {category}: {allocated:,} allocated, {spent:,} spent")
            print(f"      - Total: {total_allocated:,} allocated, {total_spent:,} spent")
        
    except Exception as e:
        print(f"    ❌ Failed to display data: {e}")
        print(f"    Raw data: {data}")

def create_minister_entity_example():
    """
    Example: Create a Minister entity with rich attributes and metadata.
    This demonstrates how to create a government minister with tabular data.
    """
    print("🏛️ Creating Minister Entity Example")
    print("=" * 50)
    
    # Get service URLs
    urls = get_service_urls()
    update_url = urls['update']
    
    # Minister data
    minister_data = {
        "id": "minister-agriculture-001",
        "name": "Minister of Agriculture and Food Security",
        "short_name": "Agriculture Minister",
        "portfolio": "Agriculture",
        "appointment_date": "2024-01-15T00:00:00Z"
    }
    
    # Personal information table
    personal_info = {
        "columns": ["field", "value", "last_updated"],
        "rows": [
            ["full_name", minister_data["name"], "2024-01-15T00:00:00Z"],
            ["short_name", minister_data["short_name"], "2024-01-15T00:00:00Z"],
            ["portfolio", minister_data["portfolio"], "2024-01-15T00:00:00Z"],
            ["appointment_date", minister_data["appointment_date"], "2024-01-15T00:00:00Z"],
            ["office_location", "Ministry of Agriculture, Colombo", "2024-01-15T00:00:00Z"],
            ["contact_email", f"{minister_data['id']}@gov.lk", "2024-01-15T00:00:00Z"],
            ["security_clearance", "Confidential", "2024-01-15T00:00:00Z"],
            ["education_background", "PhD in Agricultural Sciences", "2024-01-15T00:00:00Z"],
            ["years_of_experience", "15", "2024-01-15T00:00:00Z"]
        ]
    }
    
    # Performance metrics table
    performance_metrics = {
        "columns": ["metric", "target", "actual", "period", "status"],
        "rows": [
            ["crop_yield_improvement", "20%", "18%", "Q1-2024", "On Track"],
            ["farmer_support_programs", "10", "8", "Q1-2024", "In Progress"],
            ["food_security_index", "85%", "82%", "Q1-2024", "Good"],
            ["rural_development_projects", "25", "22", "Q1-2024", "Good"],
            ["sustainable_farming_adoption", "60%", "55%", "Q1-2024", "Good"]
        ]
    }
    
    # Budget allocation table
    budget_allocation = {
        "columns": ["category", "allocated_amount", "spent_amount", "remaining", "fiscal_year"],
        "rows": [
            ["crop_subsidies", 150000000, 120000000, 30000000, "2024"],
            ["research_and_development", 80000000, 50000000, 30000000, "2024"],
            ["farmer_training_programs", 30000000, 25000000, 5000000, "2024"],
            ["infrastructure_development", 200000000, 150000000, 50000000, "2024"],
            ["emergency_food_reserve", 100000000, 20000000, 80000000, "2024"]
        ]
    }
    
    # Create the entity payload
    payload = {
        "id": minister_data["id"],
        "kind": {"major": "Organization", "minor": "Minister"},
        "created": minister_data["appointment_date"],
        "terminated": "",
        "name": {
            "startTime": minister_data["appointment_date"],
            "endTime": "",
            "value": minister_data["name"]
        },
        "metadata": [
            {"key": "portfolio", "value": minister_data["portfolio"]},
            {"key": "appointment_date", "value": minister_data["appointment_date"]},
            {"key": "entity_type", "value": "government_minister"},
            {"key": "hierarchy_level", "value": "minister"},
            {"key": "ministry", "value": "Ministry of Agriculture"},
            {"key": "responsibility", "value": "Food Security and Rural Development"},
            {"key": "reporting_to", "value": "Prime Minister"},
            # Attribute name mappings for human-readable display
            {"key": "personal_information", "value": "Personal Information of Minister"},
            {"key": "performance_metrics", "value": "Performance Metrics and KPIs"},
            {"key": "budget_allocation", "value": "Budget Allocation and Financial Planning"}
        ],
        "attributes": [
            {
                "key": "personal_information",
                "value": {
                    "values": [
                        {
                            "startTime": minister_data["appointment_date"],
                            "endTime": "",
                            "value": personal_info
                        }
                    ]
                }
            },
            {
                "key": "performance_metrics",
                "value": {
                    "values": [
                        {
                            "startTime": "2024-01-01T00:00:00Z",
                            "endTime": "",
                            "value": performance_metrics
                        }
                    ]
                }
            },
            {
                "key": "budget_allocation",
                "value": {
                    "values": [
                        {
                            "startTime": "2024-01-01T00:00:00Z",
                            "endTime": "2024-12-31T23:59:59Z",
                            "value": budget_allocation
                        }
                    ]
                }
            }
        ],
        "relationships": []
    }
    
    print(f"📋 Creating Minister: {minister_data['name']}")
    print(f"🏢 Portfolio: {minister_data['portfolio']}")
    print(f"📅 Appointment Date: {minister_data['appointment_date']}")
    
    # Send the request
    try:
        response = requests.post(update_url, json=payload)
        
        if response.status_code in [200, 201]:
            print(f"✅ Successfully created Minister: {minister_data['id']}")
            print(f"📊 Entity includes:")
            print(f"   - Personal Information: {len(personal_info['rows'])} records")
            print(f"   - Performance Metrics: {len(performance_metrics['rows'])} records")
            print(f"   - Budget Allocation: {len(budget_allocation['rows'])} records")
            print(f"   - Metadata: {len(payload['metadata'])} fields")
            return True
        else:
            print(f"❌ Failed to create Minister: {response.status_code}")
            print(f"Error: {response.text}")
            return False
            
    except Exception as e:
        print(f"❌ Error creating entity: {e}")
        return False

def query_minister_example():
    """
    Example: Query the created minister entity.
    """
    print("\n🔍 Querying Minister Entity Example")
    print("=" * 50)
    
    # Get service URLs
    urls = get_service_urls()
    query_url = urls['query']
    
    # Query for the specific minister
    search_url = f"{query_url}/search"
    payload = {
        "id": "minister-agriculture-001"
    }
    
    try:
        response = requests.post(search_url, json=payload)
        
        if response.status_code == 200:
            data = response.json()
            print("✅ Successfully queried minister entity")
            print(f"📊 Response: {json.dumps(data, indent=2)}")
            return True
        else:
            print(f"❌ Failed to query minister: {response.status_code}")
            print(f"Error: {response.text}")
            return False
            
    except Exception as e:
        print(f"❌ Error querying entity: {e}")
        return False

def query_minister_attributes_example():
    """
    Example: Query specific attributes of the minister.
    """
    print("\n📊 Querying Minister Attributes Example")
    print("=" * 50)
    
    # Get service URLs
    urls = get_service_urls()
    query_url = urls['query']
    
    entity_id = "minister-agriculture-001"
    attributes_to_query = [
        "personal_information",
        "performance_metrics",
        "budget_allocation"
    ]
    
    # First, get the entity metadata to retrieve display names
    print("🔍 Retrieving entity metadata for display names...")
    entity_metadata = None
    try:
        search_url = f"{query_url}/search"
        search_payload = {"id": entity_id}
        search_response = requests.post(search_url, json=search_payload)
        
        if search_response.status_code == 200:
            entity_data = search_response.json()
            if 'body' in entity_data and len(entity_data['body']) > 0:
                entity_metadata = entity_data['body'][0].get('metadata', [])
                print(f"✅ Retrieved {len(entity_metadata)} metadata entries")
                # Show the attribute name mappings
                for meta in entity_metadata:
                    if meta.get('key') in attributes_to_query:
                        print(f"  📋 {meta.get('key')} → {meta.get('value')}")
            else:
                print("⚠️ No entity data found")
        else:
            print(f"⚠️ Failed to retrieve entity metadata: {search_response.status_code}")
    except Exception as e:
        print(f"⚠️ Error retrieving entity metadata: {e}")
    
    for attr_name in attributes_to_query:
        print(f"\n🔍 Querying attribute: {attr_name}")
        attr_url = f"{query_url}/{entity_id}/attributes/{attr_name}"
        
        try:
            response = requests.get(attr_url)
            
            if response.status_code == 200:
                data = response.json()
                print(f"✅ Successfully retrieved {attr_name}")
                print(f"📅 Time Range: {data.get('start', 'N/A')} to {data.get('end', 'N/A')}")
                
                # Decode the protobuf value
                raw_value = data.get('value', {})
                decoded_value = decode_protobuf_any_value(raw_value)
                
                print(f"📊 Decoded Value:")
                if isinstance(decoded_value, dict) and 'columns' in decoded_value and 'rows' in decoded_value:
                    # Display as tabular data with metadata for display names
                    display_tabular_data(decoded_value, attr_name, entity_metadata)
                else:
                    print(f"    Raw decoded data: {json.dumps(decoded_value, indent=2)}")
            else:
                print(f"❌ Failed to query {attr_name}: {response.status_code}")
                
        except Exception as e:
            print(f"❌ Error querying attribute {attr_name}: {e}")

def test_protobuf_decoding():
    """
    Test function to debug protobuf decoding with a sample value.
    """
    print("\n🧪 Testing Protobuf Decoding")
    print("=" * 50)
    
    # Sample protobuf value from your output
    sample_value = '{"typeUrl":"type.googleapis.com/google.protobuf.Struct","value":"0AB5050A046461746112AC051AA9057B22636F6C756D6E73223A5B226964222C226669656C64222C2276616C7565222C226C6173745F75706461746564225D2C22726F7773223A5B5B312C2266756C6C5F6E616D65222C224D696E6973746572206F66204167726963756C7475726520616E6420466F6F64205365637572697479222C22323032342D30312D31355430303A30303A30305A225D2C5B322C2273686F72745F6E616D65222C224167726963756C74757265204D696E6973746572222C22323032342D30312D31355430303A30303A30305A225D2C5B332C22706F7274666F6C696F222C224167726963756C74757265222C22323032342D30312D31355430303A30303A30305A225D2C5B342C226170706F696E746D656E745F64617465222C22323032342D30312D31355430303A30303A30305A222C22323032342D30312D31355430303A30303A30305A225D2C5B352C226F66666963655F6C6F636174696F6E222C224D696E6973747279206F66204167726963756C747572652C20436F6C6F6D626F222C22323032342D30312D31355430303A30303A30305A225D2C5B362C22636F6E746163745F656D61696C222C226D696E69737465722D6167726963756C747572652D30303140676F762E6C6B222C22323032342D30312D31355430303A30303A30305A225D2C5B372C2273656375726974795F636C656172616E6365222C22436F6E666964656E7469616C222C22323032342D30312D31355430303A30303A30305A225D2C5B382C22656475636174696F6E5F6261636B67726F756E64222C2250684420696E204167726963756C747572616C20536369656E636573222C22323032342D30312D31355430303A30303A30305A225D2C5B392C2279656172735F6F665F657870657269656E6365222C223135222C22323032342D30312D31355430303A30303A30305A225D5D7D"}'
    
    print("🔍 Testing with sample protobuf value...")
    decoded = decode_protobuf_any_value(sample_value)
    print(f"📊 Decoded result: {json.dumps(decoded, indent=2)}")
    
    if isinstance(decoded, dict) and 'columns' in decoded and 'rows' in decoded:
        print("✅ Successfully decoded tabular data!")
        display_tabular_data(decoded, "test_attribute")
    else:
        print("❌ Failed to decode as tabular data")
        print(f"Type: {type(decoded)}")
        print(f"Content: {decoded}")

def main():
    """
    Main function to demonstrate entity creation and querying.
    """
    print("🏛️ Custom Meta Search Example")
    print("Creating and querying a Minister entity with attributes and metadata")
    print("=" * 70)
    
    # Test protobuf decoding first
    test_protobuf_decoding()
    
    # Step 1: Create the minister entity
    creation_success = create_minister_entity_example()
    
    if creation_success:
        # Step 2: Query the entity
        query_success = query_minister_example()
        
        if query_success:
            # Step 3: Query specific attributes
            query_minister_attributes_example()
            
            print("\n🎉 Example completed successfully!")
            print("=" * 50)
            print("📊 Summary:")
            print("  ✅ Created Minister entity with 3 attributes")
            print("  ✅ Added 7 metadata fields")
            print("  ✅ Demonstrated entity querying")
            print("  ✅ Demonstrated attribute querying")
        else:
            print("\n❌ Entity querying failed!")
    else:
        print("\n❌ Entity creation failed!")

if __name__ == "__main__":
    main()
