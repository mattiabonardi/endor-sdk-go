package repository

import (
	"context"

	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
)

// ResourceInstanceRepositoryOptions defines configuration options for ResourceInstanceRepository
type ResourceInstanceRepositoryOptions struct {
	// AutoGenerateID determines whether IDs should be auto-generated or provided by the user
	// When true (default): Empty IDs are auto-generated using primitive.ObjectID.Hex()
	// When false: IDs must be provided by the user, empty IDs cause BadRequestError
	AutoGenerateID *bool
}

type ResourceInstanceRepositoryInterface[T sdk.ResourceInstanceInterface] interface {
	Instance(ctx context.Context, dto sdk.ReadInstanceDTO) (*sdk.ResourceInstance[T], error)
	List(ctx context.Context, dto sdk.ReadDTO) ([]sdk.ResourceInstance[T], error)
	Create(ctx context.Context, dto sdk.CreateDTO[sdk.ResourceInstance[T]]) (*sdk.ResourceInstance[T], error)
	Delete(ctx context.Context, dto sdk.ReadInstanceDTO) error
	Update(ctx context.Context, dto sdk.UpdateByIdDTO[sdk.ResourceInstance[T]]) (*sdk.ResourceInstance[T], error)
}

type ResourceInstanceRepository[T sdk.ResourceInstanceInterface] struct {
	repository ResourceInstanceRepositoryInterface[T]
}

// NewResourceInstanceRepository creates a new repository with default options
// Default behavior: AutoGenerateID = true (auto-generate ObjectID.Hex() as string)
func NewResourceInstanceRepository[T sdk.ResourceInstanceInterface](resourceId string, options ResourceInstanceRepositoryOptions) *ResourceInstanceRepository[T] {
	if options.AutoGenerateID == nil {
		def := true
		options.AutoGenerateID = &def
	}
	return &ResourceInstanceRepository[T]{
		repository: NewMongoResourceInstanceRepository[T](resourceId, options),
	}
}

func (r *ResourceInstanceRepository[T]) Instance(ctx context.Context, dto sdk.ReadInstanceDTO) (*sdk.ResourceInstance[T], error) {
	return r.repository.Instance(ctx, dto)
}

func (r *ResourceInstanceRepository[T]) List(ctx context.Context, dto sdk.ReadDTO) ([]sdk.ResourceInstance[T], error) {
	return r.repository.List(ctx, dto)
}

func (r *ResourceInstanceRepository[T]) Create(ctx context.Context, dto sdk.CreateDTO[sdk.ResourceInstance[T]]) (*sdk.ResourceInstance[T], error) {
	return r.repository.Create(ctx, dto)
}

func (r *ResourceInstanceRepository[T]) Delete(ctx context.Context, dto sdk.ReadInstanceDTO) error {
	return r.repository.Delete(ctx, dto)
}

func (r *ResourceInstanceRepository[T]) Update(ctx context.Context, dto sdk.UpdateByIdDTO[sdk.ResourceInstance[T]]) (*sdk.ResourceInstance[T], error) {
	return r.repository.Update(ctx, dto)
}

// ResourceInstanceSpecializedRepositoryInterface defines interface for specialized repositories
type ResourceInstanceSpecializedRepositoryInterface[T sdk.ResourceInstanceInterface, C sdk.ResourceInstanceSpecializedInterface] interface {
	Instance(ctx context.Context, dto sdk.ReadInstanceDTO) (*sdk.ResourceInstanceSpecialized[T, C], error)
	List(ctx context.Context, dto sdk.ReadDTO) ([]sdk.ResourceInstanceSpecialized[T, C], error)
	Create(ctx context.Context, dto sdk.CreateDTO[sdk.ResourceInstanceSpecialized[T, C]]) (*sdk.ResourceInstanceSpecialized[T, C], error)
	Delete(ctx context.Context, dto sdk.ReadInstanceDTO) error
	Update(ctx context.Context, dto sdk.UpdateByIdDTO[sdk.ResourceInstanceSpecialized[T, C]]) (*sdk.ResourceInstanceSpecialized[T, C], error)
}

// ResourceInstanceSpecializedRepository handles specialized resource operations independently
type ResourceInstanceSpecializedRepository[T sdk.ResourceInstanceInterface, C sdk.ResourceInstanceSpecializedInterface] struct {
	repository ResourceInstanceSpecializedRepositoryInterface[T, C]
}

// NewResourceInstanceSpecializedRepository creates a new specialized repository
func NewResourceInstanceSpecializedRepository[T sdk.ResourceInstanceInterface, C sdk.ResourceInstanceSpecializedInterface](
	resourceId string,
	options ResourceInstanceRepositoryOptions,
) *ResourceInstanceSpecializedRepository[T, C] {
	if options.AutoGenerateID == nil {
		def := true
		options.AutoGenerateID = &def
	}
	return &ResourceInstanceSpecializedRepository[T, C]{
		repository: NewMongoResourceInstanceSpecializedRepository[T, C](resourceId, options),
	}
}

// Instance retrieves a specialized resource instance
func (r *ResourceInstanceSpecializedRepository[T, C]) Instance(ctx context.Context, dto sdk.ReadInstanceDTO) (*sdk.ResourceInstanceSpecialized[T, C], error) {
	return r.repository.Instance(ctx, dto)
}

// List retrieves a list of specialized resource instances
func (r *ResourceInstanceSpecializedRepository[T, C]) List(ctx context.Context, dto sdk.ReadDTO) ([]sdk.ResourceInstanceSpecialized[T, C], error) {
	return r.repository.List(ctx, dto)
}

// Create creates a new specialized resource instance
func (r *ResourceInstanceSpecializedRepository[T, C]) Create(ctx context.Context, dto sdk.CreateDTO[sdk.ResourceInstanceSpecialized[T, C]]) (*sdk.ResourceInstanceSpecialized[T, C], error) {
	return r.repository.Create(ctx, dto)
}

// Delete deletes a specialized resource instance
func (r *ResourceInstanceSpecializedRepository[T, C]) Delete(ctx context.Context, dto sdk.ReadInstanceDTO) error {
	return r.repository.Delete(ctx, dto)
}

// Update updates a specialized resource instance
func (r *ResourceInstanceSpecializedRepository[T, C]) Update(ctx context.Context, dto sdk.UpdateByIdDTO[sdk.ResourceInstanceSpecialized[T, C]]) (*sdk.ResourceInstanceSpecialized[T, C], error) {
	return r.repository.Update(ctx, dto)
}
