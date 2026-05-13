package sdk_entity

import (
	"fmt"

	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
)

func NewEntityHandler(microServiceId string, module string, handlers *[]sdk.EndorHandlerInterface, logger *sdk.Logger, priority int, repository func(session sdk.Session) sdk.EntityRepositoryInterface) sdk.EndorHandlerInterface {
	entityService := EntityHandler{
		handlers: handlers,
	}

	return NewEndorBaseHandler[*sdk.Entity]("entity", "t(sdk.entity.handler.title)").
		WithPriority(priority).
		WithRepository(func(session sdk.Session, container sdk.EndorDIContainerInterface) sdk.EndorRepositoryInterface {
			if repository != nil {
				return repository(session)
			}
			return &EndorEntityRepository{session: session}
		}).
		WithActions(map[string]sdk.EndorHandlerActionInterface{
			"schema": sdk.NewAction(
				entityService.schema,
				"t(sdk.entity.handler.actions.schema)",
			),
			"instance": sdk.NewAction(
				entityService.instance,
				"t(sdk.entity.handler.actions.instance)",
			),
			"list": sdk.NewAction(
				entityService.list,
				"t(sdk.entity.handler.actions.list)",
			),
		})
}

type EntityHandler struct {
	handlers *[]sdk.EndorHandlerInterface
}

func resolveEntityTranslations(resolveExpr func(string) string, e sdk.Entity) sdk.Entity {
	copy := e
	copy.Title = resolveExpr(e.Title)
	copy.Description = resolveExpr(e.Description)
	copy.Schema = resolveExpr(e.Schema)
	resolvedCats := make([]sdk.Category, len(e.Categories))
	for i, cat := range e.Categories {
		cat.Title = resolveExpr(cat.Title)
		cat.Description = resolveExpr(cat.Description)
		cat.Schema = resolveExpr(cat.Schema)
		resolvedCats[i] = cat
	}
	copy.Categories = resolvedCats
	return copy
}

func (h *EntityHandler) schema(c *sdk.EndorContext[sdk.NoPayload]) (*sdk.Response[any], error) {
	return sdk.NewResponseBuilder[any]().AddSchema(sdk.NewSchema(sdk.Entity{})).Build(), nil
}

func (h *EntityHandler) list(c *sdk.EndorContext[sdk.NoPayload]) (*sdk.Response[[]sdk.Entity], error) {
	repo, ok := c.DIContainer.GetRepositories()["entity"].(sdk.EntityRepositoryInterface)
	if !ok {
		return nil, sdk.NewInternalServerError(fmt.Errorf("entity repository not available"))
	}
	entities, err := repo.List()
	if err != nil {
		return nil, err
	}
	resolved := make([]sdk.Entity, len(entities))
	for i, entity := range entities {
		resolved[i] = resolveEntityTranslations(c.ResolveTExpr, entity)
	}
	return sdk.NewResponseBuilder[[]sdk.Entity]().AddData(&resolved).AddSchema(sdk.NewSchema(sdk.Entity{})).Build(), nil
}

func (h *EntityHandler) instance(c *sdk.EndorContext[sdk.ReadInstanceDTO]) (*sdk.Response[sdk.Entity], error) {
	repo, ok := c.DIContainer.GetRepositories()["entity"].(sdk.EntityRepositoryInterface)
	if !ok {
		return nil, sdk.NewInternalServerError(fmt.Errorf("entity repository not available"))
	}
	entity, err := repo.Instance(c.Payload)
	if err != nil {
		return nil, err
	}
	var resolved sdk.Entity
	if entity != nil {
		resolved = resolveEntityTranslations(c.ResolveTExpr, *entity)
	}
	return sdk.NewResponseBuilder[sdk.Entity]().AddData(&resolved).AddSchema(sdk.NewSchema(sdk.Entity{})).Build(), nil
}
