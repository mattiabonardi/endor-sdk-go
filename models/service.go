package models

type EndorServiceMethodFunc[I any, O any] func(context EndorServiceContext[I]) (Response[O], error)

type EndorServiceMethodHandler[I any, O any] struct {
	Payload     I
	HandlerFunc EndorServiceMethodFunc[I, O]
	Public      bool
}

type EndorServiceContext[T any] struct {
	Session Session
	Payload T
}

type Session struct {
	Id    string `json:"id"`
	User  string `json:"user"`
	Email string `json:"email"`
	App   string `json:"app"`
}

type EndorService struct {
	Resource string
	Methods  map[string]EndorServiceMethodHandler[any, any]
}

func (h EndorService) AddMethod(name string, method EndorServiceMethodHandler[any, any]) {
	h.Methods[name] = method
}

type ResourceService[T any] interface {
	Instance(id string, options IntanceOptions) (T, error)
	List(options ListOptions) ([]T, error)
	Create(resource T) (T, error)
	Update(id string, resource T) (T, error)
	Delete(id string) error
}
