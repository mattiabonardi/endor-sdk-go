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
		Resource:    "resource",
		Description: "Resource",
		Methods: map[string]sdk.EndorServiceAction{
			"schema": sdk.NewAction(
				resourceService.schema,
				"Get the schema of the resource",
			),
			string(sdk.ResourceTypeBase) + "/schema": sdk.NewAction(
				resourceService.schema,
				"Get the schema of the resource of type base",
			),
			string(sdk.ResourceTypeSpecialized) + "/schema": sdk.NewAction(
				resourceService.resourceSpecializedSchema,
				"Get the schema of the resource of type "+string(sdk.ResourceTypeSpecialized),
			),
			"list": sdk.NewAction(
				resourceService.list,
				"Search for available resources",
			),
			string(sdk.ResourceTypeBase) + "/list": sdk.NewAction(
				resourceService.resourceBaseList,
				"Search for available resources of type base",
			),
			string(sdk.ResourceTypeSpecialized) + "/list": sdk.NewAction(
				resourceService.resourceSpecializedList,
				"Search for available resources of type "+string(sdk.ResourceTypeSpecialized),
			),
			"instance": sdk.NewAction(
				resourceService.instance,
				"Get the specified instance of resources",
			),
			string(sdk.ResourceTypeBase) + "/instance": sdk.NewAction(
				resourceService.resourceBaseInstance,
				"Get the specified instance of resources of type "+string(sdk.ResourceTypeBase),
			),
			string(sdk.ResourceTypeSpecialized) + "/instance": sdk.NewAction(
				resourceService.resourceSpecializedInstance,
				"Get the specified instance of resources of type "+string(sdk.ResourceTypeSpecialized),
			),
		},
	}
	if configuration.GetConfig().HybridResourcesEnabled || configuration.GetConfig().DynamicResourcesEnabled {
		service.Methods[string(sdk.ResourceTypeBase)+"/update"] = sdk.NewAction(resourceService.resourceBaseUpdate, "Update an existing resource of type "+string(sdk.ResourceTypeBase))
		service.Methods[string(sdk.ResourceTypeSpecialized)+"/update"] = sdk.NewAction(resourceService.resourceSpecializedUpdate, "Update an existing resource of type "+string(sdk.ResourceTypeSpecialized))
	}
	if configuration.GetConfig().DynamicResourcesEnabled {
		service.Methods[string(sdk.ResourceTypeBase)+"/create"] = sdk.NewAction(resourceService.resourceBaseCreate, "Create a new resource "+string(sdk.ResourceTypeBase))
		service.Methods[string(sdk.ResourceTypeSpecialized)+"/create"] = sdk.NewAction(resourceService.resourceSpecializedCreate, "Create a new resource "+string(sdk.ResourceTypeSpecialized))
		service.Methods["delete"] = sdk.NewAction(resourceService.delete, "Delete an existing resource")
	}
	return service
}

type ResourceService struct {
	microServiceId string
	services       *[]sdk.EndorServiceInterface
}

func (h *ResourceService) resourceSpecializedSchema(c *sdk.EndorContext[sdk.NoPayload]) (*sdk.Response[any], error) {
	return sdk.NewResponseBuilder[any]().AddSchema(sdk.NewSchema(&sdk.ResourceSpecialized{})).Build(), nil
}

func (h *ResourceService) schema(c *sdk.EndorContext[sdk.NoPayload]) (*sdk.Response[any], error) {
	return sdk.NewResponseBuilder[any]().AddSchema(sdk.NewSchema(&sdk.Resource{})).Build(), nil
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

func (h *ResourceService) resourceSpecializedList(c *sdk.EndorContext[sdk.NoPayload]) (*sdk.Response[[]sdk.ResourceInterface], error) {
	resources, err := NewEndorServiceRepository(h.microServiceId, h.services).ResourceList()
	if err != nil {
		return nil, err
	}
	filtered := make([]sdk.ResourceInterface, 0, len(resources))
	for _, r := range resources {
		if r.GetID() != "resource" && r.GetID() != "resource-action" && r.GetType() == sdk.ResourceTypeSpecialized {
			filtered = append(filtered, r)
		}
	}
	resources = filtered
	return sdk.NewResponseBuilder[[]sdk.ResourceInterface]().AddData(&resources).AddSchema(sdk.NewSchema(&sdk.ResourceSpecialized{})).Build(), nil
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
	return sdk.NewResponseBuilder[sdk.ResourceInterface]().AddData(&resource.resource).AddSchema(sdk.NewSchema(&sdk.ResourceSpecialized{})).Build(), nil
}

func (h *ResourceService) resourceSpecializedInstance(c *sdk.EndorContext[sdk.ReadInstanceDTO]) (*sdk.Response[sdk.ResourceInterface], error) {
	resource, err := NewEndorServiceRepository(h.microServiceId, h.services).Instance(c.Payload)
	if err != nil {
		return nil, err
	}
	if resource.resource.GetType() != sdk.ResourceTypeSpecialized {
		return nil, sdk.NewNotFoundError(fmt.Errorf("resource %s of type %s not found", resource.resource.GetID(), sdk.ResourceTypeSpecialized))
	}
	return sdk.NewResponseBuilder[sdk.ResourceInterface]().AddData(&resource.resource).AddSchema(sdk.NewSchema(&sdk.ResourceSpecialized{})).Build(), nil
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

func (h *ResourceService) resourceSpecializedCreate(c *sdk.EndorContext[sdk.CreateDTO[sdk.ResourceInterface]]) (*sdk.Response[sdk.ResourceInterface], error) {
	// force type
	c.Payload.Data.SetType(sdk.ResourceTypeSpecialized)
	err := NewEndorServiceRepository(h.microServiceId, h.services).Create(c.Payload)
	if err != nil {
		return nil, err
	}
	return sdk.NewResponseBuilder[sdk.ResourceInterface]().AddData(&c.Payload.Data).AddSchema(sdk.NewSchema(&sdk.ResourceSpecialized{})).AddMessage(sdk.NewMessage(sdk.ResponseMessageGravityInfo, fmt.Sprintf("resource %s created", c.Payload.Data.GetID()))).Build(), nil
}

func (h *ResourceService) resourceBaseUpdate(c *sdk.EndorContext[sdk.UpdateByIdDTO[sdk.ResourceInterface]]) (*sdk.Response[sdk.ResourceInterface], error) {
	resource, err := NewEndorServiceRepository(h.microServiceId, h.services).UpdateOne(c.Payload)
	if err != nil {
		return nil, err
	}
	return sdk.NewResponseBuilder[sdk.ResourceInterface]().AddData(resource).AddSchema(sdk.NewSchema(&sdk.Resource{})).AddMessage(sdk.NewMessage(sdk.ResponseMessageGravityInfo, "resource updated")).Build(), nil
}

func (h *ResourceService) resourceSpecializedUpdate(c *sdk.EndorContext[sdk.UpdateByIdDTO[sdk.ResourceInterface]]) (*sdk.Response[sdk.ResourceInterface], error) {
	resource, err := NewEndorServiceRepository(h.microServiceId, h.services).UpdateOne(c.Payload)
	if err != nil {
		return nil, err
	}
	return sdk.NewResponseBuilder[sdk.ResourceInterface]().AddData(resource).AddSchema(sdk.NewSchema(&sdk.ResourceSpecialized{})).AddMessage(sdk.NewMessage(sdk.ResponseMessageGravityInfo, "resource updated")).Build(), nil
}

func (h *ResourceService) delete(c *sdk.EndorContext[sdk.ReadInstanceDTO]) (*sdk.Response[sdk.Resource], error) {
	err := NewEndorServiceRepository(h.microServiceId, h.services).DeleteOne(c.Payload)
	if err != nil {
		return nil, err
	}
	return sdk.NewResponseBuilder[sdk.Resource]().AddMessage(sdk.NewMessage(sdk.ResponseMessageGravityInfo, fmt.Sprintf("resource %s deleted", c.Payload.Id))).Build(), nil
}
