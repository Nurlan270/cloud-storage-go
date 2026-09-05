package dto

type RegisterUserRequest struct {
	Username string `json:"username" validate:"required,min=3,max=30,username"`
	Password string `json:"password" validate:"required,min=6,max=70"`
}

type LoginUserRequest struct {
	Username string `json:"username" validate:"required,min=3,max=30,username"`
	Password string `json:"password" validate:"required,min=6,max=70"`
}
