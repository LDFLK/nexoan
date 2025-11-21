# Ingestion API

## Generate Open API 

This will generate the endpoints for the Ingestion API server using the OpenAPI specification. 
The OpenAPI specification is the base for public API for Ingestion API.

> 💡 Note: Always make sure the contract has the expected endpoints and request params
> before working on the code. The generated endpoints should not be editable at all. 
> Maybe the only changes that can be done is adding error handlers, but request and response
> must be defined in the contract. 


```bash
bal openapi -i ../contracts/rest/ingestion_api.yaml --mode service
```

## Generate GRPC Stubs

The client stub generated here will be sending and receiving values via Grpc. 
This will send requests to the corresponding CORE server endpoint. 

```bash
bal grpc --mode client --input ../core-api/protos/types_v1.proto --output .
```

> 💡 **Note**  
> At the generation make sure to remove any sample code generated to show how to use the API. Because that might add an unnecessary main file. 

## Set Environmental Variables

Following are the default values you should use. 

```bash
export CORE_SERVICE_HOST=localhost
export CORE_SERVICE_PORT=50051
export INGESTION_SERVICE_HOST=localhost
export INGESTION_SERVICE_PORT=8080
```

## Development

```bash
cd ingestion-api
cp env.template .env
# update the required fields to set the environment variables
source .env
bal test
```

## Run Test

Make sure the CORE server is running. (`cd design/core-api; ./core-server`)

```bash
# Run all tests in the current package
bal test

# Run tests with verbose output
bal test --test-report

# Run a specific test file
bal test tests/service_test.bal

# Run a specific test function
bal test --tests testMetadataHandling

# Run tests and generate a coverage report
bal test --code-coverage
```

## Run Service

```bash
cd ingestion-api
bal run
```

At the moment the port is hardcoded to 8080. This must be configurable via a config file.

