package sdk_entity_aggregation

import (
	"context"

	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
)

// AggregationPipeline is the top-level payload for the aggregation execute action.
// It is a sequence of EntityPipelineStages executed serially in declaration order.
// Each stage is an independent aggregation unit; a stage may declare dependencies
// on previously completed stages via DependsOn.
//
// Example:
//
//	[
//	  { "id": "grouped_orders", "entity": "order",    "pipeline": [{ "$group": { "id": "$customerId", "total": { "$sum": "$amount" } } }] },
//	  { "id": "customers",      "entity": "customer", "pipeline": [] },
//	  { "dependsOn": ["grouped_orders", "customers"], "pipeline": [{ "$mergeResults": { "on": "id" } }] }
//	]
type AggregationPipeline []EntityPipelineStage

// EntityPipelineStage is an independent aggregation unit. It fetches documents
// for Entity from the repository registry and applies Pipeline stages in sequence.
// When Entity is empty the stage starts with no documents, which is the intended
// pattern for post-processing stages (e.g. those that only run $mergeResults).
//
// ID is an optional stable identifier for this stage. When set, other stages
// reference it via DependsOn using this ID. When omitted, Entity is used as
// the fallback identifier (Entity must then be unique within the pipeline).
type EntityPipelineStage struct {
	ID        string      `json:"id,omitempty"`
	Entity    string      `json:"entity,omitempty"`
	DependsOn []string    `json:"dependsOn,omitempty"`
	Pipeline  []StageSpec `json:"pipeline"`
}

// StageSpec represents a single pipeline stage as a key→value map.
// Supported operators: $match, $group, $mergeResults.
type StageSpec map[string]interface{}

// EntityStageHandler is an optional callback that, when provided to
// AggregationEngine, fully replaces the built-in repository-fetch + in-memory
// pipeline execution for every EntityPipelineStage whose Entity field is
// non-empty. The callback is responsible for returning the complete result
// for that stage (including any pipeline operators it chooses to apply),
// the derived output schema, and the resolved entity reference group.
// Use WithEntityStageHandler to attach it to an engine.
type EntityStageHandler func(ctx context.Context, stage EntityPipelineStage) ([]map[string]interface{}, *sdk.Schema, sdk.EntityRefererenceGroup, error)

// MergeResultsOptions configures the $mergeResults operator, which joins the
// results of the stages listed in the enclosing EntityPipelineStage.DependsOn.
type MergeResultsOptions struct {
	// On is the field name used as the join key.
	On string `json:"on"`
	// Fields lists which fields to copy from each source doc.
	// When empty, all fields are merged.
	Fields []string `json:"fields"`
}
