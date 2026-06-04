package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

func RequestLogger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		requestID, _ := c.Get(RequestIDKey)
		email, _ := c.Get("email")
		latency := time.Since(start)

		logger.Info(
			"http_request",
			"requestId", requestID,
			"email", email,
			"method", c.Request.Method,
			"path", c.FullPath(),
			"status", c.Writer.Status(),
			"latencyMs", float64(latency.Milliseconds()),
		)
	}
}
