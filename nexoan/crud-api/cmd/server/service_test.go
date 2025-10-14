package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"lk/datafoundation/crud-api/db/config"
	mongorepository "lk/datafoundation/crud-api/db/repository/mongo"
	neo4jrepository "lk/datafoundation/crud-api/db/repository/neo4j"
	postgres "lk/datafoundation/crud-api/db/repository/postgres"
	pb "lk/datafoundation/crud-api/lk/datafoundation/crud-api"

	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

var server *Server

// createNameValue is a helper function to properly create a TimeBasedValue for Name field
func createNameValue(startTime, name string) *pb.TimeBasedValue {
	value, _ := anypb.New(&wrapperspb.StringValue{Value: name})
	return &pb.TimeBasedValue{
		StartTime: startTime,
		Value:     value,
	}
}

// TestMain sets up the actual MongoDB, Neo4j, and PostgreSQL repositories before running the tests
func TestMain(m *testing.M) {
	// Load environment variables for database configurations
	neo4jConfig := &config.Neo4jConfig{
		URI:      os.Getenv("NEO4J_URI"),
		Username: os.Getenv("NEO4J_USER"),
		Password: os.Getenv("NEO4J_PASSWORD"),
	}

	mongoConfig := &config.MongoConfig{
		URI:        os.Getenv("MONGO_URI"),
		DBName:     os.Getenv("MONGO_DB_NAME"),
		Collection: os.Getenv("MONGO_COLLECTION"),
	}

	postgresConfig := &postgres.Config{
		Host:     os.Getenv("POSTGRES_HOST"),
		Port:     os.Getenv("POSTGRES_PORT"),
		User:     os.Getenv("POSTGRES_USER"),
		Password: os.Getenv("POSTGRES_PASSWORD"),
		DBName:   os.Getenv("POSTGRES_DB"),
		SSLMode:  os.Getenv("POSTGRES_SSL_MODE"),
	}

	// Initialize Neo4j repository
	ctx := context.Background()
	neo4jRepo, err := neo4jrepository.NewNeo4jRepository(ctx, neo4jConfig)
	if err != nil {
		log.Fatalf("Failed to initialize Neo4j repository: %v", err)
	}
	defer neo4jRepo.Close(ctx)

	// Initialize MongoDB repository
	mongoRepo := mongorepository.NewMongoRepository(ctx, mongoConfig)
	if mongoRepo == nil {
		log.Fatalf("Failed to initialize MongoDB repository")
	}

	// Initialize PostgreSQL repository
	postgresRepo, err := postgres.NewPostgresRepository(*postgresConfig)
	if err != nil {
		log.Fatalf("Failed to initialize PostgreSQL repository: %v", err)
	}
	defer postgresRepo.Close()

	// Create the server with the initialized repositories
	server = &Server{
		mongoRepo:    mongoRepo,
		neo4jRepo:    neo4jRepo,
		postgresRepo: postgresRepo,
	}

	// Run the tests
	code := m.Run()

	// Exit with the test result code
	os.Exit(code)
}

// TestServiceCreateEntity tests creating an entity through the service layer
func TestServiceCreateEntity(t *testing.T) {
	ctx := context.Background()

	// Create a simple entity
	entity := &pb.Entity{
		Id:      "service_test_entity_1",
		Kind:    &pb.Kind{Major: "Person", Minor: "Minister"},
		Name:    createNameValue("2025-03-18T00:00:00Z", "John Doe"),
		Created: "2025-03-18T00:00:00Z",
	}

	// Create the entity
	resp, err := server.CreateEntity(ctx, entity)
	if err != nil {
		t.Fatalf("CreateEntity() error = %v", err)
	}
	if resp == nil {
		t.Fatal("CreateEntity() returned nil response")
	}
	if resp.Id != entity.Id {
		t.Errorf("CreateEntity() response ID = %v, want %v", resp.Id, entity.Id)
	}

	log.Printf("Successfully created entity: %v", resp.Id)
}

// TestServiceReadEntity tests reading an entity through the service layer
func TestServiceReadEntity(t *testing.T) {
	ctx := context.Background()

	// First create an entity to read
	entity := &pb.Entity{
		Id: "service_test_entity_2",
		Kind: &pb.Kind{
			Major: "Person",
			Minor: "Minister",
		},
		Name:    createNameValue("2025-03-18T00:00:00Z", "Jane Doe"),
		Created: "2025-03-18T00:00:00Z",
	}

	_, err := server.CreateEntity(ctx, entity)
	if err != nil {
		t.Fatalf("CreateEntity() error = %v", err)
	}

	// Read the entity back
	readReq := &pb.ReadEntityRequest{
		Entity: &pb.Entity{Id: "service_test_entity_2"},
		Output: []string{}, // Request basic info only
	}

	readResp, err := server.ReadEntity(ctx, readReq)
	if err != nil {
		t.Fatalf("ReadEntity() error = %v", err)
	}
	if readResp == nil {
		t.Fatal("ReadEntity() returned nil response")
	}
	if readResp.Id != entity.Id {
		t.Errorf("ReadEntity() response ID = %v, want %v", readResp.Id, entity.Id)
	}
	if readResp.Kind.Major != "Person" {
		t.Errorf("ReadEntity() response Kind.Major = %v, want Person", readResp.Kind.Major)
	}

	log.Printf("Successfully read entity: %v", readResp.Id)
}

// TestServiceCreateEntityWithRelationships tests creating entities with relationships
func TestServiceCreateEntityWithRelationships(t *testing.T) {
	ctx := context.Background()

	// Create first entity
	entity1 := &pb.Entity{
		Id: "service_test_entity_3",
		Kind: &pb.Kind{
			Major: "Person",
			Minor: "Employee",
		},
		Name:    createNameValue("2025-03-18T00:00:00Z", "Alice"),
		Created: "2025-03-18T00:00:00Z",
	}

	_, err := server.CreateEntity(ctx, entity1)
	if err != nil {
		t.Fatalf("CreateEntity(entity1) error = %v", err)
	}

	// Create second entity
	entity2 := &pb.Entity{
		Id: "service_test_entity_4",
		Kind: &pb.Kind{
			Major: "Person",
			Minor: "Employee",
		},
		Name:    createNameValue("2025-03-18T00:00:00Z", "Bob"),
		Created: "2025-03-18T00:00:00Z",
	}

	_, err = server.CreateEntity(ctx, entity2)
	if err != nil {
		t.Fatalf("CreateEntity(entity2) error = %v", err)
	}

	// Create third entity with relationship to entity2
	entity3 := &pb.Entity{
		Id: "service_test_entity_5",
		Kind: &pb.Kind{
			Major: "Person",
			Minor: "Employee",
		},
		Name:    createNameValue("2025-03-18T00:00:00Z", "Charlie"),
		Created: "2025-03-18T00:00:00Z",
		Relationships: map[string]*pb.Relationship{
			"service_test_rel_1": {
				Id:              "service_test_rel_1",
				Name:            "KNOWS",
				RelatedEntityId: "service_test_entity_4",
				StartTime:       "2025-03-18T00:00:00Z",
			},
		},
	}

	resp, err := server.CreateEntity(ctx, entity3)
	if err != nil {
		t.Fatalf("CreateEntity(entity3 with relationships) error = %v", err)
	}
	if resp == nil {
		t.Fatal("CreateEntity() returned nil response")
	}

	log.Printf("Successfully created entity with relationships: %v", resp.Id)
}

// TestServiceReadEntityWithRelationships tests reading entities with relationships
func TestServiceReadEntityWithRelationships(t *testing.T) {
	ctx := context.Background()

	// Create two entities
	entity1 := &pb.Entity{
		Id: "service_test_entity_6",
		Kind: &pb.Kind{
			Major: "Person",
			Minor: "Employee",
		},
		Name:    createNameValue("2025-03-18T00:00:00Z", "David"),
		Created: "2025-03-18T00:00:00Z",
	}

	entity2 := &pb.Entity{
		Id: "service_test_entity_7",
		Kind: &pb.Kind{
			Major: "Person",
			Minor: "Employee",
		},
		Name:    createNameValue("2025-03-18T00:00:00Z", "Eve"),
		Created: "2025-03-18T00:00:00Z",
		Relationships: map[string]*pb.Relationship{
			"service_test_rel_2": {
				Id:              "service_test_rel_2",
				Name:            "REPORTS_TO",
				RelatedEntityId: "service_test_entity_6",
				StartTime:       "2025-03-18T00:00:00Z",
			},
		},
	}

	_, err := server.CreateEntity(ctx, entity1)
	if err != nil {
		t.Fatalf("CreateEntity(entity1) error = %v", err)
	}

	_, err = server.CreateEntity(ctx, entity2)
	if err != nil {
		t.Fatalf("CreateEntity(entity2 with relationship) error = %v", err)
	}

	// Read entity with relationships
	readReq := &pb.ReadEntityRequest{
		Entity: &pb.Entity{Id: "service_test_entity_7"},
		Output: []string{"relationships"},
	}

	readResp, err := server.ReadEntity(ctx, readReq)
	if err != nil {
		t.Fatalf("ReadEntity() error = %v", err)
	}
	if readResp == nil {
		t.Fatal("ReadEntity() returned nil response")
	}
	if len(readResp.Relationships) == 0 {
		t.Error("ReadEntity() returned no relationships")
	}

	// Verify the relationship exists
	if _, exists := readResp.Relationships["service_test_rel_2"]; !exists {
		t.Error("Expected relationship 'service_test_rel_2' not found")
	}

	log.Printf("Successfully read entity with relationships: %v", readResp.Id)
}

// TestServiceUpdateEntity tests updating an entity through the service layer
func TestServiceUpdateEntity(t *testing.T) {
	ctx := context.Background()

	// Create an entity first
	entity := &pb.Entity{
		Id: "service_test_entity_8",
		Kind: &pb.Kind{
			Major: "Person",
			Minor: "Employee",
		},
		Name:    createNameValue("2025-03-18T00:00:00Z", "Mary"),
		Created: "2025-03-18T00:00:00Z",
	}

	_, err := server.CreateEntity(ctx, entity)
	if err != nil {
		t.Fatalf("CreateEntity() error = %v", err)
	}

	// Update the entity
	updateReq := &pb.UpdateEntityRequest{
		Id: "service_test_entity_8",
		Entity: &pb.Entity{
			Id:         "service_test_entity_8",
			Name:       createNameValue("2025-03-18T00:00:00Z", "Mary Updated"),
			Terminated: "2025-12-31T00:00:00Z",
		},
	}

	updateResp, err := server.UpdateEntity(ctx, updateReq)
	if err != nil {
		t.Fatalf("UpdateEntity() error = %v", err)
	}
	if updateResp == nil {
		t.Fatal("UpdateEntity() returned nil response")
	}
	if updateResp.Terminated != "2025-12-31T00:00:00Z" {
		t.Errorf("UpdateEntity() Terminated = %v, want 2025-12-31T00:00:00Z", updateResp.Terminated)
	}

	log.Printf("Successfully updated entity: %v", updateResp.Id)
}

// TestServiceReadEntities tests filtering entities through the service layer
func TestServiceReadEntities(t *testing.T) {
	ctx := context.Background()

	// Create multiple entities of the same kind
	for i := 1; i <= 3; i++ {
		entity := &pb.Entity{
			Id: fmt.Sprintf("service_test_ministry_%d", i),
			Kind: &pb.Kind{
				Major: "Organization",
				Minor: "Ministry",
			},
			Name:    createNameValue("2025-03-18T00:00:00Z", fmt.Sprintf("Ministry %d", i)),
			Created: "2025-03-18T00:00:00Z",
		}

		_, err := server.CreateEntity(ctx, entity)
		if err != nil {
			t.Fatalf("CreateEntity(ministry %d) error = %v", i, err)
		}
	}

	// Filter entities by Kind
	readReq := &pb.ReadEntityRequest{
		Entity: &pb.Entity{
			Kind: &pb.Kind{
				Major: "Organization",
				Minor: "Ministry",
			},
		},
	}

	listResp, err := server.ReadEntities(ctx, readReq)
	if err != nil {
		t.Fatalf("ReadEntities() error = %v", err)
	}
	if listResp == nil {
		t.Fatal("ReadEntities() returned nil response")
	}
	if len(listResp.Entities) < 3 {
		t.Errorf("ReadEntities() returned %d entities, want at least 3", len(listResp.Entities))
	}

	log.Printf("Successfully filtered entities: found %d entities", len(listResp.Entities))
}

// TestServiceReadEntityById tests filtering a single entity by ID
func TestServiceReadEntityById(t *testing.T) {
	ctx := context.Background()

	// Create an entity
	entity := &pb.Entity{
		Id: "service_test_entity_9",
		Kind: &pb.Kind{
			Major: "Person",
			Minor: "Employee",
		},
		Name:    createNameValue("2025-03-18T00:00:00Z", "Frank"),
		Created: "2025-03-18T00:00:00Z",
	}

	_, err := server.CreateEntity(ctx, entity)
	if err != nil {
		t.Fatalf("CreateEntity() error = %v", err)
	}

	// Filter by ID
	readReq := &pb.ReadEntityRequest{
		Entity: &pb.Entity{
			Id: "service_test_entity_9",
		},
	}

	listResp, err := server.ReadEntities(ctx, readReq)
	if err != nil {
		t.Fatalf("ReadEntities() error = %v", err)
	}
	if listResp == nil {
		t.Fatal("ReadEntities() returned nil response")
	}
	if len(listResp.Entities) != 1 {
		t.Errorf("ReadEntities() returned %d entities, want 1", len(listResp.Entities))
	}
	if len(listResp.Entities) > 0 && listResp.Entities[0].Id != "service_test_entity_9" {
		t.Errorf("ReadEntities() returned entity with ID %v, want service_test_entity_9", listResp.Entities[0].Id)
	}

	log.Printf("Successfully filtered entity by ID: %v", listResp.Entities[0].Id)
}

// TestServiceCreateEntityWithMetadata tests creating an entity with metadata
func TestServiceCreateEntityWithMetadata(t *testing.T) {
	ctx := context.Background()

	metadata := make(map[string]*anypb.Any)
	metadata["department"] = &anypb.Any{
		TypeUrl: "type.googleapis.com/google.protobuf.StringValue",
		Value:   []byte("Engineering"),
	}

	entity := &pb.Entity{
		Id: "service_test_entity_10",
		Kind: &pb.Kind{
			Major: "Person",
			Minor: "Employee",
		},
		Name:     createNameValue("2025-03-18T00:00:00Z", "Grace"),
		Created:  "2025-03-18T00:00:00Z",
		Metadata: metadata,
	}

	resp, err := server.CreateEntity(ctx, entity)
	if err != nil {
		t.Fatalf("CreateEntity() error = %v", err)
	}
	if resp == nil {
		t.Fatal("CreateEntity() returned nil response")
	}

	// Read back with metadata
	readReq := &pb.ReadEntityRequest{
		Entity: &pb.Entity{Id: "service_test_entity_10"},
		Output: []string{"metadata"},
	}

	readResp, err := server.ReadEntity(ctx, readReq)
	if err != nil {
		t.Fatalf("ReadEntity() error = %v", err)
	}
	if readResp == nil {
		t.Fatal("ReadEntity() returned nil response")
	}
	if len(readResp.Metadata) == 0 {
		t.Error("ReadEntity() returned no metadata")
	}

	log.Printf("Successfully created and read entity with metadata: %v", readResp.Id)
}

// TestServiceUpdateEntityAddRelationship tests adding a relationship to an existing entity via UpdateEntity
func TestServiceUpdateEntityAddRelationship(t *testing.T) {
	ctx := context.Background()

	// Create two entities first
	entity1 := &pb.Entity{
		Id: "service_test_entity_11",
		Kind: &pb.Kind{
			Major: "Person",
			Minor: "Employee",
		},
		Name:    createNameValue("2025-03-18T00:00:00Z", "Henry"),
		Created: "2025-03-18T00:00:00Z",
	}

	entity2 := &pb.Entity{
		Id: "service_test_entity_12",
		Kind: &pb.Kind{
			Major: "Person",
			Minor: "Employee",
		},
		Name:    createNameValue("2025-03-18T00:00:00Z", "Iris"),
		Created: "2025-03-18T00:00:00Z",
	}

	_, err := server.CreateEntity(ctx, entity1)
	if err != nil {
		t.Fatalf("CreateEntity(entity1) error = %v", err)
	}

	_, err = server.CreateEntity(ctx, entity2)
	if err != nil {
		t.Fatalf("CreateEntity(entity2) error = %v", err)
	}

	// Now update entity1 to add a relationship to entity2
	updateReq := &pb.UpdateEntityRequest{
		Id: "service_test_entity_11",
		Entity: &pb.Entity{
			Id: "service_test_entity_11",
			Relationships: map[string]*pb.Relationship{
				"service_test_rel_3": {
					Id:              "service_test_rel_3",
					Name:            "MANAGES",
					RelatedEntityId: "service_test_entity_12",
					StartTime:       "2025-04-01T00:00:00Z",
				},
			},
		},
	}

	updateResp, err := server.UpdateEntity(ctx, updateReq)
	if err != nil {
		t.Fatalf("UpdateEntity() error = %v", err)
	}
	if updateResp == nil {
		t.Fatal("UpdateEntity() returned nil response")
	}

	// Read back to verify relationship was added
	readReq := &pb.ReadEntityRequest{
		Entity: &pb.Entity{Id: "service_test_entity_11"},
		Output: []string{"relationships"},
	}

	readResp, err := server.ReadEntity(ctx, readReq)
	if err != nil {
		t.Fatalf("ReadEntity() error = %v", err)
	}
	if len(readResp.Relationships) == 0 {
		t.Error("ReadEntity() returned no relationships after update")
	}
	if _, exists := readResp.Relationships["service_test_rel_3"]; !exists {
		t.Error("Expected relationship 'service_test_rel_3' not found after update")
	}

	log.Printf("Successfully added relationship via UpdateEntity: %v", readResp.Id)
}

// TestServiceUpdateEntityModifyRelationship tests updating an existing relationship via UpdateEntity
func TestServiceUpdateEntityModifyRelationship(t *testing.T) {
	ctx := context.Background()

	// Create two entities with a relationship
	entity1 := &pb.Entity{
		Id: "service_test_entity_13",
		Kind: &pb.Kind{
			Major: "Person",
			Minor: "Employee",
		},
		Name:    createNameValue("2025-03-18T00:00:00Z", "Jack"),
		Created: "2025-03-18T00:00:00Z",
	}

	entity2 := &pb.Entity{
		Id: "service_test_entity_14",
		Kind: &pb.Kind{
			Major: "Person",
			Minor: "Employee",
		},
		Name:    createNameValue("2025-03-18T00:00:00Z", "Kate"),
		Created: "2025-03-18T00:00:00Z",
		Relationships: map[string]*pb.Relationship{
			"service_test_rel_4": {
				Id:              "service_test_rel_4",
				Name:            "WORKS_WITH",
				RelatedEntityId: "service_test_entity_13",
				StartTime:       "2025-03-18T00:00:00Z",
			},
		},
	}

	_, err := server.CreateEntity(ctx, entity1)
	if err != nil {
		t.Fatalf("CreateEntity(entity1) error = %v", err)
	}

	_, err = server.CreateEntity(ctx, entity2)
	if err != nil {
		t.Fatalf("CreateEntity(entity2 with relationship) error = %v", err)
	}

	// Update the relationship to add an EndTime date
	updateReq := &pb.UpdateEntityRequest{
		Id: "service_test_entity_14",
		Entity: &pb.Entity{
			Id: "service_test_entity_14",
			Relationships: map[string]*pb.Relationship{
				"service_test_rel_4": {
					Id:      "service_test_rel_4",
					EndTime: "2025-12-31T00:00:00Z",
				},
			},
		},
	}

	updateResp, err := server.UpdateEntity(ctx, updateReq)
	if err != nil {
		t.Fatalf("UpdateEntity() error = %v", err)
	}
	if updateResp == nil {
		t.Fatal("UpdateEntity() returned nil response")
	}

	// Read back to verify relationship was updated
	readReq := &pb.ReadEntityRequest{
		Entity: &pb.Entity{
			Id: "service_test_entity_14",
			Relationships: map[string]*pb.Relationship{
				"service_test_rel_4": {Id: "service_test_rel_4"},
			},
		},
		Output: []string{"relationships"},
	}

	readResp, err := server.ReadEntity(ctx, readReq)
	if err != nil {
		t.Fatalf("ReadEntity() error = %v", err)
	}
	if len(readResp.Relationships) == 0 {
		t.Error("ReadEntity() returned no relationships")
	}

	rel, exists := readResp.Relationships["service_test_rel_4"]
	if !exists {
		t.Fatal("Expected relationship 'service_test_rel_4' not found")
	}
	if rel.EndTime != "2025-12-31T00:00:00Z" {
		t.Errorf("Relationship EndTime = %v, want 2025-12-31T00:00:00Z", rel.EndTime)
	}

	log.Printf("Successfully updated relationship via UpdateEntity: %v", readResp.Id)
}

// TestServiceUpdateEntityMultipleRelationships tests adding multiple relationships via UpdateEntity
func TestServiceUpdateEntityMultipleRelationships(t *testing.T) {
	ctx := context.Background()

	// Create three entities
	entity1 := &pb.Entity{
		Id: "service_test_entity_15",
		Kind: &pb.Kind{
			Major: "Person",
			Minor: "Manager",
		},
		Name:    createNameValue("2025-03-18T00:00:00Z", "Leo"),
		Created: "2025-03-18T00:00:00Z",
	}

	entity2 := &pb.Entity{
		Id: "service_test_entity_16",
		Kind: &pb.Kind{
			Major: "Person",
			Minor: "Employee",
		},
		Name:    createNameValue("2025-03-18T00:00:00Z", "Mia"),
		Created: "2025-03-18T00:00:00Z",
	}

	entity3 := &pb.Entity{
		Id: "service_test_entity_17",
		Kind: &pb.Kind{
			Major: "Person",
			Minor: "Employee",
		},
		Name:    createNameValue("2025-03-18T00:00:00Z", "Nina"),
		Created: "2025-03-18T00:00:00Z",
	}

	_, err := server.CreateEntity(ctx, entity1)
	if err != nil {
		t.Fatalf("CreateEntity(entity1) error = %v", err)
	}

	_, err = server.CreateEntity(ctx, entity2)
	if err != nil {
		t.Fatalf("CreateEntity(entity2) error = %v", err)
	}

	_, err = server.CreateEntity(ctx, entity3)
	if err != nil {
		t.Fatalf("CreateEntity(entity3) error = %v", err)
	}

	// Update entity1 to add relationships to both entity2 and entity3
	updateReq := &pb.UpdateEntityRequest{
		Id: "service_test_entity_15",
		Entity: &pb.Entity{
			Id: "service_test_entity_15",
			Relationships: map[string]*pb.Relationship{
				"service_test_rel_5": {
					Id:              "service_test_rel_5",
					Name:            "SUPERVISES",
					RelatedEntityId: "service_test_entity_16",
					StartTime:       "2025-04-01T00:00:00Z",
				},
				"service_test_rel_6": {
					Id:              "service_test_rel_6",
					Name:            "SUPERVISES",
					RelatedEntityId: "service_test_entity_17",
					StartTime:       "2025-04-01T00:00:00Z",
				},
			},
		},
	}

	updateResp, err := server.UpdateEntity(ctx, updateReq)
	if err != nil {
		t.Fatalf("UpdateEntity() error = %v", err)
	}
	if updateResp == nil {
		t.Fatal("UpdateEntity() returned nil response")
	}

	// Read back to verify both relationships were added
	readReq := &pb.ReadEntityRequest{
		Entity: &pb.Entity{Id: "service_test_entity_15"},
		Output: []string{"relationships"},
	}

	readResp, err := server.ReadEntity(ctx, readReq)
	if err != nil {
		t.Fatalf("ReadEntity() error = %v", err)
	}
	if len(readResp.Relationships) < 2 {
		t.Errorf("ReadEntity() returned %d relationships, want at least 2", len(readResp.Relationships))
	}

	// Verify both relationships exist
	if _, exists := readResp.Relationships["service_test_rel_5"]; !exists {
		t.Error("Expected relationship 'service_test_rel_5' not found")
	}
	if _, exists := readResp.Relationships["service_test_rel_6"]; !exists {
		t.Error("Expected relationship 'service_test_rel_6' not found")
	}

	log.Printf("Successfully added multiple relationships via UpdateEntity: %v", readResp.Id)
}

// TestServiceCreateEntityWithMultipleRelationships tests creating an entity with multiple relationships at once
func TestServiceCreateEntityWithMultipleRelationships(t *testing.T) {
	ctx := context.Background()

	// Create target entities first
	entity1 := &pb.Entity{
		Id: "service_test_entity_18",
		Kind: &pb.Kind{
			Major: "Organization",
			Minor: "Department",
		},
		Name:    createNameValue("2025-03-18T00:00:00Z", "Engineering"),
		Created: "2025-03-18T00:00:00Z",
	}

	entity2 := &pb.Entity{
		Id: "service_test_entity_19",
		Kind: &pb.Kind{
			Major: "Organization",
			Minor: "Department",
		},
		Name:    createNameValue("2025-03-18T00:00:00Z", "Sales"),
		Created: "2025-03-18T00:00:00Z",
	}

	_, err := server.CreateEntity(ctx, entity1)
	if err != nil {
		t.Fatalf("CreateEntity(entity1) error = %v", err)
	}

	_, err = server.CreateEntity(ctx, entity2)
	if err != nil {
		t.Fatalf("CreateEntity(entity2) error = %v", err)
	}

	// Create a person entity with relationships to both departments
	entity3 := &pb.Entity{
		Id: "service_test_entity_20",
		Kind: &pb.Kind{
			Major: "Person",
			Minor: "Employee",
		},
		Name:    createNameValue("2025-03-18T00:00:00Z", "Oscar"),
		Created: "2025-03-18T00:00:00Z",
		Relationships: map[string]*pb.Relationship{
			"service_test_rel_7": {
				Id:              "service_test_rel_7",
				Name:            "MEMBER_OF",
				RelatedEntityId: "service_test_entity_18",
				StartTime:       "2025-01-01T00:00:00Z",
				EndTime:         "2025-06-30T00:00:00Z",
			},
			"service_test_rel_8": {
				Id:              "service_test_rel_8",
				Name:            "MEMBER_OF",
				RelatedEntityId: "service_test_entity_19",
				StartTime:       "2025-07-01T00:00:00Z",
			},
		},
	}

	resp, err := server.CreateEntity(ctx, entity3)
	if err != nil {
		t.Fatalf("CreateEntity(entity with multiple relationships) error = %v", err)
	}
	if resp == nil {
		t.Fatal("CreateEntity() returned nil response")
	}

	// Read back to verify both relationships were created
	readReq := &pb.ReadEntityRequest{
		Entity: &pb.Entity{Id: "service_test_entity_20"},
		Output: []string{"relationships"},
	}

	readResp, err := server.ReadEntity(ctx, readReq)
	if err != nil {
		t.Fatalf("ReadEntity() error = %v", err)
	}
	if len(readResp.Relationships) < 2 {
		t.Errorf("ReadEntity() returned %d relationships, want at least 2", len(readResp.Relationships))
	}

	// Verify both relationships exist
	rel1, exists1 := readResp.Relationships["service_test_rel_7"]
	if !exists1 {
		t.Error("Expected relationship 'service_test_rel_7' not found")
	} else if rel1.EndTime != "2025-06-30T00:00:00Z" {
		t.Errorf("Relationship 'service_test_rel_7' EndTime = %v, want 2025-06-30T00:00:00Z", rel1.EndTime)
	}

	if _, exists2 := readResp.Relationships["service_test_rel_8"]; !exists2 {
		t.Error("Expected relationship 'service_test_rel_8' not found")
	}

	log.Printf("Successfully created entity with multiple relationships: %v", readResp.Id)
}

// TestServiceCreateRelationshipWithDuplicateId tests that creating a relationship with a duplicate ID fails
func TestServiceCreateRelationshipWithDuplicateId(t *testing.T) {
	ctx := context.Background()

	// Create two entities
	entity1 := &pb.Entity{
		Id: "service_dup_entity_1",
		Kind: &pb.Kind{
			Major: "Person",
			Minor: "Employee",
		},
		Name:    createNameValue("2025-04-01T00:00:00Z", "Paul"),
		Created: "2025-04-01T00:00:00Z",
	}

	entity2 := &pb.Entity{
		Id: "service_dup_entity_2",
		Kind: &pb.Kind{
			Major: "Person",
			Minor: "Employee",
		},
		Name:    createNameValue("2025-04-01T00:00:00Z", "Quinn"),
		Created: "2025-04-01T00:00:00Z",
	}

	_, err := server.CreateEntity(ctx, entity1)
	if err != nil {
		t.Fatalf("CreateEntity(entity1) error = %v", err)
	}

	_, err = server.CreateEntity(ctx, entity2)
	if err != nil {
		t.Fatalf("CreateEntity(entity2) error = %v", err)
	}

	// Create a relationship with a specific ID
	entity3 := &pb.Entity{
		Id: "service_dup_entity_3",
		Kind: &pb.Kind{
			Major: "Person",
			Minor: "Employee",
		},
		Name:    createNameValue("2025-04-01T00:00:00Z", "Rachel"),
		Created: "2025-04-01T00:00:00Z",
		Relationships: map[string]*pb.Relationship{
			"service_duplicate_rel_id": {
				Id:              "service_duplicate_rel_id",
				Name:            "WORKS_WITH",
				RelatedEntityId: "service_dup_entity_2",
				StartTime:       "2025-04-01T00:00:00Z",
			},
		},
	}

	_, err = server.CreateEntity(ctx, entity3)
	if err != nil {
		t.Fatalf("CreateEntity(entity3 with relationship) error = %v", err)
	}

	// Read the original relationship to verify its properties
	readReq := &pb.ReadEntityRequest{
		Entity: &pb.Entity{
			Id: "service_dup_entity_3",
			Relationships: map[string]*pb.Relationship{
				"service_duplicate_rel_id": {Id: "service_duplicate_rel_id"},
			},
		},
		Output: []string{"relationships"},
	}

	readResp, err := server.ReadEntity(ctx, readReq)
	if err != nil {
		t.Fatalf("ReadEntity() error = %v", err)
	}

	originalRel := readResp.Relationships["service_duplicate_rel_id"]
	originalStartTime := originalRel.StartTime

	// Attempt to create another entity with a relationship with the SAME ID (should fail)
	entity4 := &pb.Entity{
		Id: "service_dup_entity_4",
		Kind: &pb.Kind{
			Major: "Person",
			Minor: "Employee",
		},
		Name:    createNameValue("2025-04-01T00:00:00Z", "Sam"),
		Created: "2025-04-01T00:00:00Z",
		Relationships: map[string]*pb.Relationship{
			"service_duplicate_rel_id": {
				Id:              "service_duplicate_rel_id", // Same ID!
				Name:            "MANAGES",                  // Different type
				RelatedEntityId: "service_dup_entity_2",
				StartTime:       "2025-05-01T00:00:00Z", // Different start time
			},
		},
	}

	_, err = server.CreateEntity(ctx, entity4)
	if err == nil {
		t.Error("Expected error when creating entity with duplicate relationship ID, but got none")
	} else {
		log.Printf("Duplicate relationship creation failed as expected: %v", err)
	}

	// Verify the original relationship was NOT modified
	verifyResp, err := server.ReadEntity(ctx, readReq)
	if err != nil {
		t.Fatalf("ReadEntity() error after duplicate attempt = %v", err)
	}

	verifyRel := verifyResp.Relationships["service_duplicate_rel_id"]
	if verifyRel.Name != "WORKS_WITH" {
		t.Errorf("Relationship type changed from WORKS_WITH to %v", verifyRel.Name)
	}
	if verifyRel.StartTime != originalStartTime {
		t.Errorf("Relationship StartTime changed from %v to %v", originalStartTime, verifyRel.StartTime)
	}

	log.Printf("Successfully verified that duplicate relationship IDs are rejected")
}

// TestServiceUpdateNonExistentRelationship tests that updating a non-existent relationship creates it
func TestServiceUpdateNonExistentRelationship(t *testing.T) {
	ctx := context.Background()

	// Create two entities
	entity1 := &pb.Entity{
		Id: "service_nonexistent_test_entity_1",
		Kind: &pb.Kind{
			Major: "Person",
			Minor: "Employee",
		},
		Name:    createNameValue("2025-04-01T00:00:00Z", "Tom"),
		Created: "2025-04-01T00:00:00Z",
	}

	entity2 := &pb.Entity{
		Id: "service_nonexistent_test_entity_2",
		Kind: &pb.Kind{
			Major: "Person",
			Minor: "Employee",
		},
		Name:    createNameValue("2025-04-01T00:00:00Z", "Uma"),
		Created: "2025-04-01T00:00:00Z",
	}

	_, err := server.CreateEntity(ctx, entity1)
	if err != nil {
		t.Fatalf("CreateEntity(entity1) error = %v", err)
	}

	_, err = server.CreateEntity(ctx, entity2)
	if err != nil {
		t.Fatalf("CreateEntity(entity2) error = %v", err)
	}

	// Count relationships before update (should be 0)
	readReq := &pb.ReadEntityRequest{
		Entity: &pb.Entity{Id: "service_nonexistent_test_entity_1"},
		Output: []string{"relationships"},
	}

	readResp, err := server.ReadEntity(ctx, readReq)
	if err != nil {
		t.Fatalf("ReadEntity() error = %v", err)
	}
	countBefore := len(readResp.Relationships)

	// Try to update a relationship that doesn't exist (should create it)
	updateReq := &pb.UpdateEntityRequest{
		Id: "service_nonexistent_test_entity_1",
		Entity: &pb.Entity{
			Id: "service_nonexistent_test_entity_1",
			Relationships: map[string]*pb.Relationship{
				"service_nonexistent_rel_id": {
					Id:              "service_nonexistent_rel_id",
					Name:            "MENTORS",
					RelatedEntityId: "service_nonexistent_test_entity_2",
					StartTime:       "2025-04-01T00:00:00Z",
					EndTime:         "2025-12-31T00:00:00Z",
				},
			},
		},
	}

	_, err = server.UpdateEntity(ctx, updateReq)
	if err != nil {
		t.Fatalf("UpdateEntity() error = %v (expected to succeed and create relationship)", err)
	}

	// Verify the relationship was created
	readResp, err = server.ReadEntity(ctx, readReq)
	if err != nil {
		t.Fatalf("ReadEntity() after update error = %v", err)
	}

	countAfter := len(readResp.Relationships)
	if countAfter != countBefore+1 {
		t.Errorf("Expected 1 new relationship, but count changed from %d to %d", countBefore, countAfter)
	}

	// Verify the new relationship exists with correct properties
	rel := readResp.Relationships["service_nonexistent_rel_id"]
	if rel == nil {
		t.Fatal("Relationship 'service_nonexistent_rel_id' not found after update")
	}
	if rel.Name != "MENTORS" {
		t.Errorf("Relationship Name = %v, want MENTORS", rel.Name)
	}
	if rel.RelatedEntityId != "service_nonexistent_test_entity_2" {
		t.Errorf("Relationship RelatedEntityId = %v, want service_nonexistent_test_entity_2", rel.RelatedEntityId)
	}
	if rel.EndTime != "2025-12-31T00:00:00Z" {
		t.Errorf("Relationship EndTime = %v, want 2025-12-31T00:00:00Z", rel.EndTime)
	}

	log.Printf("Successfully verified that updating non-existent relationship creates it")
}

// TestServiceUpdateRelationshipValidFields tests updating valid fields (StartTime/EndTime) on relationships
func TestServiceUpdateRelationshipValidFields(t *testing.T) {
	ctx := context.Background()

	// Create two entities with a relationship
	entity1 := &pb.Entity{
		Id: "service_update_valid_entity_1",
		Kind: &pb.Kind{
			Major: "Person",
			Minor: "Employee",
		},
		Name:    createNameValue("2025-04-01T00:00:00Z", "Uma"),
		Created: "2025-04-01T00:00:00Z",
	}

	entity2 := &pb.Entity{
		Id: "service_update_valid_entity_2",
		Kind: &pb.Kind{
			Major: "Person",
			Minor: "Employee",
		},
		Name:    createNameValue("2025-04-01T00:00:00Z", "Victor"),
		Created: "2025-04-01T00:00:00Z",
		Relationships: map[string]*pb.Relationship{
			"service_update_valid_rel": {
				Id:              "service_update_valid_rel",
				Name:            "REPORTS_TO",
				RelatedEntityId: "service_update_valid_entity_1",
				StartTime:       "2025-04-01T00:00:00Z",
			},
		},
	}

	_, err := server.CreateEntity(ctx, entity1)
	if err != nil {
		t.Fatalf("CreateEntity(entity1) error = %v", err)
	}

	_, err = server.CreateEntity(ctx, entity2)
	if err != nil {
		t.Fatalf("CreateEntity(entity2) error = %v", err)
	}

	// Count relationships before update
	readReq := &pb.ReadEntityRequest{
		Entity: &pb.Entity{Id: "service_update_valid_entity_2"},
		Output: []string{"relationships"},
	}

	readResp, err := server.ReadEntity(ctx, readReq)
	if err != nil {
		t.Fatalf("ReadEntity() error = %v", err)
	}
	countBefore := len(readResp.Relationships)

	// Update only StartTime (Created date)
	updateReq := &pb.UpdateEntityRequest{
		Id: "service_update_valid_entity_2",
		Entity: &pb.Entity{
			Id: "service_update_valid_entity_2",
			Relationships: map[string]*pb.Relationship{
				"service_update_valid_rel": {
					Id:        "service_update_valid_rel",
					StartTime: "2025-03-15T00:00:00Z",
				},
			},
		},
	}

	_, err = server.UpdateEntity(ctx, updateReq)
	if err != nil {
		t.Fatalf("UpdateEntity(update StartTime) error = %v", err)
	}

	// Verify the relationship was updated
	readResp, err = server.ReadEntity(ctx, readReq)
	if err != nil {
		t.Fatalf("ReadEntity() after StartTime update error = %v", err)
	}

	if len(readResp.Relationships) != countBefore {
		t.Errorf("Relationship count changed from %d to %d after update", countBefore, len(readResp.Relationships))
	}

	rel := readResp.Relationships["service_update_valid_rel"]
	if rel.StartTime != "2025-03-15T00:00:00Z" {
		t.Errorf("Relationship StartTime = %v, want 2025-03-15T00:00:00Z", rel.StartTime)
	}

	// Update only EndTime (Terminated date)
	updateReq2 := &pb.UpdateEntityRequest{
		Id: "service_update_valid_entity_2",
		Entity: &pb.Entity{
			Id: "service_update_valid_entity_2",
			Relationships: map[string]*pb.Relationship{
				"service_update_valid_rel": {
					Id:      "service_update_valid_rel",
					EndTime: "2025-12-31T00:00:00Z",
				},
			},
		},
	}

	_, err = server.UpdateEntity(ctx, updateReq2)
	if err != nil {
		t.Fatalf("UpdateEntity(update EndTime) error = %v", err)
	}

	// Verify EndTime was updated
	readResp, err = server.ReadEntity(ctx, readReq)
	if err != nil {
		t.Fatalf("ReadEntity() after EndTime update error = %v", err)
	}

	rel = readResp.Relationships["service_update_valid_rel"]
	if rel.EndTime != "2025-12-31T00:00:00Z" {
		t.Errorf("Relationship EndTime = %v, want 2025-12-31T00:00:00Z", rel.EndTime)
	}

	// Verify Name and RelatedEntityId haven't changed
	if rel.Name != "REPORTS_TO" {
		t.Errorf("Relationship Name changed to %v, want REPORTS_TO", rel.Name)
	}
	if rel.RelatedEntityId != "service_update_valid_entity_1" {
		t.Errorf("Relationship RelatedEntityId changed to %v, want service_update_valid_entity_1", rel.RelatedEntityId)
	}

	// Verify no new relationships were created
	if len(readResp.Relationships) != countBefore {
		t.Errorf("Relationship count changed from %d to %d after updates", countBefore, len(readResp.Relationships))
	}

	log.Printf("Successfully updated relationship with valid fields (StartTime/EndTime)")
}

// TestServiceUpdateRelationshipNoNewCreations tests that updating relationships doesn't create new ones
func TestServiceUpdateRelationshipNoNewCreations(t *testing.T) {
	ctx := context.Background()

	// Create entity with relationships
	entity1 := &pb.Entity{
		Id: "service_no_new_entity_1",
		Kind: &pb.Kind{
			Major: "Person",
			Minor: "Employee",
		},
		Name:    createNameValue("2025-04-01T00:00:00Z", "Wendy"),
		Created: "2025-04-01T00:00:00Z",
	}

	entity2 := &pb.Entity{
		Id: "service_no_new_entity_2",
		Kind: &pb.Kind{
			Major: "Person",
			Minor: "Employee",
		},
		Name:    createNameValue("2025-04-01T00:00:00Z", "Xander"),
		Created: "2025-04-01T00:00:00Z",
		Relationships: map[string]*pb.Relationship{
			"service_no_new_rel_1": {
				Id:              "service_no_new_rel_1",
				Name:            "COLLABORATES_WITH",
				RelatedEntityId: "service_no_new_entity_1",
				StartTime:       "2025-04-01T00:00:00Z",
			},
		},
	}

	_, err := server.CreateEntity(ctx, entity1)
	if err != nil {
		t.Fatalf("CreateEntity(entity1) error = %v", err)
	}

	_, err = server.CreateEntity(ctx, entity2)
	if err != nil {
		t.Fatalf("CreateEntity(entity2) error = %v", err)
	}

	// Count relationships before update
	readReq := &pb.ReadEntityRequest{
		Entity: &pb.Entity{Id: "service_no_new_entity_2"},
		Output: []string{"relationships"},
	}

	readResp, err := server.ReadEntity(ctx, readReq)
	if err != nil {
		t.Fatalf("ReadEntity() error = %v", err)
	}
	countBefore := len(readResp.Relationships)

	// Perform multiple updates
	for i := 0; i < 3; i++ {
		updateReq := &pb.UpdateEntityRequest{
			Id: "service_no_new_entity_2",
			Entity: &pb.Entity{
				Id: "service_no_new_entity_2",
				Relationships: map[string]*pb.Relationship{
					"service_no_new_rel_1": {
						Id:      "service_no_new_rel_1",
						EndTime: fmt.Sprintf("2025-12-%02dT00:00:00Z", 10+i),
					},
				},
			},
		}

		_, err = server.UpdateEntity(ctx, updateReq)
		if err != nil {
			t.Fatalf("UpdateEntity(iteration %d) error = %v", i, err)
		}
	}

	// Verify no new relationships were created
	readResp, err = server.ReadEntity(ctx, readReq)
	if err != nil {
		t.Fatalf("ReadEntity() after updates error = %v", err)
	}

	countAfter := len(readResp.Relationships)
	if countAfter != countBefore {
		t.Errorf("Relationship count changed from %d to %d after multiple updates", countBefore, countAfter)
	}

	// Verify the relationship still exists with the latest update
	rel := readResp.Relationships["service_no_new_rel_1"]
	if rel == nil {
		t.Fatal("Relationship 'service_no_new_rel_1' not found after updates")
	}
	if rel.EndTime != "2025-12-12T00:00:00Z" {
		t.Errorf("Relationship EndTime = %v, want 2025-12-12T00:00:00Z (last update)", rel.EndTime)
	}

	log.Printf("Successfully verified that updating relationships doesn't create new ones")
}

// TestServiceUpdateRelationshipBothFields tests updating both StartTime and EndTime together
func TestServiceUpdateRelationshipBothFields(t *testing.T) {
	ctx := context.Background()

	// Create two entities with a relationship
	entity1 := &pb.Entity{
		Id: "service_both_fields_entity_1",
		Kind: &pb.Kind{
			Major: "Person",
			Minor: "Employee",
		},
		Name:    createNameValue("2025-04-01T00:00:00Z", "Yara"),
		Created: "2025-04-01T00:00:00Z",
	}

	entity2 := &pb.Entity{
		Id: "service_both_fields_entity_2",
		Kind: &pb.Kind{
			Major: "Person",
			Minor: "Employee",
		},
		Name:    createNameValue("2025-04-01T00:00:00Z", "Zane"),
		Created: "2025-04-01T00:00:00Z",
		Relationships: map[string]*pb.Relationship{
			"service_both_fields_rel": {
				Id:              "service_both_fields_rel",
				Name:            "WORKS_WITH",
				RelatedEntityId: "service_both_fields_entity_1",
				StartTime:       "2025-04-01T00:00:00Z",
			},
		},
	}

	_, err := server.CreateEntity(ctx, entity1)
	if err != nil {
		t.Fatalf("CreateEntity(entity1) error = %v", err)
	}

	_, err = server.CreateEntity(ctx, entity2)
	if err != nil {
		t.Fatalf("CreateEntity(entity2) error = %v", err)
	}

	// Count relationships before update
	readReq := &pb.ReadEntityRequest{
		Entity: &pb.Entity{Id: "service_both_fields_entity_2"},
		Output: []string{"relationships"},
	}

	readResp, err := server.ReadEntity(ctx, readReq)
	if err != nil {
		t.Fatalf("ReadEntity() error = %v", err)
	}
	countBefore := len(readResp.Relationships)

	// Update both StartTime and EndTime
	updateReq := &pb.UpdateEntityRequest{
		Id: "service_both_fields_entity_2",
		Entity: &pb.Entity{
			Id: "service_both_fields_entity_2",
			Relationships: map[string]*pb.Relationship{
				"service_both_fields_rel": {
					Id:        "service_both_fields_rel",
					StartTime: "2025-02-01T00:00:00Z",
					EndTime:   "2025-11-30T00:00:00Z",
				},
			},
		},
	}

	_, err = server.UpdateEntity(ctx, updateReq)
	if err != nil {
		t.Fatalf("UpdateEntity(update both fields) error = %v", err)
	}

	// Verify both fields were updated
	readResp, err = server.ReadEntity(ctx, readReq)
	if err != nil {
		t.Fatalf("ReadEntity() after both fields update error = %v", err)
	}

	rel := readResp.Relationships["service_both_fields_rel"]
	if rel.StartTime != "2025-02-01T00:00:00Z" {
		t.Errorf("Relationship StartTime = %v, want 2025-02-01T00:00:00Z", rel.StartTime)
	}
	if rel.EndTime != "2025-11-30T00:00:00Z" {
		t.Errorf("Relationship EndTime = %v, want 2025-11-30T00:00:00Z", rel.EndTime)
	}

	// Verify Name and RelatedEntityId haven't changed
	if rel.Name != "WORKS_WITH" {
		t.Errorf("Relationship Name changed to %v, want WORKS_WITH", rel.Name)
	}
	if rel.RelatedEntityId != "service_both_fields_entity_1" {
		t.Errorf("Relationship RelatedEntityId changed to %v, want service_both_fields_entity_1", rel.RelatedEntityId)
	}

	// Verify no new relationships were created
	if len(readResp.Relationships) != countBefore {
		t.Errorf("Relationship count changed from %d to %d after update", countBefore, len(readResp.Relationships))
	}

	log.Printf("Successfully updated both StartTime and EndTime fields")
}

// TestServiceUpdateRelationshipInvalidFields tests that updating invalid fields (Name, RelatedEntityId) fails
func TestServiceUpdateRelationshipInvalidFields(t *testing.T) {
	ctx := context.Background()

	// Create two entities with a relationship
	entity1 := &pb.Entity{
		Id: "service_invalid_entity_1",
		Kind: &pb.Kind{
			Major: "Person",
			Minor: "Employee",
		},
		Name:    createNameValue("2025-04-01T00:00:00Z", "Alpha"),
		Created: "2025-04-01T00:00:00Z",
	}

	entity2 := &pb.Entity{
		Id: "service_invalid_entity_2",
		Kind: &pb.Kind{
			Major: "Person",
			Minor: "Employee",
		},
		Name:    createNameValue("2025-04-01T00:00:00Z", "Beta"),
		Created: "2025-04-01T00:00:00Z",
		Relationships: map[string]*pb.Relationship{
			"service_invalid_rel": {
				Id:              "service_invalid_rel",
				Name:            "MANAGES",
				RelatedEntityId: "service_invalid_entity_1",
				StartTime:       "2025-04-01T00:00:00Z",
			},
		},
	}

	entity3 := &pb.Entity{
		Id: "service_invalid_entity_3",
		Kind: &pb.Kind{
			Major: "Person",
			Minor: "Employee",
		},
		Name:    createNameValue("2025-04-01T00:00:00Z", "Gamma"),
		Created: "2025-04-01T00:00:00Z",
	}

	_, err := server.CreateEntity(ctx, entity1)
	if err != nil {
		t.Fatalf("CreateEntity(entity1) error = %v", err)
	}

	_, err = server.CreateEntity(ctx, entity2)
	if err != nil {
		t.Fatalf("CreateEntity(entity2) error = %v", err)
	}

	_, err = server.CreateEntity(ctx, entity3)
	if err != nil {
		t.Fatalf("CreateEntity(entity3) error = %v", err)
	}

	// Store original relationship properties
	readReq := &pb.ReadEntityRequest{
		Entity: &pb.Entity{
			Id: "service_invalid_entity_2",
			Relationships: map[string]*pb.Relationship{
				"service_invalid_rel": {Id: "service_invalid_rel"},
			},
		},
		Output: []string{"relationships"},
	}

	readResp, err := server.ReadEntity(ctx, readReq)
	if err != nil {
		t.Fatalf("ReadEntity() before update error = %v", err)
	}
	originalRel := readResp.Relationships["service_invalid_rel"]
	originalName := originalRel.Name
	originalRelatedEntityId := originalRel.RelatedEntityId

	// Try to update the relationship with invalid fields (Name and RelatedEntityId)
	// This should FAIL because only StartTime/EndTime are allowed
	updateReq := &pb.UpdateEntityRequest{
		Id: "service_invalid_entity_2",
		Entity: &pb.Entity{
			Id: "service_invalid_entity_2",
			Relationships: map[string]*pb.Relationship{
				"service_invalid_rel": {
					Id:              "service_invalid_rel",
					Name:            "SUPERVISES",               // Invalid: try to change relationship type
					RelatedEntityId: "service_invalid_entity_3", // Invalid: try to change target entity
				},
			},
		},
	}

	_, err = server.UpdateEntity(ctx, updateReq)
	if err == nil {
		t.Error("Expected error when trying to update invalid fields (Name, RelatedEntityId), but got none")
	} else {
		log.Printf("UpdateEntity failed when trying to update invalid fields (expected): %v", err)
	}

	// Read back to verify relationship hasn't changed
	readResp, err = server.ReadEntity(ctx, readReq)
	if err != nil {
		t.Fatalf("ReadEntity() after failed update error = %v", err)
	}

	rel := readResp.Relationships["service_invalid_rel"]
	if rel == nil {
		t.Fatal("Relationship 'service_invalid_rel' not found after failed update")
	}

	// Verify Name (relationship type) hasn't changed
	if rel.Name != originalName {
		t.Errorf("Relationship Name changed from %v to %v (should remain unchanged after failed update)", originalName, rel.Name)
	}

	// Verify RelatedEntityId hasn't changed
	if rel.RelatedEntityId != originalRelatedEntityId {
		t.Errorf("Relationship RelatedEntityId changed from %v to %v (should remain unchanged after failed update)", originalRelatedEntityId, rel.RelatedEntityId)
	}

	log.Printf("Successfully verified that updating invalid fields (Name/RelatedEntityId) fails")
}
