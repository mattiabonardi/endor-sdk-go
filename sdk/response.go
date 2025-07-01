package sdk

// Response with Generics
type Response[T any] struct {
	Messages []Message   `json:"messages"`
	Data     *T          `json:"data"`
	Schema   *RootSchema `json:"schema"`
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

type Meta struct {
	Default  Presentation            `json:"default"`
	Elements map[string]Presentation `json:"elements"`
}

type Presentation struct {
	Entity string `json:"entity"`
	Icon   string `json:"icon"`
}

type Message struct {
	Gravity MessageGravity `json:"gravity"`
	Value   string         `json:"value"`
}

// Message Gravity
type MessageGravity string

const (
	Info    MessageGravity = "Info"
	Warning MessageGravity = "Warning"
	Error   MessageGravity = "Error"
	Fatal   MessageGravity = "Fatal"
)

func NewMessage(gravity MessageGravity, value string) Message {
	return Message{
		Gravity: gravity,
		Value:   value,
	}
}
