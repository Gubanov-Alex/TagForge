package health

import (
	"context"
	"net/http"
	"time"

	"github.com/company/config-service/internal/database"
	"github.com/company/config-service/internal/logger"
	"github.com/company/config-service/internal/model"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// Handler handles health check endpoints
type Handler struct {
	db      *database.Connection
	redis   *redis.Client
	logger  *logger.Logger
	version string
}

// New creates a new health handler
func New(db *database.Connection, redis *redis.Client, log *logger.Logger, version string) *Handler {
	return &Handler{
		db:      db,
		redis:   redis,
		logger:  log,
		version: version,
	}
}

// Health godoc
// @Summary Health check endpoint
// @Description Returns the health status of the service and its dependencies
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} model.HealthResponse
// @Failure 503 {object} model.ErrorResponse
// @Router /health [get]
func (h *Handler) Health(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	services := make(map[string]model.ServiceHealthInfo)
	overall := "healthy"

	// Check database
	dbHealth := h.checkDatabase(ctx)
	services["database"] = dbHealth
	if dbHealth.Status != "healthy" {
		overall = "unhealthy"
	}

	// Check Redis
	redisHealth := h.checkRedis(ctx)
	services["redis"] = redisHealth
	if redisHealth.Status != "healthy" {
		overall = "unhealthy"
	}

	response := model.HealthResponse{
		Status:   overall,
		Version:  h.version,
		Services: services,
	}

	if overall == "healthy" {
		c.JSON(http.StatusOK, response)
	} else {
		c.JSON(http.StatusServiceUnavailable, response)
	}
}

// Readiness godoc
// @Summary Readiness check endpoint
// @Description Returns readiness status for Kubernetes readiness probe
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} model.SuccessResponse
// @Failure 503 {object} model.ErrorResponse
// @Router /ready [get]
func (h *Handler) Readiness(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	// Check if database is ready
	if err := h.db.HealthCheck(); err != nil {
		h.logger.Error().Err(err).Msg("Database readiness check failed")
		c.JSON(http.StatusServiceUnavailable, model.ErrorResponse{
			Error:   "service_not_ready",
			Message: "Database is not ready",
		})
		return
	}

	// Check if Redis is ready
	if err := h.redis.Ping(ctx).Err(); err != nil {
		h.logger.Error().Err(err).Msg("Redis readiness check failed")
		c.JSON(http.StatusServiceUnavailable, model.ErrorResponse{
			Error:   "service_not_ready",
			Message: "Redis is not ready",
		})
		return
	}

	c.JSON(http.StatusOK, model.SuccessResponse{
		Message: "Service is ready",
	})
}

// Liveness godoc
// @Summary Liveness check endpoint
// @Description Returns liveness status for Kubernetes liveness probe
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} model.SuccessResponse
// @Router /live [get]
func (h *Handler) Liveness(c *gin.Context) {
	c.JSON(http.StatusOK, model.SuccessResponse{
		Message: "Service is alive",
	})
}

func (h *Handler) checkDatabase(ctx context.Context) model.ServiceHealthInfo {
	start := time.Now()

	err := h.db.HealthCheck()
	latency := time.Since(start)

	info := model.ServiceHealthInfo{
		LastCheck: time.Now().Format(time.RFC3339),
		Latency:   latency.String(),
	}

	if err != nil {
		info.Status = "unhealthy"
		info.Message = err.Error()
	} else {
		info.Status = "healthy"
		info.Message = "Database connection is healthy"
	}

	return info
}

func (h *Handler) checkRedis(ctx context.Context) model.ServiceHealthInfo {
	start := time.Now()

	err := h.redis.Ping(ctx).Err()
	latency := time.Since(start)

	info := model.ServiceHealthInfo{
		LastCheck: time.Now().Format(time.RFC3339),
		Latency:   latency.String(),
	}

	if err != nil {
		info.Status = "unhealthy"
		info.Message = err.Error()
	} else {
		info.Status = "healthy"
		info.Message = "Redis connection is healthy"
	}

	return info
}
