package sdk

import (
	"github.com/gin-gonic/gin"
)

type EndorHandlerFunc[T any] func(*EndorContext[T])

type EndorServiceMethod interface {
	Register(route *gin.RouterGroup, path string)
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
	h := []EndorHandlerFunc[T]{ValidationHandler[T]}
	h = append(h, handlers...)
	return &endorServiceMethodImpl[T]{handlers: h}
}

type endorServiceMethodImpl[T any] struct {
	handlers []EndorHandlerFunc[T]
}

func (m *endorServiceMethodImpl[T]) Register(group *gin.RouterGroup, path string) {
	group.POST(path, func(c *gin.Context) {
		session := Session{
			Id:       c.GetHeader("X-User-ID"),
			Username: c.GetHeader("X-User-Session"),
		}
		ec := &EndorContext[T]{
			Index:      -1,
			GinContext: c,
			Handlers:   m.handlers,
			Data:       make(map[string]interface{}),
			Session:    session,
		}
		ec.Next()
	})
}
