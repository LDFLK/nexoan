package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	"lk/datafoundation/crud-api/db/config"
	pb "lk/datafoundation/crud-api/lk/datafoundation/crud-api"

	mongorepository "lk/datafoundation/crud-api/db/repository/mongo"
	neo4jrepository "lk/datafoundation/crud-api/db/repository/neo4j"
	postgres "lk/datafoundation/crud-api/db/repository/postgres"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/anypb"
)

// Server implements the CrudService
type Server struct {
	pb.UnimplementedCrudServiceServer
	mongoRepo *mongorepository.MongoRepository
	neo4jRepo *neo4jrepository.Neo4jRepository
}

// CreateEntity handles entity creation with metadata
func (s *Server) CreateEntity(ctx context.Context, req *pb.Entity) (*pb.Entity, error) {
	log.Printf("Creating Entity: %s", req.Id)

	// Always save the entity in MongoDB, even if it has no metadata
	// The HandleMetadata function will only process it if it has metadata
	// FIXME: https://github.com/LDFLK/nexoan/issues/120
	err := s.mongoRepo.HandleMetadata(ctx, req.Id, req)
	if err != nil {
		log.Printf("[server.CreateEntity] Error saving metadata in MongoDB: %v", err)
		return nil, err
	} else {
		log.Printf("[server.CreateEntity] Successfully saved metadata in MongoDB for entity: %s", req.Id)
	}

	// Validate required fields for Neo4j entity creation
	success, err := s.neo4jRepo.HandleGraphEntityCreation(ctx, req)
	if !success {
		log.Printf("[server.CreateEntity] Error saving entity in Neo4j: %v", err)
		return nil, err
	} else {
		log.Printf("[server.CreateEntity] Successfully saved entity in Neo4j for entity: %s", req.Id)
	}

	// Handle relationships
	err = s.neo4jRepo.HandleGraphRelationshipsCreate(ctx, req)
	if err != nil {
		log.Printf("[server.CreateEntity] Error saving relationships in Neo4j: %v", err)
		return nil, err
	} else {
		log.Printf("[server.CreateEntity] Successfully saved relationships in Neo4j for entity: %s", req.Id)
	}

	// Handle attributes
	_, err = postgres.HandleAttributes(req.Attributes)
	if err != nil {
		log.Printf("[server.CreateEntity] Error handling attributes: %v", err)
		return nil, err
	}

	// Return the complete entity including attributes
	return req, nil
}

// ReadEntity retrieves an entity's metadata
func (s *Server) ReadEntity(ctx context.Context, req *pb.ReadEntityRequest) (*pb.Entity, error) {
	log.Printf("Reading Entity: %s with output fields: %v", req.Entity.Id, req.Output)

	// Initialize a complete response entity with empty fields
	response := &pb.Entity{
		Id:            req.Entity.Id,
		Kind:          &pb.Kind{},
		Name:          &pb.TimeBasedValue{},
		Created:       "",
		Terminated:    "",
		Metadata:      make(map[string]*anypb.Any),
		Attributes:    make(map[string]*pb.TimeBasedValueList),
		Relationships: make(map[string]*pb.Relationship),
	}

	// Always fetch basic entity info from Neo4j
	kind, name, created, terminated, err := s.neo4jRepo.GetGraphEntity(ctx, req.Entity.Id)
	if err != nil {
		log.Printf("Error fetching entity info: %v", err)
		// Continue processing as we might still be able to get other information
	} else {
		response.Kind = kind
		response.Name = name
		response.Created = created
		response.Terminated = terminated
	}

	// If no output fields specified, return the entity with basic info
	if len(req.Output) == 0 {
		return response, nil
	}

	// Process each requested output field
	for _, field := range req.Output {
		log.Printf("[DEBUG] Entering switch statement for entity ID: %s", req.Entity.Id)
		switch field {
		case "metadata":
			log.Printf("[DEBUG] Processing metadata field for entity ID: %s", req.Entity.Id)
			// Get metadata from MongoDB
			metadata, err := s.mongoRepo.GetMetadata(ctx, req.Entity.Id)
			if err != nil {
				log.Printf("Error fetching metadata: %v", err)
				// Continue with other fields even if metadata fails
			} else {
				log.Printf("[DEBUG] Retrieved metadata: %+v", metadata)
				response.Metadata = metadata
			}

		case "relationships":
			// Handle relationships based on the input entity
			if req.Entity != nil && len(req.Entity.Relationships) > 0 {
				// Case 1: Validate that all relationships have a Name field
				for _, rel := range req.Entity.Relationships {
					if rel.Name == "" {
						return nil, fmt.Errorf("invalid relationship: all relationships must have a Name field")
					}
				}

				// Case 2: Call GetRelationshipsByName for each relationship
				for _, rel := range req.Entity.Relationships {
					log.Printf("Fetching related entity IDs for entity %s with relationship %s and start time %s", req.Entity.Id, rel.Name, rel.StartTime)
					relsByName, err := s.neo4jRepo.GetRelationshipsByName(ctx, req.Entity.Id, rel.Name, rel.StartTime)
					if err != nil {
						log.Printf("Error fetching related entity IDs for entity %s: %v", req.Entity.Id, err)
						continue // Continue with other relationships even if one fails
					}

					// Add the relationships to the response
					for id, relationship := range relsByName {
						response.Relationships[id] = relationship
					}
				}
			} else {
				// Case 3: If no specific relationships requested, get all relationships
				log.Printf("Fetching all relationships for entity %s", req.Entity.Id)
				graphRelationships, err := s.neo4jRepo.GetGraphRelationships(ctx, req.Entity.Id)
				if err != nil {
					log.Printf("Error fetching relationships for entity %s: %v", req.Entity.Id, err)
					// Continue with other fields even if relationships fail
				} else {
					response.Relationships = graphRelationships
				}
			}

		case "attributes":
			// TODO: Implement attribute fetching when available
			log.Printf("Attribute fetching not yet implemented")
			// Attributes map is already initialized

		case "kind", "name", "created", "terminated":
			// These fields are already fetched at the start
			continue

		default:
			log.Printf("Unknown output field requested: %s", field)
		}
	}

	return response, nil
}

// UpdateEntity modifies existing metadata
func (s *Server) UpdateEntity(ctx context.Context, req *pb.UpdateEntityRequest) (*pb.Entity, error) {
	// Extract ID from request parameter and entity data
	updateEntityID := req.Id
	updateEntity := req.Entity

	// Initialize metadata
	var metadata map[string]*anypb.Any

	// Pass the ID and metadata to HandleMetadata
	err := s.mongoRepo.HandleMetadata(ctx, updateEntityID, updateEntity)
	if err != nil {
		// Log error and continue with empty metadata
		log.Printf("[server.UpdateEntity] Error updating metadata for entity %s: %v", updateEntityID, err)
		metadata = make(map[string]*anypb.Any)
	} else {
		// Use the provided metadata
		metadata = updateEntity.Metadata
	}

	// Handle Graph Entity update if entity has required fields
	success, err := s.neo4jRepo.HandleGraphEntityUpdate(ctx, updateEntity)
	if !success {
		log.Printf("[server.UpdateEntity] Error updating graph entity for %s: %v", updateEntityID, err)
		// Continue processing despite error
	}

	// Handle Relationships update
	err = s.neo4jRepo.HandleGraphRelationshipsUpdate(ctx, updateEntity)
	if err != nil {
		log.Printf("[server.UpdateEntity] Error updating relationships for entity %s: %v", updateEntityID, err)
		// Continue processing despite error
	}

	// Read entity data from Neo4j to include in response
	kind, name, created, terminated, _ := s.neo4jRepo.GetGraphEntity(ctx, updateEntityID)

	// Get relationships from Neo4j
	relationships, _ := s.neo4jRepo.GetGraphRelationships(ctx, updateEntityID)

	// Return updated entity with all available information
	return &pb.Entity{
		Id:            updateEntity.Id,
		Kind:          kind,
		Name:          name,
		Created:       created,
		Terminated:    terminated,
		Metadata:      metadata,
		Attributes:    make(map[string]*pb.TimeBasedValueList), // Empty attributes
		Relationships: relationships,
	}, nil
}

// DeleteEntity removes metadata
func (s *Server) DeleteEntity(ctx context.Context, req *pb.EntityId) (*pb.Empty, error) {
	log.Printf("[server.DeleteEntity] Deleting Entity metadata: %s", req.Id)
	_, err := s.mongoRepo.DeleteEntity(ctx, req.Id)
	if err != nil {
		// Log error but return success
		log.Printf("[server.DeleteEntity] Error deleting metadata for entity %s: %v", req.Id, err)
	}
	// TODO: Implement Relationship Deletion in Neo4j
	// TODO: Implement Entity Deletion in Neo4j
	// TODO: Implement Attribute Deletion in Neo4j
	return &pb.Empty{}, nil
}

// ReadEntities retrieves a list of entities filtered by base attributes
func (s *Server) ReadEntities(ctx context.Context, req *pb.ReadEntityRequest) (*pb.EntityList, error) {
	if req.Entity == nil {
		return nil, fmt.Errorf("entity is required for filtering entities")
	}

	// Check if we have either an ID or Kind.Major
	if req.Entity.Id == "" && (req.Entity.Kind == nil || req.Entity.Kind.Major == "") {
		return nil, fmt.Errorf("either Entity.Id or Entity.Kind.Major is required for filtering entities")
	}

	// If we have an ID, add it to the filters
	if req.Entity.Id != "" {
		log.Printf("Filtering entities by ID: %s", req.Entity.Id)
	} else {
		log.Printf("Filtering entities by Kind.Major: %s", req.Entity.Kind.Major)
	}

	// Use HandleGraphEntityFilter to get filtered entities
	filteredEntities, err := s.neo4jRepo.HandleGraphEntityFilter(ctx, req)
	if err != nil {
		log.Printf("Error filtering entities: %v", err)
		return nil, err
	}

	// Convert filtered entities to pb.Entity format
	var entities []*pb.Entity
	for _, entity := range filteredEntities {
		pbEntity := &pb.Entity{
			Id: entity["id"].(string),
			Kind: &pb.Kind{
				Major: entity["kind"].(string),
				Minor: entity["minorKind"].(string),
			},
			Created: entity["created"].(string),
			Name: &pb.TimeBasedValue{ // How to represent time based value name?
				StartTime: entity["created"].(string),
				Value: &anypb.Any{
					TypeUrl: "type.googleapis.com/google.protobuf.StringValue",
					Value:   []byte(entity["name"].(string)),
				},
			},
		}

		// Add terminated if present
		if terminated, ok := entity["terminated"].(string); ok && terminated != "" {
			pbEntity.Terminated = terminated
			pbEntity.Name.EndTime = terminated
		}

		entities = append(entities, pbEntity)
	}

	return &pb.EntityList{
		Entities: entities,
	}, nil
}

// Start the gRPC server
func main() {
	// Initialize MongoDB config
	mongoConfig := &config.MongoConfig{
		URI:        os.Getenv("MONGO_URI"),
		DBName:     os.Getenv("MONGO_DB_NAME"),
		Collection: os.Getenv("MONGO_COLLECTION"),
	}

	// Initialize Neo4j config
	neo4jConfig := &config.Neo4jConfig{
		URI:      os.Getenv("NEO4J_URI"),
		Username: os.Getenv("NEO4J_USER"),
		Password: os.Getenv("NEO4J_PASSWORD"),
	}

	// Get host and port from environment variables with defaults
	host := os.Getenv("CRUD_SERVICE_HOST")
	if host == "" {
		host = "0.0.0.0"
	}
	port := os.Getenv("CRUD_SERVICE_PORT")
	if port == "" {
		port = "50051"
	}

	// Create MongoDB repository
	ctx := context.Background()
	mongoRepo := mongorepository.NewMongoRepository(ctx, mongoConfig)

	// Create Neo4j repository
	neo4jRepo, err := neo4jrepository.NewNeo4jRepository(ctx, neo4jConfig)
	if err != nil {
		log.Fatalf("[service.main] Failed to create Neo4j repository: %v", err)
	}
	defer neo4jRepo.Close(ctx)

	listener, err := net.Listen("tcp", host+":"+port)
	if err != nil {
		log.Fatalf("[service.main] Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	server := &Server{
		mongoRepo: mongoRepo,
		neo4jRepo: neo4jRepo,
	}

	pb.RegisterCrudServiceServer(grpcServer, server)

	// Register reflection service
	reflection.Register(grpcServer)

	log.Printf("[service.main] CRUD Service is running on %s:%s...", host, port)
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("[service.main] Failed to serve: %v", err)
	}
}
