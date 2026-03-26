#!/bin/bash
# Run all tests with coverage

set -e

echo "Running tests..."

# Run unit tests
go test ./internal/... -v -race -cover

# Run integration tests
echo ""
echo "Running integration tests..."
go test ./test/integration/... -v -race

# Generate coverage report
echo ""
echo "Generating coverage report..."
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html

echo ""
echo "Coverage report generated: coverage.html"
