package sdk

import (
	"context"
	"fmt"
)

type EndorHybridService struct {
	Resource       string
	Description    string
	Priority       *int
	metadataSchema Schema
	methodsFn      func(getSchema func() Schema, getCategorySchema func(categoryID string) Schema) map[string]EndorServiceAction
	baseModel      any                 // Modello base per ResourceInstanceRepository
	categories     map[string]Category // Cache delle categorie disponibili
}

func NewHybridService(resource, description string) EndorHybridService {
	return EndorHybridService{
		Resource:    resource,
		Description: description,
		baseModel:   &DynamicResource{}, // Default model
	}
}

// WithBaseModel permette di specificare un modello base personalizzato
func (h EndorHybridService) WithBaseModel(model any) EndorHybridService {
	h.baseModel = model
	return h
}

// Definizione dei metodi tramite funzione che riceve getSchema() e getCategorySchema()
func (h EndorHybridService) WithActions(
	fn func(getSchema func() Schema, getCategorySchema func(categoryID string) Schema) map[string]EndorServiceAction,
) EndorHybridService {
	h.methodsFn = fn
	return h
}

// WithCategories permette di configurare le categorie disponibili
func (h EndorHybridService) WithCategories(categories []Category) EndorHybridService {
	if h.categories == nil {
		h.categories = make(map[string]Category)
	}
	for _, category := range categories {
		h.categories[category.ID] = category
	}
	return h
}

// getDefaultActions fornisce le 6 azioni CRUD di default per hybrid services
func (h EndorHybridService) getDefaultActions(getSchema func() Schema) map[string]EndorServiceAction {
	// Crea repository usando DynamicResource come default (per ora)
	autogenerateID := true
	repository := NewResourceInstanceRepository[*DynamicResource](h.Resource, ResourceInstanceRepositoryOptions{
		AutoGenerateID: &autogenerateID,
	})

	return map[string]EndorServiceAction{
		"schema": NewAction(
			func(c *EndorContext[NoPayload]) (*Response[any], error) {
				return h.defaultSchema(c, getSchema)
			},
			fmt.Sprintf("Get the schema of the %s (%s)", h.Resource, h.Description),
		),
		"list": NewAction(
			func(c *EndorContext[ReadDTO]) (*Response[[]ResourceInstance[*DynamicResource]], error) {
				return h.defaultList(c, getSchema, repository)
			},
			fmt.Sprintf("Search for available list of %s (%s)", h.Resource, h.Description),
		),
		"create": NewConfigurableAction(
			EndorServiceActionOptions{
				Description:     fmt.Sprintf("Create the instance of %s (%s)", h.Resource, h.Description),
				Public:          false,
				ValidatePayload: true,
				InputSchema: &RootSchema{
					Schema: Schema{
						Type: ObjectType,
						Properties: &map[string]Schema{
							"data": h.getRootSchema(getSchema).Schema,
						},
					},
				},
			},
			func(c *EndorContext[CreateDTO[ResourceInstance[*DynamicResource]]]) (*Response[ResourceInstance[*DynamicResource]], error) {
				return h.defaultCreate(c, getSchema, repository)
			},
		),
		"instance": NewAction(
			func(c *EndorContext[ReadInstanceDTO]) (*Response[*ResourceInstance[*DynamicResource]], error) {
				return h.defaultInstance(c, getSchema, repository)
			},
			fmt.Sprintf("Get the instance of %s (%s)", h.Resource, h.Description),
		),
		"update": NewConfigurableAction(
			EndorServiceActionOptions{
				Description:     fmt.Sprintf("Update the existing instance of %s (%s)", h.Resource, h.Description),
				Public:          false,
				ValidatePayload: true,
				InputSchema: &RootSchema{
					Schema: Schema{
						Type: ObjectType,
						Properties: &map[string]Schema{
							"id": {
								Type: StringType,
							},
							"data": h.getRootSchema(getSchema).Schema,
						},
					},
				},
			},
			func(c *EndorContext[UpdateByIdDTO[ResourceInstance[*DynamicResource]]]) (*Response[ResourceInstance[*DynamicResource]], error) {
				return h.defaultUpdate(c, getSchema, repository)
			},
		),
		"delete": NewAction(
			func(c *EndorContext[ReadInstanceDTO]) (*Response[any], error) {
				return h.defaultDelete(c, repository)
			},
			fmt.Sprintf("Delete the existing instance of %s (%s)", h.Resource, h.Description),
		),
	}
}

// Conversione in EndorService, schema iniettato dal framework
func (h EndorHybridService) ToEndorService(attrs Schema) EndorService {
	h.metadataSchema = attrs
	getSchema := func() Schema { return h.metadataSchema }
	getCategorySchema := func(categoryID string) Schema {
		if category, exists := h.categories[categoryID]; exists {
			if categorySchema, err := category.UnmarshalAdditionalAttributes(); err == nil {
				return categorySchema.Schema
			}
		}
		return Schema{} // Schema vuoto se categoria non trovata o errore
	}

	var methods map[string]EndorServiceAction

	// Se non ci sono metodi personalizzati, usa quelli di default
	if h.methodsFn == nil {
		methods = h.getDefaultActions(getSchema)
	} else {
		// Inizia con i metodi di default
		methods = h.getDefaultActions(getSchema)
		// Sovrascrivi con i metodi personalizzati
		customMethods := h.methodsFn(getSchema, getCategorySchema)
		for name, action := range customMethods {
			methods[name] = action
		}
	}

	return EndorService{
		Resource:    h.Resource,
		Description: h.Description,
		Priority:    h.Priority,
		Methods:     methods,
	}
}

// Helper method per creare rootSchema
func (h EndorHybridService) getRootSchema(getSchema func() Schema) *RootSchema {
	schema := getSchema()
	rootSchema := NewSchema(h.baseModel)
	if schema.Properties != nil {
		for k, v := range *schema.Properties {
			(*rootSchema.Properties)[k] = v
		}
	}
	return rootSchema
}

// Implementazioni dei metodi di default
func (h EndorHybridService) defaultSchema(_ *EndorContext[NoPayload], getSchema func() Schema) (*Response[any], error) {
	rootSchema := h.getRootSchema(getSchema)
	return NewResponseBuilder[any]().AddSchema(rootSchema).Build(), nil
}

func (h EndorHybridService) defaultInstance(c *EndorContext[ReadInstanceDTO], getSchema func() Schema, repository *ResourceInstanceRepository[*DynamicResource]) (*Response[*ResourceInstance[*DynamicResource]], error) {
	instance, err := repository.Instance(context.TODO(), c.Payload)
	if err != nil {
		return nil, err
	}
	rootSchema := h.getRootSchema(getSchema)
	return NewResponseBuilder[*ResourceInstance[*DynamicResource]]().AddData(&instance).AddSchema(rootSchema).Build(), nil
}

func (h EndorHybridService) defaultList(c *EndorContext[ReadDTO], getSchema func() Schema, repository *ResourceInstanceRepository[*DynamicResource]) (*Response[[]ResourceInstance[*DynamicResource]], error) {
	list, err := repository.List(context.TODO(), c.Payload)
	if err != nil {
		return nil, err
	}
	rootSchema := h.getRootSchema(getSchema)
	return NewResponseBuilder[[]ResourceInstance[*DynamicResource]]().AddData(&list).AddSchema(rootSchema).Build(), nil
}

func (h EndorHybridService) defaultCreate(c *EndorContext[CreateDTO[ResourceInstance[*DynamicResource]]], getSchema func() Schema, repository *ResourceInstanceRepository[*DynamicResource]) (*Response[ResourceInstance[*DynamicResource]], error) {
	created, err := repository.Create(context.TODO(), c.Payload)
	if err != nil {
		return nil, err
	}
	rootSchema := h.getRootSchema(getSchema)
	return NewResponseBuilder[ResourceInstance[*DynamicResource]]().AddData(created).AddSchema(rootSchema).AddMessage(NewMessage(Info, fmt.Sprintf("%s created", h.Resource))).Build(), nil
}

func (h EndorHybridService) defaultUpdate(c *EndorContext[UpdateByIdDTO[ResourceInstance[*DynamicResource]]], getSchema func() Schema, repository *ResourceInstanceRepository[*DynamicResource]) (*Response[ResourceInstance[*DynamicResource]], error) {
	updated, err := repository.Update(context.TODO(), c.Payload)
	if err != nil {
		return nil, err
	}
	rootSchema := h.getRootSchema(getSchema)
	return NewResponseBuilder[ResourceInstance[*DynamicResource]]().AddData(updated).AddSchema(rootSchema).AddMessage(NewMessage(Info, fmt.Sprintf("%s updated", h.Resource))).Build(), nil
}

func (h EndorHybridService) defaultDelete(c *EndorContext[ReadInstanceDTO], repository *ResourceInstanceRepository[*DynamicResource]) (*Response[any], error) {
	err := repository.Delete(context.TODO(), c.Payload)
	if err != nil {
		return nil, err
	}
	return NewResponseBuilder[any]().AddMessage(NewMessage(Info, fmt.Sprintf("%s deleted", h.Resource))).Build(), nil
}
