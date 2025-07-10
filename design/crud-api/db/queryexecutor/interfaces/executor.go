package interfaces

import (
	"context"
	"lk/datafoundation/crud-api/pkg/schema"
	pb "lk/datafoundation/crud-api/lk/datafoundation/crud-api"
	"google.golang.org/protobuf/types/known/anypb"
)

// QueryPlan represents the execution plan for a database operation
type QueryPlan struct {
	// Operation type (e.g., "get_metadata", "create_entity", "get_relationships")
	Operation string

	// Parameters for the operation
	Parameters map[string]interface{}

	// Filtering criteria
	Filters map[string]interface{}

	// Time-based constraints
	TimeRange struct {
		StartTime string
		EndTime   string
		ActiveAt  string
	}

	// Pagination
	Pagination struct {
		Limit  int32
		Offset int32
	}
}

// MongoExecutor defines MongoDB specific operations
type MongoExecutor interface {
	// Metadata Operations with query plan
	HandleMetadata(ctx context.Context, plan *QueryPlan) error
	GetMetadata(ctx context.Context, plan *QueryPlan) (map[string]*anypb.Any, error)
}

// Neo4jExecutor defines Neo4j specific operations
type Neo4jExecutor interface {
	// Entity Operations with query plan
	GetGraphEntity(ctx context.Context, plan *QueryPlan) (*pb.Kind, *pb.TimeBasedValue, string, string, error)
	HandleGraphEntityCreation(ctx context.Context, plan *QueryPlan) (bool, error)
	HandleGraphEntityUpdate(ctx context.Context, plan *QueryPlan) (bool, error)
	HandleGraphEntityFilter(ctx context.Context, plan *QueryPlan) ([]map[string]interface{}, error)

	// Relationship Operations with query plan
	GetGraphRelationships(ctx context.Context, plan *QueryPlan) (map[string]*pb.Relationship, error)
	GetFilteredRelationships(ctx context.Context, plan *QueryPlan) (map[string]*pb.Relationship, error)
	HandleGraphRelationshipsCreate(ctx context.Context, plan *QueryPlan) error
	HandleGraphRelationshipsUpdate(ctx context.Context, plan *QueryPlan) error
}

// PostgresExecutor defines PostgreSQL specific operations
type PostgresExecutor interface {
	// Table Operations
	TableExists(ctx context.Context, plan *QueryPlan) (bool, error)
	CreateDynamicTable(ctx context.Context, plan *QueryPlan) error
	GetTableList(ctx context.Context, plan *QueryPlan) ([]string, error)
	GetSchemaOfTable(ctx context.Context, plan *QueryPlan) (*schema.SchemaInfo, error)

	// Data Operations
	InsertTabularData(ctx context.Context, plan *QueryPlan) error
	GetData(ctx context.Context, plan *QueryPlan) ([]map[string]interface{}, error)
} 