package dto

type LoginResponse struct {
	Token     string `json:"token"`
	FirstName string `json:"firstName"`
}
