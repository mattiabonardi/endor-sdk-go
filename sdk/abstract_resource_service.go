package sdk

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
		Methods: map[string]EndorServiceMethod{
			"list": NewMethod(
				service.list,
			),
			/*"instance": NewMethod(
				resourceService.instance,
			),
			"create": NewMethod(
				resourceService.create,
			),
			"update": NewMethod(
				resourceService.update,
			),
			"delete": NewMethod(
				resourceService.delete,
			),*/
		},
	}
}

func (h *AbstractResourceService) list(c *EndorContext[NoPayload]) {
	c.End(NewResponseBuilder[[]any]().AddSchema(&h.definition.Schema).Build())
}
