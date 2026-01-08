.PHONY: all build run clean test deps

# Go binary name
BINARY_NAME=blockchain

# Go module name
MODULE_NAME=github.com/aspertheghost/blockchain

# Default target
all: deps build

# Download dependencies
deps:
	go mod download

# Build the binary
build:
	go build -o $(BINARY_NAME) ./cmd/main.go

# Run the program
run: build
	./$(BINARY_NAME)

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME)
	rm -f *.db
	rm -f *.dat

# Run tests
test:
	go test ./...

# Install the binary to GOPATH/bin
install: build
	go install ./cmd/main.go

# Format code
fmt:
	go fmt ./...

# Vet code
vet:
	go vet ./...

# Show help
help:
	@echo "Available targets:"
	@echo "  all      - Download dependencies and build (default)"
	@echo "  deps     - Download Go dependencies"
	@echo "  build    - Build the binary"
	@echo "  run      - Build and run the program"
	@echo "  clean    - Remove binary, databases, and wallet files"
	@echo "  test     - Run all tests"
	@echo "  install  - Install binary to GOPATH/bin"
	@echo "  fmt      - Format Go code"
	@echo "  vet      - Run go vet"
	@echo "  help     - Show this help message"

