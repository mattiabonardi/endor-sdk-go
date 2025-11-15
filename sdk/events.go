package sdk

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// Event rappresenta un evento che può essere emesso
type Event struct {
	Name      string      `json:"name"`
	Payload   interface{} `json:"payload"`
	Timestamp int64       `json:"timestamp"`
	Source    string      `json:"source,omitempty"`
}

// EventDefinition definisce la struttura di un evento
type EventDefinition struct {
	Name          string       `json:"name"`
	Description   string       `json:"description"`
	PayloadType   reflect.Type `json:"-"`
	PayloadSchema *RootSchema  `json:"payloadSchema,omitempty"`
}

// NewEventDefinition crea una nuova definizione di evento con tipo generico
func NewEventDefinition[T any](name, description string) *EventDefinition {
	var zeroT T
	payloadType := reflect.TypeOf(zeroT)

	// Se è un puntatore, prendi il tipo sottostante
	if payloadType != nil && payloadType.Kind() == reflect.Ptr {
		payloadType = payloadType.Elem()
	}

	var schema *RootSchema
	if payloadType != nil && payloadType != reflect.TypeOf(struct{}{}) {
		schema = NewSchemaByType(payloadType)
	}

	return &EventDefinition{
		Name:          name,
		Description:   description,
		PayloadType:   payloadType,
		PayloadSchema: schema,
	}
}

// ValidatePayload valida che il payload rispetti il tipo definito
func (ed *EventDefinition) ValidatePayload(payload interface{}) error {
	if ed.PayloadType == nil {
		return nil
	}

	payloadType := reflect.TypeOf(payload)
	if payloadType != ed.PayloadType {
		// Prova a validare tramite JSON marshaling/unmarshaling
		if err := ed.validateThroughJSON(payload); err != nil {
			return fmt.Errorf("payload type mismatch for event '%s': expected %v, got %v",
				ed.Name, ed.PayloadType, payloadType)
		}
	}

	return nil
}

// validateThroughJSON valida il payload convertendolo a JSON e back
func (ed *EventDefinition) validateThroughJSON(payload interface{}) error {
	// Serializza il payload in JSON
	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload to JSON: %w", err)
	}

	// Crea un'istanza del tipo atteso
	expectedInstance := reflect.New(ed.PayloadType).Interface()

	// Deserializza il JSON nell'istanza del tipo atteso
	if err := json.Unmarshal(jsonBytes, expectedInstance); err != nil {
		return fmt.Errorf("payload cannot be converted to expected type: %w", err)
	}

	return nil
}

// EventBus interfaccia per la pubblicazione di eventi
type EventBus interface {
	// Publish pubblica un evento
	Publish(event Event) error
}

// DefaultEventBus implementazione di base dell'EventBus (no-op per ora)
type DefaultEventBus struct {
	handlers map[string][]func(Event)
}

// NewDefaultEventBus crea una nuova istanza del bus eventi di default
func NewDefaultEventBus() *DefaultEventBus {
	return &DefaultEventBus{
		handlers: make(map[string][]func(Event)),
	}
}

// Publish implementa EventBus.Publish
func (eb *DefaultEventBus) Publish(event Event) error {
	// Implementazione di base - per ora logga solamente
	// In una implementazione reale, qui ci sarebbe la logica di pubblicazione
	// (es. Redis, Kafka, RabbitMQ, etc.)

	// Chiama gli handlers registrati se ci sono
	if handlers, exists := eb.handlers[event.Name]; exists {
		for _, handler := range handlers {
			go handler(event)
		}
	}

	return nil
}
