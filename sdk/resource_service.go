package sdk

import (
	"fmt"
)

func NewResourceService(microServiceId string, services *[]EndorService, hybridServices *[]EndorHybridService) *EndorService {
	resourceService := ResourceService{
		microServiceId: microServiceId,
		services:       services,
		hybridServices: hybridServices,
	}
	service := &EndorService{
		Resource:    "resource",
		Description: "Resource",
		Methods: map[string]EndorServiceAction{
			"schema": NewAction(
				resourceService.schema,
				"Get the schema of the resource",
			),
			"list": NewAction(
				resourceService.list,
				"Search for available resources",
			),
			"instance": NewAction(
				resourceService.instance,
				"Get the specified instance of resources",
			),
		},
	}
	if GetConfig().HybridResourcesEnabled || GetConfig().DynamicResourcesEnabled {
		service.Methods["update"] = NewAction(resourceService.update, "Update an existing resource")
	}
	if GetConfig().DynamicResourcesEnabled {
		service.Methods["create"] = NewAction(resourceService.create, "Create a new resource")
		service.Methods["delete"] = NewAction(resourceService.delete, "Delete an existing resource")
	}
	return service
}

type ResourceService struct {
	microServiceId string
	services       *[]EndorService
	hybridServices *[]EndorHybridService
}

func (h *ResourceService) schema(c *EndorContext[NoPayload]) (*Response[any], error) {
	return NewResponseBuilder[any]().AddSchema(NewSchema(&Resource{})).Build(), nil
}

func (h *ResourceService) list(c *EndorContext[NoPayload]) (*Response[[]Resource], error) {
	resources, err := NewEndorServiceRepository(h.microServiceId, h.services, h.hybridServices).ResourceList()
	if err != nil {
		return nil, err
	}
	filtered := make([]Resource, 0, len(resources))
	for _, r := range resources {
		if r.ID != "resource" && r.ID != "resource-action" {
			filtered = append(filtered, r)
		}
	}
	resources = filtered
	return NewResponseBuilder[[]Resource]().AddData(&resources).AddSchema(NewSchema(&Resource{})).Build(), nil
}

func (h *ResourceService) instance(c *EndorContext[ReadInstanceDTO]) (*Response[Resource], error) {
	resource, err := NewEndorServiceRepository(h.microServiceId, h.services, h.hybridServices).Instance(c.Payload)
	if err != nil {
		return nil, err
	}
	return NewResponseBuilder[Resource]().AddData(&resource.resource).AddSchema(NewSchema(&Resource{})).Build(), nil
}

func (h *ResourceService) create(c *EndorContext[CreateDTO[Resource]]) (*Response[Resource], error) {
	err := NewEndorServiceRepository(h.microServiceId, h.services, h.hybridServices).Create(c.Payload)
	if err != nil {
		return nil, err
	}
	return NewResponseBuilder[Resource]().AddData(&c.Payload.Data).AddSchema(NewSchema(&Resource{})).AddMessage(NewMessage(Info, fmt.Sprintf("resource %s created", c.Payload.Data.ID))).Build(), nil
}

func (h *ResourceService) update(c *EndorContext[UpdateByIdDTO[Resource]]) (*Response[Resource], error) {
	resource, err := NewEndorServiceRepository(h.microServiceId, h.services, h.hybridServices).UpdateOne(c.Payload)
	if err != nil {
		return nil, err
	}
	return NewResponseBuilder[Resource]().AddData(resource).AddSchema(NewSchema(&Resource{})).AddMessage(NewMessage(Info, "resource updated")).Build(), nil
}

func (h *ResourceService) delete(c *EndorContext[ReadInstanceDTO]) (*Response[Resource], error) {
	err := NewEndorServiceRepository(h.microServiceId, h.services, h.hybridServices).DeleteOne(c.Payload)
	if err != nil {
		return nil, err
	}
	return NewResponseBuilder[Resource]().AddMessage(NewMessage(Info, fmt.Sprintf("resource %s deleted", c.Payload.Id))).Build(), nil
}
