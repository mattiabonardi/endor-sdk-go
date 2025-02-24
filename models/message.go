package models

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
