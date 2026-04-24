package sdk

import (
	"regexp"

	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk_i18n"
)

var i18nTokenRegexp = regexp.MustCompile(`t\(([^)]+)\)`)

// resolveI18nValue replaces every t(key) token inside value with its translation for the given locale.
func resolveI18nValue(locale, value string) string {
	return i18nTokenRegexp.ReplaceAllStringFunc(value, func(match string) string {
		key := match[2 : len(match)-1] // strip leading "t(" and trailing ")"
		return sdk_i18n.T(locale, key, nil)
	})
}

// Response with Generics
type Response[T any] struct {
	Messages   []ResponseMessage       `json:"messages"`
	Data       *T                      `json:"data"`
	Schema     *RootSchema             `json:"schema"`
	References *EntityRefererenceGroup `json:"references"`
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

func (h *ResponseBuilder[T]) AddReferences(references EntityRefererenceGroup) *ResponseBuilder[T] {
	h.response.References = &references
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

// ResolveTranslations resolves t(key) tokens in the schema (if present).
func (r *Response[T]) ResolveTranslations(locale string) {
	if r.Schema != nil {
		r.Schema.ResolveTranslations(locale)
	}
}
