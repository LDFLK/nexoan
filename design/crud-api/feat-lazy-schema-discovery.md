## feat: Implement Lazy Schema Discovery System

### Overview
This implements a lazy schema discovery system that efficiently explores data across our polyglot database (Neo4j and PostgreSQL) without loading full datasets.

### Detailed Function Specifications

#### 1. Query Executor Interfaces (`db/repository/query/executor/interfaces.go`)

```go
// Base Query Executor interface
type QueryExecutor interface {
    // ExecuteQuery executes a query and returns results
    ExecuteQuery(ctx context.Context, query *types.Query) (*types.QueryResult, error)
    
    // DiscoverSchema discovers and returns the schema for an entity
    DiscoverSchema(ctx context.Context, entityID string) (*types.Schema, error)
    
    // ValidateQuery validates if a query is valid against the schema
    ValidateQuery(ctx context.Context, query *types.Query) error
}

// Lazy Query Executor interface
type LazyQueryExecutor interface {
    QueryExecutor
    
    // GetEntityStructure retrieves basic entity structure from Neo4j
    GetEntityStructure(ctx context.Context, entityID string) (*types.EntityStructure, error)
    
    // GetAttributeSchema retrieves attribute schema from PostgreSQL
    GetAttributeSchema(ctx context.Context, entityID string) (*types.AttributeSchema, error)
    
    // LoadPartialData loads a limited subset of data
    LoadPartialData(ctx context.Context, query *types.Query, limit int) (*types.QueryResult, error)
    
    // GetCachedSchema retrieves schema from cache if available
    GetCachedSchema(ctx context.Context, entityID string) (*types.Schema, error)
    
    // InvalidateCache removes schema from cache
    InvalidateCache(ctx context.Context, entityID string) error
}
```

#### 2. Schema Discovery Implementation (`db/repository/query/executor/schema_discovery.go`)

```go
// SchemaDiscoverer handles schema discovery across databases
type SchemaDiscoverer struct {
    neo4jClient    *neo4j.Client
    postgresClient *postgres.Client
    schemaCache    *cache.SchemaCache
}

// Core schema discovery functions
func (sd *SchemaDiscoverer) DiscoverSchema(ctx context.Context, entityID string) (*types.Schema, error) {
    // 1. Check cache
    // 2. Get Neo4j structure
    // 3. Get PostgreSQL attributes
    // 4. Combine and cache
}

// Neo4j specific functions
func (sd *SchemaDiscoverer) discoverNeo4jSchema(ctx context.Context, entityID string) (*types.GraphSchema, error) {
    // 1. Get entity type
    // 2. Get relationships
    // 3. Get properties
}

// PostgreSQL specific functions
func (sd *SchemaDiscoverer) discoverPostgresSchema(ctx context.Context, entityID string) (*types.AttributeSchema, error) {
    // 1. Get attribute metadata
    // 2. Get data types
    // 3. Get constraints
}

// Schema combination function
func (sd *SchemaDiscoverer) combineSchemas(
    graphSchema *types.GraphSchema, 
    attrSchema *types.AttributeSchema,
) (*types.Schema, error) {
    // 1. Merge schemas
    // 2. Resolve conflicts
    // 3. Create unified view
}
```

#### 3. Lazy Executor Implementation (`db/repository/query/executor/lazy_executor.go`)

```go
// LazyExecutor implements lazy loading strategy
type LazyExecutor struct {
    discoverer *SchemaDiscoverer
    resolver   *resolver.QueryResolver
    generator  *generator.QueryGenerator
}

// Core execution functions
func (le *LazyExecutor) ExecuteQuery(ctx context.Context, query *types.Query) (*types.QueryResult, error) {
    // 1. Validate query
    // 2. Get schema
    // 3. Generate plan
    // 4. Execute
}

func (le *LazyExecutor) LoadPartialData(ctx context.Context, query *types.Query, limit int) (*types.QueryResult, error) {
    // 1. Validate limit
    // 2. Apply pagination
    // 3. Return subset
}

// Schema handling functions
func (le *LazyExecutor) GetEntityStructure(ctx context.Context, entityID string) (*types.EntityStructure, error) {
    // 1. Check Neo4j
    // 2. Get basic info
    // 3. Return structure
}

func (le *LazyExecutor) GetAttributeSchema(ctx context.Context, entityID string) (*types.AttributeSchema, error) {
    // 1. Check PostgreSQL
    // 2. Get attributes
    // 3. Return schema
}

// Cache management functions
func (le *LazyExecutor) GetCachedSchema(ctx context.Context, entityID string) (*types.Schema, error) {
    // 1. Check cache
    // 2. Validate TTL
    // 3. Return if valid
}

func (le *LazyExecutor) InvalidateCache(ctx context.Context, entityID string) error {
    // 1. Remove from cache
    // 2. Clean up resources
}
```
