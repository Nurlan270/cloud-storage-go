package app

import (
	"database/sql"

	"github.com/redis/go-redis/v9"

	conf "github.com/Nurlan270/cloud-storage-go/internal/auth_server/config"
	"github.com/Nurlan270/cloud-storage-go/internal/auth_server/repository"
	"github.com/Nurlan270/cloud-storage-go/internal/auth_server/service/auth"
	"github.com/Nurlan270/cloud-storage-go/internal/core/database"
	rdb "github.com/Nurlan270/cloud-storage-go/internal/core/redis"
)

type diContainer struct {
	//	Configuration
	conf *conf.Config

	//	Core dependencies
	db  *sql.DB
	rdb *redis.Client

	//	Repositories
	userRepo    repository.UserRepository
	sessionRepo repository.SessionRepository

	//	Services
	authSvc auth.Service
}

// All dependencies are nil - they'll be injected
// lazily when they're called first time.
func newDIContainer(conf *conf.Config) *diContainer {
	return &diContainer{conf: conf}
}

func (c *diContainer) DB() *sql.DB {
	if c.db == nil {
		c.db = database.MustConnect(c.conf.DB)
	}

	return c.db
}

func (c *diContainer) Redis() *redis.Client {
	if c.rdb == nil {
		c.rdb = rdb.MustConnect(c.conf.Redis, 1)
	}

	return c.rdb
}

func (c *diContainer) UserRepo() repository.UserRepository {
	if c.userRepo == nil {
		c.userRepo = repository.NewUserRepository(c.DB())
	}

	return c.userRepo
}

func (c *diContainer) SessionRepo() repository.SessionRepository {
	if c.sessionRepo == nil {
		c.sessionRepo = repository.NewSessionRepository(c.Redis())
	}

	return c.sessionRepo
}

func (c *diContainer) AuthService() auth.Service {
	if c.authSvc == nil {
		c.authSvc = auth.NewService(c.UserRepo(), c.SessionRepo(), c.conf)
	}

	return c.authSvc
}
