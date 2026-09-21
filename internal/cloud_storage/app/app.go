package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"

	"github.com/Nurlan270/cloud-storage-go/internal/cloud_storage/closer"
	conf "github.com/Nurlan270/cloud-storage-go/internal/cloud_storage/config"
	"github.com/Nurlan270/cloud-storage-go/internal/cloud_storage/transport/http/message"
	"github.com/Nurlan270/cloud-storage-go/internal/core/config"
	"github.com/Nurlan270/cloud-storage-go/internal/core/logger"
)

type App struct {
	conf *conf.Config
	di   *diContainer
}

func New() *App {
	app := &App{
		conf: config.MustLoad[conf.Config](),
	}

	app.di = newDIContainer(app.conf)

	app.initDeps()

	return app
}

func (a *App) initDeps() {
	deps := []func(){
		a.initLogger,
		a.setupRouter,
		a.registerRoutes,
	}

	for _, init := range deps {
		init()
	}
}

func (a *App) initLogger() {
	logger.Init(a.conf)
}

func (a *App) setupRouter() {
	r := a.di.Router()

	//	404 Custom handler
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		a.di.Render().Error(w, http.StatusNotFound, message.ErrNotFound)
	})

	//	Common middlewares
	r.Use(
		middleware.Recoverer,
		middleware.CleanPath,
		middleware.RequestID,
		middleware.ClientIPFromHeader("X-Real-IP"),
		a.di.LoggerMiddleware().Log,
	)
}

func (a *App) registerRoutes() {
	//	Middlewares
	authMW := a.di.AuthMiddleware()
	guestMW := a.di.GuestMiddleware()
	limiterMW := a.di.RateLimitMiddleware()

	//	API Routes
	a.di.Router().Route("/api", func(r chi.Router) {
		//	Auth routes
		r.Route("/auth", func(r chi.Router) {
			authHandler := a.di.AuthHandler()

			r.With(
				guestMW.Guest,
				//	Rate limit: 10 requests per 5 minutes per IP
				limiterMW.LimitByEndpoint(10, 5*time.Minute),
			).Route("/", func(r chi.Router) {
				r.Post("/sign-up", authHandler.Register)
				r.Post("/sign-in", authHandler.Login)
			})

			r.With(
				authMW.Authenticate,
				//	Rate limit: 10 requests per 5 minutes per IP
				limiterMW.LimitByEndpoint(10, 5*time.Minute),
			).Post("/sign-out", authHandler.Logout)
		})

		//	User routes
		r.Route("/user", func(r chi.Router) {
			userHandler := a.di.UserHandler()

			r.Use(
				authMW.Authenticate,
				//	Rate limit: 10 requests per 5 minutes per IP
				limiterMW.LimitByEndpoint(10, 5*time.Minute),
			)

			r.Get("/me", userHandler.Me)
		})

		//	Resource routes
		r.Route("/resource", func(r chi.Router) {
			resourceHandler := a.di.ResourceHandler()

			r.Use(
				authMW.Authenticate,
				//	Rate limit: 10 requests per 2 minutes per IP
				limiterMW.LimitByEndpoint(10, 2*time.Minute),
			)

			r.Get("/", resourceHandler.GetResourceInfo)
			r.Post("/", resourceHandler.UploadResource)
			r.Delete("/", resourceHandler.DeleteResource)
			r.Get("/search", resourceHandler.SearchResource)
		})

		//	Directory routes
		r.Route("/directory", func(r chi.Router) {
			directoryHandler := a.di.DirectoryHandler()

			r.Use(
				authMW.Authenticate,
				//	Rate limit: 30 requests per minute per IP
				limiterMW.LimitByEndpoint(30, time.Minute),
			)

			r.Post("/", directoryHandler.CreateDirectory)
			r.Get("/", directoryHandler.GetDirectoryContent)
		})
	})
}

func (a *App) Run() error {
	log := logger.Get()
	srv := a.di.HTTPServer()

	serverErrCh := make(chan error, 1)

	//	Start HTTP Server
	go func() {
		log.Info("Starting HTTP Server", zap.String("address", a.conf.HTTPServer.Address))

		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrCh <- fmt.Errorf("http: server closed unexpectedly: %w", err)
			return
		}

		serverErrCh <- nil
	}()

	//	Graceful shutdown handling
	notifyCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	select {
	case <-notifyCtx.Done():
		log.Info("Shutting down server...")

		//	Double Ctrl+C pattern.
		//	Second Ctrl+C will close server immediately.
		stop()
	case err := <-serverErrCh:
		return err
	}

	//	15 Seconds to close HTTP Server & 10 Seconds to close all other services
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("http: failed to shutdown server: %w", err)
	}

	closerCtx, closerStop := context.WithTimeout(context.Background(), 10*time.Second)
	defer closerStop()

	if err := closer.CloseAll(closerCtx); err != nil {
		return fmt.Errorf("closer: failed to close all resources: %w", err)
	}

	return nil
}
