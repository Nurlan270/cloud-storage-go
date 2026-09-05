package app

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/Nurlan270/cloud-storage-go/internal/cloud_storage/closer"
	conf "github.com/Nurlan270/cloud-storage-go/internal/cloud_storage/config"
	"github.com/Nurlan270/cloud-storage-go/internal/cloud_storage/service"
	"github.com/Nurlan270/cloud-storage-go/internal/cloud_storage/transport/http/handlers"
	"github.com/Nurlan270/cloud-storage-go/internal/cloud_storage/transport/http/middleware"
	"github.com/Nurlan270/cloud-storage-go/internal/cloud_storage/transport/http/render"
	"github.com/Nurlan270/cloud-storage-go/internal/core/database"
	"github.com/Nurlan270/cloud-storage-go/internal/core/validator"
)

type diContainer struct {
	//	Configuration
	conf *conf.Config

	//	Core dependencies
	db        *sql.DB
	validator *validator.Validate

	//	HTTP
	router chi.Router
	server *http.Server
	render *render.Render

	//	Middlewares
	authMiddleware      middleware.AuthMiddleware
	guestMiddleware     middleware.GuestMiddleware
	loggerMiddleware    middleware.LoggerMiddleware
	rateLimitMiddleware middleware.RateLimitMiddleware

	//	Handlers
	authHandler handlers.AuthHandler
	userHandler handlers.UserHandler

	//	Services
	authSvc service.AuthService
}

// All dependencies are nil - they'll be injected
// lazily when they're called first time.
func newDIContainer(conf *conf.Config) *diContainer {
	return &diContainer{conf: conf}
}

func (c *diContainer) DB() *sql.DB {
	if c.db == nil {
		c.db = database.MustConnect(c.conf.DB)

		closer.Add("Database", func() error {
			return c.db.Close()
		})
	}

	return c.db
}

func (c *diContainer) Router() chi.Router {
	if c.router == nil {
		c.router = chi.NewRouter()
	}

	return c.router
}

func (c *diContainer) HTTPServer() *http.Server {
	if c.server == nil {
		c.server = &http.Server{
			Addr:         c.conf.HTTPServer.Address,
			Handler:      c.Router(),
			ReadTimeout:  c.conf.HTTPServer.Timeout,
			WriteTimeout: c.conf.HTTPServer.Timeout,
		}
	}

	return c.server
}

func (c *diContainer) AuthHandler() handlers.AuthHandler {
	if c.authHandler == nil {
		c.authHandler = handlers.NewAuthHandler(c.AuthService(), c.Render())
	}

	return c.authHandler
}

func (c *diContainer) Render() *render.Render {
	if c.render == nil {
		c.render = render.New()
	}

	return c.render
}

func (c *diContainer) Validator() *validator.Validate {
	if c.validator == nil {
		c.validator = validator.New()
	}

	return c.validator
}

func (c *diContainer) AuthService() service.AuthService {
	if c.authSvc == nil {
		c.authSvc = service.NewAuthService(c.Validator())

		closer.Add("Auth RPC Client", func() error {
			return c.authSvc.Close()
		})
	}

	return c.authSvc
}

func (c *diContainer) GuestMiddleware() middleware.GuestMiddleware {
	if c.guestMiddleware == nil {
		c.guestMiddleware = middleware.NewGuestMiddleware(c.conf, c.AuthService(), c.Render())
	}

	return c.guestMiddleware
}

func (c *diContainer) AuthMiddleware() middleware.AuthMiddleware {
	if c.authMiddleware == nil {
		c.authMiddleware = middleware.NewAuthMiddleware(c.conf, c.AuthService(), c.Render())
	}

	return c.authMiddleware
}

func (c *diContainer) LoggerMiddleware() middleware.LoggerMiddleware {
	if c.loggerMiddleware == nil {
		c.loggerMiddleware = middleware.NewLoggerMiddleware()
	}

	return c.loggerMiddleware
}

func (c *diContainer) RateLimitMiddleware() middleware.RateLimitMiddleware {
	if c.rateLimitMiddleware == nil {
		c.rateLimitMiddleware = middleware.NewRateLimitMiddleware(c.conf.Redis, c.Render())
	}

	return c.rateLimitMiddleware
}

func (c *diContainer) UserHandler() handlers.UserHandler {
	if c.userHandler == nil {
		c.userHandler = handlers.NewUserHandler(c.Render())
	}

	return c.userHandler
}
