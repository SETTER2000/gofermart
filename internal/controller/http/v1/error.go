package v1

import "errors"

var (
	ErrNotFound             = errors.New("not found")
	ErrAlreadyExists        = errors.New("already exists")
	ErrAlreadyBeenUploaded  = errors.New("the order number has already been uploaded by another user")
	ErrAccessDenied         = errors.New(`access denied`)
	ErrBadRequest           = errors.New("bad request")
	ErrNotDataAnswer        = errors.New("no data to answer")
	ErrBadFormat            = errors.New("invalid request format")
	ErrBadFormatOrder       = errors.New("invalid order number format")
	ErrIncorrectLoginOrPass = errors.New("incorrect login or password")
)
