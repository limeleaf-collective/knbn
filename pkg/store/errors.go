package store

import "errors"

var (
	ErrInvalid   = errors.New("invalid")
	ErrDuplicate = errors.New("duplicate")
	ErrNotFound  = errors.New("not found")
	ErrStorage   = errors.New("storage")
)
