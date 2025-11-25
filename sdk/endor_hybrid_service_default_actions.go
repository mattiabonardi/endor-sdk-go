package sdk

import (
	"context"
	"fmt"
)

func getDefaultActions(resource string, schema RootSchema, resourceDescription string) map[string]EndorServiceAction {
	// Crea repository usando DynamicResource come default (per ora)
	autogenerateID := true
	repository := NewResourceInstanceRepository[*DynamicResource](resource, ResourceInstanceRepositoryOptions{
		AutoGenerateID: &autogenerateID,
	})

	return map[string]EndorServiceAction{
		"schema": NewAction(
			func(c *EndorContext[NoPayload]) (*Response[any], error) {
				return defaultSchema(c, schema)
			},
			fmt.Sprintf("Get the schema of the %s (%s)", resource, resourceDescription),
		),
		"instance": NewAction(
			func(c *EndorContext[ReadInstanceDTO]) (*Response[*ResourceInstance[*DynamicResource]], error) {
				return defaultInstance(c, schema, repository)
			},
			fmt.Sprintf("Get the instance of %s (%s)", resource, resourceDescription),
		),
		"list": NewAction(
			func(c *EndorContext[ReadDTO]) (*Response[[]ResourceInstance[*DynamicResource]], error) {
				return defaultList(c, schema, repository)
			},
			fmt.Sprintf("Search for available list of %s (%s)", resource, resourceDescription),
		),
		"create": NewConfigurableAction(
			EndorServiceActionOptions{
				Description:     fmt.Sprintf("Create the instance of %s (%s)", resource, resourceDescription),
				Public:          false,
				ValidatePayload: true,
				InputSchema: &RootSchema{
					Schema: Schema{
						Type: ObjectType,
						Properties: &map[string]Schema{
							"data": schema.Schema,
						},
					},
				},
			},
			func(c *EndorContext[CreateDTO[ResourceInstance[*DynamicResource]]]) (*Response[ResourceInstance[*DynamicResource]], error) {
				return defaultCreate(c, schema, repository, resource)
			},
		),
		"update": NewConfigurableAction(
			EndorServiceActionOptions{
				Description:     fmt.Sprintf("Update the existing instance of %s (%s)", resource, resourceDescription),
				Public:          false,
				ValidatePayload: true,
				InputSchema: &RootSchema{
					Schema: Schema{
						Type: ObjectType,
						Properties: &map[string]Schema{
							"id": {
								Type: StringType,
							},
							"data": schema.Schema,
						},
					},
				},
			},
			func(c *EndorContext[UpdateByIdDTO[ResourceInstance[*DynamicResource]]]) (*Response[ResourceInstance[*DynamicResource]], error) {
				return defaultUpdate(c, schema, repository, resource)
			},
		),
		"delete": NewAction(
			func(c *EndorContext[ReadInstanceDTO]) (*Response[any], error) {
				return defaultDelete(c, repository, resource)
			},
			fmt.Sprintf("Delete the existing instance of %s (%s)", resource, resourceDescription),
		),
	}
}

func defaultSchema(_ *EndorContext[NoPayload], schema RootSchema) (*Response[any], error) {
	return NewResponseBuilder[any]().AddSchema(&schema).Build(), nil
}

func defaultInstance(c *EndorContext[ReadInstanceDTO], schema RootSchema, repository *ResourceInstanceRepository[*DynamicResource]) (*Response[*ResourceInstance[*DynamicResource]], error) {
	instance, err := repository.Instance(context.TODO(), c.Payload)
	if err != nil {
		return nil, err
	}
	return NewResponseBuilder[*ResourceInstance[*DynamicResource]]().AddData(&instance).AddSchema(&schema).Build(), nil
}

func defaultList(c *EndorContext[ReadDTO], schema RootSchema, repository *ResourceInstanceRepository[*DynamicResource]) (*Response[[]ResourceInstance[*DynamicResource]], error) {
	list, err := repository.List(context.TODO(), c.Payload)
	if err != nil {
		return nil, err
	}
	return NewResponseBuilder[[]ResourceInstance[*DynamicResource]]().AddData(&list).AddSchema(&schema).Build(), nil
}

func defaultCreate(c *EndorContext[CreateDTO[ResourceInstance[*DynamicResource]]], schema RootSchema, repository *ResourceInstanceRepository[*DynamicResource], resource string) (*Response[ResourceInstance[*DynamicResource]], error) {
	created, err := repository.Create(context.TODO(), c.Payload)
	if err != nil {
		return nil, err
	}
	return NewResponseBuilder[ResourceInstance[*DynamicResource]]().AddData(created).AddSchema(&schema).AddMessage(NewMessage(Info, fmt.Sprintf("%s created", resource))).Build(), nil
}

func defaultUpdate(c *EndorContext[UpdateByIdDTO[ResourceInstance[*DynamicResource]]], schema RootSchema, repository *ResourceInstanceRepository[*DynamicResource], resource string) (*Response[ResourceInstance[*DynamicResource]], error) {
	updated, err := repository.Update(context.TODO(), c.Payload)
	if err != nil {
		return nil, err
	}
	return NewResponseBuilder[ResourceInstance[*DynamicResource]]().AddData(updated).AddSchema(&schema).AddMessage(NewMessage(Info, fmt.Sprintf("%s updated", resource))).Build(), nil
}

func defaultDelete(c *EndorContext[ReadInstanceDTO], repository *ResourceInstanceRepository[*DynamicResource], resource string) (*Response[any], error) {
	err := repository.Delete(context.TODO(), c.Payload)
	if err != nil {
		return nil, err
	}
	return NewResponseBuilder[any]().AddMessage(NewMessage(Info, fmt.Sprintf("%s deleted", resource))).Build(), nil
}

func getDefaultActionsForCategory(resource string, schema RootSchema, resourceDescription string, categoryID string) map[string]EndorServiceAction {
	// Per ora usa DynamicResource come base model e struct vuota come category model
	// TODO: Implementare logica per detectare il modello corretto della categoria
	autogenerateID := true

	// Crea una categoria vuota per il repository specializzato
	emptyCategory := &CategorySpecialized[struct{}]{
		ID:                   categoryID,
		Description:          "Auto-generated category",
		StaticModel:          &struct{}{},
		AdditionalAttributes: "",
	}

	repository := NewResourceInstanceSpecializedRepository[*DynamicResource, struct{}](
		resource,
		ResourceInstanceRepositoryOptions{AutoGenerateID: &autogenerateID},
		emptyCategory,
	)

	return map[string]EndorServiceAction{
		"schema": NewAction(
			func(c *EndorContext[NoPayload]) (*Response[any], error) {
				return defaultSchema(c, schema)
			},
			fmt.Sprintf("Get the schema of the %s (%s) for category %s", resource, resourceDescription, categoryID),
		),
		"list": NewAction(
			func(c *EndorContext[ReadDTO]) (*Response[[]ResourceInstanceSpecialized[*DynamicResource, struct{}]], error) {
				return defaultListSpecialized(c, schema, repository)
			},
			fmt.Sprintf("Search for available list of %s (%s) for category %s", resource, resourceDescription, categoryID),
		),
		"create": NewConfigurableAction(
			EndorServiceActionOptions{
				Description:     fmt.Sprintf("Create the instance of %s (%s) for category %s", resource, resourceDescription, categoryID),
				Public:          false,
				ValidatePayload: true,
				InputSchema: &RootSchema{
					Schema: Schema{
						Type: ObjectType,
						Properties: &map[string]Schema{
							"data": schema.Schema,
						},
					},
				},
			},
			func(c *EndorContext[CreateDTO[ResourceInstanceSpecialized[*DynamicResource, struct{}]]]) (*Response[ResourceInstanceSpecialized[*DynamicResource, struct{}]], error) {
				return defaultCreateSpecialized(c, schema, repository, resource)
			},
		),
		"instance": NewAction(
			func(c *EndorContext[ReadInstanceDTO]) (*Response[*ResourceInstanceSpecialized[*DynamicResource, struct{}]], error) {
				return defaultInstanceSpecialized(c, schema, repository)
			},
			fmt.Sprintf("Get the instance of %s (%s) for category %s", resource, resourceDescription, categoryID),
		),
		"update": NewConfigurableAction(
			EndorServiceActionOptions{
				Description:     fmt.Sprintf("Update the existing instance of %s (%s) for category %s", resource, resourceDescription, categoryID),
				Public:          false,
				ValidatePayload: true,
				InputSchema: &RootSchema{
					Schema: Schema{
						Type: ObjectType,
						Properties: &map[string]Schema{
							"id": {
								Type: StringType,
							},
							"data": schema.Schema,
						},
					},
				},
			},
			func(c *EndorContext[UpdateByIdDTO[ResourceInstanceSpecialized[*DynamicResource, struct{}]]]) (*Response[ResourceInstanceSpecialized[*DynamicResource, struct{}]], error) {
				return defaultUpdateSpecialized(c, schema, repository, resource)
			},
		),
		"delete": NewAction(
			func(c *EndorContext[ReadInstanceDTO]) (*Response[any], error) {
				return defaultDeleteSpecialized(c, repository, resource)
			},
			fmt.Sprintf("Delete the existing instance of %s (%s) for category %s", resource, resourceDescription, categoryID),
		),
	}
}

// Metodi specializzati per repository specializzati (analoghe ai metodi base)
func defaultListSpecialized(c *EndorContext[ReadDTO], schema RootSchema, repository *ResourceInstanceSpecializedRepository[*DynamicResource, struct{}]) (*Response[[]ResourceInstanceSpecialized[*DynamicResource, struct{}]], error) {
	list, err := repository.List(context.TODO(), c.Payload)
	if err != nil {
		return nil, err
	}
	return NewResponseBuilder[[]ResourceInstanceSpecialized[*DynamicResource, struct{}]]().AddData(&list).AddSchema(&schema).Build(), nil
}

func defaultCreateSpecialized(c *EndorContext[CreateDTO[ResourceInstanceSpecialized[*DynamicResource, struct{}]]], schema RootSchema, repository *ResourceInstanceSpecializedRepository[*DynamicResource, struct{}], resource string) (*Response[ResourceInstanceSpecialized[*DynamicResource, struct{}]], error) {
	created, err := repository.Create(context.TODO(), c.Payload)
	if err != nil {
		return nil, err
	}
	return NewResponseBuilder[ResourceInstanceSpecialized[*DynamicResource, struct{}]]().AddData(created).AddSchema(&schema).AddMessage(NewMessage(Info, fmt.Sprintf("%s created (category)", resource))).Build(), nil
}

func defaultInstanceSpecialized(c *EndorContext[ReadInstanceDTO], schema RootSchema, repository *ResourceInstanceSpecializedRepository[*DynamicResource, struct{}]) (*Response[*ResourceInstanceSpecialized[*DynamicResource, struct{}]], error) {
	instance, err := repository.Instance(context.TODO(), c.Payload)
	if err != nil {
		return nil, err
	}
	return NewResponseBuilder[*ResourceInstanceSpecialized[*DynamicResource, struct{}]]().AddData(&instance).AddSchema(&schema).Build(), nil
}

func defaultUpdateSpecialized(c *EndorContext[UpdateByIdDTO[ResourceInstanceSpecialized[*DynamicResource, struct{}]]], schema RootSchema, repository *ResourceInstanceSpecializedRepository[*DynamicResource, struct{}], resource string) (*Response[ResourceInstanceSpecialized[*DynamicResource, struct{}]], error) {
	updated, err := repository.Update(context.TODO(), c.Payload)
	if err != nil {
		return nil, err
	}
	return NewResponseBuilder[ResourceInstanceSpecialized[*DynamicResource, struct{}]]().AddData(updated).AddSchema(&schema).AddMessage(NewMessage(Info, fmt.Sprintf("%s updated (category)", resource))).Build(), nil
}

func defaultDeleteSpecialized(c *EndorContext[ReadInstanceDTO], repository *ResourceInstanceSpecializedRepository[*DynamicResource, struct{}], resource string) (*Response[any], error) {
	err := repository.Delete(context.TODO(), c.Payload)
	if err != nil {
		return nil, err
	}
	return NewResponseBuilder[any]().AddMessage(NewMessage(Info, fmt.Sprintf("%s deleted (category)", resource))).Build(), nil
}
