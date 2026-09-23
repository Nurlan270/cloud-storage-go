package request

import "mime/multipart"

type UploadResource struct {
	Object []*multipart.FileHeader `validate:"required,min=1"`
	Path   string                  `validate:"path"`
}

type GetResourceInfo struct {
	Path string `validate:"required,path"`
}

type DeleteResource struct {
	Path string `validate:"required,path"`
}

type DownloadResource struct {
	Path string `validate:"required,path"`
}

type SearchResource struct {
	Query string `validate:"required"`
}
