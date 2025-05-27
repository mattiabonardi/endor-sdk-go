package sdk

import "errors"

var ErrForbidden = errors.New("forbidden")
var ErrNotFound = errors.New("not found")
var ErrInternalServerError = errors.New("internal server error")
var ErrAlreadyExists = errors.New("already exist")
