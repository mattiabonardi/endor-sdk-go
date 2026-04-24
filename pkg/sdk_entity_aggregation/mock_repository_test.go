package sdk_entity_aggregation

import (
	"context"

	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
)

// testDI is the shared DI container used by all aggregation tests.
var testDI = &mockDIContainer{repos: map[string]sdk.EndorRepositoryInterface{}}

// mockDIContainer implements sdk.EndorDIContainerInterface for tests.
type mockDIContainer struct {
	repos map[string]sdk.EndorRepositoryInterface
}

func (m *mockDIContainer) GetRepositories() map[string]sdk.EndorRepositoryInterface {
	return m.repos
}

// mockRepository implements sdk.DocumentRepositoryInterface with an in-memory
// document set. It is intended for use in unit tests only.
type mockRepository struct {
	entity string
	docs   []map[string]interface{}
	schema *sdk.RootSchema
	// refDescs, when non-nil, is returned by FindReferences. Used by tests that
	// need a reference-holding "lookup" entity (e.g. "product").
	refDescs sdk.EntityReferenceGroupDescriptions
}

func newMockRepository(entity string, docs []map[string]interface{}) *mockRepository {
	return &mockRepository{entity: entity, docs: docs}
}

func (r *mockRepository) GetEntity() string { return r.entity }

func (r *mockRepository) GetSchema() *sdk.RootSchema { return r.schema }

func (r *mockRepository) FindReferences(_ context.Context, _ sdk.ReadInstancesDTO) (sdk.EntityReferenceGroupDescriptions, error) {
	if r.refDescs != nil {
		return r.refDescs, nil
	}
	return sdk.EntityReferenceGroupDescriptions{}, nil
}

// RawList returns documents as []map[string]interface{}, applying the filter from ReadDTO when
// present so that push-down $match optimizations work correctly in tests.
func (r *mockRepository) RawList(_ context.Context, dto sdk.ReadDTO) ([]map[string]interface{}, error) {
	src := r.docs
	if len(dto.Filter) > 0 {
		src = applyMatch(r.docs, dto.Filter)
	}
	result := make([]map[string]interface{}, len(src))
	copy(result, src)
	return result, nil
}

// registerMock registers the mock in the shared testDI container and returns
// a cleanup function that removes it afterwards.
func registerMock(repo *mockRepository) func() {
	testDI.repos[repo.entity] = repo
	return func() {
		delete(testDI.repos, repo.entity)
	}
}
