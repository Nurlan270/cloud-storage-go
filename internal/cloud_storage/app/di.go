package app

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/minio/minio-go/v7"

	"github.com/Nurlan270/cloud-storage-go/internal/cloud_storage/closer"
	conf "github.com/Nurlan270/cloud-storage-go/internal/cloud_storage/config"
	"github.com/Nurlan270/cloud-storage-go/internal/cloud_storage/repository"
	"github.com/Nurlan270/cloud-storage-go/internal/cloud_storage/service"
	"github.com/Nurlan270/cloud-storage-go/internal/cloud_storage/transport/http/handlers"
	"github.com/Nurlan270/cloud-storage-go/internal/cloud_storage/transport/http/middleware"
	"github.com/Nurlan270/cloud-storage-go/internal/cloud_storage/transport/http/render"
	"github.com/Nurlan270/cloud-storage-go/internal/core/database"
	coreminio "github.com/Nurlan270/cloud-storage-go/internal/core/minio"
	"github.com/Nurlan270/cloud-storage-go/internal/core/validator"
)

type diContainer struct {
	//	Configuration
	conf *conf.Config

	//	Core dependencies
	dbPool    *pgxpool.Pool
	minio     *minio.Client
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
	authHandler      handlers.AuthHandler
	userHandler      handlers.UserHandler
	resourceHandler  handlers.ResourceHandler
	directoryHandler handlers.DirectoryHandler

	//	Services
	authSvc      service.AuthService
	resourceSvc  service.ResourceService
	directorySvc service.DirectoryService

	//	Repositories
	resourceRepo  repository.ResourceRepository
	directoryRepo repository.DirectoryRepository
}

// All dependencies are nil - they'll be injected
// lazily when they're called first time.
func newDIContainer(conf *conf.Config) *diContainer {
	return &diContainer{conf: conf}
}

func (c *diContainer) DB() *pgxpool.Pool {
	if c.dbPool == nil {
		c.dbPool = database.MustConnect(c.conf.DB)

		closer.Add("Database", func() error {
			c.dbPool.Close()
			return nil
		})
	}

	return c.dbPool
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
			ReadTimeout:  c.conf.HTTPServer.ReadTimeout,
			WriteTimeout: c.conf.HTTPServer.WriteTimeout,
			IdleTimeout:  c.conf.HTTPServer.IdleTimeout,
		}
	}

	return c.server
}

func (c *diContainer) AuthHandler() handlers.AuthHandler {
	if c.authHandler == nil {
		c.authHandler = handlers.NewAuthHandler(c.conf, c.AuthService(), c.Render())
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

func (c *diContainer) ResourceHandler() handlers.ResourceHandler {
	if c.resourceHandler == nil {
		c.resourceHandler = handlers.NewResourceHandler(c.ResourceService(), c.Render(), c.Validator())
	}

	return c.resourceHandler
}

func (c *diContainer) ResourceService() service.ResourceService {
	if c.resourceSvc == nil {
		c.resourceSvc = service.NewResourceService(c.Minio(), c.DB(), c.ResourceRepo(), c.DirectoryRepo())
	}

	return c.resourceSvc
}

func (c *diContainer) ResourceRepo() repository.ResourceRepository {
	if c.resourceRepo == nil {
		c.resourceRepo = repository.NewResourceRepository(c.DB())
	}

	return c.resourceRepo
}

func (c *diContainer) DirectoryHandler() handlers.DirectoryHandler {
	if c.directoryHandler == nil {
		c.directoryHandler = handlers.NewDirectoryHandler(c.DirectoryService(), c.Render(), c.Validator())
	}

	return c.directoryHandler
}

func (c *diContainer) DirectoryService() service.DirectoryService {
	if c.directorySvc == nil {
		c.directorySvc = service.NewDirectoryService(c.Minio(), c.DB(), c.DirectoryRepo())
	}

	return c.directorySvc
}

func (c *diContainer) DirectoryRepo() repository.DirectoryRepository {
	if c.directoryRepo == nil {
		c.directoryRepo = repository.NewDirectoryRepository(c.DB())
	}

	return c.directoryRepo
}

func (c *diContainer) Minio() *minio.Client {
	if c.minio == nil {
		c.minio = coreminio.MustConnect(c.conf.Minio)
	}

	return c.minio
}
