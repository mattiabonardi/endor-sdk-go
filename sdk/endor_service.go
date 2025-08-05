package sdk

import (
	"errors"
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
)

type EndorHandlerFunc[T any, R any] func(*EndorContext[T]) (*Response[R], error)

type EndorServiceAction interface {
	CreateHTTPCallback(microserviceId string) func(c *gin.Context)
	GetOptions() EndorServiceActionOptions
}

type EndorServiceActionOptions struct {
	Description     string
	Public          bool
	ValidatePayload bool
	InputSchema     *RootSchema
}

type EndorService struct {
	Resource    string
	Description string
	Methods     map[string]EndorServiceAction
	Priority    *int

	// optionals
	Version string
}

func NewAction[T any, R any](handler EndorHandlerFunc[T, R], description string) EndorServiceAction {
	options := EndorServiceActionOptions{
		Description:     description,
		Public:          false,
		ValidatePayload: true,
	}
	// resolve input params dynamically
	var zeroT T
	tType := reflect.TypeOf(zeroT)
	if tType.Kind() == reflect.Ptr {
		tType = tType.Elem()
	}
	// convert type to schema
	if tType != reflect.TypeOf(NoPayload{}) {
		options.InputSchema = NewSchemaByType(tType)
	}
	return NewConfigurableAction(options, handler)
}

func NewConfigurableAction[T any, R any](options EndorServiceActionOptions, handler EndorHandlerFunc[T, R]) EndorServiceAction {
	return &endorServiceActionImpl[T, R]{handler: handler, options: options}
}

type endorServiceActionImpl[T any, R any] struct {
	handler EndorHandlerFunc[T, R]
	options EndorServiceActionOptions
}

func (m *endorServiceActionImpl[T, R]) CreateHTTPCallback(microserviceId string) func(c *gin.Context) {
	return func(c *gin.Context) {
		session := Session{
			Id:       c.GetHeader("x-user-session"),
			Username: c.GetHeader("x-user-id"),
		}
		ec := &EndorContext[T]{
			MicroServiceId: microserviceId,
			Session:        session,
			GinContext:     c,
		}
		var t T
		if m.options.ValidatePayload && reflect.TypeOf(t) != reflect.TypeOf(NoPayload{}) {
			if err := c.ShouldBindJSON(&ec.Payload); err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, NewDefaultResponseBuilder().AddMessage(NewMessage(Fatal, err.Error())).Build())
				return
			}
		}
		// call method
		response, err := m.handler(ec)
		if err != nil {
			var endorError *EndorError
			if errors.As(err, &endorError) {
				c.AbortWithStatusJSON(endorError.StatusCode, NewDefaultResponseBuilder().AddMessage(NewMessage(Fatal, endorError.Error())))
			} else {
				c.AbortWithStatusJSON(http.StatusInternalServerError, NewDefaultResponseBuilder().AddMessage(NewMessage(Fatal, err.Error())))
			}
		} else {
			c.Header("x-endor-microservice", microserviceId)
			c.JSON(http.StatusOK, response)
		}
	}
}

func (m *endorServiceActionImpl[T, R]) GetOptions() EndorServiceActionOptions {
	return m.options
}
