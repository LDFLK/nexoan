package mongorepository

import (
	"context"

	pb "lk/datafoundation/crud-api/lk/datafoundation/crud-api"

	"google.golang.org/protobuf/types/known/anypb"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// Add this function to handle metadata operations
func (repo *MongoRepository) HandleMetadata(ctx context.Context, entityId string, entity *pb.Entity) error {
	// Skip operations if no metadata is provided
	if entity == nil || entity.GetMetadata() == nil || len(entity.GetMetadata()) == 0 {
		return nil
	}

	// Check if entity exists
	existingEntity, err := repo.ReadEntity(ctx, entityId)
	if err != nil && err != mongo.ErrNoDocuments {
		return err
	}

	if existingEntity == nil {
		// Create new entity with all fields including metadata
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
	} else {
		// Update existing entity's metadata
		// TODO: Should we choose _id for placing our id or should we use id field separately and use that.
		// Because then it is going to be reading or deleting or whatever by filtering using an attribute not the id of the object.
		_, err = repo.UpdateEntity(ctx, existingEntity.Id, bson.M{"metadata": entity.GetMetadata()})
	}

	return err
}

// Improved GetMetadata function that handles conversion internally
func (repo *MongoRepository) GetMetadata(ctx context.Context, entityId string) (map[string]*anypb.Any, error) {
	// Use the existing ReadEntity method for consistency
	entity, err := repo.ReadEntity(ctx, entityId)
	if err != nil {
		// Return the error to the caller (including ErrNoDocuments)
		return nil, err
	}

	// Handle nil metadata (this shouldn't happen given our HandleMetadata implementation,
	// but adding as a safeguard)
	if entity.Metadata == nil {
		return make(map[string]*anypb.Any), nil
	}

	// Return the original protobuf Any metadata
	return entity.Metadata, nil
}
