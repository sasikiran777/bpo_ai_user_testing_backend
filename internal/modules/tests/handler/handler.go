package handler

import (
	"database/sql"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

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

func getUserID(c *gin.Context) (uuid.UUID, bool) {
	userIDRaw, ok := c.Get("user_id")
	if !ok {
		helpers.Error(c, http.StatusUnauthorized, "Unauthorized")
		return uuid.Nil, false
	}

	userID, ok := userIDRaw.(uuid.UUID)
	if !ok {
		helpers.Error(c, http.StatusUnauthorized, "Unauthorized")
		return uuid.Nil, false
	}

	return userID, true
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
	userID, ok := getUserID(c)
	if !ok {
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
	userID, ok := getUserID(c)
	if !ok {
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
	userID, ok := getUserID(c)
	if !ok {
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

func (h *Handler) GetUserTestResults(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	idRaw := c.Param("userTestMappingId")
	id, err := uuid.Parse(idRaw)
	if err != nil {
		helpers.Error(c, http.StatusBadRequest, "Invalid user_test_mapping_id")
		return
	}

	resp, err := h.Service.GetUserTestResults(c.Request.Context(), userID, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			helpers.Error(c, http.StatusNotFound, "User test not found")
			return
		}
		helpers.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	helpers.Success(c, http.StatusOK, "User test results fetched successfully", resp)
}

func (h *Handler) GetUserTestAudio(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	mappingIDRaw := c.Param("userTestMappingId")
	mappingID, err := uuid.Parse(mappingIDRaw)
	if err != nil {
		helpers.Error(c, http.StatusBadRequest, "Invalid user_test_mapping_id")
		return
	}

	sectionIDRaw := c.Param("sectionId")
	sectionID, err := uuid.Parse(sectionIDRaw)
	if err != nil {
		helpers.Error(c, http.StatusBadRequest, "Invalid section_id")
		return
	}

	absPath, err := h.Service.GetUserAudioFilePath(c.Request.Context(), userID, mappingID, sectionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, os.ErrNotExist) {
			helpers.Error(c, http.StatusNotFound, "Audio not found")
			return
		}
		helpers.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	c.File(absPath)
}

func (h *Handler) SaveAnswers(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
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

func (h *Handler) SaveAudioAnswer(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	userTestMappingIDRaw := c.PostForm("user_test_mapping_id")
	if userTestMappingIDRaw == "" {
		helpers.Error(c, http.StatusBadRequest, "user_test_mapping_id is required")
		return
	}
	userTestMappingID, err := uuid.Parse(userTestMappingIDRaw)
	if err != nil {
		helpers.Error(c, http.StatusBadRequest, "Invalid user_test_mapping_id")
		return
	}

	sectionIDRaw := c.PostForm("section_id")
	if sectionIDRaw == "" {
		helpers.Error(c, http.StatusBadRequest, "section_id is required")
		return
	}
	sectionID, err := uuid.Parse(sectionIDRaw)
	if err != nil {
		helpers.Error(c, http.StatusBadRequest, "Invalid section_id")
		return
	}

	question := c.PostForm("question")
	if question == "" {
		helpers.Error(c, http.StatusBadRequest, "question is required")
		return
	}

	changedWindowsCount := 0
	if v := c.PostForm("changed_windows_count"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			helpers.Error(c, http.StatusBadRequest, "invalid changed_windows_count")
			return
		}
		changedWindowsCount = n
	}

	file, err := c.FormFile("audio")
	if err != nil {
		helpers.Error(c, http.StatusBadRequest, "audio file is required")
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext == "" {
		ext = ".webm"
	}

	allowed := map[string]struct{}{
		".wav":  {},
		".mp3":  {},
		".m4a":  {},
		".webm": {},
		".ogg":  {},
		".aac":  {},
		".flac": {},
	}
	if _, ok := allowed[ext]; !ok {
		helpers.Error(c, http.StatusBadRequest, "invalid audio file type")
		return
	}

	fileName := uuid.NewString() + ext
	relPath := filepath.Join(
		"storage",
		"audio",
		userID.String(),
		userTestMappingID.String(),
		sectionID.String(),
		fileName,
	)
	absDir := filepath.Dir(relPath)
	if err = os.MkdirAll(absDir, 0o755); err != nil {
		helpers.Error(c, http.StatusInternalServerError, "failed to create upload directory")
		return
	}

	if err = c.SaveUploadedFile(file, relPath); err != nil {
		helpers.Error(c, http.StatusInternalServerError, "failed to save audio file")
		return
	}

	audioPath := filepath.ToSlash(relPath)
	row, err := h.Service.SaveAudioAnswer(
		c.Request.Context(),
		userID,
		userTestMappingID,
		sectionID,
		question,
		changedWindowsCount,
		audioPath,
	)
	if err != nil {
		_ = os.Remove(relPath)
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
		"Audio answer saved successfully",
		dto.SaveAudioAnswerResponse{ID: row.ID, AudioPath: audioPath},
	)
}

func (h *Handler) DropUserTest(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	idRaw := c.Param("userTestMappingId")
	id, err := uuid.Parse(idRaw)
	if err != nil {
		helpers.Error(c, http.StatusBadRequest, "Invalid user_test_mapping_id")
		return
	}

	if err := h.Service.DropUserTest(c.Request.Context(), userID, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			helpers.Error(c, http.StatusNotFound, "User test not found")
			return
		}
		helpers.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	helpers.Success(c, http.StatusOK, "User test dropped successfully", gin.H{"status": "dropped"})
}

func (h *Handler) GetRandomSpeakingTopic(c *gin.Context) {
	sectionIDRaw := c.Param("sectionId")
	sectionID, err := uuid.Parse(sectionIDRaw)
	if err != nil {
		helpers.Error(c, http.StatusBadRequest, "Invalid section id")
		return
	}

	topic, err := h.Service.GetRandomSpeakingTopic(c.Request.Context(), sectionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			helpers.Error(c, http.StatusNotFound, "No topic found")
			return
		}
		helpers.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	helpers.Success(c, http.StatusOK, "Speaking topic fetched successfully", dto.ToSpeakingTopicResponse(*topic))
}

func (h *Handler) GetRandomReadingComprehension(c *gin.Context) {
	sectionIDRaw := c.Param("sectionId")
	sectionID, err := uuid.Parse(sectionIDRaw)
	if err != nil {
		helpers.Error(c, http.StatusBadRequest, "Invalid section id")
		return
	}

	rc, err := h.Service.GetRandomReadingComprehension(c.Request.Context(), sectionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			helpers.Error(c, http.StatusNotFound, "No comprehension found")
			return
		}
		helpers.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	helpers.Success(c, http.StatusOK, "Reading comprehension fetched successfully", dto.ToReadingComprehensionResponse(*rc))
}
