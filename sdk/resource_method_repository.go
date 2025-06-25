package sdk

func NewResourceMethodRepository() *ResourceMethodRepository {
	return &ResourceMethodRepository{}
}

type ResourceMethodRepository struct {
}

func (h *ResourceMethodRepository) List() ([]ResourceMethod, error) {
	methods := []ResourceMethod{}
	/*services := h.registry.GetServices()
	for _, s := range services {
		for methodName, methodCallback := range s.instance.Methods {
			methods = append(methods, h.createMethod(s, methodName, methodCallback))
		}
	}*/
	return methods, nil
}

func (h *ResourceMethodRepository) Instance(dto ReadInstanceDTO) (*ResourceMethod, error) {
	/*services := h.registry.GetInternalServices()
	for path, definition := range services {
		if strings.HasSuffix(path, dto.Id) {
			return h.createMethod(definition, "", definition.callback)
		}
	}*/
	return nil, nil
}

/*func (h *ResourceMethodRepository) createMethod(s ServiceDefinition, methodName string, method EndorServiceMethod) ResourceMethod {
	methodInstance := ResourceMethod{
		ID:          path.Join(s.instance.Resource, methodName),
		Resource:    s.instance.Resource,
		Description: method.GetOptions().Description,
	}
	payload, err := resolvePayloadType(method)
	if payload != reflect.TypeOf(NoPayload{}) && err != nil {
		methodInstance.Schema = NewSchemaByType(payload)
	}
	return methodInstance
}*/
