package pattern

import "errors"

var (
	ErrNotFound     = errors.New("pattern not found")
	ErrInvalidInput = errors.New("invalid pattern input")
)
