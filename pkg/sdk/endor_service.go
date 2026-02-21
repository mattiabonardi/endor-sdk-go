package sdk

import (
	"errors"
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk_configuration"
)

type EndorHandlerFunc[T any, R any] func(*EndorContext[T]) (*Response[R], error)

type EndorHandlerActionInterface interface {
	CreateHTTPCallback(microserviceId string, entity string, action string, category string) func(c *gin.Context)
	GetOptions() EndorHandlerActionOptions
}

type EndorHandlerActionOptions struct {
	Description           string
	Public                bool
	SkipPayloadValidation bool
	InputSchema           *RootSchema
}

type EndorHandler struct {
	Entity            string
	EntityDescription string
	Actions           map[string]EndorHandlerActionInterface
	Priority          *int
	EntitySchema      RootSchema

	// optionals
	Version string
}

func (h EndorHandler) GetEntity() string {
	return h.Entity
}

func (h EndorHandler) GetEntityDescription() string {
	return h.EntityDescription
}

func (h EndorHandler) GetPriority() *int {
	return h.Priority
}

func NewAction[T any, R any](handler EndorHandlerFunc[T, R], description string) EndorHandlerActionInterface {
	options := EndorHandlerActionOptions{
		Description:           description,
		Public:                false,
		SkipPayloadValidation: false,
		InputSchema:           nil,
	}
	// resolve input params dynamically
	options.InputSchema = ResolveGenericSchema[T]()
	return NewConfigurableAction(options, handler)
}

func NewConfigurableAction[T any, R any](options EndorHandlerActionOptions, handler EndorHandlerFunc[T, R]) EndorHandlerActionInterface {
	if options.InputSchema == nil {
		options.InputSchema = ResolveGenericSchema[T]()
	}
	return &endorHandlerActionImpl[T, R]{handler: handler, options: options}
}

type endorHandlerActionImpl[T any, R any] struct {
	handler EndorHandlerFunc[T, R]
	options EndorHandlerActionOptions
}

func (m *endorHandlerActionImpl[T, R]) CreateHTTPCallback(microserviceId string, entity string, action string, categoryType string) func(c *gin.Context) {
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
		if !m.options.SkipPayloadValidation && reflect.TypeOf(t) != reflect.TypeOf(NoPayload{}) {
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
				c.JSON(endorError.StatusCode, NewDefaultResponseBuilder().AddMessage(NewMessage(ResponseMessageGravityFatal, endorError.Error())).Build())
			} else {
				logger.ErrorWithStackTrace(err)
				c.JSON(http.StatusInternalServerError, NewDefaultResponseBuilder().AddMessage(NewMessage(ResponseMessageGravityFatal, err.Error())).Build())
			}
		} else {
			c.Header("x-endor-microservice", microserviceId)
			c.JSON(http.StatusOK, response)
		}
	}
}

func (m *endorHandlerActionImpl[T, R]) GetOptions() EndorHandlerActionOptions {
	return m.options
}

// generic
type EndorHandlerInterface interface {
	GetEntity() string
	GetEntityDescription() string
	GetPriority() *int
	GetSchema() *RootSchema
}

// base
type EndorBaseHandlerInterface interface {
	EndorHandlerInterface
	WithPriority(priority int) EndorBaseHandlerInterface
	WithActions(actions map[string]EndorHandlerActionInterface) EndorBaseHandlerInterface
	ToEndorHandler() EndorHandler
}

// base specialized
type EndorBaseSpecializedHandlerInterface interface {
	EndorHandlerInterface
	WithPriority(priority int) EndorBaseSpecializedHandlerInterface
	WithActions(actions map[string]EndorHandlerActionInterface) EndorBaseSpecializedHandlerInterface
	WithCategories(categories []EndorBaseSpecializedHandlerCategoryInterface) EndorBaseSpecializedHandlerInterface
	GetCategories() []Category
	ToEndorHandler() EndorHandler
}

type EndorBaseSpecializedHandlerCategoryInterface interface {
	GetID() string
	GetDescription() string
	GetSchema() string
	GetActions() map[string]EndorHandlerActionInterface
	WithActions(actions map[string]EndorHandlerActionInterface) EndorBaseSpecializedHandlerCategoryInterface
}

// hybrid
type EndorHybridHandlerInterface interface {
	EndorHandlerInterface
	WithPriority(priority int) EndorHybridHandlerInterface
	WithActions(fn func(getSchema func() RootSchema) map[string]EndorHandlerActionInterface) EndorHybridHandlerInterface
	ToEndorHandler(metadataSchema RootSchema) EndorHandler
}

// hybrid specialized
type EndorHybridSpecializedHandlerInterface interface {
	EndorHandlerInterface
	WithPriority(priority int) EndorHybridSpecializedHandlerInterface
	WithActions(fn func(getSchema func() RootSchema) map[string]EndorHandlerActionInterface) EndorHybridSpecializedHandlerInterface
	WithHybridCategories(categories []EndorHybridSpecializedHandlerCategoryInterface) EndorHybridSpecializedHandlerInterface
	GetHybridCategories() []HybridCategory
	ToEndorHandler(metadataSchema RootSchema, categoryMetadataSchemas map[string]RootSchema, additionalCategories []DynamicCategory) EndorHandler
}

type EndorHybridSpecializedHandlerCategoryInterface interface {
	GetID() string
	GetDescription() string
	GetSchema() string
	GetActions() func(getSchema func() RootSchema) map[string]EndorHandlerActionInterface
	WithActions(actionFn func(getSchema func() RootSchema) map[string]EndorHandlerActionInterface) EndorHybridSpecializedHandlerCategoryInterface
	CreateDefaultActions(entity string, entityDescription string, metadataSchema RootSchema, categoryMetadataSchema RootSchema) map[string]EndorHandlerActionInterface
}
