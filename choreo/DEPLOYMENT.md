# DEPLOYMENT

## CRUD SERVICE

We deploy the CRUD service as a gRPC service and we have to add a few environmental variables 
and a few file mounts to make things work. 

### Environmental Variables 

1. MONGO_URI
2. MONGO_DB_NAME
3. MONGO_COLLECTION
4. NEO4J_URI
5. NEO4J_USER
6. NEO4J_PASSWORD
7. CRUD_SERVICE_HOST
8. CRUD_SERVICE_PORT
9. POSTGRES_USER
10. POSTGRES_HOST
11. POSTGRES_PORT
12. POSTGRES_DB
13. POSTGRES_SSL_MODE
14. POSTGRES_PASSWORD

### File Mounts

| # | Mount Name | Type | Mount Path | Description |
|---|------------|------|------------|-------------|
| 1 | crud-go-build-mnt | Empty Directory (In-Memory) | /home/choreouser/.cache | Go build cache directory |
| 2 | default-tmp-emptydir | Empty Directory (In-Memory) | /tmp | Temporary files directory |
| 3 | mnt-go-core-dir | Empty Directory (In-Memory) | /go | Go core directory |

### Choreo Configs

When deploying the CRUD service on thing to note is that GRPC services are not exposed through the Gateway in Choreo. So we have to choose the `PROJECT_URL` from `Manage`->`Overview` tabs in Choreo
console. Make sure to extract that URL and use it as the `crudServiceURL` config in both `Update` API and
`Query` API services.

### Choreo Configuration Groups

Note that when you add variables to your deployment configurations through Choreo Configuration Groups, 
it is a must to remove the prefix added during the deployment time, otherwise, the code won't be able to interpret it. 

For instance if you set `DB_URI` as a config parameter, once you link the configuration group to your deployment, it will have a mapping `Environment Variable` -> `Configuration Param` as follows. And let's assume your component is `SERVICE`.

`SERVICE_ENV_VAR_DB_URI` and we need to remove `SERVICE_ENV_VAR` from it to make sure the code can understand it.