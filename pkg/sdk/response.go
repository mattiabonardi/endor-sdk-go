package sdk

// Response with Generics
type Response[T any] struct {
	Messages []ResponseMessage `json:"messages"`
	Data     *T                `json:"data"`
	Schema   *RootSchema       `json:"schema"`
}

// ResponseBuilder with Generics
type ResponseBuilder[T any] struct {
	response Response[T]
}

// NewResponseBuilder initializes ResponseBuilder with generics
func NewResponseBuilder[T any]() *ResponseBuilder[T] {
	return &ResponseBuilder[T]{
		response: Response[T]{
			Messages: []ResponseMessage{},
		},
	}
}

// NewResponseBuilder initializes ResponseBuilder with generics
func NewDefaultResponseBuilder() *ResponseBuilder[map[string]any] {
	return &ResponseBuilder[map[string]any]{
		response: Response[map[string]any]{
			Messages: []ResponseMessage{},
		},
	}
}

func (h *ResponseBuilder[T]) AddMessage(message ResponseMessage) *ResponseBuilder[T] {
	h.response.Messages = append(h.response.Messages, message)
	return h
}

func (h *ResponseBuilder[T]) AddData(data *T) *ResponseBuilder[T] {
	h.response.Data = data
	return h
}

func (h *ResponseBuilder[T]) AddSchema(schema *RootSchema) *ResponseBuilder[T] {
	h.response.Schema = schema
	return h
}

func (h *ResponseBuilder[T]) Build() *Response[T] {
	return &h.response
}

type ResponseMessage struct {
	Gravity ResponseMessageGravity `json:"gravity"`
	Value   string                 `json:"value"`
}

// Message Gravity
type ResponseMessageGravity string

const (
	ResponseMessageGravityInfo    ResponseMessageGravity = "Info"
	ResponseMessageGravityWarning ResponseMessageGravity = "Warning"
	ResponseMessageGravityError   ResponseMessageGravity = "Error"
	ResponseMessageGravityFatal   ResponseMessageGravity = "Fatal"
)

func NewMessage(gravity ResponseMessageGravity, value string) ResponseMessage {
	return ResponseMessage{
		Gravity: gravity,
		Value:   value,
	}
}
