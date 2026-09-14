package message

var (
	ErrInvalidCredentials = "Username or password is invalid."

	ErrUserAlreadyExists = "User with provided username already exists."

	ErrResourceNotFound      = "Resource was not found."
	ErrResourceAlreadyExists = "Resource with provided name already exists in target directory."

	ErrInternalServer  = "Something went wrong, please try again later."
	ErrBadRequest      = "Request body is invalid."
	ErrNotFound        = "Requested resource was not found."
	ErrForbidden       = "You are not allowed to access this resource."
	ErrUnauthorized    = "You are not authorized to access this resource."
	ErrTooManyRequests = "Too many requests, slow down and try again later."
)
