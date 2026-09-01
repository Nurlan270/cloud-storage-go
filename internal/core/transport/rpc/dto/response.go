package dto

import (
	"net/http"

	"github.com/Nurlan270/cloud-storage-go/internal/core/models"
)

type RegisterUserResponse struct {
	Username      string
	SessionCookie *http.Cookie
}

type LoginUserResponse struct {
	Username      string
	SessionCookie *http.Cookie
}

type LogoutUserResponse struct {
	SessionCookie *http.Cookie
}

type GetUserFromSIDResponse struct {
	User *models.User
}
