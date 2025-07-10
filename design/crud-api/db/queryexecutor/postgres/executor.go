package postgres

import (
	"context"

	"lk/datafoundation/crud-api/db/queryexecutor/interfaces"
	"lk/datafoundation/crud-api/db/repository/postgres"
	"lk/datafoundation/crud-api/pkg/schema"
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

// TableExists checks if a table exists in the database
func (e *PostgresQueryExecutor) TableExists(ctx context.Context, plan *interfaces.QueryPlan) (bool, error) {
	tableName := plan.Parameters["tableName"].(string)
	return e.repository.TableExists(ctx, tableName)
}

// CreateDynamicTable creates a new dynamic table
func (e *PostgresQueryExecutor) CreateDynamicTable(ctx context.Context, plan *interfaces.QueryPlan) error {
	tableName := plan.Parameters["tableName"].(string)
	columns := plan.Parameters["columns"].([]postgres.Column)
	return e.repository.CreateDynamicTable(ctx, tableName, columns)
}

// GetTableList retrieves a list of tables for an entity
func (e *PostgresQueryExecutor) GetTableList(ctx context.Context, plan *interfaces.QueryPlan) ([]string, error) {
	entityID := plan.Parameters["entityId"].(string)
	return e.repository.GetTableList(ctx, entityID)
}

// GetSchemaOfTable retrieves the schema for a table
func (e *PostgresQueryExecutor) GetSchemaOfTable(ctx context.Context, plan *interfaces.QueryPlan) (*schema.SchemaInfo, error) {
	tableName := plan.Parameters["tableName"].(string)
	return e.repository.GetSchemaOfTable(ctx, tableName)
}

// InsertTabularData inserts data into a table
func (e *PostgresQueryExecutor) InsertTabularData(ctx context.Context, plan *interfaces.QueryPlan) error {
	tableName := plan.Parameters["tableName"].(string)
	entityAttributeID := plan.Parameters["entityAttributeId"].(int)
	columns := plan.Parameters["columns"].([]string)
	rows := plan.Parameters["rows"].([][]interface{})
	
	return e.repository.InsertTabularData(ctx, tableName, entityAttributeID, columns, rows)
}

// GetData retrieves data from a table with filters
func (e *PostgresQueryExecutor) GetData(ctx context.Context, plan *interfaces.QueryPlan) ([]map[string]interface{}, error) {
	tableName := plan.Parameters["tableName"].(string)
	filters := plan.Filters
	
	return e.repository.GetData(ctx, tableName, filters)
} 