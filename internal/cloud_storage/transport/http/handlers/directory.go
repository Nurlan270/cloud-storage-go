package handlers

import (
	"errors"
	"net/http"

	"github.com/Nurlan270/cloud-storage-go/internal/cloud_storage/service"
	"github.com/Nurlan270/cloud-storage-go/internal/cloud_storage/transport/http/dto/request"
	"github.com/Nurlan270/cloud-storage-go/internal/cloud_storage/transport/http/message"
	"github.com/Nurlan270/cloud-storage-go/internal/cloud_storage/transport/http/render"
	errs "github.com/Nurlan270/cloud-storage-go/internal/core/errors"
	"github.com/Nurlan270/cloud-storage-go/internal/core/validator"
)

type DirectoryHandler interface {
	CreateDirectory(w http.ResponseWriter, r *http.Request)
	GetDirectoryContent(w http.ResponseWriter, r *http.Request)
}

type directoryHandler struct {
	dirSvc   service.DirectoryService
	rend     *render.Render
	validate *validator.Validate
}

func NewDirectoryHandler(
	dirSvc service.DirectoryService,
	rend *render.Render,
	validate *validator.Validate,
) DirectoryHandler {
	return &directoryHandler{
		dirSvc:   dirSvc,
		rend:     rend,
		validate: validate,
	}
}

func (h *directoryHandler) CreateDirectory(w http.ResponseWriter, r *http.Request) {
	//	Get data
	req := request.CreateDirectory{
		Path: r.URL.Query().Get("path"),
	}

	//	Validate
	if err := h.validate.Struct(req); err != nil {
		h.rend.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	//	Create
	info, err := h.dirSvc.Create(r.Context(), req)

	if errors.Is(err, errs.ErrDirectoryAlreadyExists) {
		h.rend.Error(w, http.StatusConflict, message.ErrDirectoryAlreadyExists)
		return
	}

	if errors.Is(err, errs.ErrParentDirectoryNotExists) {
		h.rend.Error(w, http.StatusNotFound, message.ErrParentDirectoryNotExists)
		return
	}

	if err != nil {
		h.rend.Error(w, http.StatusInternalServerError, message.ErrInternalServer)
		return
	}

	h.rend.JSON(w, http.StatusCreated, info)
}

func (h *directoryHandler) GetDirectoryContent(w http.ResponseWriter, r *http.Request) {
	//	Get data
	req := request.GetDirectoryContent{
		Path: r.URL.Query().Get("path"),
	}

	//	Validate
	if err := h.validate.Struct(req); err != nil {
		h.rend.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	//	Create
	list, err := h.dirSvc.GetContent(r.Context(), req)

	if errors.Is(err, errs.ErrDirectoryNotFound) {
		h.rend.Error(w, http.StatusNotFound, message.ErrDirectoryNotExists)
		return
	}

	if err != nil {
		h.rend.Error(w, http.StatusInternalServerError, message.ErrInternalServer)
		return
	}

	h.rend.JSON(w, http.StatusOK, list)
}
