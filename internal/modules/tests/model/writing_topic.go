package model

import (
	sharedModel "ai_testing/internal/shared/models"

	"github.com/google/uuid"
)

type WritingTopic struct {
	sharedModel.BaseModel `bun:"table:writing_topics"`

	TestSectionMappingID uuid.UUID `bun:"test_section_mapping_id,type:uuid,notnull"`
	Topic                string    `bun:"topic,type:text,notnull"`
	IsActive             bool      `bun:"is_active,notnull,default:true"`
}
