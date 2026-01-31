package sdk_entity

import (
	"context"

	"github.com/mattiabonardi/endor-sdk-go/internal/repository"
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
)

type EntityInstanceRepository[T sdk.EntityInstanceInterface] struct {
	repository sdk.EntityInstanceRepositoryInterface[T]
}

// NewEntityInstanceRepository creates a new repository with default options
// Default behavior: AutoGenerateID = true (auto-generate ObjectID.Hex() as string)
func NewEntityInstanceRepository[T sdk.EntityInstanceInterface](entityId string, options sdk.EntityInstanceRepositoryOptions) *EntityInstanceRepository[T] {
	if options.AutoGenerateID == nil {
		def := true
		options.AutoGenerateID = &def
	}
	return &EntityInstanceRepository[T]{
		repository: repository.NewMongoEntityInstanceRepository[T](entityId, options),
	}
}

func (r *EntityInstanceRepository[T]) Instance(ctx context.Context, dto sdk.ReadInstanceDTO) (*sdk.EntityInstance[T], error) {
	return r.repository.Instance(ctx, dto)
}

func (r *EntityInstanceRepository[T]) List(ctx context.Context, dto sdk.ReadDTO) ([]sdk.EntityInstance[T], error) {
	return r.repository.List(ctx, dto)
}

func (r *EntityInstanceRepository[T]) Create(ctx context.Context, dto sdk.CreateDTO[sdk.EntityInstance[T]]) (*sdk.EntityInstance[T], error) {
	return r.repository.Create(ctx, dto)
}

func (r *EntityInstanceRepository[T]) Delete(ctx context.Context, dto sdk.ReadInstanceDTO) error {
	return r.repository.Delete(ctx, dto)
}

func (r *EntityInstanceRepository[T]) Update(ctx context.Context, dto sdk.UpdateById[sdk.PartialEntityInstance[T]]) (*sdk.EntityInstance[T], error) {
	return r.repository.Update(ctx, dto)
}
