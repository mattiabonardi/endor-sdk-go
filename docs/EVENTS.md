# Sistema di Gestione Eventi - EndorService

Questo documento descrive l'implementazione del sistema di gestione degli eventi in EndorService, che permette agli sviluppatori di dichiarare e emettere eventi tipizzati dalle actions.

## Panoramica

Il sistema di eventi è composto da:

1. **EventDefinition**: Definisce la struttura di un evento con nome, descrizione e tipo di payload
2. **EventBus**: Interfaccia per la pubblicazione e sottoscrizione agli eventi
3. **EndorContext.EmitEvent()**: Metodo per emettere eventi con validazione automatica del payload
4. **Validazione automatica**: Il payload viene validato rispetto al tipo definito nell'EventDefinition

## Uso Base

### 1. Definizione degli Eventi

Crea una definizione di evento utilizzando generics per il tipo di payload:

```go
package main

import "github.com/mattiabonardi/endor-sdk-go/sdk"

// Definisci la struttura del payload
type UserCreatedPayload struct {
    UserID   string `json:"userId"`
    Username string `json:"username"`
    Email    string `json:"email"`
}

// Crea la definizione dell'evento
userCreatedEvent := sdk.NewEventDefinition[UserCreatedPayload](
    "user.created",
    "Emitted when a new user is successfully created",
)
```

### 2. Creazione di Actions con Eventi

Usa i nuovi costruttori per creare actions con eventi predefiniti:

```go
// Crea un handler che può emettere eventi
createUserHandler := func(ctx *sdk.EndorContext[CreateUserRequest]) (*sdk.Response[CreateUserResponse], error) {
    // Logica di business...
    userID := "user_12345"
    
    // Emetti l'evento
    eventPayload := UserCreatedPayload{
        UserID:   userID,
        Username: ctx.Payload.Username,
        Email:    ctx.Payload.Email,
    }
    
    if err := ctx.EmitEvent("user.created", eventPayload); err != nil {
        log.Printf("Failed to emit event: %v", err)
    }
    
    return &sdk.Response[CreateUserResponse]{
        Data: &CreateUserResponse{
            UserID:  userID,
            Message: "User created successfully",
        },
    }, nil
}

// Crea l'action con gli eventi
action := sdk.NewActionWithEvents(
    createUserHandler,
    "Creates a new user",
    userCreatedEvent,
    // puoi aggiungere altri eventi...
)
```

### 3. Configurazione dell'EventBus

Configura un EventBus personalizzato o usa quello di default:

```go
// EventBus di default
eventBus := sdk.NewDefaultEventBus()

// Registra handlers per ascoltare gli eventi
eventBus.Subscribe("user.created", func(event sdk.Event) {
    log.Printf("User created: %+v", event.Payload)
    // Invia email di benvenuto, aggiorna cache, etc.
})

// Configura il server con l'EventBus
endor := sdk.NewEndorInitializer().
    WithEndorServices(&services).
    WithEventBus(eventBus).
    Build()
```

## API Dettagliata

### EventDefinition

```go
type EventDefinition struct {
    Name          string
    Description   string
    PayloadType   reflect.Type
    PayloadSchema *RootSchema
}

// NewEventDefinition crea una nuova definizione di evento
func NewEventDefinition[T any](name, description string) *EventDefinition

// ValidatePayload valida che il payload rispetti il tipo definito
func (ed *EventDefinition) ValidatePayload(payload interface{}) error
```

### EventBus

```go
type EventBus interface {
    // Publish pubblica un evento
    Publish(event Event) error
    
    // Subscribe registra un handler per un tipo di evento
    Subscribe(eventName string, handler func(Event)) error
    
    // Unsubscribe rimuove un handler
    Unsubscribe(eventName string) error
}

// DefaultEventBus implementazione di base
type DefaultEventBus struct { /* ... */ }

func NewDefaultEventBus() *DefaultEventBus
```

### EndorContext

```go
// EmitEvent pubblica un evento validando il payload
func (ec *EndorContext[T]) EmitEvent(eventName string, payload interface{}) error
```

### EndorServiceAction

Nuovi metodi aggiunti all'interfaccia:

```go
type EndorServiceAction interface {
    // ... metodi esistenti ...
    
    // AddEvent aggiunge un evento all'action
    AddEvent(eventDef *EventDefinition) EndorServiceAction
    
    // GetEvent recupera una definizione di evento
    GetEvent(name string) (*EventDefinition, bool)
}
```

## Nuovi Costruttori

### NewActionWithEvents

```go
func NewActionWithEvents[T any, R any](
    handler EndorHandlerFunc[T, R], 
    description string, 
    events ...*EventDefinition,
) EndorServiceAction
```

### NewConfigurableActionWithEvents

```go
func NewConfigurableActionWithEvents[T any, R any](
    options EndorServiceActionOptions, 
    handler EndorHandlerFunc[T, R], 
    events ...*EventDefinition,
) EndorServiceAction
```

## Validazione del Payload

Il sistema valida automaticamente i payload degli eventi:

1. **Validazione di tipo**: Verifica che il payload corrisponda al tipo definito nell'EventDefinition
2. **Validazione JSON**: Se i tipi non corrispondono esattamente, prova a convertire via JSON
3. **Schema validation**: Usa il RootSchema generato per validazioni più dettagliate

### Esempio di Validazione

```go
// Definisci evento
eventDef := sdk.NewEventDefinition[UserCreatedPayload]("user.created", "User created")

// Payload valido
validPayload := UserCreatedPayload{UserID: "123", Username: "john", Email: "john@example.com"}
err := ctx.EmitEvent("user.created", validPayload) // ✅ OK

// Payload invalido - tipo sbagliato
invalidPayload := "not a UserCreatedPayload"
err := ctx.EmitEvent("user.created", invalidPayload) // ❌ Errore
```

## Gestione degli Errori

L'emissione di eventi può fallire per:

- **Evento non definito**: L'evento non è stato dichiarato per l'action
- **Payload invalido**: Il payload non rispetta il tipo definito
- **EventBus non configurato**: Nessun EventBus è stato configurato
- **Errori di pubblicazione**: Errori nella pubblicazione dell'evento

```go
if err := ctx.EmitEvent("user.created", payload); err != nil {
    // Gestisci l'errore - potresti loggarlo, ritentare, o interrompere l'operazione
    log.Printf("Failed to emit event: %v", err)
}
```

## Best Practices

1. **Nomenclatura**: Usa nomi di eventi descrittivi e consistenti (es. `resource.action`)
2. **Payload strutturati**: Usa struct Go tipizzate per i payload degli eventi
3. **Gestione errori**: Non ignorare gli errori di emissione eventi
4. **Documentazione**: Documenta chiaramente quando e perché vengono emessi gli eventi
5. **Idempotenza**: Gli eventi dovrebbero contenere tutte le informazioni necessarie
6. **Backwards compatibility**: Considera la compatibilità quando modifichi i payload degli eventi

## Esempio Completo

Vedi `examples/events_example.go` per un esempio completo di implementazione.

## Testing

Il sistema include test completi in `sdk/events_test.go` che coprono:

- Creazione di EventDefinition
- Validazione del payload
- Funzionalità dell'EventBus
- Integrazione con EndorServiceAction
- Emissione eventi tramite EndorContext