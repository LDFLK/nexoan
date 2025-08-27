echo "Building all packages..."
go build -v ./...
echo "Building crud-service binary..."
go build -v -o crud-service cmd/server/service.go cmd/server/utils.go
echo "Build complete!"