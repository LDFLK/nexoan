package interfaces

import (
	"context"
	pb "github.com/LDFLK/nexoan/design/crud-api/lk/datafoundation/crud-api"
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
	// Time Series Operations
	HandleTimeSeriesData(ctx context.Context, plan *QueryPlan) error
	GetTimeSeriesData(ctx context.Context, plan *QueryPlan) (*pb.TimeSeriesData, error)
	AggregateTimeSeriesData(ctx context.Context, plan *QueryPlan) (*pb.TimeSeriesAggregation, error)

	// Attribute Operations
	HandleAttributeData(ctx context.Context, plan *QueryPlan) error
	GetAttributeData(ctx context.Context, plan *QueryPlan) (map[string]*pb.AttributeValue, error)
	BulkAttributeOperation(ctx context.Context, plan *QueryPlan) error

	// Transaction Operations
	BeginTransaction(ctx context.Context) error
	CommitTransaction(ctx context.Context) error
	RollbackTransaction(ctx context.Context) error

	// Query Operations
	ExecuteQuery(ctx context.Context, plan *QueryPlan) (interface{}, error)
	ExecuteBatchQuery(ctx context.Context, plan *QueryPlan) ([]interface{}, error)
} 