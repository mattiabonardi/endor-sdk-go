package sdk_entity_aggregation

import (
	"testing"
)

// TestMergeResults_AllFields merges two entity result sets on "id" without
// restricting fields — every field from every doc must appear in the output.
func TestMergeResults_AllFields(t *testing.T) {
	entityResults := map[string][]map[string]interface{}{
		"customer": customerDocs,
		"order": {
			{"id": "c1", "amount": float64(300)},
			{"id": "c2", "amount": float64(200)},
			{"id": "c3", "amount": float64(300)},
		},
	}

	result := mergeResults(
		entityResults,
		[]string{"customer", "order"},
		MergeResultsOptions{On: "id"},
	)

	if len(result) != 3 {
		t.Fatalf("expected 3 docs, got %d", len(result))
	}

	byID := indexByStringField(result, "id")

	if got := byID["c1"]["name"]; got != "Alice" {
		t.Errorf("c1 name: got %v, want Alice", got)
	}
	if got := byID["c1"]["amount"]; got != float64(300) {
		t.Errorf("c1 amount: got %v, want 300", got)
	}
	if got := byID["c2"]["country"]; got != "US" {
		t.Errorf("c2 country: got %v, want US", got)
	}
}

// TestMergeResults_SelectedFields merges on "id" keeping only "name" and "amount".
func TestMergeResults_SelectedFields(t *testing.T) {
	entityResults := map[string][]map[string]interface{}{
		"customer": customerDocs,
		"order": {
			{"id": "c1", "amount": float64(300)},
			{"id": "c2", "amount": float64(200)},
		},
	}

	result := mergeResults(
		entityResults,
		[]string{"customer", "order"},
		MergeResultsOptions{On: "id", Fields: []string{"name", "amount"}},
	)

	if len(result) != 3 {
		t.Fatalf("expected 3 docs, got %d", len(result))
	}

	byID := indexByStringField(result, "id")

	// "country" must not appear because it is not in Fields.
	if _, ok := byID["c1"]["country"]; ok {
		t.Errorf("c1 country should not be present")
	}
	if got := byID["c1"]["name"]; got != "Alice" {
		t.Errorf("c1 name: got %v, want Alice", got)
	}
	if got := byID["c1"]["amount"]; got != float64(300) {
		t.Errorf("c1 amount: got %v, want 300", got)
	}
	// c3 has no order doc — amount must be absent.
	if _, ok := byID["c3"]["amount"]; ok {
		t.Errorf("c3 amount should not be present")
	}
}

// TestMergeResults_OrderPreserved verifies that the output doc order follows
// the first-seen order from the leading entity in order slice.
func TestMergeResults_OrderPreserved(t *testing.T) {
	entityResults := map[string][]map[string]interface{}{
		"customer": {
			{"id": "c3", "name": "Carol"},
			{"id": "c1", "name": "Alice"},
			{"id": "c2", "name": "Bob"},
		},
	}

	result := mergeResults(
		entityResults,
		[]string{"customer"},
		MergeResultsOptions{On: "id"},
	)

	if len(result) != 3 {
		t.Fatalf("expected 3 docs, got %d", len(result))
	}

	want := []string{"c3", "c1", "c2"}
	for i, doc := range result {
		if got := doc["id"].(string); got != want[i] {
			t.Errorf("position %d: got %s, want %s", i, got, want[i])
		}
	}
}

// TestMergeResults_MissingEntity skips entity keys absent from entityResults.
func TestMergeResults_MissingEntity(t *testing.T) {
	entityResults := map[string][]map[string]interface{}{
		"customer": customerDocs,
	}

	result := mergeResults(
		entityResults,
		[]string{"customer", "nonexistent"},
		MergeResultsOptions{On: "id"},
	)

	if len(result) != 3 {
		t.Fatalf("expected 3 docs, got %d", len(result))
	}
}

// TestMergeResults_EmptyEntityResults returns an empty slice when no entity
// in order has results.
func TestMergeResults_EmptyEntityResults(t *testing.T) {
	result := mergeResults(
		map[string][]map[string]interface{}{},
		[]string{"customer"},
		MergeResultsOptions{On: "id"},
	)

	if len(result) != 0 {
		t.Fatalf("expected 0 docs, got %d", len(result))
	}
}

// indexByStringField indexes docs by a given string field value.
func indexByStringField(docs []map[string]interface{}, field string) map[string]map[string]interface{} {
	m := make(map[string]map[string]interface{}, len(docs))
	for _, doc := range docs {
		if v, ok := doc[field]; ok {
			m[v.(string)] = doc
		}
	}
	return m
}
