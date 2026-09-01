package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/Nurlan270/cloud-storage-go/internal/core/logger"

	mw "github.com/go-chi/chi/middleware"
)

func Log(next http.Handler) http.Handler {
	l := logger.Get().WithOptions(zap.WithCaller(false))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := l.With(
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
