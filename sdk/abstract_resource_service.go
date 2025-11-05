package sdk

import (
	"fmt"
)

type AbstractResourceService struct {
	resource   string
	definition ResourceDefinition
}

func NewAbstractResourceService(resource string, description string, definition ResourceDefinition) EndorService {
	service := AbstractResourceService{
		resource:   resource,
		definition: definition,
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
						Definitions: map[string]Schema{
							fmt.Sprintf("CreateDTO_%s", resource): {
								Type: ObjectType,
								Properties: &map[string]Schema{
									"data": definition.Schema.Schema,
								},
							},
						},
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
						Definitions: map[string]Schema{
							fmt.Sprintf("UpdateByIdDTO_%s", resource): {
								Type: ObjectType,
								Properties: &map[string]Schema{
									"id": {
										Type: StringType,
									},
									"data": definition.Schema.Schema,
								},
							},
						},
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
	return NewResponseBuilder[any]().AddSchema(h.createSchema()).Build(), nil
}

func (h *AbstractResourceService) instance(c *EndorContext[ReadInstanceDTO]) (*Response[any], error) {
	repo, err := NewAbstractResourceRepository(h.definition, nil)
	if err != nil {
		return nil, err
	}
	instance, err := repo.Instance(c.Payload)
	if err != nil {
		return nil, err
	}
	return NewResponseBuilder[any]().AddData(&instance).AddSchema(h.createSchema()).Build(), nil
}

func (h *AbstractResourceService) list(c *EndorContext[ReadDTO]) (*Response[[]any], error) {
	repo, err := NewAbstractResourceRepository(h.definition, nil)
	if err != nil {
		return nil, err
	}
	list, err := repo.List(c.Payload)
	if err != nil {
		return nil, err
	}
	return NewResponseBuilder[[]any]().AddData(&list).AddSchema(h.createSchema()).Build(), nil
}

func (h *AbstractResourceService) create(c *EndorContext[CreateDTO[any]]) (*Response[any], error) {
	repo, err := NewAbstractResourceRepository(h.definition, nil)
	if err != nil {
		return nil, err
	}
	err = repo.Create(c.Payload)
	if err != nil {
		return nil, err
	}
	return NewResponseBuilder[any]().AddData(&c.Payload.Data).AddSchema(h.createSchema()).AddMessage(NewMessage(Info, fmt.Sprintf("%s created", h.resource))).Build(), nil
}

func (h *AbstractResourceService) update(c *EndorContext[UpdateByIdDTO[any]]) (*Response[any], error) {
	repo, err := NewAbstractResourceRepository(h.definition, nil)
	if err != nil {
		return nil, err
	}
	updated, err := repo.Update(c.Payload)
	if err != nil {
		return nil, err
	}
	return NewResponseBuilder[any]().AddData(&updated).AddSchema(h.createSchema()).AddMessage(NewMessage(Info, fmt.Sprintf("%s updated", h.resource))).Build(), nil
}

func (h *AbstractResourceService) delete(c *EndorContext[DeleteByIdDTO]) (*Response[any], error) {
	repo, err := NewAbstractResourceRepository(h.definition, nil)
	if err != nil {
		return nil, err
	}
	err = repo.Delete(c.Payload)
	if err != nil {
		return nil, err
	}
	return NewResponseBuilder[any]().AddMessage(NewMessage(Info, fmt.Sprintf("%s deleted", h.resource))).Build(), nil
}

func (h *AbstractResourceService) createSchema() *RootSchema {
	schema := h.definition.Schema
	// id
	if h.definition.Id != "" {
		schema.UISchema = &UISchema{
			Id: &h.definition.Id,
		}
	}
	return &schema
}
