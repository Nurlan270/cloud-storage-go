package app

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/rpc"

	"github.com/Nurlan270/cloud-storage-go/internal/core/config"
)

type App struct {
	conf *Config
	di   *diContainer
}

func New() *App {
	app := App{
		conf: config.MustLoad[Config](),
	}

	app.di = newDIContainer(app.conf)

	return &app
}

func (a *App) Run() error {
	if err := rpc.RegisterName("AuthService", a.di.AuthService()); err != nil {
		return fmt.Errorf("rpc: failed to register auth service: %v", err)
	}

	rpc.HandleHTTP()

	l, err := net.Listen("tcp", "0.0.0.0:7070")
	if err != nil {
		return fmt.Errorf("net: failed to listen: %v", err)
	}
	defer l.Close()

	if err = http.Serve(l, nil); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("http: server closed unexpectedly: %v", err)
	}

	return nil
}
