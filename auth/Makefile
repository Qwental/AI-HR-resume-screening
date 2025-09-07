# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Binary name
BINARY_NAME=hr-avatar
BINARY_UNIX=$(BINARY_NAME)_unix

# Docker parameters
DOCKER_IMAGE=hr-avatar
DOCKER_TAG=latest
DOCKER_COMPOSE=docker-compose

.PHONY: all build clean test deps help

# Default target
all: test build

## Build the binary
build:
	$(GOBUILD) -o $(BINARY_NAME) -v ./cmd/server/main.go

## Run the application
run:
	$(GOBUILD) -o $(BINARY_NAME) -v ./cmd/server/main.go
	./$(BINARY_NAME)

## Run tests
test:
	$(GOTEST) -v ./...

## Run tests with coverage
test-coverage:
	$(GOTEST) -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out

## Clean build artifacts
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)

## Download dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

## Cross compilation for Linux
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_UNIX) -v ./cmd/server/main.go

## Docker commands

## Build Docker image
docker-build:
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

## Run single container
docker-run:
	docker run -p 8080:8080 --env-file .env $(DOCKER_IMAGE):$(DOCKER_TAG)

## Build and run with docker-compose
docker-up:
	$(DOCKER_COMPOSE) up --build -d

## Stop docker-compose services
docker-down:
	$(DOCKER_COMPOSE) down

## Restart docker-compose services
docker-restart:
	$(DOCKER_COMPOSE) restart

## View logs
docker-logs:
	$(DOCKER_COMPOSE) logs -f

## View app logs only
docker-logs-app:
	$(DOCKER_COMPOSE) logs -f app

## Start only database
docker-db:
	$(DOCKER_COMPOSE) up -d postgres redis

## Run database migrations (if you have them)
docker-migrate:
	$(DOCKER_COMPOSE) exec app ./main migrate

## Access database console
docker-db-console:
	$(DOCKER_COMPOSE) exec postgres psql -U postgres -d auth_demo

## Remove all containers and volumes
docker-clean:
	$(DOCKER_COMPOSE) down -v
	docker system prune -f

## Development setup
dev-setup:
	cp .env.example .env
	$(GOMOD) download
	$(GOMOD) tidy
	$(DOCKER_COMPOSE) up -d postgres redis

## Full development environment
dev:
	make dev-setup
	make run

## Production deployment
prod-deploy:
	$(DOCKER_COMPOSE) -f docker-compose.yml -f docker-compose.prod.yml up -d

## Show help
help:
	@echo ''
	@echo 'Usage:'
	@echo '  make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} { \
		if (/^[a-zA-Z_-]+:.*?##.*$$/) printf "  %-20s%s\n", $$1, $$2 \
	}' $(MAKEFILE_LIST)