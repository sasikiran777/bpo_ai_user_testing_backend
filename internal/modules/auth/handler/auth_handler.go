package handler

import (
	"net/http"

	"ai_testing/internal/modules/auth/constants"
	"ai_testing/internal/modules/auth/dto"
	authService "ai_testing/internal/modules/auth/service"
	"ai_testing/internal/shared/helpers"
	sharedvalidator "ai_testing/internal/shared/validator"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	Service *authService.AuthService
}

func New(
	service *authService.AuthService,
) *AuthHandler {
	return &AuthHandler{
		Service: service,
	}
}

func (h *AuthHandler) Login(c *gin.Context) {
	payload, ok := sharedvalidator.GetValidatedPayload[dto.LoginRequest](
		c,
		constants.LoginPayloadKey,
		"login payload not found",
		"invalid login payload",
	)
	if !ok {
		return
	}

	data, err := h.Service.Login(
		c.Request.Context(),
		*payload,
	)

	if err != nil {

		helpers.Error(
			c,
			http.StatusUnauthorized,
			err.Error(),
		)

		return
	}

	helpers.Success(
		c,
		http.StatusOK,
		"Login successful",
		data,
	)
}

func (h *AuthHandler) Register(c *gin.Context) {
	payload, ok := sharedvalidator.GetValidatedPayload[dto.RegisterRequest](
		c,
		constants.RegisterPayloadKey,
		"register payload not found",
		"invalid register payload",
	)
	if !ok {
		return
	}

	data, err := h.Service.Register(
		c.Request.Context(),
		*payload,
	)

	if err != nil {

		helpers.Error(
			c,
			http.StatusBadRequest,
			err.Error(),
		)

		return
	}

	helpers.Success(
		c,
		http.StatusOK,
		"Register successful",
		data,
	)
}
