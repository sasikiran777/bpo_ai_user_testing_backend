package validator

import (
	"ai_testing/internal/shared/helpers"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetValidatedPayload[T any](c *gin.Context, key, notFoundMsg, invalidMsg string) (*T, bool) {
	v, ok := c.Get(key)
	if !ok {
		helpers.Error(c, http.StatusBadRequest, notFoundMsg)
		return nil, false
	}
	payload, ok := v.(T)
	if !ok {
		helpers.Error(c, http.StatusBadRequest, invalidMsg)
		return nil, false
	}
	return &payload, true
}
