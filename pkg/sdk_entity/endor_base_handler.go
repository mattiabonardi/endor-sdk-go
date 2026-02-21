package sdk_entity

import (
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
)

type EndorBaseHandler[T sdk.EntityInstanceInterface] struct {
	entity            string
	entityDescription string
	priority          *int
	actions           map[string]sdk.EndorHandlerActionInterface
}

func (h EndorBaseHandler[T]) GetEntity() string {
	return h.entity
}

func (h EndorBaseHandler[T]) GetEntityDescription() string {
	return h.entityDescription
}

func (h EndorBaseHandler[T]) GetPriority() *int {
	return h.priority
}

func (h EndorBaseHandler[T]) GetSchema() *sdk.RootSchema {
	return getRootSchema[T]()
}

func NewEndorBaseHandler[T sdk.EntityInstanceInterface](entity, entityDescription string) sdk.EndorBaseHandlerInterface {
	return EndorBaseHandler[T]{
		entity:            entity,
		entityDescription: entityDescription,
	}
}

func (h EndorBaseHandler[T]) WithPriority(
	priority int,
) sdk.EndorBaseHandlerInterface {
	h.priority = &priority
	return h
}

func (h EndorBaseHandler[T]) WithActions(
	actions map[string]sdk.EndorHandlerActionInterface,
) sdk.EndorBaseHandlerInterface {
	h.actions = actions
	return h
}

func (h EndorBaseHandler[T]) ToEndorHandler() sdk.EndorHandler {
	var baseModel T
	rootSchema := sdk.NewSchema(baseModel)
	return sdk.EndorHandler{
		Entity:            h.entity,
		EntityDescription: h.entityDescription,
		Priority:          h.priority,
		Actions:           h.actions,
		EntitySchema:      *rootSchema,
	}
}

func getRootSchema[T sdk.EntityInstanceInterface]() *sdk.RootSchema {
	var baseModel T
	return sdk.NewSchema(baseModel)
}
