package sdk

import (
	"errors"
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/mattiabonardi/endor-sdk-go/internal/configuration"
)

type EndorHandlerFunc[T any, R any] func(*EndorContext[T]) (*Response[R], error)

type EndorServiceAction interface {
	CreateHTTPCallback(microserviceId string, resource string, action string) func(c *gin.Context)
	GetOptions() EndorServiceActionOptions
}

type EndorServiceActionOptions struct {
	Description     string
	Public          bool
	ValidatePayload bool
	InputSchema     *RootSchema
}

type EndorService struct {
	Resource         string
	Description      string
	Methods          map[string]EndorServiceAction
	Priority         *int
	ResourceMetadata bool

	// optionals
	Version string
}

func NewAction[T any, R any](handler EndorHandlerFunc[T, R], description string) EndorServiceAction {
	options := EndorServiceActionOptions{
		Description:     description,
		Public:          false,
		ValidatePayload: true,
		InputSchema:     nil,
	}
	// resolve input params dynamically
	options.InputSchema = ResolveGenericSchema[T]()
	return NewConfigurableAction(options, handler)
}

func NewConfigurableAction[T any, R any](options EndorServiceActionOptions, handler EndorHandlerFunc[T, R]) EndorServiceAction {
	if options.InputSchema == nil {
		options.InputSchema = ResolveGenericSchema[T]()
	}
	return &endorServiceActionImpl[T, R]{handler: handler, options: options}
}

type endorServiceActionImpl[T any, R any] struct {
	handler EndorHandlerFunc[T, R]
	options EndorServiceActionOptions
}

func (m *endorServiceActionImpl[T, R]) CreateHTTPCallback(microserviceId string, resource string, action string) func(c *gin.Context) {
	return func(c *gin.Context) {
		development := false
		if c.GetHeader("x-development") == "true" {
			development = true
		}
		session := Session{
			Id:          c.GetHeader("x-user-session"),
			Username:    c.GetHeader("x-user-id"),
			Development: development,
		}

		// logger
		configuration := configuration.GetConfig()
		logger := NewLogger(Config{
			LogType: LogType(configuration.LogType),
		}, LogContext{
			UserSession: session.Id,
			UserID:      session.Username,
			Resource:    resource,
			Action:      action,
		})

		// log incoming request
		logger.Info("Incoming request")

		ec := &EndorContext[T]{
			MicroServiceId: microserviceId,
			Session:        session,
			GinContext:     c,
			Logger:         *logger,
		}
		var t T
		if m.options.ValidatePayload && reflect.TypeOf(t) != reflect.TypeOf(NoPayload{}) {
			if err := c.ShouldBindJSON(&ec.Payload); err != nil {
				//TODO: implements JSON Schema validation
				logger.Error(err.Error())
				c.AbortWithStatusJSON(http.StatusBadRequest, NewDefaultResponseBuilder().AddMessage(NewMessage(ResponseMessageGravityFatal, err.Error())).Build())
				return
			}
		}
		// call method
		response, err := m.handler(ec)
		if err != nil {
			var endorError *EndorError
			if errors.As(err, &endorError) {
				logger.Error(endorError.Error())
				c.AbortWithStatusJSON(endorError.StatusCode, NewDefaultResponseBuilder().AddMessage(NewMessage(ResponseMessageGravityFatal, endorError.Error())))
			} else {
				logger.Error(err.Error())
				c.AbortWithStatusJSON(http.StatusInternalServerError, NewDefaultResponseBuilder().AddMessage(NewMessage(ResponseMessageGravityFatal, err.Error())))
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

type EndorHybridServiceCategory interface {
	GetID() string
	CreateDefaultActions(resource string, resourceDescription string, metadataSchema Schema) map[string]EndorServiceAction
}

type EndorHybridService interface {
	GetResource() string
	GetResourceDescription() string
	GetPriority() *int
	WithCategories(categories []EndorHybridServiceCategory) EndorHybridService
	WithActions(fn func(getSchema func() RootSchema) map[string]EndorServiceAction) EndorHybridService
	ToEndorService(metadataSchema Schema) EndorService
}
