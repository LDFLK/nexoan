# Update API

## Generate Open API 

This will generate the endpoints for the Update API server using the OpenAPI specification. 
The OpenAPI specification is the base for public API for Update API.

> 💡 Note: Always make sure the contract has the expected endpoints and request params
> before working on the code. The generated endpoints should not be editable at all. 
> Maybe the only changes that can be done is adding error handlers, but request and response
> must be defined in the contract. 


```bash
bal openapi -i ../contracts/rest/update_api.yaml --mode service
```

## Generate GRPC Stubs

The client stub generated here will be sending and receiving values via Grpc. 
This will send requests to the corresponding CRUD server endpoint. 

```bash
bal grpc --mode client --input ../crud-api/protos/types_v1.proto --output .
```

> 💡 **Note**  
> At the generation make sure to remove any sample code generated to show how to use the API. Because that might add an unnecessary main file. 

## Run Test

Make sure the CRUD server is running. (`cd design/crud-api; ./crud-server`)

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
cd update-api
bal run
```

At the moment the port is hardcoded to 8080. This must be configurable via a config file.

