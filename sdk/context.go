package sdk

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

type Session struct {
	Id          string `json:"id"`
	Username    string `json:"username"`
	Development bool   `json:"development"`
}

type EndorContext[T any] struct {
	MicroServiceId         string
	Session                Session
	Payload                T
	ResourceMetadataSchema RootSchema
	EventBus               EventBus
	AvailableEvents        map[string]*EventDefinition

	GinContext *gin.Context
}

// EmitEvent pubblica un evento validando il payload rispetto alla definizione
func (ec *EndorContext[T]) EmitEvent(eventName string, payload interface{}) error {
	// Verifica che l'evento sia definito per questa action
	eventDef, exists := ec.AvailableEvents[eventName]
	if !exists {
		return fmt.Errorf("event '%s' is not defined for this action", eventName)
	}

	// Valida il payload
	if err := eventDef.ValidatePayload(payload); err != nil {
		return fmt.Errorf("event payload validation failed: %w", err)
	}

	// Se non c'è un event bus configurato, restituisci errore
	if ec.EventBus == nil {
		return fmt.Errorf("no event bus configured")
	}

	// Crea l'evento
	event := Event{
		Name:      eventName,
		Payload:   payload,
		Timestamp: time.Now().Unix(),
		Source:    ec.MicroServiceId,
	}

	// Pubblica l'evento
	return ec.EventBus.Publish(event)
}

type NoPayload struct{}
