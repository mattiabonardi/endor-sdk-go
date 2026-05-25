package sdk_entity_aggregation

import (
	"fmt"

	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk_entity"
)

const aggregationEntity = "aggregation"

// NewAggregationHandler builds an EndorBaseHandlerInterface for the "aggregation"
// entity and registers the "execute" action, which runs an AggregationPipeline
// against the local RepositoryRegistry.
// Pass a non-nil EntityStageHandler to override the default entity-stage
// execution (e.g. to route stages to remote microservices in a core service).
// Additional AggregationEngineOption values are applied after the handler.
func NewAggregationHandler(priority int, handler EntityStageHandler, opts ...AggregationEngineOption) sdk.EndorBaseHandlerInterface {
	if handler != nil {
		opts = append([]AggregationEngineOption{WithEntityStageHandler(handler)}, opts...)
	}
	return sdk_entity.NewEndorBaseHandler[aggregationEntity_](
		aggregationEntity,
		"${t.sdk.aggregation.handler.title}",
	).WithPriority(priority).WithActions(map[string]sdk.EndorHandlerActionInterface{
		"execute": sdk.NewConfigurableAction(
			sdk.EndorHandlerActionOptions{
				Description:           "${t.sdk.aggregation.handler.actions.execute}",
				Public:                false,
				SkipPayloadValidation: false,
				InputSchema:           buildPipelineSchema(),
			},
			func(c *sdk.EndorContext[AggregationPipeline]) (*sdk.Response[[]map[string]interface{}], error) {
				// opts is captured from the outer scope and already includes the
				// executor option when a non-nil executor was provided.
				engine := NewAggregationEngine(c.Session, c.DIContainer, opts...)
				result, schema, refs, err := engine.Execute(c.GinContext.Request.Context(), c.Payload)
				if err != nil {
					return nil, sdk.NewBadRequestError(fmt.Errorf("aggregation failed: %w", err)).WithTranslation("sdk.aggregation.messages.failed", nil)
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
	description := "${t.sdk.aggregation.fields.pipeline_description}"
	return &sdk.RootSchema{
		Schema: sdk.Schema{
			Type:        sdk.SchemaTypeArray,
			Description: &description,
		},
	}
}
