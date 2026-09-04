package service

import (
	"errors"
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

	Close() error
}

type authService struct {
	validate *validator.Validate

	clientMu sync.RWMutex
	client   *rpc.Client

	//	Used only to reconnect client
	reconnMu sync.Mutex
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

	//	Call RPC
	if err := s.Call("Register", req, &resp); err != nil {
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

	//	Call RPC
	if err := s.Call("Login", req, &resp); err != nil {
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

	//	Call RPC
	if err := s.Call("Logout", 0, &resp); err != nil {
		logger.Get().Error("rpc: failed to call AuthService.Logout", zap.Error(err))

		return resp, err
	}

	return resp, nil
}

func (s *authService) GetUserFromSID(sid string) (rpcdto.GetUserFromSIDResponse, error) {
	var resp rpcdto.GetUserFromSIDResponse

	//	Call GetUserFromSID
	if err := s.Call("GetUserFromSID", sid, &resp); err != nil {
		if !errs.RPCErrorIs(err, errs.ErrSessionNotFound) && !errs.RPCErrorIs(err, errs.ErrSessionExpired) {
			logger.Get().Error("rpc: failed to call AuthService.GetUserFromSID", zap.Error(err))
		}

		return resp, err
	}

	return resp, nil
}

func (s *authService) Close() error {
	s.clientMu.RLock()
	client := s.client
	s.clientMu.RUnlock()

	if client == nil {
		return nil
	}

	s.clientMu.Lock()
	defer s.clientMu.Unlock()

	if s.client == nil {
		return nil
	}

	err := s.client.Close()
	s.client = nil

	return err
}

// Call is a wrapper around (rpc.Client).Call method
// which automatically gets client and reconnects on server error.
func (s *authService) Call(method string, req any, resp any) error {
	//	Get RPC Client
	client, err := s.getClient()
	if err != nil {
		return fmt.Errorf("rpc: failed to get client: %w", err)
	}

	serviceMethod := fmt.Sprintf("AuthService.%s", method)

	//	Call RPC method
	callErr := client.Call(serviceMethod, req, resp)
	if callErr == nil {
		return nil
	}

	//	Reconnect if there's connection error
	if errors.Is(callErr, rpc.ErrShutdown) {
		if reconnectErr := s.reconnect(client); reconnectErr != nil {
			return fmt.Errorf("rpc: failed to reconnect: %w", reconnectErr)
		}
	} else {
		return callErr
	}

	newClient, err := s.getClient()
	if err != nil {
		return fmt.Errorf("rpc: failed to get client: %w", err)
	}

	//	Call RPC method again after reconnecting with new client
	return newClient.Call(serviceMethod, req, resp)
}

func (s *authService) getClient() (*rpc.Client, error) {
	s.clientMu.RLock()
	client := s.client
	s.clientMu.RUnlock()

	if client != nil {
		return client, nil
	}

	s.clientMu.Lock()
	defer s.clientMu.Unlock()

	if s.client != nil {
		return s.client, nil
	}

	//	Client is nil -> connect and set new fresh client
	client, err := dialConn()
	if err != nil {
		return nil, err
	}

	s.client = client

	return s.client, nil
}

func (s *authService) reconnect(failedClient *rpc.Client) error {
	s.reconnMu.Lock()
	defer s.reconnMu.Unlock()

	s.clientMu.RLock()
	currentClient := s.client
	s.clientMu.RUnlock()

	if currentClient != failedClient {
		return nil
	}

	newClient, err := dialConn()
	if err != nil {
		return fmt.Errorf("rpc: failed to dial new client: %w", err)
	}

	s.clientMu.Lock()
	if s.client != failedClient {
		s.clientMu.Unlock()

		_ = newClient.Close()

		return nil
	}

	s.client = newClient
	s.clientMu.Unlock()

	var failedClientErr error
	if failedClient != nil {
		failedClientErr = failedClient.Close()
	}

	return failedClientErr
}

// dialConn dials the auth server and returns *rpc.Client.
func dialConn() (*rpc.Client, error) {
	c, err := rpc.DialHTTP("tcp", "auth_server:7070")
	if err != nil {
		return nil, fmt.Errorf("rpc: failed to dial auth server: %w", err)
	}

	return c, nil
}
