.PHONY: help build up down logs clean migrate test lint

# Default target
help:
	@echo "MailGo - Development Commands"
	@echo ""
	@echo "Usage:"
	@echo "  make build       Build all Docker images"
	@echo "  make up          Start all services"
	@echo "  make down        Stop all services"
	@echo "  make logs        View logs"
	@echo "  make clean       Remove all containers and volumes"
	@echo "  make migrate     Run database migrations"
	@echo "  make test        Run tests"
	@echo "  make lint        Run linters"
	@echo "  make dev-api     Run API server locally"
	@echo "  make dev-worker  Run worker locally"
	@echo "  make dev-frontend Run frontend locally"

# Build all services
build:
	docker-compose build

# Start all services
up:
	docker-compose up -d
	@echo "Waiting for services to be healthy..."
	@sleep 10
	@echo "Services started!"
	@echo "Frontend: http://localhost:3000"
	@echo "API: http://localhost:8080"
	@echo "Mailcow: http://localhost:8081"

# Stop all services
down:
	docker-compose down

# View logs
logs:
	docker-compose logs -f

# Clean everything
clean:
	docker-compose down -v
	rm -rf data/

# Run database migrations
migrate:
	docker-compose exec control-plane sh -c "cd /app && ./migrate -path migrations -database \$$DATABASE_URL up"

# Run tests (backend)
test:
	cd backend && go test -v ./...

# Run linters
lint:
	cd backend && golangci-lint run
	cd frontend && npm run lint

# Development mode - API
dev-api:
	cd backend && go run cmd/api/main.go

# Development mode - Worker
dev-worker:
	cd backend && go run cmd/worker/main.go

# Development mode - Frontend
dev-frontend:
	cd frontend && npm run dev

# Install dependencies
deps:
	cd backend && go mod download
	cd frontend && npm install

# Generate protobuf files
proto:
	cd backend && protoc --go_out=. --go-grpc_out=. proto/*.proto

# Database reset (DANGEROUS)
db-reset:
	docker-compose down -v postgres
	docker-compose up -d postgres
	@sleep 5
	$(MAKE) migrate

# Check system health
health:
	@echo "Checking API health..."
	@curl -s http://localhost:8080/health || echo "API not responding"
	@echo "\nChecking Frontend health..."
	@curl -s http://localhost:3000 > /dev/null && echo "Frontend OK" || echo "Frontend not responding"
