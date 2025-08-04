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