package service

import (
	"errors"
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
		var rpcErr rpc.ServerError
		if errors.As(err, &rpcErr) && rpcErr.Error() != errs.ErrUserAlreadyExists.Error() {
			//	If error is not ErrUserAlreadyExists then log error
			logger.Get().Error("rpc: failed to call AuthService.Register", zap.Error(err))
		}

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
		logger.Get().Error("rpc: failed to dial auth server", zap.Error(err))

		return nil, err
	}

	return client, nil
}
