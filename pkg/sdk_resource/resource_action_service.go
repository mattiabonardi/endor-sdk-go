package sdk_resource

func NewResourceActionService(microServiceId string, services *[]EndorService, hybridServices *[]EndorHybridService) *EndorService {
	resourceMethodService := ResourceActionService{
		microServiceId: microServiceId,
		services:       services,
	}
	return &EndorService{
		Resource:    "resource-action",
		Description: "Resource Action",
		Methods: map[string]EndorServiceAction{
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

type ResourceActionService struct {
	microServiceId string
	services       *[]EndorService
	hybridServices *[]EndorHybridService
}

func (h *ResourceActionService) schema(c *EndorContext[NoPayload]) (*Response[any], error) {
	return NewResponseBuilder[any]().AddSchema(NewSchema(&ResourceAction{})).Build(), nil
}

func (h *ResourceActionService) list(c *EndorContext[NoPayload]) (*Response[[]ResourceAction], error) {
	resourceMethods, err := NewEndorServiceRepository(h.microServiceId, h.services, h.hybridServices).ResourceActionList()
	if err != nil {
		return nil, err
	}
	return NewResponseBuilder[[]ResourceAction]().AddData(&resourceMethods).AddSchema(NewSchema(&ResourceAction{})).Build(), nil
}

func (h *ResourceActionService) instance(c *EndorContext[ReadInstanceDTO]) (*Response[ResourceAction], error) {
	resourceAction, err := NewEndorServiceRepository(h.microServiceId, h.services, h.hybridServices).ActionInstance(c.Payload)
	if err != nil {
		return nil, err
	}
	return NewResponseBuilder[ResourceAction]().AddData(&resourceAction.resourceAction).AddSchema(NewSchema(&ResourceAction{})).Build(), nil
}
