package handlers

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"go.uber.org/zap"

	"github.com/Nurlan270/cloud-storage-go/internal/cloud_storage/service"
	"github.com/Nurlan270/cloud-storage-go/internal/cloud_storage/transport/http/dto/request"
	"github.com/Nurlan270/cloud-storage-go/internal/cloud_storage/transport/http/message"
	"github.com/Nurlan270/cloud-storage-go/internal/cloud_storage/transport/http/render"
	errs "github.com/Nurlan270/cloud-storage-go/internal/core/errors"
	"github.com/Nurlan270/cloud-storage-go/internal/core/logger"
	"github.com/Nurlan270/cloud-storage-go/internal/core/validator"
)

type ResourceHandler interface {
	GetResourceInfo(w http.ResponseWriter, r *http.Request)
	UploadResource(w http.ResponseWriter, r *http.Request)
	DeleteResource(w http.ResponseWriter, r *http.Request)
	SearchResource(w http.ResponseWriter, r *http.Request)
	DownloadResource(w http.ResponseWriter, r *http.Request)
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
		h.rend.Error(w, http.StatusNotFound, message.ErrDirectoryNotFound)
		return
	}

	if err != nil {
		h.rend.Error(w, http.StatusInternalServerError, message.ErrInternalServer)
		return
	}

	h.rend.JSON(w, http.StatusNoContent, nil)
}

func (h *resourceHandler) DownloadResource(w http.ResponseWriter, r *http.Request) {
	//	Get data
	req := request.DownloadResource{
		Path: r.URL.Query().Get("path"),
	}

	//	Validate
	if err := h.validate.Struct(req); err != nil {
		h.rend.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	//	Download
	result, err := h.resourceSvc.Download(r.Context(), req)

	if errors.Is(err, errs.ErrResourceNotFound) {
		h.rend.Error(w, http.StatusNotFound, message.ErrResourceNotFound)
		return
	}

	if err != nil {
		h.rend.Error(w, http.StatusInternalServerError, message.ErrInternalServer)
		return
	}

	//	Close opened file
	defer result.Content.Close()

	//	Set headers
	w.Header().Set("Content-Type", result.Type)
	w.Header().Set("Content-Length", strconv.FormatInt(result.Size, 10))
	w.Header().Set("Content-Disposition", fmt.Sprintf(
		"attachment; filename=%q", result.Name,
	))

	//	Copy content into response
	if _, err = io.Copy(w, result.Content); err != nil {
		logger.Get().Error("download: failed to stream result content", zap.Error(err))
		h.rend.Error(w, http.StatusInternalServerError, message.ErrInternalServer)
	}
}
