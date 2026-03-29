package sdk_entity

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mattiabonardi/endor-sdk-go/internal/repository"
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
)

type EntityInstanceRepository[T sdk.EntityInstanceInterface] struct {
	repository sdk.EntityInstanceRepositoryInterface[T]
	entityId   string
}

// NewEntityInstanceRepository creates a new repository with default options
// Default behavior: AutoGenerateID = true (auto-generate ObjectID.Hex() as string)
func NewEntityInstanceRepository[T sdk.EntityInstanceInterface](entityId string, schema sdk.RootSchema, options sdk.EntityInstanceRepositoryOptions) *EntityInstanceRepository[T] {
	if options.AutoGenerateID == nil {
		def := true
		options.AutoGenerateID = &def
	}
	return &EntityInstanceRepository[T]{
		repository: repository.NewMongoEntityInstanceRepository[T](entityId, schema, options),
		entityId:   entityId,
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

func (r *EntityInstanceRepository[T]) Update(ctx context.Context, dto sdk.UpdateByIdDTO[sdk.PartialEntityInstance[T]]) (*sdk.EntityInstance[T], error) {
	return r.repository.Update(ctx, dto)
}

func (r *EntityInstanceRepository[T]) FindReferences(ctx context.Context, dto sdk.ReadInstancesDTO) (sdk.EntityReferenceGroupDescriptions, error) {
	return r.repository.FindReferences(ctx, dto)
}

func (r *EntityInstanceRepository[T]) GetEntity() string {
	return r.entityId
}

func (r *EntityInstanceRepository[T]) InstanceWithReferences(ctx context.Context, dto sdk.ReadInstanceDTO) (*sdk.EntityInstance[T], sdk.EntityRefererenceGroup, error) {
	return r.repository.InstanceWithReferences(ctx, dto)
}

func (r *EntityInstanceRepository[T]) ListWithReferences(ctx context.Context, dto sdk.ReadDTO) ([]sdk.EntityInstance[T], sdk.EntityRefererenceGroup, error) {
	return r.repository.ListWithReferences(ctx, dto)
}

func (r *EntityInstanceRepository[T]) ListDocuments(ctx context.Context, dto sdk.ReadDTO) ([]map[string]interface{}, error) {
	items, err := r.repository.List(ctx, dto)
	if err != nil {
		return nil, err
	}
	result := make([]map[string]interface{}, 0, len(items))
	for _, item := range items {
		data, err := json.Marshal(item)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal entity instance: %w", err)
		}
		var doc map[string]interface{}
		if err := json.Unmarshal(data, &doc); err != nil {
			return nil, fmt.Errorf("failed to unmarshal entity instance: %w", err)
		}
		result = append(result, doc)
	}
	return result, nil
}
