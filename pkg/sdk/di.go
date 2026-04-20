package sdk

type EndorDIContainer interface {
	GetRepositories() map[string]EndorRepositoryInterface
}

type RepositoryFactory func(container EndorDIContainer) EndorRepositoryInterface

func GetStaticRepository[T EntityInstanceInterface](diContainer EndorDIContainer, entityId string) StaticEntityInstanceRepositoryInterface[T] {
	repo, ok := diContainer.GetRepositories()[entityId].(StaticEntityInstanceRepositoryInterface[T])
	if !ok {
		// Gestione errore: il repository non esiste o non è del tipo corretto per T
		panic("Repository not found or type mismatch for entity: " + entityId)
	}
	return repo
}

func GetDynamicRepository[T EntityInstanceInterface](diContainer EndorDIContainer, entityId string) EntityInstanceRepositoryInterface[T] {
	repo, ok := diContainer.GetRepositories()[entityId].(EntityInstanceRepositoryInterface[T])
	if !ok {
		// Gestione errore: il repository non esiste o non è del tipo corretto per T
		panic("Repository not found or type mismatch for entity: " + entityId)
	}
	return repo
}
