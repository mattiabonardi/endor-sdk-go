package sdk_entity_aggregation

import (
	"context"
	"fmt"
	"strings"

	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
)

// extractReferenceIDsFromFlatDocs walks docs using schema to find fields annotated
// with UISchema.Entity and collects their values as deduplicated reference IDs.
// Returns a map[entityName][]ids.
func extractReferenceIDsFromFlatDocs(schema *sdk.Schema, docs []map[string]interface{}) map[string][]string {
	entityIDs := make(map[string][]string)
	if schema == nil || schema.Properties == nil {
		return entityIDs
	}
	seen := make(map[string]map[string]bool)
	for _, doc := range docs {
		for propName, propSchema := range *schema.Properties {
			if propSchema.UISchema == nil || propSchema.UISchema.Entity == nil {
				continue
			}
			val, ok := doc[propName]
			if !ok || val == nil {
				continue
			}
			id := fmt.Sprintf("%v", val)
			if id == "" {
				continue
			}
			entity := *propSchema.UISchema.Entity
			if seen[entity] == nil {
				seen[entity] = make(map[string]bool)
			}
			if !seen[entity][id] {
				seen[entity][id] = true
				entityIDs[entity] = append(entityIDs[entity], id)
			}
		}
	}
	return entityIDs
}

// resolveReferences looks up entity IDs in the RepositoryRegistry and returns
// the merged EntityRefererenceGroup.
func resolveReferences(ctx context.Context, entityIDs map[string][]string, di sdk.EndorDIContainerInterface) (sdk.EntityRefererenceGroup, error) {
	if len(entityIDs) == 0 {
		return nil, nil
	}
	refs := make(sdk.EntityRefererenceGroup)
	for entityName, ids := range entityIDs {
		repo, ok := di.GetRepositories()[entityName]
		if !ok {
			continue
		}
		descriptions, err := repo.FindReferences(ctx, sdk.ReadInstancesDTO{Ids: ids})
		if err != nil {
			return nil, fmt.Errorf("references for entity %q: %w", entityName, err)
		}
		if len(descriptions) > 0 {
			refs[entityName] = descriptions
		}
	}
	if len(refs) == 0 {
		return nil, nil
	}
	return refs, nil
}

// mergeEntityRefererenceGroups merges multiple EntityRefererenceGroup values into one.
func mergeEntityRefererenceGroups(groups ...sdk.EntityRefererenceGroup) sdk.EntityRefererenceGroup {
	merged := make(sdk.EntityRefererenceGroup)
	for _, g := range groups {
		for entity, descs := range g {
			if merged[entity] == nil {
				merged[entity] = make(sdk.EntityReferenceGroupDescriptions)
			}
			for id, desc := range descs {
				merged[entity][id] = desc
			}
		}
	}
	if len(merged) == 0 {
		return nil
	}
	return merged
}

// deriveSchemaAfterPipeline applies each stage in the pipeline to evolve the
// schema shape. Only $group stages alter the schema; $match leaves it unchanged.
func deriveSchemaAfterPipeline(schema *sdk.Schema, pipeline []StageSpec) sdk.Schema {
	current := *schema
	for _, stage := range pipeline {
		if groupSpec, ok := stage["$group"]; ok {
			if g, ok := groupSpec.(map[string]interface{}); ok {
				current = deriveSchemaAfterGroup(&current, g)
			}
		}
	}
	return current
}

// deriveSchemaAfterGroup computes the output schema of a $group stage.
// Direct field references ("$fieldName") inherit the full source schema, preserving
// UISchema.Entity and other annotations. Accumulators produce typed schemas:
// $sum/$avg/$min/$max → number, $count → integer, $push/$addToSet → array,
// $first/$last → inherit from source field (same as a direct reference).
func deriveSchemaAfterGroup(inputSchema *sdk.Schema, groupSpec map[string]interface{}) sdk.Schema {
	props := make(map[string]sdk.Schema)
	for outField, expr := range groupSpec {
		props[outField] = inferGroupFieldSchema(inputSchema, expr)
	}
	return sdk.Schema{Type: sdk.SchemaTypeObject, Properties: &props}
}

// inferGroupFieldSchema returns the Schema for a single $group output field expression.
func inferGroupFieldSchema(inputSchema *sdk.Schema, expr interface{}) sdk.Schema {
	// Direct field reference: inherit the full source schema.
	if s, ok := expr.(string); ok && strings.HasPrefix(s, "$") {
		return inheritSourceSchema(inputSchema, s[1:])
	}
	// Accumulator object.
	if accMap, ok := expr.(map[string]interface{}); ok {
		for op, fieldExpr := range accMap {
			switch op {
			case "$sum", "$avg", "$min", "$max":
				return sdk.Schema{Type: sdk.SchemaTypeNumber}
			case "$count":
				return sdk.Schema{Type: sdk.SchemaTypeInteger}
			case "$push", "$addToSet":
				return sdk.Schema{Type: sdk.SchemaTypeArray}
			case "$first", "$last":
				if s, ok := fieldExpr.(string); ok && strings.HasPrefix(s, "$") {
					return inheritSourceSchema(inputSchema, s[1:])
				}
				return sdk.Schema{Type: sdk.SchemaTypeString}
			}
		}
	}
	// Literal constant (e.g. fixed-value group key).
	return sdk.Schema{Type: sdk.SchemaTypeString}
}

// inheritSourceSchema looks up fieldName in inputSchema.Properties and returns
// its schema (preserving UISchema.Entity and all other annotations). Falls back
// to a plain string schema when the field is not found.
func inheritSourceSchema(inputSchema *sdk.Schema, fieldName string) sdk.Schema {
	if inputSchema != nil && inputSchema.Properties != nil {
		if srcSchema, ok := (*inputSchema.Properties)[fieldName]; ok {
			return srcSchema
		}
	}
	return sdk.Schema{Type: sdk.SchemaTypeString}
}

// mergeSchemas merges the Properties of all schemas identified by dependsOn into
// a single object schema. Later entries overwrite earlier ones on key conflicts.
func mergeSchemas(stageSchemas map[string]*sdk.Schema, dependsOn []string) *sdk.Schema {
	allProps := make(map[string]sdk.Schema)
	for _, stageID := range dependsOn {
		s, ok := stageSchemas[stageID]
		if !ok || s == nil || s.Properties == nil {
			continue
		}
		for k, v := range *s.Properties {
			// First-writer-wins: the first stage that introduces a field
			// determines its schema (preserving UISchema, entity refs, etc.).
			if _, exists := allProps[k]; !exists {
				allProps[k] = v
			}
		}
	}
	if len(allProps) == 0 {
		return nil
	}
	return &sdk.Schema{Type: sdk.SchemaTypeObject, Properties: &allProps}
}
