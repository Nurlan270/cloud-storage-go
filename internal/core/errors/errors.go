package errors

import "errors"

var (
	ErrUserAlreadyExists = errors.New("user already exists")
)

type ErrValidation struct {
	Message string
}

func (e ErrValidation) Error() string {
	return e.Message
}
