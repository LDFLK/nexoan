package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

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

	log.Printf("[server.UpdateEntity] Updating Entity: %s", updateEntityID)

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
	if req.Entity == nil || req.Entity.Kind == nil || req.Entity.Kind.Major == "" {
		return nil, fmt.Errorf("Kind.Major is required for filtering entities")
	}

	log.Printf("Filtering entities by Kind.Major: %s", req.Entity.Kind.Major)

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

// TabularDataRequest represents incoming tabular data
type TabularDataRequest struct {
	Headers     []string            // Column headers
	Data        [][]string          // Row data
	Validation  *ValidationRules    // Optional validation rules
	Options     *ProcessingOptions  // Optional processing options
}

// ValidationRules defines validation requirements
type ValidationRules struct {
	RequiredFields    []string          // Fields that must have values
	UniqueFields     []string          // Fields that must be unique
	PatternRules     map[string]string // Regex patterns for fields
	MaxRows          int               // Maximum allowed rows
	MinRows          int               // Minimum required rows
}

// ProcessingOptions defines data processing options
type ProcessingOptions struct {
	TrimWhitespace    bool   // Whether to trim whitespace
	NullValue         string // String to treat as null
	DateFormat        string // Expected date format
	SanitizeSpecials  bool   // Whether to sanitize special characters
}

// TabularValidationResult represents validation result
type TabularValidationResult struct {
	IsValid         bool
	ErrorMessages   []string
	ColumnTypes     map[string]string
	EntityType      string
	SanitizedData   [][]string           // Cleaned data
	Statistics      *ValidationStats      // Validation statistics
	Warnings        []string             // Non-critical issues
}

// ValidationStats provides statistical information
type ValidationStats struct {
	TotalRows        int
	ValidRows        int
	InvalidRows      int
	EmptyFields      map[string]int      // Count of empty fields per column
	UniqueValues     map[string]int      // Count of unique values per column
	DataQualityScore float64             // Overall quality score (0-1)
}

// sanitizeData cleans the input data
func (s *Server) sanitizeData(req *TabularDataRequest) [][]string {
	sanitized := make([][]string, len(req.Data))
	options := req.Options
	if options == nil {
		options = &ProcessingOptions{
			TrimWhitespace: true,
			NullValue: "NULL",
			DateFormat: "2006-01-02",
			SanitizeSpecials: true,
		}
	}

	for i, row := range req.Data {
		sanitized[i] = make([]string, len(row))
		for j, cell := range row {
			// Apply sanitization based on options
			value := cell
			if options.TrimWhitespace {
				value = strings.TrimSpace(value)
			}
			if options.SanitizeSpecials {
				value = sanitizeSpecialChars(value)
			}
			if value == "" || value == options.NullValue {
				value = ""
			}
			sanitized[i][j] = value
		}
	}
	return sanitized
}

// sanitizeSpecialChars removes or escapes special characters
func sanitizeSpecialChars(value string) string {
	// Remove control characters
	value = strings.Map(func(r rune) rune {
		if r < 32 || r == 127 {
			return -1
		}
		return r
	}, value)

	// Escape special characters
	value = strings.ReplaceAll(value, "\\'", "'")
	value = strings.ReplaceAll(value, "\\\"", "\"")
	
	return value
}

// validateDataRules checks data against validation rules
func (s *Server) validateDataRules(req *TabularDataRequest, sanitizedData [][]string) []string {
	var errors []string
	rules := req.Validation
	if rules == nil {
		return errors
	}

	// Check row count limits
	if rules.MaxRows > 0 && len(sanitizedData) > rules.MaxRows {
		errors = append(errors, fmt.Sprintf("row count exceeds maximum limit of %d", rules.MaxRows))
	}
	if rules.MinRows > 0 && len(sanitizedData) < rules.MinRows {
		errors = append(errors, fmt.Sprintf("row count below minimum requirement of %d", rules.MinRows))
	}

	// Create header index map
	headerIndex := make(map[string]int)
	for i, header := range req.Headers {
		headerIndex[header] = i
	}

	// Check required fields
	for _, field := range rules.RequiredFields {
		if idx, exists := headerIndex[field]; exists {
			for rowIdx, row := range sanitizedData {
				if row[idx] == "" {
					errors = append(errors, fmt.Sprintf("required field '%s' is empty in row %d", field, rowIdx+1))
				}
			}
		} else {
			errors = append(errors, fmt.Sprintf("required field '%s' not found in headers", field))
		}
	}

	// Check unique fields
	for _, field := range rules.UniqueFields {
		if idx, exists := headerIndex[field]; exists {
			values := make(map[string]int)
			for rowIdx, row := range sanitizedData {
				if row[idx] != "" {
					if prevRow, isDuplicate := values[row[idx]]; isDuplicate {
						errors = append(errors, fmt.Sprintf("duplicate value '%s' in field '%s' at rows %d and %d", 
							row[idx], field, prevRow+1, rowIdx+1))
					}
					values[row[idx]] = rowIdx
				}
			}
		}
	}

	// Check pattern rules
	for field, pattern := range rules.PatternRules {
		if idx, exists := headerIndex[field]; exists {
			regex, err := regexp.Compile(pattern)
			if err != nil {
				errors = append(errors, fmt.Sprintf("invalid pattern for field '%s': %v", field, err))
				continue
			}
			for rowIdx, row := range sanitizedData {
				if row[idx] != "" && !regex.MatchString(row[idx]) {
					errors = append(errors, fmt.Sprintf("value '%s' in field '%s' at row %d does not match required pattern", 
						row[idx], field, rowIdx+1))
				}
			}
		}
	}

	return errors
}

// calculateStatistics generates validation statistics
func (s *Server) calculateStatistics(req *TabularDataRequest, sanitizedData [][]string) *ValidationStats {
	stats := &ValidationStats{
		TotalRows:    len(sanitizedData),
		ValidRows:    len(sanitizedData),
		InvalidRows:  0,
		EmptyFields:  make(map[string]int),
		UniqueValues: make(map[string]int),
	}

	// Calculate empty and unique values per column
	for colIdx, header := range req.Headers {
		uniqueValues := make(map[string]bool)
		emptyCount := 0
		
		for _, row := range sanitizedData {
			if row[colIdx] == "" {
				emptyCount++
			} else {
				uniqueValues[row[colIdx]] = true
			}
		}
		
		stats.EmptyFields[header] = emptyCount
		stats.UniqueValues[header] = len(uniqueValues)
	}

	// Calculate data quality score (0-1)
	totalFields := len(req.Headers) * len(sanitizedData)
	emptyFields := 0
	for _, count := range stats.EmptyFields {
		emptyFields += count
	}
	
	if totalFields > 0 {
		stats.DataQualityScore = 1 - (float64(emptyFields) / float64(totalFields))
	}

	return stats
}

// Enhance the main HandleTabularData method
func (s *Server) HandleTabularData(ctx context.Context, req *TabularDataRequest) (*TabularValidationResult, error) {
	log.Printf("Processing tabular data with %d columns and %d rows", len(req.Headers), len(req.Data))
	
	result := &TabularValidationResult{
		IsValid:       true,
		ErrorMessages: []string{},
		ColumnTypes:   make(map[string]string),
		Warnings:      []string{},
	}

	// Step 1: Basic structure validation
	if err := s.validateTabularStructure(req); err != nil {
		result.IsValid = false
		result.ErrorMessages = append(result.ErrorMessages, err.Error())
		return result, nil
	}

	// Step 2: Data sanitization
	result.SanitizedData = s.sanitizeData(req)

	// Step 3: Apply validation rules
	if ruleErrors := s.validateDataRules(req, result.SanitizedData); len(ruleErrors) > 0 {
		result.IsValid = false
		result.ErrorMessages = append(result.ErrorMessages, ruleErrors...)
	}

	// Step 4: Type inference
	columnTypes, err := s.inferTabularTypes(ctx, req)
	if err != nil {
		result.IsValid = false
		result.ErrorMessages = append(result.ErrorMessages, fmt.Sprintf("type inference error: %v", err))
	} else {
		result.ColumnTypes = columnTypes
	}

	// Step 5: Entity type detection
	entityType, err := s.detectEntityType(req.Headers)
	if err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("entity detection warning: %v", err))
	}
	result.EntityType = entityType

	// Step 6: Calculate statistics
	result.Statistics = s.calculateStatistics(req, result.SanitizedData)

	return result, nil
}

// validateTabularStructure validates basic table structure
func (s *Server) validateTabularStructure(req *TabularDataRequest) error {
	if len(req.Headers) == 0 {
		return fmt.Errorf("no headers provided")
	}

	if len(req.Data) == 0 {
		return fmt.Errorf("no data rows provided")
	}

	// Check header uniqueness
	headerMap := make(map[string]bool)
	for _, header := range req.Headers {
		if header == "" {
			return fmt.Errorf("empty header found")
		}
		if headerMap[header] {
			return fmt.Errorf("duplicate header found: %s", header)
		}
		headerMap[header] = true
	}

	// Check row consistency
	headerCount := len(req.Headers)
	for i, row := range req.Data {
		if len(row) != headerCount {
			return fmt.Errorf("inconsistent column count in row %d: expected %d, got %d", 
				i+1, headerCount, len(row))
		}
	}

	return nil
}

// inferTabularTypes infers column types using existing type inference
func (s *Server) inferTabularTypes(ctx context.Context, req *TabularDataRequest) (map[string]string, error) {
	columnTypes := make(map[string]string)
	
	// Use sample rows for type inference
	sampleSize := 5
	if len(req.Data) < sampleSize {
		sampleSize = len(req.Data)
	}

	for colIndex, header := range req.Headers {
		// Get sample values for this column
		samples := make([]string, sampleSize)
		for i := 0; i < sampleSize; i++ {
			samples[i] = req.Data[i][colIndex]
		}
		
		// Infer type for this column
		colType := inferColumnType(samples)
		columnTypes[header] = colType
	}

	return columnTypes, nil
}

// detectEntityType tries to determine the entity type from headers
func (s *Server) detectEntityType(headers []string) (string, error) {
	// Common entity indicators in headers
	entityIndicators := map[string]string{
		"employee": "Employee",
		"user":    "User",
		"product": "Product",
		"order":   "Order",
	}

	// Look for ID patterns
	for _, header := range headers {
		headerLower := strings.ToLower(header)
		for indicator, entityType := range entityIndicators {
			if strings.Contains(headerLower, indicator) && strings.Contains(headerLower, "id") {
				return entityType, nil
			}
		}
	}

	return "Unknown", nil
}

// inferColumnType determines the type of data in a column
func inferColumnType(samples []string) string {
	isNumber := true
	isDate := true
	
	for _, sample := range samples {
		// Skip empty values
		if sample == "" {
			continue
		}

		// Check if number
		if _, err := strconv.ParseFloat(sample, 64); err != nil {
			isNumber = false
		}

		// Check if date (simple check)
		if _, err := time.Parse("2006-01-02", sample); err != nil {
			isDate = false
		}
	}

	if isDate {
		return "date"
	}
	if isNumber {
		return "number"
	}
	return "string"
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
