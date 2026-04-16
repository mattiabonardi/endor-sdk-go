package sdk_entity

import (
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
)

func NewEntityHandler(microServiceId string, handlers *[]sdk.EndorHandlerInterface, repository *sdk.EntityRepositoryInterface, logger *sdk.Logger, priority int, hybridEntitiesEnabled bool, dynamicEntitiesEnabled bool) sdk.EndorHandlerInterface {
	var repo sdk.EntityRepositoryInterface
	if repository == nil {
		// Use the singleton repository to ensure cache consistency
		if singletonRepo := GetEndorHandlerRepository(); singletonRepo != nil {
			repo = singletonRepo
		} else {
			// Fallback: initialize if not yet initialized (should not happen in normal flow)
			repo = InitEndorHandlerRepository(microServiceId, handlers, logger)
		}
	} else {
		repo = *repository
	}
	entityService := EntityHandler{
		microServiceId: microServiceId,
		handlers:       handlers,
		repository:     repo,
	}

	// hybrid actions
	hybridActions := map[string]sdk.EndorHandlerActionInterface{
		"schema": sdk.NewAction(
			entityService.schema(sdk.NewSchema(&sdk.EntityHybrid{})),
			"Get the schema of the entity of type "+string(sdk.EntityTypeHybrid),
		),
		"instance": sdk.NewAction(
			entityService.instance(sdk.EntityTypeHybrid, sdk.NewSchema(&sdk.EntityHybrid{})),
			"Get the specified instance of entities of type "+string(sdk.EntityTypeHybrid),
		),
		"list": sdk.NewAction(
			entityService.list(sdk.EntityTypeHybrid, sdk.NewSchema(&sdk.EntityHybrid{})),
			"Search for available entities of type "+string(sdk.EntityTypeHybrid),
		),
	}
	// hybrid specialized actions
	hybridSpecializedActions := map[string]sdk.EndorHandlerActionInterface{
		"schema": sdk.NewAction(
			entityService.schema(sdk.NewSchema(&sdk.EntityHybridSpecialized{})),
			"Get the schema of the entity of type "+string(sdk.EntityTypeHybridSpecialized),
		),
		"instance": sdk.NewAction(
			entityService.instance(sdk.EntityTypeHybridSpecialized, sdk.NewSchema(&sdk.EntityHybridSpecialized{})),
			"Get the specified instance of entities of type "+string(sdk.EntityTypeHybridSpecialized),
		),
		"list": sdk.NewAction(
			entityService.list(sdk.EntityTypeHybridSpecialized, sdk.NewSchema(&sdk.EntityHybridSpecialized{})),
			"Search for available entities of type "+string(sdk.EntityTypeHybridSpecialized),
		),
	}
	// dynamic category actions
	dynamicActions := map[string]sdk.EndorHandlerActionInterface{}
	// dynamic specialized category actions
	dynamicSpecializedActions := map[string]sdk.EndorHandlerActionInterface{}

	if dynamicEntitiesEnabled {
		dynamicActions["schema"] = sdk.NewAction(entityService.schema(entityService.getDynamicSchema(&sdk.EntityHybrid{})), "Get the schema of the entity of type "+string(sdk.EntityTypeDynamic))
		dynamicActions["instance"] = sdk.NewAction(entityService.instance(sdk.EntityTypeDynamic, sdk.NewSchema(&sdk.EntityHybrid{})), "Get the specified instance of entities of type "+string(sdk.EntityTypeDynamic))
		dynamicActions["list"] = sdk.NewAction(entityService.list(sdk.EntityTypeDynamic, sdk.NewSchema(&sdk.EntityHybrid{})), "Search for available entities of type "+string(sdk.EntityTypeDynamic))

		dynamicSpecializedActions["schema"] = sdk.NewAction(entityService.schema(entityService.getDynamicSchema(&sdk.EntityHybridSpecialized{})), "Get the schema of the entity of type "+string(sdk.EntityTypeDynamicSpecialized))
		dynamicSpecializedActions["instance"] = sdk.NewAction(entityService.instance(sdk.EntityTypeDynamicSpecialized, sdk.NewSchema(&sdk.EntityHybridSpecialized{})), "Get the specified instance of entities of type "+string(sdk.EntityTypeDynamicSpecialized))
		dynamicSpecializedActions["list"] = sdk.NewAction(entityService.list(sdk.EntityTypeDynamicSpecialized, sdk.NewSchema(&sdk.EntityHybridSpecialized{})), "Search for available entities of type "+string(sdk.EntityTypeDynamicSpecialized))
	}

	return NewEndorBaseSpecializedHandler[*sdk.Entity]("entity", "Entity").
		WithPriority(priority).
		WithActions(map[string]sdk.EndorHandlerActionInterface{
			"schema": sdk.NewAction(
				entityService.schema(sdk.NewSchema(&sdk.Entity{})),
				"Get the schema of the entity",
			),
			"instance": sdk.NewAction(
				entityService.instance("", sdk.NewSchema(&sdk.Entity{})),
				"Get the specified instance of entities",
			),
			"list": sdk.NewAction(
				entityService.list("", sdk.NewSchema(&sdk.Entity{})),
				"Search for available entities",
			)}).WithCategories(
		[]sdk.EndorBaseSpecializedHandlerCategoryInterface{
			NewEndorBaseSpecializedHandlerCategory[*sdk.Entity](string(sdk.EntityTypeBase), "Base").
				WithActions(map[string]sdk.EndorHandlerActionInterface{
					"schema": sdk.NewAction(
						entityService.schema(sdk.NewSchema(&sdk.Entity{})),
						"Get the schema of the entity of type "+string(sdk.EntityTypeBase),
					),
					"instance": sdk.NewAction(
						entityService.instance(sdk.EntityTypeBase, sdk.NewSchema(&sdk.Entity{})),
						"Get the specified instance of entities of type "+string(sdk.EntityTypeBase),
					),
					"list": sdk.NewAction(
						entityService.list(sdk.EntityTypeBase, sdk.NewSchema(&sdk.Entity{})),
						"Search for available entities of type "+string(sdk.EntityTypeBase),
					),
				}),
			NewEndorBaseSpecializedHandlerCategory[*sdk.EntitySpecialized](string(sdk.EntityTypeBaseSpecialized), "Base specialized").
				WithActions(map[string]sdk.EndorHandlerActionInterface{
					"schema": sdk.NewAction(
						entityService.schema(sdk.NewSchema(&sdk.EntitySpecialized{})),
						"Get the schema of the entity of type "+string(sdk.EntityTypeBaseSpecialized),
					),
					"instance": sdk.NewAction(
						entityService.instance(sdk.EntityTypeBaseSpecialized, sdk.NewSchema(&sdk.EntitySpecialized{})),
						"Get the specified instance of entities of type "+string(sdk.EntityTypeBaseSpecialized),
					),
					"list": sdk.NewAction(
						entityService.list(sdk.EntityTypeBaseSpecialized, sdk.NewSchema(&sdk.EntitySpecialized{})),
						"Search for available entities of type "+string(sdk.EntityTypeBaseSpecialized),
					),
				}),
			NewEndorBaseSpecializedHandlerCategory[*sdk.EntityHybrid](string(sdk.EntityTypeHybrid), "Hybrid").
				WithActions(hybridActions),
			NewEndorBaseSpecializedHandlerCategory[*sdk.EntityHybridSpecialized](string(sdk.EntityTypeHybridSpecialized), "Hybrid specialized").
				WithActions(hybridSpecializedActions),
			NewEndorBaseSpecializedHandlerCategory[*sdk.EntityHybrid](string(sdk.EntityTypeDynamic), "Dynamic").
				WithActions(dynamicActions),
			NewEndorBaseSpecializedHandlerCategory[*sdk.EntityHybridSpecialized](string(sdk.EntityTypeDynamicSpecialized), "Dynamic specialized").
				WithActions(dynamicSpecializedActions),
		},
	)
}

type EntityHandler struct {
	microServiceId string
	handlers       *[]sdk.EndorHandlerInterface
	repository     sdk.EntityRepositoryInterface
}

func (h *EntityHandler) schema(schema *sdk.RootSchema) func(c *sdk.EndorContext[sdk.NoPayload]) (*sdk.Response[any], error) {
	return func(c *sdk.EndorContext[sdk.NoPayload]) (*sdk.Response[any], error) {
		return sdk.NewResponseBuilder[any]().AddSchema(schema).Build(), nil
	}
}

func (h *EntityHandler) list(entityType sdk.EntityType, schema *sdk.RootSchema) func(c *sdk.EndorContext[sdk.NoPayload]) (*sdk.Response[[]sdk.EntityInterface], error) {
	return func(c *sdk.EndorContext[sdk.NoPayload]) (*sdk.Response[[]sdk.EntityInterface], error) {
		entities, err := h.repository.List(&entityType)
		if err != nil {
			return nil, err
		}
		return sdk.NewResponseBuilder[[]sdk.EntityInterface]().AddData(&entities).AddSchema(schema).Build(), nil
	}
}

func (h *EntityHandler) instance(entityType sdk.EntityType, schema *sdk.RootSchema) func(c *sdk.EndorContext[sdk.ReadInstanceDTO]) (*sdk.Response[sdk.EntityInterface], error) {
	return func(c *sdk.EndorContext[sdk.ReadInstanceDTO]) (*sdk.Response[sdk.EntityInterface], error) {
		entity, err := h.repository.Instance(&entityType, c.Payload)
		if err != nil {
			return nil, err
		}
		return sdk.NewResponseBuilder[sdk.EntityInterface]().AddData(entity).AddSchema(schema).Build(), nil
	}
}

func (h *EntityHandler) getDynamicSchema(baseSchema sdk.EntityInterface) *sdk.RootSchema {
	schema := sdk.NewSchema(baseSchema)
	readOnly := false
	properties := *schema.Schema.Properties
	// define id as readOnly
	idSchema := properties["id"]
	idSchema.ReadOnly = &readOnly
	properties["id"] = idSchema
	// define description as readOnly
	descriptionSchema := properties["description"]
	descriptionSchema.ReadOnly = &readOnly
	properties["description"] = descriptionSchema
	// define service as readOnly
	serviceSchema := properties["service"]
	serviceSchema.ReadOnly = &readOnly
	query := "$filter(dynamicResourceEnabled == true)"
	serviceSchema.UISchema.Query = &query
	properties["service"] = serviceSchema
	return schema
}
