package sdk_entity_aggregation

import "testing"

// ---------------------------------------------------------------------------
// applyMatch
// ---------------------------------------------------------------------------

func TestApplyMatch_EqualityDirect(t *testing.T) {
	result := applyMatch(orderDocs, map[string]interface{}{"status": "completed"})
	if len(result) != 4 {
		t.Errorf("expected 4, got %d", len(result))
	}
}

func TestApplyMatch_NoMatch(t *testing.T) {
	result := applyMatch(orderDocs, map[string]interface{}{"status": "cancelled"})
	if len(result) != 0 {
		t.Errorf("expected 0, got %d", len(result))
	}
}

func TestApplyMatch_EmptyFilter(t *testing.T) {
	result := applyMatch(orderDocs, map[string]interface{}{})
	if len(result) != len(orderDocs) {
		t.Errorf("expected %d, got %d", len(orderDocs), len(result))
	}
}

func TestApplyMatch_And(t *testing.T) {
	result := applyMatch(orderDocs, map[string]interface{}{
		"$and": []interface{}{
			map[string]interface{}{"status": "completed"},
			map[string]interface{}{"customerId": "c1"},
		},
	})
	if len(result) != 2 {
		t.Errorf("expected 2, got %d", len(result))
	}
}

func TestApplyMatch_Or(t *testing.T) {
	result := applyMatch(orderDocs, map[string]interface{}{
		"$or": []interface{}{
			map[string]interface{}{"customerId": "c1"},
			map[string]interface{}{"customerId": "c3"},
		},
	})
	if len(result) != 3 {
		t.Errorf("expected 3, got %d", len(result))
	}
}

func TestApplyMatch_Nor(t *testing.T) {
	// excludes c1 and c3 → only c2 docs remain
	result := applyMatch(orderDocs, map[string]interface{}{
		"$nor": []interface{}{
			map[string]interface{}{"customerId": "c1"},
			map[string]interface{}{"customerId": "c3"},
		},
	})
	if len(result) != 2 {
		t.Errorf("expected 2, got %d", len(result))
	}
}

func TestApplyMatch_MultipleFields(t *testing.T) {
	result := applyMatch(orderDocs, map[string]interface{}{
		"status":     "pending",
		"customerId": "c2",
	})
	if len(result) != 1 {
		t.Errorf("expected 1, got %d", len(result))
	}
}

// ---------------------------------------------------------------------------
// matchCondition — comparison operators
// ---------------------------------------------------------------------------

func TestMatchCondition_Eq(t *testing.T) {
	if !matchCondition(float64(100), map[string]interface{}{"$eq": float64(100)}) {
		t.Error("$eq: expected true")
	}
	if matchCondition(float64(100), map[string]interface{}{"$eq": float64(200)}) {
		t.Error("$eq: expected false")
	}
}

func TestMatchCondition_Ne(t *testing.T) {
	if !matchCondition(float64(100), map[string]interface{}{"$ne": float64(200)}) {
		t.Error("$ne: expected true")
	}
	if matchCondition(float64(100), map[string]interface{}{"$ne": float64(100)}) {
		t.Error("$ne: expected false")
	}
}

func TestMatchCondition_Gt(t *testing.T) {
	if !matchCondition(float64(200), map[string]interface{}{"$gt": float64(100)}) {
		t.Error("$gt: expected true")
	}
	if matchCondition(float64(100), map[string]interface{}{"$gt": float64(100)}) {
		t.Error("$gt equal: expected false")
	}
	if matchCondition(float64(50), map[string]interface{}{"$gt": float64(100)}) {
		t.Error("$gt less: expected false")
	}
}

func TestMatchCondition_Gte(t *testing.T) {
	if !matchCondition(float64(100), map[string]interface{}{"$gte": float64(100)}) {
		t.Error("$gte equal: expected true")
	}
	if !matchCondition(float64(200), map[string]interface{}{"$gte": float64(100)}) {
		t.Error("$gte greater: expected true")
	}
	if matchCondition(float64(50), map[string]interface{}{"$gte": float64(100)}) {
		t.Error("$gte less: expected false")
	}
}

func TestMatchCondition_Lt(t *testing.T) {
	if !matchCondition(float64(50), map[string]interface{}{"$lt": float64(100)}) {
		t.Error("$lt: expected true")
	}
	if matchCondition(float64(100), map[string]interface{}{"$lt": float64(100)}) {
		t.Error("$lt equal: expected false")
	}
}

func TestMatchCondition_Lte(t *testing.T) {
	if !matchCondition(float64(100), map[string]interface{}{"$lte": float64(100)}) {
		t.Error("$lte equal: expected true")
	}
	if matchCondition(float64(200), map[string]interface{}{"$lte": float64(100)}) {
		t.Error("$lte greater: expected false")
	}
}

func TestMatchCondition_In(t *testing.T) {
	if !matchCondition("completed", map[string]interface{}{"$in": []interface{}{"completed", "pending"}}) {
		t.Error("$in: expected true")
	}
	if matchCondition("cancelled", map[string]interface{}{"$in": []interface{}{"completed", "pending"}}) {
		t.Error("$in: expected false")
	}
}

func TestMatchCondition_Nin(t *testing.T) {
	if !matchCondition("cancelled", map[string]interface{}{"$nin": []interface{}{"completed", "pending"}}) {
		t.Error("$nin: expected true")
	}
	if matchCondition("completed", map[string]interface{}{"$nin": []interface{}{"completed", "pending"}}) {
		t.Error("$nin: expected false")
	}
}

func TestMatchCondition_Exists(t *testing.T) {
	if !matchCondition("Alice", map[string]interface{}{"$exists": true}) {
		t.Error("$exists true: expected true for non-nil value")
	}
	if matchCondition(nil, map[string]interface{}{"$exists": true}) {
		t.Error("$exists true: expected false for nil")
	}
	if !matchCondition(nil, map[string]interface{}{"$exists": false}) {
		t.Error("$exists false: expected true for nil")
	}
}

func TestMatchCondition_Regex(t *testing.T) {
	if !matchCondition("completed", map[string]interface{}{"$regex": "mplet"}) {
		t.Error("$regex: expected substring match")
	}
	if matchCondition("pending", map[string]interface{}{"$regex": "mplet"}) {
		t.Error("$regex: expected no match")
	}
}

func TestMatchCondition_PlainEquality(t *testing.T) {
	if !matchCondition("completed", "completed") {
		t.Error("plain equality: expected true")
	}
	if matchCondition("completed", "pending") {
		t.Error("plain equality: expected false")
	}
}

// ---------------------------------------------------------------------------
// applyMatch integration — operators against orderDocs
// ---------------------------------------------------------------------------

func TestApplyMatch_Gte_OnOrders(t *testing.T) {
	result := applyMatch(orderDocs, map[string]interface{}{
		"amount": map[string]interface{}{"$gte": float64(150)},
	})
	// amount >= 150: 150, 200, 300 → 3 docs
	if len(result) != 3 {
		t.Errorf("expected 3, got %d", len(result))
	}
}

func TestApplyMatch_In_OnOrders(t *testing.T) {
	result := applyMatch(orderDocs, map[string]interface{}{
		"customerId": map[string]interface{}{"$in": []interface{}{"c1", "c3"}},
	})
	// c1 has 2 docs, c3 has 1 → 3 total
	if len(result) != 3 {
		t.Errorf("expected 3, got %d", len(result))
	}
}

func TestApplyMatch_Regex_OnCustomers(t *testing.T) {
	result := applyMatch(customerDocs, map[string]interface{}{
		"name": map[string]interface{}{"$regex": "li"},
	})
	// "Alice" contains "li" → 1 doc
	if len(result) != 1 {
		t.Errorf("expected 1, got %d", len(result))
	}
	if result[0]["name"] != "Alice" {
		t.Errorf("expected Alice, got %v", result[0]["name"])
	}
}
