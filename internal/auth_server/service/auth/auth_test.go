package auth

import (
	"context"
	"database/sql"
	"github.com/Nurlan270/cloud-storage-go/internal/auth_server/testutil"
	errs "github.com/Nurlan270/cloud-storage-go/internal/core/errors"
	"github.com/Nurlan270/cloud-storage-go/internal/core/models"
	"github.com/Nurlan270/cloud-storage-go/internal/core/repository"
	httpdto "github.com/Nurlan270/cloud-storage-go/internal/core/transport/http/dto"
	rpcdto "github.com/Nurlan270/cloud-storage-go/internal/core/transport/rpc/dto"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/require"
	"log"
	"os"
	"testing"
)

var (
	db       *sql.DB
	userRepo repository.UserRepository
	sessRepo repository.SessionRepository
	authSvc  Service
)

func TestMain(m *testing.M) {
	conf := testutil.NewTestConfig()

	var (
		err     error
		cleanup func() error
	)

	//	Setup test database
	db, cleanup, err = testutil.NewTestDB(context.Background(), conf.DB)
	if err != nil {
		log.Fatalf("failed to create test database: %s", err)
	}

	//	Arrange
	userRepo = repository.NewUserRepository(db)
	sessRepo = repository.NewSessionRepository(db)

	authSvc = NewService(userRepo, sessRepo, conf)

	//	Run tests
	code := m.Run()

	//	Clean test database
	if err = cleanup(); err != nil {
		log.Printf("failed to cleanup: %s", err)
	}

	os.Exit(code)
}

func TestAuthService(t *testing.T) {
	t.Run("it creates new user on register", func(t *testing.T) {
		testutil.CleanupDatabase(t, db)

		req := httpdto.RegisterUserRequest{
			Username: "john_doe",
			Password: "secret123",
		}

		var resp rpcdto.RegisterUserResponse

		err := authSvc.Register(req, &resp)

		require.NoErrorf(t, err, "Should register new user")

		expected := &models.User{
			ID:       1,
			Username: req.Username,
		}

		actual, err := userRepo.GetUserFromUsername(resp.Username)

		require.NoError(t, err)
		require.Truef(t, cmp.Equal(expected, actual, cmp.Options{
			cmpopts.IgnoreFields(models.User{}, "Password"),
		}), "Should find just created user")
	})

	t.Run("it returns error if user already exists on register", func(t *testing.T) {
		testutil.CleanupDatabase(t, db)

		//	User 1
		user1 := httpdto.RegisterUserRequest{
			Username: "john_doe",
			Password: "secret123",
		}

		var resp1 rpcdto.RegisterUserResponse

		err := authSvc.Register(user1, &resp1)

		require.NoErrorf(t, err, "Should register first user")

		//	User 2
		user2 := httpdto.RegisterUserRequest{
			Username: "john_doe",
			Password: "another123",
		}

		var resp2 rpcdto.RegisterUserResponse

		err = authSvc.Register(user2, &resp2)

		require.ErrorIsf(t, err, errs.ErrUserAlreadyExists, "Should return error for second user")
	})
}
