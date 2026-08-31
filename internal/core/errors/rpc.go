package errors

import (
	"errors"
	"net/rpc"
)

// RPCErrorIs returns true if target error matches
// with err rpc.ServerError, otherwise false.
func RPCErrorIs(err error, target error) bool {
	var rpcErr rpc.ServerError

	if errors.As(err, &rpcErr) && rpcErr.Error() == target.Error() {
		return true
	}

	return false
}
