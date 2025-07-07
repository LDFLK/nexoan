package executor

import (
	"context"
	pb "lk/datafoundation/crud-api/lk/datafoundation/crud-api"
)

// QueryExecutor is the main interface for executing queries.
// It is implemented by both the orchestrating LazyExecutor and the database-specific executors.
type QueryExecutor interface {
	// Execute executes a query and returns results
	Execute(ctx context.Context, query *pb.Query) (*pb.QueryResult, error)

	// ValidateQuery validates if a query is valid against the schema
	ValidateQuery(ctx context.Context, query *pb.Query) error
} 