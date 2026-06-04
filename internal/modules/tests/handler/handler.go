package handler

import (
	"database/sql"
	"errors"
	"net/http"

	"ai_testing/internal/modules/tests/dto"
	"ai_testing/internal/modules/tests/service"
	"ai_testing/internal/shared/helpers"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	Service *service.Service
}

func New(service *service.Service) *Handler {
	return &Handler{Service: service}
}

func (h *Handler) List(c *gin.Context) {
	tests, err := h.Service.List(c.Request.Context())
	if err != nil {
		helpers.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	resp := make([]dto.TestResponse, 0, len(tests))
	for _, t := range tests {
		resp = append(resp, dto.ToTestResponse(t))
	}

	helpers.Success(c, http.StatusOK, "Tests fetched successfully", resp)
}

func (h *Handler) ListForUser(c *gin.Context) {
	userIDRaw, ok := c.Get("user_id")
	if !ok {
		helpers.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	userID, ok := userIDRaw.(uuid.UUID)
	if !ok {
		helpers.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	resp, err := h.Service.ListForUser(c.Request.Context(), userID)
	if err != nil {
		helpers.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	helpers.Success(c, http.StatusOK, "User tests fetched successfully", resp)
}

func (h *Handler) GetSectionsByTestID(c *gin.Context) {
	testIDRaw := c.Param("testId")
	testID, err := uuid.Parse(testIDRaw)
	if err != nil {
		helpers.Error(c, http.StatusBadRequest, "Invalid test id")
		return
	}

	test, err := h.Service.GetSectionsByTestID(c.Request.Context(), testID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			helpers.Error(c, http.StatusNotFound, "Test not found")
			return
		}
		helpers.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	helpers.Success(c, http.StatusOK, "Test fetched successfully", dto.ToTestResponse(*test))
}

func (h *Handler) CreateUserTestMapping(c *gin.Context) {
	userIDRaw, ok := c.Get("user_id")
	if !ok {
		helpers.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	userID, ok := userIDRaw.(uuid.UUID)
	if !ok {
		helpers.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	testIDRaw := c.Param("testId")
	testID, err := uuid.Parse(testIDRaw)
	if err != nil {
		helpers.Error(c, http.StatusBadRequest, "Invalid test id")
		return
	}

	var payload dto.CreateUserTestMappingRequest
	if err = c.ShouldBindJSON(&payload); err != nil {
		helpers.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	mapping, err := h.Service.CreateUserTestMapping(
		c.Request.Context(),
		userID,
		testID,
		payload.MicroPhonePermission,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			helpers.Error(c, http.StatusNotFound, "Test not found")
			return
		}
		helpers.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	helpers.Success(c, http.StatusOK, "User test created successfully", dto.ToUserTestMappingResponse(*mapping))
}

func (h *Handler) GetUserTestStatus(c *gin.Context) {
	userIDRaw, ok := c.Get("user_id")
	if !ok {
		helpers.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	userID, ok := userIDRaw.(uuid.UUID)
	if !ok {
		helpers.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	testIDRaw := c.Param("testId")
	testID, err := uuid.Parse(testIDRaw)
	if err != nil {
		helpers.Error(c, http.StatusBadRequest, "Invalid test id")
		return
	}

	mapping, err := h.Service.GetUserTestStatus(c.Request.Context(), userID, testID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			helpers.Error(c, http.StatusNotFound, "User test not found")
			return
		}
		helpers.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	helpers.Success(c, http.StatusOK, "User test status fetched successfully", dto.ToUserTestMappingResponse(*mapping))
}

func (h *Handler) SaveAnswers(c *gin.Context) {
	userIDRaw, ok := c.Get("user_id")
	if !ok {
		helpers.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	userID, ok := userIDRaw.(uuid.UUID)
	if !ok {
		helpers.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var payload dto.SaveAnswersRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		helpers.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	if len(payload.Questions) != len(payload.Answers) {
		helpers.Error(c, http.StatusBadRequest, "questions and answers length mismatch")
		return
	}

	if len(payload.Questions) == 0 {
		helpers.Error(c, http.StatusBadRequest, "questions cannot be empty")
		return
	}

	row, err := h.Service.SaveAnswers(c.Request.Context(), userID, payload)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			helpers.Error(c, http.StatusNotFound, "Not found")
			return
		}
		helpers.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	helpers.Success(
		c,
		http.StatusOK,
		"Answers saved successfully",
		dto.SaveAnswersResponse{ID: row.ID},
	)
}
