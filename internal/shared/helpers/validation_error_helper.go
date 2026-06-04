package helpers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func ValidationError(
	c *gin.Context,
	errors map[string]string,
) {

	c.JSON(
		http.StatusUnprocessableEntity,
		gin.H{
			"success": false,
			"message": "Validation failed",
			"errors":  errors,
		},
	)
}
