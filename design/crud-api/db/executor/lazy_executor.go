package executor

import (
	"context"
	"fmt"
	pb "lk/datafoundation/crud-api/lk/datafoundation/crud-api"
)

// LazyExecutor implements the lazy loading and query routing strategy.
// It orchestrates query execution by routing queries to the correct database executor.
type LazyExecutor struct {
	postgresExecutor QueryExecutor
	mongoExecutor    QueryExecutor
	neo4jExecutor    QueryExecutor
}

// NewLazyExecutor creates a new LazyExecutor
func NewLazyExecutor(postgresExecutor, mongoExecutor, neo4jExecutor QueryExecutor) QueryExecutor {
	return &LazyExecutor{
		postgresExecutor: postgresExecutor,
		mongoExecutor:    mongoExecutor,
		neo4jExecutor:    neo4jExecutor,
	}
}

// Execute determines the correct executor and delegates the query.
func (le *LazyExecutor) Execute(ctx context.Context, query *pb.Query) (*pb.QueryResult, error) {
	// TODO: Implement the schema discovery and query planning logic
	// to determine the correct executor.
	// For now, we will just route to a default executor as a placeholder.

	// Placeholder logic:
	fmt.Println("LazyExecutor is routing the query...")
	// Defaulting to Postgres for now
	return le.postgresExecutor.Execute(ctx, query)
}

// ValidateQuery determines the correct executor and delegates validation.
func (le *LazyExecutor) ValidateQuery(ctx context.Context, query *pb.Query) error {
	// TODO: Implement the schema discovery and query planning logic
	// to determine the correct executor for validation.

	// Placeholder logic:
	fmt.Println("LazyExecutor is routing the validation...")
	// Defaulting to Postgres for now
	return le.postgresExecutor.ValidateQuery(ctx, query)
} 