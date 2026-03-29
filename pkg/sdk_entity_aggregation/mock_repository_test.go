package sdk_entity_aggregation

import (
	"context"

	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
)

// mockRepository implements sdk.DocumentRepositoryInterface with an in-memory
// document set. It is intended for use in unit tests only.
type mockRepository struct {
	entity string
	docs   []map[string]interface{}
}

func newMockRepository(entity string, docs []map[string]interface{}) *mockRepository {
	return &mockRepository{entity: entity, docs: docs}
}

func (r *mockRepository) GetEntity() string { return r.entity }

func (r *mockRepository) FindReferences(_ context.Context, _ sdk.ReadInstancesDTO) (sdk.EntityReferenceGroupDescriptions, error) {
	return sdk.EntityReferenceGroupDescriptions{}, nil
}

// ListDocuments returns documents, applying the filter from ReadDTO when present
// so that push-down $match optimizations work correctly in tests.
func (r *mockRepository) ListDocuments(_ context.Context, dto sdk.ReadDTO) ([]map[string]interface{}, error) {
	if len(dto.Filter) == 0 {
		return r.docs, nil
	}
	return applyMatch(r.docs, dto.Filter), nil
}

// registerMock registers the mock in the global RepositoryRegistry and returns
// a cleanup function that removes it afterwards.
func registerMock(repo *mockRepository) func() {
	sdk.GetRepositoryRegistry().Register(repo.entity, repo)
	return func() {
		sdk.GetRepositoryRegistry().Register(repo.entity, nil)
	}
}
