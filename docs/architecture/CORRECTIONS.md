# Documentation Corrections

This document tracks corrections made to the architecture documentation based on actual codebase review.

## Date: October 14, 2024

### Correction 1: Relationship Types are Dynamic, Not Predefined

**Issue**: Original documentation listed specific relationship types (REPORTS_TO, MANAGES, WORKS_ON, etc.) as if they were built-in types.

**Reality**: Nexoan uses a **completely dynamic relationship system**. There are NO predefined relationship types.

**How it works**:
- Users provide a `name` field in the Relationship protobuf message
- The system dynamically creates Neo4j relationships using that name as the type
- Example: `name: "reports_to"` creates `[:reports_to]` relationship in Neo4j
- Example: `name: "depends_on"` creates `[:depends_on]` relationship in Neo4j

**Code Evidence** (`nexoan/crud-api/db/repository/neo4j/neo4j_client.go:215`):
```go
createQuery := `MATCH (p {Id: $parentID}), (c {Id: $childID})
                MERGE (p)-[r:` + rel.Name + ` {Id: $relationshipID}]->(c)
                SET r.Created = datetime($startDate)`
```

The relationship type is **string concatenated** into the Cypher query at runtime.

### Correction 2: Relationship Property Names

**Issue**: Original documentation showed incorrect property names for Neo4j relationships.

**What the API uses** (Protobuf/JSON):
- `id` (lowercase)
- `name` (becomes relationship TYPE, not a property)
- `startTime`
- `endTime`
- `direction` (not stored in Neo4j)

**What Neo4j actually stores**:
```cypher
{
    Id: String,              // Uppercase "I"
    Created: DateTime,       // Not "startTime"
    Terminated: DateTime     // Not "endTime", nullable
}
```

**Key insight**: The `name` field is NOT stored as a property - it becomes the relationship TYPE in the Cypher pattern.

### Correction 3: Direction Field

**Issue**: Documentation suggested `direction` was stored as a relationship property.

**Reality**: `direction` is NOT stored in Neo4j. It's:
1. Implicitly determined by the graph structure (source → target)
2. Returned by the application layer when reading relationships
3. Computed based on whether the relationship is outgoing or incoming relative to the queried entity

**Code Evidence** (`nexoan/crud-api/db/repository/neo4j/neo4j_client.go:418-429`):
```cypher
-- Returns "OUTGOING" or "INCOMING" as computed value, not from storage
MATCH (e {Id: $entityID})-[r]->(related)
RETURN type(r) AS type, related.Id AS relatedID, "OUTGOING" AS direction
UNION
MATCH (e {Id: $entityID})<-[r]-(related)
RETURN type(r) AS type, related.Id AS relatedID, "INCOMING" AS direction
```

### Correction 4: Examples are Just Examples

**Issue**: Examples like "reports_to", "depends_on", "works_on" appeared to be special or built-in.

**Reality**: These are merely examples from:
- `reports_to`: Used in E2E organizational chart tests (`test_orgchart_test.py`)
- `depends_on`: Used in unit tests for package dependencies (`inference_test.go`, `schema_test.go`)

Users can create **any relationship type** with any name they want. There is no validation, no schema enforcement, and no predefined list.

## Impact on Documentation

### Files Updated:
1. `/docs/architecture/database-schemas.md`
   - Clarified dynamic relationship system
   - Fixed property names
   - Updated Cypher query examples
   - Added implementation code references

### Sections Updated:
- "Relationship Types" - Now explains dynamic nature
- "Relationship Structure" - Now shows correct Neo4j properties
- "Cypher Queries" - Now uses correct field names and explains dynamic types
- Examples - Now clearly marked as examples, not built-in types

## Verification

To verify these corrections, see:

1. **Relationship creation**: `nexoan/crud-api/db/repository/neo4j/neo4j_client.go:192-260`
2. **Relationship reading**: `nexoan/crud-api/db/repository/neo4j/neo4j_client.go:407-478`
3. **Test examples**: `nexoan/tests/e2e/test_orgchart_test.py:651-657`
4. **Unit test examples**: `nexoan/crud-api/pkg/storageinference/inference_test.go:422-425`

## Lessons Learned

1. **Always verify against actual code**: Don't assume based on examples in tests
2. **Check implementation, not just interfaces**: The Protobuf definitions don't show how data is actually stored
3. **Look for string concatenation**: Dynamic systems often inject values into queries
4. **Distinguish API from storage**: Field names can differ between API layer and storage layer

## Future Improvements

Consider adding to codebase:
1. Code comments explaining the dynamic relationship system
2. Examples in README showing how to create custom relationship types
3. Validation (optional) for relationship names (e.g., alphanumeric only, no spaces)
4. Documentation in protobuf file about name → type conversion

---

**Corrected By**: Architecture Documentation Review  
**Date**: October 14, 2024  
**Verified**: Yes, against actual codebase implementation

