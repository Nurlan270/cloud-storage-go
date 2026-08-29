package app

import (
	"net/http"

	"github.com/go-chi/chi"

	"github.com/Nurlan270/cloud-storage-go/internal/core/config"
)

type App struct {
	conf *Config
	di   *diContainer
}

func New() *App {
	app := &App{
		conf: config.MustLoad[Config](),
	}

	app.di = newDIContainer(app.conf)

	app.initDeps()

	return app
}

func (a *App) initDeps() {
	deps := []func(){
		a.registerRoutes,
	}

	for _, init := range deps {
		init()
	}
}

func (a *App) registerRoutes() {
	a.di.Router().Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	a.di.Router().Route("/api", func(r chi.Router) {
		//	Auth routes
		r.Route("/auth", func(r chi.Router) {
			authHandler := a.di.AuthHandler()

			r.Post("/sign-up", authHandler.Register)
		})
	})
}

func (a *App) Run() error {
	return a.di.HTTPServer().ListenAndServe()
}
