package validator

import (
	"ai_testing/internal/shared/helpers"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

func AuthValidator[T any](key string, dto T) gin.HandlerFunc {
	return func(c *gin.Context) {
		var payload T

		if err := c.ShouldBindBodyWith(&payload, binding.JSON); err != nil {
			helpers.ValidationError(
				c,
				helpers.FormatValidationErrors(err),
			)
			c.Abort()
			return
		}

		c.Set(key, payload)
		c.Next()
	}
}
