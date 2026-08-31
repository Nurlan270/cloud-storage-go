package auth

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/iancoleman/strcase"
	"github.com/lib/pq"
	"github.com/lib/pq/pqerror"
	"golang.org/x/crypto/bcrypt"

	"github.com/Nurlan270/cloud-storage-go/internal/auth_server/config"
	errs "github.com/Nurlan270/cloud-storage-go/internal/core/errors"
	"github.com/Nurlan270/cloud-storage-go/internal/core/models"
	httpdto "github.com/Nurlan270/cloud-storage-go/internal/core/transport/http/dto"
	rpcdto "github.com/Nurlan270/cloud-storage-go/internal/core/transport/rpc/dto"

	guuid "github.com/google/uuid"
)

type UserRepository interface {
	CreateUser(user *models.User) (*models.User, error)
}

type SessionRepository interface {
	CreateSession(session *models.Session) (*models.Session, error)
}

type Service interface {
	Register(req httpdto.RegisterUserRequest, resp *rpcdto.RegisterUserResponse) error
}

type service struct {
	userRepo UserRepository
	sessRepo SessionRepository
	appConf  *config.Config
}

func NewService(userRepo UserRepository, sessRepo SessionRepository, appConf *config.Config) Service {
	return &service{
		userRepo: userRepo,
		sessRepo: sessRepo,
		appConf:  appConf,
	}
}

func (s *service) Register(req httpdto.RegisterUserRequest, resp *rpcdto.RegisterUserResponse) error {
	//	Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("bcrypt: failed to hash password: %w", err)
	}

	user := &models.User{
		Username: req.Username,
		Password: string(hashedPassword),
	}

	//	Create user
	dbUser, err := s.userRepo.CreateUser(user)
	if err != nil {
		if uniqErr := pq.As(err, pqerror.UniqueViolation); uniqErr != nil {
			//	User with provided username already exists
			return errs.ErrUserAlreadyExists
		}

		return fmt.Errorf("user repo: failed to create user: %w", err)
	}

	//	Generate UUID of Session
	uuid, err := generateUUID()
	if err != nil {
		return fmt.Errorf("uuid: failed to generate: %w", err)
	}

	expiresAt := time.Now().Add(s.appConf.Session.ExpiresIn).UTC()
	sess := &models.Session{
		UUID:      uuid,
		UserID:    dbUser.ID,
		ExpiresAt: expiresAt,
	}

	//	Create session
	dbSession, err := s.sessRepo.CreateSession(sess)
	if err != nil {
		return fmt.Errorf("session repo: failed to create session: %w", err)
	}

	//	Create session cookie
	cookieName := buildCookieName(s.appConf.GetAppName())
	*resp = rpcdto.RegisterUserResponse{
		Username: dbUser.Username,
		SessionCookie: &http.Cookie{
			Name:     cookieName,
			Value:    dbSession.UUID,
			Expires:  expiresAt,
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		},
	}

	return nil
}

func generateUUID() (string, error) {
	uuid, err := guuid.NewRandom()
	if err != nil {
		return "", err
	}

	return uuid.String(), nil
}

func buildCookieName(s string) string {
	str := strcase.ToSnake(s)
	return strings.Trim(str, "_") + "_session"
}
