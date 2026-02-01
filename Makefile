.PHONY: all build test clean run-basic run-filesystem run-todo run-composite run-skills

all: test build

build:
	@echo "Building Deep Agents Go..."
	@go build -o bin/basic ./cmd/examples/basic
	@go build -o bin/filesystem ./cmd/examples/filesystem
	@go build -o bin/todo ./cmd/examples/todo
	@go build -o bin/composite ./cmd/examples/composite
	@go build -o bin/skills ./cmd/examples/skills
	@go build -o bin/bash ./cmd/examples/bash
	@go build -o bin/deepagents ./cmd/deepagents
	@echo "Build complete!"

test:
	@echo "Running tests..."
	@go test -v -cover ./...

test-coverage:
	@echo "Running tests with coverage..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -rf workspace/
	@rm -rf workspace_todo/
	@rm -f coverage.out coverage.html
	@echo "Clean complete!"

run-basic:
	@echo "Running basic example..."
	@go run ./cmd/examples/basic/main.go

run-filesystem:
	@echo "Running filesystem example..."
	@go run ./cmd/examples/filesystem/main.go

run-todo:
	@echo "Running todo example..."
	@go run ./cmd/examples/todo/main.go

run-composite:
	@echo "Running composite example..."
	@go run ./cmd/examples/composite/main.go

run-skills:
	@echo "Running skills example..."
	@go run ./cmd/examples/skills/main.go

run-bash:
	@echo "Running bash example..."
	@go run ./cmd/examples/bash/main.go

fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@echo "Format complete!"

lint:
	@echo "Running linter..."
	@golangci-lint run ./...

deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy
	@echo "Dependencies updated!"

help:
	@echo "Deep Agents Go - Makefile commands:"
	@echo "  make build          - Build all examples"
	@echo "  make test           - Run all tests"
	@echo "  make test-coverage  - Run tests with coverage report"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make run-basic      - Run basic example"
	@echo "  make run-filesystem - Run filesystem example"
	@echo "  make run-todo       - Run todo example"
	@echo "  make run-composite  - Run composite example"
	@echo "  make run-skills     - Run skills example"
	@echo "  make fmt            - Format code"
	@echo "  make lint           - Run linter"
	@echo "  make deps           - Update dependencies"
