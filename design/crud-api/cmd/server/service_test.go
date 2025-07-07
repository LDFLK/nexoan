package main

import (
	"context"
	"log"
	"os"
	"testing"

	"lk/datafoundation/crud-api/db/config"
	mongorepository "lk/datafoundation/crud-api/db/repository/mongo"
	neo4jrepository "lk/datafoundation/crud-api/db/repository/neo4j"
	postgres "lk/datafoundation/crud-api/db/repository/postgres"
)

var server *Server

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

// FIXME: https://github.com/LDFLK/nexoan/issues/119

// TestEmptyEntity tests creating an entity with empty metadata, attributes, and relationships
// TODO: Think more about what is an empty entity, how empty it can be, define empty entity
// func TestEmptyEntity(t *testing.T) {
// 	// Create an entity with only ID field and empty data
// 	emptyEntity := &pb.Entity{
// 		Id: "empty-entity-test-001",

// 		// All other fields (metadata, attributes, relationships) are nil/empty
// 	}

// 	// Call CreateEntity
// 	ctx := context.Background()
// 	resp, err := server.CreateEntity(ctx, emptyEntity)

// 	// Validation should fail for Neo4j but MongoDB should succeed
// 	// The function should not return an error since MongoDB storage worked
// 	assert.NoError(t, err, "Expected no error even though Neo4j validation fails")
// 	assert.NotNil(t, resp, "Response should not be nil")
// 	assert.Equal(t, emptyEntity.Id, resp.Id, "Entity ID should match")

// 	// Now read back the entity from MongoDB to verify it was stored
// 	readEntity := &pb.Entity{
// 		Id: emptyEntity.Id,
// 	}
// 	readResp, err := server.ReadEntity(ctx, readEntity)
// 	assert.NoError(t, err, "Expected no error when reading back the entity")
// 	assert.NotNil(t, readResp, "Read response should not be nil")
// 	assert.Equal(t, emptyEntity.Id, readResp.Id, "Read entity ID should match")

// 	// Clean up - delete the test entity
// 	deleteReq := &pb.EntityId{
// 		Id: emptyEntity.Id,
// 	}
// 	_, err = server.DeleteEntity(ctx, deleteReq)
// 	assert.NoError(t, err, "Error deleting test entity")
// }

// TestCreateEntity tests the CreateEntity function
// TODO: FIX BUG @zaeema-n there is a bug in the CreateEntity function
func TestCreateEntity(t *testing.T) {

	// Encode Name as *anypb.Any
	// nameValue, err := anypb.New(&anypb.Any{Value: []byte("John Doe")})
	// assert.NoError(t, err)

	// // Define test cases
	// tests := []struct {
	// 	name    string
	// 	entity  *pb.Entity
	// 	wantErr bool
	// }{
	// 	{
	// 		name: "CreateEntity_Success",
	// 		entity: &pb.Entity{
	// 			Id: "123",
	// 			Kind: &pb.Kind{
	// 				Major: "Person",
	// 				Minor: "Employee",
	// 			},
	// 			Name:       &pb.TimeBasedValue{Value: nameValue},
	// 			Created:    "2025-03-20",
	// 			Terminated: "",
	// 		},
	// 		wantErr: false,
	// 	},
	// 	{
	// 		name: "CreateEntity_MissingId",
	// 		entity: &pb.Entity{
	// 			Id: "",
	// 			Kind: &pb.Kind{
	// 				Major: "Person",
	// 				Minor: "Employee",
	// 			},
	// 			Name:       &pb.TimeBasedValue{Value: nameValue},
	// 			Created:    "2025-03-20",
	// 			Terminated: "",
	// 		},
	// 		wantErr: true,
	// 	},
	// }

	// // Run test cases
	// for _, tt := range tests {
	// 	t.Run(tt.name, func(t *testing.T) {
	// 		ctx := context.Background()

	// 		resp, err := server.CreateEntity(ctx, tt.entity)
	// 		if (err != nil) != tt.wantErr {
	// 			t.Errorf("CreateEntity() error = %v, wantErr %v", err, tt.wantErr)
	// 		}

	// 		if !tt.wantErr {
	// 			// Verify the response matches the input entity
	// 			assert.Equal(t, tt.entity, resp, "Expected response to match the input entity")
	// 		}
	// 	})
	// }
}
