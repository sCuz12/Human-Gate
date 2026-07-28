package identity

import "errors"

var (
	ErrUnauthorized       = errors.New("unauthorized")
	ErrForbidden          = errors.New("forbidden")
	ErrInvalidUserID      = errors.New("invalid user id")
	ErrInvalidWorkspaceID = errors.New("invalid workspace id")
)
