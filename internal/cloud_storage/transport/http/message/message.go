package message

var (
	ErrInvalidRequestBody = "Request body is invalid."
	ErrInvalidCredentials = "Username or password is invalid."
	ErrUserAlreadyExists  = "User with provided username already exists."

	ErrInternalServer = "Something went wrong, please try again later."
	ErrNotFound       = "Requested resource was not found."
)
