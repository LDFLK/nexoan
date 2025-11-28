# OpenGIN Data Model

## The "Entity"

The core unit of data in OpenGIN is the **Entity**. An Entity represents any object in the system (e.g., a Person, a Device, an Event).

### JSON Structure (Ingestion)

```json
{
    "id": "unique-entity-id",
    "kind": {
        "major": "Category",
        "minor": "SubCategory"
    },
    "name": {
        "value": "Entity Name",
        "startTime": "2024-01-01T00:00:00Z",
        "endTime": "" 
    },
    "metadata": [
        {"key": "source", "value": "import-job-1"},
        {"key": "priority", "value": "high"}
    ],
    "attributes": [
        {
            "key": "status",
            "values": [
                {
                    "value": "active",
                    "startTime": "2024-01-01T00:00:00Z",
                    "endTime": "2024-02-01T00:00:00Z"
                },
                {
                    "value": "inactive",
                    "startTime": "2024-02-01T00:00:00Z",
                    "endTime": ""
                }
            ]
        }
    ],
    "relationships": [
        {
            "key": "reports_to",
            "value": {
                "relatedEntityId": "manager-id",
                "startTime": "2024-01-01T00:00:00Z",
                "endTime": ""
            }
        }
    ]
}
```

## Storage Mapping

### 1. MongoDB (Metadata)
- **Content**: The `metadata` array.
- **Format**: Stored as a JSON document keyed by `_id` (Entity ID).
- **Use Case**: Flexible, unstructured data that doesn't need complex querying or time-series tracking.

### 2. Neo4j (Graph)
- **Content**: 
    - **Node**: The Entity itself (`id`, `kind`, `name`, `created`, `terminated`).
    - **Edges**: The `relationships` array.
- **Format**: 
    - Node Label: `Entity`
    - Edge Type: The relationship key (e.g., `REPORTS_TO`).
- **Use Case**: Graph traversals, finding connected entities, pathfinding.

### 3. PostgreSQL (Attributes)
- **Content**: The `attributes` array.
- **Format**: Relational tables.
    - `entity_attributes`: Maps Entity ID to Attribute ID.
    - `attr_<type>`: Stores the actual values with time ranges (`start_time`, `end_time`).
- **Use Case**: Time-series data, historical values, strict schema enforcement for values.

## TimeBasedValue

A key concept in OpenGIN is the `TimeBasedValue`. Almost all data (names, attributes, relationships) has a temporal dimension.

- **startTime**: When this value became true.
- **endTime**: When this value ceased to be true (empty string or null means "currently true").

This allows OpenGIN to answer questions like "Who was John's manager in 2023?" or "What was the status of this device last week?".
