package sdk

import (
	"context"
)

type ResourceInstanceRepositoryInterface[T any, ID comparable] interface {
	Instance(ctx context.Context, dto ReadInstanceDTO[ID]) (*ResourceInstance[T], error)
	List(ctx context.Context, dto ReadDTO) ([]ResourceInstance[T], error)
	Create(ctx context.Context, dto CreateDTO[ResourceInstance[T]]) (*ResourceInstance[T], error)
	Delete(ctx context.Context, dto DeleteByIdDTO[ID]) error
	Update(ctx context.Context, dto UpdateByIdDTO[ResourceInstance[T], ID]) (*ResourceInstance[T], error)
}

type ResourceInstanceRepository[T any, ID comparable] struct {
	repository ResourceInstanceRepositoryInterface[T, ID]
}

func NewResourceInstanceRepository[T any, ID comparable](resourceId string) *ResourceInstanceRepository[T, ID] {
	return &ResourceInstanceRepository[T, ID]{
		repository: NewMongoResourceInstanceRepository[T, ID](resourceId),
	}
}

func (r *ResourceInstanceRepository[T, ID]) Instance(ctx context.Context, dto ReadInstanceDTO[ID]) (*ResourceInstance[T], error) {
	return r.repository.Instance(ctx, dto)
}

func (r *ResourceInstanceRepository[T, ID]) List(ctx context.Context, dto ReadDTO) ([]ResourceInstance[T], error) {
	return r.repository.List(ctx, dto)
}

func (r *ResourceInstanceRepository[T, ID]) Create(ctx context.Context, dto CreateDTO[ResourceInstance[T]]) (*ResourceInstance[T], error) {
	return r.repository.Create(ctx, dto)
}

func (r *ResourceInstanceRepository[T, ID]) Delete(ctx context.Context, dto DeleteByIdDTO[ID]) error {
	return r.repository.Delete(ctx, dto)
}

func (r *ResourceInstanceRepository[T, ID]) Update(ctx context.Context, dto UpdateByIdDTO[ResourceInstance[T], ID]) (*ResourceInstance[T], error) {
	return r.repository.Update(ctx, dto)
}
