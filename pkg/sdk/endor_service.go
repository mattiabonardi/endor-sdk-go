package sdk

import (
	"errors"
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk_configuration"
)

type EndorHandlerFunc[T any, R any] func(*EndorContext[T]) (*Response[R], error)

type EndorServiceActionInterface interface {
	CreateHTTPCallback(microserviceId string, entity string, action string, category string) func(c *gin.Context)
	GetOptions() EndorServiceActionOptions
}

type EndorServiceActionOptions struct {
	Description     string
	Public          bool
	ValidatePayload bool
	InputSchema     *RootSchema
}

type EndorService struct {
	Entity            string
	EntityDescription string
	Actions           map[string]EndorServiceActionInterface
	Priority          *int
	EntitySchema      RootSchema

	// optionals
	Version string
}

func (h EndorService) GetEntity() string {
	return h.Entity
}

func (h EndorService) GetEntityDescription() string {
	return h.EntityDescription
}

func (h EndorService) GetPriority() *int {
	return h.Priority
}

func NewAction[T any, R any](handler EndorHandlerFunc[T, R], description string) EndorServiceActionInterface {
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

func NewConfigurableAction[T any, R any](options EndorServiceActionOptions, handler EndorHandlerFunc[T, R]) EndorServiceActionInterface {
	if options.InputSchema == nil {
		options.InputSchema = ResolveGenericSchema[T]()
	}
	return &endorServiceActionImpl[T, R]{handler: handler, options: options}
}

type endorServiceActionImpl[T any, R any] struct {
	handler EndorHandlerFunc[T, R]
	options EndorServiceActionOptions
}

func (m *endorServiceActionImpl[T, R]) CreateHTTPCallback(microserviceId string, entity string, action string, categoryType string) func(c *gin.Context) {
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
		configuration := sdk_configuration.GetConfig()
		logger := NewLogger(LogConfig{
			LogType: LogType(configuration.LogType),
		}, LogContext{
			UserSession: session.Id,
			UserID:      session.Username,
			Entity:      entity,
			Action:      action,
		})

		// log incoming request
		logger.Info("Incoming request")

		ec := &EndorContext[T]{
			MicroServiceId: microserviceId,
			Session:        session,
			GinContext:     c,
			Logger:         *logger,
			CategoryType:   categoryType,
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

// generic
type EndorServiceInterface interface {
	GetEntity() string
	GetEntityDescription() string
	GetPriority() *int
}

// base
type EndorBaseServiceInterface interface {
	EndorServiceInterface
	WithPriority(priority int) EndorBaseServiceInterface
	WithActions(actions map[string]EndorServiceActionInterface) EndorBaseServiceInterface
	ToEndorService() EndorService
}

// base specialized
type EndorBaseSpecializedServiceInterface interface {
	EndorServiceInterface
	WithPriority(priority int) EndorBaseSpecializedServiceInterface
	WithActions(actions map[string]EndorServiceActionInterface) EndorBaseSpecializedServiceInterface
	WithCategories(categories []EndorBaseSpecializedServiceCategoryInterface) EndorBaseSpecializedServiceInterface
	ToEndorService() EndorService
}

type EndorBaseSpecializedServiceCategoryInterface interface {
	GetID() string
	GetActions() map[string]EndorServiceActionInterface
	WithActions(actions map[string]EndorServiceActionInterface) EndorBaseSpecializedServiceCategoryInterface
}

// hybrid
type EndorHybridServiceInterface interface {
	EndorServiceInterface
	WithPriority(priority int) EndorHybridServiceInterface
	WithActions(fn func(getSchema func() RootSchema) map[string]EndorServiceActionInterface) EndorHybridServiceInterface
	ToEndorService(metadataSchema Schema) EndorService
}

// hybrid specialized
type EndorHybridSpecializedServiceInterface interface {
	EndorServiceInterface
	WithPriority(priority int) EndorHybridSpecializedServiceInterface
	WithActions(fn func(getSchema func() RootSchema) map[string]EndorServiceActionInterface) EndorHybridSpecializedServiceInterface
	WithCategories(categories []EndorHybridSpecializedServiceCategoryInterface) EndorHybridSpecializedServiceInterface
	ToEndorService(metadataSchema Schema, categoryMetadataSchemas map[string]Schema) EndorService
}

type EndorHybridSpecializedServiceCategoryInterface interface {
	GetID() string
	GetActions() func(getSchema func() RootSchema) map[string]EndorServiceActionInterface
	CreateDefaultActions(entity string, entityDescription string, metadataSchema Schema, categoryMetadataSchema Schema) map[string]EndorServiceActionInterface
}
