package sdk

type Session struct {
	Id       string `json:"id"`
	Username string `json:"username"`
}

type EndorContext[T any] struct {
	MicroServiceId string
	Session        Session
	Payload        T
}

type NoPayload struct{}
