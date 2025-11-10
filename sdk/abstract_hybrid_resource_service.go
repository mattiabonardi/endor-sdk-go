package sdk

import (
	"context"
	"fmt"
)

type AbstractHybridResourceService struct {
	resource   string
	repository *ResourceInstanceRepository[*DynamicResource]
}

func NewAbstractHybridResourceService(resource string, description string) EndorHybridService {
	autogenerateID := true
	service := AbstractHybridResourceService{
		resource: resource,
		repository: NewResourceInstanceRepository[*DynamicResource](resource, ResourceInstanceRepositoryOptions{
			AutoGenerateID: &autogenerateID,
		}),
	}

	return NewHybridService(resource, description).WithActions(
		func(getSchema func() Schema) map[string]EndorServiceAction {
			// Otteniamo lo schema dinamico al momento della chiamata
			schema := getSchema()

			// Creiamo la rootSchema che combina DynamicResource con gli attributi aggiuntivi
			rootSchema := NewSchema(DynamicResource{})
			if schema.Properties != nil {
				for k, v := range *schema.Properties {
					(*rootSchema.Properties)[k] = v
				}
			}

			return map[string]EndorServiceAction{
				"schema": NewAction(
					func(c *EndorContext[NoPayload]) (*Response[any], error) {
						return service.schema(c, rootSchema)
					},
					fmt.Sprintf("Get the schema of the %s (%s)", resource, description),
				),
				"list": NewAction(
					func(c *EndorContext[ReadDTO]) (*Response[[]ResourceInstance[*DynamicResource]], error) {
						return service.list(c, rootSchema)
					},
					fmt.Sprintf("Search for available list of %s (%s)", resource, description),
				),
				"create": NewConfigurableAction(
					EndorServiceActionOptions{
						Description:     fmt.Sprintf("Create the instance of %s (%s)", resource, description),
						Public:          false,
						ValidatePayload: true,
						InputSchema: &RootSchema{
							Schema: Schema{
								Type: ObjectType,
								Properties: &map[string]Schema{
									"data": rootSchema.Schema,
								},
							},
						},
					},
					func(c *EndorContext[CreateDTO[ResourceInstance[*DynamicResource]]]) (*Response[ResourceInstance[*DynamicResource]], error) {
						return service.create(c, rootSchema)
					},
				),
				"instance": NewAction(
					func(c *EndorContext[ReadInstanceDTO]) (*Response[*ResourceInstance[*DynamicResource]], error) {
						return service.instance(c, rootSchema)
					},
					fmt.Sprintf("Get the instance of %s (%s)", resource, description),
				),
				"update": NewConfigurableAction(
					EndorServiceActionOptions{
						Description:     fmt.Sprintf("Update the existing instance of %s (%s)", resource, description),
						Public:          false,
						ValidatePayload: true,
						InputSchema: &RootSchema{
							Schema: Schema{
								Type: ObjectType,
								Properties: &map[string]Schema{
									"id": {
										Type: StringType,
									},
									"data": rootSchema.Schema,
								},
							},
						},
					},
					func(c *EndorContext[UpdateByIdDTO[ResourceInstance[*DynamicResource]]]) (*Response[ResourceInstance[*DynamicResource]], error) {
						return service.update(c, rootSchema)
					},
				),
				"delete": NewAction(
					func(c *EndorContext[ReadInstanceDTO]) (*Response[any], error) {
						return service.delete(c)
					},
					fmt.Sprintf("Delete the existing instance of %s (%s)", resource, description),
				),
			}
		},
	)
}

func (h *AbstractHybridResourceService) schema(_ *EndorContext[NoPayload], rootSchema *RootSchema) (*Response[any], error) {
	return NewResponseBuilder[any]().AddSchema(rootSchema).Build(), nil
}

func (h *AbstractHybridResourceService) instance(c *EndorContext[ReadInstanceDTO], rootSchema *RootSchema) (*Response[*ResourceInstance[*DynamicResource]], error) {
	instance, err := h.repository.Instance(context.TODO(), c.Payload)
	if err != nil {
		return nil, err
	}
	return NewResponseBuilder[*ResourceInstance[*DynamicResource]]().AddData(&instance).AddSchema(rootSchema).Build(), nil
}

func (h *AbstractHybridResourceService) list(c *EndorContext[ReadDTO], rootSchema *RootSchema) (*Response[[]ResourceInstance[*DynamicResource]], error) {
	list, err := h.repository.List(context.TODO(), c.Payload)
	if err != nil {
		return nil, err
	}
	return NewResponseBuilder[[]ResourceInstance[*DynamicResource]]().AddData(&list).AddSchema(rootSchema).Build(), nil
}

func (h *AbstractHybridResourceService) create(c *EndorContext[CreateDTO[ResourceInstance[*DynamicResource]]], rootSchema *RootSchema) (*Response[ResourceInstance[*DynamicResource]], error) {
	created, err := h.repository.Create(context.TODO(), c.Payload)
	if err != nil {
		return nil, err
	}
	return NewResponseBuilder[ResourceInstance[*DynamicResource]]().AddData(created).AddSchema(rootSchema).AddMessage(NewMessage(Info, fmt.Sprintf("%s created", h.resource))).Build(), nil
}

func (h *AbstractHybridResourceService) update(c *EndorContext[UpdateByIdDTO[ResourceInstance[*DynamicResource]]], rootSchema *RootSchema) (*Response[ResourceInstance[*DynamicResource]], error) {
	updated, err := h.repository.Update(context.TODO(), c.Payload)
	if err != nil {
		return nil, err
	}
	return NewResponseBuilder[ResourceInstance[*DynamicResource]]().AddData(updated).AddSchema(rootSchema).AddMessage(NewMessage(Info, fmt.Sprintf("%s updated", h.resource))).Build(), nil
}

func (h *AbstractHybridResourceService) delete(c *EndorContext[ReadInstanceDTO]) (*Response[any], error) {
	err := h.repository.Delete(context.TODO(), c.Payload)
	if err != nil {
		return nil, err
	}
	return NewResponseBuilder[any]().AddMessage(NewMessage(Info, fmt.Sprintf("%s deleted", h.resource))).Build(), nil
}
