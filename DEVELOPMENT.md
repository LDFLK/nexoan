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


### Docker (Standalone): WIP

```bash
docker build -t all-services-test-standalone -f Dockerfile .
```

```bash
docker run --rm all-services-test-standalone
```

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
