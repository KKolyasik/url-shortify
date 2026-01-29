package config

import "errors"

var (
	ErrInvalidNetAddress = errors.New("incorrect net address: expected host:port")
	ErrInvalidBaseURL    = errors.New("base url must include scheme and host")
	ErrInvalidFileName   = errors.New("invalid filename")
)
