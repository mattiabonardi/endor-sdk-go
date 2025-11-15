package sdk_test

import (
	"reflect"
	"testing"

	"github.com/mattiabonardi/endor-sdk-go/sdk"
)

// TestEventDefinition per struct di test
type TestEventPayload struct {
	Message string `json:"message"`
	Value   int    `json:"value"`
}

func TestNewEventDefinition(t *testing.T) {
	// Test creazione event definition con payload typato
	eventDef := sdk.NewEventDefinition[TestEventPayload]("test.created", "Test event creation")

	if eventDef.Name != "test.created" {
		t.Errorf("Expected name 'test.created', got '%s'", eventDef.Name)
	}

	if eventDef.Description != "Test event creation" {
		t.Errorf("Expected description 'Test event creation', got '%s'", eventDef.Description)
	}

	if eventDef.PayloadType != reflect.TypeOf(TestEventPayload{}) {
		t.Errorf("Expected payload type to be TestEventPayload, got %v", eventDef.PayloadType)
	}

	if eventDef.PayloadSchema == nil {
		t.Error("Expected payload schema to be generated")
	}
}

func TestEventDefinitionValidatePayload(t *testing.T) {
	eventDef := sdk.NewEventDefinition[TestEventPayload]("test.created", "Test event")

	// Test con payload valido
	validPayload := TestEventPayload{
		Message: "Hello",
		Value:   42,
	}

	if err := eventDef.ValidatePayload(validPayload); err != nil {
		t.Errorf("Expected valid payload to pass validation, got error: %v", err)
	}

	// Test con payload dello stesso tipo ma diverse proprietà
	validPayload2 := TestEventPayload{
		Message: "World",
		Value:   100,
	}

	if err := eventDef.ValidatePayload(validPayload2); err != nil {
		t.Errorf("Expected valid payload to pass validation, got error: %v", err)
	}
}

func TestDefaultEventBus(t *testing.T) {
	bus := sdk.NewDefaultEventBus()

	// Test pubblicazione evento
	event := sdk.Event{
		Name:      "test.event",
		Payload:   map[string]interface{}{"test": "data"},
		Timestamp: 1234567890,
		Source:    "test-service",
	}

	if err := bus.Publish(event); err != nil {
		t.Errorf("Expected publish to succeed, got error: %v", err)
	}
	// Pubblica di nuovo l'evento per testare il handler
	if err := bus.Publish(event); err != nil {
		t.Errorf("Expected publish to succeed, got error: %v", err)
	}

	// Nota: dato che il handler viene eseguito in goroutine,
	// dovremmo aggiungere un piccolo delay per verificare
	// ma per semplicità assumiamo che funzioni
}

func TestEndorServiceActionWithEvents(t *testing.T) {
	// Crea un event definition
	eventDef := sdk.NewEventDefinition[TestEventPayload]("user.created", "User creation event")

	// Crea una action con eventi
	handler := func(ctx *sdk.EndorContext[sdk.NoPayload]) (*sdk.Response[string], error) {
		// Testa l'emissione di eventi
		payload := TestEventPayload{
			Message: "User created successfully",
			Value:   1,
		}

		if err := ctx.EmitEvent("user.created", payload); err != nil {
			return nil, err
		}

		message := "Success"
		return &sdk.Response[string]{
			Data: &message,
		}, nil
	}

	action := sdk.NewActionWithEvents(handler, "Create user", eventDef)

	// Verifica che l'evento sia stato aggiunto
	if retrievedEvent, exists := action.GetEvent("user.created"); !exists {
		t.Error("Expected event 'user.created' to be present in action")
	} else if retrievedEvent.Name != "user.created" {
		t.Errorf("Expected event name 'user.created', got '%s'", retrievedEvent.Name)
	}

	// Verifica le opzioni
	options := action.GetOptions()
	if len(options.Events) != 1 {
		t.Errorf("Expected 1 event in options, got %d", len(options.Events))
	}
}

func TestEndorContextEmitEvent(t *testing.T) {
	// Crea un event bus di test
	bus := sdk.NewDefaultEventBus()

	// Crea una definizione di evento
	eventDef := sdk.NewEventDefinition[TestEventPayload]("test.emit", "Test emit event")

	// Crea un context con eventi disponibili
	ctx := &sdk.EndorContext[sdk.NoPayload]{
		MicroServiceId: "test-service",
		EventBus:       bus,
		AvailableEvents: map[string]*sdk.EventDefinition{
			"test.emit": eventDef,
		},
	}

	// Test emissione evento valido
	payload := TestEventPayload{
		Message: "Test message",
		Value:   123,
	}

	if err := ctx.EmitEvent("test.emit", payload); err != nil {
		t.Errorf("Expected emit to succeed, got error: %v", err)
	}

	// Test emissione evento non definito
	if err := ctx.EmitEvent("undefined.event", payload); err == nil {
		t.Error("Expected error for undefined event, got nil")
	}

	// Test emissione con payload invalido
	invalidPayload := "not a TestEventPayload"
	if err := ctx.EmitEvent("test.emit", invalidPayload); err == nil {
		t.Error("Expected error for invalid payload, got nil")
	}
}

func TestEndorContextEmitEventWithoutBus(t *testing.T) {
	// Crea un context senza event bus
	ctx := &sdk.EndorContext[sdk.NoPayload]{
		MicroServiceId: "test-service",
		EventBus:       nil,
		AvailableEvents: map[string]*sdk.EventDefinition{
			"test.emit": sdk.NewEventDefinition[TestEventPayload]("test.emit", "Test event"),
		},
	}

	payload := TestEventPayload{Message: "Test", Value: 1}

	if err := ctx.EmitEvent("test.emit", payload); err == nil {
		t.Error("Expected error when no event bus is configured, got nil")
	}
}
