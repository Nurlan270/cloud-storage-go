package middleware

import (
	"net/http"

	"github.com/Nurlan270/cloud-storage-go/internal/cloud_storage/config"
	"github.com/Nurlan270/cloud-storage-go/internal/cloud_storage/transport/http/context"
	"github.com/Nurlan270/cloud-storage-go/internal/cloud_storage/transport/http/message"
	"github.com/Nurlan270/cloud-storage-go/internal/cloud_storage/transport/http/render"
	errs "github.com/Nurlan270/cloud-storage-go/internal/core/errors"
	corehttp "github.com/Nurlan270/cloud-storage-go/internal/core/transport/http"
	rpcdto "github.com/Nurlan270/cloud-storage-go/internal/core/transport/rpc/dto"
)

type AuthMiddleware interface {
	Authenticate(next http.Handler) http.Handler
}

type AuthService interface {
	GetUserFromSID(sid string) (rpcdto.GetUserFromSIDResponse, error)
}

type authMiddleware struct {
	appConf *config.Config
	authSvc AuthService
	rend    *render.Render
}

func NewAuthMiddleware(appConf *config.Config, authSvc AuthService, rend *render.Render) AuthMiddleware {
	return &authMiddleware{
		appConf: appConf,
		authSvc: authSvc,
		rend:    rend,
	}
}

func (m *authMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := r.Cookie(corehttp.BuildSessionCookieName(m.appConf))
		if err != nil {
			//	No Session cookie was found
			m.rend.Error(w, http.StatusUnauthorized, message.ErrUnauthorized)
			return
		}

		resp, err := m.authSvc.GetUserFromSID(session.Value)
		if errs.RPCErrorIs(err, errs.ErrSessionNotFound) || errs.RPCErrorIs(err, errs.ErrSessionExpired) {
			//	Session is not valid
			m.rend.Error(w, http.StatusUnauthorized, message.ErrUnauthorized)
			return
		}

		if err != nil {
			m.rend.Error(w, http.StatusInternalServerError, message.ErrInternalServer)
			return
		}

		//	Put authenticated user into request context
		ctx := context.NewUserContext(r.Context(), resp.User)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
