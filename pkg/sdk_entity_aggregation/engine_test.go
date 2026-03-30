package sdk_entity_aggregation

import (
	"context"
	"encoding/json"
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

	p := pipeline(t,
		map[string]interface{}{
			"entity": "order",
			"pipeline": []interface{}{
				map[string]interface{}{
					"$group": map[string]interface{}{
						"id": "$customerId",
					},
				},
			},
		},
	)

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

	p := pipeline(t,
		map[string]interface{}{
			"entity": "order",
			"pipeline": []interface{}{
				map[string]interface{}{
					"$group": map[string]interface{}{
						"id":    "$customerId",
						"total": map[string]interface{}{"$sum": "$amount"},
					},
				},
			},
		},
	)

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
	p := pipeline(t,
		map[string]interface{}{
			"entity": "order",
			"pipeline": []interface{}{
				map[string]interface{}{
					"$group": map[string]interface{}{
						"id":    "$customerId",
						"total": map[string]interface{}{"$sum": "$amount"},
					},
				},
			},
		},
		map[string]interface{}{
			"entity":   "customer",
			"pipeline": []interface{}{},
		},
		map[string]interface{}{
			"$mergeResults": map[string]interface{}{
				"on": "id",
			},
		},
	)

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

// #region helpers

func pipeline(t *testing.T, stages ...interface{}) AggregationPipeline {
	t.Helper()
	p := make(AggregationPipeline, len(stages))
	for i, s := range stages {
		b, err := json.Marshal(s)
		if err != nil {
			t.Fatalf("pipeline marshal stage %d: %v", i, err)
		}
		p[i] = json.RawMessage(b)
	}
	return p
}

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
