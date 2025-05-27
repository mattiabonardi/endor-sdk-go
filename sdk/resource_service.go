package sdk

import "fmt"

func NewResourceService(services []EndorService) EndorService {
	resourceService := ResourceService{
		Services: services,
	}
	return EndorService{
		Resource:    "resource",
		Description: "Resource",
		Methods: map[string]EndorServiceMethod{
			"list": NewMethod(
				AuthorizationHandler,
				resourceService.list,
			),
			"instance": NewMethod(
				ValidationHandler,
				AuthorizationHandler,
				resourceService.instance,
			),
		},
	}
}

type ResourceService struct {
	Services []EndorService
}

func (h *ResourceService) list(c *EndorContext[ResourceListDTO]) {
	if c.Payload.App != c.Session.App && c.Session.App != "admin" {
		c.Forbidden(fmt.Errorf("you can't access to resources of others application"))
		return
	}
	resources, err := NewResourceRepository(h.Services).List(c.Payload)
	if err != nil {
		c.InternalServerError(err)
		return
	}
	c.End(NewResponseBuilder[[]Resource]().AddData(&resources).AddSchema(NewSchema(&Resource{})).Build())
}

func (h *ResourceService) instance(c *EndorContext[ResourceInstanceDTO]) {
	if c.Payload.App != c.Session.App && c.Session.App != "admin" {
		c.Forbidden(fmt.Errorf("you can't access to resources of others application"))
		return
	}
	resource, err := NewResourceRepository(h.Services).Instance(c.Payload)
	if err != nil {
		c.InternalServerError(err)
		return
	}
	if resource == nil {
		c.NotFound(fmt.Errorf("not found resource %s", c.Payload.Id))
	} else {
		c.End(NewResponseBuilder[Resource]().AddData(resource).AddSchema(NewSchema(&Resource{})).Build())
	}
}
