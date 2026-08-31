package dto

import "net/http"

type RegisterUserResponse struct {
	Username      string
	SessionCookie *http.Cookie
}
