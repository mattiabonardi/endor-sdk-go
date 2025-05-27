package sdk

import "github.com/mattiabonardi/endor-sdk-go/sdk/utils"

func NewResourceRepository(services []EndorService) *ResourceRepository {
	return &ResourceRepository{
		Services: services,
	}
}

type ResourceRepository struct {
	Services []EndorService
}

func (h *ResourceRepository) List(options ResourceListDTO) ([]Resource, error) {
	resources := []Resource{}
	for _, service := range h.Services {
		if len(service.Apps) == 0 || utils.StringElemMatch(service.Apps, options.App) {
			resource := Resource{
				ID:          service.Resource,
				Description: service.Description,
			}
			for methodName, method := range service.Methods {
				payload, _ := resolvePayloadType(method)
				requestSchema := NewSchemaByType(payload)
				if methodName == "create" {
					stringSchema, _ := requestSchema.ToYAML()
					resource.Schema = stringSchema
				}
			}
			resources = append(resources, resource)
		}
	}
	return resources, nil
}

func (h *ResourceRepository) Instance(options ResourceInstanceDTO) (*Resource, error) {
	for _, service := range h.Services {
		if service.Resource == options.Id {
			if len(service.Apps) == 0 || utils.StringElemMatch(service.Apps, options.App) {
				resource := Resource{
					ID:          service.Resource,
					Description: service.Description,
				}
				for methodName, method := range service.Methods {
					payload, _ := resolvePayloadType(method)
					requestSchema := NewSchemaByType(payload)
					if methodName == "create" {
						stringSchema, _ := requestSchema.ToYAML()
						resource.Schema = stringSchema
					}
				}
				return &resource, nil
			}
		}
	}
	return nil, nil
}
