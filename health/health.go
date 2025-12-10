// Package health provides common health check handlers for all services.
package health

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Status represents health check status
type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusUnhealthy Status = "unhealthy"
	StatusDegraded  Status = "degraded"
)

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    Status                   `json:"status"`
	Timestamp string                   `json:"timestamp"`
	Checks    map[string]ComponentCheck `json:"checks,omitempty"`
}

// ComponentCheck represents a single component's health
type ComponentCheck struct {
	Status  Status `json:"status"`
	Message string `json:"message,omitempty"`
	Latency string `json:"latency,omitempty"`
}

// Checker interface for health check components
type Checker interface {
	Name() string
	Check(ctx context.Context) ComponentCheck
}

// Handler holds health check dependencies
type Handler struct {
	checkers []Checker
	mu       sync.RWMutex
}

// NewHandler creates a new health handler
func NewHandler() *Handler {
	return &Handler{
		checkers: make([]Checker, 0),
	}
}

// AddChecker adds a health checker
func (h *Handler) AddChecker(checker Checker) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.checkers = append(h.checkers, checker)
}

// HealthHandler returns the /health endpoint handler (liveness probe)
func (h *Handler) HealthHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, HealthResponse{
			Status:    StatusHealthy,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		})
	}
}

// ReadyHandler returns the /ready endpoint handler (readiness probe)
func (h *Handler) ReadyHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		h.mu.RLock()
		checkers := h.checkers
		h.mu.RUnlock()

		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		checks := make(map[string]ComponentCheck)
		overallStatus := StatusHealthy

		for _, checker := range checkers {
			check := checker.Check(ctx)
			checks[checker.Name()] = check

			if check.Status == StatusUnhealthy {
				overallStatus = StatusUnhealthy
			} else if check.Status == StatusDegraded && overallStatus == StatusHealthy {
				overallStatus = StatusDegraded
			}
		}

		statusCode := http.StatusOK
		if overallStatus == StatusUnhealthy {
			statusCode = http.StatusServiceUnavailable
		}

		c.JSON(statusCode, HealthResponse{
			Status:    overallStatus,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Checks:    checks,
		})
	}
}

// RegisterRoutes registers health check routes
func (h *Handler) RegisterRoutes(router *gin.Engine) {
	router.GET("/health", h.HealthHandler())
	router.GET("/ready", h.ReadyHandler())
}

// DatabaseChecker checks database connectivity
type DatabaseChecker struct {
	db *gorm.DB
}

// NewDatabaseChecker creates a new database health checker
func NewDatabaseChecker(db *gorm.DB) *DatabaseChecker {
	return &DatabaseChecker{db: db}
}

// Name returns the checker name
func (d *DatabaseChecker) Name() string {
	return "database"
}

// Check performs the database health check
func (d *DatabaseChecker) Check(ctx context.Context) ComponentCheck {
	start := time.Now()

	sqlDB, err := d.db.DB()
	if err != nil {
		return ComponentCheck{
			Status:  StatusUnhealthy,
			Message: "Failed to get database connection: " + err.Error(),
		}
	}

	if err := sqlDB.PingContext(ctx); err != nil {
		return ComponentCheck{
			Status:  StatusUnhealthy,
			Message: "Database ping failed: " + err.Error(),
		}
	}

	return ComponentCheck{
		Status:  StatusHealthy,
		Message: "Database connection OK",
		Latency: time.Since(start).String(),
	}
}

// RedisChecker interface for Redis health check
type RedisChecker struct {
	pingFunc func(ctx context.Context) error
}

// NewRedisChecker creates a new Redis health checker
func NewRedisChecker(pingFunc func(ctx context.Context) error) *RedisChecker {
	return &RedisChecker{pingFunc: pingFunc}
}

// Name returns the checker name
func (r *RedisChecker) Name() string {
	return "redis"
}

// Check performs the Redis health check
func (r *RedisChecker) Check(ctx context.Context) ComponentCheck {
	start := time.Now()

	if err := r.pingFunc(ctx); err != nil {
		return ComponentCheck{
			Status:  StatusUnhealthy,
			Message: "Redis ping failed: " + err.Error(),
		}
	}

	return ComponentCheck{
		Status:  StatusHealthy,
		Message: "Redis connection OK",
		Latency: time.Since(start).String(),
	}
}

// SimpleHealthHandler returns simple /health handler without dependencies
func SimpleHealthHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	}
}

// SimpleReadyHandler returns simple /ready handler without dependencies
func SimpleReadyHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ready",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	}
}
