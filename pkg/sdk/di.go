package sdk

import (
	"fmt"
)

type EndorDIContainerInterface interface {
	GetRepositories() map[string]EndorRepositoryInterface
}

type RepositoryFactory func(session Session, container EndorDIContainerInterface) EndorRepositoryInterface

func GetStaticRepository[T EntityInstanceInterface](diContainer EndorDIContainerInterface, entityId string) (StaticEntityInstanceRepositoryInterface[T], error) {
	repo, ok := diContainer.GetRepositories()[entityId].(StaticEntityInstanceRepositoryInterface[T])
	if !ok {
		return nil, fmt.Errorf("repository for entity %s not found", entityId)
	}
	return repo, nil
}

func GetDynamicRepository[T EntityInstanceInterface](diContainer EndorDIContainerInterface, entityId string) (EntityInstanceRepositoryInterface[T], error) {
	repo, ok := diContainer.GetRepositories()[entityId].(EntityInstanceRepositoryInterface[T])
	if !ok {
		return nil, fmt.Errorf("repository for entity %s not found", entityId)
	}
	return repo, nil
}
