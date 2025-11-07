package sdk

import (
	"context"
)

type ResourceInstanceRepositoryInterface[ID comparable, T ResourceInstanceInterface[ID]] interface {
	Instance(ctx context.Context, dto ReadInstanceDTO[ID]) (*ResourceInstance[ID, T], error)
	List(ctx context.Context, dto ReadDTO) ([]ResourceInstance[ID, T], error)
	Create(ctx context.Context, dto CreateDTO[ResourceInstance[ID, T]]) (*ResourceInstance[ID, T], error)
	Delete(ctx context.Context, dto DeleteByIdDTO[ID]) error
	Update(ctx context.Context, dto UpdateByIdDTO[ID, ResourceInstance[ID, T]]) (*ResourceInstance[ID, T], error)
}

type ResourceInstanceRepository[ID comparable, T ResourceInstanceInterface[ID]] struct {
	repository ResourceInstanceRepositoryInterface[ID, T]
}

func NewResourceInstanceRepository[ID comparable, T ResourceInstanceInterface[ID]](resourceId string) *ResourceInstanceRepository[ID, T] {
	return &ResourceInstanceRepository[ID, T]{
		repository: NewMongoResourceInstanceRepository[ID, T](resourceId),
	}
}

func (r *ResourceInstanceRepository[ID, T]) Instance(ctx context.Context, dto ReadInstanceDTO[ID]) (*ResourceInstance[ID, T], error) {
	return r.repository.Instance(ctx, dto)
}

func (r *ResourceInstanceRepository[ID, T]) List(ctx context.Context, dto ReadDTO) ([]ResourceInstance[ID, T], error) {
	return r.repository.List(ctx, dto)
}

func (r *ResourceInstanceRepository[ID, T]) Create(ctx context.Context, dto CreateDTO[ResourceInstance[ID, T]]) (*ResourceInstance[ID, T], error) {
	return r.repository.Create(ctx, dto)
}

func (r *ResourceInstanceRepository[ID, T]) Delete(ctx context.Context, dto DeleteByIdDTO[ID]) error {
	return r.repository.Delete(ctx, dto)
}

func (r *ResourceInstanceRepository[ID, T]) Update(ctx context.Context, dto UpdateByIdDTO[ID, ResourceInstance[ID, T]]) (*ResourceInstance[ID, T], error) {
	return r.repository.Update(ctx, dto)
}
