package model

import (
	"ai_testing/internal/shared/models"

	"github.com/google/uuid"
)

type UserQuestionMapping struct {
	models.BaseModel `bun:"table:user_question_mappings"`

	UserTestMappingID    uuid.UUID `bun:"user_test_mapping_id,type:uuid,notnull"`
	TestSectionMappingID uuid.UUID `bun:"test_section_mapping_id,type:uuid,notnull"`
	Question             []string  `bun:"question,type:jsonb,notnull,default:'[]'::jsonb"`
	UserAnswer           []string  `bun:"user_answer,type:jsonb,notnull,default:'[]'::jsonb"`
	TestNotes            []string  `bun:"test_notes,type:jsonb,notnull,default:'[]'::jsonb"`
	MarksObtained        int       `bun:"marks_obtained"`
	AIFeedback           string    `bun:"ai_feedback,type:text"`
	ChangedWindowsCount  int       `bun:"changed_windows_count"`
	HasGraded            bool      `bun:"has_graded,notnull,default:false"`
}
