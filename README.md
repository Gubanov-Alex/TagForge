# Tag Management API Service

A high-performance microservice for managing tags, their versioning, and deployment across environments. Built with Go, PostgreSQL, Redis, and Kafka for scalability and reliability.

## 🚀 Features

- **CRUD Operations**: Complete tag lifecycle management
- **Version Control**: Git-like versioning system for tags
- **Environment Management**: Support for dev, staging, and production environments
- **Tag Templates**: Pre-built templates for Google Analytics, Facebook Pixel, and custom tags
- **Event Streaming**: Real-time event processing with Apache Kafka
- **Audit Trail**: Complete history of all tag modifications
- **Rate Limiting**: API protection with Redis-based rate limiting
- **Caching**: High-performance caching layer for frequently accessed data
- **API Documentation**: Auto-generated Swagger/OpenAPI documentation
- **Health Monitoring**: Comprehensive health checks and metrics

## 🛠 Technology Stack

- **Language**: Go 1.21+
- **Web Framework**: Gin
- **Database**: PostgreSQL
- **Cache**: Redis
- **Message Broker**: Apache Kafka
- **API Documentation**: Swagger/OpenAPI
- **Containerization**: Docker & Docker Compose
- **Migration Tool**: golang-migrate
- **Monitoring**: Prometheus metrics
- **Tracing**: OpenTelemetry

## 📋 Prerequisites

- Go 1.21 or higher
- Docker and Docker Compose
- Make (optional, for using Makefile commands)

## 🏗 Project Structure

```
tag-management-api/
├── cmd/
│   └── server/
│       └── main.go                 # Application entrypoint
├── internal/
│   ├── api/
│   │   ├── v1/
│   │   │   ├── handler/           # HTTP handlers
│   │   │   ├── middleware/        # Custom middleware
│   │   │   └── router.go          # Route definitions
│   │   └── health/                # Health check handlers
│   ├── config/
│   │   └── config.go              # Configuration structure
│   ├── database/
│   │   ├── connection.go          # DB connection setup
│   │   └── migrations.go          # Migration runner
│   ├── model/
│   │   ├── tag.go                 # Tag entity
│   │   ├── environment.go         # Environment entity
│   │   └── template.go            # Template entity
│   ├── repository/
│   │   └── ...                    # Data access layer
│   ├── service/
│   │   └── ...                    # Business logic layer
│   └── kafka/
│       ├── producer.go            # Kafka producer
│       └── consumer.go            # Kafka consumers
├── pkg/
│   ├── errors/                    # Custom error types
│   └── logger/                    # Structured logging
├── migrations/                    # Database migrations
├── deployments/
│   ├── docker-compose.yml         # Local development stack
│   └── Dockerfile                 # Application container
├── docs/
│   └── swagger/                   # API documentation
├── Makefile                       # Development tasks
├── go.mod                         # Go dependencies
└── README.md                      # This file
```

## 🚦 Quick Start

### 1. Clone the repository
```bash
git clone https://github.com/yourusername/tag-management-api.git
cd tag-management-api
```

### 2. Set up environment variables
```bash
cp .env.example .env
# Edit .env with your configuration
```

### 3. Start the infrastructure
```bash
make up
```

This will start:
- PostgreSQL (port 5432)
- Redis (port 6379)
- Kafka & Zookeeper (ports 9092, 2181)
- API Service (port 8080)

### 4. Run database migrations
```bash
make migrate-up
```

### 5. Access the API
- API: http://localhost:8080
- Swagger UI: http://localhost:8080/swagger/index.html
- Health Check: http://localhost:8080/health

## 🔧 Development

### Available Make Commands

```bash
make build          # Build the application
make run           # Run the application locally
make test          # Run tests
make test-coverage # Run tests with coverage
make lint          # Run linter
make fmt           # Format code
make swagger       # Generate Swagger documentation
make migrate-up    # Apply database migrations
make migrate-down  # Rollback last migration
make docker-build  # Build Docker image
make docker-run    # Run application in Docker
make up            # Start all services with docker-compose
make down          # Stop all services
```

### Running Tests
```bash
# Run all tests
make test

# Run specific tests
go test -v -run TestCreateTag ./internal/api/v1/handler

# Run with race detector
go test -race ./...

# Generate coverage report
make test-coverage
```

### API Examples

#### Create a Tag
```bash
curl -X POST http://localhost:8080/api/v1/tags \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "name": "Google Analytics",
    "type": "analytics",
    "template_id": "ga-template",
    "config": {
      "tracking_id": "UA-123456-1"
    }
  }'
```

#### Get Tag by ID
```bash
curl -X GET http://localhost:8080/api/v1/tags/{tag_id} \
  -H "Authorization: Bearer $TOKEN"
```

#### Deploy Tag to Environment
```bash
curl -X POST http://localhost:8080/api/v1/tags/{tag_id}/deploy \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "environment": "production",
    "version": "v1.2.0"
  }'
```

## 📊 Architecture Overview

### Event Flow
```
Client Request → API Gateway → Tag Service → PostgreSQL
                     ↓              ↓
                Rate Limiter    Kafka Producer
                  (Redis)           ↓
                               Event Stream
                                    ↓
                          ┌─────────┴─────────┐
                          ↓                   ↓
                    Audit Logger        Tag Processor
                     (Consumer)          (Consumer)
```

### Kafka Topics
- `tags.events.created` - New tag creation events
- `tags.events.updated` - Tag update events
- `tags.events.deleted` - Tag deletion events
- `tags.events.deployed` - Tag deployment events
- `tags.commands.validate` - Validation commands
- `tags.events.dlq` - Dead letter queue

## 🔐 Security

- JWT-based authentication
- Rate limiting per user/endpoint
- Input validation and sanitization
- SQL injection protection
- XSS prevention
- CORS configuration

## 📈 Performance

- Handles 1000+ requests per second
- Sub-100ms response time for cached requests
- Horizontal scaling through Kafka partitions
- Connection pooling for database efficiency
- Redis caching for frequently accessed data

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Coding Standards
- Follow [Effective Go](https://go.dev/doc/effective_go) guidelines
- Use `gofmt` and `golangci-lint`
- Write tests for new features
- Update documentation as needed

## 📝 API Documentation

Full API documentation is available via Swagger UI when the service is running:
http://localhost:8080/swagger/index.html

## 🔍 Monitoring

### Health Checks
- `/health` - Basic health check
- `/health/ready` - Readiness probe
- `/health/live` - Liveness probe

### Metrics
Prometheus metrics are exposed at `/metrics`

## 🚀 Deployment

### Docker
```bash
docker build -t tag-management-api:latest .
docker run -p 8080:8080 tag-management-api:latest
```

### Kubernetes
```bash
kubectl apply -f deployments/k8s/
```

## 📞 Support

For questions and support, please open an issue in the GitHub repository.

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- OpenTag team for inspiration
- Go community for excellent tools and libraries
- Contributors and maintainers

---

**Note**: This is a demonstration project showcasing modern Go microservice architecture with event-driven design patterns.
