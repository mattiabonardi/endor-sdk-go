package sdk

import (
	"fmt"
	"net/http"
)

type EndorError struct {
	StatusCode      int
	InternalErr     error
	TranslationKey  string
	TranslationArgs map[string]any
}

func (e *EndorError) Error() string {
	return fmt.Sprintf("%s", e.InternalErr.Error())
}

func (e *EndorError) Unwrap() error {
	return e.InternalErr
}

// WithTranslation attaches an i18n translation key and optional named interpolation args.
// The translation is resolved at response time using the request locale.
func (e *EndorError) WithTranslation(key string, args map[string]any) *EndorError {
	e.TranslationKey = key
	e.TranslationArgs = args
	return e
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

func NewConflictError(err error) *EndorError {
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
