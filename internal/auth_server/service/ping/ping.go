package ping

type Service interface {
	Ping(_ int, pong *bool) error
}

type service struct{}

func NewService() Service {
	return &service{}
}

// Ping is used just to determine whether client is still alive or not
// It does not return any information beside pong.
func (s *service) Ping(_ int, pong *bool) error {
	*pong = true
	return nil
}
