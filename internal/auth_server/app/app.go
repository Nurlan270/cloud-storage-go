package app

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/rpc"

	conf "github.com/Nurlan270/cloud-storage-go/internal/auth_server/config"
	"github.com/Nurlan270/cloud-storage-go/internal/core/config"
)

type App struct {
	conf *conf.Config
	di   *diContainer
}

func New() *App {
	app := App{
		conf: config.MustLoad[conf.Config](),
	}

	app.di = newDIContainer(app.conf)

	app.initDeps()

	return &app
}

func (a *App) initDeps() {
	deps := []func(){
		a.initRPC,
	}

	for _, init := range deps {
		init()
	}
}

func (a *App) initRPC() {
	if err := rpc.RegisterName("AuthService", a.di.AuthService()); err != nil {
		panic(fmt.Sprintf("rpc: failed to register auth service: %v", err))
	}

	if err := rpc.RegisterName("PingService", a.di.PingService()); err != nil {
		panic(fmt.Sprintf("rpc: failed to register ping service: %v", err))
	}

	rpc.HandleHTTP()
}

func (a *App) Run() error {
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
