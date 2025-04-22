package models

type EndorHandlerFunc func(*EndorContext)

type EndorService struct {
	Resource string
	Methods  map[string][]EndorHandlerFunc
}

func NewEndorService(resource string) *EndorService {
	return &EndorService{Resource: resource, Methods: make(map[string][]EndorHandlerFunc)}
}

func (s *EndorService) Handle(key string, handlers ...EndorHandlerFunc) {
	s.Methods[key] = handlers
}
