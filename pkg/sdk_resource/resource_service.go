package sdk_resource

import (
	"fmt"

	"github.com/mattiabonardi/endor-sdk-go/internal/configuration"
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
)

func NewResourceService(microServiceId string, services *[]sdk.EndorServiceInterface) *sdk.EndorService {
	resourceService := ResourceService{
		microServiceId: microServiceId,
		services:       services,
	}
	service := &sdk.EndorService{
		Resource:            "resource",
		ResourceDescription: "Resource",
		Actions: map[string]sdk.EndorServiceActionInterface{
			"schema": sdk.NewAction(
				resourceService.schema,
				"Get the schema of the resource",
			),
			string(sdk.ResourceTypeBase) + "/schema": sdk.NewAction(
				resourceService.schema,
				"Get the schema of the resource of type base",
			),
			string(sdk.ResourceTypeHybrid) + "/schema": sdk.NewAction(
				resourceService.resourceHybridSchema,
				"Get the schema of the resource of type "+string(sdk.ResourceTypeHybrid),
			),
			string(sdk.ResourceTypeHybridSpecialized) + "/schema": sdk.NewAction(
				resourceService.resourceHybridSpecializedSchema,
				"Get the schema of the resource of type "+string(sdk.ResourceTypeHybridSpecialized),
			),
			"list": sdk.NewAction(
				resourceService.list,
				"Search for available resources",
			),
			string(sdk.ResourceTypeBase) + "/list": sdk.NewAction(
				resourceService.resourceBaseList,
				"Search for available resources of type base",
			),
			string(sdk.ResourceTypeHybrid) + "/list": sdk.NewAction(
				resourceService.resourceHybridList,
				"Search for available resources of type "+string(sdk.ResourceTypeHybrid),
			),
			string(sdk.ResourceTypeHybridSpecialized) + "/list": sdk.NewAction(
				resourceService.resourceHybridSpecializedList,
				"Search for available resources of type "+string(sdk.ResourceTypeHybridSpecialized),
			),
			"instance": sdk.NewAction(
				resourceService.instance,
				"Get the specified instance of resources",
			),
			string(sdk.ResourceTypeBase) + "/instance": sdk.NewAction(
				resourceService.resourceBaseInstance,
				"Get the specified instance of resources of type "+string(sdk.ResourceTypeBase),
			),
			string(sdk.ResourceTypeHybrid) + "/instance": sdk.NewAction(
				resourceService.resourceHybridInstance,
				"Get the specified instance of resources of type "+string(sdk.ResourceTypeHybrid),
			),
			string(sdk.ResourceTypeHybridSpecialized) + "/instance": sdk.NewAction(
				resourceService.resourceHybridSpecializedInstance,
				"Get the specified instance of resources of type "+string(sdk.ResourceTypeHybridSpecialized),
			),
		},
	}
	if configuration.GetConfig().HybridResourcesEnabled || configuration.GetConfig().DynamicResourcesEnabled {
		service.Actions[string(sdk.ResourceTypeBase)+"/update"] = sdk.NewAction(resourceService.resourceBaseUpdate, "Update an existing resource of type "+string(sdk.ResourceTypeBase))
		service.Actions[string(sdk.ResourceTypeHybrid)+"/update"] = sdk.NewAction(resourceService.resourceHybridUpdate, "Update an existing resource of type "+string(sdk.ResourceTypeHybrid))
		service.Actions[string(sdk.ResourceTypeHybridSpecialized)+"/update"] = sdk.NewAction(resourceService.resourceHybridSpecializedUpdate, "Update an existing resource of type "+string(sdk.ResourceTypeHybridSpecialized))
	}
	if configuration.GetConfig().DynamicResourcesEnabled {
		service.Actions[string(sdk.ResourceTypeBase)+"/create"] = sdk.NewAction(resourceService.resourceBaseCreate, "Create a new resource "+string(sdk.ResourceTypeBase))
		service.Actions[string(sdk.ResourceTypeHybrid)+"/create"] = sdk.NewAction(resourceService.resourceHybridCreate, "Create a new resource "+string(sdk.ResourceTypeHybrid))
		service.Actions[string(sdk.ResourceTypeHybridSpecialized)+"/create"] = sdk.NewAction(resourceService.resourceHybridSpecializedCreate, "Create a new resource "+string(sdk.ResourceTypeHybridSpecialized))
		service.Actions["delete"] = sdk.NewAction(resourceService.delete, "Delete an existing resource")
	}
	return service
}

type ResourceService struct {
	microServiceId string
	services       *[]sdk.EndorServiceInterface
}

func (h *ResourceService) schema(c *sdk.EndorContext[sdk.NoPayload]) (*sdk.Response[any], error) {
	return sdk.NewResponseBuilder[any]().AddSchema(sdk.NewSchema(&sdk.Resource{})).Build(), nil
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
		if r.GetID() != "resource" && r.GetID() != "resource-action" {
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
		if r.GetID() != "resource" && r.GetID() != "resource-action" && r.GetType() == sdk.ResourceTypeBase {
			filtered = append(filtered, r)
		}
	}
	resources = filtered
	return sdk.NewResponseBuilder[[]sdk.ResourceInterface]().AddData(&resources).AddSchema(sdk.NewSchema(&sdk.Resource{})).Build(), nil
}

func (h *ResourceService) resourceHybridList(c *sdk.EndorContext[sdk.NoPayload]) (*sdk.Response[[]sdk.ResourceInterface], error) {
	resources, err := NewEndorServiceRepository(h.microServiceId, h.services).ResourceList()
	if err != nil {
		return nil, err
	}
	filtered := make([]sdk.ResourceInterface, 0, len(resources))
	for _, r := range resources {
		if r.GetID() != "resource" && r.GetID() != "resource-action" && r.GetType() == sdk.ResourceTypeHybrid {
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
		if r.GetID() != "resource" && r.GetID() != "resource-action" && r.GetType() == sdk.ResourceTypeHybridSpecialized {
			filtered = append(filtered, r)
		}
	}
	resources = filtered
	return sdk.NewResponseBuilder[[]sdk.ResourceInterface]().AddData(&resources).AddSchema(sdk.NewSchema(&sdk.ResourceHybridSpecialized{})).Build(), nil
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
	if resource.resource.GetType() != sdk.ResourceTypeBase {
		return nil, sdk.NewNotFoundError(fmt.Errorf("resource %s of type %s not found", resource.resource.GetID(), sdk.ResourceTypeBase))
	}
	return sdk.NewResponseBuilder[sdk.ResourceInterface]().AddData(&resource.resource).AddSchema(sdk.NewSchema(&sdk.ResourceHybridSpecialized{})).Build(), nil
}

func (h *ResourceService) resourceHybridInstance(c *sdk.EndorContext[sdk.ReadInstanceDTO]) (*sdk.Response[sdk.ResourceInterface], error) {
	resource, err := NewEndorServiceRepository(h.microServiceId, h.services).Instance(c.Payload)
	if err != nil {
		return nil, err
	}
	if resource.resource.GetType() != sdk.ResourceTypeHybrid {
		return nil, sdk.NewNotFoundError(fmt.Errorf("resource %s of type %s not found", resource.resource.GetID(), sdk.ResourceTypeHybrid))
	}
	return sdk.NewResponseBuilder[sdk.ResourceInterface]().AddData(&resource.resource).AddSchema(sdk.NewSchema(&sdk.ResourceHybrid{})).Build(), nil
}

func (h *ResourceService) resourceHybridSpecializedInstance(c *sdk.EndorContext[sdk.ReadInstanceDTO]) (*sdk.Response[sdk.ResourceInterface], error) {
	resource, err := NewEndorServiceRepository(h.microServiceId, h.services).Instance(c.Payload)
	if err != nil {
		return nil, err
	}
	if resource.resource.GetType() != sdk.ResourceTypeHybridSpecialized {
		return nil, sdk.NewNotFoundError(fmt.Errorf("resource %s of type %s not found", resource.resource.GetID(), sdk.ResourceTypeHybridSpecialized))
	}
	return sdk.NewResponseBuilder[sdk.ResourceInterface]().AddData(&resource.resource).AddSchema(sdk.NewSchema(&sdk.ResourceHybridSpecialized{})).Build(), nil
}

func (h *ResourceService) resourceBaseCreate(c *sdk.EndorContext[sdk.CreateDTO[sdk.ResourceInterface]]) (*sdk.Response[sdk.ResourceInterface], error) {
	// force type
	c.Payload.Data.SetType(sdk.ResourceTypeBase)
	err := NewEndorServiceRepository(h.microServiceId, h.services).Create(c.Payload)
	if err != nil {
		return nil, err
	}
	return sdk.NewResponseBuilder[sdk.ResourceInterface]().AddData(&c.Payload.Data).AddSchema(sdk.NewSchema(&sdk.Resource{})).AddMessage(sdk.NewMessage(sdk.ResponseMessageGravityInfo, fmt.Sprintf("resource %s created", c.Payload.Data.GetID()))).Build(), nil
}

func (h *ResourceService) resourceHybridCreate(c *sdk.EndorContext[sdk.CreateDTO[sdk.ResourceInterface]]) (*sdk.Response[sdk.ResourceInterface], error) {
	// force type
	c.Payload.Data.SetType(sdk.ResourceTypeHybrid)
	err := NewEndorServiceRepository(h.microServiceId, h.services).Create(c.Payload)
	if err != nil {
		return nil, err
	}
	return sdk.NewResponseBuilder[sdk.ResourceInterface]().AddData(&c.Payload.Data).AddSchema(sdk.NewSchema(&sdk.ResourceHybrid{})).AddMessage(sdk.NewMessage(sdk.ResponseMessageGravityInfo, fmt.Sprintf("resource %s created", c.Payload.Data.GetID()))).Build(), nil
}

func (h *ResourceService) resourceHybridSpecializedCreate(c *sdk.EndorContext[sdk.CreateDTO[sdk.ResourceInterface]]) (*sdk.Response[sdk.ResourceInterface], error) {
	// force type
	c.Payload.Data.SetType(sdk.ResourceTypeHybridSpecialized)
	err := NewEndorServiceRepository(h.microServiceId, h.services).Create(c.Payload)
	if err != nil {
		return nil, err
	}
	return sdk.NewResponseBuilder[sdk.ResourceInterface]().AddData(&c.Payload.Data).AddSchema(sdk.NewSchema(&sdk.ResourceHybridSpecialized{})).AddMessage(sdk.NewMessage(sdk.ResponseMessageGravityInfo, fmt.Sprintf("resource %s created", c.Payload.Data.GetID()))).Build(), nil
}

func (h *ResourceService) resourceBaseUpdate(c *sdk.EndorContext[sdk.UpdateByIdDTO[sdk.ResourceInterface]]) (*sdk.Response[sdk.ResourceInterface], error) {
	resource, err := NewEndorServiceRepository(h.microServiceId, h.services).UpdateOne(c.Payload)
	if err != nil {
		return nil, err
	}
	return sdk.NewResponseBuilder[sdk.ResourceInterface]().AddData(resource).AddSchema(sdk.NewSchema(&sdk.Resource{})).AddMessage(sdk.NewMessage(sdk.ResponseMessageGravityInfo, "resource updated")).Build(), nil
}

func (h *ResourceService) resourceHybridUpdate(c *sdk.EndorContext[sdk.UpdateByIdDTO[sdk.ResourceInterface]]) (*sdk.Response[sdk.ResourceInterface], error) {
	resource, err := NewEndorServiceRepository(h.microServiceId, h.services).UpdateOne(c.Payload)
	if err != nil {
		return nil, err
	}
	return sdk.NewResponseBuilder[sdk.ResourceInterface]().AddData(resource).AddSchema(sdk.NewSchema(&sdk.ResourceHybrid{})).AddMessage(sdk.NewMessage(sdk.ResponseMessageGravityInfo, "resource updated")).Build(), nil
}

func (h *ResourceService) resourceHybridSpecializedUpdate(c *sdk.EndorContext[sdk.UpdateByIdDTO[sdk.ResourceInterface]]) (*sdk.Response[sdk.ResourceInterface], error) {
	resource, err := NewEndorServiceRepository(h.microServiceId, h.services).UpdateOne(c.Payload)
	if err != nil {
		return nil, err
	}
	return sdk.NewResponseBuilder[sdk.ResourceInterface]().AddData(resource).AddSchema(sdk.NewSchema(&sdk.ResourceHybridSpecialized{})).AddMessage(sdk.NewMessage(sdk.ResponseMessageGravityInfo, "resource updated")).Build(), nil
}

func (h *ResourceService) delete(c *sdk.EndorContext[sdk.ReadInstanceDTO]) (*sdk.Response[sdk.Resource], error) {
	err := NewEndorServiceRepository(h.microServiceId, h.services).DeleteOne(c.Payload)
	if err != nil {
		return nil, err
	}
	return sdk.NewResponseBuilder[sdk.Resource]().AddMessage(sdk.NewMessage(sdk.ResponseMessageGravityInfo, fmt.Sprintf("resource %s deleted", c.Payload.Id))).Build(), nil
}
