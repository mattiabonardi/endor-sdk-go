package sdk

import (
	"context"
)

// StaticResourceInstanceRepositoryOptions defines configuration options for StaticResourceInstanceRepository
// Mirrors ResourceInstanceRepositoryOptions for consistency
type StaticResourceInstanceRepositoryOptions struct {
	// AutoGenerateID determines whether IDs should be auto-generated or provided by the user
	// When true (default): Empty IDs are auto-generated using primitive.ObjectID.Hex()
	// When false: IDs must be provided by the user, empty IDs cause BadRequestError
	AutoGenerateID *bool
}

// StaticResourceInstanceRepositoryInterface defines CRUD operations for working directly with model type T
// without the ResourceInstance[T] wrapper. This provides a simpler interface for cases where
// the full resource instance structure (with metadata) is not needed.
type StaticResourceInstanceRepositoryInterface[T ResourceInstanceInterface] interface {
	Instance(ctx context.Context, dto ReadInstanceDTO) (T, error)
	List(ctx context.Context, dto ReadDTO) ([]T, error)
	Create(ctx context.Context, dto CreateDTO[T]) (T, error)
	Delete(ctx context.Context, dto ReadInstanceDTO) error
	Update(ctx context.Context, dto UpdateByIdDTO[T]) (T, error)
}

// StaticResourceInstanceRepository provides a repository implementation that works directly
// with the model type T instead of ResourceInstance[T]. This is useful for simpler use cases
// where the metadata functionality is not required.
type StaticResourceInstanceRepository[T ResourceInstanceInterface] struct {
	repository StaticResourceInstanceRepositoryInterface[T]
}

// NewStaticResourceInstanceRepository creates a new static repository with default options
// Default behavior: AutoGenerateID = true (auto-generate ObjectID.Hex() as string)
func NewStaticResourceInstanceRepository[T ResourceInstanceInterface](resourceId string, options StaticResourceInstanceRepositoryOptions) *StaticResourceInstanceRepository[T] {
	if options.AutoGenerateID == nil {
		def := true
		options.AutoGenerateID = &def
	}
	return &StaticResourceInstanceRepository[T]{
		repository: NewMongoStaticResourceInstanceRepository[T](resourceId, options),
	}
}

func (r *StaticResourceInstanceRepository[T]) Instance(ctx context.Context, dto ReadInstanceDTO) (T, error) {
	return r.repository.Instance(ctx, dto)
}

func (r *StaticResourceInstanceRepository[T]) List(ctx context.Context, dto ReadDTO) ([]T, error) {
	return r.repository.List(ctx, dto)
}

func (r *StaticResourceInstanceRepository[T]) Create(ctx context.Context, dto CreateDTO[T]) (T, error) {
	return r.repository.Create(ctx, dto)
}

func (r *StaticResourceInstanceRepository[T]) Delete(ctx context.Context, dto ReadInstanceDTO) error {
	return r.repository.Delete(ctx, dto)
}

func (r *StaticResourceInstanceRepository[T]) Update(ctx context.Context, dto UpdateByIdDTO[T]) (T, error) {
	return r.repository.Update(ctx, dto)
}
