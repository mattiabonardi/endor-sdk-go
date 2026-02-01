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

type EntityInstanceRepositoryInterface[T EntityInstanceInterface] interface {
	Instance(ctx context.Context, dto ReadInstanceDTO) (*EntityInstance[T], error)
	List(ctx context.Context, dto ReadDTO) ([]EntityInstance[T], error)
	Create(ctx context.Context, dto CreateDTO[EntityInstance[T]]) (*EntityInstance[T], error)
	Delete(ctx context.Context, dto ReadInstanceDTO) error
	Update(ctx context.Context, dto UpdateByIdDTO[PartialEntityInstance[T]]) (*EntityInstance[T], error)
}

// StaticEntityInstanceRepositoryOptions defines configuration options for StaticEntityInstanceRepository
// Mirrors EntityInstanceRepositoryOptions for consistency
type StaticEntityInstanceRepositoryOptions struct {
	// AutoGenerateID determines whether IDs should be auto-generated or provided by the user
	// When true (default): Empty IDs are auto-generated using primitive.ObjectID.Hex()
	// When false: IDs must be provided by the user, empty IDs cause BadRequestError
	AutoGenerateID *bool
}

// StaticEntityInstanceRepositoryInterface defines CRUD operations for working directly with model type T
// without the EntityInstance[T] wrapper. This provides a simpler interface for cases where
// the full entity instance structure (with metadata) is not needed.
type StaticEntityInstanceRepositoryInterface[T EntityInstanceInterface] interface {
	Instance(ctx context.Context, dto ReadInstanceDTO) (T, error)
	List(ctx context.Context, dto ReadDTO) ([]T, error)
	Create(ctx context.Context, dto CreateDTO[T]) (T, error)
	Delete(ctx context.Context, dto ReadInstanceDTO) error
	Update(ctx context.Context, dto UpdateByIdDTO[map[string]interface{}]) (T, error)
}

type ReadInstanceDTO struct {
	Id string `json:"id,omitempty"`
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
