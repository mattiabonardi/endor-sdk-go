package sdk_entity

import (
	"context"
	"strings"

	"github.com/mattiabonardi/endor-sdk-go/internal/repository"
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
	"go.mongodb.org/mongo-driver/bson"
)

type EntityInstanceRepository[T sdk.EntityInstanceInterface] struct {
	repository sdk.EntityInstanceRepositoryInterface[T]
	entityId   string
	schema     sdk.RootSchema
}

// NewEntityInstanceRepository creates a new repository with default options
// Default behavior: AutoGenerateID = true (auto-generate ObjectID.Hex() as string)
func NewEntityInstanceRepository[T sdk.EntityInstanceInterface](entityId string, schema sdk.RootSchema, options sdk.EntityInstanceRepositoryOptions, session sdk.Session, di sdk.EndorDIContainerInterface) *EntityInstanceRepository[T] {
	if options.AutoGenerateID == nil {
		def := true
		options.AutoGenerateID = &def
	}
	entity, _, _ := strings.Cut(entityId, "/")

	return &EntityInstanceRepository[T]{
		repository: repository.NewMongoEntityInstanceRepository[T](entity, schema, options, session, di),
		entityId:   entityId,
		schema:     schema,
	}
}

func (r *EntityInstanceRepository[T]) Instance(ctx context.Context, dto sdk.ReadInstanceDTO) (*sdk.EntityInstance[T], error) {
	return r.repository.Instance(ctx, dto)
}

func (r *EntityInstanceRepository[T]) RawList(ctx context.Context, dto sdk.ReadDTO) ([]bson.M, error) {
	return r.repository.RawList(ctx, dto)
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

func (r *EntityInstanceRepository[T]) GetSchema() *sdk.RootSchema {
	return &r.schema
}

func (r *EntityInstanceRepository[T]) InstanceWithReferences(ctx context.Context, dto sdk.ReadInstanceDTO) (*sdk.EntityInstance[T], sdk.EntityRefererenceGroup, error) {
	return r.repository.InstanceWithReferences(ctx, dto)
}

func (r *EntityInstanceRepository[T]) ListWithReferences(ctx context.Context, dto sdk.ReadDTO) ([]sdk.EntityInstance[T], sdk.EntityRefererenceGroup, error) {
	return r.repository.ListWithReferences(ctx, dto)
}
