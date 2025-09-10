# Neo4j Migration Guide: Docker Container to Aura

This guide provides step-by-step instructions for migrating a Neo4j database from a Docker container to Neo4j Aura.

## Prerequisites

- Docker installed and running
- Access to your Neo4j Docker container
- Neo4j Aura account and database instance
- Neo4j Aura connection URI, username, and password

## Migration Steps

### 1. Stop the Neo4j Container

Make sure the neo4j container is stopped before creating the dump.

### 2. Identify where the data is stored in the container

Find the directory path that's mapped to /data in your neo4j container, this is where the data is stored. 

Run `docker inspect <container_name>` and look for the Source path in the Mounts section.

This will look something like this: `/var/lib/docker/volumes/neo4j_data/_data`

### 3. Create a Local Dump Folder

Create a folder on your local machine to store the dump file.

### 4. Create a Database Dump

Run the following command to create a dump from your Neo4j Docker container:  

```bash
docker run --rm \
--volume=/var/lib/docker/volumes/neo4j_data/_data:/data \
--volume=/Users/your_username/Documents/neo4j_dump:/backups \
neo4j/neo4j-admin:latest \
neo4j-admin database dump neo4j --to-path=/backups
```

**Important:** Replace the following placeholders with your actual values:
- `/var/lib/docker/volumes/neo4j_data/_data` with the correct path to the data in your container.
- `/Users/your_username/Documents/neo4j_dump` with the correct path to your local folder to store the dump file.

### 4. Verify the Database Dump was created

Navigate to the local folder specified previously and check that a dump file has been created inside.

### 4. Upload the Dump to Neo4j Aura

Run the following command to upload the dump to your Neo4j Aura instance:

```bash
docker run --rm \
--volume=/Users/your_username/Documents/neo4j_dump:/dump \
neo4j/neo4j-admin:latest \
neo4j-admin database upload neo4j \
--from-path=/dump \
--to-uri=<neo4j_uri> \
--to-user=<neo4j_user> \
--to-password=<neo4j_password> \
--overwrite-destination=true
```

**Important:** Replace the following placeholders with your actual values:
- `/Users/your_username/Documents/neo4j_dump` with the correct path to your local folder with the dump file.
- `<neo4j_uri>`,`<neo4j_user>`,`<neo4j_password>` with your actual aura db credentials

### 5. Verify the Migration

Connect to your Neo4j Aura instance and verify that the new data has been inserted.
