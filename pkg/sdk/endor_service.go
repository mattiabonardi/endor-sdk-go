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

func (h EndorService) GetResource() string {
	return h.Resource
}

func (h EndorService) GetResourceDescription() string {
	return h.Description
}

func (h EndorService) GetPriority() *int {
	return h.Priority
}

func (h EndorService) ToEndorService(schema Schema) EndorService {
	return h
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
				logger.ErrorWithStackTrace(err)
				c.AbortWithStatusJSON(http.StatusBadRequest, NewDefaultResponseBuilder().AddMessage(NewMessage(ResponseMessageGravityFatal, err.Error())).Build())
				return
			}
		}
		// call method
		response, err := m.handler(ec)
		if err != nil {
			var endorError *EndorError
			if errors.As(err, &endorError) {
				logger.ErrorWithStackTrace(endorError)
				c.AbortWithStatusJSON(endorError.StatusCode, NewDefaultResponseBuilder().AddMessage(NewMessage(ResponseMessageGravityFatal, endorError.Error())))
			} else {
				logger.ErrorWithStackTrace(err)
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

type EndorServiceInterface interface {
	GetResource() string
	GetResourceDescription() string
	GetPriority() *int
	ToEndorService(metadataSchema Schema) EndorService
}

type EndorHybridServiceInterface interface {
	EndorServiceInterface
	WithActions(fn func(getSchema func() RootSchema) map[string]EndorServiceAction) EndorHybridServiceInterface
	ToEndorService(metadataSchema Schema) EndorService
}

type EndorHybridSpecializedServiceInterface interface {
	EndorServiceInterface
	WithCategories(categories []EndorHybridSpecializedServiceCategoryInterface) EndorHybridSpecializedServiceInterface
	ToEndorService(metadataSchema Schema) EndorService
}

type EndorHybridSpecializedServiceCategoryInterface interface {
	GetID() string
	CreateDefaultActions(resource string, resourceDescription string, metadataSchema Schema) map[string]EndorServiceAction
}
