package helpers

import (
	"ai_testing/internal/shared/responses"

	"github.com/gin-gonic/gin"
)

func Success(
	c *gin.Context,
	status int,
	message string,
	data any,
) {

	c.JSON(status, responses.SuccessResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}
