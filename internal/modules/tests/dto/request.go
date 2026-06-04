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
	ChangedWindowsCount int       `json:"changed_windows_count"`
}
