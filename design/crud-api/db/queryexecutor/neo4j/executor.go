package neo4j

import (
	"context"
	"github.com/LDFLK/nexoan/design/crud-api/db/queryexecutor/interfaces"
	"github.com/LDFLK/nexoan/design/crud-api/db/repository/neo4j"
	pb "github.com/LDFLK/nexoan/design/crud-api/lk/datafoundation/crud-api"
)

// Neo4jQueryExecutor implements the Neo4jExecutor interface
type Neo4jQueryExecutor struct {
	repository *neo4j.Neo4jRepository
}

// NewNeo4jQueryExecutor creates a new Neo4j query executor
func NewNeo4jQueryExecutor(repository *neo4j.Neo4jRepository) interfaces.Neo4jExecutor {
	return &Neo4jQueryExecutor{
		repository: repository,
	}
}

// Entity Operations
func (e *Neo4jQueryExecutor) GetGraphEntity(ctx context.Context, plan *interfaces.QueryPlan) (*pb.Kind, *pb.TimeBasedValue, string, string, error) {
	entityId := plan.Parameters["entityId"].(string)
	return e.repository.GetGraphEntity(ctx, entityId)
}

func (e *Neo4jQueryExecutor) HandleGraphEntityCreation(ctx context.Context, plan *interfaces.QueryPlan) (bool, error) {
	entity := plan.Parameters["entity"].(*pb.Entity)
	return e.repository.HandleGraphEntityCreation(ctx, entity)
}

func (e *Neo4jQueryExecutor) HandleGraphEntityUpdate(ctx context.Context, plan *interfaces.QueryPlan) (bool, error) {
	entity := plan.Parameters["entity"].(*pb.Entity)
	return e.repository.HandleGraphEntityUpdate(ctx, entity)
}

func (e *Neo4jQueryExecutor) HandleGraphEntityFilter(ctx context.Context, plan *interfaces.QueryPlan) ([]map[string]interface{}, error) {
	req := plan.Parameters["request"].(*pb.ReadEntityRequest)
	return e.repository.HandleGraphEntityFilter(ctx, req)
}

// Relationship Operations
func (e *Neo4jQueryExecutor) GetGraphRelationships(ctx context.Context, plan *interfaces.QueryPlan) (map[string]*pb.Relationship, error) {
	entityId := plan.Parameters["entityId"].(string)
	return e.repository.GetGraphRelationships(ctx, entityId)
}

func (e *Neo4jQueryExecutor) GetFilteredRelationships(ctx context.Context, plan *interfaces.QueryPlan) (map[string]*pb.Relationship, error) {
	entityId := plan.Parameters["entityId"].(string)
	relationshipId := plan.Parameters["relationshipId"].(string)
	relationship := plan.Parameters["relationship"].(string)
	relatedEntityId := plan.Parameters["relatedEntityId"].(string)
	startTime := plan.TimeRange.StartTime
	endTime := plan.TimeRange.EndTime
	direction := plan.Parameters["direction"].(string)
	activeAt := plan.TimeRange.ActiveAt

	return e.repository.GetFilteredRelationships(ctx, entityId, relationshipId, 
		relationship, relatedEntityId, startTime, endTime, direction, activeAt)
}

func (e *Neo4jQueryExecutor) HandleGraphRelationshipsCreate(ctx context.Context, plan *interfaces.QueryPlan) error {
	entity := plan.Parameters["entity"].(*pb.Entity)
	return e.repository.HandleGraphRelationshipsCreate(ctx, entity)
}

func (e *Neo4jQueryExecutor) HandleGraphRelationshipsUpdate(ctx context.Context, plan *interfaces.QueryPlan) error {
	entity := plan.Parameters["entity"].(*pb.Entity)
	return e.repository.HandleGraphRelationshipsUpdate(ctx, entity)
} 