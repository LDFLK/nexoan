package executor

import (
	"context"
	pb "lk/datafoundation/crud-api/lk/datafoundation/crud-api"
	"google.golang.org/protobuf/types/known/anypb"
)

// QueryExecutor is the base interface with common query methods
type QueryExecutor interface {
	// Common query methods that all executors must implement
	Execute(ctx context.Context, query *pb.Query) (*pb.QueryResult, error)
	ValidateQuery(ctx context.Context, query *pb.Query) error
}

// EntityReader defines methods for reading entity data
type EntityReader interface {
	ReadEntity(ctx context.Context, entityID string) (*pb.Entity, error)
}

// MetadataReader defines methods for reading metadata
type MetadataReader interface {
	GetMetadata(ctx context.Context, entityId string) (map[string]*anypb.Any, error)
}

// GraphReader defines methods for reading graph entities
type GraphReader interface {
	GetGraphEntity(ctx context.Context, entityId string) (*pb.Kind, *pb.TimeBasedValue, string, string, error)
	ReadGraphEntity(ctx context.Context, entityID string) (map[string]interface{}, error)
}

// RelationshipReader defines methods for reading relationships
type RelationshipReader interface {
	GetGraphRelationships(ctx context.Context, entityId string) (map[string]*pb.Relationship, error)
	GetRelationshipsByName(ctx context.Context, entityId, relationship, ts string) (map[string]*pb.Relationship, error)
	ReadRelatedGraphEntityIds(ctx context.Context, entityID, relationship, ts string) ([]map[string]interface{}, error)
	ReadRelationships(ctx context.Context, entityID string) ([]map[string]interface{}, error)
	ReadRelationship(ctx context.Context, relationshipID string) (map[string]interface{}, error)
	HandleGraphEntityFilter(ctx context.Context, req *pb.ReadEntityRequest) ([]map[string]interface{}, error)
}

// MongoExecutor combines base query execution with MongoDB-specific operations
type MongoExecutor interface {
	QueryExecutor  // Inherits base query methods
	EntityReader   // Can read entities
	MetadataReader // Can read metadata
}

// Neo4jExecutor combines base query execution with Neo4j-specific operations
type Neo4jExecutor interface {
	QueryExecutor       // Inherits base query methods
	GraphReader        // Can read graph entities
	RelationshipReader // Can read relationships
}

// PostgresExecutor only implements base query operations
type PostgresExecutor interface {
	QueryExecutor // Inherits base query methods
	// Note: PostgreSQL read operations are handled in repository layer
} 