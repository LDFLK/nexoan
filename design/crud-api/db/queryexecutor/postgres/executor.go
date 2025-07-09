package postgres

import (
	"context"
	"github.com/LDFLK/nexoan/design/crud-api/db/queryexecutor/interfaces"
	"github.com/LDFLK/nexoan/design/crud-api/db/repository/postgres"
	pb "github.com/LDFLK/nexoan/design/crud-api/lk/datafoundation/crud-api"
)

// PostgresQueryExecutor implements the PostgresExecutor interface
type PostgresQueryExecutor struct {
	repository *postgres.PostgresRepository
}

// NewPostgresQueryExecutor creates a new PostgreSQL query executor
func NewPostgresQueryExecutor(repository *postgres.PostgresRepository) interfaces.PostgresExecutor {
	return &PostgresQueryExecutor{
		repository: repository,
	}
}

// HandleAttributeData handles attribute data operations
func (e *PostgresQueryExecutor) HandleAttributeData(ctx context.Context, plan *interfaces.QueryPlan) error {
	entityId := plan.Parameters["entityId"].(string)
	attributes := plan.Parameters["attributes"].(map[string]*pb.AttributeValue)
	return e.repository.HandleAttributeData(ctx, entityId, attributes)
}

// GetAttributeData retrieves attribute data
func (e *PostgresQueryExecutor) GetAttributeData(ctx context.Context, plan *interfaces.QueryPlan) (map[string]*pb.AttributeValue, error) {
	entityId := plan.Parameters["entityId"].(string)
	return e.repository.GetAttributeData(ctx, entityId)
} 