package request

type CreateDirectory struct {
	Path string `validate:"required,path"`
}

type GetDirectoryContent struct {
	Path string `validate:"path"`
}
