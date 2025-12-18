package sdk_entity

import (
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
)

type EndorBaseService[T sdk.EntityInstanceInterface] struct {
	entity            string
	entityDescription string
	priority          *int
	actions           map[string]sdk.EndorServiceActionInterface
}

func (h EndorBaseService[T]) GetEntity() string {
	return h.entity
}

func (h EndorBaseService[T]) GetEntityDescription() string {
	return h.entityDescription
}

func (h EndorBaseService[T]) GetPriority() *int {
	return h.priority
}

func NewEndorBaseService[T sdk.EntityInstanceInterface](entity, entityDescription string) sdk.EndorBaseServiceInterface {
	return EndorBaseService[T]{
		entity:            entity,
		entityDescription: entityDescription,
	}
}

func (h EndorBaseService[T]) WithPriority(
	priority int,
) sdk.EndorBaseServiceInterface {
	h.priority = &priority
	return h
}

func (h EndorBaseService[T]) WithActions(
	actions map[string]sdk.EndorServiceActionInterface,
) sdk.EndorBaseServiceInterface {
	h.actions = actions
	return h
}

func (h EndorBaseService[T]) ToEndorService() sdk.EndorService {
	var baseModel T
	rootSchema := sdk.NewSchema(baseModel)
	return sdk.EndorService{
		Entity:            h.entity,
		EntityDescription: h.entityDescription,
		Priority:          h.priority,
		Actions:           h.actions,
		EntitySchema:      *rootSchema,
	}
}
