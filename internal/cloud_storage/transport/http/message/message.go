package message

var (
	ErrInvalidRequestBody = "Request body is invalid."
	ErrInvalidCredentials = "Username or password is invalid."
	ErrUserAlreadyExists  = "User with provided username already exists."

	ErrInternalServer = "Something went wrong, please try again later."
	ErrNotFound       = "Requested resource was not found."
	ErrForbidden      = "You are not allowed to access this resource."
	ErrUnauthorized   = "You are not authorized to access this resource."
)
