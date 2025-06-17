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
  -e CRUD_SERVICE_HOST="$CRUD_SERVICE_HOST" \
  -e CRUD_SERVICE_PORT="$CRUD_SERVICE_PORT" \
  ldf-choreo-crud-service
```

2. Start the Update Service:
```bash
# Build the update service image
docker build -t ldf-choreo-update-service -f Dockerfile.update.choreo .

# Run the update service using environment variables
docker run -d -p 8080:8080 \
  --name ldf-choreo-update-service \
  -e CRUD_SERVICE_URL="http://host.docker.internal:$CRUD_SERVICE_PORT" \
  -e UPDATE_SERVICE_HOST="$UPDATE_SERVICE_HOST" \
  -e UPDATE_SERVICE_PORT="$UPDATE_SERVICE_PORT" \
  ldf-choreo-update-service
```

3. Start the Query Service:
```bash
# Build the update service image
docker build -t ldf-choreo-query-service -f Dockerfile.query.choreo .

# Run the update service using environment variables
docker run -d -p 8081:8081 \
  --name ldf-choreo-query-service \
  -e CRUD_SERVICE_URL="http://host.docker.internal:$CRUD_SERVICE_PORT" \
  -e QUERY_SERVICE_HOST="$QUERY_SERVICE_HOST" \
  -e QUERY_SERVICE_PORT="$QUERY_SERVICE_PORT" \
  ldf-choreo-query-service
```

### Required Environment Variables

#### CRUD Service
- `NEO4J_URI`: Connection URI for Neo4j database
- `NEO4J_USER`: Username for Neo4j authentication
- `NEO4J_PASSWORD`: Password for Neo4j authentication
- `MONGO_URI`: Connection URI for MongoDB
- `MONGO_DB_NAME`: MongoDB database name
- `MONGO_COLLECTION`: MongoDB collection name
- `MONGO_ADMIN_USER`: MongoDB admin username
- `MONGO_ADMIN_PASSWORD`: MongoDB admin password
- `CRUD_SERVICE_HOST`: Host address to bind the service (default: 0.0.0.0)
- `CRUD_SERVICE_PORT`: Port to expose the gRPC service (default: 50051)

#### Update Service
- `CRUD_SERVICE_URL`: URL of the CRUD service (required)
- `UPDATE_SERVICE_HOST`: Host address to bind the service (default: 0.0.0.0)
- `UPDATE_SERVICE_PORT`: Port to expose the service (default: 8080)

### Testing the Services

1. Verify CRUD Service:
```bash
# Test CRUD service health
curl http://localhost:$CRUD_SERVICE_PORT/health
```

2. Test Update Service:
```bash
# Create an entity
curl -X POST http://localhost:$UPDATE_SERVICE_PORT/entities \
  -H "Content-Type: application/json" \
  -d '{"id":"123","kind":{"major":"example","minor":"test"}}'
```

3. Test Query Service:


### Troubleshooting

1. If services can't connect:
   - Ensure both containers are running (`docker ps`)
   - Check container logs (`docker logs <container-id>`)
   - Verify ports are not in use
   - On macOS, ensure `host.docker.internal` is properly configured
   - Verify MongoDB and Neo4j connections for CRUD service

2. Common Issues:
   - Port conflicts: Change port mappings if needed
   - Connection refused: Check if services are running and accessible
   - Database connection issues: Verify MongoDB and Neo4j credentials
   - Permission issues: Ensure proper file permissions in containers


## Choreo Deployment 

When we want to work on a feature which is branched out from the main as a
feature-x, but we keep working on that feature by locally checking out and
making consistent pull requests to that branch. We should adopt a custom for that
which doesn't confuses the branches. 

For instance for our Choreo development make sure to do a code freeze from the `main` 
branch and then make a `choreo-rc-<version>` at `upstream/nexoan` or we refer as `ldf/nexoan` repo. 

From that branch checkout a local branch

```bash
git checkout -b ldf-choreo-rc-<version> ldf/choreo-rc-<version>
```

Then create your branch to work on things. You can always use this `ldf-choreo-rc-<version>`
branch to keep things in sync with the `ldf/choreo-rc-<version>` and branch out from 
it to work on the next feature. This is useful when you're working on making significant changes
to the build configurations at deployment. 

Now checkout your branch as follows. 

```bash
git checkout -b vibhatha-choreo-rc-0.1.0
```

> ⚠️ **Warning:** When you create a pull request, make sure to target the `choreo-rc-<version>` branch instead of the `main` branch.

## Deployment

### CRUD Service

### Ingestion Service

### Query Service

## References

1. https://wso2.com/choreo/docs/develop-components/develop-components-with-git/
2. https://wso2.com/choreo/docs/devops-and-ci-cd/manage-configurations-and-secrets/
3. https://wso2.com/choreo/docs/choreo-concepts/ci-cd/
4. [Test Runner](https://wso2.com/choreo/docs/testing/test-components-with-test-runner/)
5. [Choreo Examples](https://github.com/wso2/choreo-samples)
6. [Expose TCP Server via a Service](https://wso2.com/choreo/docs/develop-components/develop-services/expose-a-tcp-server-via-a-service/#step-2-build-and-deploy)
