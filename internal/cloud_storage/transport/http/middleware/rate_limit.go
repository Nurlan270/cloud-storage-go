package middleware

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"

	"github.com/Nurlan270/cloud-storage-go/internal/cloud_storage/transport/http/message"
	"github.com/Nurlan270/cloud-storage-go/internal/cloud_storage/transport/http/render"
	"github.com/Nurlan270/cloud-storage-go/internal/core/redis"

	httprateredis "github.com/go-chi/httprate-redis"
)

type RateLimitMiddleware interface {
	Limit(requestLimit int, windowLength time.Duration) func(next http.Handler) http.Handler
	LimitByEndpoint(requestLimit int, windowLength time.Duration) func(next http.Handler) http.Handler
}

type rateLimitMiddleware struct {
	conf redis.Config
	rend *render.Render
}

func NewRateLimitMiddleware(conf redis.Config, rend *render.Render) RateLimitMiddleware {
	return &rateLimitMiddleware{
		conf: conf,
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
		setHeaders(),
		m.setRedisLimiter(),
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
		setHeaders(),
		m.setRedisLimiter(),
		m.renderResponse(),
	)
}

func (m *rateLimitMiddleware) renderResponse() httprate.Option {
	return httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) {
		m.rend.Error(w, http.StatusTooManyRequests, message.ErrTooManyRequests)
	})
}

func setHeaders() httprate.Option {
	return httprate.WithResponseHeaders(httprate.ResponseHeaders{
		Limit:      "X-RateLimit-Limit",
		Remaining:  "X-RateLimit-Remaining",
		RetryAfter: "Retry-After",
		Reset:      "", // omit
		Increment:  "", // omit
	})
}

func (m *rateLimitMiddleware) setRedisLimiter() httprate.Option {
	return httprateredis.WithRedisLimitCounter(&httprateredis.Config{
		PrefixKey: "rate_limit",
		Host:      m.conf.Host,
		Port:      m.conf.Port,
		Password:  m.conf.Password,
		MaxIdle:   10,
		MaxActive: 10,
		DBIndex:   0,
	})
}

func clientIPKey(r *http.Request) (string, error) {
	return "limit:" + httprate.CanonicalizeIP(middleware.GetClientIP(r.Context())), nil
}
