## feat: Implement Lazy Schema Discovery System

### Overview
This implements a lazy schema discovery system that efficiently explores data across our polyglot database (Neo4j, PostgreSQL, and MongoDB) without loading full datasets.

### Detailed Function Specifications

#### 1. Query Executor Interfaces (`db/repository/query/executor/interfaces.go`)

```go
// QueryExecutor is the main interface for executing queries.
// It is implemented by both the orchestrating LazyExecutor and the database-specific executors.
type QueryExecutor interface {
    // Execute executes a query and returns results
    Execute(ctx context.Context, query *types.Query) (*types.QueryResult, error)
    
    // ValidateQuery validates if a query is valid against the schema
    ValidateQuery(ctx context.Context, query *types.Query) error
}

// Lazy Query Executor interface
type LazyQueryExecutor interface {
    QueryExecutor
    
    // GetEntityStructure retrieves basic entity structure from Neo4j
    GetEntityStructure(ctx context.Context, entityID string) (*types.EntityStructure, error)
    
    // GetAttributeSchema retrieves attribute schema from PostgreSQL and MongoDB
    GetAttributeSchema(ctx context.Context, entityID string) (*types.AttributeSchema, error)
    
    // LoadPartialData loads a limited subset of data
    LoadPartialData(ctx context.Context, query *types.Query, limit int) (*types.QueryResult, error)
}
```

#### 2. Schema Discovery Implementation (`db/repository/query/executor/schema_discovery.go`)

```go
// SchemaDiscoverer handles schema discovery across databases
type SchemaDiscoverer struct {
    neo4jClient    *neo4j.Client
    postgresClient *postgres.Client
    mongoClient    *mongo.Client
}

// Core schema discovery functions
func (sd *SchemaDiscoverer) DiscoverSchema(ctx context.Context, entityID string) (*types.Schema, error) {
    // 1. Get Neo4j structure
    // 2. Get PostgreSQL attributes
    // 3. Get MongoDB collections and fields
    // 4. Combine schemas
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

// MongoDB specific functions
func (sd *SchemaDiscoverer) discoverMongoSchema(ctx context.Context, entityID string) (*types.DocumentSchema, error) {
    // 1. Get collection structure
    // 2. Sample documents for field discovery
    // 3. Infer data types
    // 4. Identify indexes and unique constraints
}

// Schema combination function
func (sd *SchemaDiscoverer) combineSchemas(
    graphSchema *types.GraphSchema, 
    attrSchema *types.AttributeSchema,
    docSchema *types.DocumentSchema,
) (*types.Schema, error) {
    // 1. Merge schemas from all sources
    // 2. Resolve conflicts between different storage types
    // 3. Create unified view
    // 4. Handle data type mappings between systems
}
```

#### 3. Lazy Executor (Orchestrator) Implementation (`db/repository/query/executor/lazy_executor.go`)

```go
// LazyExecutor implements the lazy loading and query routing strategy.
// It orchestrates query execution by routing queries to the correct database executor.
type LazyExecutor struct {
    discoverer       *SchemaDiscoverer
    queryPlan        *plan.QueryPlan
    generator        *generator.QueryGenerator
    postgresExecutor QueryExecutor
    mongoExecutor    QueryExecutor
    neo4jExecutor    QueryExecutor
}

// Core execution functions
func (le *LazyExecutor) Execute(ctx context.Context, query *types.Query) (*types.QueryResult, error) {
    // 1. Validate query
    // 2. Discover schema if not already known
    // 3. Generate a query plan to determine the target database
    // 4. Based on the plan, delegate the query to the single appropriate database-specific executor
    // 5. Return the result directly from the selected executor
}

func (le *LazyExecutor) LoadPartialData(ctx context.Context, query *types.Query, limit int) (*types.QueryResult, error) {
    // 1. Validate limit
    // 2. Apply pagination
    // 3. Return subset from appropriate data source
}

// Schema handling functions
func (le *LazyExecutor) GetEntityStructure(ctx context.Context, entityID string) (*types.EntityStructure, error) {
    // 1. Check Neo4j
    // 2. Get basic info
    // 3. Return structure
}

func (le *LazyExecutor) GetAttributeSchema(ctx context.Context, entityID string) (*types.AttributeSchema, error) {
    // 1. Check PostgreSQL and MongoDB
    // 2. Get attributes and document fields
    // 3. Return combined schema
}
```

#### 4. Database-Specific Executors

These executors are responsible for handling queries for a single database type. They all implement the `QueryExecutor` interface.

##### 4.1. PostgreSQL Executor (`db/repository/query/executor/postgres_executor.go`)

```go
// PostgresExecutor handles query execution against PostgreSQL
type PostgresExecutor struct {
    db *postgres.Client // Assuming a postgres client
}

// NewPostgresExecutor creates a new PostgresExecutor
func NewPostgresExecutor(client *postgres.Client) QueryExecutor {
    return &PostgresExecutor{db: client}
}

// Execute runs a query against PostgreSQL
func (e *PostgresExecutor) Execute(ctx context.Context, query *types.Query) (*types.QueryResult, error) {
    // 1. Translate the relevant parts of the query into SQL
    // 2. Execute the SQL query
    // 3. Convert the SQL result into a *types.QueryResult
    // 4. Return the result
}

// ValidateQuery validates the PostgreSQL-specific parts of a query
func (e *PostgresExecutor) ValidateQuery(ctx context.Context, query *types.Query) error {
    // Implementation for validation
}
```

##### 4.2. MongoDB Executor (`db/repository/query/executor/mongo_executor.go`)

```go
// MongoExecutor handles query execution against MongoDB
type MongoExecutor struct {
    db *mongo.Client // Assuming a mongo client
}

// NewMongoExecutor creates a new MongoExecutor
func NewMongoExecutor(client *mongo.Client) QueryExecutor {
    return &MongoExecutor{db: client}
}

// Execute runs a query against MongoDB
func (e *MongoExecutor) Execute(ctx context.Context, query *types.Query) (*types.QueryResult, error) {
    // 1. Translate the relevant parts of the query into a MongoDB query (e.g., an aggregation pipeline)
    // 2. Execute the MongoDB query
    // 3. Convert the result into a *types.QueryResult
    // 4. Return the result
}

// ValidateQuery validates the MongoDB-specific parts of a query
func (e *MongoExecutor) ValidateQuery(ctx context.Context, query *types.Query) error {
    // Implementation for validation
}
```

##### 4.3. Neo4j Executor (`db/repository/query/executor/neo4j_executor.go`)

```go
// Neo4jExecutor handles query execution against Neo4j
type Neo4jExecutor struct {
    db *neo4j.Client // Assuming a neo4j client
}

// NewNeo4jExecutor creates a new Neo4jExecutor
func NewNeo4jExecutor(client *neo4j.Client) QueryExecutor {
    return &Neo4jExecutor{db: client}
}

// Execute runs a query against Neo4j
func (e *Neo4jExecutor) Execute(ctx context.Context, query *types.Query) (*types.QueryResult, error) {
    // 1. Translate the relevant parts of the query into Cypher
    // 2. Execute the Cypher query
    // 3. Convert the result into a *types.QueryResult
    // 4. Return the result
}

// ValidateQuery validates the Neo4j-specific parts of a query
func (e *Neo4jExecutor) ValidateQuery(ctx context.Context, query *types.Query) error {
    // Implementation for validation
}
```
