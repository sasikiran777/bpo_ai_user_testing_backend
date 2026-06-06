package dto

import (
	testsmodel "ai_testing/internal/modules/tests/model"
	usersmodel "ai_testing/internal/modules/users/model"
	"encoding/json"
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

type SaveAudioAnswerResponse struct {
	ID        uuid.UUID `json:"id"`
	AudioPath string    `json:"audio_path"`
}

type SpeakingTopicResponse struct {
	ID                   uuid.UUID `json:"id"`
	TestSectionMappingID uuid.UUID `json:"test_section_mapping_id"`
	Topic                string    `json:"topic"`
}

type ReadingComprehensionResponse struct {
	ID                   uuid.UUID       `json:"id"`
	TestSectionMappingID uuid.UUID       `json:"test_section_mapping_id"`
	Title                string          `json:"title"`
	Passage              string          `json:"passage"`
	Questions            json.RawMessage `json:"questions"`
}

type UserTestSectionResultResponse struct {
	TestSectionResponse

	Questions           []string `json:"questions"`
	Answers             []string `json:"answers"`
	TestNotes           []string `json:"test_notes"`
	MarksObtained       int      `json:"marks_obtained"`
	AIFeedback          string   `json:"ai_feedback"`
	ChangedWindowsCount int      `json:"changed_windows_count"`
	HasGraded           bool     `json:"has_graded"`
}

type UserTestResultsResponse struct {
	UserTestMapping    UserTestMappingResponse         `json:"user_test_mapping"`
	Test               TestResponse                    `json:"test"`
	TotalMarksObtained int                             `json:"total_marks_obtained"`
	TotalMaxMarks      int                             `json:"total_max_marks"`
	Sections           []UserTestSectionResultResponse `json:"sections"`
	GradingCompleted   bool                            `json:"grading_completed"`
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

func ToSpeakingTopicResponse(t testsmodel.SpeakingTopic) SpeakingTopicResponse {
	return SpeakingTopicResponse{
		ID:                   t.ID,
		TestSectionMappingID: t.TestSectionMappingID,
		Topic:                t.Topic,
	}
}

func ToReadingComprehensionResponse(r testsmodel.ReadingComprehension) ReadingComprehensionResponse {
	q := r.Questions
	if q == nil {
		q = json.RawMessage("[]")
	}
	return ReadingComprehensionResponse{
		ID:                   r.ID,
		TestSectionMappingID: r.TestSectionMappingID,
		Title:                r.Title,
		Passage:              r.Passage,
		Questions:            q,
	}
}

func ToUserTestResultsResponse(
	test testsmodel.Test,
	mapping usersmodel.UserTestMapping,
	sections []testsmodel.TestSectionMapping,
	answersBySection map[uuid.UUID]usersmodel.UserQuestionMapping,
) UserTestResultsResponse {
	totalMax := 0
	totalObtained := 0
	outSections := make([]UserTestSectionResultResponse, 0, len(sections))
	for _, s := range sections {
		totalMax += s.MaxMarks
		a, ok := answersBySection[s.ID]
		if !ok {
			outSections = append(outSections, UserTestSectionResultResponse{
				TestSectionResponse: ToTestSectionResponse(s),
				Questions:           []string{},
				Answers:             []string{},
				TestNotes:           []string{},
				MarksObtained:       0,
				AIFeedback:          "",
				ChangedWindowsCount: 0,
				HasGraded:           false,
			})
			continue
		}
		totalObtained += a.MarksObtained
		outSections = append(outSections, UserTestSectionResultResponse{
			TestSectionResponse: ToTestSectionResponse(s),
			Questions:           a.Question,
			Answers:             a.UserAnswer,
			TestNotes:           a.TestNotes,
			MarksObtained:       a.MarksObtained,
			AIFeedback:          a.AIFeedback,
			ChangedWindowsCount: a.ChangedWindowsCount,
			HasGraded:           a.HasGraded,
		})
	}

	return UserTestResultsResponse{
		UserTestMapping:    ToUserTestMappingResponse(mapping),
		Test:               ToTestResponse(test),
		TotalMarksObtained: totalObtained,
		TotalMaxMarks:      totalMax,
		Sections:           outSections,
		GradingCompleted:   mapping.GradingCompleted,
	}
}
