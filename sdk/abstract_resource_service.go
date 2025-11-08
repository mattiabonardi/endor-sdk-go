package sdk

import (
	"context"
	"fmt"
)

type AbstractResourceService struct {
	resource   string
	rootSchema *RootSchema
	repository *ResourceInstanceRepository[DynamicResource]
}

func NewAbstractResourceService(resource string, description string, additionalAttributes RootSchema) EndorService {
	rootSchema := NewSchema(DynamicResource{})
	// merge additional attributes
	for k, v := range *additionalAttributes.Schema.Properties {
		(*rootSchema.Definitions["DynamicResource"].Properties)[k] = v
	}
	for k, v := range additionalAttributes.Definitions {
		rootSchema.Definitions[k] = v
	}
	autogenerateID := true
	service := AbstractResourceService{
		resource:   resource,
		rootSchema: rootSchema,
		repository: NewResourceInstanceRepository[DynamicResource](resource, ResourceInstanceRepositoryOptions{
			AutoGenerateID: &autogenerateID,
		}),
	}
	return EndorService{
		Resource:    resource,
		Description: description,
		Methods: map[string]EndorServiceAction{
			"schema": NewAction(
				service.schema,
				fmt.Sprintf("Get the schema of the %s (%s)", resource, description),
			),
			"list": NewAction(
				service.list,
				fmt.Sprintf("Search for available list of %s (%s)", resource, description),
			),
			"create": NewConfigurableAction(
				EndorServiceActionOptions{
					Description:     fmt.Sprintf("Create the instance of %s (%s)", resource, description),
					Public:          false,
					ValidatePayload: true,
					InputSchema: &RootSchema{
						Schema: Schema{
							Reference: fmt.Sprintf("#/$defs/CreateDTO_%s", resource),
						},
						Definitions: func() map[string]Schema {
							defs := map[string]Schema{
								fmt.Sprintf("CreateDTO_%s", resource): {
									Type: ObjectType,
									Properties: &map[string]Schema{
										"data": rootSchema.Schema,
									},
								},
							}
							for k, v := range rootSchema.Definitions {
								defs[k] = v
							}
							return defs
						}(),
					},
				},
				service.create,
			),
			"instance": NewAction(
				service.instance,
				fmt.Sprintf("Get the instance of %s (%s)", resource, description),
			),
			"update": NewConfigurableAction(
				EndorServiceActionOptions{
					Description:     fmt.Sprintf("Update the existing instance of %s (%s)", resource, description),
					Public:          false,
					ValidatePayload: true,
					InputSchema: &RootSchema{
						Schema: Schema{
							Reference: fmt.Sprintf("#/$defs/UpdateByIdDTO_%s", resource),
						},
						Definitions: func() map[string]Schema {
							defs := map[string]Schema{
								fmt.Sprintf("UpdateByIdDTO_%s", resource): {
									Type: ObjectType,
									Properties: &map[string]Schema{
										"id": {
											Type: StringType,
										},
										"data": rootSchema.Schema,
									},
								},
							}
							for k, v := range rootSchema.Definitions {
								defs[k] = v
							}
							return defs
						}(),
					},
				},
				service.update,
			),
			"delete": NewAction(
				service.delete,
				fmt.Sprintf("Delete the existing instance of %s (%s)", resource, description),
			),
		},
	}
}

func (h *AbstractResourceService) schema(c *EndorContext[NoPayload]) (*Response[any], error) {
	return NewResponseBuilder[any]().AddSchema(h.rootSchema).Build(), nil
}

func (h *AbstractResourceService) instance(c *EndorContext[ReadInstanceDTO]) (*Response[*ResourceInstance[DynamicResource]], error) {
	instance, err := h.repository.Instance(context.TODO(), c.Payload)
	if err != nil {
		return nil, err
	}
	return NewResponseBuilder[*ResourceInstance[DynamicResource]]().AddData(&instance).AddSchema(h.rootSchema).Build(), nil
}

func (h *AbstractResourceService) list(c *EndorContext[ReadDTO]) (*Response[[]ResourceInstance[DynamicResource]], error) {
	list, err := h.repository.List(context.TODO(), c.Payload)
	if err != nil {
		return nil, err
	}
	return NewResponseBuilder[[]ResourceInstance[DynamicResource]]().AddData(&list).AddSchema(h.rootSchema).Build(), nil
}

func (h *AbstractResourceService) create(c *EndorContext[CreateDTO[ResourceInstance[DynamicResource]]]) (*Response[ResourceInstance[DynamicResource]], error) {
	_, err := h.repository.Create(context.TODO(), c.Payload)
	if err != nil {
		return nil, err
	}
	return NewResponseBuilder[ResourceInstance[DynamicResource]]().AddData(&c.Payload.Data).AddSchema(h.rootSchema).AddMessage(NewMessage(Info, fmt.Sprintf("%s created", h.resource))).Build(), nil
}

func (h *AbstractResourceService) update(c *EndorContext[UpdateByIdDTO[ResourceInstance[DynamicResource]]]) (*Response[ResourceInstance[DynamicResource]], error) {
	updated, err := h.repository.Update(context.TODO(), c.Payload)
	if err != nil {
		return nil, err
	}
	return NewResponseBuilder[ResourceInstance[DynamicResource]]().AddData(updated).AddSchema(h.rootSchema).AddMessage(NewMessage(Info, fmt.Sprintf("%s updated", h.resource))).Build(), nil
}

func (h *AbstractResourceService) delete(c *EndorContext[ReadInstanceDTO]) (*Response[any], error) {
	err := h.repository.Delete(context.TODO(), c.Payload)
	if err != nil {
		return nil, err
	}
	return NewResponseBuilder[any]().AddMessage(NewMessage(Info, fmt.Sprintf("%s deleted", h.resource))).Build(), nil
}
