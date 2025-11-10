package sdk

type EndorHybridService struct {
	Resource       string
	Description    string
	Priority       *int
	metadataSchema Schema
	methodsFn      func(getSchema func() Schema) map[string]EndorServiceAction
}

func NewHybridService(resource, description string) EndorHybridService {
	return EndorHybridService{
		Resource:    resource,
		Description: description,
	}
}

// Definizione dei metodi tramite funzione che riceve getSchema()
func (h EndorHybridService) WithActions(
	fn func(getSchema func() Schema) map[string]EndorServiceAction,
) EndorHybridService {
	h.methodsFn = fn
	return h
}

// Conversione in EndorService, schema iniettato dal framework
func (h EndorHybridService) ToEndorService(attrs Schema) EndorService {
	h.metadataSchema = attrs
	getSchema := func() Schema { return h.metadataSchema }

	methods := h.methodsFn(getSchema)

	return EndorService{
		Resource:    h.Resource,
		Description: h.Description,
		Priority:    h.Priority,
		Methods:     methods,
	}
}
