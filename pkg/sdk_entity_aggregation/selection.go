package sdk_entity_aggregation

import (
	"fmt"
	"strings"
)

// applyMatch filters documents using MongoDB-style filter operators.
func applyMatch(docs []map[string]interface{}, filter map[string]interface{}) []map[string]interface{} {
	result := docs[:0:0]
	for _, doc := range docs {
		if matchDocument(doc, filter) {
			result = append(result, doc)
		}
	}
	return result
}

func matchDocument(doc map[string]interface{}, filter map[string]interface{}) bool {
	for field, condition := range filter {
		switch field {
		case "$and":
			clauses, ok := toSliceOfMaps(condition)
			if !ok {
				return false
			}
			for _, clause := range clauses {
				if !matchDocument(doc, clause) {
					return false
				}
			}
		case "$or":
			clauses, ok := toSliceOfMaps(condition)
			if !ok {
				return false
			}
			matched := false
			for _, clause := range clauses {
				if matchDocument(doc, clause) {
					matched = true
					break
				}
			}
			if !matched {
				return false
			}
		case "$nor":
			clauses, ok := toSliceOfMaps(condition)
			if !ok {
				return false
			}
			for _, clause := range clauses {
				if matchDocument(doc, clause) {
					return false
				}
			}
		default:
			docVal := getFieldValue(doc, field)
			if !matchCondition(docVal, condition) {
				return false
			}
		}
	}
	return true
}

func matchCondition(value interface{}, condition interface{}) bool {
	condMap, ok := condition.(map[string]interface{})
	if !ok {
		// Plain equality.
		return equals(value, condition)
	}
	for op, operand := range condMap {
		switch op {
		case "$eq":
			if !equals(value, operand) {
				return false
			}
		case "$ne":
			if equals(value, operand) {
				return false
			}
		case "$gt":
			if compareValues(value, operand) <= 0 {
				return false
			}
		case "$gte":
			if compareValues(value, operand) < 0 {
				return false
			}
		case "$lt":
			if compareValues(value, operand) >= 0 {
				return false
			}
		case "$lte":
			if compareValues(value, operand) > 0 {
				return false
			}
		case "$in":
			arr := toSlice(operand)
			found := false
			for _, v := range arr {
				if equals(value, v) {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		case "$nin":
			arr := toSlice(operand)
			for _, v := range arr {
				if equals(value, v) {
					return false
				}
			}
		case "$exists":
			wantExists := toBool(operand)
			exists := value != nil
			if exists != wantExists {
				return false
			}
		case "$regex":
			// Basic prefix/contains match using strings.Contains for simplicity.
			if pattern, ok := operand.(string); ok {
				str := fmt.Sprintf("%v", value)
				if !strings.Contains(str, pattern) {
					return false
				}
			}
		}
	}
	return true
}

// applyGroup groups documents by the id expression and computes accumulators.
func applyGroup(docs []map[string]interface{}, groupSpec map[string]interface{}) ([]map[string]interface{}, error) {
	idExpr := groupSpec["id"]

	type groupEntry struct {
		key   string
		idVal interface{}
		docs  []map[string]interface{}
	}
	order := []string{}
	groups := map[string]*groupEntry{}

	for _, doc := range docs {
		idVal := resolveExpr(doc, idExpr)
		key := fmt.Sprintf("%v", idVal)
		if _, exists := groups[key]; !exists {
			groups[key] = &groupEntry{key: key, idVal: idVal}
			order = append(order, key)
		}
		groups[key].docs = append(groups[key].docs, doc)
	}

	result := make([]map[string]interface{}, 0, len(groups))
	for _, key := range order {
		entry := groups[key]
		output := map[string]interface{}{"id": entry.idVal}
		for field, accExpr := range groupSpec {
			if field == "id" {
				continue
			}
			val, err := applyAccumulator(entry.docs, accExpr)
			if err != nil {
				return nil, fmt.Errorf("accumulator %q: %w", field, err)
			}
			output[field] = val
		}
		result = append(result, output)
	}
	return result, nil
}
