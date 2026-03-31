package sdk_entity_aggregation

import (
	"fmt"

	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk_entity"
)

const aggregationEntity = "aggregation"
const aggregationEntityDescription = "Distributed aggregation pipeline over registered entity repositories"

// NewAggregationHandler builds an EndorBaseHandlerInterface for the "aggregation"
// entity and registers the "execute" action, which runs an AggregationPipeline
// against the local RepositoryRegistry.
func NewAggregationHandler(priority int, opts ...AggregationEngineOption) sdk.EndorBaseHandlerInterface {
	engine := NewAggregationEngine(opts...)

	return sdk_entity.NewEndorBaseHandler[aggregationEntity_](
		aggregationEntity,
		aggregationEntityDescription,
	).WithPriority(priority).WithActions(map[string]sdk.EndorHandlerActionInterface{
		"execute": sdk.NewConfigurableAction(
			sdk.EndorHandlerActionOptions{
				Description:           "Execute an aggregation pipeline over entity repositories",
				Public:                false,
				SkipPayloadValidation: false,
				InputSchema:           buildPipelineSchema(),
			},
			func(c *sdk.EndorContext[AggregationPipeline]) (*sdk.Response[[]map[string]interface{}], error) {
				result, schema, refs, err := engine.Execute(c.GinContext.Request.Context(), c.Payload)
				if err != nil {
					return nil, sdk.NewBadRequestError(fmt.Errorf("aggregation failed: %w", err))
				}
				return sdk.NewResponseBuilder[[]map[string]interface{}]().AddData(&result).
					AddReferences(refs).AddSchema(schema).Build(), nil
			},
		),
	})
}

// aggregationEntity_ is a minimal entity type used only to satisfy the
// EndorBaseHandler generic constraint. It is not stored in any repository.
type aggregationEntity_ struct{}

func (a aggregationEntity_) GetID() any { return nil }

// buildPipelineSchema returns a descriptive JSON Schema for the AggregationPipeline payload.
func buildPipelineSchema() *sdk.RootSchema {
	description := "Array of pipeline stages. Each stage is either an entity stage " +
		"{ entity, pipeline } or the top-level $mergeResults operator."
	return &sdk.RootSchema{
		Schema: sdk.Schema{
			Type:        sdk.SchemaTypeArray,
			Description: &description,
		},
	}
}
