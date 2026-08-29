package app

import (
	"database/sql"

	"github.com/Nurlan270/cloud-storage-go/internal/auth_server/service/auth"
	"github.com/Nurlan270/cloud-storage-go/internal/core/database"
	"github.com/Nurlan270/cloud-storage-go/internal/core/repository"
)

type diContainer struct {
	//	Configuration
	conf *Config

	//	Core dependencies
	//todo: replace with pool
	db *sql.DB

	//	Repositories
	userRepo repository.UserRepository

	//	Services
	authSvc auth.AuthService
}

// All dependencies are nil - they'll be injected
// lazily when they're called first time.
func newDIContainer(conf *Config) *diContainer {
	return &diContainer{conf: conf}
}

func (c *diContainer) DB() *sql.DB {
	if c.db == nil {
		c.db = database.MustConnect(c.conf.DB)
	}

	return c.db
}

func (c *diContainer) UserRepo() repository.UserRepository {
	if c.userRepo == nil {
		c.userRepo = repository.NewUserRepository(c.DB())
	}

	return c.userRepo
}

func (c *diContainer) AuthService() auth.AuthService {
	if c.authSvc == nil {
		c.authSvc = auth.NewAuthService(c.UserRepo())
	}

	return c.authSvc
}
