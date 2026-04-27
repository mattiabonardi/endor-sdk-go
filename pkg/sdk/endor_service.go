package sdk

import (
	"errors"
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk_configuration"
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk_i18n"
)

type EndorHandlerFunc[T any, R any] func(*EndorContext[T]) (*Response[R], error)

type EndorHandlerActionInterface interface {
	CreateHTTPCallback(microserviceId string, entity string, action string, category string, session Session, container EndorDIContainerInterface) func(c *gin.Context)
	GetOptions() EndorHandlerActionOptions
}

type EndorHandlerActionOptions struct {
	Description           string
	Public                bool
	SkipPayloadValidation bool
	InputSchema           *RootSchema
}

type EndorHandler struct {
	Entity              string
	EntityTitle         string
	EntityDescription   string
	Actions             map[string]EndorHandlerActionInterface
	Priority            *int
	EntitySchema        RootSchema
	RepositoryFactories map[string]RepositoryFactory

	// optionals
	Version string
}

func (h EndorHandler) GetEntity() string {
	return h.Entity
}

func (h EndorHandler) GetEntityTitle() string {
	return h.EntityTitle
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

func (m *endorHandlerActionImpl[T, R]) CreateHTTPCallback(microserviceId string, entity string, action string, categoryType string, session Session, container EndorDIContainerInterface) func(c *gin.Context) {
	return func(c *gin.Context) {
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
			Locale:         sdk_i18n.NormalizeLocale(c.GetHeader("Accept-Language")),
			GinContext:     c,
			Logger:         *logger,
			CategoryType:   categoryType,
			DIContainer:    container,
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
				message := endorError.Error()
				if endorError.TranslationKey != "" {
					message = sdk_i18n.T(ec.Locale, endorError.TranslationKey, endorError.TranslationArgs)
				}
				c.JSON(endorError.StatusCode, NewDefaultResponseBuilder().AddMessage(NewMessage(ResponseMessageGravityFatal, message)).Build())
			} else {
				logger.ErrorWithStackTrace(err)
				c.JSON(http.StatusInternalServerError, NewDefaultResponseBuilder().AddMessage(NewMessage(ResponseMessageGravityFatal, err.Error())).Build())
			}
		} else {
			response.ResolveTranslations(ec.Locale)
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
	GetEntityTitle() string
	GetEntityDescription() string
	GetPriority() *int
	GetSchema() *RootSchema
}

// base
type EndorBaseHandlerInterface interface {
	EndorHandlerInterface
	WithExtendedDescription(description string) EndorBaseHandlerInterface
	WithPriority(priority int) EndorBaseHandlerInterface
	WithActions(actions map[string]EndorHandlerActionInterface) EndorBaseHandlerInterface
	WithRepository(fn RepositoryFactory) EndorBaseHandlerInterface
	ToEndorHandler() EndorHandler
}

// base specialized
type EndorBaseSpecializedHandlerInterface interface {
	EndorHandlerInterface
	WithExtendedDescription(description string) EndorBaseSpecializedHandlerInterface
	WithPriority(priority int) EndorBaseSpecializedHandlerInterface
	WithActions(actions map[string]EndorHandlerActionInterface) EndorBaseSpecializedHandlerInterface
	WithCategories(categories []EndorBaseSpecializedHandlerCategoryInterface) EndorBaseSpecializedHandlerInterface
	WithRepository(fn RepositoryFactory) EndorBaseSpecializedHandlerInterface
	GetCategories() []Category
	ToEndorHandler() EndorHandler
}

type EndorBaseSpecializedHandlerCategoryInterface interface {
	GetID() string
	GetTitle() string
	GetDescription() string
	GetSchema() string
	GetActions() map[string]EndorHandlerActionInterface
	GetRepository() RepositoryFactory
	WithExtendedDescription(description string) EndorBaseSpecializedHandlerCategoryInterface
	WithActions(actions map[string]EndorHandlerActionInterface) EndorBaseSpecializedHandlerCategoryInterface
	WithRepository(fn RepositoryFactory) EndorBaseSpecializedHandlerCategoryInterface
}

// hybrid
type EndorHybridHandlerInterface interface {
	EndorHandlerInterface
	WithExtendedDescription(description string) EndorHybridHandlerInterface
	WithPriority(priority int) EndorHybridHandlerInterface
	WithActions(fn func(getSchema func() RootSchema) map[string]EndorHandlerActionInterface) EndorHybridHandlerInterface
	ToEndorHandler(metadataSchema RootSchema) EndorHandler
}

// hybrid specialized
type EndorHybridSpecializedHandlerInterface interface {
	EndorHandlerInterface
	WithExtendedDescription(description string) EndorHybridSpecializedHandlerInterface
	WithPriority(priority int) EndorHybridSpecializedHandlerInterface
	WithActions(fn func(getSchema func() RootSchema) map[string]EndorHandlerActionInterface) EndorHybridSpecializedHandlerInterface
	WithHybridCategories(categories []EndorHybridSpecializedHandlerCategoryInterface) EndorHybridSpecializedHandlerInterface
	GetHybridCategories() []HybridCategory
	ToEndorHandler(metadataSchema RootSchema, categoryMetadataSchemas map[string]RootSchema, additionalCategories []DynamicCategory) EndorHandler
}

type EndorHybridSpecializedHandlerCategoryInterface interface {
	GetID() string
	GetTitle() string
	GetDescription() string
	GetSchema() string
	GetActions() func(getSchema func() RootSchema) map[string]EndorHandlerActionInterface
	GetRepository() RepositoryFactory
	WithExtendedDescription(description string) EndorHybridSpecializedHandlerCategoryInterface
	WithActions(actionFn func(getSchema func() RootSchema) map[string]EndorHandlerActionInterface) EndorHybridSpecializedHandlerCategoryInterface
	CreateDefaultActions(entity string, entityDescription string, metadataSchema RootSchema, categoryMetadataSchema RootSchema) map[string]EndorHandlerActionInterface
}
