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
	"github.com/Nurlan270/cloud-storage-go/internal/core/transport/http/dto"
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
		a.registerRoutes,
	}

	for _, init := range deps {
		init()
	}
}

func (a *App) initLogger() {
	logger.Init(a.conf)
}

func (a *App) registerRoutes() {
	r := a.di.Router()

	//	Common middlewares
	r.Use(
		middleware.Recoverer,
		middleware.CleanPath,
		middleware.RequestID,
		mw.Log,
	)

	//	404 Custom handler
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		a.di.Render().JSON(w, http.StatusNotFound, dto.ErrorResponse{
			Message: message.ErrNotFound,
		})
	})

	//	API Routes
	r.Route("/api", func(r chi.Router) {
		//	Auth routes
		r.Route("/auth", func(r chi.Router) {
			authHandler := a.di.AuthHandler()

			r.Post("/sign-up", authHandler.Register)
		})
	})
}

func (a *App) Run() error {
	srv := a.di.HTTPServer()

	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("http: server closed unexpectedly: %v", err)
	}

	return nil
}
