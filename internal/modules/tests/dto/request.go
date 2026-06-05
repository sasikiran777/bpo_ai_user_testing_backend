package dto

import "github.com/google/uuid"

type CreateUserTestMappingRequest struct {
	MicroPhonePermission bool `json:"micro_phone_permission" binding:"required"`
}

type SaveAnswersRequest struct {
	UserTestMappingID   uuid.UUID `json:"user_test_mapping_id" binding:"required"`
	SectionID           uuid.UUID `json:"section_id" binding:"required"`
	Questions           []string  `json:"questions" binding:"required"`
	Answers             []string  `json:"answers" binding:"required"`
	TestNotes           []string  `json:"test_notes"`
	ChangedWindowsCount int       `json:"changed_windows_count"`
}

type SaveAudioAnswerRequest struct {
	UserTestMappingID   string `form:"user_test_mapping_id"`
	SectionID           string `form:"section_id"`
	Question            string `form:"question"`
	ChangedWindowsCount int    `form:"changed_windows_count"`
}
