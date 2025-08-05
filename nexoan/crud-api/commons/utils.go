package commons

import (
	"encoding/json"
	pb "lk/datafoundation/crud-api/lk/datafoundation/crud-api"
	"lk/datafoundation/crud-api/pkg/storageinference"
	"log"
	"time"

	"context"
	"fmt"
	"os"

	"lk/datafoundation/crud-api/db/config"
	mongorepository "lk/datafoundation/crud-api/db/repository/mongo"
	neo4jrepository "lk/datafoundation/crud-api/db/repository/neo4j"

	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// CreateTimeBasedValue creates a TimeBasedValue with a string value
func CreateTimeBasedValue(startTime, endTime, value string) *pb.TimeBasedValue {
	return &pb.TimeBasedValue{
		StartTime: startTime,
		EndTime:   endTime,
		Value:     ConvertStringToAny(value),
	}
}

func ConvertStringToAny(value string) *anypb.Any {
	any, _ := anypb.New(&wrapperspb.StringValue{
		Value: value,
	})
	return any
}

// ExtractStringFromAny extracts a string value from *anypb.Any
func ExtractStringFromAny(anyValue *anypb.Any) string {
	if anyValue == nil {
		return ""
	}
	var wrapper wrapperspb.StringValue
	if err := anyValue.UnmarshalTo(&wrapper); err == nil {
		return wrapper.Value
	}
	return ""
}

// ConvertStorageTypeStringToEnum converts a storage type string to StorageType enum
func ConvertStorageTypeStringToEnum(storageTypeStr string) storageinference.StorageType {
	switch storageTypeStr {
	case "TabularData":
		return storageinference.TabularData
	case "GraphData":
		return storageinference.GraphData
	case "MapData":
		return storageinference.MapData
	case "ListData":
		return storageinference.ListData
	case "ScalarData":
		return storageinference.ScalarData
	default:
		return storageinference.UnknownData
	}
}

// ExtractAttributeMetadataFields extracts attribute metadata fields from a MongoDB entity
func ExtractAttributeMetadataFields(entity *pb.Entity) (storageTypeStr, storagePath, updatedStr string, schemaMap map[string]interface{}) {
	// attribute_id, attribute_name, storage_type, storage_path, updated, schema
	if entity == nil || entity.Metadata == nil {
		return "", "", "", make(map[string]interface{})
	}

	metadataMap := entity.Metadata

	// Extract basic fields
	storageTypeStr = ExtractStringFromAny(metadataMap["storage_type"])
	storagePath = ExtractStringFromAny(metadataMap["storage_path"])
	updatedStr = ExtractStringFromAny(metadataMap["updated"])
	schemaStr := ExtractStringFromAny(metadataMap["schema"])

	// Convert schema JSON string to map
	schemaMap, err := ConvertJSONStringToMap(schemaStr)
	if err != nil {
		// Return empty map if conversion fails
		schemaMap = make(map[string]interface{})
	}

	return storageTypeStr, storagePath, updatedStr, schemaMap
}

// ConvertMapToAnyMap converts map[string]interface{} to map[string]*anypb.Any
func ConvertMapToAnyMap(input map[string]interface{}) map[string]*anypb.Any {
	result := make(map[string]*anypb.Any)

	for key, value := range input {
		switch v := value.(type) {
		case string:
			result[key] = ConvertStringToAny(v)
		case int, int32, int64:
			result[key] = ConvertStringToAny(fmt.Sprintf("%v", v))
		case float32, float64:
			result[key] = ConvertStringToAny(fmt.Sprintf("%v", v))
		case bool:
			result[key] = ConvertStringToAny(fmt.Sprintf("%v", v))
		case map[string]interface{}:
			// Handle nested maps recursively
			nestedMap := ConvertMapToAnyMap(v)
			// For nested maps, we'll store them as individual key-value pairs with a prefix
			for nestedKey, nestedValue := range nestedMap {
				result[key+"_"+nestedKey] = nestedValue
			}
		case []interface{}:
			// Handle slices by converting to JSON string
			result[key] = ConvertStringToAny(fmt.Sprintf("%v", v))
		default:
			// For complex types, convert to JSON string
			result[key] = ConvertStringToAny(fmt.Sprintf("%v", v))
		}
	}

	return result
}

// ConvertMapToAny converts map[string]interface{} to *anypb.Any
func ConvertMapToAny(input map[string]interface{}) *anypb.Any {
	// For now, convert the map to a JSON-like string representation
	// In a more sophisticated implementation, you could use a proper protobuf message
	jsonStr := fmt.Sprintf("%v", input)
	return ConvertStringToAny(jsonStr)
}

// ConvertJSONStringToMap converts a JSON string to map[string]interface{}
func ConvertJSONStringToMap(jsonStr string) (map[string]interface{}, error) {
	if jsonStr == "" {
		return make(map[string]interface{}), nil
	}

	var result map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &result)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON string: %w", err)
	}
	return result, nil
}

// ParseTimestamp parses a timestamp string to time.Time, returns zero value if empty or invalid
func ParseTimestamp(timestampStr string, context string) time.Time {
	if timestampStr == "" {
		return time.Time{} // Return zero value for empty strings
	}

	parsed, err := time.Parse(time.RFC3339, timestampStr)
	if err != nil {
		log.Printf("[Commons.ParseTimestamp] Warning: Failed to parse timestamp '%s' for %s, using zero value", timestampStr, context)
		return time.Time{} // Return zero value for invalid timestamps
	}

	return parsed
}

// GetNeo4jConfig creates a Neo4jConfig from environment variables
func GetNeo4jConfig() *config.Neo4jConfig {
	return &config.Neo4jConfig{
		URI:      os.Getenv("NEO4J_URI"),
		Username: os.Getenv("NEO4J_USER"),
		Password: os.Getenv("NEO4J_PASSWORD"),
	}
}

// GetMongoConfig creates a MongoConfig from environment variables
func GetMongoConfig() *config.MongoConfig {
	return &config.MongoConfig{
		URI:        os.Getenv("MONGO_URI"),
		DBName:     os.Getenv("MONGO_DB_NAME"),
		Collection: os.Getenv("MONGO_COLLECTION"),
	}
}

// GetNeo4jRepository retrieves a Neo4j repository
func GetNeo4jRepository(ctx context.Context) (*neo4jrepository.Neo4jRepository, error) {
	cfg := GetNeo4jConfig()
	repo, err := neo4jrepository.NewNeo4jRepository(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("[Commons] failed to create Neo4j repository: %w", err)
	}
	return repo, nil
}

// GetMongoRepository retrieves a Mongo repository
// TODO: Handle errors better
func GetMongoRepository(ctx context.Context) *mongorepository.MongoRepository {
	cfg := GetMongoConfig()
	return mongorepository.NewMongoRepository(ctx, cfg)
}
