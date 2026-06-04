package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func CORS(allowedOrigins []string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(allowedOrigins))
	for _, o := range allowedOrigins {
		allowed[o] = struct{}{}
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" {
			if _, ok := allowed[origin]; ok {
				c.Header("Access-Control-Allow-Origin", origin)
				c.Header("Vary", "Origin")
				c.Header("Access-Control-Allow-Credentials", "true")
				c.Header("Access-Control-Expose-Headers", "X-Request-Id")
			}
		}

		if c.Request.Method == http.MethodOptions {
			c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")

			reqHeaders := strings.TrimSpace(c.GetHeader("Access-Control-Request-Headers"))
			if reqHeaders == "" {
				reqHeaders = "Authorization,Content-Type,Accept,X-Request-Id"
			}
			c.Header("Access-Control-Allow-Headers", reqHeaders)
			c.Header("Access-Control-Max-Age", "86400")

			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

