package render

import (
	"net/http"

	"github.com/unrolled/render"
	"go.uber.org/zap"

	"github.com/Nurlan270/cloud-storage-go/internal/cloud_storage/transport/http/dto/response"
	"github.com/Nurlan270/cloud-storage-go/internal/core/logger"
)

type Render struct {
	*render.Render

	log *logger.Logger
}

func New() *Render {
	rend := render.New(render.Options{
		IndentJSON:                true,
		DisableHTTPErrorRendering: true,
	})

	return &Render{
		rend,
		logger.Get(),
	}
}

func (r *Render) Error(w http.ResponseWriter, statusCode int, msg string) {
	err := r.Render.JSON(w, statusCode, response.Error{
		Message: msg,
	})
	if err != nil {
		r.log.Error("render: failed to render error JSON response", zap.Error(err))
	}
}

func (r *Render) JSON(w http.ResponseWriter, statusCode int, data any) {
	err := r.Render.JSON(w, statusCode, data)
	if err != nil {
		r.log.Error("render: failed to render JSON response", zap.Error(err))
	}
}
