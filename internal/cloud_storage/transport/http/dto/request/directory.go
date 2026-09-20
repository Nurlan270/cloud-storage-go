package request

type CreateDirectory struct {
	Path string `validate:"required,path"`
}
