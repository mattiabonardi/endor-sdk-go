package sdk_entity

import (
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
)

type EndorBaseHandler[T sdk.EntityInstanceInterface] struct {
	entity            string
	entityTitle       string
	entityDescription string
	priority          *int
	actions           map[string]sdk.EndorHandlerActionInterface
	repositoryFactory sdk.RepositoryFactory
}

func (h EndorBaseHandler[T]) GetEntity() string {
	return h.entity
}

func (h EndorBaseHandler[T]) GetEntityTitle() string {
	return h.entityTitle
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

func NewEndorBaseHandler[T sdk.EntityInstanceInterface](entity, entityTitle string) sdk.EndorBaseHandlerInterface {
	return EndorBaseHandler[T]{
		entity:      entity,
		entityTitle: entityTitle,
	}
}

func (h EndorBaseHandler[T]) WithExtendedDescription(
	description string,
) sdk.EndorBaseHandlerInterface {
	h.entityDescription = description
	return h
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

func (h EndorBaseHandler[T]) WithRepository(
	fn sdk.RepositoryFactory,
) sdk.EndorBaseHandlerInterface {
	h.repositoryFactory = fn
	return h
}

func (h EndorBaseHandler[T]) ToEndorHandler() sdk.EndorHandler {
	var baseModel T
	rootSchema := sdk.NewSchema(baseModel)
	return sdk.EndorHandler{
		Entity:              h.entity,
		EntityTitle:         h.entityTitle,
		EntityDescription:   h.entityDescription,
		Priority:            h.priority,
		Actions:             h.actions,
		EntitySchema:        *rootSchema,
		RepositoryFactories: map[string]sdk.RepositoryFactory{h.entity: h.repositoryFactory},
	}
}

func getRootSchema[T sdk.EntityInstanceInterface]() *sdk.RootSchema {
	var baseModel T
	return sdk.NewSchema(baseModel)
}
