package sdk_entity_aggregation

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
)

// AggregationEngine executes aggregation pipelines against the local
// RepositoryRegistry. Supported entity-level stages: $match, $group.
// Supported top-level operators: $mergeResults.
type AggregationEngine struct{}

// NewAggregationEngine returns a ready-to-use AggregationEngine.
func NewAggregationEngine() *AggregationEngine {
	return &AggregationEngine{}
}

// Execute runs the aggregation pipeline and returns the final document set.
func (e *AggregationEngine) Execute(ctx context.Context, pipeline AggregationPipeline) ([]map[string]interface{}, error) {
	entityResults := map[string][]map[string]interface{}{}
	entityOrder := []string{}
	currentResult := []map[string]interface{}{}
	hasResult := false

	for _, rawStage := range pipeline {
		// Try to decode as an entity stage (presence of "entity" field).
		var entityStage EntityPipelineStage
		if err := json.Unmarshal(rawStage, &entityStage); err == nil && entityStage.Entity != "" {
			docs, err := e.executeEntityStage(ctx, entityStage)
			if err != nil {
				return nil, fmt.Errorf("entity stage %q: %w", entityStage.Entity, err)
			}
			if _, seen := entityResults[entityStage.Entity]; !seen {
				entityOrder = append(entityOrder, entityStage.Entity)
			}
			entityResults[entityStage.Entity] = docs
			continue
		}

		// Decode as a top-level operator map (one key per stage).
		var stageMap map[string]json.RawMessage
		if err := json.Unmarshal(rawStage, &stageMap); err != nil {
			return nil, fmt.Errorf("invalid pipeline stage: %w", err)
		}

		for op, value := range stageMap {
			switch op {
			case "$mergeResults":
				var opts MergeResultsOptions
				if err := json.Unmarshal(value, &opts); err != nil {
					return nil, fmt.Errorf("$mergeResults: %w", err)
				}
				currentResult = mergeResults(entityResults, entityOrder, opts)
				hasResult = true
			}
		}
	}

	// If no $mergeResults was applied, return the single entity result directly.
	if !hasResult {
		if len(entityOrder) == 1 {
			return entityResults[entityOrder[0]], nil
		}
		return currentResult, nil
	}
	return currentResult, nil
}

// executeEntityStage fetches documents for the entity and applies the
// pipeline stages in-memory. A leading $match is pushed down to the
// repository as a filter for efficiency.
func (e *AggregationEngine) executeEntityStage(ctx context.Context, stage EntityPipelineStage) ([]map[string]interface{}, error) {
	repo, ok := sdk.GetDocumentRepository(stage.Entity)
	if !ok {
		return nil, fmt.Errorf("entity %q not found in repository registry", stage.Entity)
	}

	// Push down a leading $match to the repository filter.
	filter := map[string]interface{}{}
	inMemoryFrom := 0

	if len(stage.Pipeline) > 0 {
		if matchVal, ok := stage.Pipeline[0]["$match"]; ok {
			if f, ok := matchVal.(map[string]interface{}); ok {
				filter = f
				inMemoryFrom = 1
			}
		}
	}

	docs, err := repo.ListDocuments(ctx, sdk.ReadDTO{Filter: filter})
	if err != nil {
		return nil, err
	}

	// Apply remaining stages in-memory.
	for _, s := range stage.Pipeline[inMemoryFrom:] {
		docs, err = e.applyEntityStage(docs, s)
		if err != nil {
			return nil, err
		}
	}
	return docs, nil
}

func (e *AggregationEngine) applyEntityStage(docs []map[string]interface{}, stage StageSpec) ([]map[string]interface{}, error) {
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
	return docs, nil
}

// ---------------------------------------------------------------------------
// In-memory stage implementations
// ---------------------------------------------------------------------------

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
