package sdk

import (
	"github.com/gin-gonic/gin"
)

type EndorHandlerFunc[T any] func(*EndorContext[T])

type EndorServiceMethod interface {
	Register(route *gin.RouterGroup, path string, microserviceId string)
	GetOptions() EndorMethodOptions
}

type EndorMethodOptions struct {
	Public     bool
	MethodType string
}

type EndorService struct {
	Resource    string
	Description string
	Methods     map[string]EndorServiceMethod
	Priority    *int

	// optionals
	Version string
}

func NewMethod[T any](handlers ...EndorHandlerFunc[T]) EndorServiceMethod {
	return NewConfigurableMethod(EndorMethodOptions{}, handlers...)
}

func NewConfigurableMethod[T any](options EndorMethodOptions, handlers ...EndorHandlerFunc[T]) EndorServiceMethod {
	if options.MethodType == "GET" {
		return &endorServiceMethodImpl[T]{handlers: handlers, options: options}
	}
	h := []EndorHandlerFunc[T]{ValidationHandler[T]}
	h = append(h, handlers...)
	if options.MethodType == "" {
		options.MethodType = "POST"
	}
	return &endorServiceMethodImpl[T]{handlers: h, options: options}
}

type endorServiceMethodImpl[T any] struct {
	handlers []EndorHandlerFunc[T]
	options  EndorMethodOptions
}

func (m *endorServiceMethodImpl[T]) Register(group *gin.RouterGroup, path string, microserviceId string) {
	callback := func(c *gin.Context) {
		session := Session{
			Id:       c.GetHeader("X-User-ID"),
			Username: c.GetHeader("X-User-Session"),
		}
		ec := &EndorContext[T]{
			MicroServiceId: microserviceId,
			Index:          -1,
			GinContext:     c,
			Handlers:       m.handlers,
			Data:           make(map[string]interface{}),
			Session:        session,
		}
		ec.Next()
	}
	if m.options.MethodType == "POST" {
		group.POST(path, callback)
	}
	if m.options.MethodType == "GET" {
		group.GET(path, callback)
	}
}

func (m *endorServiceMethodImpl[T]) GetOptions() EndorMethodOptions {
	return m.options
}
