package sdk

import (
	"fmt"
	"net/http"
)

type EndorError struct {
	StatusCode  int
	InternalErr error
}

func (e *EndorError) Error() string {
	return fmt.Sprintf("%v", e.InternalErr)
}

func (e *EndorError) Unwrap() error {
	return e.InternalErr
}

// Factories
func NewBadRequestError(err error) *EndorError {
	return &EndorError{StatusCode: http.StatusBadRequest, InternalErr: err}
}

func NewNotFoundError(err error) *EndorError {
	return &EndorError{StatusCode: http.StatusNotFound, InternalErr: err}
}

func NewInternalServerError(err error) *EndorError {
	return &EndorError{StatusCode: http.StatusInternalServerError, InternalErr: err}
}

func NewConfictError(err error) *EndorError {
	return &EndorError{StatusCode: http.StatusConflict, InternalErr: err}
}

func NewForbiddenError(err error) *EndorError {
	return &EndorError{StatusCode: http.StatusForbidden, InternalErr: err}
}

func NewUnauthorizedError(err error) *EndorError {
	return &EndorError{StatusCode: http.StatusUnauthorized, InternalErr: err}
}

func NewGenericError(status int, err error) *EndorError {
	return &EndorError{StatusCode: status, InternalErr: err}
}
