package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Nurlan270/cloud-storage-go/internal/cloud_storage/config"
	"github.com/Nurlan270/cloud-storage-go/internal/cloud_storage/service"
	"github.com/Nurlan270/cloud-storage-go/internal/cloud_storage/transport/http/dto/response"
	"github.com/Nurlan270/cloud-storage-go/internal/cloud_storage/transport/http/message"
	"github.com/Nurlan270/cloud-storage-go/internal/cloud_storage/transport/http/render"
	errs "github.com/Nurlan270/cloud-storage-go/internal/core/errors"
	corehttp "github.com/Nurlan270/cloud-storage-go/internal/core/transport/http"
	coredto "github.com/Nurlan270/cloud-storage-go/internal/core/transport/http/dto"
)

type AuthHandler interface {
	Register(w http.ResponseWriter, r *http.Request)
	Login(w http.ResponseWriter, r *http.Request)
	Logout(w http.ResponseWriter, r *http.Request)
}

type authHandler struct {
	appConf *config.Config
	authSvc service.AuthService
	rend    *render.Render
}

func NewAuthHandler(appConf *config.Config, authSvc service.AuthService, rend *render.Render) AuthHandler {
	return &authHandler{
		appConf: appConf,
		authSvc: authSvc,
		rend:    rend,
	}
}

//nolint:dupl
func (h *authHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req coredto.RegisterUserRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.rend.Error(w, http.StatusBadRequest, message.ErrBadRequest)
		return
	}

	resp, err := h.authSvc.RegisterUser(req)

	var validationErr errs.ErrValidation
	if errors.As(err, &validationErr) {
		h.rend.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if errs.RPCErrorIs(err, errs.ErrUserAlreadyExists) {
		h.rend.Error(w, http.StatusConflict, message.ErrUserAlreadyExists)
		return
	}

	if err != nil {
		h.rend.Error(w, http.StatusInternalServerError, message.ErrInternalServer)
		return
	}

	//	Set session cookie
	http.SetCookie(w, resp.SessionCookie)

	h.rend.JSON(w, http.StatusCreated, response.RegisterUser{
		Username: resp.Username,
	})
}

//nolint:dupl
func (h *authHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req coredto.LoginUserRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.rend.Error(w, http.StatusBadRequest, message.ErrBadRequest)
		return
	}

	resp, err := h.authSvc.LoginUser(req)

	var validationErr errs.ErrValidation
	if errors.As(err, &validationErr) {
		h.rend.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if errs.RPCErrorIs(err, errs.ErrInvalidCredentials) {
		h.rend.Error(w, http.StatusUnauthorized, message.ErrInvalidCredentials)
		return
	}

	if err != nil {
		h.rend.Error(w, http.StatusInternalServerError, message.ErrInternalServer)
		return
	}

	//	Set session cookie
	http.SetCookie(w, resp.SessionCookie)

	h.rend.JSON(w, http.StatusOK, response.LoginUser{
		Username: resp.Username,
	})
}

func (h *authHandler) Logout(w http.ResponseWriter, r *http.Request) {
	session, err := r.Cookie(corehttp.BuildSessionCookieName(h.appConf))
	if err != nil || session.Valid() != nil {
		//	No Session cookie was found, or it was malformed
		h.rend.Error(w, http.StatusUnauthorized, message.ErrUnauthorized)
		return
	}

	req := coredto.LogoutUserRequest{
		SID: session.Value,
	}

	//	Logout user
	resp, err := h.authSvc.LogoutUser(req)
	if err != nil {
		h.rend.Error(w, http.StatusInternalServerError, message.ErrInternalServer)
		return
	}

	//	Remove session cookie
	http.SetCookie(w, resp.SessionCookie)

	h.rend.JSON(w, http.StatusNoContent, nil)
}
