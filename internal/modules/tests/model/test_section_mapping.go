package model

import (
	sharedModel "ai_testing/internal/shared/models"

	"github.com/google/uuid"
)

type TestSectionMapping struct {
	sharedModel.BaseModel `bun:"table:test_section_mappings"`

	TestID      uuid.UUID `bun:"test_id,type:uuid,notnull"`
	Name        string    `bun:"name,notnull"`
	Description string    `bun:"description"`
	MaxMarks    int       `bun:"max_marks,notnull"`
	MaxTime     int       `bun:"max_time,notnull"`
	IsActive    bool      `bun:"is_active,notnull,default:true"`
}
