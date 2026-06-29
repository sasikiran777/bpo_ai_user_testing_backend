package model

import (
	"ai_testing/internal/shared/models"
	"time"

	"github.com/google/uuid"
)

type UserTestMapping struct {
	models.BaseModel `bun:"table:user_test_mappings"`

	UserID               uuid.UUID `bun:"user_id,type:uuid,notnull"`
	TestID               uuid.UUID `bun:"test_id,type:uuid,notnull"`
	Status               string    `bun:"status,default:'initialized'"`
	ResetCount           int       `bun:"reset_count,notnull,default:0"`
	StartedAt            time.Time `bun:"started_at,nullzero"`
	CompletedAt          time.Time `bun:"completed_at,nullzero"`
	DidSystemComplete    bool      `bun:"did_system_complete,default:false"`
	MicroPhonePermission bool      `bun:"micro_phone_permission,default:true"`
	GradingCompleted     bool      `bun:"grading_completed,notnull,default:false"`
}
