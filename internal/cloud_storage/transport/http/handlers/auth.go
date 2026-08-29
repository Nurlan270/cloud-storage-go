package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/unrolled/render"

	"github.com/Nurlan270/cloud-storage-go/internal/cloud_storage/service"
	"github.com/Nurlan270/cloud-storage-go/internal/cloud_storage/transport/dto"
	"github.com/Nurlan270/cloud-storage-go/internal/cloud_storage/transport/message"
	coredto "github.com/Nurlan270/cloud-storage-go/internal/core/dto"
	errs "github.com/Nurlan270/cloud-storage-go/internal/core/errors"
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
	var req coredto.RegisterUserRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.rend.JSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Message: message.ParseData,
		})

		return
	}

	username, err := h.authSvc.RegisterUser(req)

	var validationErr errs.ErrValidation
	if errors.As(err, &validationErr) {
		h.rend.JSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Message: err.Error(),
		})

		return
	}

	if errors.Is(err, errs.ErrUserAlreadyExists) {
		h.rend.JSON(w, http.StatusConflict, dto.ErrorResponse{
			Message: message.UserAlreadyExists,
		})

		return
	}

	if err != nil {
		h.rend.JSON(w, http.StatusInternalServerError, dto.ErrorResponse{
			Message: message.InternalError,
		})

		return
	}

	h.rend.JSON(w, http.StatusOK, coredto.RegisterUserResponse{
		Username: username,
	})
}
