package middleware

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"

	"github.com/Nurlan270/cloud-storage-go/internal/cloud_storage/transport/http/message"
	"github.com/Nurlan270/cloud-storage-go/internal/cloud_storage/transport/http/render"
)

type RateLimitMiddleware interface {
	Limit(requestLimit int, windowLength time.Duration) func(next http.Handler) http.Handler
	LimitByEndpoint(requestLimit int, windowLength time.Duration) func(next http.Handler) http.Handler
}

type rateLimitMiddleware struct {
	rend *render.Render
}

func NewRateLimitMiddleware(rend *render.Render) RateLimitMiddleware {
	return &rateLimitMiddleware{
		rend: rend,
	}
}

func (m *rateLimitMiddleware) Limit(
	requestLimit int,
	windowLength time.Duration,
) func(next http.Handler) http.Handler {
	return httprate.LimitBy(
		requestLimit,
		windowLength,
		clientIPKey,
		m.renderResponse(),
	)
}

func (m *rateLimitMiddleware) LimitByEndpoint(
	requestLimit int,
	windowLength time.Duration,
) func(next http.Handler) http.Handler {
	return httprate.LimitBy(
		requestLimit,
		windowLength,
		httprate.JoinKeys(clientIPKey, httprate.KeyByEndpoint),
		m.renderResponse(),
	)
}

func (m *rateLimitMiddleware) renderResponse() httprate.Option {
	return httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) {
		m.rend.Error(w, http.StatusTooManyRequests, message.ErrTooManyRequests)
	})
}

func clientIPKey(r *http.Request) (string, error) {
	return "limit:" + httprate.CanonicalizeIP(middleware.GetClientIP(r.Context())), nil
}
