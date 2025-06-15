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

## Set Environmental Variables

Following are the default values you should use. 

```bash
export CRUD_SERVICE_HOST=localhost
export CRUD_SERVICE_PORT=50051
export UPDATE_SERVICE_HOST=localhost
export UPDATE_SERVICE_PORT=8080
```

## Development

```bash
cd design/update-api
cp env.template .env
# update the required fields to set the environment variables
source .env
bal test
```

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

# Update API Service

This service provides an API for updating entities in the LDF Architecture.

## SSL Certificate Setup

The service uses SSL certificates for secure communication. Here's how to set up the certificates for local development and testing:

### 1. Generate SSL Certificates

The certificates are stored in the `certs` directory. To generate new certificates:

```bash
# Create certs directory if it doesn't exist
mkdir -p certs

# Generate self-signed certificate
cd certs
openssl req -x509 -newkey rsa:4096 -keyout server.key -out server.crt -days 365 -nodes \
    -subj "/CN=localhost" \
    -addext "subjectAltName = DNS:localhost,IP:127.0.0.1"
```

This will create:
- `server.key`: The private key
- `server.crt`: The self-signed certificate

### 2. Certificate Configuration

The certificates are used in two places:

1. **gRPC Client Configuration** (in `tests/service_test.bal`):
```ballerina
grpc:ClientConfiguration grpcConfig = {
    secureSocket: {
        enable: true,
        cert: "../certs/server.crt"
    }
};
```

2. **HTTP Client Configuration** (in `tests/service_test.bal`):
```ballerina
http:ClientConfiguration httpConfig = {
    httpVersion: "2.0",
    secureSocket: {
        enable: true,
        cert: "../certs/server.crt"
    }
};
```

### 3. Environment Variables

The service uses the following environment variables:

- `CRUD_SERVICE_URL`: The URL of the CRUD service (e.g., `https://localhost:8080`)
- `CRUD_SERVICE_HOST`: The host of the CRUD service (fallback if URL not provided)
- `CRUD_SERVICE_PORT`: The port of the CRUD service (fallback if URL not provided)
- `UPDATE_SERVICE_HOST`: The host of the update service
- `UPDATE_SERVICE_PORT`: The port of the update service

### 4. Running Tests

To run the tests:

```bash
# Set environment variables
export CRUD_SERVICE_URL="https://localhost:8080"  # or your actual CRUD service URL
export UPDATE_SERVICE_HOST="localhost"
export UPDATE_SERVICE_PORT="8081"

# Run tests
bal test
```

### 5. Troubleshooting

If you encounter SSL certificate issues:

1. Verify that the certificates are in the correct location (`certs/server.crt` and `certs/server.key`)
2. Check that the certificate paths in the configurations are correct
3. Ensure the CRUD service is configured to use the same certificates
4. If using a different hostname, update the certificate's Subject Alternative Name (SAN)

### 6. Security Notes

- The self-signed certificates are for development and testing only
- For production, use proper SSL certificates from a trusted Certificate Authority
- Never commit private keys to version control
- Consider using environment variables for certificate paths in production

## Development

### Prerequisites

- Ballerina 2201.7.0 or later
- OpenSSL for certificate generation

### Building

```bash
bal build
```

### Running

```bash
bal run
```

## License

[Your License Here]

