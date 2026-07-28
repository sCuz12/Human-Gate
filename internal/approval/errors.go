package approval

import "errors"

var (
	ErrInvalidRequest = errors.New("invalid approval request")
	ErrForbidden      = errors.New("user cannot access approval request")
	ErrNotFound       = errors.New("approval request not found")
	ErrResolved       = errors.New("approval request already resolved")
	ErrExpired        = errors.New("approval request expired")
)
