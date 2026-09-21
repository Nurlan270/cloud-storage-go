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

type ResourceHandler interface {
	GetResourceInfo(w http.ResponseWriter, r *http.Request)
	UploadResource(w http.ResponseWriter, r *http.Request)
	DeleteResource(w http.ResponseWriter, r *http.Request)
	SearchResource(w http.ResponseWriter, r *http.Request)
}

type resourceHandler struct {
	resourceSvc service.ResourceService
	rend        *render.Render
	validate    *validator.Validate
}

func NewResourceHandler(
	resourceSvc service.ResourceService,
	rend *render.Render,
	validate *validator.Validate,
) ResourceHandler {
	return &resourceHandler{
		resourceSvc: resourceSvc,
		rend:        rend,
		validate:    validate,
	}
}

func (h *resourceHandler) UploadResource(w http.ResponseWriter, r *http.Request) {
	//	Parse
	var maxMemory int64 = 10 << 20 // Upload 10 MiB into memory, rest goes to disk
	if err := r.ParseMultipartForm(maxMemory); err != nil {
		h.rend.Error(w, http.StatusBadRequest, message.ErrBadRequest)
		return
	}

	//	Get data
	req := request.UploadResource{
		Object: r.MultipartForm.File["object"],
		Path:   r.URL.Query().Get("path"),
	}

	//	Validate
	if err := h.validate.Struct(req); err != nil {
		h.rend.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	//	Upload
	list, err := h.resourceSvc.Upload(r.Context(), req)

	if errors.Is(err, errs.ErrResourceAlreadyExists) {
		h.rend.Error(w, http.StatusConflict, message.ErrResourceAlreadyExists)
		return
	}

	if err != nil {
		h.rend.Error(w, http.StatusInternalServerError, message.ErrInternalServer)
		return
	}

	h.rend.JSON(w, http.StatusCreated, list)
}

func (h *resourceHandler) GetResourceInfo(w http.ResponseWriter, r *http.Request) {
	//	Get data
	req := request.GetResourceInfo{
		Path: r.URL.Query().Get("path"),
	}

	//	Validate
	if err := h.validate.Struct(req); err != nil {
		h.rend.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	//	Get info
	info, err := h.resourceSvc.GetInfo(r.Context(), req)

	if errors.Is(err, errs.ErrResourceNotFound) {
		h.rend.Error(w, http.StatusNotFound, message.ErrResourceNotFound)
		return
	}

	if err != nil {
		h.rend.Error(w, http.StatusInternalServerError, message.ErrInternalServer)
		return
	}

	h.rend.JSON(w, http.StatusOK, info)
}

func (h *resourceHandler) SearchResource(w http.ResponseWriter, r *http.Request) {
	//	Get data
	req := request.SearchResource{
		Query: r.URL.Query().Get("query"),
	}

	//	Validate
	if err := h.validate.Struct(req); err != nil {
		h.rend.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	//	Search
	list, err := h.resourceSvc.Search(r.Context(), req)
	if err != nil {
		h.rend.Error(w, http.StatusInternalServerError, message.ErrInternalServer)
		return
	}

	h.rend.JSON(w, http.StatusOK, list)
}

func (h *resourceHandler) DeleteResource(w http.ResponseWriter, r *http.Request) {
	//	Get data
	req := request.DeleteResource{
		Path: r.URL.Query().Get("path"),
	}

	//	Validate
	if err := h.validate.Struct(req); err != nil {
		h.rend.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	//	Delete
	err := h.resourceSvc.Delete(r.Context(), req)

	if errors.Is(err, errs.ErrResourceNotFound) {
		h.rend.Error(w, http.StatusNotFound, message.ErrResourceNotFound)
		return
	}

	if errors.Is(err, errs.ErrDirectoryNotFound) {
		h.rend.Error(w, http.StatusNotFound, message.ErrDirectoryNotExists)
		return
	}

	if err != nil {
		h.rend.Error(w, http.StatusInternalServerError, message.ErrInternalServer)
		return
	}

	h.rend.JSON(w, http.StatusNoContent, nil)
}
