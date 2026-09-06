package auth

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/lib/pq"
	"github.com/lib/pq/pqerror"
	"golang.org/x/crypto/bcrypt"

	"github.com/Nurlan270/cloud-storage-go/internal/auth_server/config"
	errs "github.com/Nurlan270/cloud-storage-go/internal/core/errors"
	"github.com/Nurlan270/cloud-storage-go/internal/core/models"
	corehttp "github.com/Nurlan270/cloud-storage-go/internal/core/transport/http"
	httpdto "github.com/Nurlan270/cloud-storage-go/internal/core/transport/http/dto"
	rpcdto "github.com/Nurlan270/cloud-storage-go/internal/core/transport/rpc/dto"

	guuid "github.com/google/uuid"
)

type UserRepository interface {
	CreateUser(user *models.User) (*models.User, error)
	GetUserFromUserID(userID uint64) (*models.User, error)
	GetUserFromUsername(username string) (*models.User, error)
}

type SessionRepository interface {
	CreateSession(session *models.Session) (*models.Session, error)
	GetSessionFromSID(sid string) (*models.Session, error)
}

type Service interface {
	Register(req httpdto.RegisterUserRequest, resp *rpcdto.RegisterUserResponse) error
	Login(req httpdto.LoginUserRequest, resp *rpcdto.LoginUserResponse) error
	Logout(_ int, resp *rpcdto.LogoutUserResponse) error
	GetUserFromSID(sid string, resp *rpcdto.GetUserFromSIDResponse) error
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
	repoUser, err := s.userRepo.CreateUser(user)
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

	expiresIn := s.appConf.Session.ExpiresIn
	sess := &models.Session{
		UUID:      uuid,
		UserID:    repoUser.ID,
		ExpiresIn: expiresIn,
	}

	//	Create session
	repoSession, err := s.sessRepo.CreateSession(sess)
	if err != nil {
		return fmt.Errorf("session repo: failed to create session: %w", err)
	}

	//	Create session cookie
	expiresAt := time.Now().Add(expiresIn).UTC()
	cookieName := corehttp.BuildSessionCookieName(s.appConf)
	*resp = rpcdto.RegisterUserResponse{
		Username: repoUser.Username,
		SessionCookie: &http.Cookie{
			Name:     cookieName,
			Value:    repoSession.UUID,
			Expires:  expiresAt,
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		},
	}

	return nil
}

func (s *service) Login(req httpdto.LoginUserRequest, resp *rpcdto.LoginUserResponse) error {
	//	Get user
	repoUser, err := s.userRepo.GetUserFromUsername(req.Username)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFound) {
			//	User with provided username wasn't found
			return errs.ErrInvalidCredentials
		}

		return fmt.Errorf("user repo: failed to get user: %w", err)
	}

	//	Compare password hashes
	if err = bcrypt.CompareHashAndPassword([]byte(repoUser.Password), []byte(req.Password)); err != nil {
		//	Provided password is invalid
		return errs.ErrInvalidCredentials
	}

	//	Generate UUID of Session
	uuid, err := generateUUID()
	if err != nil {
		return fmt.Errorf("uuid: failed to generate: %w", err)
	}

	expiresIn := s.appConf.Session.ExpiresIn
	expiresAt := time.Now().Add(expiresIn).UTC()
	sess := &models.Session{
		UUID:      uuid,
		UserID:    repoUser.ID,
		ExpiresIn: expiresIn,
	}

	//	Create session
	repoSession, err := s.sessRepo.CreateSession(sess)
	if err != nil {
		return fmt.Errorf("session repo: failed to create session: %w", err)
	}

	//	Create session cookie
	cookieName := corehttp.BuildSessionCookieName(s.appConf)
	*resp = rpcdto.LoginUserResponse{
		Username: repoUser.Username,
		SessionCookie: &http.Cookie{
			Name:     cookieName,
			Value:    repoSession.UUID,
			Expires:  expiresAt,
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		},
	}

	return nil
}

func (s *service) Logout(_ int, resp *rpcdto.LogoutUserResponse) error {
	//	Create session cookie with negative values (to logout user)
	cookieName := corehttp.BuildSessionCookieName(s.appConf)
	*resp = rpcdto.LogoutUserResponse{
		SessionCookie: &http.Cookie{
			Name:     cookieName,
			Value:    "",
			MaxAge:   -1,
			Expires:  time.Unix(0, 0),
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		},
	}

	return nil
}

func (s *service) GetUserFromSID(sid string, resp *rpcdto.GetUserFromSIDResponse) error {
	session, err := s.sessRepo.GetSessionFromSID(sid)
	if err != nil {
		if errors.Is(err, errs.ErrSessionInvalid) {
			return errs.ErrSessionInvalid
		}

		return fmt.Errorf("session repo: failed to get session: %w", err)
	}

	user, err := s.userRepo.GetUserFromUserID(session.UserID)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFound) {
			return errs.ErrUserNotFound
		}

		return fmt.Errorf("user repo: failed to get user: %w", err)
	}

	*resp = rpcdto.GetUserFromSIDResponse{
		User: user,
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
