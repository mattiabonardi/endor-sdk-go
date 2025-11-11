package sdk

import (
	"context"
)

// ResourceInstanceRepositoryOptions defines configuration options for ResourceInstanceRepository
type ResourceInstanceRepositoryOptions struct {
	// AutoGenerateID determines whether IDs should be auto-generated or provided by the user
	// When true (default): Empty IDs are auto-generated using primitive.ObjectID.Hex()
	// When false: IDs must be provided by the user, empty IDs cause BadRequestError
	AutoGenerateID *bool
}

type ResourceInstanceRepositoryInterface[T ResourceInstanceInterface] interface {
	Instance(ctx context.Context, dto ReadInstanceDTO) (*ResourceInstance[T], error)
	List(ctx context.Context, dto ReadDTO) ([]ResourceInstance[T], error)
	Create(ctx context.Context, dto CreateDTO[ResourceInstance[T]]) (*ResourceInstance[T], error)
	Delete(ctx context.Context, dto ReadInstanceDTO) error
	Update(ctx context.Context, dto UpdateByIdDTO[ResourceInstance[T]]) (*ResourceInstance[T], error)
}

type ResourceInstanceRepository[T ResourceInstanceInterface] struct {
	repository ResourceInstanceRepositoryInterface[T]
}

// NewResourceInstanceRepository creates a new repository with default options
// Default behavior: AutoGenerateID = true (auto-generate ObjectID.Hex() as string)
func NewResourceInstanceRepository[T ResourceInstanceInterface](resourceId string, options ResourceInstanceRepositoryOptions) *ResourceInstanceRepository[T] {
	if options.AutoGenerateID == nil {
		def := true
		options.AutoGenerateID = &def
	}
	return &ResourceInstanceRepository[T]{
		repository: NewMongoResourceInstanceRepository[T](resourceId, options),
	}
}

func (r *ResourceInstanceRepository[T]) Instance(ctx context.Context, dto ReadInstanceDTO) (*ResourceInstance[T], error) {
	return r.repository.Instance(ctx, dto)
}

func (r *ResourceInstanceRepository[T]) List(ctx context.Context, dto ReadDTO) ([]ResourceInstance[T], error) {
	return r.repository.List(ctx, dto)
}

func (r *ResourceInstanceRepository[T]) Create(ctx context.Context, dto CreateDTO[ResourceInstance[T]]) (*ResourceInstance[T], error) {
	return r.repository.Create(ctx, dto)
}

func (r *ResourceInstanceRepository[T]) Delete(ctx context.Context, dto ReadInstanceDTO) error {
	return r.repository.Delete(ctx, dto)
}

func (r *ResourceInstanceRepository[T]) Update(ctx context.Context, dto UpdateByIdDTO[ResourceInstance[T]]) (*ResourceInstance[T], error) {
	return r.repository.Update(ctx, dto)
}
