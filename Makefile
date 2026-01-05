.PHONY: help dev test build docker migrate-up migrate-down lint clean

# Default target
.DEFAULT_GOAL := help

## help: Show this help message
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

## dev: Start development environment
dev:
	docker-compose up -d postgres redis nats
	@echo "Waiting for services to be ready..."
	@sleep 3
	@echo "Development environment is ready!"
	@echo "PostgreSQL: localhost:5432"
	@echo "Redis: localhost:6379"
	@echo "NATS: localhost:4222"

## dev-down: Stop development environment
dev-down:
	docker-compose down

## migrate-up: Run database migrations up
migrate-up:
	@echo "Running migrations..."
	docker-compose exec -T postgres psql -U hpmmo -d hpmmo < migrations/001_initial.up.sql
	@echo "Migrations completed!"

## migrate-down: Rollback last migration
migrate-down:
	@echo "Rolling back migrations..."
	docker-compose exec -T postgres psql -U hpmmo -d hpmmo < migrations/001_initial.down.sql
	@echo "Rollback completed!"

## seed: Seed database with initial data
seed:
	@echo "Seeding database..."
	docker-compose exec -T postgres psql -U hpmmo -d hpmmo < scripts/seed_items.sql
	docker-compose exec -T postgres psql -U hpmmo -d hpmmo < scripts/seed_zones.sql
	@echo "Database seeded!"

## test: Run all tests
test:
	go test -v -race -cover ./...

## test-integration: Run integration tests
test-integration:
	docker-compose up -d
	sleep 3
	go test -v -tags=integration ./...

## build: Build all services
build:
	@echo "Building services..."
	CGO_ENABLED=0 GOOS=linux go build -o bin/api ./cmd/api
	CGO_ENABLED=0 GOOS=linux go build -o bin/zone ./cmd/zone
	@echo "Build completed!"

## build-local: Build for local development
build-local:
	@echo "Building for local development..."
	go build -o bin/api ./cmd/api
	go build -o bin/zone ./cmd/zone
	@echo "Build completed!"

## run-api: Run API server locally
run-api: build-local
	./bin/api

## run-zone: Run Zone server locally
run-zone: build-local
	./bin/zone

## docker-build: Build Docker images
docker-build:
	docker build -t hp-mmo-api:latest -f deployments/docker/Dockerfile.api .
	docker build -t hp-mmo-zone:latest -f deployments/docker/Dockerfile.zone .

## docker-push: Push Docker images
docker-push: docker-build
	docker push hp-mmo-api:latest
	docker push hp-mmo-zone:latest

## lint: Run linter
lint:
	golangci-lint run --timeout 5m

## fmt: Format code
fmt:
	go fmt ./...
	goimports -w .

## clean: Clean build artifacts
clean:
	rm -rf bin/
	rm -rf tmp/
	go clean -cache

## logs: Tail docker-compose logs
logs:
	docker-compose logs -f

## psql: Connect to PostgreSQL
psql:
	docker-compose exec postgres psql -U hpmmo -d hpmmo

## redis-cli: Connect to Redis
redis-cli:
	docker-compose exec redis redis-cli

## generate: Generate code (mocks, etc.)
generate:
	go generate ./...
