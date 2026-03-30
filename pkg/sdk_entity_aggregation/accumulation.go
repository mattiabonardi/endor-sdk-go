package sdk_entity_aggregation

import "fmt"

func applyAccumulator(docs []map[string]interface{}, accExpr interface{}) (interface{}, error) {
	accMap, ok := accExpr.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("accumulator must be an object, got %T", accExpr)
	}
	for op, fieldExpr := range accMap {
		switch op {
		case "$sum":
			sum := float64(0)
			for _, doc := range docs {
				val := resolveExpr(doc, fieldExpr)
				if f, ok := toFloat64(val); ok {
					sum += f
				}
			}
			return sum, nil
		case "$avg":
			sum := float64(0)
			count := 0
			for _, doc := range docs {
				val := resolveExpr(doc, fieldExpr)
				if f, ok := toFloat64(val); ok {
					sum += f
					count++
				}
			}
			if count == 0 {
				return nil, nil
			}
			return sum / float64(count), nil
		case "$min":
			var min interface{}
			for _, doc := range docs {
				val := resolveExpr(doc, fieldExpr)
				if min == nil || compareValues(val, min) < 0 {
					min = val
				}
			}
			return min, nil
		case "$max":
			var max interface{}
			for _, doc := range docs {
				val := resolveExpr(doc, fieldExpr)
				if max == nil || compareValues(val, max) > 0 {
					max = val
				}
			}
			return max, nil
		case "$first":
			if len(docs) == 0 {
				return nil, nil
			}
			return resolveExpr(docs[0], fieldExpr), nil
		case "$last":
			if len(docs) == 0 {
				return nil, nil
			}
			return resolveExpr(docs[len(docs)-1], fieldExpr), nil
		case "$push":
			arr := make([]interface{}, 0, len(docs))
			for _, doc := range docs {
				arr = append(arr, resolveExpr(doc, fieldExpr))
			}
			return arr, nil
		case "$addToSet":
			seen := map[string]bool{}
			arr := []interface{}{}
			for _, doc := range docs {
				v := resolveExpr(doc, fieldExpr)
				k := fmt.Sprintf("%v", v)
				if !seen[k] {
					seen[k] = true
					arr = append(arr, v)
				}
			}
			return arr, nil
		case "$count":
			return len(docs), nil
		}
	}
	return nil, fmt.Errorf("unknown accumulator operator in %v", accExpr)
}
