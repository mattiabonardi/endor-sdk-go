package sdk

import (
	"context"
	"sync"
)

// EntityInstanceRepositoryOptions defines configuration options for EntityInstanceRepository
type EntityInstanceRepositoryOptions struct {
	// AutoGenerateID determines whether IDs should be auto-generated or provided by the user
	// When true (default): Empty IDs are auto-generated using primitive.ObjectID.Hex()
	// When false: IDs must be provided by the user, empty IDs cause BadRequestError
	AutoGenerateID *bool
}

// RepositoryInterface defines the common operations shared by all repository types.
// It is the type returned by RepositoryRegistry.Get, allowing callers to perform
// common operations (e.g. resolving references) without knowing the concrete repository type.
type EndorRepositoryInterface interface {
	FindReferences(ctx context.Context, ids ReadInstancesDTO) (EntityReferenceGroupDescriptions, error)
	GetEntity() string
}

type EntityInstanceRepositoryInterface[T EntityInstanceInterface] interface {
	EndorRepositoryInterface
	Instance(ctx context.Context, dto ReadInstanceDTO) (*EntityInstance[T], error)
	List(ctx context.Context, dto ReadDTO) ([]EntityInstance[T], error)
	Create(ctx context.Context, dto CreateDTO[EntityInstance[T]]) (*EntityInstance[T], error)
	Delete(ctx context.Context, dto ReadInstanceDTO) error
	Update(ctx context.Context, dto UpdateByIdDTO[PartialEntityInstance[T]]) (*EntityInstance[T], error)

	InstanceWithReferences(ctx context.Context, dto ReadInstanceDTO) (*EntityInstance[T], EntityRefererenceGroup, error)
	ListWithReferences(ctx context.Context, dto ReadDTO) ([]EntityInstance[T], EntityRefererenceGroup, error)
}

// StaticEntityInstanceRepositoryOptions defines configuration options for StaticEntityInstanceRepository
// Mirrors EntityInstanceRepositoryOptions for consistency
type StaticEntityInstanceRepositoryOptions[T EntityInstanceInterface] struct {
	// AutoGenerateID determines whether IDs should be auto-generated or provided by the user
	// When true (default): Empty IDs are auto-generated using primitive.ObjectID.Hex()
	// When false: IDs must be provided by the user, empty IDs cause BadRequestError
	AutoGenerateID *bool

	Hooks StaticEntityInstanceRepositoryOptionsHooks[T]
}

type StaticEntityInstanceRepositoryOptionsHooks[T EntityInstanceInterface] struct {
	AfterFind func(entity T) error
}

// StaticEntityInstanceRepositoryInterface defines CRUD operations for working directly with model type T
// without the EntityInstance[T] wrapper. This provides a simpler interface for cases where
// the full entity instance structure (with metadata) is not needed.
type StaticEntityInstanceRepositoryInterface[T EntityInstanceInterface] interface {
	EndorRepositoryInterface
	Instance(ctx context.Context, dto ReadInstanceDTO) (T, error)
	List(ctx context.Context, dto ReadDTO) ([]T, error)
	Create(ctx context.Context, dto CreateDTO[T]) (T, error)
	Delete(ctx context.Context, dto ReadInstanceDTO) error
	Update(ctx context.Context, dto UpdateByIdDTO[map[string]interface{}]) (T, error)

	InstanceWithReferences(ctx context.Context, dto ReadInstanceDTO) (T, EntityRefererenceGroup, error)
	ListWithReferences(ctx context.Context, dto ReadDTO) ([]T, EntityRefererenceGroup, error)
}

type ReadInstanceDTO struct {
	Id string `json:"id,omitempty"`
}

type ReadInstancesDTO struct {
	Ids []string `json:"ids,omitempty"`
}

type CreateDTO[T any] struct {
	Data T `json:"data" binding:"required"`
}

type ReadDTO struct {
	Filter     map[string]interface{} `json:"filter"`
	Projection map[string]interface{} `json:"projection"`
}

// UpdateById defines the structure for updates with a generic data type
type UpdateByIdDTO[T any] struct {
	Id   string `json:"id,omitempty"`
	Data T      `json:"data" binding:"required"`
}

// #region Entity References

// define a map that contains for each entity id the relative description map
type EntityRefererenceGroup map[string]EntityReferenceGroupDescriptions

// define a map that contains for each id the relative description
type EntityReferenceGroupDescriptions map[string]string

// #endregion

// #region Repository Registry

// RepositoryRegistry is a unified, thread-safe singleton registry that holds both
// EntityInstanceRepositoryInterface and StaticEntityInstanceRepositoryInterface instances,
// keyed by a string name (typically the entity type name).
// Use Get to retrieve the common RepositoryInterface (e.g. to call FindReferences).
// Use the typed helpers GetEntityInstanceRepository / GetStaticEntityInstanceRepository
// when the concrete repository type is known.
type RepositoryRegistry struct {
	mu           sync.RWMutex
	repositories map[string]EndorRepositoryInterface
}

var (
	repositoryRegistryInstance *RepositoryRegistry
	repositoryRegistryOnce     sync.Once
)

// GetRepositoryRegistry returns the singleton RepositoryRegistry.
func GetRepositoryRegistry() *RepositoryRegistry {
	repositoryRegistryOnce.Do(func() {
		repositoryRegistryInstance = &RepositoryRegistry{
			repositories: make(map[string]EndorRepositoryInterface),
		}
	})
	return repositoryRegistryInstance
}

// Register stores a repository (either EntityInstanceRepositoryInterface or
// StaticEntityInstanceRepositoryInterface) under the given name.
func (r *RepositoryRegistry) Register(name string, repo EndorRepositoryInterface) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.repositories[name] = repo
}

// Get retrieves the RepositoryInterface stored under the given name.
// Use this when only common operations (e.g. FindReferences) are needed.
func (r *RepositoryRegistry) Get(name string) (EndorRepositoryInterface, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	repo, ok := r.repositories[name]
	return repo, ok
}

// DocumentRepositoryInterface extends EndorRepositoryInterface with the ability
// to list raw documents as a slice of maps, enabling aggregation pipelines to
// query any entity without knowing its concrete type at compile time.
type DocumentRepositoryInterface interface {
	EndorRepositoryInterface
	ListDocuments(ctx context.Context, dto ReadDTO) ([]map[string]interface{}, error)
}

// GetDocumentRepository retrieves a repository from the singleton RepositoryRegistry
// and asserts it to DocumentRepositoryInterface.
// Returns (nil, false) if the name is not registered or the type does not match.
func GetDocumentRepository(name string) (DocumentRepositoryInterface, bool) {
	repo, ok := GetRepositoryRegistry().Get(name)
	if !ok {
		return nil, false
	}
	docRepo, ok := repo.(DocumentRepositoryInterface)
	return docRepo, ok
}

// GetEntityInstanceRepository retrieves a repository from the singleton RepositoryRegistry
// and asserts it to EntityInstanceRepositoryInterface[T].
// Returns (nil, false) if the name is not registered or the type does not match.
func GetEntityInstanceRepository[T EntityInstanceInterface](name string) (EntityInstanceRepositoryInterface[T], bool) {
	repo, ok := GetRepositoryRegistry().Get(name)
	if !ok {
		return nil, false
	}
	typed, ok := repo.(EntityInstanceRepositoryInterface[T])
	return typed, ok
}

// GetStaticEntityInstanceRepository retrieves a repository from the singleton RepositoryRegistry
// and asserts it to StaticEntityInstanceRepositoryInterface[T].
// Returns (nil, false) if the name is not registered or the type does not match.
func GetStaticEntityInstanceRepository[T EntityInstanceInterface](name string) (StaticEntityInstanceRepositoryInterface[T], bool) {
	repo, ok := GetRepositoryRegistry().Get(name)
	if !ok {
		return nil, false
	}
	typed, ok := repo.(StaticEntityInstanceRepositoryInterface[T])
	return typed, ok
}

// #endregion
