package auth

import (
	"github.com/lib/pq"
	"github.com/lib/pq/pqerror"
	"golang.org/x/crypto/bcrypt"

	"github.com/Nurlan270/cloud-storage-go/internal/core/dto"
	errs "github.com/Nurlan270/cloud-storage-go/internal/core/errors"
	"github.com/Nurlan270/cloud-storage-go/internal/core/models"
)

type UserRepository interface {
	CreateUser(username, password string) (*models.User, error)
}

type AuthService interface {
	Register(req dto.RegisterUserRequest, resp *dto.RegisterUserResponse) error
}

type authService struct {
	repo UserRepository
}

func NewAuthService(repo UserRepository) AuthService {
	return &authService{
		repo: repo,
	}
}

func (s *authService) Register(req dto.RegisterUserRequest, resp *dto.RegisterUserResponse) error {
	//	Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	//	Create user
	u, err := s.repo.CreateUser(req.Username, string(hashedPassword))
	if err != nil {
		errUniq := pq.As(err, pqerror.UniqueViolation)
		if errUniq != nil {
			//	User with provided username already exists
			return errs.ErrUserAlreadyExists
		}

		return err
	}

	*resp = dto.RegisterUserResponse{
		Username: u.Username,
	}

	return nil
}
