package middleware

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/middleware"
	"go.uber.org/zap"

	"github.com/Nurlan270/cloud-storage-go/internal/core/logger"
)

func Log(next http.Handler) http.Handler {
	l := logger.Get().WithOptions(zap.WithCaller(false))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := l.With(
			zap.String("ip", r.RemoteAddr),
			zap.String("user_agent", r.UserAgent()),
			zap.String("request_id", middleware.GetReqID(r.Context())),
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.Time("timestamp", time.Now().UTC()),
		)

		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		t1 := time.Now()
		defer func() {
			log.Info("Incoming request",
				zap.Int("status", ww.Status()),
				zap.Duration("duration", time.Since(t1).Round(time.Millisecond)),
			)
		}()

		next.ServeHTTP(ww, r)
	})
}
