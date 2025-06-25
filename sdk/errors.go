package sdk

import "errors"

var ErrBadRequest = errors.New("bad request")
var ErrUnauthorized = errors.New("unauthorized")
var ErrForbidden = errors.New("forbidden")
var ErrNotFound = errors.New("not found")
var ErrInternalServerError = errors.New("internal server error")
var ErrAlreadyExists = errors.New("already exist")
