package sdk_entity

import (
	"fmt"

	"github.com/mattiabonardi/endor-sdk-go/internal/configuration"
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
)

func NewEntityService(microServiceId string, services *[]sdk.EndorServiceInterface) sdk.EndorServiceInterface {
	entityService := EntityService{
		microServiceId: microServiceId,
		services:       services,
	}

	// hybrid category actions
	hybridCategoryActions := map[string]sdk.EndorServiceActionInterface{
		"schema": sdk.NewAction(
			entityService.entityHybridSchema,
			"Get the schema of the entity of type "+string(sdk.EntityTypeHybrid),
		),
		"instance": sdk.NewAction(
			entityService.entityHybridInstance,
			"Get the specified instance of entities of type "+string(sdk.EntityTypeHybrid),
		),
		"list": sdk.NewAction(
			entityService.entityHybridList,
			"Search for available entities of type "+string(sdk.EntityTypeHybrid),
		),
	}
	// hybrid specialized actions
	hybridSpecializedCategoryActions := map[string]sdk.EndorServiceActionInterface{
		"schema": sdk.NewAction(
			entityService.entityHybridSpecializedSchema,
			"Get the schema of the entity of type "+string(sdk.EntityTypeHybridSpecialized),
		),
		"instance": sdk.NewAction(
			entityService.entityHybridSpecializedInstance,
			"Get the specified instance of entities of type "+string(sdk.EntityTypeHybridSpecialized),
		),
		"list": sdk.NewAction(
			entityService.entityHybridSpecializedList,
			"Search for available entities of type "+string(sdk.EntityTypeHybridSpecialized),
		),
	}
	// dynamic category actions
	dynamicCategoryActions := map[string]sdk.EndorServiceActionInterface{
		"schema": sdk.NewAction(
			entityService.entityHybridSchema,
			"Get the schema of the entity of type "+string(sdk.EntityTypeDynamic),
		),
		"instance": sdk.NewAction(
			entityService.entityDynamicInstance,
			"Get the specified instance of entities of type "+string(sdk.EntityTypeDynamic),
		),
		"list": sdk.NewAction(
			entityService.entityDynamicList,
			"Search for available entities of type "+string(sdk.EntityTypeDynamic),
		),
	}
	// dynamic specialized category actions
	dynamicSpecializedCategoryActions := map[string]sdk.EndorServiceActionInterface{
		"schema": sdk.NewAction(
			entityService.entityHybridSpecializedSchema,
			"Get the schema of the entity of type "+string(sdk.EntityTypeDynamicSpecialized),
		),
		"instance": sdk.NewAction(
			entityService.entityDynamicSpecializedInstance,
			"Get the specified instance of entities of type "+string(sdk.EntityTypeDynamicSpecialized),
		),
		"list": sdk.NewAction(
			entityService.entityDynamicSpecializedList,
			"Search for available entities of type "+string(sdk.EntityTypeDynamicSpecialized),
		),
	}

	if configuration.GetConfig().HybridEntitiesEnabled || configuration.GetConfig().DynamicEntitiesEnabled {
		hybridCategoryActions["update"] = sdk.NewAction(entityService.entityHybridUpdate, "Update an existing entity of type "+string(sdk.EntityTypeHybrid))
		hybridSpecializedCategoryActions["update"] = sdk.NewAction(entityService.entityHybridSpecializedUpdate, "Update an existing entity of type "+string(sdk.EntityTypeHybridSpecialized))
	}
	if configuration.GetConfig().DynamicEntitiesEnabled {
		dynamicCategoryActions["create"] = sdk.NewAction(entityService.entityDynamicCreate, "Create a new entity "+string(sdk.EntityTypeDynamic))
		dynamicCategoryActions["update"] = sdk.NewAction(entityService.entityDynamicUpdate, "Update an existing entity of type "+string(sdk.EntityTypeDynamic))
		dynamicCategoryActions["delete"] = sdk.NewAction(entityService.entityDynamicDelete, "Delete an existing entity "+string(sdk.EntityTypeDynamic))

		dynamicSpecializedCategoryActions["create"] = sdk.NewAction(entityService.entityDynamicSpecializedCreate, "Create a new entity "+string(sdk.EntityTypeDynamicSpecialized))
		dynamicSpecializedCategoryActions["update"] = sdk.NewAction(entityService.entityDynamicSpecializedUpdate, "Update an existing entity of type "+string(sdk.EntityTypeDynamicSpecialized))
		dynamicSpecializedCategoryActions["delete"] = sdk.NewAction(entityService.entityDynamicDelete, "Delete an existing entity "+string(sdk.EntityTypeDynamicSpecialized))
	}

	return NewEndorBaseSpecializedService[*sdk.Entity]("entity", "Entity").
		WithActions(map[string]sdk.EndorServiceActionInterface{
			"schema": sdk.NewAction(
				entityService.schema,
				"Get the schema of the entity",
			),
			"list": sdk.NewAction(
				entityService.list,
				"Search for available entities",
			),
			"instance": sdk.NewAction(
				entityService.instance,
				"Get the specified instance of entities",
			)}).WithCategories(
		[]sdk.EndorBaseSpecializedServiceCategoryInterface{
			NewEndorBaseSpecializedServiceCategory[*sdk.Entity](string(sdk.EntityTypeBase), "Base").
				WithActions(map[string]sdk.EndorServiceActionInterface{
					"schema": sdk.NewAction(
						entityService.schema,
						"Get the schema of the entity of type "+string(sdk.EntityTypeBase),
					),
					"instance": sdk.NewAction(
						entityService.entityBaseInstance,
						"Get the specified instance of entities of type "+string(sdk.EntityTypeBase),
					),
					"list": sdk.NewAction(
						entityService.entityBaseList,
						"Search for available entities of type "+string(sdk.EntityTypeBase),
					),
				}),
			NewEndorBaseSpecializedServiceCategory[*sdk.EntitySpecialized](string(sdk.EntityTypeBaseSpecialized), "Base specialized").
				WithActions(map[string]sdk.EndorServiceActionInterface{
					"schema": sdk.NewAction(
						entityService.entityBaseSpecializedSchema,
						"Get the schema of the entity of type "+string(sdk.EntityTypeBaseSpecialized),
					),
					"instance": sdk.NewAction(
						entityService.entityBaseSpecializedInstance,
						"Get the specified instance of entities of type "+string(sdk.EntityTypeBaseSpecialized),
					),
					"list": sdk.NewAction(
						entityService.entityBaseSpecializedList,
						"Search for available entities of type "+string(sdk.EntityTypeBaseSpecialized),
					),
				}),
			NewEndorBaseSpecializedServiceCategory[*sdk.EntityHybrid](string(sdk.EntityTypeHybrid), "Hybrid").
				WithActions(hybridCategoryActions),
			NewEndorBaseSpecializedServiceCategory[*sdk.EntityHybridSpecialized](string(sdk.EntityTypeHybridSpecialized), "Hybrid specialized").
				WithActions(hybridSpecializedCategoryActions),
			NewEndorBaseSpecializedServiceCategory[*sdk.EntityHybrid](string(sdk.EntityTypeDynamic), "Dynamic").
				WithActions(dynamicCategoryActions),
			NewEndorBaseSpecializedServiceCategory[*sdk.EntityHybridSpecialized](string(sdk.EntityTypeDynamicSpecialized), "Dynamic specialized").
				WithActions(dynamicSpecializedCategoryActions),
		},
	)
}

type EntityService struct {
	microServiceId string
	services       *[]sdk.EndorServiceInterface
}

func (h *EntityService) schema(c *sdk.EndorContext[sdk.NoPayload]) (*sdk.Response[any], error) {
	return sdk.NewResponseBuilder[any]().AddSchema(sdk.NewSchema(&sdk.Entity{})).Build(), nil
}

func (h *EntityService) entityBaseSpecializedSchema(c *sdk.EndorContext[sdk.NoPayload]) (*sdk.Response[any], error) {
	return sdk.NewResponseBuilder[any]().AddSchema(sdk.NewSchema(&sdk.EntitySpecialized{})).Build(), nil
}

func (h *EntityService) entityHybridSchema(c *sdk.EndorContext[sdk.NoPayload]) (*sdk.Response[any], error) {
	return sdk.NewResponseBuilder[any]().AddSchema(sdk.NewSchema(&sdk.EntityHybrid{})).Build(), nil
}

func (h *EntityService) entityHybridSpecializedSchema(c *sdk.EndorContext[sdk.NoPayload]) (*sdk.Response[any], error) {
	return sdk.NewResponseBuilder[any]().AddSchema(sdk.NewSchema(&sdk.EntityHybridSpecialized{})).Build(), nil
}

func (h *EntityService) list(c *sdk.EndorContext[sdk.NoPayload]) (*sdk.Response[[]sdk.EntityInterface], error) {
	entities, err := NewEndorServiceRepository(h.microServiceId, h.services).EntityList()
	if err != nil {
		return nil, err
	}
	filtered := make([]sdk.EntityInterface, 0, len(entities))
	for _, r := range entities {
		if r.GetID() != "entity" && r.GetID() != "entity-action" {
			filtered = append(filtered, r)
		}
	}
	entities = filtered
	return sdk.NewResponseBuilder[[]sdk.EntityInterface]().AddData(&entities).AddSchema(sdk.NewSchema(&sdk.Entity{})).Build(), nil
}

func (h *EntityService) entityBaseList(c *sdk.EndorContext[sdk.NoPayload]) (*sdk.Response[[]sdk.EntityInterface], error) {
	entities, err := NewEndorServiceRepository(h.microServiceId, h.services).EntityList()
	if err != nil {
		return nil, err
	}
	filtered := make([]sdk.EntityInterface, 0, len(entities))
	for _, r := range entities {
		if r.GetID() != "entity" && r.GetID() != "entity-action" && r.GetCategoryType() == string(sdk.EntityTypeBase) {
			filtered = append(filtered, r)
		}
	}
	entities = filtered
	return sdk.NewResponseBuilder[[]sdk.EntityInterface]().AddData(&entities).AddSchema(sdk.NewSchema(&sdk.Entity{})).Build(), nil
}

func (h *EntityService) entityBaseSpecializedList(c *sdk.EndorContext[sdk.NoPayload]) (*sdk.Response[[]sdk.EntityInterface], error) {
	entities, err := NewEndorServiceRepository(h.microServiceId, h.services).EntityList()
	if err != nil {
		return nil, err
	}
	filtered := make([]sdk.EntityInterface, 0, len(entities))
	for _, r := range entities {
		if r.GetID() != "entity" && r.GetID() != "entity-action" && r.GetCategoryType() == string(sdk.EntityTypeBaseSpecialized) {
			filtered = append(filtered, r)
		}
	}
	entities = filtered
	return sdk.NewResponseBuilder[[]sdk.EntityInterface]().AddData(&entities).AddSchema(sdk.NewSchema(&sdk.EntitySpecialized{})).Build(), nil
}

func (h *EntityService) entityHybridList(c *sdk.EndorContext[sdk.NoPayload]) (*sdk.Response[[]sdk.EntityInterface], error) {
	entities, err := NewEndorServiceRepository(h.microServiceId, h.services).EntityList()
	if err != nil {
		return nil, err
	}
	filtered := make([]sdk.EntityInterface, 0, len(entities))
	for _, r := range entities {
		if r.GetID() != "entity" && r.GetID() != "entity-action" && r.GetCategoryType() == string(sdk.EntityTypeHybrid) {
			filtered = append(filtered, r)
		}
	}
	entities = filtered
	return sdk.NewResponseBuilder[[]sdk.EntityInterface]().AddData(&entities).AddSchema(sdk.NewSchema(&sdk.EntityHybrid{})).Build(), nil
}

func (h *EntityService) entityHybridSpecializedList(c *sdk.EndorContext[sdk.NoPayload]) (*sdk.Response[[]sdk.EntityInterface], error) {
	entities, err := NewEndorServiceRepository(h.microServiceId, h.services).EntityList()
	if err != nil {
		return nil, err
	}
	filtered := make([]sdk.EntityInterface, 0, len(entities))
	for _, r := range entities {
		if r.GetID() != "entity" && r.GetID() != "entity-action" && r.GetCategoryType() == string(sdk.EntityTypeHybridSpecialized) {
			filtered = append(filtered, r)
		}
	}
	entities = filtered
	return sdk.NewResponseBuilder[[]sdk.EntityInterface]().AddData(&entities).AddSchema(sdk.NewSchema(&sdk.EntityHybridSpecialized{})).Build(), nil
}

func (h *EntityService) entityDynamicList(c *sdk.EndorContext[sdk.NoPayload]) (*sdk.Response[[]sdk.EntityInterface], error) {
	entities, err := NewEndorServiceRepository(h.microServiceId, h.services).EntityList()
	if err != nil {
		return nil, err
	}
	filtered := make([]sdk.EntityInterface, 0, len(entities))
	for _, r := range entities {
		if r.GetID() != "entity" && r.GetID() != "entity-action" && r.GetCategoryType() == string(sdk.EntityTypeDynamic) {
			filtered = append(filtered, r)
		}
	}
	entities = filtered
	return sdk.NewResponseBuilder[[]sdk.EntityInterface]().AddData(&entities).AddSchema(sdk.NewSchema(&sdk.EntityHybrid{})).Build(), nil
}

func (h *EntityService) entityDynamicSpecializedList(c *sdk.EndorContext[sdk.NoPayload]) (*sdk.Response[[]sdk.EntityInterface], error) {
	entities, err := NewEndorServiceRepository(h.microServiceId, h.services).EntityList()
	if err != nil {
		return nil, err
	}
	filtered := make([]sdk.EntityInterface, 0, len(entities))
	for _, r := range entities {
		if r.GetID() != "entity" && r.GetID() != "entity-action" && r.GetCategoryType() == string(sdk.EntityTypeDynamicSpecialized) {
			filtered = append(filtered, r)
		}
	}
	entities = filtered
	return sdk.NewResponseBuilder[[]sdk.EntityInterface]().AddData(&entities).AddSchema(sdk.NewSchema(&sdk.EntityHybrid{})).Build(), nil
}

func (h *EntityService) instance(c *sdk.EndorContext[sdk.ReadInstanceDTO]) (*sdk.Response[sdk.EntityInterface], error) {
	entity, err := NewEndorServiceRepository(h.microServiceId, h.services).Instance(c.Payload)
	if err != nil {
		return nil, err
	}
	return sdk.NewResponseBuilder[sdk.EntityInterface]().AddData(&entity.entity).AddSchema(sdk.NewSchema(&sdk.Entity{})).Build(), nil
}

func (h *EntityService) entityBaseInstance(c *sdk.EndorContext[sdk.ReadInstanceDTO]) (*sdk.Response[sdk.EntityInterface], error) {
	entity, err := NewEndorServiceRepository(h.microServiceId, h.services).Instance(c.Payload)
	if err != nil {
		return nil, err
	}
	if entity.entity.GetCategoryType() != string(sdk.EntityTypeBase) {
		return nil, sdk.NewNotFoundError(fmt.Errorf("entity %s of type %s not found", entity.entity.GetID(), sdk.EntityTypeBase))
	}
	return sdk.NewResponseBuilder[sdk.EntityInterface]().AddData(&entity.entity).AddSchema(sdk.NewSchema(&sdk.Entity{})).Build(), nil
}

func (h *EntityService) entityBaseSpecializedInstance(c *sdk.EndorContext[sdk.ReadInstanceDTO]) (*sdk.Response[sdk.EntityInterface], error) {
	entity, err := NewEndorServiceRepository(h.microServiceId, h.services).Instance(c.Payload)
	if err != nil {
		return nil, err
	}
	if entity.entity.GetCategoryType() != string(sdk.EntityTypeBaseSpecialized) {
		return nil, sdk.NewNotFoundError(fmt.Errorf("entity %s of type %s not found", entity.entity.GetID(), sdk.EntityTypeBaseSpecialized))
	}
	return sdk.NewResponseBuilder[sdk.EntityInterface]().AddData(&entity.entity).AddSchema(sdk.NewSchema(&sdk.EntitySpecialized{})).Build(), nil
}

func (h *EntityService) entityHybridInstance(c *sdk.EndorContext[sdk.ReadInstanceDTO]) (*sdk.Response[sdk.EntityInterface], error) {
	entity, err := NewEndorServiceRepository(h.microServiceId, h.services).Instance(c.Payload)
	if err != nil {
		return nil, err
	}
	if entity.entity.GetCategoryType() != string(sdk.EntityTypeHybrid) {
		return nil, sdk.NewNotFoundError(fmt.Errorf("entity %s of type %s not found", entity.entity.GetID(), sdk.EntityTypeHybrid))
	}
	return sdk.NewResponseBuilder[sdk.EntityInterface]().AddData(&entity.entity).AddSchema(sdk.NewSchema(&sdk.EntityHybrid{})).Build(), nil
}

func (h *EntityService) entityHybridSpecializedInstance(c *sdk.EndorContext[sdk.ReadInstanceDTO]) (*sdk.Response[sdk.EntityInterface], error) {
	entity, err := NewEndorServiceRepository(h.microServiceId, h.services).Instance(c.Payload)
	if err != nil {
		return nil, err
	}
	if entity.entity.GetCategoryType() != string(sdk.EntityTypeHybridSpecialized) {
		return nil, sdk.NewNotFoundError(fmt.Errorf("entity %s of type %s not found", entity.entity.GetID(), sdk.EntityTypeHybridSpecialized))
	}
	return sdk.NewResponseBuilder[sdk.EntityInterface]().AddData(&entity.entity).AddSchema(sdk.NewSchema(&sdk.EntityHybridSpecialized{})).Build(), nil
}

func (h *EntityService) entityDynamicInstance(c *sdk.EndorContext[sdk.ReadInstanceDTO]) (*sdk.Response[sdk.EntityInterface], error) {
	entity, err := NewEndorServiceRepository(h.microServiceId, h.services).Instance(c.Payload)
	if err != nil {
		return nil, err
	}
	if entity.entity.GetCategoryType() != string(sdk.EntityTypeDynamic) {
		return nil, sdk.NewNotFoundError(fmt.Errorf("entity %s of type %s not found", entity.entity.GetID(), sdk.EntityTypeDynamic))
	}
	return sdk.NewResponseBuilder[sdk.EntityInterface]().AddData(&entity.entity).AddSchema(sdk.NewSchema(&sdk.EntityHybrid{})).Build(), nil
}

func (h *EntityService) entityDynamicSpecializedInstance(c *sdk.EndorContext[sdk.ReadInstanceDTO]) (*sdk.Response[sdk.EntityInterface], error) {
	entity, err := NewEndorServiceRepository(h.microServiceId, h.services).Instance(c.Payload)
	if err != nil {
		return nil, err
	}
	if entity.entity.GetCategoryType() != string(sdk.EntityTypeDynamicSpecialized) {
		return nil, sdk.NewNotFoundError(fmt.Errorf("entity %s of type %s not found", entity.entity.GetID(), sdk.EntityTypeDynamicSpecialized))
	}
	return sdk.NewResponseBuilder[sdk.EntityInterface]().AddData(&entity.entity).AddSchema(sdk.NewSchema(&sdk.EntityHybrid{})).Build(), nil
}

func (h *EntityService) entityDynamicCreate(c *sdk.EndorContext[sdk.CreateDTO[sdk.EntityInterface]]) (*sdk.Response[sdk.EntityInterface], error) {
	// force type
	c.Payload.Data.SetCategoryType(string(sdk.EntityTypeDynamic))
	err := NewEndorServiceRepository(h.microServiceId, h.services).Create(c.Payload)
	if err != nil {
		return nil, err
	}
	return sdk.NewResponseBuilder[sdk.EntityInterface]().AddData(&c.Payload.Data).AddSchema(sdk.NewSchema(&sdk.EntityHybrid{})).AddMessage(sdk.NewMessage(sdk.ResponseMessageGravityInfo, fmt.Sprintf("entity %s created", c.Payload.Data.GetID()))).Build(), nil
}

func (h *EntityService) entityDynamicSpecializedCreate(c *sdk.EndorContext[sdk.CreateDTO[sdk.EntityInterface]]) (*sdk.Response[sdk.EntityInterface], error) {
	// force type
	c.Payload.Data.SetCategoryType(string(sdk.EntityTypeDynamicSpecialized))
	err := NewEndorServiceRepository(h.microServiceId, h.services).Create(c.Payload)
	if err != nil {
		return nil, err
	}
	return sdk.NewResponseBuilder[sdk.EntityInterface]().AddData(&c.Payload.Data).AddSchema(sdk.NewSchema(&sdk.EntityHybridSpecialized{})).AddMessage(sdk.NewMessage(sdk.ResponseMessageGravityInfo, fmt.Sprintf("entity %s created", c.Payload.Data.GetID()))).Build(), nil
}

func (h *EntityService) entityHybridUpdate(c *sdk.EndorContext[sdk.UpdateByIdDTO[sdk.EntityHybrid]]) (*sdk.Response[sdk.EntityInterface], error) {
	entity, err := NewEndorServiceRepository(h.microServiceId, h.services).Update(c.Payload.Id, &c.Payload.Data)
	if err != nil {
		return nil, err
	}
	return sdk.NewResponseBuilder[sdk.EntityInterface]().AddData(entity).AddSchema(sdk.NewSchema(&sdk.EntityHybrid{})).AddMessage(sdk.NewMessage(sdk.ResponseMessageGravityInfo, "entity updated")).Build(), nil
}

func (h *EntityService) entityHybridSpecializedUpdate(c *sdk.EndorContext[sdk.UpdateByIdDTO[sdk.EntityHybridSpecialized]]) (*sdk.Response[sdk.EntityInterface], error) {
	entity, err := NewEndorServiceRepository(h.microServiceId, h.services).Update(c.Payload.Id, &c.Payload.Data)
	if err != nil {
		return nil, err
	}
	return sdk.NewResponseBuilder[sdk.EntityInterface]().AddData(entity).AddSchema(sdk.NewSchema(&sdk.EntityHybridSpecialized{})).AddMessage(sdk.NewMessage(sdk.ResponseMessageGravityInfo, "entity updated")).Build(), nil
}

func (h *EntityService) entityDynamicUpdate(c *sdk.EndorContext[sdk.UpdateByIdDTO[sdk.EntityInterface]]) (*sdk.Response[sdk.EntityInterface], error) {
	entity, err := NewEndorServiceRepository(h.microServiceId, h.services).Update(c.Payload.Id, c.Payload.Data)
	if err != nil {
		return nil, err
	}
	return sdk.NewResponseBuilder[sdk.EntityInterface]().AddData(entity).AddSchema(sdk.NewSchema(&sdk.EntityHybrid{})).AddMessage(sdk.NewMessage(sdk.ResponseMessageGravityInfo, "entity updated")).Build(), nil
}

func (h *EntityService) entityDynamicSpecializedUpdate(c *sdk.EndorContext[sdk.UpdateByIdDTO[sdk.EntityInterface]]) (*sdk.Response[sdk.EntityInterface], error) {
	entity, err := NewEndorServiceRepository(h.microServiceId, h.services).Update(c.Payload.Id, c.Payload.Data)
	if err != nil {
		return nil, err
	}
	return sdk.NewResponseBuilder[sdk.EntityInterface]().AddData(entity).AddSchema(sdk.NewSchema(&sdk.EntityHybridSpecialized{})).AddMessage(sdk.NewMessage(sdk.ResponseMessageGravityInfo, "entity updated")).Build(), nil
}

func (h *EntityService) entityDynamicDelete(c *sdk.EndorContext[sdk.ReadInstanceDTO]) (*sdk.Response[sdk.Entity], error) {
	err := NewEndorServiceRepository(h.microServiceId, h.services).Delete(c.Payload)
	if err != nil {
		return nil, err
	}
	return sdk.NewResponseBuilder[sdk.Entity]().AddMessage(sdk.NewMessage(sdk.ResponseMessageGravityInfo, fmt.Sprintf("entity %s deleted", c.Payload.Id))).Build(), nil
}
