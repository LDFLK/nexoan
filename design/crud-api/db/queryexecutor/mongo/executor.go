package mongo

import (
	"context"
	"lk/datafoundation/crud-api/db/queryexecutor/interfaces"
	mongorepo "lk/datafoundation/crud-api/db/repository/mongo"
	pb "lk/datafoundation/crud-api/lk/datafoundation/crud-api"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/protobuf/types/known/anypb"
)

// MongoQueryExecutor implements the MongoExecutor interface
type MongoQueryExecutor struct {
	repository *mongorepo.MongoRepository
}

// NewMongoQueryExecutor creates a new MongoDB query executor
func NewMongoQueryExecutor(repository *mongorepo.MongoRepository) interfaces.MongoExecutor {
	return &MongoQueryExecutor{
		repository: repository,
	}
}

// HandleMetadata handles metadata operations for an entity
func (e *MongoQueryExecutor) HandleMetadata(ctx context.Context, plan *interfaces.QueryPlan) error {
	entityId := plan.Parameters["entityId"].(string)
	entity := plan.Parameters["entity"].(*pb.Entity)
	return e.repository.HandleMetadata(ctx, entityId, entity)
}

// GetMetadata retrieves metadata for an entity
func (e *MongoQueryExecutor) GetMetadata(ctx context.Context, plan *interfaces.QueryPlan) (map[string]*anypb.Any, error) {
	entityId := plan.Parameters["entityId"].(string)
	return e.repository.GetMetadata(ctx, entityId)
}

// CreateEntity creates a new entity with metadata
func (e *MongoQueryExecutor) CreateEntity(ctx context.Context, entity *pb.Entity) (*mongo.InsertOneResult, error) {
	return e.repository.CreateEntity(ctx, entity)
}

// ReadEntity retrieves an entity and its metadata
func (e *MongoQueryExecutor) ReadEntity(ctx context.Context, entityId string) (*pb.Entity, error) {
	return e.repository.ReadEntity(ctx, entityId)
}

// UpdateEntity updates an entity's fields
func (e *MongoQueryExecutor) UpdateEntity(ctx context.Context, entityId string, updates map[string]interface{}) (*mongo.UpdateResult, error) {
	return e.repository.UpdateEntity(ctx, entityId, updates)
}

// DeleteEntity removes an entity
func (e *MongoQueryExecutor) DeleteEntity(ctx context.Context, entityId string) (*mongo.DeleteResult, error) {
	return e.repository.DeleteEntity(ctx, entityId)
} 