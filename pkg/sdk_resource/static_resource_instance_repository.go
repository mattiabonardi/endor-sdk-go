package sdk_resource

import (
	"context"

	"github.com/mattiabonardi/endor-sdk-go/internal/repository"
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
)

// StaticResourceInstanceRepository provides a repository implementation that works directly
// with the model type T instead of ResourceInstance[T]. This is useful for simpler use cases
// where the metadata functionality is not required.
type StaticResourceInstanceRepository[T sdk.ResourceInstanceInterface] struct {
	repository sdk.StaticResourceInstanceRepositoryInterface[T]
}

// NewStaticResourceInstanceRepository creates a new static repository with default options
// Default behavior: AutoGenerateID = true (auto-generate ObjectID.Hex() as string)
func NewStaticResourceInstanceRepository[T sdk.ResourceInstanceInterface](resourceId string, options sdk.StaticResourceInstanceRepositoryOptions) *StaticResourceInstanceRepository[T] {
	if options.AutoGenerateID == nil {
		def := true
		options.AutoGenerateID = &def
	}
	return &StaticResourceInstanceRepository[T]{
		repository: repository.NewMongoStaticResourceInstanceRepository[T](resourceId, options),
	}
}

func (r *StaticResourceInstanceRepository[T]) Instance(ctx context.Context, dto sdk.ReadInstanceDTO) (T, error) {
	return r.repository.Instance(ctx, dto)
}

func (r *StaticResourceInstanceRepository[T]) List(ctx context.Context, dto sdk.ReadDTO) ([]T, error) {
	return r.repository.List(ctx, dto)
}

func (r *StaticResourceInstanceRepository[T]) Create(ctx context.Context, dto sdk.CreateDTO[T]) (T, error) {
	return r.repository.Create(ctx, dto)
}

func (r *StaticResourceInstanceRepository[T]) Delete(ctx context.Context, dto sdk.ReadInstanceDTO) error {
	return r.repository.Delete(ctx, dto)
}

func (r *StaticResourceInstanceRepository[T]) Update(ctx context.Context, dto sdk.UpdateByIdDTO[T]) (T, error) {
	return r.repository.Update(ctx, dto)
}
