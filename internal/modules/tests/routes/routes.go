package routes

import (
	"ai_testing/internal/modules/tests/handler"

	"github.com/gin-gonic/gin"
)

func RegisterTestRoutes(r *gin.RouterGroup, h *handler.Handler) {
	r.GET("", h.List)
	r.GET("/my-tests", h.ListForUser)
	r.POST("/my-tests/save-answers", h.SaveAnswers)
	r.POST("/my-tests/save-audio", h.SaveAudioAnswer)
	r.POST("/my-tests/:testId", h.CreateUserTestMapping)
	r.GET("/my-tests/:testId", h.GetUserTestStatus)
	r.GET("/:testId", h.GetSectionsByTestID)
}
