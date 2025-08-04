# Choreo 

## Local Development and Testing

### Prerequisites

- Docker installed
- Git repository cloned
- Ports 50051 (CRUD service) and 8080 (Update service) available
- MongoDB and Neo4j instances (for CRUD service)

### Environment Setup

Set up your environment variables in your terminal:

```bash
# MongoDB Configuration
export MONGO_URI="mongodb+srv://username:password@your-cluster.mongodb.net/?retryWrites=true&w=majority"
export MONGO_DB_NAME="your-mongo-db-name"
export MONGO_COLLECTION="your-mongo-collection-name"
export MONGO_ADMIN_USER="your-mongo-admin-username"
export MONGO_ADMIN_PASSWORD="your-mongo-admin-password"

# Neo4j Configuration
export NEO4J_URI="neo4j+s://your-neo4j-instance.databases.neo4j.io"
export NEO4J_USER="your-neo4j-username"
export NEO4J_PASSWORD="your-neo4j-password"

# Service Configuration
export CRUD_SERVICE_HOST="0.0.0.0"
export CRUD_SERVICE_PORT="50051"
export UPDATE_SERVICE_HOST="0.0.0.0"
export UPDATE_SERVICE_PORT="8080"
export QUERY_SERVICE_HOST="0.0.0.0"
export QUERY_SERVICE_PORT="8081"
```

### Running Services Locally

1. Start the CRUD Service:
```bash
# Build the CRUD service image
# For ARM64 (Apple Silicon):
docker build --platform linux/arm64 -t ldf-choreo-crud-service -f Dockerfile.crud.choreo .
# For AMD64:
docker build --platform linux/amd64 -t ldf-choreo-crud-service -f Dockerfile.crud.choreo .

# Run the CRUD service using environment variables
docker run -d \
  --name ldf-choreo-crud-service \
  -p 50051:50051 \
  -e NEO4J_URI="$NEO4J_URI" \
  -e NEO4J_USER="$NEO4J_USER" \
  -e NEO4J_PASSWORD="$NEO4J_PASSWORD" \
  -e MONGO_URI="$MONGO_URI" \
  -e MONGO_DB_NAME="$MONGO_DB_NAME" \
  -e MONGO_COLLECTION="$MONGO_COLLECTION" \
  -e MONGO_ADMIN_USER="$MONGO_ADMIN_USER" \
  -e MONGO_ADMIN_PASSWORD="$MONGO_ADMIN_PASSWORD" \
  -e POSTGRES_HOST="$POSTGRES_HOST" \
  -e POSTGRES_PORT="$POSTGRES_PORT" \
  -e POSTGRES_USER="$POSTGRES_USER" \
  -e POSTGRES_PASSWORD="$POSTGRES_PASSWORD" \
  -e POSTGRES_DB="$POSTGRES_DB" \
  -e POSTGRES_SSL_MODE="$POSTGRES_SSL_MODE" \
  -e POSTGRES_TEST_DB_URI="$POSTGRES_TEST_DB_URI" \
  -e CRUD_SERVICE_HOST="$CRUD_SERVICE_HOST" \
  -e CRUD_SERVICE_PORT="$CRUD_SERVICE_PORT" \
  ldf-choreo-crud-service
```

## References

1. https://wso2.com/choreo/docs/develop-components/develop-components-with-git/
2. https://wso2.com/choreo/docs/devops-and-ci-cd/manage-configurations-and-secrets/
3. https://wso2.com/choreo/docs/choreo-concepts/ci-cd/
4. [Test Runner](https://wso2.com/choreo/docs/testing/test-components-with-test-runner/)
5. [Choreo Examples](https://github.com/wso2/choreo-samples)
6. [Expose TCP Server via a Service](https://wso2.com/choreo/docs/develop-components/develop-services/expose-a-tcp-server-via-a-service/#step-2-build-and-deploy)