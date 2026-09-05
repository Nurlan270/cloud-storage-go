package dto

type ErrorResponse struct {
	Message string `json:"message"`
}

type RegisterUserResponse struct {
	Username string `json:"username"`
}

type LoginUserResponse struct {
	Username string `json:"username"`
}

type UserMeResponse struct {
	Username string `json:"username"`
}
