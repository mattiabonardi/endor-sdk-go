package sdk

// Response with Generics
type Response[T any] struct {
	Messages []Message `json:"messages"`
	Data     T         `json:"data"`
	Schema   Schema    `json:"schema"`
	Meta     Meta      `json:"meta"`
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

func (h *ResponseBuilder[T]) AddSchema(schema Schema) *ResponseBuilder[T] {
	h.response.Schema = schema
	return h
}

func (h *ResponseBuilder[T]) AddMeta(meta Meta) *ResponseBuilder[T] {
	h.response.Meta = meta
	return h
}

func (h *ResponseBuilder[T]) Build() *Response[T] {
	return &h.response
}

type SchemaTypeName string

const (
	StringType  SchemaTypeName = "string"
	NumberType  SchemaTypeName = "number"
	BooleanType SchemaTypeName = "boolean"
	ObjectType  SchemaTypeName = "object"
	ArrayType   SchemaTypeName = "array"
)

type Schema struct {
	Type       SchemaTypeName    `json:"type"`
	Properties map[string]Schema `json:"properties,omitempty"`
	Items      *Schema           `json:"items,omitempty"`
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
	Info  MessageGravity = "Info"
	Debug MessageGravity = "Debug"
	Error MessageGravity = "Error"
	Fatal MessageGravity = "Fatal"
)

func NewMessage(gravity MessageGravity, value string) Message {
	return Message{
		Gravity: gravity,
		Value:   value,
	}
}
