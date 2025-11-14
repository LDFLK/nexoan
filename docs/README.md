# Nexoan Documentation

Welcome to the Nexoan data management system documentation. This folder contains comprehensive guides for understanding, deploying, and maintaining the Nexoan system.

## 📚 Documentation Overview

### Core System Documentation

- **[How It Works](how_it_works.md)** - Complete end-to-end data flow from JSON input to database storage
- **[Architecture](architecture.md)** - System design and component overview
- **[Data Types](datatype.md)** - Type inference system and supported data types
- **[Storage Types](storage.md)** - Detailed guide to supported data storage formats

### Database Management

#### Backup Procedures
- **[MongoDB Backup](database/BACKUP_MONGODB.md)** - MongoDB backup and restore procedures
- **[Neo4j Backup](database/BACKUP_NEO4J.md)** - Neo4j backup and restore procedures  
- **[PostgreSQL Backup](database/BACKUP_POSTGRES.md)** - PostgreSQL backup and restore procedures

#### Deployment
- **[Backup Integration](deployment/BACKUP_INTEGRATION.md)** - Instructions on running the script for backing up and restoring all databases from github.

### Development & Operations

- **[UX Guidelines](ux.md)** - User experience best practices and guidelines
- **[Known Issues](issues.md)** - Current limitations, bugs, and workarounds
- **[Release Lifecycle](release_life_cycle.md)** - Development and release process

## 🚀 Quick Start

1. **Understanding the System**: Start with [How It Works](how_it_works.md) to understand the complete data flow
2. **Architecture Overview**: Read [Architecture](architecture.md) to understand system components
3. **Data Types**: Review [Data Types](datatype.md) to understand type inference
4. **Storage Formats**: Check [Storage Types](storage.md) for supported data formats

## 🔧 System Features

- **Multi-Database Support**: MongoDB, Neo4j, and PostgreSQL
- **Automatic Type Inference**: Intelligently determines data types from JSON
- **Flexible Storage**: Supports tabular, graph, list, map, and scalar data
- **Relationship Management**: Handles complex entity relationships
- **Data Consistency**: Maintains integrity across multiple storage systems

## 📋 Supported Storage Types

| Type | Description | Use Case |
|------|-------------|----------|
| **Tabular** | Structured data in table format | CSV-like data, relational data |
| **Graph** | Network of nodes and relationships | Social networks, hierarchies |
| **List** | Ordered collections of items | Arrays, sequences |
| **Map** | Key-value pair collections | Configuration, metadata |
| **Scalar** | Single values | Simple data points |

## 🗄️ Database Support

| Database | Purpose | Documentation |
|----------|---------|---------------|
| **MongoDB** | Metadata storage | [Backup Guide](database/BACKUP_MONGODB.md) |
| **Neo4j** | Graph entities and relationships | [Backup Guide](database/BACKUP_NEO4J.md) |
| **PostgreSQL** | Structured data | [Backup Guide](database/BACKUP_POSTGRES.md) |

## 🛠️ Development

### Prerequisites
- Docker and Docker Compose
- Go 1.19+
- Ballerina
- Access to MongoDB, Neo4j, and PostgreSQL instances

### Getting Started
1. Clone this repository
2. Review the [Architecture](architecture.md) documentation
3. Set up your databases following the backup guides
4. Deploy using the provided Docker configurations

## 📞 Support

- **Known Issues**: Check [Known Issues](issues.md) for common problems
- **Architecture Questions**: Review [Architecture](architecture.md)
- **Data Format Questions**: See [Storage Types](storage.md)

---

*This documentation is maintained alongside the codebase. For the most up-to-date information, always refer to the latest version in this repository.*
