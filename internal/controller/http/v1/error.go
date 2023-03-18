package v1

import "errors"

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
	ErrAccessDenied  = errors.New(`access denied`)
	ErrBadRequest    = errors.New("bad request")
)

type response struct {
	Error string `json:"error" example:"message"`
}
