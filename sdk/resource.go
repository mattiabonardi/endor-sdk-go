package sdk

import (
	"encoding/json"
	"fmt"
	"reflect"

	"gopkg.in/yaml.v3"
)

// Category rappresenta una categoria con attributi dinamici specifici
type Category struct {
	ID                   string `json:"id" bson:"id" schema:"title=Category ID"`
	Description          string `json:"description" bson:"description" schema:"title=Category Description"`
	AdditionalAttributes string `json:"additionalAttributes" bson:"additionalAttributes" schema:"title=Additional category attributes schema,format=yaml"`
}

// CategorySpecialized rappresenta una categoria con supporto per modello statico
type CategorySpecialized[C any] struct {
	ID                   string `json:"id" bson:"id" schema:"title=Category ID"`
	Description          string `json:"description" bson:"description" schema:"title=Category Description"`
	StaticModel          *C     `json:"-" bson:"-"` // Modello statico (solo per typing, non serializzato)
	AdditionalAttributes string `json:"additionalAttributes" bson:"additionalAttributes" schema:"title=Additional category attributes schema,format=yaml"`
}

// ToCategory converte una CategorySpecialized in una Category per backward compatibility
func (c *CategorySpecialized[C]) ToCategory() Category {
	return Category{
		ID:                   c.ID,
		Description:          c.Description,
		AdditionalAttributes: c.AdditionalAttributes,
	}
}

// NewCategorySpecialized crea una nuova categoria specializzata con modello statico
func NewCategorySpecialized[C any](id, description string, staticModel *C, additionalAttributes string) CategorySpecialized[C] {
	return CategorySpecialized[C]{
		ID:                   id,
		Description:          description,
		StaticModel:          staticModel,
		AdditionalAttributes: additionalAttributes,
	}
}

// GetStaticModelSchema restituisce lo schema del modello statico
func (c *CategorySpecialized[C]) GetStaticModelSchema() *RootSchema {
	if c.StaticModel == nil {
		return nil
	}
	return NewSchema(*c.StaticModel)
}

// GetCombinedSchema combina lo schema del modello statico con gli attributi aggiuntivi
func (c *CategorySpecialized[C]) GetCombinedSchema() (*RootSchema, error) {
	var combinedSchema *RootSchema

	// Inizia con lo schema del modello statico se presente
	if c.StaticModel != nil {
		combinedSchema = NewSchema(*c.StaticModel)
	} else {
		combinedSchema = &RootSchema{
			Schema: Schema{
				Type:       ObjectType,
				Properties: &map[string]Schema{},
			},
		}
	}

	// Aggiungi gli attributi aggiuntivi se presenti
	if c.AdditionalAttributes != "" {
		var additionalSchema RootSchema
		err := yaml.Unmarshal([]byte(c.AdditionalAttributes), &additionalSchema)
		if err != nil {
			return nil, fmt.Errorf("failed to parse Category AdditionalAttributes YAML: %w", err)
		}

		// Combina i due schemi
		if additionalSchema.Schema.Properties != nil {
			for k, v := range *additionalSchema.Schema.Properties {
				(*combinedSchema.Schema.Properties)[k] = v
			}
		}
	}

	return combinedSchema, nil
}

// UnmarshalAdditionalAttributes deserializza gli attributi aggiuntivi della categoria
func (c *Category) UnmarshalAdditionalAttributes() (*RootSchema, error) {
	var schema RootSchema
	err := yaml.Unmarshal([]byte(c.AdditionalAttributes), &schema)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Category AdditionalAttributes YAML: %w", err)
	}
	return &schema, nil
}

type Resource struct {
	ID                   string     `json:"id" bson:"_id" schema:"title=Id"`
	Description          string     `json:"description" schema:"title=Description"`
	Service              string     `json:"service" schema:"title=Service" ui-schema:"resource=microservice"`
	AdditionalAttributes string     `json:"additionalAttributes" schema:"title=Additional attributes schema,format=yaml"` // YAML string, raw
	Categories           []Category `json:"categories,omitempty" bson:"categories,omitempty" schema:"title=Categories"`
}

func (h *Resource) UnmarshalAdditionalAttributes() (*RootSchema, error) {
	var schema RootSchema
	err := yaml.Unmarshal([]byte(h.AdditionalAttributes), &schema)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ResourceDefinition YAML: %w", err)
	}
	return &schema, nil
}

// GetCategoryByID trova una categoria per ID
func (h *Resource) GetCategoryByID(categoryID string) (*Category, bool) {
	for i, category := range h.Categories {
		if category.ID == categoryID {
			return &h.Categories[i], true
		}
	}
	return nil, false
}

// GetCategorySchema restituisce lo schema combinato di base e categoria specifica
func (h *Resource) GetCategorySchema(categoryID string) (*RootSchema, error) {
	// Schema di base della risorsa
	baseSchema, err := h.UnmarshalAdditionalAttributes()
	if err != nil {
		return nil, err
	}

	// Se non c'è categoria specificata, restituisci solo lo schema di base
	if categoryID == "" {
		return baseSchema, nil
	}

	// Trova la categoria
	category, found := h.GetCategoryByID(categoryID)
	if !found {
		return nil, fmt.Errorf("category '%s' not found", categoryID)
	}

	// Schema della categoria
	categorySchema, err := category.UnmarshalAdditionalAttributes()
	if err != nil {
		return nil, err
	}

	// Combina gli schemi (categoria estende il base)
	combined := &RootSchema{
		Schema: Schema{
			Type:       ObjectType,
			Properties: &map[string]Schema{},
		},
	}

	// Aggiungi proprietà di base
	if baseSchema.Properties != nil {
		for k, v := range *baseSchema.Properties {
			(*combined.Properties)[k] = v
		}
	}

	// Aggiungi proprietà della categoria (sovrascrivono se esistenti)
	if categorySchema.Properties != nil {
		for k, v := range *categorySchema.Properties {
			(*combined.Properties)[k] = v
		}
	}

	return combined, nil
}

type ResourceAction struct {
	// version/resource/action
	ID          string `json:"id" schema:"title=Id"`
	Resource    string `json:"resource" schema:"title=Resource" ui-schema:"resource=resource"`
	Description string `json:"description" schema:"title=Description"`
	InputSchema string `json:"inputSchema" schema:"title=Input schema,format=yaml"`
}

type ResourceInstanceInterface interface {
	GetID() *string
	SetID(id string)
}

type ResourceInstance[T ResourceInstanceInterface] struct {
	This     T              `bson:",inline"`
	Metadata map[string]any `bson:"metadata,omitempty"`
}

// ToSchema genera lo schema della risorsa, includendo lo schema dei metadata
func (d *ResourceInstance[T]) ToSchema(metadataSchema *Schema) *RootSchema {
	t := reflect.TypeOf(d.This)
	// Se T è un puntatore, ottieni il tipo sottostante
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	rootSchema := NewSchema(reflect.New(t).Interface())
	if metadataSchema != nil && metadataSchema.Properties != nil {
		if rootSchema.Schema.Properties == nil {
			props := make(map[string]Schema)
			rootSchema.Schema.Properties = &props
		}
		for k, v := range *metadataSchema.Properties {
			(*rootSchema.Schema.Properties)[k] = v
		}
	}
	return rootSchema
}

func (d ResourceInstance[T]) MarshalJSON() ([]byte, error) {
	base, err := json.Marshal(d.This)
	if err != nil {
		return nil, err
	}

	var baseMap map[string]any
	if err := json.Unmarshal(base, &baseMap); err != nil {
		return nil, err
	}

	for k, v := range d.Metadata {
		if _, exists := baseMap[k]; !exists {
			baseMap[k] = v
		}
	}

	return json.Marshal(baseMap)
}

func (d *ResourceInstance[T]) UnmarshalJSON(data []byte) error {
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	var temp T
	t := reflect.TypeOf(temp)
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	known := map[string]struct{}{}
	for i := 0; i < t.NumField(); i++ {
		tag := t.Field(i).Tag.Get("json")
		if tag == "" {
			tag = t.Field(i).Name
		}
		known[tag] = struct{}{}
	}

	baseMap := make(map[string]any)
	metaMap := make(map[string]any)

	for k, v := range raw {
		if _, isKnown := known[k]; isKnown {
			baseMap[k] = v
		} else {
			metaMap[k] = v
		}
	}

	baseBytes, err := json.Marshal(baseMap)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(baseBytes, &d.This); err != nil {
		return err
	}

	d.Metadata = metaMap
	return nil
}

func (d *ResourceInstance[T]) GetID() *string {
	return d.This.GetID()
}

type DynamicResource struct {
	Id          string `json:"id" bson:"_id" schema:"title=Id,readOnly=true" ui-schema:"hidden=true"`
	Description string `json:"description" bson:"description" schema:"title=Description"`
}

func (h *DynamicResource) GetID() *string {
	return &h.Id
}

func (h *DynamicResource) SetID(id string) {
	h.Id = id
}

// ResourceInstanceSpecialization rappresenta la specializzazione di una categoria
type ResourceInstanceSpecialization[C any] struct {
	This     C                      `json:",inline" bson:"this"`
	Metadata map[string]interface{} `json:",inline" bson:"metadata,omitempty"`
}

// ResourceInstanceSpecialized rappresenta un'istanza di risorsa specializzata con categoria
type ResourceInstanceSpecialized[T ResourceInstanceInterface, C any] struct {
	This           T                                  `json:",inline" bson:"this"`
	Metadata       map[string]interface{}             `json:",inline" bson:"metadata,omitempty"`
	Specialization *ResourceInstanceSpecialization[C] `json:",inline" bson:"specialization,omitempty"`
}

// GetID implementa ResourceInstanceInterface
func (r *ResourceInstanceSpecialized[T, C]) GetID() *string {
	return r.This.GetID()
}

// SetID implementa ResourceInstanceInterface
func (r *ResourceInstanceSpecialized[T, C]) SetID(id string) {
	r.This.SetID(id)
}

// GetCategoryID restituisce l'ID della categoria se presente nella specializzazione
func (r *ResourceInstanceSpecialized[T, C]) GetCategoryID() string {
	if r.Specialization == nil {
		return ""
	}
	// Implementazione specifica dipenderà da come gestiamo l'ID categoria nella specializzazione
	// Per ora restituiamo una stringa vuota, sarà definita meglio quando estenderemo Category
	return ""
}
