## Development

For the development mode make sure you `source` the file containing the secrets. For instance 
you can keep a secret file like `ldf.testing.profile`

```bash
export MONGO_TESTING_DB_URI=""
export MONGO_TESTING_DB=""
export MONGO_TESTING_COLLECTION=""

export NEO4J_TESTING_DB_URI=""
export NEO4J_TESTING_USERNAME=""
export NEO4J_TESTING_PASSWORD=""

export POSTGRES_TESTING_HOST=""
export POSTGRES_TESTING_PORT=""
export POSTGRES_TESTING_USER=""
export POSTGRES_TESTING_PASSWORD=""
export POSTGRES_TESTING_DB=""
```

`config.env` or secrets in Github would make up `NEO4J_AUTH=${NEO4J_TESTING_USERNAME}/${NEO4J_TESTING_PASSWORD}`.

In the same terminal or ssh session, do the following;

This will start instances of the MongoDB, Neo4j, and PostgreSQL database servers.

### Start the Database Servers

```bash
docker compose up --build
```

- MongoDB can be accessed at `mongodb://localhost:27017`
- Neo4j can be accessed at `http://localhost:7474/browser/` for the web interface or `bolt://localhost:7687` for the bolt protocol
- PostgreSQL can be accessed at `localhost:5432`

### Shutdown the Database Servers

```bash
docker compose down -v
```

### BackUp Server Data (TODO)


### Restore Server Data (TODO)


### Docker Compose

Use the `docker compose` to up the services to run tests and to check the current version of the software is working. 



#### Up the Services

`docker compose up` 

#### Down the Services

`docker compose down` 

#### Get services up independently 

MongoDB Service

`docker compose up -d mongodb`

Neo4j Service 

`docker compose up -d neo4j`

PostgreSQL Service

`docker compose up -d postgres`

Build CRUD Service

`docker compose build crud` 

And to up it `docker compose up crud`

### Docker Health Checks and Service Startup Timing

The Docker Compose configuration includes health checks for all services to ensure proper startup sequencing and service readiness. Understanding these health checks is crucial for development and troubleshooting.

#### Health Check Configuration

Each service has specific health check settings optimized for its startup characteristics:

**CRUD Service (gRPC):**
```yaml
healthcheck:
  test: ["CMD", "nc", "-zv", "localhost", "50051"]
  interval: 15s      # Check every 15 seconds
  timeout: 10s       # Each check times out after 10 seconds
  retries: 5         # Allow 5 consecutive failures before marking unhealthy
  start_period: 120s # Grace period of 2 minutes for Go tests to complete
```

**Update/Query Services (HTTP):**
```yaml
healthcheck:
  test: ["CMD", "nc", "-zv", "localhost", "8080"]  # or 8081 for query
  interval: 15s      # Check every 15 seconds
  timeout: 10s       # Each check times out after 10 seconds
  retries: 5         # Allow 5 consecutive failures before marking unhealthy
  start_period: 120s # Grace period of 2 minutes for Ballerina compilation and tests
```

#### Service Startup Timeline

Understanding the startup sequence is important for debugging:

1. **Database Services (MongoDB, Neo4j, PostgreSQL)**: Start first and become healthy within 30-60 seconds
2. **CRUD Service**: 
   - Starts after databases are healthy
   - Runs Go tests (can take 60-90 seconds)
   - Starts gRPC server on port 50051
   - Health checks ignored for first 120 seconds
3. **Update/Query Services**:
   - Start after CRUD service is healthy
   - Compile Ballerina code and run tests
   - Start HTTP servers on ports 8080/8081
   - Health checks ignored for first 120 seconds

#### Monitoring Service Health

**Check overall service status:**
```bash
docker compose ps
```

**Monitor health status in real-time:**
```bash
watch -n 2 'docker compose ps --format "table {{.Name}}\t{{.Status}}\t{{.Health}}"'
```

**Check specific service health:**
```bash
# Check CRUD service health
docker inspect crud --format='{{.State.Health.Status}}'

# Get detailed health check history
docker inspect crud | jq '.[0].State.Health.Log[-3:]'
```

**Monitor service logs during startup:**
```bash
# Watch for CRUD service startup
docker logs -f crud | grep -E "(test|CRUD Service is running)"

# Watch for Update service startup
docker logs -f update | grep -E "(Compiling|Running executable)"
```

#### Troubleshooting Health Check Issues

**Common Issues:**

1. **Service marked as unhealthy but logs show it's running:**
   - Check if the health check command is appropriate for the service type
   - gRPC services need port checks (`nc -z`), not HTTP checks (`wget`)
   - Verify the service is listening on the expected port

2. **Service fails health check during startup:**
   - Increase `start_period` if services need more time for tests/compilation
   - Check if required dependencies (netcat, curl) are installed in the container

3. **Intermittent health check failures:**
   - Increase `retries` to allow for temporary network issues
   - Adjust `interval` and `timeout` based on service response characteristics

**Debug Commands:**
```bash
# Test health check manually
docker exec crud nc -zv localhost 50051

# Check if service is listening on expected port
docker exec crud netstat -tlnp | grep 50051

# Verify container has required tools
docker exec crud which nc
```

#### Health Check Best Practices

- **Use appropriate health checks**: Port checks for gRPC, HTTP checks for REST APIs
- **Set realistic timeouts**: Account for service initialization time (tests, compilation)
- **Monitor startup logs**: Watch for specific startup messages to verify service readiness
- **Test health checks manually**: Use `docker exec` to run health check commands directly
- **Consider service dependencies**: Ensure dependent services are healthy before starting

### Using CRUD API services via Ballerina

When using any CRUD services such as `ReadEntity`, `UpdateEntity` etc via Ballerina (for example in the query api or update api layer) pay special attention to the name field in Entity objects.

The name field is a TimeBasedValue of the following structure:

```protobuf
message TimeBasedValue {
    string startTime = 1;
    string endTime = 2;
    google.protobuf.Any value = 3; // Storing any type of value
}
```

Note that when creating the Entity, if you don't pass the name field, the "Any" value inside will default to a null value. This will cause Ballerina to throw an error as it can't handle null values in this context. Thus, always ensure that when passing an empty name field you must include the field with an empty string for the value part.

For example, this will throw an error as the name field is not present:

```bal
Entity relFilterName = {
        id: entityId,
        relationships: [{key: "", value: {name: "linked"}}]
    };
```

But this will not throw an error as though the name is empty, the field itself is still present:

```bal
Entity relFilterName = {
        id: entityId,
        name: {
            value: check pbAny:pack("")
        },
        relationships: [{key: "", value: {name: "linked"}}]
    };
```

* Note this doesn't apply to other fields. If you don't want to include a field's value, you don't need to pass the field at all. 

## Debug with Choreo

```bash
choreo connect --project ldf_sandbox_vibhatha --component crud-service
```