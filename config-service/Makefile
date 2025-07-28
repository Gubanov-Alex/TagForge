.PHONY: help build run test lint fmt swagger migrate clean up down logs shell

# Variables
APP_NAME := config-service
BINARY_NAME := main
DOCKER_COMPOSE_FILE := deployments/docker-compose.yml
MIGRATION_PATH := migrations
DATABASE_URL := postgres://postgres:postgres@localhost:5431/config_service?sslmode=disable

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
up: ## Start infrastructure services (without app)
	@echo "$(BLUE)Starting infrastructure services...$(RESET)"
	docker-compose -f $(DOCKER_COMPOSE_FILE) up -d postgres redis zookeeper kafka kafka-ui prometheus grafana
	@echo "$(GREEN)Infrastructure services started!$(RESET)"
	@echo "$(YELLOW)Waiting for services to be ready...$(RESET)"
	@sleep 30
	@make status-infra

up-full: build-docker ## Start all services including application
	@echo "$(BLUE)Starting all services with application...$(RESET)"
	docker-compose -f $(DOCKER_COMPOSE_FILE) up -d
	@echo "$(GREEN)All services started!$(RESET)"
	@echo "$(YELLOW)Waiting for services to be ready...$(RESET)"
	@sleep 30
	@make status

down: ## Stop all services
	@echo "$(BLUE)Stopping all services...$(RESET)"
	docker-compose -f $(DOCKER_COMPOSE_FILE) down
	@echo "$(GREEN)All services stopped!$(RESET)"

restart: down up ## Restart infrastructure services

status: ## Show status of all services
	@echo "$(CYAN)Service Status:$(RESET)"
	docker-compose -f $(DOCKER_COMPOSE_FILE) ps

status-infra: ## Show status of infrastructure services only
	@echo "$(CYAN)Infrastructure Status:$(RESET)"
	docker-compose -f $(DOCKER_COMPOSE_FILE) ps postgres redis zookeeper kafka kafka-ui prometheus grafana

logs: ## Show logs from all services
	docker-compose -f $(DOCKER_COMPOSE_FILE) logs -f

logs-infra: ## Show logs from infrastructure services
	docker-compose -f $(DOCKER_COMPOSE_FILE) logs -f postgres redis kafka

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
	@if command -v swag >/dev/null 2>&1; then \
		swag init -g cmd/server/main.go --output docs/swagger; \
	else \
		echo "$(YELLOW)swag not found in PATH, using go run...$(RESET)"; \
		go run github.com/swaggo/swag/cmd/swag@latest init -g cmd/server/main.go --output docs/swagger; \
	fi
	@echo "$(GREEN)Swagger documentation generated!$(RESET)"

docs-serve: ## Serve documentation locally
	@echo "$(BLUE)Starting documentation server...$(RESET)"
	@echo "$(CYAN)Swagger UI: http://localhost:8080/swagger/index.html$(RESET)"
	@make run

##@ Utilities
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

##@ Quick Start
install-tools: ## Install required development tools
	@echo "$(BLUE)Installing development tools...$(RESET)"
	go install github.com/swaggo/swag/cmd/swag@latest
	go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	@echo "$(GREEN)Development tools installed!$(RESET)"
	@echo "$(YELLOW)Make sure $(shell go env GOPATH)/bin is in your PATH$(RESET)"

setup: install-tools up migrate-up swagger ## Complete project setup
	@echo "$(GREEN)Project setup completed!$(RESET)"
	@echo "$(CYAN)Next steps:$(RESET)"
	@echo "  - API Documentation: http://localhost:8080/swagger/index.html"
	@echo "  - Run application: make run"
	@echo "  - View logs: make logs-infra"

health: ## Check service health
	@echo "$(BLUE)Checking service health...$(RESET)"
	@curl -s http://localhost:8080/health | jq . 2>/dev/null || \
		echo "$(RED)Service not running or jq not installed$(RESET)"
