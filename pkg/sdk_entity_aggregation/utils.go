package sdk_entity_aggregation

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"strings"
)

// generateStageID produces a unique stage identifier by appending 4 random bytes
// (hex-encoded) to the entity name. This ensures two stages targeting the same
// entity never collide in stageResults.
func generateStageID(entity string) string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	if entity != "" {
		return fmt.Sprintf("%s_%x", entity, b)
	}
	return fmt.Sprintf("stage_%x", b)
}

// remarshal round-trips a value through JSON to decode an interface{} into a
// typed struct without requiring the caller to hold a json.RawMessage.
func remarshal(src interface{}, dst interface{}) error {
	b, err := json.Marshal(src)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, dst)
}

// resolveExpr evaluates a field reference ("$fieldName") or returns a literal.
func resolveExpr(doc map[string]interface{}, expr interface{}) interface{} {
	if s, ok := expr.(string); ok && strings.HasPrefix(s, "$") {
		return getFieldValue(doc, s[1:])
	}
	return expr
}

// getFieldValue resolves a dot-notation path in a document.
func getFieldValue(doc map[string]interface{}, path string) interface{} {
	parts := strings.SplitN(path, ".", 2)
	val, ok := doc[parts[0]]
	if !ok {
		return nil
	}
	if len(parts) == 1 {
		return val
	}
	if nested, ok := val.(map[string]interface{}); ok {
		return getFieldValue(nested, parts[1])
	}
	return nil
}

// equals compares two values for equality, normalising numeric types.
func equals(a, b interface{}) bool {
	fa, okA := toFloat64(a)
	fb, okB := toFloat64(b)
	if okA && okB {
		return fa == fb
	}
	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}

// compareValues returns -1, 0, or 1 (numeric-aware).
func compareValues(a, b interface{}) int {
	fa, okA := toFloat64(a)
	fb, okB := toFloat64(b)
	if okA && okB {
		if fa < fb {
			return -1
		}
		if fa > fb {
			return 1
		}
		return 0
	}
	sa := fmt.Sprintf("%v", a)
	sb := fmt.Sprintf("%v", b)
	if sa < sb {
		return -1
	}
	if sa > sb {
		return 1
	}
	return 0
}

func toFloat64(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int32:
		return float64(n), true
	case int64:
		return float64(n), true
	case uint:
		return float64(n), true
	case uint32:
		return float64(n), true
	case uint64:
		return float64(n), true
	default:
		return 0, false
	}
}

func toSliceOfMaps(v interface{}) ([]map[string]interface{}, bool) {
	arr, ok := v.([]interface{})
	if !ok {
		return nil, false
	}
	result := make([]map[string]interface{}, 0, len(arr))
	for _, item := range arr {
		m, ok := item.(map[string]interface{})
		if !ok {
			return nil, false
		}
		result = append(result, m)
	}
	return result, true
}

func toBool(v interface{}) bool {
	switch b := v.(type) {
	case bool:
		return b
	default:
		f, ok := toFloat64(v)
		return ok && f != 0
	}
}

func toSlice(v interface{}) []interface{} {
	switch arr := v.(type) {
	case []interface{}:
		return arr
	default:
		return nil
	}
}

func entityListToSliceOfMaps(items []interface{}) ([]map[string]interface{}, error) {
	result := make([]map[string]interface{}, 0, len(items))
	for _, item := range items {
		data, err := json.Marshal(item)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal entity instance: %w", err)
		}
		var doc map[string]interface{}
		if err := json.Unmarshal(data, &doc); err != nil {
			return nil, fmt.Errorf("failed to unmarshal entity instance: %w", err)
		}
		result = append(result, doc)
	}
	return result, nil
}
