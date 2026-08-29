package dto

type RegisterUserRequest struct {
	Username string `json:"username" validate:"required,min=3,max=32,username"`
	Password string `json:"password" validate:"required,min=6"`
}
