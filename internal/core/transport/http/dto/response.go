package dto

type ErrorResponse struct {
	Message string `json:"message"`
}

type RegisterUserResponse struct {
	Username string `json:"username"`
}
