package model

import (
	"encoding/json"

	sharedModel "ai_testing/internal/shared/models"

	"github.com/google/uuid"
)

type ReadingComprehension struct {
	sharedModel.BaseModel `bun:"table:reading_comprehensions"`

	TestSectionMappingID uuid.UUID       `bun:"test_section_mapping_id,type:uuid,notnull"`
	Title                string          `bun:"title,type:text,notnull"`
	Passage              string          `bun:"passage,type:text,notnull"`
	Questions            json.RawMessage `bun:"questions,type:jsonb,notnull,default:'[]'::jsonb"`
	IsActive             bool            `bun:"is_active,notnull,default:true"`
}

