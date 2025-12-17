package repository

import (
	"context"

	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
)

// EntityInstanceRepositoryOptions defines configuration options for EntityInstanceRepository
type EntityInstanceRepositoryOptions struct {
	// AutoGenerateID determines whether IDs should be auto-generated or provided by the user
	// When true (default): Empty IDs are auto-generated using primitive.ObjectID.Hex()
	// When false: IDs must be provided by the user, empty IDs cause BadRequestError
	AutoGenerateID *bool
}

type EntityInstanceRepositoryInterface[T sdk.EntityInstanceInterface] interface {
	Instance(ctx context.Context, dto sdk.ReadInstanceDTO) (*sdk.EntityInstance[T], error)
	List(ctx context.Context, dto sdk.ReadDTO) ([]sdk.EntityInstance[T], error)
	Create(ctx context.Context, dto sdk.CreateDTO[sdk.EntityInstance[T]]) (*sdk.EntityInstance[T], error)
	Delete(ctx context.Context, dto sdk.ReadInstanceDTO) error
	Update(ctx context.Context, dto sdk.UpdateByIdDTO[sdk.EntityInstance[T]]) (*sdk.EntityInstance[T], error)
}

type EntityInstanceRepository[T sdk.EntityInstanceInterface] struct {
	repository EntityInstanceRepositoryInterface[T]
}

// NewEntityInstanceRepository creates a new repository with default options
// Default behavior: AutoGenerateID = true (auto-generate ObjectID.Hex() as string)
func NewEntityInstanceRepository[T sdk.EntityInstanceInterface](entityId string, options EntityInstanceRepositoryOptions) *EntityInstanceRepository[T] {
	if options.AutoGenerateID == nil {
		def := true
		options.AutoGenerateID = &def
	}
	return &EntityInstanceRepository[T]{
		repository: NewMongoEntityInstanceRepository[T](entityId, options),
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

func (r *EntityInstanceRepository[T]) Update(ctx context.Context, dto sdk.UpdateByIdDTO[sdk.EntityInstance[T]]) (*sdk.EntityInstance[T], error) {
	return r.repository.Update(ctx, dto)
}
