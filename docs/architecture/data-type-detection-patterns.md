# Core Storage Type Detection Patterns

This document describes the JSON patterns used by the Nexoan system to automatically detect and classify the three core storage types when attributes are fed into the system.

## Overview

The system uses a hierarchical detection approach with the following precedence order for the three core storage types:
1. **Tabular Data** (highest priority)
2. **Graph Data** 
3. **Document/Map Data** (lowest priority)

## Detection Patterns

### 1. Tabular Data Pattern

**Detection Criteria**: Structure contains both `columns` and `rows` fields where:
- `columns` is an array of strings representing column names
- `rows` is an array of arrays representing data rows

**Example**:
```json
{
  "columns": ["id", "name", "department", "salary"],
  "rows": [
    [1, "John Doe", "Engineering", 75000],
    [2, "Jane Smith", "Marketing", 65000]
  ]
}
```

### 2. Graph Data Pattern

**Detection Criteria**: Structure contains both `nodes` and `edges` fields where:
- `nodes` is an array of node objects with properties
- `edges` is an array of edge objects with source and target references

**Example**:
```json
{
  "nodes": [
    {"id": "user1", "type": "user", "properties": {"name": "Alice", "age": 30}},
    {"id": "user2", "type": "user", "properties": {"name": "Bob", "age": 25}},
    {"id": "post1", "type": "post", "properties": {"title": "Hello", "content": "World"}}
  ],
  "edges": [
    {"source": "user1", "target": "user2", "type": "follows", "properties": {"since": "2024-01-01"}},
    {"source": "user1", "target": "post1", "type": "created", "properties": {"timestamp": "2024-03-20T10:00:00Z"}}
  ]
}
```

### 3. Document/Map Data Pattern

**Detection Criteria**: Object with key-value pairs that doesn't match tabular or graph patterns.

**Example**:
```json
{
  "user": {
    "name": "John",
    "age": 30,
    "address": {
      "city": "New York",
      "zip": "10001"
    }
  },
  "settings": {
    "theme": "dark",
    "notifications": true
  }
}
```

## Implementation Details

### Detection Algorithm

The system follows this detection sequence for the three core storage types:

1. **Check for Tabular Structure**:
   - Verify presence of `columns` field (must be array)
   - Verify presence of `rows` field (must be array)
   - Return `TabularData` if both conditions met

2. **Check for Graph Structure**:
   - Verify presence of `nodes` field
   - Verify presence of `edges` field
   - Return `GraphData` if both conditions met

3. **Default to Document/Map Structure**:
   - If object with fields but no specific pattern, return `MapData`

### Storage Type Constants

```go
const (
    TabularData StorageType = "tabular"
    MapData     StorageType = "map"
    GraphData   StorageType = "graph"
    UnknownData StorageType = "unknown"
)
```

This pattern-based detection system enables automatic data classification and appropriate storage backend selection in the Nexoan platform for the three core storage types: Tabular (PostgreSQL), Graph (Neo4j), and Document/Map (MongoDB).