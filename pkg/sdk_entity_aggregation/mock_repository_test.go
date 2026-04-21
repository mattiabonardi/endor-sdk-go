package sdk_entity_aggregation

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
)

// testDI is the shared DI container used by all aggregation tests.
var testDI = &mockDIContainer{repos: map[string]sdk.EndorRepositoryInterface{}}

// mockDIContainer implements sdk.EndorDIContainer for tests.
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

// RawList returns documents as []bson.M, applying the filter from ReadDTO when
// present so that push-down $match optimizations work correctly in tests.
func (r *mockRepository) RawList(_ context.Context, dto sdk.ReadDTO) ([]bson.M, error) {
	src := r.docs
	if len(dto.Filter) > 0 {
		src = applyMatch(r.docs, dto.Filter)
	}
	result := make([]bson.M, len(src))
	for i, d := range src {
		result[i] = bson.M(d)
	}
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
