package model

import (
	"ai_testing/internal/shared/models"
)

type User struct {
	models.BaseModel `bun:"table:users"`

	FirstName             string `bun:"first_name,notnull"`
	MiddleName            string `bun:"middle_name,nullzero"`
	LastName              string `bun:"last_name,notnull"`
	CountryCode           string `bun:"country_code,notnull"`
	Phone                 string `bun:"phone,notnull"`
	Email                 string `bun:"email,notnull,unique"`
	Password              string `bun:"password,notnull"`
	TypeOfPositionDesired string `bun:"type_of_position_desired,notnull"`
	ExpInMonths           int    `bun:"exp_in_months,notnull,default:0"`
}
