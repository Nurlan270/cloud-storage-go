package errors

import (
	"errors"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")

	ErrInvalidCredentials = errors.New("invalid credentials")

	ErrSessionInvalid = errors.New("session invalid")

	ErrResourceNotFound      = errors.New("resource not found")
	ErrResourceAlreadyExists = errors.New("resource already exists")

	ErrDirectoryNotFound       = errors.New("directory not found")
	ErrDirectoryAlreadyExists  = errors.New("directory already exists")
	ErrParentDirectoryNotFound = errors.New("parent directory does not exist")
)

type ErrValidation struct {
	Message string
}

func (e ErrValidation) Error() string {
	return e.Message
}
