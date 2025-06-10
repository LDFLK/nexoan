package neo4jrepository

import (
	"context"
	"fmt"
	"log"

	pb "lk/datafoundation/crud-api/lk/datafoundation/crud-api" // Replace with your actual protobuf package

	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// GetEntityDetailsFromNeo4j retrieves entity information from Neo4j database
func (repo *Neo4jRepository) GetGraphEntity(ctx context.Context, entityId string) (*pb.Kind, *pb.TimeBasedValue, string, string, error) {
	// Try to get additional entity information from Neo4j
	var kind *pb.Kind
	var name *pb.TimeBasedValue
	var created string
	var terminated string

	// Attempt to read from Neo4j, but don't fail if it's not available
	entityMap, err := repo.ReadGraphEntity(ctx, entityId)
	if err == nil && entityMap != nil {

		// Entity found in Neo4j, extract information
		if majorKindValue, ok := entityMap["MajorKind"]; ok {
			kind = &pb.Kind{
				Major: majorKindValue.(string),
			}
		}

		if minorKindValue, ok := entityMap["MinorKind"]; ok {
			if kind == nil {
				kind = &pb.Kind{}
			}
			kind.Minor = minorKindValue.(string)
		}

		if nameValue, ok := entityMap["Name"]; ok {
			// Create a TimeBasedValue with string value
			value, _ := anypb.New(&wrapperspb.StringValue{
				Value: nameValue.(string),
			})

			name = &pb.TimeBasedValue{
				StartTime: entityMap["Created"].(string),
				Value:     value,
			}

			// Add EndTime if available
			if termValue, ok := entityMap["Terminated"]; ok {
				name.EndTime = termValue.(string)
			}
		}

		if createdValue, ok := entityMap["Created"]; ok {
			created = createdValue.(string)
		}

		if termValue, ok := entityMap["Terminated"]; ok {
			terminated = termValue.(string)
		}
	}

	return kind, name, created, terminated, err
}

// GetGraphRelationships retrieves relationships for an entity from Neo4j
func (repo *Neo4jRepository) GetGraphRelationships(ctx context.Context, entityId string) (map[string]*pb.Relationship, error) {
	relationships := make(map[string]*pb.Relationship)
	// Retrieve relationships from Neo4j
	relData, err := repo.ReadRelationships(ctx, entityId)
	if err != nil {
		log.Printf("[neo4j_handler.GetGraphRelationships] Error reading relationships for entity %s: %v", entityId, err)
		return relationships, fmt.Errorf("[neo4j_handler.GetGraphRelationships] error reading relationships: %v", err)
	}

	// Process each relationship
	// TODO: Holding relationship and defining the content needs to be
	//  revalidated. Discuss and confirm.
	//  Also build a rule based validation for the relationship content.
	for _, rel := range relData {
		relType, ok1 := rel["type"].(string)
		relatedID, ok2 := rel["relatedID"].(string)
		created, ok3 := rel["Created"].(string)
		relID, ok4 := rel["relationshipID"].(string)
		direction, ok5 := rel["direction"].(string)

		if !ok1 || !ok2 || !ok3 || !ok4 || !ok5 || direction != "OUTGOING" {
			continue // Skip if any required field is missing
		}

		// Create relationship object
		relationship := &pb.Relationship{
			Id:              relID,
			Name:            relType,
			RelatedEntityId: relatedID,
			StartTime:       created,
		}

		// Add termination date if available
		if terminated, ok := rel["Terminated"].(string); ok && terminated != "" {
			relationship.EndTime = terminated
		}

		// Store in map with unique key
		relationships[relID] = relationship
	}

	return relationships, nil
}

func (repo *Neo4jRepository) GetRelationshipsByName(ctx context.Context, entityId string, relationship string, ts string) (map[string]*pb.Relationship, error) {
	// Validate input parameters
	if entityId == "" {
		return nil, fmt.Errorf("entityId cannot be empty")
	}
	if relationship == "" {
		return nil, fmt.Errorf("relationship type cannot be empty")
	}
	if ts == "" {
		return nil, fmt.Errorf("timestamp cannot be empty")
	}

	// Call ReadRelatedGraphEntityIds from neo4j_client.go
	relationshipData, err := repo.ReadRelatedGraphEntityIds(ctx, entityId, relationship, ts)
	if err != nil {
		log.Printf("[GetEntityIdsByRelationship] Error fetching related relationships for entity %s with relationship %s: %v", entityId, relationship, err)
		return nil, err
	}

	// Convert the list of relationships into a map[string]*pb.Relationship
	relationships := make(map[string]*pb.Relationship)
	for _, rel := range relationshipData {
		relID, ok1 := rel["Id"].(string)
		relatedEntityID, ok2 := rel["RelatedEntityId"].(string)
		startTime, ok3 := rel["StartTime"].(string)
		endTime, _ := rel["EndTime"].(string) // Optional field
		name, ok4 := rel["Name"].(string)

		// Ensure required fields are present
		if !ok1 || !ok2 || !ok3 || !ok4 {
			log.Printf("[GetEntityIdsByRelationship] Skipping relationship due to missing required fields: %v", rel)
			continue
		}

		// Create a pb.Relationship object
		relationships[relID] = &pb.Relationship{
			Id:              relID,
			RelatedEntityId: relatedEntityID,
			StartTime:       startTime,
			EndTime:         endTime,
			Name:            name,
		}
	}

	return relationships, nil
}

// validateGraphEntityCreation checks if an entity has all required fields for Neo4j storage
func validateGraphEntityCreation(entity *pb.Entity) bool {
	// Check if Kind is present and has a Major value
	if entity.Kind == nil || entity.Kind.GetMajor() == "" {
		log.Printf("[neo4j_handler.validateGraphEntityCreation] Skipping Neo4j entity creation for %s: Missing or empty Kind.Major", entity.Id)
		return false
	}

	// Check if Name is present and has a Value
	if entity.Name == nil || entity.Name.GetValue() == nil {
		log.Printf("[neo4j_handler.validateGraphEntityCreation] Skipping Neo4j entity creation for %s: Missing or empty Name.Value", entity.Id)
		return false
	}

	// Check if Created date is present
	if entity.Created == "" {
		log.Printf("[neo4j_handler.validateGraphEntityCreation] Skipping Neo4j entity creation for %s: Missing Created date", entity.Id)
		return false
	}

	return true
}

// HandleGraphEntityCreation creates a new entity in Neo4j
func (repo *Neo4jRepository) HandleGraphEntityCreation(ctx context.Context, entity *pb.Entity) (bool, error) {
	// Validate required fields for Neo4j entity creation
	if !validateGraphEntityCreation(entity) {
		log.Printf("[neo4j_handler.HandleGraphEntityCreation] Entity %s saved in MongoDB only, skipping Neo4j due to missing required fields", entity.Id)
		return false, fmt.Errorf("[neo4j_handler.HandleGraphEntityCreation] missing required fields for Neo4j entity creation")
	}

	log.Printf("[neo4j_handler.HandleGraphEntityCreation] Creating new entity in Neo4j: %s", entity.Id)

	// Prepare data for Neo4j with safety checks
	entityMap := map[string]interface{}{
		"Id": entity.Id,
	}

	// Validate and extract the Kind field
	if entity.Kind == nil || entity.Kind.GetMajor() == "" || entity.Kind.GetMinor() == "" {
		return false, fmt.Errorf("[neo4j_handler.HandleGraphEntityCreation] missing or invalid Kind.Major or Kind.Minor for entity %s", entity.Id)
	}

	kind := &pb.Kind{
		Major: entity.Kind.GetMajor(),
		Minor: entity.Kind.GetMinor(),
	}

	if entity.Name != nil && entity.Name.GetValue() != nil {
		// Unpack the Any value to get the actual string
		var stringValue wrapperspb.StringValue
		err := entity.Name.GetValue().UnmarshalTo(&stringValue)
		if err != nil {
			// If we can't unmarshal to StringValue, try to get the raw value
			if entity.Name.GetValue().GetTypeUrl() == "type.googleapis.com/google.protobuf.StringValue" {
				// The value is already a StringValue, try to get the raw bytes
				rawValue := entity.Name.GetValue().GetValue()
				if len(rawValue) > 0 {
					// The first byte is the length, followed by the actual string
					if len(rawValue) > 1 {
						entityMap["Name"] = string(rawValue[1:])
						log.Printf("Using raw value from Any: %s\n", string(rawValue[1:]))
					}
				}
			} else {
				fmt.Printf("Error unpacking Name value for entity %s: %v\n", entity.Id, err)
				return false, fmt.Errorf("[neo4j_handler.HandleGraphEntityCreation] error unpacking Name value: %v", err)
			}
		} else {
			// Successfully unpacked to StringValue
			entityMap["Name"] = stringValue.Value
			log.Printf("Using unpacked StringValue: %s\n", stringValue.Value)
		}
	}

	// Handle other fields
	if entity.Created != "" {
		entityMap["Created"] = entity.Created
	}

	if entity.Terminated != "" {
		entityMap["Terminated"] = entity.Terminated
	}

	// Create the entity
	result, err := repo.CreateGraphEntity(ctx, kind, entityMap)
	if err != nil {
		log.Printf("[neo4j_handler.HandleGraphEntityCreation] Error creating entity in Neo4j: %v", err)
		return false, err
	} else {
		log.Printf("[neo4j_handler.HandleGraphEntityCreation] Successfully created entity in Neo4j: %s", entity.Id)
		return result != nil, nil // Success if we got a non-nil result
	}
}

// HandleGraphEntityUpdate updates an existing entity in Neo4j
func (repo *Neo4jRepository) HandleGraphEntityUpdate(ctx context.Context, entity *pb.Entity) (bool, error) {
	// Validate required fields for Neo4j entity update
	if entity.Id == "" {
		log.Printf("[neo4j_handler.HandleGraphEntityUpdate] Entity ID is required for Neo4j entity update")
		return false, fmt.Errorf("[neo4j_handler.HandleGraphEntityUpdate] entity ID is required")
	}

	log.Printf("[neo4j_handler.HandleGraphEntityUpdate] Updating existing entity in Neo4j: %s", entity.Id)

	// Prepare data for Neo4j with safety checks
	entityMap := map[string]interface{}{
		"Id": entity.Id,
	}

	// Handle Name field safely
	if entity.Name != nil && entity.Name.GetValue() != nil {
		// Unpack the Any value to get the actual string
		var stringValue wrapperspb.StringValue
		err := entity.Name.GetValue().UnmarshalTo(&stringValue)
		if err != nil {
			log.Printf("[neo4j_handler.HandleGraphEntityUpdate] Error unpacking Name value for entity %s: %v", entity.Id, err)
			return false, fmt.Errorf("[neo4j_handler.HandleGraphEntityUpdate] error unpacking Name value: %v", err)
		}
		// Get the actual string value from the StringValue and check it's not empty
		if stringValue.Value != "" {
			entityMap["Name"] = stringValue.Value
		}
	}

	// Handle other fields
	if entity.Created != "" {
		entityMap["Created"] = entity.Created
	}

	if entity.Terminated != "" {
		entityMap["Terminated"] = entity.Terminated
	}

	// Update the entity
	result, err := repo.UpdateGraphEntity(ctx, entity.Id, entityMap)
	log.Printf("[neo4j_handler.HandleGraphEntityUpdate] Entity map for update: %+v", entityMap)
	if err != nil {
		log.Printf("[neo4j_handler.HandleGraphEntityUpdate] Error updating entity in Neo4j: %v", err)
		return false, err
	} else {
		log.Printf("[neo4j_handler.HandleGraphEntityUpdate] Successfully updated entity in Neo4j: %s", entity.Id)
		log.Printf("[neo4j_handler.HandleGraphEntityUpdate] Update result: %+v", result)
		return result != nil, nil // Success if we got a non-nil result
	}
}

// HandleGraphRelationshipsCreate handles creating new relationships
func (repo *Neo4jRepository) HandleGraphRelationshipsCreate(ctx context.Context, entity *pb.Entity) error {
	if len(entity.Relationships) == 0 {
		log.Printf("[neo4j_handler.HandleGraphRelationshipsCreate] No relationships to process for entity: %s", entity.Id)
		return nil
	}

	log.Printf("[neo4j_handler.HandleGraphRelationshipsCreate] Processing %d relationships for entity: %s", len(entity.Relationships), entity.Id)

	// First, process all child entities
	for _, relationship := range entity.Relationships {
		if relationship == nil || relationship.RelatedEntityId == "" {
			continue
		}

		// Check if the child entity exists
		childEntityMap, err := repo.ReadGraphEntity(ctx, relationship.RelatedEntityId)
		if err != nil || childEntityMap == nil {
			log.Printf("[neo4j_handler.HandleGraphRelationshipsCreate] Child entity %s does not exist in Neo4j. Make sure to create it first.",
				relationship.RelatedEntityId)
			return fmt.Errorf("[neo4j_handler.HandleGraphRelationshipsCreate] child entity %s does not exist", relationship.RelatedEntityId)
		}
		log.Printf("[neo4j_handler.HandleGraphRelationshipsCreate] Child entity %s exists in Neo4j", relationship.RelatedEntityId)

		// Create the relationship
		_, err = repo.CreateRelationship(ctx, entity.Id, relationship)
		if err != nil {
			log.Printf("[neo4j_handler.HandleGraphRelationshipsCreate] Error creating relationship from %s to %s: %v",
				entity.Id, relationship.RelatedEntityId, err)
			return err
		}
		log.Printf("[neo4j_handler.HandleGraphRelationshipsCreate] Successfully created relationship from %s to %s",
			entity.Id, relationship.RelatedEntityId)
	}

	return nil
}

// HandleGraphRelationshipsUpdate handles updating existing relationships
func (repo *Neo4jRepository) HandleGraphRelationshipsUpdate(ctx context.Context, entity *pb.Entity) error {
	log.Printf("[neo4j_handler.HandleGraphRelationshipsUpdate] Received entity: %+v", entity)

	if len(entity.Relationships) == 0 {
		log.Printf("[neo4j_handler.HandleGraphRelationshipsUpdate] No relationships to process for entity: %s", entity.Id)
		return nil
	}

	log.Printf("[neo4j_handler.HandleGraphRelationshipsUpdate] Processing %d relationships for entity: %s", len(entity.Relationships), entity.Id)

	// First verify the parent entity exists
	parentEntity, err := repo.ReadGraphEntity(ctx, entity.Id)
	if err != nil || parentEntity == nil {
		log.Printf("[neo4j_handler.HandleGraphRelationshipsUpdate] Parent entity %s does not exist in Neo4j", entity.Id)
		return fmt.Errorf("[neo4j_handler.HandleGraphRelationshipsUpdate] parent entity %s does not exist", entity.Id)
	}

	for _, relationship := range entity.Relationships {
		if relationship == nil {
			continue
		}

		// For updates, we only need the ID
		if relationship.Id != "" {
			relationshipData := map[string]interface{}{
				"Terminated": relationship.EndTime,
			}
			log.Printf("[neo4j_handler.HandleGraphRelationshipsUpdate] Update data: %+v", relationshipData)
			// Try to update if we have an ID
			_, err = repo.UpdateRelationship(ctx, relationship.Id, relationshipData)
			if err == nil {
				log.Printf("[neo4j_handler.HandleGraphRelationshipsUpdate] Successfully updated relationship %s", relationship.Id)
				continue
			}
			log.Printf("[neo4j_handler.HandleGraphRelationshipsUpdate] Failed to update relationship: %v", err)
		}

		// For creation, we need the related entity ID
		if relationship.RelatedEntityId == "" {
			continue
		}

		// Check if the child entity exists
		childEntityMap, err := repo.ReadGraphEntity(ctx, relationship.RelatedEntityId)
		if err != nil || childEntityMap == nil {
			log.Printf("[neo4j_handler.HandleGraphRelationshipsUpdate] Child entity %s does not exist in Neo4j. Make sure to create it first.",
				relationship.RelatedEntityId)
			return fmt.Errorf("[neo4j_handler.HandleGraphRelationshipsUpdate] child entity %s does not exist", relationship.RelatedEntityId)
		}
		log.Printf("[neo4j_handler.HandleGraphRelationshipsUpdate] Child entity %s exists in Neo4j", relationship.RelatedEntityId)

		// Either no ID or update failed, try to create
		_, createErr := repo.CreateRelationship(ctx, entity.Id, relationship)
		if createErr != nil {
			log.Printf("[neo4j_handler.HandleGraphRelationshipsUpdate] Error creating relationship from %s to %s: %v",
				entity.Id, relationship.RelatedEntityId, createErr)
			return createErr
		}
		log.Printf("[neo4j_handler.HandleGraphRelationshipsUpdate] Successfully created new relationship from %s to %s",
			entity.Id, relationship.RelatedEntityId)
	}

	return nil
}

// HandleGraphEntityFilter processes a ReadEntityRequest and calls FilterEntities
func (repo *Neo4jRepository) HandleGraphEntityFilter(ctx context.Context, req *pb.ReadEntityRequest) ([]map[string]interface{}, error) {
	if req == nil || req.Entity == nil {
		return nil, fmt.Errorf("invalid request: ReadEntityRequest or Entity is nil")
	}

	// Extract filters from the request
	filters := make(map[string]interface{})

	// If ID is present, only use that filter
	if req.Entity.Id != "" {
		filters["id"] = req.Entity.Id
	} else {
		// Only add other filters if we're not filtering by ID
		// Add name if present
		if req.Entity.Name != nil && req.Entity.Name.Value != nil {
			var stringValue wrapperspb.StringValue
			if err := req.Entity.Name.Value.UnmarshalTo(&stringValue); err == nil {
				filters["name"] = stringValue.Value
			}
		}

		// Add created timestamp if present
		if req.Entity.Created != "" {
			filters["created"] = req.Entity.Created
		}

		// Add terminated timestamp if present
		if req.Entity.Terminated != "" {
			filters["terminated"] = req.Entity.Terminated
		}
	}

	// Call FilterEntities with the extracted filters
	return repo.FilterEntities(ctx, req.Entity.Kind, filters)
}
