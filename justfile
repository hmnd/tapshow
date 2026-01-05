version := `cat VERSION`

# Build the binary using zig as the C/C++ compiler
build:
    CC="zig cc" CXX="zig c++" go build -ldflags "-X main.version={{version}}" -o bin/tapshow ./cmd/tapshow

# Run the application
run: build
    ./bin/tapshow

# Run all tests
test:
    go test ./...

# Run tests with verbose output
test-verbose:
    go test -v ./...

# Run tests with coverage
test-coverage:
    go test -cover ./...

# Clean build artifacts
clean:
    rm -rf bin/
