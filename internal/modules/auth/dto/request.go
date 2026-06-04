package dto

import "ai_testing/internal/modules/users/model"

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type RegisterRequest struct {
	FirstName      string   `json:"first_name" binding:"required"`
	LastName       string   `json:"last_name" binding:"required"`
	Phone          string   `json:"phone" binding:"required"`
	Email          string   `json:"email" binding:"required,email"`
	Password       string   `json:"password" binding:"required,min=6"`
	TotalExpMonths int      `json:"total_exp_months" binding:"required,min=0"`
	Skills         []string `json:"skills" binding:"required,min=1"`
	PastJobTitle   string   `json:"past_job_title"`
	Company        string   `json:"company"`
}

func (r *RegisterRequest) ToUserModel() *model.User {
	return &model.User{
		Email:     r.Email,
		Password:  r.Password,
		FirstName: r.FirstName,
		LastName:  r.LastName,
	}
}

func (r *RegisterRequest) ToUserProfileModel() *model.UserProfile {
	return &model.UserProfile{
		UserID:         r.ToUserModel().ID,
		TotalExpMonths: r.TotalExpMonths,
		Skills:         r.Skills,
		PastJobTitle:   r.PastJobTitle,
		Company:        r.Company,
	}
}
