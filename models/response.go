package models

type Response struct {
	Messages []Message `json:"messages"`
	Data     any       `json:"data"`
}

type ResponseBuilder struct {
	response Response
}

func NewResponseBuilder() *ResponseBuilder {
	return &ResponseBuilder{
		response: Response{
			Messages: []Message{},
			Data:     nil,
		},
	}
}

func (h *ResponseBuilder) AddMessage(message Message) *ResponseBuilder {
	h.response.Messages = append(h.response.Messages, message)
	return h
}

func (h *ResponseBuilder) AddData(data interface{}) *ResponseBuilder {
	h.response.Data = data
	return h
}

func (h *ResponseBuilder) Build() *Response {
	return &h.response
}
