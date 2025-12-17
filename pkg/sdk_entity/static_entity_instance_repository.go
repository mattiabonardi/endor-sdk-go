package sdk_entity

import (
	"context"

	"github.com/mattiabonardi/endor-sdk-go/internal/repository"
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
)

// StaticEntityInstanceRepository provides a repository implementation that works directly
// with the model type T instead of EntityInstance[T]. This is useful for simpler use cases
// where the metadata functionality is not required.
type StaticEntityInstanceRepository[T sdk.EntityInstanceInterface] struct {
	repository sdk.StaticEntityInstanceRepositoryInterface[T]
}

// NewStaticEntityInstanceRepository creates a new static repository with default options
// Default behavior: AutoGenerateID = true (auto-generate ObjectID.Hex() as string)
func NewStaticEntityInstanceRepository[T sdk.EntityInstanceInterface](entityId string, options sdk.StaticEntityInstanceRepositoryOptions) *StaticEntityInstanceRepository[T] {
	if options.AutoGenerateID == nil {
		def := true
		options.AutoGenerateID = &def
	}
	return &StaticEntityInstanceRepository[T]{
		repository: repository.NewMongoStaticEntityInstanceRepository[T](entityId, options),
	}
}

func (r *StaticEntityInstanceRepository[T]) Instance(ctx context.Context, dto sdk.ReadInstanceDTO) (T, error) {
	return r.repository.Instance(ctx, dto)
}

func (r *StaticEntityInstanceRepository[T]) List(ctx context.Context, dto sdk.ReadDTO) ([]T, error) {
	return r.repository.List(ctx, dto)
}

func (r *StaticEntityInstanceRepository[T]) Create(ctx context.Context, dto sdk.CreateDTO[T]) (T, error) {
	return r.repository.Create(ctx, dto)
}

func (r *StaticEntityInstanceRepository[T]) Delete(ctx context.Context, dto sdk.ReadInstanceDTO) error {
	return r.repository.Delete(ctx, dto)
}

func (r *StaticEntityInstanceRepository[T]) Update(ctx context.Context, dto sdk.UpdateByIdDTO[T]) (T, error) {
	return r.repository.Update(ctx, dto)
}
