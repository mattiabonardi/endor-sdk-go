package sdk_entity_aggregation

import (
	"context"
	"fmt"
	"testing"
)

// orderDocs is the shared test dataset used by aggregation tests.
var orderDocs = []map[string]interface{}{
	{"id": "1", "customerId": "c1", "status": "completed", "amount": float64(100)},
	{"id": "2", "customerId": "c1", "status": "completed", "amount": float64(200)},
	{"id": "3", "customerId": "c2", "status": "completed", "amount": float64(150)},
	{"id": "4", "customerId": "c2", "status": "pending", "amount": float64(50)},
	{"id": "5", "customerId": "c3", "status": "completed", "amount": float64(300)},
}

// entityResults built from the shared orderDocs dataset plus extra customer docs.
var customerDocs = []map[string]interface{}{
	{"id": "c1", "name": "Alice", "country": "IT"},
	{"id": "c2", "name": "Bob", "country": "US"},
	{"id": "c3", "name": "Carol", "country": "FR"},
}

// #region selection

func TestGroupBy_ByCustomer(t *testing.T) {
	cleanup := registerMock(newMockRepository("order", orderDocs))
	defer cleanup()

	p := AggregationPipeline{
		{
			Entity: "order",
			Pipeline: []StageSpec{
				{"$group": map[string]interface{}{"id": "$customerId"}},
			},
		},
	}

	result, err := NewAggregationEngine().Execute(context.Background(), p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 3 {
		t.Fatalf("expected 3 groups, got %d", len(result))
	}

	byID := indexByID(result)

	if _, ok := byID["c1"]; !ok {
		t.Errorf("expected group c1")
	}
	if _, ok := byID["c2"]; !ok {
		t.Errorf("expected group c2")
	}
	if _, ok := byID["c3"]; !ok {
		t.Errorf("expected group c3")
	}
}

// #endregion

// #region accumulation

func TestGroupBy_ByCustomer_WithSum(t *testing.T) {
	cleanup := registerMock(newMockRepository("order", orderDocs))
	defer cleanup()

	p := AggregationPipeline{
		{
			Entity: "order",
			Pipeline: []StageSpec{
				{"$group": map[string]interface{}{
					"id":    "$customerId",
					"total": map[string]interface{}{"$sum": "$amount"},
				}},
			},
		},
	}

	result, err := NewAggregationEngine().Execute(context.Background(), p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 3 {
		t.Fatalf("expected 3 groups, got %d", len(result))
	}

	byID := indexByID(result)

	if got := byID["c1"]["total"].(float64); got != 300 {
		t.Errorf("c1 total: got %v, want 300", got)
	}
	if got := byID["c2"]["total"].(float64); got != 200 {
		t.Errorf("c2 total: got %v, want 200", got)
	}
	if got := byID["c3"]["total"].(float64); got != 300 {
		t.Errorf("c3 total: got %v, want 300", got)
	}
}

// #endregion

// #region combination

func TestMergeResults(t *testing.T) {
	cleanupOrders := registerMock(newMockRepository("order", orderDocs))
	defer cleanupOrders()
	cleanupCustomers := registerMock(newMockRepository("customer", customerDocs))
	defer cleanupCustomers()

	// Group orders by customerId → each doc gets "id" = customerId.
	// Then merge with customer docs (which also carry "id") to get a combined view.
	p := AggregationPipeline{
		{
			ID:     "grouped_orders",
			Entity: "order",
			Pipeline: []StageSpec{
				{"$group": map[string]interface{}{
					"id":    "$customerId",
					"total": map[string]interface{}{"$sum": "$amount"},
				}},
			},
		},
		{
			ID:       "customers",
			Entity:   "customer",
			Pipeline: []StageSpec{},
		},
		{
			DependsOn: []string{"grouped_orders", "customers"},
			Pipeline: []StageSpec{
				{"$mergeResults": map[string]interface{}{"on": "id"}},
			},
		},
	}

	result, err := NewAggregationEngine().Execute(context.Background(), p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 3 {
		t.Fatalf("expected 3 merged docs, got %d", len(result))
	}

	byID := indexByID(result)

	if got := byID["c1"]["total"].(float64); got != 300 {
		t.Errorf("c1 total: got %v, want 300", got)
	}
	if got := byID["c1"]["name"].(string); got != "Alice" {
		t.Errorf("c1 name: got %v, want Alice", got)
	}
	if got := byID["c2"]["total"].(float64); got != 200 {
		t.Errorf("c2 total: got %v, want 200", got)
	}
	if got := byID["c2"]["country"].(string); got != "US" {
		t.Errorf("c2 country: got %v, want US", got)
	}
	if got := byID["c3"]["name"].(string); got != "Carol" {
		t.Errorf("c3 name: got %v, want Carol", got)
	}
}

// #endregion

// #region entity_stage_handler

// TestEntityStageHandler_ReplacesBuiltinLogic verifies that, when an
// EntityStageHandler is provided via WithEntityStageHandler, the engine calls
// the callback for every entity stage and uses its returned docs instead of
// hitting the repository registry. The callback here computes a result from the
// stage metadata so the test is self-contained and does not need a mock repo.
func TestEntityStageHandler_ReplacesBuiltinLogic(t *testing.T) {
	// computedByEntity simulates what a master query layer would do: it returns
	// a synthetic document set whose content depends on the entity name.
	computedByEntity := map[string][]map[string]interface{}{
		"order": {
			{"id": "c1", "total": float64(300)},
			{"id": "c2", "total": float64(200)},
			{"id": "c3", "total": float64(300)},
		},
		"customer": {
			{"id": "c1", "name": "Alice"},
			{"id": "c2", "name": "Bob"},
			{"id": "c3", "name": "Carol"},
		},
	}

	handler := func(_ context.Context, stage EntityPipelineStage) ([]map[string]interface{}, error) {
		docs, ok := computedByEntity[stage.Entity]
		if !ok {
			return []map[string]interface{}{}, nil
		}
		return docs, nil
	}

	p := AggregationPipeline{
		{ID: "orders", Entity: "order", Pipeline: []StageSpec{}},
		{ID: "customers", Entity: "customer", Pipeline: []StageSpec{}},
		{
			DependsOn: []string{"orders", "customers"},
			Pipeline:  []StageSpec{{"$mergeResults": map[string]interface{}{"on": "id"}}},
		},
	}

	engine := NewAggregationEngine(WithEntityStageHandler(handler))
	result, err := engine.Execute(context.Background(), p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 3 {
		t.Fatalf("expected 3 merged docs, got %d", len(result))
	}

	byID := indexByID(result)

	if got := byID["c1"]["total"].(float64); got != 300 {
		t.Errorf("c1 total: got %v, want 300", got)
	}
	if got := byID["c1"]["name"].(string); got != "Alice" {
		t.Errorf("c1 name: got %v, want Alice", got)
	}
	if got := byID["c2"]["total"].(float64); got != 200 {
		t.Errorf("c2 total: got %v, want 200", got)
	}
	if got := byID["c2"]["name"].(string); got != "Bob" {
		t.Errorf("c2 name: got %v, want Bob", got)
	}
	if got := byID["c3"]["name"].(string); got != "Carol" {
		t.Errorf("c3 name: got %v, want Carol", got)
	}
}

// TestEntityStageHandler_OwnsFullStage verifies that the EntityStageHandler
// receives the complete EntityPipelineStage — including its Pipeline — and that
// its return value is used as-is, without the engine re-applying any in-memory
// operators. This models a master microservice that forwards the entire stage
// (entity + pipeline) to a child microservice via HTTP and returns its result.
func TestEntityStageHandler_OwnsFullStage(t *testing.T) {
	// The handler simulates a child microservice that has already executed the
	// $group+$sum locally and returns the aggregated result directly.
	handler := func(_ context.Context, stage EntityPipelineStage) ([]map[string]interface{}, error) {
		// Assert that the full pipeline is forwarded to the handler.
		if len(stage.Pipeline) != 1 {
			return nil, fmt.Errorf("expected 1 pipeline stage forwarded, got %d", len(stage.Pipeline))
		}
		// Return the already-aggregated result (as a child microservice would).
		return []map[string]interface{}{
			{"id": "c1", "total": float64(300)},
			{"id": "c2", "total": float64(150)},
		}, nil
	}

	p := AggregationPipeline{
		{
			Entity: "order",
			Pipeline: []StageSpec{
				{"$group": map[string]interface{}{
					"id":    "$customerId",
					"total": map[string]interface{}{"$sum": "$amount"},
				}},
			},
		},
	}

	engine := NewAggregationEngine(WithEntityStageHandler(handler))
	result, err := engine.Execute(context.Background(), p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("expected 2 docs, got %d", len(result))
	}

	byID := indexByID(result)

	if got := byID["c1"]["total"].(float64); got != 300 {
		t.Errorf("c1 total: got %v, want 300", got)
	}
	if got := byID["c2"]["total"].(float64); got != 150 {
		t.Errorf("c2 total: got %v, want 150", got)
	}
}

// #endregion

// indexByID indexes grouped results by the "id" field produced by $group.
func indexByID(docs []map[string]interface{}) map[string]map[string]interface{} {
	m := make(map[string]map[string]interface{}, len(docs))
	for _, doc := range docs {
		key := ""
		if id := doc["id"]; id != nil {
			key = id.(string)
		}
		m[key] = doc
	}
	return m
}

// #endregion
