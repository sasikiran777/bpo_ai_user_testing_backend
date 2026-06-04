package routes

import (
	"ai_testing/internal/modules/auth/dto"
	authHandler "ai_testing/internal/modules/auth/handler"
	authValidator "ai_testing/internal/modules/auth/validator"

	"ai_testing/internal/modules/auth/constants"

	"github.com/gin-gonic/gin"
)

func RegisterAuthRoutes(
	router *gin.RouterGroup,
	h *authHandler.AuthHandler,
) {

	loginValidator := authValidator.AuthValidator(constants.LoginPayloadKey, dto.LoginRequest{})
	router.POST("/login", loginValidator, h.Login)

	registerValidator := authValidator.AuthValidator(constants.RegisterPayloadKey, dto.RegisterRequest{})
	router.POST("/register", registerValidator, h.Register)
}
