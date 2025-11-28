# What is OpenGIN?

OpenGIN is a flexible, high-performance data platform designed to handle complex entity relationships and temporal data. It serves as a backend engine for applications requiring robust data modeling, ingestion, and retrieval capabilities.

## Core Concepts

- **Polyglot Persistence**: OpenGIN leverages the strengths of multiple databases:
    - **MongoDB**: Stores flexible metadata (JSON documents).
    - **Neo4j**: Manages entities and their relationships (Graph).
    - **PostgreSQL**: Handles time-series attribute data (Relational).

- **Entity Model**: The central unit of data is an "Entity".
    - **ID**: Unique identifier.
    - **Kind**: Classification (Major/Minor types).
    - **Name**: Temporal name values.
    - **Metadata**: Key-value pairs (stored in Mongo).
    - **Attributes**: Time-based values (stored in Postgres).
    - **Relationships**: Connections to other entities (stored in Neo4j).

- **Architecture**:
    - **Ingestion API (Ballerina)**: REST API for creating/updating entities. Converts JSON to Protobuf.
    - **Read API (Ballerina)**: REST API for retrieving entities.
    - **Core API (Go)**: The central brain. Handles business logic and orchestrates data storage/retrieval across the three databases via gRPC.

## Key Features

- **Temporal Data**: Native support for time-based values (startTime, endTime) for attributes and relationships.
- **Graph Capabilities**: Powerful relationship traversal and querying via Neo4j.
- **Scalability**: Microservices architecture allows independent scaling of components.
- **Strict Contracts**: Uses Protobuf for internal communication and OpenAPI for external REST APIs.

## Purpose

OpenGIN aims to provide a "Universal Data Engine" that abstracts the complexity of managing polyglot persistence and temporal graph data, offering a simple API for developers to build complex data-driven applications.
