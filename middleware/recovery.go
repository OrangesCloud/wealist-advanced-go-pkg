package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Recovery returns a middleware that recovers from panics
func Recovery(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Get request ID if available
				requestID := GetRequestID(c)

				// Log the panic
				logger.Error("Panic recovered",
					zap.String("request_id", requestID),
					zap.Any("error", err),
					zap.String("stack", string(debug.Stack())),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
				)

				// Return 500 error
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": map[string]interface{}{
						"code":    "INTERNAL_ERROR",
						"message": "Internal server error",
					},
					"requestId": requestID,
				})
			}
		}()

		c.Next()
	}
}
