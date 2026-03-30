package sdk_entity_aggregation

import "fmt"

func mergeResults(
	entityResults map[string][]map[string]interface{},
	order []string,
	opts MergeResultsOptions,
) []map[string]interface{} {
	merged := map[string]map[string]interface{}{}
	keyOrder := []string{}

	for _, entity := range order {
		docs, ok := entityResults[entity]
		if !ok {
			continue
		}
		for _, doc := range docs {
			key := fmt.Sprintf("%v", doc[opts.On])
			if _, exists := merged[key]; !exists {
				merged[key] = map[string]interface{}{opts.On: doc[opts.On]}
				keyOrder = append(keyOrder, key)
			}
			if len(opts.Fields) > 0 {
				for _, field := range opts.Fields {
					if val, ok := doc[field]; ok {
						merged[key][field] = val
					}
				}
			} else {
				for k, v := range doc {
					merged[key][k] = v
				}
			}
		}
	}

	result := make([]map[string]interface{}, 0, len(merged))
	for _, key := range keyOrder {
		result = append(result, merged[key])
	}
	return result
}
