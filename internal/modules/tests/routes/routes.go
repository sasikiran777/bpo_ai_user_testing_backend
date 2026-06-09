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
	r.GET("/my-tests/results/:userTestMappingId", h.GetUserTestResults)
	r.GET("/my-tests/audio/:userTestMappingId/sections/:sectionId", h.GetUserTestAudio)
	r.POST("/my-tests/:testId", h.CreateUserTestMapping)
	r.GET("/my-tests/:testId", h.GetUserTestStatus)
	r.PATCH("/my-tests/:userTestMappingId/drop", h.DropUserTest)
	r.GET("/sections/:sectionId/speaking-topic", h.GetRandomSpeakingTopic)
	r.GET("/sections/:sectionId/writing-topic", h.GetRandomWritingTopic)
	r.GET("/sections/:sectionId/reading", h.GetRandomReadingComprehension)
	r.GET("/:testId", h.GetSectionsByTestID)
}
