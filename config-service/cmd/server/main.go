package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	// Import generated swagger docs
	_ "github.com/company/config-service/docs/swagger"
	"github.com/company/config-service/internal/api/health"
	"github.com/company/config-service/internal/config"
	"github.com/company/config-service/internal/database"
	"github.com/company/config-service/internal/logger"
	"github.com/company/config-service/pkg/metrics"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

var (
	version   = "dev"
	buildTime = "unknown"
	gitCommit = "unknown"
)

// @title Config Service API
// @version 1.0
// @description A configuration management service for cloud-native applications
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.example.com/support
// @contact.email support@example.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	log := logger.New(logger.Config{
		Level:  cfg.Logger.Level,
		Format: cfg.Logger.Format,
	})
	logger.SetGlobal(log)

	log.Info().
		Str("version", version).
		Str("build_time", buildTime).
		Str("git_commit", gitCommit).
		Str("environment", cfg.Server.Environment).
		Msg("Starting Config Service")

	// Initialize database connection
	db, err := database.New(cfg.Database, log)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer db.Close()

	// Run database migrations
	migrationRunner, err := database.NewMigrationRunner(db, cfg.Database, log)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create migration runner")
	}
	defer migrationRunner.Close()

	if err := migrationRunner.Up(); err != nil {
		log.Fatal().Err(err).Msg("Failed to run database migrations")
	}

	// Initialize Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.GetRedisAddr(),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	defer redisClient.Close()

	// Test Redis connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to Redis")
	}

	// Initialize metrics
	metricsCollector := metrics.New()

	// Set Gin mode based on environment
	if cfg.Server.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize Gin router
	router := gin.New()

	// Add middleware
	router.Use(gin.Recovery())
	router.Use(metricsCollector.Middleware())
	router.Use(corsMiddleware())
	router.Use(requestIDMiddleware())
	router.Use(loggingMiddleware(log))

	// Health check endpoints (no versioning)
	healthHandler := health.New(db, redisClient, log, version)
	router.GET("/health", healthHandler.Health)
	router.GET("/ready", healthHandler.Readiness)
	router.GET("/live", healthHandler.Liveness)

	// Metrics endpoint
	if cfg.Metrics.Enabled {
		router.GET(cfg.Metrics.Path, gin.WrapH(promhttp.Handler()))
	}

	// Swagger documentation (only in non-production environments)
	if cfg.Server.Environment != "production" {
		router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		v1.GET("/ping", pingHandler)
		v1.GET("/environments", getEnvironments(db))
		v1.GET("/tags", getTags(db))
	}

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Start server in a goroutine
	go func() {
		log.Info().
			Str("host", cfg.Server.Host).
			Str("port", cfg.Server.Port).
			Msg("Starting HTTP server")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	// Start metrics updater in a goroutine
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				stats := db.Stats()
				metricsCollector.UpdateDBConnections(
					stats.OpenConnections,
					stats.Idle,
					stats.InUse,
				)
			case <-ctx.Done():
				return
			}
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down server...")

	// Create a context with timeout for graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Gracefully shutdown the server
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatal().Err(err).Msg("Server forced to shutdown")
	}

	log.Info().Msg("Server exited")
}

// corsMiddleware adds CORS headers
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// requestIDMiddleware adds a unique request ID to each request
func requestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}

		c.Header("X-Request-ID", requestID)
		c.Set("request_id", requestID)
		c.Next()
	}
}

// loggingMiddleware logs HTTP requests
func loggingMiddleware(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		requestID, _ := c.Get("request_id")

		// Process request
		c.Next()

		// Log request
		latency := time.Since(start)
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method
		userAgent := c.Request.UserAgent()

		if raw != "" {
			path = path + "?" + raw
		}

		log.InfoWithFields("HTTP request", map[string]interface{}{
			"request_id": requestID,
			"status":     statusCode,
			"latency":    latency.String(),
			"client_ip":  clientIP,
			"method":     method,
			"path":       path,
			"user_agent": userAgent,
			"body_size":  c.Writer.Size(),
		})
	}
}

// generateRequestID generates a unique request ID
func generateRequestID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// pingHandler godoc
// @Summary Ping endpoint
// @Description Returns pong with version and timestamp
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/ping [get]
func pingHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
		"version": version,
		"time":    time.Now().Format(time.RFC3339),
	})
}

// getEnvironments godoc
// @Summary Get all environments
// @Description Retrieve all available environments
// @Tags environments
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/environments [get]
func getEnvironments(db *database.Connection) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := db.DB.Query("SELECT id, name, slug, description, active, priority FROM environments ORDER BY priority DESC")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		var environments []gin.H
		for rows.Next() {
			var id int64
			var name, slug, description string
			var active bool
			var priority int

			if err := rows.Scan(&id, &name, &slug, &description, &active, &priority); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			environments = append(environments, gin.H{
				"id":          id,
				"name":        name,
				"slug":        slug,
				"description": description,
				"active":      active,
				"priority":    priority,
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"environments": environments,
			"total":        len(environments),
		})
	}
}

// getTags godoc
// @Summary Get all tags
// @Description Retrieve all available tags
// @Tags tags
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/tags [get]
func getTags(db *database.Connection) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := db.DB.Query("SELECT id, name, description, color FROM tags ORDER BY name")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		var tags []gin.H
		for rows.Next() {
			var id int64
			var name, description, color string

			if err := rows.Scan(&id, &name, &description, &color); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			tags = append(tags, gin.H{
				"id":          id,
				"name":        name,
				"description": description,
				"color":       color,
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"tags":  tags,
			"total": len(tags),
		})
	}
}
