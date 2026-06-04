package helpers

import (
	"ai_testing/internal/shared/responses"

	"github.com/gin-gonic/gin"
)

func Error(
	c *gin.Context,
	status int,
	message string,
) {

	c.JSON(status, responses.ErrorResponse{
		Success: false,
		Message: message,
	})
}
