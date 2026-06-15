package dto

import "ai_testing/internal/modules/users/model"

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type RegisterRequest struct {
	FirstName             string `json:"first_name" binding:"required"`
	MiddleName            string `json:"middle_name"`
	LastName              string `json:"last_name" binding:"required"`
	CountryCode           string `json:"country_code" binding:"required"`
	Phone                 string `json:"phone" binding:"required"`
	Email                 string `json:"email" binding:"required,email"`
	Password              string `json:"password" binding:"required,min=6"`
	TypeOfPositionDesired string `json:"type_of_position_desired" binding:"required"`
	ExpInMonths           int    `json:"exp_in_months" binding:"required,min=0"`
}

func (r *RegisterRequest) ToUserModel() *model.User {
	return &model.User{
		FirstName:             r.FirstName,
		MiddleName:            r.MiddleName,
		LastName:              r.LastName,
		CountryCode:           r.CountryCode,
		Phone:                 r.Phone,
		Email:                 r.Email,
		Password:              r.Password,
		TypeOfPositionDesired: r.TypeOfPositionDesired,
		ExpInMonths:           r.ExpInMonths,
	}
}
