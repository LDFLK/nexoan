# Nexoan Architecture Diagrams

This document contains various architectural diagrams in Mermaid format for the Nexoan system.

---

## 1. System Architecture - High Level

```mermaid
graph TB
    subgraph "Client Layer"
        Client[Web/Mobile/Desktop Clients]
    end
    
    subgraph "API Layer"
        UpdateAPI[Update API<br/>Ballerina:8080<br/>CREATE/UPDATE/DELETE]
        QueryAPI[Query API<br/>Ballerina:8081<br/>READ/QUERY]
        SwaggerUI[Swagger UI<br/>API Documentation]
    end
    
    subgraph "Service Layer"
        CRUD[CRUD Service<br/>Go gRPC:50051]
        
        subgraph "CRUD Components"
            Server[gRPC Server<br/>- CreateEntity<br/>- ReadEntity<br/>- UpdateEntity<br/>- DeleteEntity]
            Engine[Engine<br/>- AttributeProcessor<br/>- GraphMetadataManager<br/>- TypeInference<br/>- StorageInference]
            Repos[Repository Layer<br/>- MongoRepo<br/>- Neo4jRepo<br/>- PostgresRepo]
        end
    end
    
    subgraph "Database Layer"
        MongoDB[(MongoDB:27017<br/>Metadata Storage<br/>Key-Value Pairs)]
        Neo4j[(Neo4j:7687<br/>Graph Storage<br/>Entities & Relationships)]
        PostgreSQL[(PostgreSQL:5432<br/>Attribute Storage<br/>Time-Series Data)]
    end
    
    subgraph "Supporting Services"
        Cleanup[Cleanup Service<br/>Database Cleanup]
        Backup[Backup/Restore<br/>Version Management]
    end
    
    Client -->|HTTP/REST + JSON| UpdateAPI
    Client -->|HTTP/REST + JSON| QueryAPI
    Client -.->|View Docs| SwaggerUI
    
    UpdateAPI -->|gRPC + Protobuf| CRUD
    QueryAPI -->|gRPC + Protobuf| CRUD
    
    CRUD --> Server
    Server --> Engine
    Server --> Repos
    
    Repos -->|BSON| MongoDB
    Repos -->|Bolt/Cypher| Neo4j
    Repos -->|SQL| PostgreSQL
    
    Cleanup -.->|Cleanup| MongoDB
    Cleanup -.->|Cleanup| Neo4j
    Cleanup -.->|Cleanup| PostgreSQL
    
    Backup -.->|Backup/Restore| MongoDB
    Backup -.->|Backup/Restore| Neo4j
    Backup -.->|Backup/Restore| PostgreSQL
    
    style Client fill:#e1f5ff
    style UpdateAPI fill:#fff4e6
    style QueryAPI fill:#fff4e6
    style CRUD fill:#f3e5f5
    style MongoDB fill:#e8f5e9
    style Neo4j fill:#e8f5e9
    style PostgreSQL fill:#e8f5e9
```

---

## 2. Create Entity Data Flow

```mermaid
sequenceDiagram
    participant Client
    participant UpdateAPI as Update API<br/>(Ballerina)
    participant CRUD as CRUD Service<br/>(Go gRPC)
    participant Mongo as MongoDB
    participant Neo4j
    participant Postgres as PostgreSQL
    
    Client->>UpdateAPI: POST /entities<br/>(JSON payload)
    activate UpdateAPI
    
    Note over UpdateAPI: Convert JSON<br/>to Protobuf Entity
    
    UpdateAPI->>CRUD: gRPC: CreateEntity(Entity)
    activate CRUD
    
    par Save to Databases
        CRUD->>Mongo: Save metadata
        activate Mongo
        Mongo-->>CRUD: Success
        deactivate Mongo
    and
        CRUD->>Neo4j: Create entity node
        activate Neo4j
        Neo4j-->>CRUD: Node created
        deactivate Neo4j
    and
        CRUD->>Neo4j: Create relationships
        activate Neo4j
        Neo4j-->>CRUD: Relationships created
        deactivate Neo4j
    and
        CRUD->>Postgres: Process & save attributes
        activate Postgres
        Note over Postgres: Infer types<br/>Create schemas<br/>Store values
        Postgres-->>CRUD: Attributes saved
        deactivate Postgres
    end
    
    CRUD-->>UpdateAPI: Entity (Protobuf)
    deactivate CRUD
    
    Note over UpdateAPI: Convert Protobuf<br/>to JSON
    
    UpdateAPI-->>Client: 201 Created<br/>(JSON response)
    deactivate UpdateAPI
```

---

## 3. Read Entity Data Flow

```mermaid
sequenceDiagram
    participant Client
    participant QueryAPI as Query API<br/>(Ballerina)
    participant CRUD as CRUD Service<br/>(Go gRPC)
    participant Mongo as MongoDB
    participant Neo4j
    participant Postgres as PostgreSQL
    
    Client->>QueryAPI: GET /entities/{id}?output=metadata,relationships
    activate QueryAPI
    
    QueryAPI->>CRUD: gRPC: ReadEntity(id, output=[metadata, relationships])
    activate CRUD
    
    Note over CRUD: Always fetch<br/>basic entity info
    CRUD->>Neo4j: Get entity (id, kind, name, created)
    activate Neo4j
    Neo4j-->>CRUD: Entity info
    deactivate Neo4j
    
    alt Output includes metadata
        CRUD->>Mongo: Get metadata by ID
        activate Mongo
        Mongo-->>CRUD: Metadata document
        deactivate Mongo
    end
    
    alt Output includes relationships
        CRUD->>Neo4j: Get relationships for entity
        activate Neo4j
        Neo4j-->>CRUD: Related entities
        deactivate Neo4j
    end
    
    alt Output includes attributes
        CRUD->>Postgres: Get attributes for entity
        activate Postgres
        Postgres-->>CRUD: Attribute values
        deactivate Postgres
    end
    
    Note over CRUD: Assemble complete<br/>entity from parts
    
    CRUD-->>QueryAPI: Entity (Protobuf)
    deactivate CRUD
    
    Note over QueryAPI: Convert Protobuf<br/>to JSON
    
    QueryAPI-->>Client: 200 OK<br/>(JSON response)
    deactivate QueryAPI
```

---

## 4. Component Architecture

```mermaid
graph TB
    subgraph "Update API - nexoan/update-api/"
        UA_Service[update_api_service.bal<br/>REST Endpoints]
        UA_Types[types_v1_pb.bal<br/>Protobuf Types]
        UA_Utils[utils/<br/>Helper Functions]
        
        UA_Service --> UA_Types
        UA_Service --> UA_Utils
    end
    
    subgraph "Query API - nexoan/query-api/"
        QA_Service[query_api_service.bal<br/>REST Endpoints]
        QA_Types[types_v1_pb.bal<br/>Protobuf Types]
        QA_Utils[types.bal<br/>Type Definitions]
        
        QA_Service --> QA_Types
        QA_Service --> QA_Utils
    end
    
    subgraph "CRUD Service - nexoan/crud-api/"
        subgraph "cmd/server/"
            Server[service.go<br/>gRPC Server Implementation]
            Utils[utils.go<br/>Helper Functions]
        end
        
        subgraph "engine/"
            AttrResolver[attribute_resolver.go<br/>Attribute Processing]
            GraphMgr[graph_metadata_manager.go<br/>Graph Metadata]
        end
        
        subgraph "pkg/"
            TypeInf[typeinference/<br/>Type Detection]
            StorageInf[storageinference/<br/>Storage Type Detection]
            Schema[schema/<br/>Schema Management]
        end
        
        subgraph "db/repository/"
            MongoRepo[mongo/<br/>mongodb_client.go<br/>metadata_handler.go]
            Neo4jRepo[neo4j/<br/>neo4j_client.go<br/>graph_entity_handler.go]
            PostgresRepo[postgres/<br/>postgres_client.go<br/>data_handler.go]
        end
        
        subgraph "protos/"
            Proto[types_v1.proto<br/>Protobuf Definitions]
        end
        
        Server --> AttrResolver
        Server --> GraphMgr
        Server --> MongoRepo
        Server --> Neo4jRepo
        Server --> PostgresRepo
        
        AttrResolver --> TypeInf
        AttrResolver --> StorageInf
        AttrResolver --> Schema
        AttrResolver --> PostgresRepo
    end
    
    UA_Service -->|gRPC| Server
    QA_Service -->|gRPC| Server
    
    MongoRepo -.->|BSON| MongoDB[(MongoDB)]
    Neo4jRepo -.->|Cypher| Neo4j[(Neo4j)]
    PostgresRepo -.->|SQL| PostgreSQL[(PostgreSQL)]
    
    style Server fill:#bbdefb
    style MongoRepo fill:#c8e6c9
    style Neo4jRepo fill:#c8e6c9
    style PostgresRepo fill:#c8e6c9
```

---

## 5. Data Storage Distribution

```mermaid
graph LR
    subgraph "Entity Data"
        Entity[Entity<br/>id: entity123<br/>kind: Person/Employee<br/>name: John Doe<br/>metadata: dept=Eng<br/>attributes: salary=100k<br/>relationships: reports_to]
    end
    
    subgraph "MongoDB"
        MetaDoc[metadata Collection<br/>{<br/>  _id: entity123,<br/>  metadata: {<br/>    department: Engineering,<br/>    role: Engineer<br/>  }<br/>}]
    end
    
    subgraph "Neo4j"
        EntityNode[entity123:Entity Node<br/>{<br/>  id: entity123,<br/>  kind_major: Person,<br/>  kind_minor: Employee,<br/>  name: John Doe,<br/>  created: 2024-01-01<br/>}]
        
        RelEdge[REPORTS_TO Relationship<br/>{<br/>  id: rel123,<br/>  startTime: 2024-01-01<br/>}]
        
        ManagerNode[manager123:Entity]
        
        EntityNode -->|RelEdge| ManagerNode
    end
    
    subgraph "PostgreSQL"
        AttrSchemas[attribute_schemas<br/>{<br/>  kind_major: Person,<br/>  attr_name: salary,<br/>  data_type: int,<br/>  storage_type: scalar<br/>}]
        
        EntityAttrs[entity_attributes<br/>{<br/>  entity_id: entity123,<br/>  attr_name: salary<br/>}]
        
        AttrTable[attr_Person_salary<br/>{<br/>  entity_id: entity123,<br/>  start_time: 2024-01,<br/>  end_time: NULL,<br/>  value: 100000<br/>}]
        
        AttrSchemas -.-> EntityAttrs
        EntityAttrs -.-> AttrTable
    end
    
    Entity -->|Metadata| MetaDoc
    Entity -->|Entity & Relationships| EntityNode
    Entity -->|Attributes| AttrSchemas
    
    style Entity fill:#fff9c4
    style MetaDoc fill:#c8e6c9
    style EntityNode fill:#bbdefb
    style RelEdge fill:#bbdefb
    style AttrSchemas fill:#f8bbd0
```

---

## 6. Type Inference Flow

```mermaid
flowchart TD
    Start([Receive Attribute Value]) --> CheckType{Check Value Type}
    
    CheckType -->|Number| IsDecimal{Has Decimal<br/>or Scientific?}
    IsDecimal -->|Yes| Float[Infer: float]
    IsDecimal -->|No| Integer[Infer: int]
    
    CheckType -->|String| CheckSpecial{Match Special<br/>Pattern?}
    CheckSpecial -->|Date Pattern| Date[Infer: date]
    CheckSpecial -->|Time Pattern| Time[Infer: time]
    CheckSpecial -->|DateTime Pattern| DateTime[Infer: datetime]
    CheckSpecial -->|No Match| String[Infer: string]
    
    CheckType -->|Boolean| Bool[Infer: bool]
    CheckType -->|Null| Null[Infer: null]
    CheckType -->|Array| Array[Infer: array<br/>Determine element type]
    CheckType -->|Object| Object[Infer: map/object<br/>Process nested]
    
    Float --> StorageInfer[Storage Type Inference]
    Integer --> StorageInfer
    Date --> StorageInfer
    Time --> StorageInfer
    DateTime --> StorageInfer
    String --> StorageInfer
    Bool --> StorageInfer
    Null --> StorageInfer
    Array --> StorageInfer
    Object --> StorageInfer
    
    StorageInfer --> CheckStructure{Check Structure}
    
    CheckStructure -->|Has columns & rows| Tabular[Storage: TABULAR<br/>→ PostgreSQL table]
    CheckStructure -->|Has nodes & edges| Graph[Storage: GRAPH<br/>→ Neo4j subgraph]
    CheckStructure -->|Has items array| List[Storage: LIST<br/>→ PostgreSQL array/JSONB]
    CheckStructure -->|Single value| Scalar[Storage: SCALAR<br/>→ PostgreSQL column]
    CheckStructure -->|Key-value pairs| Map[Storage: MAP<br/>→ PostgreSQL JSONB]
    
    Tabular --> SavePostgres[Save to PostgreSQL]
    Graph --> SaveNeo4j[Save to Neo4j]
    List --> SavePostgres
    Scalar --> SavePostgres
    Map --> SavePostgres
    
    SavePostgres --> End([Complete])
    SaveNeo4j --> End
    
    style Start fill:#e1f5ff
    style End fill:#c8e6c9
    style Tabular fill:#fff9c4
    style Graph fill:#fff9c4
    style List fill:#fff9c4
    style Scalar fill:#fff9c4
    style Map fill:#fff9c4
```

---

## 7. Deployment Architecture

```mermaid
graph TB
    subgraph "Docker Host"
        subgraph "ldf-network (Bridge Network)"
            subgraph "API Containers"
                UpdateCont[update<br/>Container<br/>Port: 8080]
                QueryCont[query<br/>Container<br/>Port: 8081]
            end
            
            subgraph "Service Containers"
                CrudCont[crud<br/>Container<br/>Port: 50051]
            end
            
            subgraph "Database Containers"
                MongoCont[mongodb<br/>Container<br/>Port: 27017]
                Neo4jCont[neo4j<br/>Container<br/>Ports: 7474, 7687]
                PostgresCont[postgres<br/>Container<br/>Port: 5432]
            end
            
            subgraph "Support Containers"
                CleanupCont[cleanup<br/>Container<br/>Profile: cleanup]
            end
            
            UpdateCont -->|gRPC| CrudCont
            QueryCont -->|gRPC| CrudCont
            
            CrudCont -->|MongoDB Wire| MongoCont
            CrudCont -->|Bolt| Neo4jCont
            CrudCont -->|PostgreSQL Wire| PostgresCont
            
            CleanupCont -.->|Cleanup| MongoCont
            CleanupCont -.->|Cleanup| Neo4jCont
            CleanupCont -.->|Cleanup| PostgresCont
        end
        
        subgraph "Docker Volumes"
            MongoVol[mongodb_data<br/>mongodb_config<br/>mongodb_backup]
            Neo4jVol[neo4j_data<br/>neo4j_logs<br/>neo4j_import]
            PostgresVol[postgres_data<br/>postgres_backup]
        end
        
        MongoCont -.->|Mount| MongoVol
        Neo4jCont -.->|Mount| Neo4jVol
        PostgresCont -.->|Mount| PostgresVol
    end
    
    Internet[Internet/Local Network] -->|HTTP 8080| UpdateCont
    Internet -->|HTTP 8081| QueryCont
    Internet -.->|Dev Access 27017| MongoCont
    Internet -.->|Dev Access 7474/7687| Neo4jCont
    Internet -.->|Dev Access 5432| PostgresCont
    
    style UpdateCont fill:#fff4e6
    style QueryCont fill:#fff4e6
    style CrudCont fill:#f3e5f5
    style MongoCont fill:#e8f5e9
    style Neo4jCont fill:#e8f5e9
    style PostgresCont fill:#e8f5e9
    style MongoVol fill:#ffebee
    style Neo4jVol fill:#ffebee
    style PostgresVol fill:#ffebee
```

---

## 8. Entity Lifecycle State Machine

```mermaid
stateDiagram-v2
    [*] --> Created: POST /entities
    
    Created --> Active: Entity exists in all DBs
    
    Active --> BeingRead: GET /entities/{id}
    BeingRead --> Active: Return data
    
    Active --> BeingUpdated: PUT /entities/{id}
    BeingUpdated --> Active: Update complete
    
    Active --> BeingDeleted: DELETE /entities/{id}
    BeingDeleted --> Deleted: Remove from all DBs
    
    Active --> Terminated: Set terminated timestamp
    Terminated --> Historical: Query with activeAt
    
    Deleted --> [*]
    
    note right of Created
        Metadata → MongoDB
        Entity → Neo4j
        Relationships → Neo4j
        Attributes → PostgreSQL
    end note
    
    note right of BeingRead
        Selective field retrieval
        based on output parameter:
        - metadata
        - relationships
        - attributes
    end note
    
    note right of Terminated
        Entity marked as terminated
        but remains in databases
        for historical queries
    end note
```

---

## 9. Backup and Restore Workflow

```mermaid
flowchart TD
    Start([Backup Initiated]) --> CheckDatabases{Check Database<br/>Status}
    
    CheckDatabases -->|All Healthy| Parallel[Parallel Backup]
    CheckDatabases -->|Unhealthy| Error1[Error: Database Unavailable]
    
    Parallel --> BackupMongo[Backup MongoDB<br/>mongodump → nexoan.tar.gz]
    Parallel --> BackupNeo4j[Backup Neo4j<br/>neo4j-admin dump → neo4j.dump]
    Parallel --> BackupPostgres[Backup PostgreSQL<br/>pg_dump → nexoan.tar.gz]
    
    BackupMongo --> LocalStore[Store in Local<br/>Backup Directory]
    BackupNeo4j --> LocalStore
    BackupPostgres --> LocalStore
    
    LocalStore --> CommitGit{Commit to<br/>data-backups repo?}
    
    CommitGit -->|Yes| GitCommit[Git commit & push]
    CommitGit -->|No| LocalOnly[Local backup only]
    
    GitCommit --> CreatePR[Create Pull Request]
    CreatePR --> PRMerge[Merge to main]
    PRMerge --> CreateRelease[Create GitHub Release<br/>Tag: version]
    
    CreateRelease --> Complete([Backup Complete])
    LocalOnly --> Complete
    
    Complete -.-> RestoreStart([Restore Initiated])
    
    RestoreStart --> RestoreSource{Restore Source?}
    
    RestoreSource -->|Local| LocalRestore[Read from local<br/>backup directory]
    RestoreSource -->|GitHub| DownloadRelease[Download from<br/>GitHub release]
    
    LocalRestore --> ExtractFiles[Extract backup files]
    DownloadRelease --> ExtractFiles
    
    ExtractFiles --> ParallelRestore[Parallel Restore]
    
    ParallelRestore --> RestoreMongo[Restore MongoDB<br/>mongorestore]
    ParallelRestore --> RestoreNeo4j[Restore Neo4j<br/>neo4j-admin load]
    ParallelRestore --> RestorePostgres[Restore PostgreSQL<br/>pg_restore]
    
    RestoreMongo --> Verify[Verify Data Integrity]
    RestoreNeo4j --> Verify
    RestorePostgres --> Verify
    
    Verify --> RestoreComplete([Restore Complete])
    
    Error1 --> End([End with Error])
    
    style Start fill:#e1f5ff
    style Complete fill:#c8e6c9
    style RestoreComplete fill:#c8e6c9
    style Error1 fill:#ffcdd2
    style CreateRelease fill:#fff9c4
```

---

## 10. Attribute Processing Pipeline

```mermaid
flowchart LR
    subgraph "Input"
        AttrInput[Attribute Data<br/>{<br/>  name: salary,<br/>  value: 100000,<br/>  startTime: 2024-01<br/>}]
    end
    
    subgraph "Type Inference"
        TypeCheck[Infer Data Type<br/>→ int]
    end
    
    subgraph "Storage Inference"
        StorageCheck[Infer Storage Type<br/>→ SCALAR]
    end
    
    subgraph "Schema Management"
        CheckSchema{Schema<br/>Exists?}
        CreateSchema[Create Schema<br/>in attribute_schemas]
        UseSchema[Use Existing Schema]
        
        CheckSchema -->|No| CreateSchema
        CheckSchema -->|Yes| UseSchema
    end
    
    subgraph "Table Management"
        CheckTable{Table<br/>Exists?}
        CreateTable[Create Table<br/>attr_Kind_AttrName]
        UseTable[Use Existing Table]
        
        CheckTable -->|No| CreateTable
        CheckTable -->|Yes| UseTable
    end
    
    subgraph "Data Insertion"
        InsertData[Insert Attribute Value<br/>with time range]
    end
    
    subgraph "Link Entity"
        LinkEntity[Link in<br/>entity_attributes table]
    end
    
    AttrInput --> TypeCheck
    TypeCheck --> StorageCheck
    StorageCheck --> CheckSchema
    
    CreateSchema --> CheckTable
    UseSchema --> CheckTable
    
    CreateTable --> InsertData
    UseTable --> InsertData
    
    InsertData --> LinkEntity
    
    LinkEntity --> Output[Attribute Stored]
    
    style AttrInput fill:#e1f5ff
    style TypeCheck fill:#fff9c4
    style StorageCheck fill:#fff9c4
    style CreateSchema fill:#ffecb3
    style CreateTable fill:#ffecb3
    style Output fill:#c8e6c9
```

---

## Diagram Usage

### Viewing Diagrams

These Mermaid diagrams can be viewed in:
1. **GitHub** - Native Mermaid rendering
2. **VS Code** - With Mermaid extension
3. **Mermaid Live Editor** - https://mermaid.live
4. **Documentation Sites** - Most support Mermaid rendering

### Exporting Diagrams

To export as PNG/SVG:
```bash
# Install Mermaid CLI
npm install -g @mermaid-js/mermaid-cli

# Export diagram
mmdc -i diagrams.md -o output.png
```

### Updating Diagrams

When updating the architecture:
1. Update the relevant Mermaid diagram in this file
2. Update the ASCII diagrams in `overview.md` if needed
3. Ensure consistency across all documentation
4. Commit changes with descriptive message

---

**Document Version**: 1.0  
**Last Updated**: October 2024  
**Format**: Mermaid Diagram Language

