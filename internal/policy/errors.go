package policy

import "errors"

var (
	ErrInvalidRequest = errors.New("invalid policy request")
	ErrForbidden      = errors.New("user cannot manage policies")
	ErrNotFound       = errors.New("policy not found")
)
