package main

import (
	"context"
	"os"
	"testing"

// 	"lk/datafoundation/crud-api/db/config"
// 	mongorepository "lk/datafoundation/crud-api/db/repository/mongo"
// 	neo4jrepository "lk/datafoundation/crud-api/db/repository/neo4j"
// )

var server *Server

// TestMain sets up the actual MongoDB and Neo4j repositories before running the tests
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

	// Create the server with the initialized repositories
	server = &Server{
		mongoRepo: mongoRepo,
		neo4jRepo: neo4jRepo,
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

func TestHandleTabularData(t *testing.T) {
	tests := []struct {
		name    string
		req     *TabularDataRequest
		want    *TabularValidationResult
		wantErr bool
	}{
		{
			name: "Valid Employee Data with Validation Rules",
			req: &TabularDataRequest{
				Headers: []string{"EmployeeID", "Name", "Salary", "JoinDate"},
				Data: [][]string{
					{"EMP001", "John Doe", "75000", "2023-01-15"},
					{"EMP002", "Jane Smith", "85000", "2023-02-01"},
				},
				Validation: &ValidationRules{
					RequiredFields: []string{"EmployeeID", "Name"},
					UniqueFields:   []string{"EmployeeID"},
					PatternRules:   map[string]string{
						"EmployeeID": "^EMP\\d{3}$",
					},
					MinRows: 1,
					MaxRows: 1000,
				},
				Options: &ProcessingOptions{
					TrimWhitespace:   true,
					NullValue:        "NULL",
					DateFormat:       "2006-01-02",
					SanitizeSpecials: true,
				},
			},
			want: &TabularValidationResult{
				IsValid: true,
				ColumnTypes: map[string]string{
					"EmployeeID": "string",
					"Name":       "string",
					"Salary":     "number",
					"JoinDate":   "date",
				},
				EntityType: "Employee",
			},
			wantErr: false,
		},
		{
			name: "Data with Validation Errors",
			req: &TabularDataRequest{
				Headers: []string{"EmployeeID", "Name", "Salary"},
				Data: [][]string{
					{"INVALID", "", "75000"},
					{"EMP002", "", "NOT_NUMBER"},
				},
				Validation: &ValidationRules{
					RequiredFields: []string{"Name"},
					PatternRules:   map[string]string{
						"EmployeeID": "^EMP\\d{3}$",
					},
				},
			},
			want: &TabularValidationResult{
				IsValid:       false,
				ErrorMessages: []string{
					"required field 'Name' is empty in row 1",
					"required field 'Name' is empty in row 2",
					"value 'INVALID' in field 'EmployeeID' at row 1 does not match required pattern",
				},
			},
			wantErr: false,
		},
		{
			name: "Data with Special Characters",
			req: &TabularDataRequest{
				Headers: []string{"ID", "Description"},
				Data: [][]string{
					{"1", "Test\nwith\nnewlines"},
					{"2", "Test with \"quotes\""},
				},
				Options: &ProcessingOptions{
					TrimWhitespace:   true,
					SanitizeSpecials: true,
				},
			},
			want: &TabularValidationResult{
				IsValid: true,
			},
			wantErr: false,
		},
		{
			name: "Empty Data",
			req: &TabularDataRequest{
				Headers: []string{},
				Data:    [][]string{},
			},
			want: &TabularValidationResult{
				IsValid:       false,
				ErrorMessages: []string{"no headers provided"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{}
			got, err := s.HandleTabularData(context.Background(), tt.req)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("HandleTabularData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if got.IsValid != tt.want.IsValid {
				t.Errorf("HandleTabularData() IsValid = %v, want %v", got.IsValid, tt.want.IsValid)
			}
			
			// Check error messages if validation failed
			if !got.IsValid && tt.want.ErrorMessages != nil {
				if len(got.ErrorMessages) != len(tt.want.ErrorMessages) {
					t.Errorf("HandleTabularData() ErrorMessages count = %v, want %v", 
						len(got.ErrorMessages), len(tt.want.ErrorMessages))
				}
				for i, wantMsg := range tt.want.ErrorMessages {
					if i >= len(got.ErrorMessages) || got.ErrorMessages[i] != wantMsg {
						t.Errorf("HandleTabularData() ErrorMessage[%d] = %v, want %v", 
							i, got.ErrorMessages[i], wantMsg)
					}
				}
			}
			
			// Check statistics for valid cases
			if got.IsValid && got.Statistics != nil {
				if got.Statistics.TotalRows != len(tt.req.Data) {
					t.Errorf("HandleTabularData() Statistics.TotalRows = %v, want %v",
						got.Statistics.TotalRows, len(tt.req.Data))
				}
				
				if got.Statistics.DataQualityScore < 0 || got.Statistics.DataQualityScore > 1 {
					t.Errorf("HandleTabularData() Statistics.DataQualityScore = %v, should be between 0 and 1",
						got.Statistics.DataQualityScore)
				}
			}
			
			// Check column types for valid cases with specified types
			if tt.want.ColumnTypes != nil {
				for header, wantType := range tt.want.ColumnTypes {
					if gotType := got.ColumnTypes[header]; gotType != wantType {
						t.Errorf("HandleTabularData() column %s type = %v, want %v", 
							header, gotType, wantType)
					}
				}
			}
			
			if tt.want.EntityType != "" && got.EntityType != tt.want.EntityType {
				t.Errorf("HandleTabularData() EntityType = %v, want %v", 
					got.EntityType, tt.want.EntityType)
			}
		})
	}
}
