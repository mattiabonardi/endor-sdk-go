package models

// Response with Generics
type Response[T any] struct {
	Messages []Message `json:"messages"`
	Data     T         `json:"data"`
}

// ResponseBuilder with Generics
type ResponseBuilder[T any] struct {
	response Response[T]
}

// NewResponseBuilder initializes ResponseBuilder with generics
func NewResponseBuilder[T any]() *ResponseBuilder[T] {
	return &ResponseBuilder[T]{
		response: Response[T]{
			Messages: []Message{},
		},
	}
}

// NewResponseBuilder initializes ResponseBuilder with generics
func NewDefaultResponseBuilder() *ResponseBuilder[map[string]any] {
	return &ResponseBuilder[map[string]any]{
		response: Response[map[string]any]{
			Messages: []Message{},
		},
	}
}

func (h *ResponseBuilder[T]) AddMessage(message Message) *ResponseBuilder[T] {
	h.response.Messages = append(h.response.Messages, message)
	return h
}

func (h *ResponseBuilder[T]) AddData(data T) *ResponseBuilder[T] {
	h.response.Data = data
	return h
}

func (h *ResponseBuilder[T]) Build() *Response[T] {
	return &h.response
}
