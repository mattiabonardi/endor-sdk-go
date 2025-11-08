package sdk

import (
	"context"
)

type ResourceInstanceRepositoryInterface[T ResourceInstanceInterface] interface {
	Instance(ctx context.Context, dto ReadInstanceDTO) (*ResourceInstance[T], error)
	List(ctx context.Context, dto ReadDTO) ([]ResourceInstance[T], error)
	Create(ctx context.Context, dto CreateDTO[ResourceInstance[T]]) (*ResourceInstance[T], error)
	Delete(ctx context.Context, dto DeleteByIdDTO) error
	Update(ctx context.Context, dto UpdateByIdDTO[ResourceInstance[T]]) (*ResourceInstance[T], error)
}

type ResourceInstanceRepository[T ResourceInstanceInterface] struct {
	repository ResourceInstanceRepositoryInterface[T]
}

func NewResourceInstanceRepository[T ResourceInstanceInterface](resourceId string) *ResourceInstanceRepository[T] {
	return &ResourceInstanceRepository[T]{
		repository: NewMongoResourceInstanceRepository[T](resourceId),
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

func (r *ResourceInstanceRepository[T]) Delete(ctx context.Context, dto DeleteByIdDTO) error {
	return r.repository.Delete(ctx, dto)
}

func (r *ResourceInstanceRepository[T]) Update(ctx context.Context, dto UpdateByIdDTO[ResourceInstance[T]]) (*ResourceInstance[T], error) {
	return r.repository.Update(ctx, dto)
}
