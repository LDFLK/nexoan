# Neo4j Migration Guide: Docker Container to Aura

This guide provides step-by-step instructions for migrating a Neo4j database from a Docker container to Neo4j Aura.

## Prerequisites

- Docker installed and running
- Access to your Neo4j Docker container
- Neo4j Aura account and database instance
- Neo4j Aura connection URI, username, and password

## Migration Steps

### 1. Check Neo4j Version in Container

```bash
docker exec <container_name> neo4j --version
```

Example:
```bash
docker exec neo4j_export neo4j --version
# Output: 5.12.0
```

### 2. Create a Dump Directory in the Container

```bash
docker exec <container_name> mkdir -p /var/lib/neo4j/dumps
```

Example:
```bash
docker exec neo4j_export mkdir -p /var/lib/neo4j/dumps
```

### 3. Stop the Neo4j Database (Not the Container)

```bash
docker exec -it <container_name> cypher-shell -u neo4j -p <container_password>
```

Then in the Cypher shell:
```cypher
STOP DATABASE neo4j;
:exit
```

### 4. Create a Database Dump

For Neo4j 4.x:
```bash
docker exec <container_name> neo4j-admin dump --database=neo4j --to=/var/lib/neo4j/dumps
```

For Neo4j 5.x:
```bash
docker exec <container_name> neo4j-admin database dump neo4j --to-path=/var/lib/neo4j/dumps
```

Example:
```bash
docker exec neo4j_export neo4j-admin database dump neo4j --to-path=/var/lib/neo4j/dumps
```

### 5. Verify the Dump File Was Created

```bash
docker exec <container_name> ls -la /var/lib/neo4j/dumps
```

Example:
```bash
docker exec neo4j_export ls -la /var/lib/neo4j/dumps
# Should show neo4j.dump file
```

### 6. Upload the Dump to Neo4j Aura

For Neo4j 5.x:
```bash
docker exec <container_name> neo4j-admin database upload neo4j --from-path=/var/lib/neo4j/dumps --to-uri=<aura_connection_uri> --overwrite-destination=true --to-user=neo4j --to-password="<your_aura_password>"
```

Example:
```bash
docker exec neo4j_export neo4j-admin database upload neo4j --from-path=/var/lib/neo4j/dumps --to-uri=neo4j+s://4dbc2ec5.databases.neo4j.io --overwrite-destination=true --to-user=neo4j --to-password="YsgI6WFU6n-M0fyLurwUlIR0TVnM-KuGZMKScqHMQ6s"
```

Add `--verbose` flag for detailed output if needed.

### 7. Alternative: Manual Upload

If the direct upload doesn't work, you can extract the dump file and upload it manually:

1. Copy the dump file from the container to your local machine:
   ```bash
   docker cp <container_name>:/var/lib/neo4j/dumps/neo4j.dump ./neo4j.dump
   ```

2. Upload the dump file through the Neo4j Aura web interface.

### 8. Restart the Neo4j Database in the Container (If Needed)

```bash
docker exec -it <container_name> cypher-shell -u neo4j -p <container_password>
```

Then in the Cypher shell:
```cypher
START DATABASE neo4j;
:exit
```

### 9. Verify the Migration

Connect to your Neo4j Aura instance using the Neo4j Browser at:
```
https://<your-instance-id>.databases.neo4j.io
```

Run a query to verify your data is present:
```cypher
MATCH (n) RETURN count(n);
```
