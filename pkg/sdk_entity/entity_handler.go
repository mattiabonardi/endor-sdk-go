package sdk_entity

import (
	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
)

func NewEntityHandler(microServiceId string, module string, handlers *[]sdk.EndorHandlerInterface, repository *sdk.EntityRepositoryInterface, logger *sdk.Logger, priority int) sdk.EndorHandlerInterface {
	// Resolve the repository lazily at action-call time so that InitEndorHandlerRepository
	// (called later in server.Init with the correct projectLocalesFS) is always the first
	// initializer and the nil-FS fallback path is never triggered.
	repoAccessor := func() sdk.EntityRepositoryInterface {
		if repository != nil {
			return *repository
		}
		if r := GetEndorHandlerRepository(); r != nil {
			return r
		}
		return nil
	}
	entityService := EntityHandler{
		handlers:     handlers,
		repoAccessor: repoAccessor,
	}

	schema := sdk.NewSchema(&sdk.Entity{})
	return NewEndorBaseHandler[*sdk.Entity]("entity", "t(sdk.entity.handler.title)").
		WithPriority(priority).
		WithActions(map[string]sdk.EndorHandlerActionInterface{
			"schema": sdk.NewAction(
				entityService.schema(schema),
				"t(sdk.entity.handler.actions.schema)",
			),
			"instance": sdk.NewAction(
				entityService.instance("", schema),
				"t(sdk.entity.handler.actions.instance)",
			),
			"list": sdk.NewAction(
				entityService.list("", schema),
				"t(sdk.entity.handler.actions.list)",
			),
		})
}

type EntityHandler struct {
	handlers     *[]sdk.EndorHandlerInterface
	repoAccessor func() sdk.EntityRepositoryInterface
}

func resolveEntityTranslations(resolveExpr func(string) string, entity sdk.EntityInterface) sdk.EntityInterface {
	e, ok := entity.(*sdk.Entity)
	if !ok {
		return entity
	}
	copy := *e
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
	return &copy
}

func (h *EntityHandler) schema(schema *sdk.RootSchema) func(c *sdk.EndorContext[sdk.NoPayload]) (*sdk.Response[any], error) {
	return func(c *sdk.EndorContext[sdk.NoPayload]) (*sdk.Response[any], error) {
		return sdk.NewResponseBuilder[any]().AddSchema(schema).Build(), nil
	}
}

func (h *EntityHandler) list(entityType sdk.EntityType, schema *sdk.RootSchema) func(c *sdk.EndorContext[sdk.NoPayload]) (*sdk.Response[[]sdk.EntityInterface], error) {
	return func(c *sdk.EndorContext[sdk.NoPayload]) (*sdk.Response[[]sdk.EntityInterface], error) {
		entities, err := h.repoAccessor().List(c.Session, &entityType)
		if err != nil {
			return nil, err
		}
		resolved := make([]sdk.EntityInterface, len(entities))
		for i, entity := range entities {
			resolved[i] = resolveEntityTranslations(c.ResolveTExpr, entity)
		}
		return sdk.NewResponseBuilder[[]sdk.EntityInterface]().AddData(&resolved).AddSchema(schema).Build(), nil
	}
}

func (h *EntityHandler) instance(entityType sdk.EntityType, schema *sdk.RootSchema) func(c *sdk.EndorContext[sdk.ReadInstanceDTO]) (*sdk.Response[sdk.EntityInterface], error) {
	return func(c *sdk.EndorContext[sdk.ReadInstanceDTO]) (*sdk.Response[sdk.EntityInterface], error) {
		entity, err := h.repoAccessor().Instance(c.Session, &entityType, c.Payload)
		if err != nil {
			return nil, err
		}
		var resolved sdk.EntityInterface
		if entity != nil {
			resolved = resolveEntityTranslations(c.ResolveTExpr, *entity)
		}
		return sdk.NewResponseBuilder[sdk.EntityInterface]().AddData(&resolved).AddSchema(schema).Build(), nil
	}
}
