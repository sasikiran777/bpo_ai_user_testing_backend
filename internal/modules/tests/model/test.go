package model

import sharedModel "ai_testing/internal/shared/models"

type Test struct {
	sharedModel.BaseModel `bun:"table:tests"`

	Name        string   `bun:"name,notnull"`
	Code        string   `bun:"code,notnull,unique"`
	Description string   `bun:"description"`
	Instruction []string `bun:"instruction,type:jsonb,notnull,default:'[]'::jsonb"`

	IsActive bool `bun:"is_active,notnull,default:false"`

	Sections []TestSectionMapping `bun:"-"`
}
