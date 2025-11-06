package sdk

func NewResourceActionService(microServiceId string, services *[]EndorService, databaseName string) *EndorService {
	resourceMethodService := ResourceActionService{
		microServiceId: microServiceId,
		services:       services,
		databaseName:   databaseName,
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
	databaseName   string
}

func (h *ResourceActionService) schema(c *EndorContext[NoPayload]) (*Response[any], error) {
	return NewResponseBuilder[any]().AddSchema(NewSchema(&ResourceAction{})).Build(), nil
}

func (h *ResourceActionService) list(c *EndorContext[NoPayload]) (*Response[[]ResourceAction], error) {
	resourceMethods, err := NewEndorServiceRepository(h.microServiceId, h.services, h.databaseName).ResourceActionList()
	if err != nil {
		return nil, err
	}
	return NewResponseBuilder[[]ResourceAction]().AddData(&resourceMethods).AddSchema(NewSchema(&ResourceAction{})).Build(), nil
}

func (h *ResourceActionService) instance(c *EndorContext[ReadInstanceDTO[string]]) (*Response[ResourceAction], error) {
	resourceAction, err := NewEndorServiceRepository(h.microServiceId, h.services, h.databaseName).ActionInstance(c.Payload)
	if err != nil {
		return nil, err
	}
	return NewResponseBuilder[ResourceAction]().AddData(&resourceAction.resourceAction).AddSchema(NewSchema(&ResourceAction{})).Build(), nil
}
