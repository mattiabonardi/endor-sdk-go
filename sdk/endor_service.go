package sdk

import (
	"errors"
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
)

type EndorHandlerFunc[T any, R any] func(*EndorContext[T]) (*Response[R], error)

type EndorServiceAction interface {
	CreateHTTPCallback(microserviceId string, eventBus EventBus) func(c *gin.Context)
	GetOptions() EndorServiceActionOptions
	AddEvent(eventDef *EventDefinition) EndorServiceAction
	GetEvent(name string) (*EventDefinition, bool)
}

type EndorServiceActionOptions struct {
	Description     string
	Public          bool
	ValidatePayload bool
	InputSchema     *RootSchema
	Events          map[string]*EventDefinition
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
		Events:          make(map[string]*EventDefinition),
	}
	// resolve input params dynamically
	options.InputSchema = resolveInputSchema[T]()
	return NewConfigurableAction(options, handler)
}

func NewConfigurableAction[T any, R any](options EndorServiceActionOptions, handler EndorHandlerFunc[T, R]) EndorServiceAction {
	if options.InputSchema == nil {
		options.InputSchema = resolveInputSchema[T]()
	}
	if options.Events == nil {
		options.Events = make(map[string]*EventDefinition)
	}
	return &endorServiceActionImpl[T, R]{handler: handler, options: options}
}

type endorServiceActionImpl[T any, R any] struct {
	handler EndorHandlerFunc[T, R]
	options EndorServiceActionOptions
}

func (m *endorServiceActionImpl[T, R]) CreateHTTPCallback(microserviceId string, eventBus EventBus) func(c *gin.Context) {
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
		// Recupera categoryID dal context Gin se presente
		var categoryID *string
		if catID, exists := c.Get("categoryID"); exists {
			if catIDStr, ok := catID.(string); ok {
				categoryID = &catIDStr
			}
		}

		ec := &EndorContext[T]{
			MicroServiceId:  microserviceId,
			Session:         session,
			EventBus:        eventBus,
			AvailableEvents: m.options.Events,
			CategoryID:      categoryID,
			GinContext:      c,
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

func (m *endorServiceActionImpl[T, R]) AddEvent(eventDef *EventDefinition) EndorServiceAction {
	if m.options.Events == nil {
		m.options.Events = make(map[string]*EventDefinition)
	}
	m.options.Events[eventDef.Name] = eventDef
	return m
}

func (m *endorServiceActionImpl[T, R]) GetEvent(name string) (*EventDefinition, bool) {
	if m.options.Events == nil {
		return nil, false
	}
	eventDef, exists := m.options.Events[name]
	return eventDef, exists
}

func resolveInputSchema[T any]() *RootSchema {
	var zeroT T
	tType := reflect.TypeOf(zeroT)
	if tType.Kind() == reflect.Ptr {
		tType = tType.Elem()
	}
	// convert type to schema
	if tType != nil && tType != reflect.TypeOf(NoPayload{}) {
		return NewSchemaByType(tType)
	}
	return nil
}

// NewActionWithEvents crea una nuova action con eventi predefiniti
func NewActionWithEvents[T any, R any](
	handler EndorHandlerFunc[T, R],
	description string,
	events ...*EventDefinition,
) EndorServiceAction {
	action := NewAction(handler, description)
	for _, event := range events {
		action.AddEvent(event)
	}
	return action
}

// NewConfigurableActionWithEvents crea una nuova action configurabile con eventi predefiniti
func NewConfigurableActionWithEvents[T any, R any](
	options EndorServiceActionOptions,
	handler EndorHandlerFunc[T, R],
	events ...*EventDefinition,
) EndorServiceAction {
	action := NewConfigurableAction(options, handler)
	for _, event := range events {
		action.AddEvent(event)
	}
	return action
}
