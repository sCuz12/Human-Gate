package apikeys

import "errors"

var (
	ErrMissingAPIKey      = errors.New("missing api key")
	ErrInvalidAPIKey      = errors.New("invalid api key")
	ErrAPIKeyUnauthorized = errors.New("api key missing required scope")
)
