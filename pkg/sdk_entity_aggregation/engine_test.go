package sdk_entity_aggregation

import (
	"context"
	"encoding/json"
	"testing"
)

// orderDocs is the shared test dataset used by group-by tests.
var orderDocs = []map[string]interface{}{
	{"_id": "1", "customerId": "c1", "status": "completed", "amount": float64(100)},
	{"_id": "2", "customerId": "c1", "status": "completed", "amount": float64(200)},
	{"_id": "3", "customerId": "c2", "status": "completed", "amount": float64(150)},
	{"_id": "4", "customerId": "c2", "status": "pending", "amount": float64(50)},
	{"_id": "5", "customerId": "c3", "status": "completed", "amount": float64(300)},
}

func pipeline(t *testing.T, stages ...interface{}) AggregationPipeline {
	t.Helper()
	p := make(AggregationPipeline, len(stages))
	for i, s := range stages {
		b, err := json.Marshal(s)
		if err != nil {
			t.Fatalf("pipeline marshal stage %d: %v", err, i)
		}
		p[i] = json.RawMessage(b)
	}
	return p
}

func TestGroupBy_ByCustomer(t *testing.T) {
	cleanup := registerMock(newMockRepository("order", orderDocs))
	defer cleanup()

	p := pipeline(t,
		map[string]interface{}{
			"entity": "order",
			"pipeline": []interface{}{
				map[string]interface{}{
					"$group": map[string]interface{}{
						"_id": "$customerId",
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

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func indexByID(docs []map[string]interface{}) map[string]map[string]interface{} {
	m := make(map[string]map[string]interface{}, len(docs))
	for _, doc := range docs {
		key := ""
		if id := doc["_id"]; id != nil {
			key = id.(string)
		}
		m[key] = doc
	}
	return m
}
