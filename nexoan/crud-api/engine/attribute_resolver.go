package engine

import (
	"context"
	"fmt"
	pb "lk/datafoundation/crud-api/lk/datafoundation/crud-api"
	"lk/datafoundation/crud-api/pkg/storageinference"

	"time"

	"google.golang.org/protobuf/types/known/anypb"
)

// AttributeResolver interface defines the contract for all attribute resolvers
type AttributeResolver interface {
	Initialize() error
	CreateResolve(ctx context.Context, entityID, attrName string, value *pb.TimeBasedValue) error
	ReadResolve(ctx context.Context, entityID, attrName string, value *pb.TimeBasedValue) error
	UpdateResolve(ctx context.Context, entityID, attrName string, value *pb.TimeBasedValue) error
	DeleteResolve(ctx context.Context, entityID, attrName string, value *pb.TimeBasedValue) error
	Finalize() error
}

// BaseAttributeResolver provides common functionality for all resolvers
type BaseAttributeResolver struct {
	storageInferrer *storageinference.StorageInferrer
}

func (r *BaseAttributeResolver) Initialize() error {
	r.storageInferrer = &storageinference.StorageInferrer{}
	return nil
}

func (r *BaseAttributeResolver) Finalize() error {
	return nil
}

// EntityAttributeProcessor handles the main processing of Entity objects
type EntityAttributeProcessor struct {
	resolvers    map[storageinference.StorageType]AttributeResolver
	graphManager *GraphMetadataManager
}

// NewEntityAttributeProcessor creates a new processor with all resolvers initialized
func NewEntityAttributeProcessor() *EntityAttributeProcessor {
	processor := &EntityAttributeProcessor{
		resolvers:    make(map[storageinference.StorageType]AttributeResolver),
		graphManager: NewGraphMetadataManager(),
	}

	// Initialize all resolvers
	processor.resolvers[storageinference.GraphData] = &GraphAttributeResolver{}
	processor.resolvers[storageinference.TabularData] = &TabularAttributeResolver{}
	processor.resolvers[storageinference.MapData] = &DocumentAttributeResolver{}

	// Initialize each resolver
	for _, resolver := range processor.resolvers {
		if err := resolver.Initialize(); err != nil {
			fmt.Printf("Warning: failed to initialize resolver: %v\n", err)
		}
	}

	return processor
}

// GetResolver returns the resolver for a specific storage type
func (p *EntityAttributeProcessor) GetResolver(storageType storageinference.StorageType) (AttributeResolver, bool) {
	resolver, exists := p.resolvers[storageType]
	return resolver, exists
}

// ProcessEntityAttributes processes all attributes in an Entity
func (p *EntityAttributeProcessor) ProcessEntityAttributes(ctx context.Context, entity *pb.Entity, operation string) error {
	if entity == nil || entity.Attributes == nil {
		return nil
	}

	// Process each attribute
	for attrName, timeBasedValueList := range entity.Attributes {
		fmt.Printf("DEBUG: Processing attribute %s\n", attrName)
		if timeBasedValueList == nil {
			continue
		}

		// Process each time-based value
		for _, value := range timeBasedValueList.Values {
			if value == nil || value.Value == nil {
				continue
			}

			// Determine storage type
			storageType, err := p.determineStorageType(value.Value)
			fmt.Printf("DEBUG: Determined storage type for attribute %s: %s\n", attrName, storageType)
			if err != nil {
				return fmt.Errorf("error determining storage type for attribute %s: %v", attrName, err)
			}

			// Create or update graph metadata BEFORE processing the attribute
			if err := p.handleAttributeLookUp(ctx, entity.Id, attrName, storageType, operation); err != nil {
				return fmt.Errorf("error handling graph metadata for attribute %s: %v", attrName, err)
			}

			// Get appropriate resolver
			resolver, exists := p.resolvers[storageType]
			if !exists {
				fmt.Printf("Warning: no resolver found for storage type %s, skipping attribute %s\n", storageType, attrName)
				continue
			}

			// Execute the appropriate operation
			if err := p.executeOperation(ctx, resolver, operation, entity.Id, attrName, value); err != nil {
				return fmt.Errorf("error executing %s operation for attribute %s: %v", operation, attrName, err)
			}
		}
	}

	return nil
}

// handleAttributeLookUp handles the attribute look up operations
// This is the first step in the attribute processing pipeline.
// It creates the attribute look up metadata and the attribute node in the graph.
// It also creates the IS_ATTRIBUTE relationship between the entity and the attribute.
// It also creates the attribute metadata in the document database.
func (p *EntityAttributeProcessor) handleAttributeLookUp(ctx context.Context, entityID, attrName string, storageType storageinference.StorageType, operation string) error {
	// Generate attribute metadata
	fmt.Printf("DEBUG: Handling graph metadata for attribute %s\n", attrName)
	attributeID := GenerateAttributeID(entityID, attrName)
	storagePath := GenerateStoragePath(entityID, attrName, storageType)

	metadata := &AttributeMetadata{
		EntityID:      entityID,
		AttributeID:   attributeID,
		AttributeName: attrName,
		StorageType:   storageType,
		StoragePath:   storagePath,
		Updated:       time.Now(),
		Schema:        make(map[string]interface{}), // TODO: Extract schema from value
	}

	switch operation {
	case "create":
		// Create attribute node in graph
		if err := p.graphManager.CreateAttribute(ctx, metadata); err != nil {
			metadata.Created = time.Now()
			return fmt.Errorf("failed to create attribute node: %v", err)
		}

	case "update":
		// Update attribute metadata in graph
		if err := p.graphManager.UpdateAttribute(ctx, metadata); err != nil {
			metadata.Updated = time.Now()
			return fmt.Errorf("failed to update attribute metadata: %v", err)
		}

	case "delete":
		// Delete attribute node and relationships from graph
		if err := p.graphManager.DeleteAttribute(ctx, entityID, attrName); err != nil {
			metadata.Updated = time.Now()
			return fmt.Errorf("failed to delete attribute node: %v", err)
		}

	case "read":
		// For read operations, we might want to verify the attribute exists in the graph
		// This is optional but can be useful for validation
		_, err := p.graphManager.GetAttribute(ctx, entityID, attrName)
		if err != nil {
			fmt.Printf("Warning: attribute %s not found in graph metadata for entity %s\n", attrName, entityID)
		}
	}

	return nil
}

// determineStorageType determines the storage type of a TimeBasedValue
func (p *EntityAttributeProcessor) determineStorageType(anyValue *anypb.Any) (storageinference.StorageType, error) {
	if anyValue == nil {
		return storageinference.UnknownData, fmt.Errorf("anyValue is nil")
	}

	// Use the storage inference logic to determine type
	storageInferrer := &storageinference.StorageInferrer{}
	return storageInferrer.InferType(anyValue)
}

// executeOperation executes the appropriate CRUD operation
func (p *EntityAttributeProcessor) executeOperation(ctx context.Context, resolver AttributeResolver, operation, entityID, attrName string, value *pb.TimeBasedValue) error {
	if resolver == nil {
		return fmt.Errorf("resolver is nil")
	}

	switch operation {
		case "create":
			return resolver.CreateResolve(ctx, entityID, attrName, value)
		case "read":
			return resolver.ReadResolve(ctx, entityID, attrName, value)
		case "update":
			return resolver.UpdateResolve(ctx, entityID, attrName, value)
		case "delete":
			return resolver.DeleteResolve(ctx, entityID, attrName, value)
		default:
			return fmt.Errorf("unknown operation: %s", operation)
	}
}

// GraphAttributeResolver handles graph data structures with nodes and edges
type GraphAttributeResolver struct {
	BaseAttributeResolver
}

func (r *GraphAttributeResolver) CreateResolve(ctx context.Context, entityID, attrName string, value *pb.TimeBasedValue) error {
	// TODO: implement graph-specific create logic
	// - Validate graph structure (nodes and edges)
	// - Store in graph database (Neo4j)
	// - Handle graph relationships
	fmt.Printf("Creating graph attribute %s for entity %s\n", attrName, entityID)
	return nil
}

func (r *GraphAttributeResolver) ReadResolve(ctx context.Context, entityID, attrName string, value *pb.TimeBasedValue) error {
	// TODO: implement graph-specific read logic
	// - Query graph database
	// - Retrieve nodes and edges
	// - Return graph structure
	fmt.Printf("Reading graph attribute %s for entity %s\n", attrName, entityID)
	return nil
}

func (r *GraphAttributeResolver) UpdateResolve(ctx context.Context, entityID, attrName string, value *pb.TimeBasedValue) error {
	// TODO: implement graph-specific update logic
	// - Update nodes and edges
	// - Handle graph modifications
	// - Maintain graph consistency
	fmt.Printf("Updating graph attribute %s for entity %s\n", attrName, entityID)
	return nil
}

func (r *GraphAttributeResolver) DeleteResolve(ctx context.Context, entityID, attrName string, value *pb.TimeBasedValue) error {
	// TODO: implement graph-specific delete logic
	// - Remove nodes and edges
	// - Clean up relationships
	// - Handle cascading deletes
	fmt.Printf("Deleting graph attribute %s for entity %s\n", attrName, entityID)
	return nil
}

// TabularAttributeResolver handles tabular data structures with columns and rows
type TabularAttributeResolver struct {
	BaseAttributeResolver
}

func (r *TabularAttributeResolver) CreateResolve(ctx context.Context, entityID, attrName string, value *pb.TimeBasedValue) error {
	// TODO: implement tabular-specific create logic
	// - Validate tabular structure (columns and rows)
	// - Create or update database table
	// - Insert data rows
	fmt.Printf("Creating tabular attribute %s for entity %s\n", attrName, entityID)
	return nil
}

func (r *TabularAttributeResolver) ReadResolve(ctx context.Context, entityID, attrName string, value *pb.TimeBasedValue) error {
	// TODO: implement tabular-specific read logic
	// - Query database table
	// - Retrieve rows and columns
	// - Return tabular structure
	fmt.Printf("Reading tabular attribute %s for entity %s\n", attrName, entityID)
	return nil
}

func (r *TabularAttributeResolver) UpdateResolve(ctx context.Context, entityID, attrName string, value *pb.TimeBasedValue) error {
	// TODO: implement tabular-specific update logic
	// - Update table schema if needed
	// - Update data rows
	// - Handle schema evolution
	fmt.Printf("Updating tabular attribute %s for entity %s\n", attrName, entityID)
	return nil
}

func (r *TabularAttributeResolver) DeleteResolve(ctx context.Context, entityID, attrName string, value *pb.TimeBasedValue) error {
	// TODO: implement tabular-specific delete logic
	// - Delete data rows
	// - Optionally drop table
	// - Clean up schema
	fmt.Printf("Deleting tabular attribute %s for entity %s\n", attrName, entityID)
	return nil
}

// DocumentAttributeResolver handles document/map data structures with key-value pairs
type DocumentAttributeResolver struct {
	BaseAttributeResolver
}

func (r *DocumentAttributeResolver) CreateResolve(ctx context.Context, entityID, attrName string, value *pb.TimeBasedValue) error {
	// TODO: implement document-specific create logic
	// - Validate document structure
	// - Store in document database (MongoDB)
	// - Handle document indexing
	fmt.Printf("Creating document attribute %s for entity %s\n", attrName, entityID)
	return nil
}

func (r *DocumentAttributeResolver) ReadResolve(ctx context.Context, entityID, attrName string, value *pb.TimeBasedValue) error {
	// TODO: implement document-specific read logic
	// - Query document database
	// - Retrieve document structure
	// - Return key-value pairs
	fmt.Printf("Reading document attribute %s for entity %s\n", attrName, entityID)
	return nil
}

func (r *DocumentAttributeResolver) UpdateResolve(ctx context.Context, entityID, attrName string, value *pb.TimeBasedValue) error {
	// TODO: implement document-specific update logic
	// - Update document fields
	// - Handle partial updates
	// - Maintain document consistency
	fmt.Printf("Updating document attribute %s for entity %s\n", attrName, entityID)
	return nil
}

func (r *DocumentAttributeResolver) DeleteResolve(ctx context.Context, entityID, attrName string, value *pb.TimeBasedValue) error {
	// TODO: implement document-specific delete logic
	// - Remove document
	// - Clean up indexes
	// - Handle cascading deletes
	fmt.Printf("Deleting document attribute %s for entity %s\n", attrName, entityID)
	return nil
}
