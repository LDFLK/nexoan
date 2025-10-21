# OpenGIN Architecture Documentation

Welcome to the OpenGIN architecture documentation. This folder contains comprehensive documentation about the system's architecture, components, data flows, and design decisions.

---

## Documentation Structure

### 1. [Overview](./overview.md) - **START HERE**
Complete system architecture overview including:
- High-level architecture diagram
- Layer-by-layer breakdown (API, Service, Database)
- Data model and storage strategy
- Data flow sequences
- Type and storage inference systems
- Technology stack
- Key features and design decisions

**Recommended for**: Everyone - developers, architects, stakeholders

---

### 2. [Diagrams](./diagrams.md)
Visual architecture diagrams in Mermaid format:
- System architecture (high-level)
- Create entity data flow
- Read entity data flow
- Component architecture
- Data storage distribution
- Type inference flow
- Deployment architecture
- Entity lifecycle state machine
- Backup and restore workflow
- Attribute processing pipeline

**Recommended for**: Visual learners, architects, presentations

---

### 3. [CRUD Service Details](./crud-service-details.md)
In-depth documentation of the CRUD Service:
- Directory structure
- gRPC server implementation
- All service methods (CreateEntity, ReadEntity, UpdateEntity, DeleteEntity)
- Repository layer (MongoDB, Neo4j, PostgreSQL)
- Engine layer (AttributeProcessor, TypeInference, StorageInference)
- Configuration and environment variables
- Testing strategies
- Performance considerations
- Error handling and logging

**Recommended for**: Backend developers, service implementers

---

### 4. [API Layer Details](./api-layer-details.md)
Complete API layer documentation:
- Update API (CREATE, UPDATE, DELETE operations)
- Query API (READ, QUERY operations)
- Request/response formats
- JSON to Protobuf conversion
- OpenAPI contracts
- Swagger UI documentation
- Query parameters and filtering
- Temporal queries (activeAt)
- Error handling
- Best practices

**Recommended for**: API consumers, frontend developers, integration engineers

---

### 5. [Database Schemas](./database-schemas.md)
Detailed database schema documentation:
- **MongoDB**: Collections, document structures, indexes
- **Neo4j**: Node types, relationship types, Cypher queries
- **PostgreSQL**: Core tables, dynamic attribute tables, type mapping
- Cross-database consistency
- Schema evolution strategies
- Backup and restore procedures
- Performance optimization

**Recommended for**: Database administrators, backend developers, data engineers

---

## Quick Navigation

### By Role

**I'm a new developer** → Start with [Overview](./overview.md), then [Diagrams](./diagrams.md)

**I'm working on APIs** → Read [API Layer Details](./api-layer-details.md)

**I'm working on backend** → Read [CRUD Service Details](./crud-service-details.md)

**I'm working on databases** → Read [Database Schemas](./database-schemas.md)

**I'm presenting the architecture** → Use [Diagrams](./diagrams.md) and [Overview](./overview.md)

### By Task

**Understanding data flow** → [Overview](./overview.md) + [Diagrams](./diagrams.md)

**Adding new endpoint** → [API Layer Details](./api-layer-details.md)

**Adding new entity type** → [Database Schemas](./database-schemas.md) + [CRUD Service Details](./crud-service-details.md)

**Debugging data storage** → [Database Schemas](./database-schemas.md)

**Performance tuning** → [CRUD Service Details](./crud-service-details.md) + [Database Schemas](./database-schemas.md)

**Understanding types** → [Overview](./overview.md) + [CRUD Service Details](./crud-service-details.md)

---

## Key Concepts

### Multi-Database Strategy

OpenGIN uses three databases, each optimized for specific data types:

| Database | Purpose | Reason |
|----------|---------|--------|
| **MongoDB** | Metadata | Schema-less, flexible key-value storage |
| **Neo4j** | Entities & Relationships | Optimized graph traversal, Cypher queries |
| **PostgreSQL** | Attributes | ACID compliance, time-series data, strong typing |

### Layered Architecture

```
Client Layer (HTTP/JSON)
    ↓
API Layer (Update API, Query API)
    ↓  gRPC/Protobuf
Service Layer (CRUD Service)
    ↓  Native Protocols
Database Layer (MongoDB, Neo4j, PostgreSQL)
```

### Time-Based Data

All attributes and relationships support temporal tracking:
- `startTime`: When value/relationship became effective
- `endTime`: When value/relationship ended (NULL = current)
- `activeAt` queries: Retrieve data as it existed at specific point in time

### Type and Storage Inference

The system automatically:
1. **Infers data types**: int, float, string, bool, date, time, datetime
2. **Determines storage strategy**: SCALAR, LIST, MAP, TABULAR, GRAPH
3. **Creates schemas dynamically**: No manual schema definition needed

---

## Architecture Principles

### 1. Separation of Concerns
- APIs handle HTTP/JSON ↔ gRPC/Protobuf conversion
- CRUD Service orchestrates business logic
- Repositories handle database-specific operations

### 2. Database Specialization
- Each database does what it does best
- No single database bottleneck
- Independent scaling of each database

### 3. Contract-First API Design
- OpenAPI specifications define APIs
- Code generated from contracts
- Swagger UI for documentation

### 4. Type Safety
- Protobuf for internal communication
- Strong typing in Go (CRUD Service)
- Type inference for flexibility

### 5. Temporal Support
- All data versioned by time
- Historical queries supported
- Audit trail built-in

---

## Common Patterns

### Entity Creation Flow
```
Client → Update API → CRUD Service → [MongoDB, Neo4j, PostgreSQL] → Response
```

### Entity Query Flow
```
Client → Query API → CRUD Service → Fetch from DBs based on output param → Response
```

### Selective Retrieval
```
GET /v1/entities/{id}?output=metadata,relationships
```
Only fetches requested fields, reducing load and improving performance.

### Temporal Query
```
GET /v1/entities/{id}/attributes?name=salary&activeAt=2024-03-15T00:00:00Z
```
Returns attribute value as it was on specific date.

---

## Technology Stack Summary

| Layer | Technology | Language |
|-------|-----------|----------|
| Update API | Ballerina | Ballerina |
| Query API | Ballerina | Ballerina |
| CRUD Service | Go + gRPC | Go |
| MongoDB | MongoDB 5.0+ | - |
| Neo4j | Neo4j 5.x | Cypher |
| PostgreSQL | PostgreSQL 14+ | SQL |
| Messaging | Protobuf | IDL |
| Container | Docker + Compose | YAML |
| Testing | Various | Go, Ballerina, Python |

---

## Architecture Diagrams at a Glance

### System Architecture
```
┌─────────────┐
│   Clients   │
└──────┬──────┘
       │ HTTP/JSON
┌──────┴──────────────┐
│    API Layer        │
│ Update API | Query  │
└──────┬──────────────┘
       │ gRPC/Protobuf
┌──────┴──────────────┐
│   CRUD Service      │
│  (Orchestration)    │
└──────┬──────────────┘
       │ Native Protocols
┌──────┴──────────────────────────┐
│ MongoDB | Neo4j | PostgreSQL    │
│ Metadata| Graph | Attributes     │
└─────────────────────────────────┘
```

See [Diagrams](./diagrams.md) for detailed visual representations.

---

## Development Workflow

### 1. Understanding the System
- Read [Overview](./overview.md)
- Review [Diagrams](./diagrams.md)
- Understand data flow

### 2. Setting Up Development Environment
- Clone repository
- Start databases: `docker-compose up -d mongodb neo4j postgres`
- Start CRUD service: `cd nexoan/crud-api && ./crud-service`
- Start APIs: Update API (port 8080), Query API (port 8081)

### 3. Making Changes

**Adding new API endpoint**:
1. Update OpenAPI contract in `nexoan/contracts/rest/`
2. Regenerate service code
3. Implement endpoint logic
4. Update [API Layer Details](./api-layer-details.md)

**Adding CRUD Service feature**:
1. Implement in appropriate layer (server, engine, repository)
2. Add tests
3. Update [CRUD Service Details](./crud-service-details.md)

**Modifying database schema**:
1. Consider impact across all databases
2. Update schema migration scripts
3. Update [Database Schemas](./database-schemas.md)

### 4. Testing
- Unit tests: `go test ./...` or `bal test`
- Integration tests: E2E tests in `nexoan/tests/e2e/`
- Database tests: Ensure all databases are running

### 5. Documentation
- Update relevant architecture docs
- Add examples to appropriate sections
- Keep diagrams in sync with changes

---

## Performance Considerations

### API Layer
- Use selective retrieval (`output` parameter)
- Filter at the source (server-side filtering)
- Use temporal queries to reduce data transfer

### Service Layer
- Connection pooling for all databases
- Parallel operations where possible
- Efficient Protobuf serialization

### Database Layer
- Proper indexing on all tables/collections
- Cypher query optimization
- PostgreSQL query planning

See individual docs for detailed optimization strategies.

---

## Security Considerations

### Current State
- Development mode: No authentication
- All endpoints publicly accessible
- No encryption in transit (within Docker network)

### Planned Enhancements
- JWT authentication
- Role-based access control (RBAC)
- TLS/SSL for external communication
- API rate limiting
- Field-level access control

---

## Monitoring and Observability

### Logging
- Structured logging in all services
- Log aggregation (planned)
- Error tracking

### Metrics (Planned)
- Request rates
- Response times
- Database query performance
- Error rates

### Tracing (Planned)
- Distributed tracing with OpenTelemetry
- End-to-end request tracking

---

## Troubleshooting

### Common Issues

**Entity not found**:
- Check if entity exists in Neo4j
- Verify entity ID is correct
- Check if entity was deleted

**Attribute not saving**:
- Check type inference logs
- Verify PostgreSQL connection
- Check if table was created

**Relationship not showing**:
- Verify both entities exist in Neo4j
- Check relationship direction
- Use temporal query to check if relationship was active

**Metadata missing**:
- Check MongoDB connection
- Verify entity ID matches
- Check if metadata was provided in create request

---

## Contributing to Documentation

### When to Update Documentation

Update architecture docs when:
- Adding new services or components
- Changing data flow or structure
- Modifying database schemas
- Adding new features
- Changing APIs or contracts

### Documentation Style

- Clear, concise language
- Include code examples
- Add diagrams where helpful
- Cross-reference related docs
- Keep examples up to date

### Diagram Updates

When updating Mermaid diagrams:
1. Test rendering at https://mermaid.live
2. Ensure consistency with other diagrams
3. Update ASCII diagrams in overview if needed
4. Commit with descriptive message

---

## Related Documentation

### In This Repository

- [Main README](../../README.md) - Project overview and quick start
- [How It Works](../how_it_works.md) - Detailed data flow
- [Data Types](../datatype.md) - Type inference system
- [Storage Types](../storage.md) - Storage inference system
- [Deployment Guide](../deployment/BACKUP_INTEGRATION.md) - Backup and restore
- [Core API README](../../nexoan/crud-api/README.md) - Polyglot Database Query Processing
- [Ingestion API README](../../nexoan/update-api/README.md) - Ingestion API setup
- [Read API README](../../nexoan/query-api/README.md) - Read API setup

### External Resources

- [Ballerina Documentation](https://ballerina.io/learn/)
- [gRPC Documentation](https://grpc.io/docs/)
- [Protocol Buffers](https://developers.google.com/protocol-buffers)
- [MongoDB Documentation](https://docs.mongodb.com/)
- [Neo4j Documentation](https://neo4j.com/docs/)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [Docker Documentation](https://docs.docker.com/)
- [Mermaid Documentation](https://mermaid.js.org/)

---

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 1.0.0 - alpha | October 2024 | Initial comprehensive architecture documentation |

---

## Contact and Support

For questions about the architecture:
- Review this documentation first
- Check related service READMEs
- Review code comments and tests
- Consult the development team

---

**Last Updated**: October 2024  
**Documentation Status**: ✅ Complete and Current  
**Maintained By**: OpenGIN Development Team

