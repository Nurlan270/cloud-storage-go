package app

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"

	conf "github.com/Nurlan270/cloud-storage-go/internal/cloud_storage/config"
	"github.com/Nurlan270/cloud-storage-go/internal/cloud_storage/transport/http/message"
	mw "github.com/Nurlan270/cloud-storage-go/internal/cloud_storage/transport/http/middleware"
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
		mw.Log,
	)
}

func (a *App) registerRoutes() {
	//	Middlewares
	authMW := mw.NewAuthMiddleware(a.conf, a.di.AuthService(), a.di.Render())
	guestMW := mw.NewGuestMiddleware(a.conf, a.di.AuthService(), a.di.Render())

	//	API Routes
	a.di.Router().Route("/api", func(r chi.Router) {
		//	Auth routes
		r.Route("/auth", func(r chi.Router) {
			//todo: add rate limiter
			authHandler := a.di.AuthHandler()

			r.With(guestMW.Guest).Post("/sign-up", authHandler.Register)
			r.With(guestMW.Guest).Post("/sign-in", authHandler.Login)

			r.With(authMW.Authenticate).Post("/sign-out", authHandler.Logout)
		})
	})
}

func (a *App) Run() error {
	srv := a.di.HTTPServer()

	//todo: add graceful shutdown
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("http: server closed unexpectedly: %v", err)
	}

	return nil
}
