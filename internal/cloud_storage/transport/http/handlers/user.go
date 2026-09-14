package handlers

import (
	"net/http"

	"github.com/Nurlan270/cloud-storage-go/internal/cloud_storage/transport/http/context"
	"github.com/Nurlan270/cloud-storage-go/internal/cloud_storage/transport/http/dto/response"
	"github.com/Nurlan270/cloud-storage-go/internal/cloud_storage/transport/http/render"
)

type UserHandler interface {
	Me(w http.ResponseWriter, r *http.Request)
}

type userHandler struct {
	rend *render.Render
}

func NewUserHandler(rend *render.Render) UserHandler {
	return &userHandler{rend: rend}
}

func (h *userHandler) Me(w http.ResponseWriter, r *http.Request) {
	u := context.UserFromRequest(r)

	h.rend.JSON(w, http.StatusOK, response.UserMe{
		Username: u.Username,
	})
}
