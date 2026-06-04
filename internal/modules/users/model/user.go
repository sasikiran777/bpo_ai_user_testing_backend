package model

import (
	"ai_testing/internal/shared/models"
)

type User struct {
	models.BaseModel `bun:"table:users"`

	FirstName string `bun:"first_name,notnull"`
	LastName  string `bun:"last_name,notnull"`
	Email     string `bun:"email,notnull,unique"`
	Phone     string `bun:"phone,nullzero"`
	Password  string `bun:"password,notnull"`
}
