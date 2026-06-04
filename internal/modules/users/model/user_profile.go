package model

import (
	"github.com/google/uuid"

	"ai_testing/internal/shared/models"
)

type UserProfile struct {
	models.BaseModel `bun:"table:user_profiles"`

	UserID         uuid.UUID `bun:"user_id,type:uuid,notnull,unique"`
	TotalExpMonths int       `bun:"total_exp_months,notnull,default:0"`
	Skills         []string  `bun:"skills,type:jsonb,notnull,default:'[]'::jsonb"`
	PastJobTitle   string    `bun:"past_job_title,nullzero"`
	Company        string    `bun:"company,nullzero"`
}
