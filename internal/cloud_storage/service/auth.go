package service

import (
	"fmt"
	"net/rpc"
	"sync"

	"go.uber.org/zap"

	errs "github.com/Nurlan270/cloud-storage-go/internal/core/errors"
	"github.com/Nurlan270/cloud-storage-go/internal/core/logger"
	"github.com/Nurlan270/cloud-storage-go/internal/core/transport/http/dto"
	rpcdto "github.com/Nurlan270/cloud-storage-go/internal/core/transport/rpc/dto"
	"github.com/Nurlan270/cloud-storage-go/internal/core/validator"
)

type AuthService interface {
	RegisterUser(req dto.RegisterUserRequest) (rpcdto.RegisterUserResponse, error)
	LoginUser(req dto.LoginUserRequest) (rpcdto.LoginUserResponse, error)
	LogoutUser() (rpcdto.LogoutUserResponse, error)
	GetUserFromSID(sid string) (rpcdto.GetUserFromSIDResponse, error)
}

type authService struct {
	validate *validator.Validate

	//	This mutex is used if client disconnects.
	//	In this case it will lock until it reconnects back
	mu     sync.Mutex
	client *rpc.Client
}

func NewAuthService(validate *validator.Validate) AuthService {
	return &authService{
		validate: validate,
	}
}

func (s *authService) RegisterUser(req dto.RegisterUserRequest) (rpcdto.RegisterUserResponse, error) {
	var resp rpcdto.RegisterUserResponse

	//	Validate request
	if err := s.validate.Struct(req); err != nil {
		return resp, s.validate.MapError(err)
	}

	//	Get RPC Client
	client, err := s.getClient()
	if err != nil {
		logger.Get().Error("rpc: failed to get client", zap.Error(err))

		return resp, err
	}

	//	Call register
	err = client.Call("AuthService.Register", req, &resp)
	if err != nil {
		if !errs.RPCErrorIs(err, errs.ErrUserAlreadyExists) {
			//	If error is not ErrUserAlreadyExists then log error
			logger.Get().Error("rpc: failed to call AuthService.Register", zap.Error(err))
		}

		return resp, err
	}

	return resp, nil
}

func (s *authService) LoginUser(req dto.LoginUserRequest) (rpcdto.LoginUserResponse, error) {
	var resp rpcdto.LoginUserResponse

	//	Validate request
	if err := s.validate.Struct(req); err != nil {
		return resp, s.validate.MapError(err)
	}

	//	Get RPC Client
	client, err := s.getClient()
	if err != nil {
		logger.Get().Error("rpc: failed to get client", zap.Error(err))

		return resp, err
	}

	//	Call login
	err = client.Call("AuthService.Login", req, &resp)
	if err != nil {
		if !errs.RPCErrorIs(err, errs.ErrInvalidCredentials) {
			//	If error is not ErrInvalidCredentials then log error
			logger.Get().Error("rpc: failed to call AuthService.Login", zap.Error(err))
		}

		return resp, err
	}

	return resp, nil
}

func (s *authService) LogoutUser() (rpcdto.LogoutUserResponse, error) {
	var resp rpcdto.LogoutUserResponse

	//	Get RPC Client
	client, err := s.getClient()
	if err != nil {
		logger.Get().Error("rpc: failed to get client", zap.Error(err))

		return resp, err
	}

	//	Call logout
	err = client.Call("AuthService.Logout", 0, &resp)
	if err != nil {
		logger.Get().Error("rpc: failed to call AuthService.Logout", zap.Error(err))

		return resp, err
	}

	return resp, nil
}

func (s *authService) GetUserFromSID(sid string) (rpcdto.GetUserFromSIDResponse, error) {
	var resp rpcdto.GetUserFromSIDResponse

	//	Get RPC Client
	client, err := s.getClient()
	if err != nil {
		logger.Get().Error("rpc: failed to get client", zap.Error(err))

		return resp, err
	}

	//	Call GetUserFromSID
	err = client.Call("AuthService.GetUserFromSID", sid, &resp)
	if err != nil {
		logger.Get().Error("rpc: failed to call AuthService.GetUserFromSID", zap.Error(err))

		return resp, err
	}

	return resp, nil
}

func (s *authService) getClient() (*rpc.Client, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.client == nil {
		client, err := dialAuthServer()
		if err != nil {
			return nil, err
		}

		s.client = client
	} else {
		//	Try to ping
		var pong bool
		if err := s.client.Call("PingService.Ping", 0, &pong); err != nil {
			//	Client disconnected -> reconnect
			client, err := dialAuthServer()
			if err != nil {
				return nil, err
			}

			s.client = client
		}
	}

	return s.client, nil
}

func dialAuthServer() (*rpc.Client, error) {
	client, err := rpc.DialHTTP("tcp", "auth_server:7070")
	if err != nil {
		return nil, fmt.Errorf("rpc: failed to dial auth server: %w", err)
	}

	return client, nil
}
