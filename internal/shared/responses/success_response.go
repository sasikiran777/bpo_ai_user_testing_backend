package responses

type SuccessResponse struct {
	Success bool   `json:"success" example:"true"`
	Message string `json:"message" example:"Request successful"`
	Data    any    `json:"data"`
}
