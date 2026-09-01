package middleware

import (
	"net/http"

	"github.com/Nurlan270/cloud-storage-go/internal/cloud_storage/config"
	"github.com/Nurlan270/cloud-storage-go/internal/cloud_storage/transport/http/message"
	"github.com/Nurlan270/cloud-storage-go/internal/cloud_storage/transport/http/render"
	errs "github.com/Nurlan270/cloud-storage-go/internal/core/errors"
	corehttp "github.com/Nurlan270/cloud-storage-go/internal/core/transport/http"
)

type GuestMiddleware interface {
	Guest(next http.Handler) http.Handler
}

type guestMiddleware struct {
	appConf *config.Config
	authSvc AuthService
	rend    *render.Render
}

func NewGuestMiddleware(appConf *config.Config, authSvc AuthService, rend *render.Render) GuestMiddleware {
	return &guestMiddleware{
		appConf: appConf,
		authSvc: authSvc,
		rend:    rend,
	}
}

func (m *guestMiddleware) Guest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := r.Cookie(corehttp.BuildSessionCookieName(m.appConf))
		if err != nil {
			//	No Session cookie was found
			next.ServeHTTP(w, r)
			return
		}

		_, err = m.authSvc.GetUserFromSID(session.Value)
		if errs.RPCErrorIs(err, errs.ErrSessionNotFound) || errs.RPCErrorIs(err, errs.ErrSessionExpired) {
			//	Session is not valid
			next.ServeHTTP(w, r)
			return
		}

		if err != nil {
			m.rend.Error(w, http.StatusInternalServerError, message.ErrInternalServer)
			return
		}

		m.rend.Error(w, http.StatusForbidden, message.ErrForbidden)
	})
}
