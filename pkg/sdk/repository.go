package sdk

import (
	"context"
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
	RawList(ctx context.Context, dto ReadDTO) ([]map[string]interface{}, error)
	GetSchema() *RootSchema
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
