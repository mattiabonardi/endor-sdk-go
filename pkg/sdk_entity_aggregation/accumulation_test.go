package sdk_entity_aggregation

import "testing"

func TestAccumulator_Sum(t *testing.T) {
	val, err := applyAccumulator(orderDocs, map[string]interface{}{"$sum": "$amount"})
	if err != nil {
		t.Fatal(err)
	}
	if val.(float64) != 800 {
		t.Errorf("got %v, want 800", val)
	}
}

func TestAccumulator_Avg(t *testing.T) {
	val, err := applyAccumulator(orderDocs, map[string]interface{}{"$avg": "$amount"})
	if err != nil {
		t.Fatal(err)
	}
	if val.(float64) != 160 {
		t.Errorf("got %v, want 160", val)
	}
}

func TestAccumulator_Min(t *testing.T) {
	val, err := applyAccumulator(orderDocs, map[string]interface{}{"$min": "$amount"})
	if err != nil {
		t.Fatal(err)
	}
	if val.(float64) != 50 {
		t.Errorf("got %v, want 50", val)
	}
}

func TestAccumulator_Max(t *testing.T) {
	val, err := applyAccumulator(orderDocs, map[string]interface{}{"$max": "$amount"})
	if err != nil {
		t.Fatal(err)
	}
	if val.(float64) != 300 {
		t.Errorf("got %v, want 300", val)
	}
}

func TestAccumulator_Count(t *testing.T) {
	val, err := applyAccumulator(orderDocs, map[string]interface{}{"$count": 1})
	if err != nil {
		t.Fatal(err)
	}
	if val.(int) != 5 {
		t.Errorf("got %v, want 5", val)
	}
}

func TestAccumulator_First(t *testing.T) {
	val, err := applyAccumulator(orderDocs, map[string]interface{}{"$first": "$amount"})
	if err != nil {
		t.Fatal(err)
	}
	if val.(float64) != 100 {
		t.Errorf("got %v, want 100", val)
	}
}

func TestAccumulator_Last(t *testing.T) {
	val, err := applyAccumulator(orderDocs, map[string]interface{}{"$last": "$amount"})
	if err != nil {
		t.Fatal(err)
	}
	if val.(float64) != 300 {
		t.Errorf("got %v, want 300", val)
	}
}
