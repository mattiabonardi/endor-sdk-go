package sdk

import (
	"fmt"
	"net/http"
)

type EndorError struct {
	StatusCode  int
	Message     string
	InternalErr error
}

func (e *EndorError) Error() string {
	return fmt.Sprintf("%s: %v", e.Message, e.InternalErr)
}

func (e *EndorError) Unwrap() error {
	return e.InternalErr
}

// Factories
func NewBadRequestError(msg string, err error) *EndorError {
	return &EndorError{StatusCode: http.StatusBadRequest, Message: msg, InternalErr: err}
}

func NewNotFoundError(msg string, err error) *EndorError {
	return &EndorError{StatusCode: http.StatusNotFound, Message: msg, InternalErr: err}
}

func NewInternalServerError(msg string, err error) *EndorError {
	return &EndorError{StatusCode: http.StatusInternalServerError, Message: msg, InternalErr: err}
}

func NewConfictErorr(msg string, err error) *EndorError {
	return &EndorError{StatusCode: http.StatusConflict, Message: msg, InternalErr: err}
}

func NewForbiddenErorr(msg string, err error) *EndorError {
	return &EndorError{StatusCode: http.StatusForbidden, Message: msg, InternalErr: err}
}

func NewUnauthorizedErorr(msg string, err error) *EndorError {
	return &EndorError{StatusCode: http.StatusUnauthorized, Message: msg, InternalErr: err}
}

func NewGenericError(msg string, status int, err error) *EndorError {
	return &EndorError{StatusCode: status, Message: msg, InternalErr: err}
}
