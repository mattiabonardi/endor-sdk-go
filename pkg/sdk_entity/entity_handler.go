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

	return NewEndorBaseSpecializedHandler[*sdk.Entity]("entity", "t(sdk.entity.handler.title)").
		WithPriority(priority).
		WithActions(map[string]sdk.EndorHandlerActionInterface{
			"schema": sdk.NewAction(
				entityService.schema(sdk.NewSchema(&sdk.Entity{})),
				"t(sdk.entity.handler.actions.schema)",
			),
			"instance": sdk.NewAction(
				entityService.instance("", sdk.NewSchema(&sdk.Entity{})),
				"t(sdk.entity.handler.actions.instance)",
			),
			"list": sdk.NewAction(
				entityService.list("", sdk.NewSchema(&sdk.Entity{})),
				"t(sdk.entity.handler.actions.list)",
			)}).WithCategories(
		[]sdk.EndorBaseSpecializedHandlerCategoryInterface{
			NewEndorBaseSpecializedHandlerCategory[*sdk.Entity](string(sdk.EntityTypeBase), "t(sdk.entity.handler.categories.base.title)").
				WithActions(map[string]sdk.EndorHandlerActionInterface{
					"schema": sdk.NewAction(
						entityService.schema(sdk.NewSchema(&sdk.Entity{})),
						"t(sdk.entity.handler.actions.schema) "+string(sdk.EntityTypeBase),
					),
					"instance": sdk.NewAction(
						entityService.instance(sdk.EntityTypeBase, sdk.NewSchema(&sdk.Entity{})),
						"t(sdk.entity.handler.actions.instance) "+string(sdk.EntityTypeBase),
					),
					"list": sdk.NewAction(
						entityService.list(sdk.EntityTypeBase, sdk.NewSchema(&sdk.Entity{})),
						"t(sdk.entity.handler.actions.list) "+string(sdk.EntityTypeBase),
					),
				}),
			NewEndorBaseSpecializedHandlerCategory[*sdk.EntitySpecialized](string(sdk.EntityTypeBaseSpecialized), "t(sdk.entity.handler.categories.base_specialized.title)").
				WithActions(map[string]sdk.EndorHandlerActionInterface{
					"schema": sdk.NewAction(
						entityService.schema(sdk.NewSchema(&sdk.EntitySpecialized{})),
						"t(sdk.entity.handler.actions.schema) "+string(sdk.EntityTypeBaseSpecialized),
					),
					"instance": sdk.NewAction(
						entityService.instance(sdk.EntityTypeBaseSpecialized, sdk.NewSchema(&sdk.EntitySpecialized{})),
						"t(sdk.entity.handler.actions.instance) "+string(sdk.EntityTypeBaseSpecialized),
					),
					"list": sdk.NewAction(
						entityService.list(sdk.EntityTypeBaseSpecialized, sdk.NewSchema(&sdk.EntitySpecialized{})),
						"t(sdk.entity.handler.actions.list) "+string(sdk.EntityTypeBaseSpecialized),
					),
				}),
			NewEndorBaseSpecializedHandlerCategory[*sdk.EntityHybrid](string(sdk.EntityTypeHybrid), "t(sdk.entity.handler.categories.hybrid.title)").
				WithActions(map[string]sdk.EndorHandlerActionInterface{
					"schema": sdk.NewAction(
						entityService.schema(sdk.NewSchema(&sdk.EntityHybrid{})),
						"t(sdk.entity.handler.actions.schema) "+string(sdk.EntityTypeHybrid),
					),
					"instance": sdk.NewAction(
						entityService.instance(sdk.EntityTypeHybrid, sdk.NewSchema(&sdk.EntityHybrid{})),
						"t(sdk.entity.handler.actions.instance) "+string(sdk.EntityTypeHybrid),
					),
					"list": sdk.NewAction(
						entityService.list(sdk.EntityTypeHybrid, sdk.NewSchema(&sdk.EntityHybrid{})),
						"t(sdk.entity.handler.actions.list) "+string(sdk.EntityTypeHybrid),
					),
				}),
			NewEndorBaseSpecializedHandlerCategory[*sdk.EntityHybridSpecialized](string(sdk.EntityTypeHybridSpecialized), "t(sdk.entity.handler.categories.hybrid_specialized.title)").
				WithActions(map[string]sdk.EndorHandlerActionInterface{
					"schema": sdk.NewAction(
						entityService.schema(sdk.NewSchema(&sdk.EntityHybridSpecialized{})),
						"t(sdk.entity.handler.actions.schema) "+string(sdk.EntityTypeHybridSpecialized),
					),
					"instance": sdk.NewAction(
						entityService.instance(sdk.EntityTypeHybridSpecialized, sdk.NewSchema(&sdk.EntityHybridSpecialized{})),
						"t(sdk.entity.handler.actions.instance) "+string(sdk.EntityTypeHybridSpecialized),
					),
					"list": sdk.NewAction(
						entityService.list(sdk.EntityTypeHybridSpecialized, sdk.NewSchema(&sdk.EntityHybridSpecialized{})),
						"t(sdk.entity.handler.actions.list) "+string(sdk.EntityTypeHybridSpecialized),
					),
				}),
			NewEndorBaseSpecializedHandlerCategory[*sdk.EntityHybrid](string(sdk.EntityTypeDynamic), "t(sdk.entity.handler.categories.dynamic.title)").
				WithActions(map[string]sdk.EndorHandlerActionInterface{
					"schema": sdk.NewAction(
						entityService.schema(sdk.NewSchema(&sdk.EntityHybrid{})),
						"t(sdk.entity.handler.actions.schema) "+string(sdk.EntityTypeDynamic)),
					"instance": sdk.NewAction(
						entityService.instance(sdk.EntityTypeDynamic, sdk.NewSchema(&sdk.EntityHybrid{})),
						"t(sdk.entity.handler.actions.instance) "+string(sdk.EntityTypeDynamic)),
					"list": sdk.NewAction(
						entityService.list(sdk.EntityTypeDynamic, sdk.NewSchema(&sdk.EntityHybrid{})),
						"t(sdk.entity.handler.actions.list) "+string(sdk.EntityTypeDynamic)),
				}),
			NewEndorBaseSpecializedHandlerCategory[*sdk.EntityHybridSpecialized](string(sdk.EntityTypeDynamicSpecialized), "t(sdk.entity.handler.categories.dynamic_specialized.title)").
				WithActions(map[string]sdk.EndorHandlerActionInterface{
					"schema": sdk.NewAction(
						entityService.schema(sdk.NewSchema(&sdk.EntityHybridSpecialized{})),
						"t(sdk.entity.handler.actions.schema) "+string(sdk.EntityTypeDynamicSpecialized)),
					"instance": sdk.NewAction(
						entityService.instance(sdk.EntityTypeDynamicSpecialized, sdk.NewSchema(&sdk.EntityHybridSpecialized{})),
						"t(sdk.entity.handler.actions.instance) "+string(sdk.EntityTypeDynamicSpecialized)),
					"list": sdk.NewAction(
						entityService.list(sdk.EntityTypeDynamicSpecialized, sdk.NewSchema(&sdk.EntityHybridSpecialized{})),
						"t(sdk.entity.handler.actions.list) "+string(sdk.EntityTypeDynamicSpecialized)),
				}),
		},
	)
}

type EntityHandler struct {
	handlers     *[]sdk.EndorHandlerInterface
	repoAccessor func() sdk.EntityRepositoryInterface
}

func resolveEntityTranslations(resolveExpr func(string) string, entity sdk.EntityInterface) sdk.EntityInterface {
	switch e := entity.(type) {
	case *sdk.Entity:
		copy := *e
		copy.Title = resolveExpr(e.Title)
		copy.Description = resolveExpr(e.Description)
		copy.Schema = resolveExpr(e.Schema)
		return &copy
	case *sdk.EntitySpecialized:
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
	case *sdk.EntityHybrid:
		copy := *e
		copy.Title = resolveExpr(e.Title)
		copy.Description = resolveExpr(e.Description)
		copy.Schema = resolveExpr(e.Schema)
		copy.AdditionalSchema = resolveExpr(e.AdditionalSchema)
		return &copy
	case *sdk.EntityHybridSpecialized:
		copy := *e
		copy.Title = resolveExpr(e.Title)
		copy.Description = resolveExpr(e.Description)
		copy.Schema = resolveExpr(e.Schema)
		copy.AdditionalSchema = resolveExpr(e.AdditionalSchema)
		resolvedCats := make([]sdk.HybridCategory, len(e.Categories))
		for i, cat := range e.Categories {
			cat.Title = resolveExpr(cat.Title)
			cat.Description = resolveExpr(cat.Description)
			cat.Schema = resolveExpr(cat.Schema)
			cat.AdditionalSchema = resolveExpr(cat.AdditionalSchema)
			resolvedCats[i] = cat
		}
		copy.Categories = resolvedCats
		return &copy
	}
	return entity
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
