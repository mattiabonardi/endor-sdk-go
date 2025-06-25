package sdk

func NewResourceMethodService() EndorResource {
	resourceMethodService := ResourceMethodService{}
	return EndorResource{
		Resource:    "resource-method",
		Description: "Resource Method",
		Methods: map[string]EndorResourceAction{
			"schema": NewAction(
				resourceMethodService.schema,
				"Get the schema of the resource method",
			),
			"list": NewAction(
				resourceMethodService.list,
				"Search for available resources",
			),
			"instance": NewAction(
				resourceMethodService.instance,
				"Get the specified instance of resources",
			),
		},
	}
}

type ResourceMethodService struct{}

func (h *ResourceMethodService) schema(c *EndorContext[NoPayload]) (*Response[any], error) {
	return NewResponseBuilder[any]().AddSchema(NewSchema(&ResourceMethod{})).Build(), nil
}

func (h *ResourceMethodService) list(c *EndorContext[NoPayload]) (*Response[[]ResourceMethod], error) {
	resourceMethods, err := NewResourceMethodRepository().List()
	if err != nil {
		return nil, err
	}
	return NewResponseBuilder[[]ResourceMethod]().AddData(&resourceMethods).AddSchema(NewSchema(&ResourceMethod{})).Build(), nil
}

func (h *ResourceMethodService) instance(c *EndorContext[ReadInstanceDTO]) (*Response[ResourceMethod], error) {
	resourceMethod, err := NewResourceMethodRepository().Instance(c.Payload)
	if err != nil {
		return nil, err
	}
	return NewResponseBuilder[ResourceMethod]().AddData(resourceMethod).AddSchema(NewSchema(&ResourceMethod{})).Build(), nil
}
