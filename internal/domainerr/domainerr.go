package domainerr

import "errors"

var (
	ErrURLAlreadyExists   = errors.New("url already exists")
	ErrShortCodeCollision = errors.New("short code collision")
	ErrURLDeleted         = errors.New("url deleted")
)

type URLAlreadyExistsError struct {
	ShortCode string
}

func (e *URLAlreadyExistsError) Error() string {
	return ErrURLAlreadyExists.Error()
}
