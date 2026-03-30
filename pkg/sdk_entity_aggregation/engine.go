package sdk_entity_aggregation

import (
	"context"
	"fmt"

	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
)

// AggregationEngine executes aggregation pipelines against the local
// RepositoryRegistry. Supported operators (usable at any StageSpec level):
// $match, $group, $mergeResults.
type AggregationEngine struct{}

// NewAggregationEngine returns a ready-to-use AggregationEngine.
func NewAggregationEngine() *AggregationEngine {
	return &AggregationEngine{}
}

// stageExecContext carries shared state needed by operators during stage execution.
type stageExecContext struct {
	// stageResults holds the output of each completed EntityPipelineStage,
	// keyed by stage ID (ID field if set, Entity name otherwise).
	// Used by $mergeResults to reference other stages.
	stageResults map[string][]map[string]interface{}
	// dependsOn is the DependsOn list of the enclosing EntityPipelineStage,
	// which $mergeResults uses as the ordered set of stage IDs to merge.
	dependsOn []string
}

// Execute runs all EntityPipelineStages in declaration order and returns the
// result of the last stage. Each stage stores its output keyed by stageID so
// that subsequent stages can reference it via DependsOn / $mergeResults.
func (e *AggregationEngine) Execute(ctx context.Context, pipeline AggregationPipeline) ([]map[string]interface{}, error) {
	stageResults := map[string][]map[string]interface{}{}
	lastResult := []map[string]interface{}{}

	for _, stage := range pipeline {
		// Assign a unique ID once so all subsequent uses within this iteration
		// refer to the same key. An explicit ID is preserved as-is.
		if stage.ID == "" {
			stage.ID = generateStageID(stage.Entity)
		}
		docs, err := e.executeEntityStage(ctx, stage, stageResults)
		if err != nil {
			return nil, fmt.Errorf("stage %q: %w", stage.ID, err)
		}
		stageResults[stage.ID] = docs
		lastResult = docs
	}
	return lastResult, nil
}

// executeEntityStage fetches the initial document set for the stage (when Entity
// is registered in the repository registry) and then applies each StageSpec in
// sequence. A leading $match is pushed down to the repository as a filter.
// Stages with an empty Entity start with an empty document set.
func (e *AggregationEngine) executeEntityStage(
	ctx context.Context,
	stage EntityPipelineStage,
	stageResults map[string][]map[string]interface{},
) ([]map[string]interface{}, error) {
	docs := []map[string]interface{}{}
	pipeline := stage.Pipeline

	if stage.Entity != "" {
		repo, ok := sdk.GetDocumentRepository(stage.Entity)
		if !ok {
			return nil, fmt.Errorf("entity %q not found in repository registry", stage.Entity)
		}

		// Push down a leading $match to the repository filter for efficiency.
		filter := map[string]interface{}{}
		if len(pipeline) > 0 {
			if matchVal, ok := pipeline[0]["$match"]; ok {
				if f, ok := matchVal.(map[string]interface{}); ok {
					filter = f
					pipeline = pipeline[1:]
				}
			}
		}

		var err error
		docs, err = repo.ListDocuments(ctx, sdk.ReadDTO{Filter: filter})
		if err != nil {
			return nil, err
		}
	}

	execCtx := stageExecContext{
		stageResults: stageResults,
		dependsOn:    stage.DependsOn,
	}

	var err error
	for _, s := range pipeline {
		docs, err = e.applyStage(docs, s, execCtx)
		if err != nil {
			return nil, err
		}
	}
	return docs, nil
}

// applyStage applies a single StageSpec operator to the working document set.
// All supported operators are resolved here — there is no separate top-level
// vs entity-level distinction.
func (e *AggregationEngine) applyStage(docs []map[string]interface{}, stage StageSpec, ctx stageExecContext) ([]map[string]interface{}, error) {
	if matchSpec, ok := stage["$match"]; ok {
		if f, ok := matchSpec.(map[string]interface{}); ok {
			return applyMatch(docs, f), nil
		}
	}
	if groupSpec, ok := stage["$group"]; ok {
		if g, ok := groupSpec.(map[string]interface{}); ok {
			return applyGroup(docs, g)
		}
	}
	if mergeSpec, ok := stage["$mergeResults"]; ok {
		var opts MergeResultsOptions
		if err := remarshal(mergeSpec, &opts); err != nil {
			return nil, fmt.Errorf("$mergeResults: %w", err)
		}
		return mergeResults(ctx.stageResults, ctx.dependsOn, opts), nil
	}
	return docs, nil
}
