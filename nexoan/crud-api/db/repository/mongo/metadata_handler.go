package mongorepository

import (
	"context"
	"log"

	pb "lk/datafoundation/crud-api/lk/datafoundation/crud-api"

	"google.golang.org/protobuf/types/known/anypb"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// ...existing code...

// CreateMetadata creates a new entity with metadata if it doesn't exist.
func (repo *MongoRepository) HandleMetadataCreation(ctx context.Context, entityId string, entity *pb.Entity) error {
	if entity == nil || entity.GetMetadata() == nil || len(entity.GetMetadata()) == 0 {
		return nil
	}

	existingEntity, err := repo.ReadEntity(ctx, entityId)
	if err != nil && err != mongo.ErrNoDocuments {
		return err
	}
	if existingEntity != nil {
		return mongo.ErrNoDocuments // Or return a custom error: entity already exists
	}

	newEntity := &pb.Entity{
		Id:            entityId,
		Metadata:      entity.GetMetadata(),
		Kind:          entity.Kind,
		Created:       entity.Created,
		Terminated:    entity.Terminated,
		Name:          entity.Name,
		Attributes:    entity.Attributes,
		Relationships: entity.Relationships,
	}
	_, err = repo.CreateEntity(ctx, newEntity)
	return err
}

// UpdateMetadata updates metadata for an existing entity.
func (repo *MongoRepository) HandleMetadataUpdate(ctx context.Context, entityId string, metadata map[string]*anypb.Any) error {
	if len(metadata) == 0 {
		return nil
	}

	existingEntity, err := repo.ReadEntity(ctx, entityId)
	if err != nil {
		return err
	}
	if existingEntity == nil {
		return mongo.ErrNoDocuments // Or return a custom error: entity not found
	}

	_, err = repo.UpdateEntity(ctx, entityId, bson.M{"metadata": metadata})
	return err
}

// Improved GetMetadata function that handles conversion internally
func (repo *MongoRepository) GetMetadata(ctx context.Context, entityId string) (map[string]*anypb.Any, error) {
	// Use the existing ReadEntity method for consistency
	entity, err := repo.ReadEntity(ctx, entityId)
	if err != nil {
		// Log error and return empty metadata map
		log.Printf("Error retrieving metadata for entity %s: %v", entityId, err)
		metadata := make(map[string]*anypb.Any)
		return metadata, nil
	}

	// Handle nil metadata (this shouldn't happen given our HandleMetadata implementation,
	// but adding as a safeguard)
	if entity.Metadata == nil {
		return make(map[string]*anypb.Any), nil
	}

	// Return the original protobuf Any metadata
	return entity.Metadata, nil
}
