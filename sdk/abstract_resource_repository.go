package sdk

import (
	"fmt"
)

type AbstractResourceRepository struct {
	resourceDefinition ResourceDefinition
	sliceContexts      []ResourceSliceContext
}

func NewAbstractResourceRepository(def ResourceDefinition, customizedRepositories map[string]ResurceRepositoryInterface) (*AbstractResourceRepository, error) {
	sliceContexts := []ResourceSliceContext{}
	for _, ds := range def.DataSources {
		switch ds.GetType() {
		case "mongodb":
			mongoDS, ok := ds.(*MongoDataSource)
			if !ok {
				return nil, fmt.Errorf("expected *MongoDataSource, got %T", ds)
			}
			mongoRepo, err := NewMongoAbstractResourceRepository(GetConfig().EndorDynamicResourceDBName, def, *mongoDS)
			if err != nil {
				return nil, err
			}
			sliceContexts = append(sliceContexts, ResourceSliceContext{
				dataSource: ds,
				repository: mongoRepo,
			})
		default:
			return nil, fmt.Errorf("unsupported data source type: %s", ds.GetType())
		}
	}

	//TODO: handle multiple repositories
	if len(sliceContexts) == 0 {
		return nil, fmt.Errorf("datasource not defined")
	}

	return &AbstractResourceRepository{
		resourceDefinition: def,
		sliceContexts:      sliceContexts,
	}, nil
}

func (r *AbstractResourceRepository) Instance(dto ReadInstanceDTO) (any, error) {
	return r.sliceContexts[0].repository.Instance(dto)
}

func (r *AbstractResourceRepository) List(dto ReadDTO) ([]any, error) {
	return r.sliceContexts[0].repository.List(dto)
}

func (r *AbstractResourceRepository) Create(dto CreateDTO[any]) error {
	return r.sliceContexts[0].repository.Create(dto)
}

func (r *AbstractResourceRepository) Update(dto UpdateByIdDTO[any]) (any, error) {
	return r.sliceContexts[0].repository.Update(dto)
}

func (r *AbstractResourceRepository) Delete(dto DeleteByIdDTO) error {
	return r.sliceContexts[0].repository.Delete(dto)
}
