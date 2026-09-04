package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/Nurlan270/cloud-storage-go/internal/core/logger"

	mw "github.com/go-chi/chi/v5/middleware"
)

type LoggerMiddleware interface {
	Log(next http.Handler) http.Handler
}

type loggerMiddleware struct {
	log *zap.Logger
}

func NewLoggerMiddleware() LoggerMiddleware {
	return &loggerMiddleware{
		log: logger.Get().WithOptions(zap.WithCaller(false)),
	}
}

func (m *loggerMiddleware) Log(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := m.log.With(
			zap.String("ip", r.RemoteAddr),
			zap.String("user_agent", r.UserAgent()),
			zap.String("request_id", mw.GetReqID(r.Context())),
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.Time("timestamp", time.Now().UTC()),
		)

		ww := mw.NewWrapResponseWriter(w, r.ProtoMajor)

		t1 := time.Now()
		defer func() {
			log.Info("Incoming request",
				zap.Int("status", ww.Status()),
				zap.Duration("duration_ms", time.Since(t1).Round(time.Millisecond)),
			)
		}()

		next.ServeHTTP(ww, r)
	})
}
