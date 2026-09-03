package app

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/go-chi/chi"

	"github.com/Nurlan270/cloud-storage-go/internal/cloud_storage/closer"
	conf "github.com/Nurlan270/cloud-storage-go/internal/cloud_storage/config"
	"github.com/Nurlan270/cloud-storage-go/internal/cloud_storage/service"
	"github.com/Nurlan270/cloud-storage-go/internal/cloud_storage/transport/http/handlers"
	"github.com/Nurlan270/cloud-storage-go/internal/cloud_storage/transport/http/render"
	"github.com/Nurlan270/cloud-storage-go/internal/core/database"
	"github.com/Nurlan270/cloud-storage-go/internal/core/validator"
)

type diContainer struct {
	//	Configuration
	conf *conf.Config

	//	Core dependencies
	//todo: replace with pool
	db        *sql.DB
	validator *validator.Validate

	//	HTTP
	router chi.Router
	server *http.Server
	render *render.Render

	//	Services
	authSvc service.AuthService

	//	Handlers
	authHandler handlers.AuthHandler
}

// All dependencies are nil - they'll be injected
// lazily when they're called first time.
func newDIContainer(conf *conf.Config) *diContainer {
	return &diContainer{conf: conf}
}

func (c *diContainer) DB() *sql.DB {
	if c.db == nil {
		c.db = database.MustConnect(c.conf.DB)

		closer.Add("Database", func(_ context.Context) error {
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
	}

	return c.authSvc
}
