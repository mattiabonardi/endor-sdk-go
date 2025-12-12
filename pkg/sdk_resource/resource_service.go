package sdk_resource

import (
	"fmt"

	"github.com/mattiabonardi/endor-sdk-go/internal/configuration"
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
)

func NewResourceService(microServiceId string, services *[]sdk.EndorServiceInterface) sdk.EndorServiceInterface {
	resourceService := ResourceService{
		microServiceId: microServiceId,
		services:       services,
	}

	// hybrid category actions
	hybridCategoryActions := map[string]sdk.EndorServiceActionInterface{
		"schema": sdk.NewAction(
			resourceService.resourceHybridSchema,
			"Get the schema of the resource of type "+string(sdk.ResourceTypeHybrid),
		),
		"instance": sdk.NewAction(
			resourceService.resourceHybridInstance,
			"Get the specified instance of resources of type "+string(sdk.ResourceTypeHybrid),
		),
		"list": sdk.NewAction(
			resourceService.resourceHybridList,
			"Search for available resources of type "+string(sdk.ResourceTypeHybrid),
		),
	}
	// hybrid specialized actions
	hybridSpecializedCategoryActions := map[string]sdk.EndorServiceActionInterface{
		"schema": sdk.NewAction(
			resourceService.resourceHybridSpecializedSchema,
			"Get the schema of the resource of type "+string(sdk.ResourceTypeHybridSpecialized),
		),
		"instance": sdk.NewAction(
			resourceService.resourceHybridSpecializedInstance,
			"Get the specified instance of resources of type "+string(sdk.ResourceTypeHybridSpecialized),
		),
		"list": sdk.NewAction(
			resourceService.resourceHybridSpecializedList,
			"Search for available resources of type "+string(sdk.ResourceTypeHybridSpecialized),
		),
	}
	// dynamic category actions
	dynamicCategoryActions := map[string]sdk.EndorServiceActionInterface{
		"schema": sdk.NewAction(
			resourceService.resourceHybridSchema,
			"Get the schema of the resource of type "+string(sdk.ResourceTypeDynamic),
		),
		"instance": sdk.NewAction(
			resourceService.resourceDynamicInstance,
			"Get the specified instance of resources of type "+string(sdk.ResourceTypeDynamic),
		),
		"list": sdk.NewAction(
			resourceService.resourceDynamicList,
			"Search for available resources of type "+string(sdk.ResourceTypeDynamic),
		),
	}
	// dynamic specialized category actions
	dynamicSpecializedCategoryActions := map[string]sdk.EndorServiceActionInterface{
		"schema": sdk.NewAction(
			resourceService.resourceHybridSpecializedSchema,
			"Get the schema of the resource of type "+string(sdk.ResourceTypeDynamicSpecialized),
		),
		"instance": sdk.NewAction(
			resourceService.resourceDynamicSpecializedInstance,
			"Get the specified instance of resources of type "+string(sdk.ResourceTypeDynamicSpecialized),
		),
		"list": sdk.NewAction(
			resourceService.resourceDynamicSpecializedList,
			"Search for available resources of type "+string(sdk.ResourceTypeDynamicSpecialized),
		),
	}

	if configuration.GetConfig().HybridResourcesEnabled || configuration.GetConfig().DynamicResourcesEnabled {
		hybridCategoryActions["update"] = sdk.NewAction(resourceService.resourceHybridUpdate, "Update an existing resource of type "+string(sdk.ResourceTypeHybrid))
		hybridSpecializedCategoryActions["update"] = sdk.NewAction(resourceService.resourceHybridSpecializedUpdate, "Update an existing resource of type "+string(sdk.ResourceTypeHybridSpecialized))
	}
	if configuration.GetConfig().DynamicResourcesEnabled {
		dynamicCategoryActions["create"] = sdk.NewAction(resourceService.resourceDynamicCreate, "Create a new resource "+string(sdk.ResourceTypeDynamic))
		dynamicCategoryActions["update"] = sdk.NewAction(resourceService.resourceDynamicUpdate, "Update an existing resource of type "+string(sdk.ResourceTypeDynamic))
		dynamicCategoryActions["delete"] = sdk.NewAction(resourceService.resourceDynamicDelete, "Delete an existing resource "+string(sdk.ResourceTypeDynamic))

		dynamicSpecializedCategoryActions["create"] = sdk.NewAction(resourceService.resourceDynamicSpecializedCreate, "Create a new resource "+string(sdk.ResourceTypeDynamicSpecialized))
		dynamicSpecializedCategoryActions["update"] = sdk.NewAction(resourceService.resourceDynamicSpecializedUpdate, "Update an existing resource of type "+string(sdk.ResourceTypeDynamicSpecialized))
		dynamicSpecializedCategoryActions["delete"] = sdk.NewAction(resourceService.resourceDynamicDelete, "Delete an existing resource "+string(sdk.ResourceTypeDynamicSpecialized))
	}

	return NewEndorBaseSpecializedService[*sdk.Resource]("resource", "Resource").
		WithActions(map[string]sdk.EndorServiceActionInterface{
			"schema": sdk.NewAction(
				resourceService.schema,
				"Get the schema of the resource",
			),
			"list": sdk.NewAction(
				resourceService.list,
				"Search for available resources",
			),
			"instance": sdk.NewAction(
				resourceService.instance,
				"Get the specified instance of resources",
			)}).WithCategories(
		[]sdk.EndorBaseSpecializedServiceCategoryInterface{
			NewEndorBaseSpecializedServiceCategory[*sdk.Resource](string(sdk.ResourceTypeBase), "Base").
				WithActions(map[string]sdk.EndorServiceActionInterface{
					"schema": sdk.NewAction(
						resourceService.schema,
						"Get the schema of the resource of type "+string(sdk.ResourceTypeBase),
					),
					"instance": sdk.NewAction(
						resourceService.resourceBaseInstance,
						"Get the specified instance of resources of type "+string(sdk.ResourceTypeBase),
					),
					"list": sdk.NewAction(
						resourceService.resourceBaseList,
						"Search for available resources of type "+string(sdk.ResourceTypeBase),
					),
				}),
			NewEndorBaseSpecializedServiceCategory[*sdk.ResourceSpecialized](string(sdk.ResourceTypeBaseSpecialized), "Base specialized").
				WithActions(map[string]sdk.EndorServiceActionInterface{
					"schema": sdk.NewAction(
						resourceService.resourceBaseSpecializedSchema,
						"Get the schema of the resource of type "+string(sdk.ResourceTypeBaseSpecialized),
					),
					"instance": sdk.NewAction(
						resourceService.resourceBaseSpecializedInstance,
						"Get the specified instance of resources of type "+string(sdk.ResourceTypeBaseSpecialized),
					),
					"list": sdk.NewAction(
						resourceService.resourceBaseSpecializedList,
						"Search for available resources of type "+string(sdk.ResourceTypeBaseSpecialized),
					),
				}),
			NewEndorBaseSpecializedServiceCategory[*sdk.ResourceHybrid](string(sdk.ResourceTypeHybrid), "Hybrid").
				WithActions(hybridCategoryActions),
			NewEndorBaseSpecializedServiceCategory[*sdk.ResourceHybridSpecialized](string(sdk.ResourceTypeHybridSpecialized), "Hybrid specialized").
				WithActions(hybridSpecializedCategoryActions),
			NewEndorBaseSpecializedServiceCategory[*sdk.ResourceHybridSpecialized](string(sdk.ResourceTypeDynamic), "Dynamic").
				WithActions(dynamicCategoryActions),
			NewEndorBaseSpecializedServiceCategory[*sdk.ResourceHybridSpecialized](string(sdk.ResourceTypeDynamicSpecialized), "Dynamic specialized").
				WithActions(dynamicSpecializedCategoryActions),
		},
	)
}

type ResourceService struct {
	microServiceId string
	services       *[]sdk.EndorServiceInterface
}

func (h *ResourceService) schema(c *sdk.EndorContext[sdk.NoPayload]) (*sdk.Response[any], error) {
	return sdk.NewResponseBuilder[any]().AddSchema(sdk.NewSchema(&sdk.Resource{})).Build(), nil
}

func (h *ResourceService) resourceBaseSpecializedSchema(c *sdk.EndorContext[sdk.NoPayload]) (*sdk.Response[any], error) {
	return sdk.NewResponseBuilder[any]().AddSchema(sdk.NewSchema(&sdk.ResourceSpecialized{})).Build(), nil
}

func (h *ResourceService) resourceHybridSchema(c *sdk.EndorContext[sdk.NoPayload]) (*sdk.Response[any], error) {
	return sdk.NewResponseBuilder[any]().AddSchema(sdk.NewSchema(&sdk.ResourceHybrid{})).Build(), nil
}

func (h *ResourceService) resourceHybridSpecializedSchema(c *sdk.EndorContext[sdk.NoPayload]) (*sdk.Response[any], error) {
	return sdk.NewResponseBuilder[any]().AddSchema(sdk.NewSchema(&sdk.ResourceHybridSpecialized{})).Build(), nil
}

func (h *ResourceService) list(c *sdk.EndorContext[sdk.NoPayload]) (*sdk.Response[[]sdk.ResourceInterface], error) {
	resources, err := NewEndorServiceRepository(h.microServiceId, h.services).ResourceList()
	if err != nil {
		return nil, err
	}
	filtered := make([]sdk.ResourceInterface, 0, len(resources))
	for _, r := range resources {
		if *r.GetID() != "resource" && *r.GetID() != "resource-action" {
			filtered = append(filtered, r)
		}
	}
	resources = filtered
	return sdk.NewResponseBuilder[[]sdk.ResourceInterface]().AddData(&resources).AddSchema(sdk.NewSchema(&sdk.Resource{})).Build(), nil
}

func (h *ResourceService) resourceBaseList(c *sdk.EndorContext[sdk.NoPayload]) (*sdk.Response[[]sdk.ResourceInterface], error) {
	resources, err := NewEndorServiceRepository(h.microServiceId, h.services).ResourceList()
	if err != nil {
		return nil, err
	}
	filtered := make([]sdk.ResourceInterface, 0, len(resources))
	for _, r := range resources {
		if *r.GetID() != "resource" && *r.GetID() != "resource-action" && *r.GetCategoryType() == string(sdk.ResourceTypeBase) {
			filtered = append(filtered, r)
		}
	}
	resources = filtered
	return sdk.NewResponseBuilder[[]sdk.ResourceInterface]().AddData(&resources).AddSchema(sdk.NewSchema(&sdk.Resource{})).Build(), nil
}

func (h *ResourceService) resourceBaseSpecializedList(c *sdk.EndorContext[sdk.NoPayload]) (*sdk.Response[[]sdk.ResourceInterface], error) {
	resources, err := NewEndorServiceRepository(h.microServiceId, h.services).ResourceList()
	if err != nil {
		return nil, err
	}
	filtered := make([]sdk.ResourceInterface, 0, len(resources))
	for _, r := range resources {
		if *r.GetID() != "resource" && *r.GetID() != "resource-action" && *r.GetCategoryType() == string(sdk.ResourceTypeBaseSpecialized) {
			filtered = append(filtered, r)
		}
	}
	resources = filtered
	return sdk.NewResponseBuilder[[]sdk.ResourceInterface]().AddData(&resources).AddSchema(sdk.NewSchema(&sdk.ResourceSpecialized{})).Build(), nil
}

func (h *ResourceService) resourceHybridList(c *sdk.EndorContext[sdk.NoPayload]) (*sdk.Response[[]sdk.ResourceInterface], error) {
	resources, err := NewEndorServiceRepository(h.microServiceId, h.services).ResourceList()
	if err != nil {
		return nil, err
	}
	filtered := make([]sdk.ResourceInterface, 0, len(resources))
	for _, r := range resources {
		if *r.GetID() != "resource" && *r.GetID() != "resource-action" && *r.GetCategoryType() == string(sdk.ResourceTypeHybrid) {
			filtered = append(filtered, r)
		}
	}
	resources = filtered
	return sdk.NewResponseBuilder[[]sdk.ResourceInterface]().AddData(&resources).AddSchema(sdk.NewSchema(&sdk.ResourceHybrid{})).Build(), nil
}

func (h *ResourceService) resourceHybridSpecializedList(c *sdk.EndorContext[sdk.NoPayload]) (*sdk.Response[[]sdk.ResourceInterface], error) {
	resources, err := NewEndorServiceRepository(h.microServiceId, h.services).ResourceList()
	if err != nil {
		return nil, err
	}
	filtered := make([]sdk.ResourceInterface, 0, len(resources))
	for _, r := range resources {
		if *r.GetID() != "resource" && *r.GetID() != "resource-action" && *r.GetCategoryType() == string(sdk.ResourceTypeHybridSpecialized) {
			filtered = append(filtered, r)
		}
	}
	resources = filtered
	return sdk.NewResponseBuilder[[]sdk.ResourceInterface]().AddData(&resources).AddSchema(sdk.NewSchema(&sdk.ResourceHybridSpecialized{})).Build(), nil
}

func (h *ResourceService) resourceDynamicList(c *sdk.EndorContext[sdk.NoPayload]) (*sdk.Response[[]sdk.ResourceInterface], error) {
	resources, err := NewEndorServiceRepository(h.microServiceId, h.services).ResourceList()
	if err != nil {
		return nil, err
	}
	filtered := make([]sdk.ResourceInterface, 0, len(resources))
	for _, r := range resources {
		if *r.GetID() != "resource" && *r.GetID() != "resource-action" && *r.GetCategoryType() == string(sdk.ResourceTypeDynamic) {
			filtered = append(filtered, r)
		}
	}
	resources = filtered
	return sdk.NewResponseBuilder[[]sdk.ResourceInterface]().AddData(&resources).AddSchema(sdk.NewSchema(&sdk.ResourceHybrid{})).Build(), nil
}

func (h *ResourceService) resourceDynamicSpecializedList(c *sdk.EndorContext[sdk.NoPayload]) (*sdk.Response[[]sdk.ResourceInterface], error) {
	resources, err := NewEndorServiceRepository(h.microServiceId, h.services).ResourceList()
	if err != nil {
		return nil, err
	}
	filtered := make([]sdk.ResourceInterface, 0, len(resources))
	for _, r := range resources {
		if *r.GetID() != "resource" && *r.GetID() != "resource-action" && *r.GetCategoryType() == string(sdk.ResourceTypeDynamicSpecialized) {
			filtered = append(filtered, r)
		}
	}
	resources = filtered
	return sdk.NewResponseBuilder[[]sdk.ResourceInterface]().AddData(&resources).AddSchema(sdk.NewSchema(&sdk.ResourceHybrid{})).Build(), nil
}

func (h *ResourceService) instance(c *sdk.EndorContext[sdk.ReadInstanceDTO]) (*sdk.Response[sdk.ResourceInterface], error) {
	resource, err := NewEndorServiceRepository(h.microServiceId, h.services).Instance(c.Payload)
	if err != nil {
		return nil, err
	}
	return sdk.NewResponseBuilder[sdk.ResourceInterface]().AddData(&resource.resource).AddSchema(sdk.NewSchema(&sdk.Resource{})).Build(), nil
}

func (h *ResourceService) resourceBaseInstance(c *sdk.EndorContext[sdk.ReadInstanceDTO]) (*sdk.Response[sdk.ResourceInterface], error) {
	resource, err := NewEndorServiceRepository(h.microServiceId, h.services).Instance(c.Payload)
	if err != nil {
		return nil, err
	}
	if *resource.resource.GetCategoryType() != string(sdk.ResourceTypeBase) {
		return nil, sdk.NewNotFoundError(fmt.Errorf("resource %s of type %s not found", *resource.resource.GetID(), sdk.ResourceTypeBase))
	}
	return sdk.NewResponseBuilder[sdk.ResourceInterface]().AddData(&resource.resource).AddSchema(sdk.NewSchema(&sdk.Resource{})).Build(), nil
}

func (h *ResourceService) resourceBaseSpecializedInstance(c *sdk.EndorContext[sdk.ReadInstanceDTO]) (*sdk.Response[sdk.ResourceInterface], error) {
	resource, err := NewEndorServiceRepository(h.microServiceId, h.services).Instance(c.Payload)
	if err != nil {
		return nil, err
	}
	if *resource.resource.GetCategoryType() != string(sdk.ResourceTypeBaseSpecialized) {
		return nil, sdk.NewNotFoundError(fmt.Errorf("resource %s of type %s not found", *resource.resource.GetID(), sdk.ResourceTypeBaseSpecialized))
	}
	return sdk.NewResponseBuilder[sdk.ResourceInterface]().AddData(&resource.resource).AddSchema(sdk.NewSchema(&sdk.ResourceSpecialized{})).Build(), nil
}

func (h *ResourceService) resourceHybridInstance(c *sdk.EndorContext[sdk.ReadInstanceDTO]) (*sdk.Response[sdk.ResourceInterface], error) {
	resource, err := NewEndorServiceRepository(h.microServiceId, h.services).Instance(c.Payload)
	if err != nil {
		return nil, err
	}
	if *resource.resource.GetCategoryType() != string(sdk.ResourceTypeHybrid) {
		return nil, sdk.NewNotFoundError(fmt.Errorf("resource %s of type %s not found", *resource.resource.GetID(), sdk.ResourceTypeHybrid))
	}
	return sdk.NewResponseBuilder[sdk.ResourceInterface]().AddData(&resource.resource).AddSchema(sdk.NewSchema(&sdk.ResourceHybrid{})).Build(), nil
}

func (h *ResourceService) resourceHybridSpecializedInstance(c *sdk.EndorContext[sdk.ReadInstanceDTO]) (*sdk.Response[sdk.ResourceInterface], error) {
	resource, err := NewEndorServiceRepository(h.microServiceId, h.services).Instance(c.Payload)
	if err != nil {
		return nil, err
	}
	if *resource.resource.GetCategoryType() != string(sdk.ResourceTypeHybridSpecialized) {
		return nil, sdk.NewNotFoundError(fmt.Errorf("resource %s of type %s not found", *resource.resource.GetID(), sdk.ResourceTypeHybridSpecialized))
	}
	return sdk.NewResponseBuilder[sdk.ResourceInterface]().AddData(&resource.resource).AddSchema(sdk.NewSchema(&sdk.ResourceHybridSpecialized{})).Build(), nil
}

func (h *ResourceService) resourceDynamicInstance(c *sdk.EndorContext[sdk.ReadInstanceDTO]) (*sdk.Response[sdk.ResourceInterface], error) {
	resource, err := NewEndorServiceRepository(h.microServiceId, h.services).Instance(c.Payload)
	if err != nil {
		return nil, err
	}
	if *resource.resource.GetCategoryType() != string(sdk.ResourceTypeDynamic) {
		return nil, sdk.NewNotFoundError(fmt.Errorf("resource %s of type %s not found", *resource.resource.GetID(), sdk.ResourceTypeDynamic))
	}
	return sdk.NewResponseBuilder[sdk.ResourceInterface]().AddData(&resource.resource).AddSchema(sdk.NewSchema(&sdk.ResourceHybrid{})).Build(), nil
}

func (h *ResourceService) resourceDynamicSpecializedInstance(c *sdk.EndorContext[sdk.ReadInstanceDTO]) (*sdk.Response[sdk.ResourceInterface], error) {
	resource, err := NewEndorServiceRepository(h.microServiceId, h.services).Instance(c.Payload)
	if err != nil {
		return nil, err
	}
	if *resource.resource.GetCategoryType() != string(sdk.ResourceTypeDynamicSpecialized) {
		return nil, sdk.NewNotFoundError(fmt.Errorf("resource %s of type %s not found", *resource.resource.GetID(), sdk.ResourceTypeDynamicSpecialized))
	}
	return sdk.NewResponseBuilder[sdk.ResourceInterface]().AddData(&resource.resource).AddSchema(sdk.NewSchema(&sdk.ResourceHybrid{})).Build(), nil
}

func (h *ResourceService) resourceDynamicCreate(c *sdk.EndorContext[sdk.CreateDTO[sdk.ResourceInterface]]) (*sdk.Response[sdk.ResourceInterface], error) {
	// force type
	c.Payload.Data.SetCategoryType(string(sdk.ResourceTypeDynamic))
	err := NewEndorServiceRepository(h.microServiceId, h.services).Create(c.Payload)
	if err != nil {
		return nil, err
	}
	return sdk.NewResponseBuilder[sdk.ResourceInterface]().AddData(&c.Payload.Data).AddSchema(sdk.NewSchema(&sdk.ResourceHybrid{})).AddMessage(sdk.NewMessage(sdk.ResponseMessageGravityInfo, fmt.Sprintf("resource %s created", *c.Payload.Data.GetID()))).Build(), nil
}

func (h *ResourceService) resourceDynamicSpecializedCreate(c *sdk.EndorContext[sdk.CreateDTO[sdk.ResourceInterface]]) (*sdk.Response[sdk.ResourceInterface], error) {
	// force type
	c.Payload.Data.SetCategoryType(string(sdk.ResourceTypeDynamicSpecialized))
	err := NewEndorServiceRepository(h.microServiceId, h.services).Create(c.Payload)
	if err != nil {
		return nil, err
	}
	return sdk.NewResponseBuilder[sdk.ResourceInterface]().AddData(&c.Payload.Data).AddSchema(sdk.NewSchema(&sdk.ResourceHybridSpecialized{})).AddMessage(sdk.NewMessage(sdk.ResponseMessageGravityInfo, fmt.Sprintf("resource %s created", *c.Payload.Data.GetID()))).Build(), nil
}

func (h *ResourceService) resourceHybridUpdate(c *sdk.EndorContext[sdk.UpdateByIdDTO[sdk.ResourceHybrid]]) (*sdk.Response[sdk.ResourceInterface], error) {
	resource, err := NewEndorServiceRepository(h.microServiceId, h.services).Update(c.Payload.Id, &c.Payload.Data)
	if err != nil {
		return nil, err
	}
	return sdk.NewResponseBuilder[sdk.ResourceInterface]().AddData(resource).AddSchema(sdk.NewSchema(&sdk.ResourceHybrid{})).AddMessage(sdk.NewMessage(sdk.ResponseMessageGravityInfo, "resource updated")).Build(), nil
}

func (h *ResourceService) resourceHybridSpecializedUpdate(c *sdk.EndorContext[sdk.UpdateByIdDTO[sdk.ResourceHybridSpecialized]]) (*sdk.Response[sdk.ResourceInterface], error) {
	resource, err := NewEndorServiceRepository(h.microServiceId, h.services).Update(c.Payload.Id, &c.Payload.Data)
	if err != nil {
		return nil, err
	}
	return sdk.NewResponseBuilder[sdk.ResourceInterface]().AddData(resource).AddSchema(sdk.NewSchema(&sdk.ResourceHybridSpecialized{})).AddMessage(sdk.NewMessage(sdk.ResponseMessageGravityInfo, "resource updated")).Build(), nil
}

func (h *ResourceService) resourceDynamicUpdate(c *sdk.EndorContext[sdk.UpdateByIdDTO[sdk.ResourceInterface]]) (*sdk.Response[sdk.ResourceInterface], error) {
	resource, err := NewEndorServiceRepository(h.microServiceId, h.services).Update(c.Payload.Id, c.Payload.Data)
	if err != nil {
		return nil, err
	}
	return sdk.NewResponseBuilder[sdk.ResourceInterface]().AddData(resource).AddSchema(sdk.NewSchema(&sdk.ResourceHybrid{})).AddMessage(sdk.NewMessage(sdk.ResponseMessageGravityInfo, "resource updated")).Build(), nil
}

func (h *ResourceService) resourceDynamicSpecializedUpdate(c *sdk.EndorContext[sdk.UpdateByIdDTO[sdk.ResourceInterface]]) (*sdk.Response[sdk.ResourceInterface], error) {
	resource, err := NewEndorServiceRepository(h.microServiceId, h.services).Update(c.Payload.Id, c.Payload.Data)
	if err != nil {
		return nil, err
	}
	return sdk.NewResponseBuilder[sdk.ResourceInterface]().AddData(resource).AddSchema(sdk.NewSchema(&sdk.ResourceHybridSpecialized{})).AddMessage(sdk.NewMessage(sdk.ResponseMessageGravityInfo, "resource updated")).Build(), nil
}

func (h *ResourceService) resourceDynamicDelete(c *sdk.EndorContext[sdk.ReadInstanceDTO]) (*sdk.Response[sdk.Resource], error) {
	err := NewEndorServiceRepository(h.microServiceId, h.services).Delete(c.Payload)
	if err != nil {
		return nil, err
	}
	return sdk.NewResponseBuilder[sdk.Resource]().AddMessage(sdk.NewMessage(sdk.ResponseMessageGravityInfo, fmt.Sprintf("resource %s deleted", c.Payload.Id))).Build(), nil
}
