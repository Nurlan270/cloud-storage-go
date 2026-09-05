package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Nurlan270/cloud-storage-go/internal/cloud_storage/service"
	"github.com/Nurlan270/cloud-storage-go/internal/cloud_storage/transport/http/message"
	"github.com/Nurlan270/cloud-storage-go/internal/cloud_storage/transport/http/render"
	errs "github.com/Nurlan270/cloud-storage-go/internal/core/errors"
	"github.com/Nurlan270/cloud-storage-go/internal/core/transport/http/dto"
)

type AuthHandler interface {
	Register(w http.ResponseWriter, r *http.Request)
	Login(w http.ResponseWriter, r *http.Request)
	Logout(w http.ResponseWriter, r *http.Request)
}

type handler struct {
	authSvc service.AuthService
	rend    *render.Render
}

func NewAuthHandler(authSvc service.AuthService, rend *render.Render) AuthHandler {
	return &handler{
		authSvc: authSvc,
		rend:    rend,
	}
}

//nolint:dupl
func (h *handler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterUserRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.rend.Error(w, http.StatusBadRequest, message.ErrInvalidRequestBody)
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

	h.rend.JSON(w, http.StatusCreated, dto.RegisterUserResponse{
		Username: resp.Username,
	})
}

//nolint:dupl
func (h *handler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginUserRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.rend.Error(w, http.StatusBadRequest, message.ErrInvalidRequestBody)
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

	h.rend.JSON(w, http.StatusOK, dto.LoginUserResponse{
		Username: resp.Username,
	})
}

func (h *handler) Logout(w http.ResponseWriter, r *http.Request) {
	resp, err := h.authSvc.LogoutUser()
	if err != nil {
		h.rend.Error(w, http.StatusInternalServerError, message.ErrInternalServer)
		return
	}

	//	Remove session cookie
	http.SetCookie(w, resp.SessionCookie)

	h.rend.JSON(w, http.StatusNoContent, nil)
}
