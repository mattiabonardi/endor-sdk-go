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

	// optionals
	Version string
	Apps    []string
}

func NewMethod[T any](handlers ...EndorHandlerFunc[T]) EndorServiceMethod {
	return &endorServiceMethodImpl[T]{handlers: handlers}
}

type endorServiceMethodImpl[T any] struct {
	handlers []EndorHandlerFunc[T]
}

func (m *endorServiceMethodImpl[T]) Register(group *gin.RouterGroup, path string) {
	group.POST(path, func(c *gin.Context) {
		ec := &EndorContext[T]{
			Index:      -1,
			GinContext: c,
			Handlers:   m.handlers,
			Data:       make(map[string]interface{}),
			Session: Session{
				App: c.Param("app"),
			},
		}
		ec.Next()
	})
}
