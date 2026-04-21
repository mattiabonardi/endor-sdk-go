package sdk

import "fmt"

type EndorDIContainer interface {
	GetRepositories() map[string]EndorRepositoryInterface
}

type RepositoryFactory func(container EndorDIContainer) EndorRepositoryInterface

func GetStaticRepository[T EntityInstanceInterface](diContainer EndorDIContainer, entityId string) (StaticEntityInstanceRepositoryInterface[T], error) {
	repo, ok := diContainer.GetRepositories()[entityId].(StaticEntityInstanceRepositoryInterface[T])
	if !ok {
		return nil, fmt.Errorf("repository for entity %s not found", entityId)
	}
	return repo, nil
}

func GetDynamicRepository[T EntityInstanceInterface](diContainer EndorDIContainer, entityId string) (EntityInstanceRepositoryInterface[T], error) {
	repo, ok := diContainer.GetRepositories()[entityId].(EntityInstanceRepositoryInterface[T])
	if !ok {
		return nil, fmt.Errorf("repository for entity %s not found", entityId)
	}
	return repo, nil
}
