package sdk_entity

import (
	"fmt"

	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk_configuration"
)

func NewEntityService(microServiceId string, services *[]sdk.EndorServiceInterface, repository *sdk.EntityRepositoryInterface, logger *sdk.Logger, priority int) sdk.EndorServiceInterface {
	var repo sdk.EntityRepositoryInterface
	if repository == nil {
		repo = NewEndorServiceRepository(microServiceId, services, logger)
	} else {
		repo = *repository
	}
	entityService := EntityService{
		microServiceId: microServiceId,
		services:       services,
		repository:     repo,
	}

	// hybrid category actions
	hybridActions := map[string]sdk.EndorServiceActionInterface{
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
	hybridSpecializedActions := map[string]sdk.EndorServiceActionInterface{
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
	dynamicActions := map[string]sdk.EndorServiceActionInterface{}
	// dynamic specialized category actions
	dynamicSpecializedActions := map[string]sdk.EndorServiceActionInterface{}

	if sdk_configuration.GetConfig().HybridEntitiesEnabled || sdk_configuration.GetConfig().DynamicEntitiesEnabled {
		hybridActions["update"] = sdk.NewAction(entityService.updateHybrid, "Update an existing entity of type "+string(sdk.EntityTypeHybrid))
		hybridSpecializedActions["update"] = sdk.NewAction(entityService.updateHybridSpecialized, "Update an existing entity of type "+string(sdk.EntityTypeHybridSpecialized))
	}
	if sdk_configuration.GetConfig().DynamicEntitiesEnabled {
		dynamicActions["schema"] = sdk.NewAction(entityService.schema(entityService.getDynamicSchema()), "Get the schema of the entity of type "+string(sdk.EntityTypeDynamic))
		dynamicActions["instance"] = sdk.NewAction(entityService.instance(sdk.EntityTypeDynamic, sdk.NewSchema(&sdk.EntityHybrid{})), "Get the specified instance of entities of type "+string(sdk.EntityTypeDynamic))
		dynamicActions["list"] = sdk.NewAction(entityService.list(sdk.EntityTypeDynamic, sdk.NewSchema(&sdk.EntityHybrid{})), "Search for available entities of type "+string(sdk.EntityTypeDynamic))
		dynamicActions["create"] = sdk.NewAction(entityService.createDynamic, "Create a new entity "+string(sdk.EntityTypeDynamic))
		dynamicActions["update"] = sdk.NewAction(entityService.updateDynamic, "Update an existing entity of type "+string(sdk.EntityTypeDynamic))
		dynamicActions["delete"] = sdk.NewAction(entityService.delete(sdk.EntityTypeDynamic), "Delete an existing entity "+string(sdk.EntityTypeDynamic))

		dynamicSpecializedActions["schema"] = sdk.NewAction(entityService.schema(entityService.getDynamicSpecializedSchema()), "Get the schema of the entity of type "+string(sdk.EntityTypeDynamicSpecialized))
		dynamicSpecializedActions["instance"] = sdk.NewAction(entityService.instance(sdk.EntityTypeDynamicSpecialized, sdk.NewSchema(&sdk.EntityHybridSpecialized{})), "Get the specified instance of entities of type "+string(sdk.EntityTypeDynamicSpecialized))
		dynamicSpecializedActions["list"] = sdk.NewAction(entityService.list(sdk.EntityTypeDynamicSpecialized, sdk.NewSchema(&sdk.EntityHybridSpecialized{})), "Search for available entities of type "+string(sdk.EntityTypeDynamicSpecialized))
		dynamicSpecializedActions["create"] = sdk.NewAction(entityService.createDynamicSpecalized, "Create a new entity "+string(sdk.EntityTypeDynamicSpecialized))
		dynamicSpecializedActions["update"] = sdk.NewAction(entityService.updateDynamicSpecialized, "Update an existing entity of type "+string(sdk.EntityTypeDynamicSpecialized))
		dynamicSpecializedActions["delete"] = sdk.NewAction(entityService.delete(sdk.EntityTypeDynamicSpecialized), "Delete an existing entity "+string(sdk.EntityTypeDynamicSpecialized))
	}

	return NewEndorBaseSpecializedService[*sdk.Entity]("entity", "Entity").
		WithPriority(priority).
		WithActions(map[string]sdk.EndorServiceActionInterface{
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
		[]sdk.EndorBaseSpecializedServiceCategoryInterface{
			NewEndorBaseSpecializedServiceCategory[*sdk.Entity](string(sdk.EntityTypeBase), "Base").
				WithActions(map[string]sdk.EndorServiceActionInterface{
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
			NewEndorBaseSpecializedServiceCategory[*sdk.EntitySpecialized](string(sdk.EntityTypeBaseSpecialized), "Base specialized").
				WithActions(map[string]sdk.EndorServiceActionInterface{
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
			NewEndorBaseSpecializedServiceCategory[*sdk.EntityHybrid](string(sdk.EntityTypeHybrid), "Hybrid").
				WithActions(hybridActions),
			NewEndorBaseSpecializedServiceCategory[*sdk.EntityHybridSpecialized](string(sdk.EntityTypeHybridSpecialized), "Hybrid specialized").
				WithActions(hybridSpecializedActions),
			NewEndorBaseSpecializedServiceCategory[*sdk.EntityHybrid](string(sdk.EntityTypeDynamic), "Dynamic").
				WithActions(dynamicActions),
			NewEndorBaseSpecializedServiceCategory[*sdk.EntityHybridSpecialized](string(sdk.EntityTypeDynamicSpecialized), "Dynamic specialized").
				WithActions(dynamicSpecializedActions),
		},
	)
}

type EntityService struct {
	microServiceId string
	services       *[]sdk.EndorServiceInterface
	repository     sdk.EntityRepositoryInterface
}

func (h *EntityService) schema(schema *sdk.RootSchema) func(c *sdk.EndorContext[sdk.NoPayload]) (*sdk.Response[any], error) {
	return func(c *sdk.EndorContext[sdk.NoPayload]) (*sdk.Response[any], error) {
		return sdk.NewResponseBuilder[any]().AddSchema(schema).Build(), nil
	}
}

func (h *EntityService) list(entityType sdk.EntityType, schema *sdk.RootSchema) func(c *sdk.EndorContext[sdk.NoPayload]) (*sdk.Response[[]sdk.EntityInterface], error) {
	return func(c *sdk.EndorContext[sdk.NoPayload]) (*sdk.Response[[]sdk.EntityInterface], error) {
		entities, err := h.repository.List(&entityType)
		if err != nil {
			return nil, err
		}
		return sdk.NewResponseBuilder[[]sdk.EntityInterface]().AddData(&entities).AddSchema(schema).Build(), nil
	}
}

func (h *EntityService) instance(entityType sdk.EntityType, schema *sdk.RootSchema) func(c *sdk.EndorContext[sdk.ReadInstanceDTO]) (*sdk.Response[sdk.EntityInterface], error) {
	return func(c *sdk.EndorContext[sdk.ReadInstanceDTO]) (*sdk.Response[sdk.EntityInterface], error) {
		entity, err := h.repository.Instance(&entityType, c.Payload)
		if err != nil {
			return nil, err
		}
		return sdk.NewResponseBuilder[sdk.EntityInterface]().AddData(entity).AddSchema(schema).Build(), nil
	}
}

func (h *EntityService) createDynamic(c *sdk.EndorContext[sdk.CreateDTO[sdk.EntityHybrid]]) (*sdk.Response[sdk.EntityInterface], error) {
	return h.create(sdk.CreateDTO[sdk.EntityInterface]{
		Data: &c.Payload.Data,
	}, sdk.EntityTypeDynamic, sdk.NewSchema(sdk.EntityHybrid{}))
}

func (h *EntityService) createDynamicSpecalized(c *sdk.EndorContext[sdk.CreateDTO[sdk.EntityHybrid]]) (*sdk.Response[sdk.EntityInterface], error) {
	return h.create(sdk.CreateDTO[sdk.EntityInterface]{
		Data: &c.Payload.Data,
	}, sdk.EntityTypeDynamicSpecialized, sdk.NewSchema(sdk.EntityHybridSpecialized{}))
}

func (h *EntityService) create(dto sdk.CreateDTO[sdk.EntityInterface], entityType sdk.EntityType, schema *sdk.RootSchema) (*sdk.Response[sdk.EntityInterface], error) {
	entity, err := h.repository.Create(&entityType, dto)
	if err != nil {
		return nil, err
	}
	return sdk.NewResponseBuilder[sdk.EntityInterface]().AddData(entity).AddSchema(sdk.NewSchema(schema)).AddMessage(sdk.NewMessage(sdk.ResponseMessageGravityInfo, fmt.Sprintf("entity %s created", dto.Data.GetID()))).Build(), nil
}

func (h *EntityService) updateHybrid(c *sdk.EndorContext[sdk.UpdateByIdDTO[sdk.EntityHybrid]]) (*sdk.Response[sdk.EntityInterface], error) {
	return h.update(sdk.UpdateByIdDTO[sdk.EntityInterface]{
		Id:   c.Payload.Data.ID,
		Data: &c.Payload.Data,
	}, sdk.EntityTypeHybrid, sdk.NewSchema(sdk.EntityHybrid{}))
}

func (h *EntityService) updateHybridSpecialized(c *sdk.EndorContext[sdk.UpdateByIdDTO[sdk.EntityHybridSpecialized]]) (*sdk.Response[sdk.EntityInterface], error) {
	return h.update(sdk.UpdateByIdDTO[sdk.EntityInterface]{
		Id:   c.Payload.Data.ID,
		Data: &c.Payload.Data,
	}, sdk.EntityTypeHybridSpecialized, sdk.NewSchema(sdk.EntityHybridSpecialized{}))
}

func (h *EntityService) updateDynamic(c *sdk.EndorContext[sdk.UpdateByIdDTO[sdk.EntityHybrid]]) (*sdk.Response[sdk.EntityInterface], error) {
	return h.update(sdk.UpdateByIdDTO[sdk.EntityInterface]{
		Id:   c.Payload.Data.ID,
		Data: &c.Payload.Data,
	}, sdk.EntityTypeDynamic, sdk.NewSchema(sdk.EntityHybrid{}))
}

func (h *EntityService) updateDynamicSpecialized(c *sdk.EndorContext[sdk.UpdateByIdDTO[sdk.EntityHybridSpecialized]]) (*sdk.Response[sdk.EntityInterface], error) {
	return h.update(sdk.UpdateByIdDTO[sdk.EntityInterface]{
		Id:   c.Payload.Data.ID,
		Data: &c.Payload.Data,
	}, sdk.EntityTypeDynamicSpecialized, sdk.NewSchema(sdk.EntityHybridSpecialized{}))
}

func (h *EntityService) update(dto sdk.UpdateByIdDTO[sdk.EntityInterface], entityType sdk.EntityType, schema *sdk.RootSchema) (*sdk.Response[sdk.EntityInterface], error) {
	entity, err := h.repository.Update(&entityType, dto)
	if err != nil {
		return nil, err
	}
	return sdk.NewResponseBuilder[sdk.EntityInterface]().AddData(entity).AddSchema(sdk.NewSchema(schema)).AddMessage(sdk.NewMessage(sdk.ResponseMessageGravityInfo, fmt.Sprintf("entity %s updated", dto.Id))).Build(), nil
}

func (h *EntityService) delete(entityType sdk.EntityType) func(c *sdk.EndorContext[sdk.ReadInstanceDTO]) (*sdk.Response[sdk.NoPayload], error) {
	return func(c *sdk.EndorContext[sdk.ReadInstanceDTO]) (*sdk.Response[sdk.NoPayload], error) {
		err := h.repository.Delete(&entityType, c.Payload)
		if err != nil {
			return nil, err
		}
		return sdk.NewResponseBuilder[sdk.NoPayload]().AddMessage(sdk.NewMessage(sdk.ResponseMessageGravityInfo, fmt.Sprintf("entity %s deleted", c.Payload.Id))).Build(), nil
	}
}

func (h *EntityService) getDynamicSchema() *sdk.RootSchema {
	schema := sdk.NewSchema(sdk.EntityHybrid{})
	// define service as readOnly
	properties := *schema.Schema.Properties
	serviceSchema := properties["service"]
	readOnly := false
	serviceSchema.ReadOnly = &readOnly
	properties["service"] = serviceSchema
	return schema
}

func (h *EntityService) getDynamicSpecializedSchema() *sdk.RootSchema {
	schema := sdk.NewSchema(sdk.EntityHybridSpecialized{})
	// define service as readOnly
	properties := *schema.Schema.Properties
	serviceSchema := properties["service"]
	readOnly := false
	serviceSchema.ReadOnly = &readOnly
	properties["service"] = serviceSchema
	return schema
}
