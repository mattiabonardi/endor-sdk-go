package sdk

func NewResourceService() EndorService {
	resourceService := ResourceService{}
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

type ResourceService struct{}

func (h *ResourceService) list(c *EndorContext[NoPayload]) {
	//c.End(sdk.NewResponseBuilder[[]models.App]().AddData(&apps).AddSchema(sdk.NewSchema(&models.App{})).Build())
}

func (h *ResourceService) instance(c *EndorContext[ReadInstanceDTO]) {
	//c.End(sdk.NewResponseBuilder[models.App]().AddData(app).AddSchema(sdk.NewSchema(&models.App{})).Build())
}
