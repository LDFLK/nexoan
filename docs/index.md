# OpenGIN Documentation 

Welcome to the **OpenGIN (Open General Information Network)** documentation. This is your central hub for understanding, implementing, and contributing to the OpenGIN platform.

## Quick Start

**New to OpenGIN?** Start here:

1. **[Architecture Overview](./architecture/overview.md)** - High-level system understanding
2. **[How It Works](./how_it_works.md)** - End-to-end data flow explanation  
3. **[Getting Started Guide](../README.md)** - Setup and installation

---

## Documentation Structure

### Architecture & Design

| Document | Description | Audience |
|----------|-------------|----------|
| **[Architecture Overview](./architecture/overview.md)** | Complete system architecture, data flows, and design decisions | Everyone |
| **[API Layer Details](./architecture/api-layer-details.md)** | Complete API documentation and contracts | API Consumers, Frontend Devs |
| **[Database Schemas](./architecture/database-schemas.md)** | MongoDB, Neo4j, PostgreSQL schemas | Database Admins, Backend Devs |

### Core Systems

| Document | Description | Audience |
|----------|-------------|----------|
| **[How It Works](./how_it_works.md)** | End-to-end data flow and processing | Developers, Architects |
| **[Data Types](./datatype.md)** | Type inference system and supported types | Developers |
| **[Storage Types](./storage.md)** | Storage inference and data organization | Backend Developers |

### Database & Storage

| Document | Description | Audience |
|----------|-------------|----------|
| **[MongoDB Backup](./database/BACKUP_MONGODB.md)** | MongoDB backup and restore procedures | Database Admins |
| **[Neo4j Backup](./database/BACKUP_NEO4J.md)** | Neo4j backup and restore procedures | Database Admins |
| **[PostgreSQL Backup](./database/BACKUP_POSTGRES.md)** | PostgreSQL backup and restore procedures | Database Admins |

### Deployment & Operations

| Document | Description | Audience |
|----------|-------------|----------|
| **[Release Lifecycle](./release_life_cycle.md)** | Versioning, release stages, and deployment | DevOps, Release Managers |
| **[Backup Integration](./deployment/BACKUP_INTEGRATION.md)** | Backup and restore workflows | Operations Team |

### Limitations

| Document | Description | Audience |
|----------|-------------|----------|
| **[Limitations](./limitations.md)** | Known limitations | All Users |

---

## Role-Based Navigation

### 👨‍💻 **I'm a Developer**

**Getting Started:**
- [Architecture Overview](./architecture/overview.md) → [How It Works](./how_it_works.md)

**Working on APIs:**
- [API Layer Details](./architecture/api-layer-details.md) → [Data Types](./datatype.md) → [Storage Types](./storage.md)

### 🏗️ **I'm an Architect**

**System Design:**
- [Architecture Overview](./architecture/overview.md) → [Architecture Diagrams](./architecture/diagrams.md) → [Database Schemas](./architecture/database-schemas.md)

**Understanding Data Flow:**
- [How It Works](./how_it_works.md) → [Architecture Overview](./architecture/overview.md) → [Architecture Diagrams](./architecture/diagrams.md)

### 🗄️ **I'm a Database Administrator**

**Database Management:**
- [Database Schemas](./architecture/database-schemas.md) → [MongoDB Backup](./database/BACKUP_MONGODB.md) → [Neo4j Backup](./database/BACKUP_NEO4J.md) → [PostgreSQL Backup](./database/BACKUP_POSTGRES.md)

**Backup & Recovery:**
- [Backup Integration](./deployment/BACKUP_INTEGRATION.md) → [MongoDB Backup](./database/BACKUP_MONGODB.md) → [Neo4j Backup](./database/BACKUP_NEO4J.md) → [PostgreSQL Backup](./database/BACKUP_POSTGRES.md)

### 🚀 **I'm a DevOps Engineer**

**Deployment & Operations:**
- [Release Lifecycle](./release_life_cycle.md) → [Backup Integration](./deployment/BACKUP_INTEGRATION.md) → [Architecture Overview](./architecture/overview.md)

### 👥 **I'm a Product Manager**

**Understanding the Platform:**
- [Architecture Overview](./architecture/overview.md) → [How It Works](./how_it_works.md) → [UX Guidelines](./ux.md)

---

## Task-Based Navigation

### **Understanding the System**
- [Architecture Overview](./architecture/overview.md) + [Architecture Diagrams](./architecture/diagrams.md) + [How It Works](./how_it_works.md)

### **Adding New Features**
- [API Layer Details](./architecture/api-layer-details.md) + [Database Schemas](./architecture/database-schemas.md)

### **Debugging Issues**
- [How It Works](./how_it_works.md) + [Database Schemas](./architecture/database-schemas.md) + [Limitations](./limitations.md)

### **Performance Optimization**
- [Database Schemas](./architecture/database-schemas.md) + [Architecture Overview](./architecture/overview.md)

### **Data Migration**
- [Database Schemas](./architecture/database-schemas.md) + [Backup Integration](./deployment/BACKUP_INTEGRATION.md)

### **API Integration**
- [API Layer Details](./architecture/api-layer-details.md) + [Data Types](./datatype.md) + [Storage Types](./storage.md)

---

## Architecture at a Glance

### Polyglot Database Strategy
OpenGIN uses three specialized databases:

| Database | Purpose | Use Case |
|----------|---------|----------|
| **MongoDB** | Metadata | Schema-less, flexible key-value storage |
| **Neo4j** | Entities & Relationships | Graph traversal, relationship queries |
| **PostgreSQL** | Attributes | ACID compliance, time-series data, strong typing |

### Layered Architecture
```
┌─────────────────┐
│   Client Layer  │ (HTTP/JSON)
└─────────┬───────┘
          │
┌─────────▼───────┐
│     API Layer   │ (Ingestion API, Read API)
└─────────┬───────┘
          │ gRPC/Protobuf
┌─────────▼───────┐
│     Core API    │ (Orchestration)
└─────────┬───────┘
          │ Native Protocols
┌─────────▼───────────────────┐
│ MongoDB │ Neo4j │ PostgreSQL│
│ Metadata│ Graph │ Attributes│
└─────────────────────────────┘
```

### Key Features
- **Temporal Support**: All data versioned by time with `startTime`/`endTime`
- **Type Inference**: Automatic data type detection (int, float, string, bool, date, time, datetime)
- **Storage Inference**: Automatic storage strategy determination (SCALAR, LIST, MAP, TABULAR, GRAPH)
- **Polyglot Database**: Each database optimized for specific data types
- **Contract-First**: OpenAPI specifications with Swagger UI

---

## Technology Stack

| Layer | Technology | Language |
|-------|-----------|----------|
| **Ingestion API** | Ballerina | Ballerina |
| **Read API** | Ballerina | Ballerina |
| **Core API** | Go + gRPC | Go |
| **MongoDB** | MongoDB 5.0+ | - |
| **Neo4j** | Neo4j 5.x | Cypher |
| **PostgreSQL** | PostgreSQL 14+ | SQL |
| **Messaging** | Protobuf | IDL |
| **Container** | Docker + Compose | YAML |

---

## Quick Development Setup

### Prerequisites
- Docker and Docker Compose
- Go 1.19+ (for CRUD service)
- Ballerina (for APIs)

### Start the System
```bash
# Start databases
docker-compose up -d mongodb neo4j postgres

# Start CRUD service
cd opengin/core-api && ./core-service

# Start APIs
# Ingestion API: http://localhost:8080
# Read API: http://localhost:8081
```

### Test the System
```bash
# Run E2E tests
cd opengin/tests/e2e && ./run_e2e.sh

# Run performance tests
cd perf && python performance_test.py
```

---

## Common Patterns

### Entity Creation Flow
```
Client → Ingestion API → Core API → [MongoDB, Neo4j, PostgreSQL] → Response
```

### Entity Query Flow
```
Client → Read API → Core API → Fetch from DBs based on output param → Response
```

### Selective Retrieval
```
GET /v1/entities/{id}?output=metadata,relationships
```

### Temporal Query
```
GET /v1/entities/{id}/attributes?name=salary&activeAt=2024-03-15T00:00:00Z
```

---

## Development Workflow

### 1. **Understanding Changes**
- Review [Architecture Overview](./architecture/overview.md)
- Check [How It Works](./how_it_works.md) for data flow
- Consult [Database Schemas](./architecture/database-schemas.md) for data structure

### 2. **Making Changes**
- **API Changes**: Update OpenAPI contracts in `opengin/contracts/rest/`
- **Service Changes**: Modify CRUD service in `opengin/crud-api/`
- **Database Changes**: Consider impact across all three databases

### 3. **Testing**
- Unit tests: `go test ./...` or `bal test`
- Integration tests: E2E tests in `opengin/tests/e2e/`
- Performance tests: `perf/performance_test.py`

### 4. **Documentation**
- Update relevant architecture docs
- Keep diagrams in sync with changes
- Add examples to appropriate sections

---

## Support & Contributing

### Getting Help
1. Review this documentation first
2. Check [Limitations](./limitations.md) for known problems
3. Review service-specific READMEs in `opengin/` directories
4. Consult the development team

### Contributing
1. Follow the development workflow above
2. Update documentation when making changes
3. Keep diagrams in sync with code changes
4. Test thoroughly before submitting changes

---

## Documentation Status

| Section | Status | Last Updated |
|---------|--------|--------------|
| Architecture | ✅ Complete | October 2024 |
| APIs | ✅ Complete | October 2024 |
| Databases | ✅ Complete | October 2024 |
| Deployment | ✅ Complete | October 2024 |
| Troubleshooting | ✅ Complete | October 2024 |

---

**Last Updated**: October 2024  
**Documentation Status**: ✅ Complete and Current  
**Maintained By**: OpenGIN Development Team

