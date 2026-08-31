package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/rpc"

	"github.com/unrolled/render"

	"github.com/Nurlan270/cloud-storage-go/internal/cloud_storage/service"
	"github.com/Nurlan270/cloud-storage-go/internal/cloud_storage/transport/http/message"
	errs "github.com/Nurlan270/cloud-storage-go/internal/core/errors"
	"github.com/Nurlan270/cloud-storage-go/internal/core/transport/http/dto"
)

type AuthHandler interface {
	Register(w http.ResponseWriter, r *http.Request)
}

type authHandler struct {
	authSvc service.AuthService
	rend    *render.Render
}

func NewAuthHandler(authSvc service.AuthService, rend *render.Render) AuthHandler {
	return &authHandler{
		authSvc: authSvc,
		rend:    rend,
	}
}

func (h *authHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterUserRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.rend.JSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Message: message.ErrParseData,
		})

		return
	}

	resp, err := h.authSvc.RegisterUser(req)

	var validationErr errs.ErrValidation
	if errors.As(err, &validationErr) {
		h.rend.JSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Message: err.Error(),
		})

		return
	}

	var rpcErr rpc.ServerError
	if errors.As(err, &rpcErr) && rpcErr.Error() == errs.ErrUserAlreadyExists.Error() {
		h.rend.JSON(w, http.StatusConflict, dto.ErrorResponse{
			Message: message.ErrUserAlreadyExists,
		})

		return
	}

	if err != nil {
		h.rend.JSON(w, http.StatusInternalServerError, dto.ErrorResponse{
			Message: message.ErrInternalServer,
		})

		return
	}

	//	Set session cookie
	http.SetCookie(w, resp.SessionCookie)

	h.rend.JSON(w, http.StatusCreated, dto.RegisterUserResponse{
		Username: resp.Username,
	})
}
