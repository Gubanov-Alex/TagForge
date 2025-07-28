.PHONY: help build run test lint fmt swagger migrate clean up down logs shell

# Variables
APP_NAME := config-service
BINARY_NAME := main
DOCKER_COMPOSE_FILE := deployments/docker-compose.yml
MIGRATION_PATH := migrations
DATABASE_URL := postgres://postgres:postgres@localhost:5432/config_service?sslmode=disable

# Colors for output
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[0;33m
BLUE := \033[0;34m
MAGENTA := \033[0;35m
CYAN := \033[0;36m
WHITE := \033[0;37m
RESET := \033[0m

# Default target
help: ## Show this help message
	@echo "$(CYAN)Available commands:$(RESET)"
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make $(GREEN)<target>$(RESET)\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  $(GREEN)%-15s$(RESET) %s\n", $$1, $$2 } /^##@/ { printf "\n$(MAGENTA)%s$(RESET)\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development
build: ## Build the application binary
	@echo "$(BLUE)Building $(APP_NAME)...$(RESET)"
	go build -ldflags="-s -w" -o bin/$(BINARY_NAME) cmd/server/main.go
	@echo "$(GREEN)Build completed: bin/$(BINARY_NAME)$(RESET)"

build-docker: ## Build Docker image
	@echo "$(BLUE)Building Docker image...$(RESET)"
	docker build -f deployments/Dockerfile -t $(APP_NAME):latest .
	@echo "$(GREEN)Docker image built: $(APP_NAME):latest$(RESET)"

run: ## Run the application locally
	@echo "$(BLUE)Running $(APP_NAME)...$(RESET)"
	go run cmd/server/main.go

run-docker: build-docker ## Run the application in Docker
	@echo "$(BLUE)Running $(APP_NAME) in Docker...$(RESET)"
	docker run --rm -p 8080:8080 \
		-e DATABASE_HOST=host.docker.internal \
		-e REDIS_HOST=host.docker.internal \
		-e KAFKA_BROKERS=host.docker.internal:9092 \
		$(APP_NAME):latest

##@ Infrastructure
up: ## Start all services with Docker Compose
	@echo "$(BLUE)Starting all services...$(RESET)"
	docker-compose -f $(DOCKER_COMPOSE_FILE) up -d
	@echo "$(GREEN)All services started!$(RESET)"
	@echo "$(YELLOW)Waiting for services to be ready...$(RESET)"
	@sleep 10
	@make status

down: ## Stop all services
	@echo "$(BLUE)Stopping all services...$(RESET)"
	docker-compose -f $(DOCKER_COMPOSE_FILE) down
	@echo "$(GREEN)All services stopped!$(RESET)"

restart: down up ## Restart all services

status: ## Show status of all services
	@echo "$(CYAN)Service Status:$(RESET)"
	docker-compose -f $(DOCKER_COMPOSE_FILE) ps

logs: ## Show logs from all services
	docker-compose -f $(DOCKER_COMPOSE_FILE) logs -f

logs-app: ## Show application logs only
	docker-compose -f $(DOCKER_COMPOSE_FILE) logs -f config-service

##@ Database
migrate-up: ## Run database migrations
	@echo "$(BLUE)Running database migrations...$(RESET)"
	migrate -path $(MIGRATION_PATH) -database "$(DATABASE_URL)" up
	@echo "$(GREEN)Migrations completed!$(RESET)"

migrate-down: ## Rollback one migration
	@echo "$(BLUE)Rolling back one migration...$(RESET)"
	migrate -path $(MIGRATION_PATH) -database "$(DATABASE_URL)" down 1
	@echo "$(GREEN)Rollback completed!$(RESET)"

migrate-reset: ## Reset all migrations
	@echo "$(RED)WARNING: This will reset all migrations!$(RESET)"
	@read -p "Are you sure? [y/N] " -n 1 -r; \
	if [[ $REPLY =~ ^[Yy]$ ]]; then \
		echo "$(BLUE)Resetting migrations...$(RESET)"; \
		migrate -path $(MIGRATION_PATH) -database "$(DATABASE_URL)" down -all; \
		echo "$(GREEN)All migrations reset!$(RESET)"; \
	else \
		echo "$(YELLOW)Operation cancelled.$(RESET)"; \
	fi

migrate-version: ## Show current migration version
	migrate -path $(MIGRATION_PATH) -database "$(DATABASE_URL)" version

migrate-create: ## Create new migration (usage: make migrate-create NAME=create_users_table)
	@if [ -z "$(NAME)" ]; then \
		echo "$(RED)Error: NAME is required$(RESET)"; \
		echo "Usage: make migrate-create NAME=create_users_table"; \
		exit 1; \
	fi
	@echo "$(BLUE)Creating migration: $(NAME)$(RESET)"
	migrate create -ext sql -dir $(MIGRATION_PATH) -seq $(NAME)
	@echo "$(GREEN)Migration files created!$(RESET)"

##@ Testing
test: ## Run all tests
	@echo "$(BLUE)Running tests...$(RESET)"
	go test -v -race -cover ./...

test-coverage: ## Run tests with coverage report
	@echo "$(BLUE)Running tests with coverage...$(RESET)"
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Coverage report generated: coverage.html$(RESET)"

test-integration: ## Run integration tests
	@echo "$(BLUE)Running integration tests...$(RESET)"
	go test -v -tags=integration ./...

benchmark: ## Run benchmarks
	@echo "$(BLUE)Running benchmarks...$(RESET)"
	go test -bench=. -benchmem ./...

##@ Code Quality
lint: ## Run linters
	@echo "$(BLUE)Running linters...$(RESET)"
	golangci-lint run
	@echo "$(GREEN)Linting completed!$(RESET)"

fmt: ## Format code
	@echo "$(BLUE)Formatting code...$(RESET)"
	gofmt -s -w .
	goimports -w .
	@echo "$(GREEN)Code formatted!$(RESET)"

vet: ## Run go vet
	@echo "$(BLUE)Running go vet...$(RESET)"
	go vet ./...

mod-tidy: ## Tidy go modules
	@echo "$(BLUE)Tidying go modules...$(RESET)"
	go mod tidy
	@echo "$(GREEN)Modules tidied!$(RESET)"

##@ Documentation
swagger: ## Generate Swagger documentation
	@echo "$(BLUE)Generating Swagger documentation...$(RESET)"
	swag init -g cmd/server/main.go --output docs/swagger
	@echo "$(GREEN)Swagger documentation generated!$(RESET)"

docs-serve: ## Serve documentation locally
	@echo "$(BLUE)Starting documentation server...$(RESET)"
	@echo "$(CYAN)Swagger UI: http://localhost:8080/swagger/index.html$(RESET)"
	@make run

##@ Utilities
shell: ## Open shell in running container
	docker-compose -f $(DOCKER_COMPOSE_FILE) exec config-service sh

shell-db: ## Open PostgreSQL shell
	docker-compose -f $(DOCKER_COMPOSE_FILE) exec postgres psql -U postgres -d config_service

shell-redis: ## Open Redis CLI
	docker-compose -f $(DOCKER_COMPOSE_FILE) exec redis redis-cli

clean: ## Clean build artifacts
	@echo "$(BLUE)Cleaning build artifacts...$(RESET)"
	rm -rf bin/
	rm -f coverage.out coverage.html
	docker system prune -f
	@echo "$(GREEN)Clean completed!$(RESET)"

clean-all: clean ## Clean everything including volumes
	@echo "$(RED)WARNING: This will remove all Docker volumes!$(RESET)"
	@read -p "Are you sure? [y/N] " -n 1 -r; \
	if [[ $REPLY =~ ^[Yy]$ ]]; then \
		echo "$(BLUE)Removing all volumes...$(RESET)"; \
		docker-compose -f $(DOCKER_COMPOSE_FILE) down -v; \
		docker volume prune -f; \
		echo "$(GREEN)All volumes removed!$(RESET)"; \
	else \
		echo "$(YELLOW)Operation cancelled.$(RESET)"; \
	fi

##@ Monitoring
monitoring-up: ## Start monitoring stack (Prometheus + Grafana)
	@echo "$(BLUE)Starting monitoring stack...$(RESET)"
	docker-compose -f $(DOCKER_COMPOSE_FILE) up -d prometheus grafana
	@echo "$(GREEN)Monitoring stack started!$(RESET)"
	@echo "$(CYAN)Prometheus: http://localhost:9090$(RESET)"
	@echo "$(CYAN)Grafana: http://localhost:3000 (admin/admin)$(RESET)"

kafka-topics: ## List Kafka topics
	docker-compose -f $(DOCKER_COMPOSE_FILE) exec kafka kafka-topics --bootstrap-server localhost:9092 --list

kafka-console-producer: ## Start Kafka console producer
	docker-compose -f $(DOCKER_COMPOSE_FILE) exec kafka kafka-console-producer --bootstrap-server localhost:9092 --topic config-events

kafka-console-consumer: ## Start Kafka console consumer
	docker-compose -f $(DOCKER_COMPOSE_FILE) exec kafka kafka-console-consumer --bootstrap-server localhost:9092 --topic config-events --from-beginning

##@ Production
build-prod: ## Build production binary
	@echo "$(BLUE)Building production binary...$(RESET)"
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
		-ldflags="-s -w -extldflags '-static'" \
		-a -installsuffix cgo \
		-o bin/$(BINARY_NAME)-linux-amd64 \
		cmd/server/main.go
	@echo "$(GREEN)Production binary built: bin/$(BINARY_NAME)-linux-amd64$(RESET)"

deploy-staging: build-prod ## Deploy to staging environment
	@echo "$(BLUE)Deploying to staging...$(RESET)"
	# Add your staging deployment commands here
	@echo "$(GREEN)Deployed to staging!$(RESET)"

deploy-prod: build-prod ## Deploy to production environment
	@echo "$(RED)WARNING: Deploying to production!$(RESET)"
	@read -p "Are you sure? [y/N] " -n 1 -r; \
	if [[ $REPLY =~ ^[Yy]$ ]]; then \
		echo "$(BLUE)Deploying to production...$(RESET)"; \
		# Add your production deployment commands here; \
		echo "$(GREEN)Deployed to production!$(RESET)"; \
	else \
		echo "$(YELLOW)Deployment cancelled.$(RESET)"; \
	fi

##@ CI/CD
ci-test: fmt vet lint test ## Run CI tests locally
	@echo "$(GREEN)All CI checks passed!$(RESET)"

pre-commit: ci-test swagger ## Run pre-commit checks
	@echo "$(GREEN)Ready for commit!$(RESET)"

##@ Information
version: ## Show application version
	@go run cmd/server/main.go -version 2>/dev/null || echo "Run 'make build' first"

deps: ## Show dependencies
	@echo "$(CYAN)Go Dependencies:$(RESET)"
	go list -m all

env: ## Show environment variables
	@echo "$(CYAN)Environment Variables:$(RESET)"
	@echo "DATABASE_URL=$(DATABASE_URL)"
	@echo "MIGRATION_PATH=$(MIGRATION_PATH)"
	@echo "APP_NAME=$(APP_NAME)"

health: ## Check service health
	@echo "$(BLUE)Checking service health...$(RESET)"
	@curl -s http://localhost:8080/health | jq . 2>/dev/null || \
		echo "$(RED)Service not running or jq not installed$(RESET)"

api-test: ## Test API endpoints
	@echo "$(BLUE)Testing API endpoints...$(RESET)"
	@echo "$(CYAN)Health endpoint:$(RESET)"
	@curl -s http://localhost:8080/health || echo "$(RED)Failed$(RESET)"
	@echo ""
	@echo "$(CYAN)Ready endpoint:$(RESET)"
	@curl -s http://localhost:8080/ready || echo "$(RED)Failed$(RESET)"
	@echo ""
	@echo "$(CYAN)Metrics endpoint:$(RESET)"
	@curl -s http://localhost:8080/metrics | head -5 || echo "$(RED)Failed$(RESET)"

##@ Quick Start
install-tools: ## Install required development tools
	@echo "$(BLUE)Installing development tools...$(RESET)"
	go install github.com/swaggo/swag/cmd/swag@latest
	go install github.com/golang/mock/mockgen@latest
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.54.2
	go install golang.org/x/tools/cmd/goimports@latest
	@echo "$(GREEN)Development tools installed!$(RESET)"

setup: install-tools up migrate-up swagger ## Complete project setup
	@echo "$(GREEN)Project setup completed!$(RESET)"
	@echo "$(CYAN)Next steps:$(RESET)"
	@echo "  - API Documentation: http://localhost:8080/swagger/index.html"
	@echo "  - Kafka UI: http://localhost:8090"
	@echo "  - Prometheus: http://localhost:9090"
	@echo "  - Grafana: http://localhost:3000"
	@echo "  - Run tests: make test"
	@echo "  - View logs: make logs-app"
