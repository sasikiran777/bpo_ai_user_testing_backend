package dto

import (
	testsmodel "ai_testing/internal/modules/tests/model"
	usersmodel "ai_testing/internal/modules/users/model"
	"time"

	"github.com/google/uuid"
)

type TestSectionResponse struct {
	ID          uuid.UUID `json:"id"`
	TestID      uuid.UUID `json:"test_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	MaxMarks    int       `json:"max_marks"`
	MaxTime     int       `json:"max_time"`
	IsActive    bool      `json:"is_active"`
}

type TestResponse struct {
	ID          uuid.UUID             `json:"id"`
	Name        string                `json:"name"`
	Code        string                `json:"code"`
	Description string                `json:"description"`
	Instruction []string              `json:"instruction"`
	IsActive    bool                  `json:"is_active"`
	Sections    []TestSectionResponse `json:"sections"`
}

type TestWithUserStatusResponse struct {
	TestResponse

	Attempted         bool       `json:"attempted"`
	UserTestMappingID *uuid.UUID `json:"user_test_mapping_id"`
	Status            string     `json:"status"`
	ResetCount        int        `json:"reset_count"`
	CompletedAt       *time.Time `json:"completed_at"`
	GradingCompleted  bool       `json:"grading_completed"`
}

type UserTestMappingResponse struct {
	ID                   uuid.UUID  `json:"id"`
	UserID               uuid.UUID  `json:"user_id"`
	TestID               uuid.UUID  `json:"test_id"`
	Status               string     `json:"status"`
	ResetCount           int        `json:"reset_count"`
	StartedAt            *time.Time `json:"started_at"`
	CompletedAt          *time.Time `json:"completed_at"`
	DidSystemComplete    bool       `json:"did_system_complete"`
	MicroPhonePermission bool       `json:"micro_phone_permission"`
	GradingCompleted     bool       `json:"grading_completed"`
}

type SaveAnswersResponse struct {
	ID uuid.UUID `json:"id"`
}

func ToTestResponse(t testsmodel.Test) TestResponse {
	res := TestResponse{
		ID:          t.ID,
		Name:        t.Name,
		Code:        t.Code,
		Description: t.Description,
		Instruction: t.Instruction,
		IsActive:    t.IsActive,
		Sections:    []TestSectionResponse{},
	}

	if len(t.Sections) == 0 {
		return res
	}

	res.Sections = make([]TestSectionResponse, 0, len(t.Sections))
	for _, s := range t.Sections {
		res.Sections = append(res.Sections, ToTestSectionResponse(s))
	}

	return res
}

func ToTestWithUserStatusResponse(
	t testsmodel.Test,
	mapping *usersmodel.UserTestMapping,
) TestWithUserStatusResponse {
	resp := TestWithUserStatusResponse{
		TestResponse: ToTestResponse(t),
		Attempted:    false,
		Status:       "not_attempted",
		ResetCount:   0,
		CompletedAt:  nil,
	}

	if mapping == nil {
		return resp
	}

	resp.Attempted = true
	id := mapping.ID
	resp.UserTestMappingID = &id
	resp.Status = mapping.Status
	resp.ResetCount = mapping.ResetCount
	if !mapping.CompletedAt.IsZero() {
		tm := mapping.CompletedAt
		resp.CompletedAt = &tm
	}
	resp.GradingCompleted = mapping.GradingCompleted
	return resp
}

func ToTestSectionResponse(s testsmodel.TestSectionMapping) TestSectionResponse {
	return TestSectionResponse{
		ID:          s.ID,
		TestID:      s.TestID,
		Name:        s.Name,
		Description: s.Description,
		MaxMarks:    s.MaxMarks,
		MaxTime:     s.MaxTime,
		IsActive:    s.IsActive,
	}
}

func ToUserTestMappingResponse(m usersmodel.UserTestMapping) UserTestMappingResponse {
	resp := UserTestMappingResponse{
		ID:                   m.ID,
		UserID:               m.UserID,
		TestID:               m.TestID,
		Status:               m.Status,
		ResetCount:           m.ResetCount,
		DidSystemComplete:    m.DidSystemComplete,
		MicroPhonePermission: m.MicroPhonePermission,
		GradingCompleted:     m.GradingCompleted,
	}

	if !m.StartedAt.IsZero() {
		tm := m.StartedAt
		resp.StartedAt = &tm
	}
	if !m.CompletedAt.IsZero() {
		tm := m.CompletedAt
		resp.CompletedAt = &tm
	}

	return resp
}
