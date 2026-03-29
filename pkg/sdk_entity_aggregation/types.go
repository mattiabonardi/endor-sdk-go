package sdk_entity_aggregation

import "encoding/json"

// AggregationPipeline is the top-level payload for the aggregation execute action.
// Each element is either an EntityPipelineStage (has "entity" key) or a
// post-processing operator ($mergeResults).
//
// Example:
//
//	[
//	  { "entity": "order", "pipeline": [{ "$match": { "status": "completed" } }, { "$group": { "_id": "$customerId", "totalSpent": { "$sum": "$amount" } } }] },
//	  { "entity": "review", "pipeline": [{ "$match": { "rating": { "$gte": 4 } } }, { "$group": { "_id": "$userId", "positiveReviews": { "$sum": 1 } } }] },
//	  { "$mergeResults": { "on": "_id", "fields": ["totalSpent", "positiveReviews"] } }
//	]
type AggregationPipeline []json.RawMessage

// EntityPipelineStage targets a specific local entity and applies a sequence
// of pipeline stages to its data.
type EntityPipelineStage struct {
	Entity   string      `json:"entity"`
	Pipeline []StageSpec `json:"pipeline"`
}

// StageSpec represents a single pipeline stage as a raw key→value map.
// Supported keys: $match, $group.
type StageSpec map[string]interface{}

// MergeResultsOptions configures the $mergeResults top-level operator, which
// joins results from multiple entity stages on a common key.
type MergeResultsOptions struct {
	// On is the field name used as the join key.
	On string `json:"on"`
	// Fields lists which fields to include from each entity result.
	// When empty, all fields are merged.
	Fields []string `json:"fields"`
}
