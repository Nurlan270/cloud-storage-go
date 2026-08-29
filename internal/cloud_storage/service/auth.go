package service

import (
	"fmt"
	"net/rpc"

	"github.com/Nurlan270/cloud-storage-go/internal/core/dto"
	"github.com/Nurlan270/cloud-storage-go/internal/core/validator"
)

type AuthService interface {
	RegisterUser(req dto.RegisterUserRequest) (string, error)
}

type authService struct {
	validate *validator.Validate
}

func NewAuthService(validate *validator.Validate) AuthService {
	return &authService{
		validate: validate,
	}
}

func (s *authService) RegisterUser(req dto.RegisterUserRequest) (string, error) {
	if err := s.validate.Struct(req); err != nil {
		return "", s.validate.MapError(err)
	}

	client, err := rpc.DialHTTP("tcp", "auth_server:7070")
	if err != nil {
		return "", fmt.Errorf("rpc: failed connect: %s", err)
	}

	var username string

	err = client.Call("AuthService.Register", req, &username)
	if err != nil {
		return "", fmt.Errorf("rpc: failed call AuthService.Register: %s", err)
	}

	return username, nil
}
