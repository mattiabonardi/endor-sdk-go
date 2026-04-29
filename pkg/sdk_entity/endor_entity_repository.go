package sdk_entity

import (
	"fmt"
	"path"
	"sync"

	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
)

// Singleton instance and initialization sync
var (
	endorHandlerRepositoryInstance *EndorHandlerRepository
	endorHandlerRepositoryOnce     sync.Once
)

// GetEndorHandlerRepository returns the singleton EndorHandlerRepository instance.
func GetEndorHandlerRepository() *EndorHandlerRepository {
	return endorHandlerRepositoryInstance
}

// InitEndorHandlerRepository initializes the singleton EndorHandlerRepository.
// It also initializes the underlying RegistryCore engine if not yet done.
func InitEndorHandlerRepository(module string, version string, internalEndorHandlers *[]sdk.EndorHandlerInterface, logger *sdk.Logger) *EndorHandlerRepository {
	endorHandlerRepositoryOnce.Do(func() {
		core := InitRegistryCore(module, version, internalEndorHandlers, logger)
		endorHandlerRepositoryInstance = &EndorHandlerRepository{core: core}
	})
	return endorHandlerRepositoryInstance
}

// NewEndorHandlerRepository returns the singleton EndorHandlerRepository instance.
// Deprecated: Use InitEndorHandlerRepository for explicit initialization or GetEndorHandlerRepository to retrieve it.
func NewEndorHandlerRepository(module string, version string, internalEndorHandlers *[]sdk.EndorHandlerInterface, logger *sdk.Logger) *EndorHandlerRepository {
	return InitEndorHandlerRepository(module, version, internalEndorHandlers, logger)
}

// EndorHandlerRepository provides entity-level access to the handler registry.
// It delegates all resolution to RegistryCore, passing the session so the core
// can transparently choose the production or ephemeral dev dictionary.
type EndorHandlerRepository struct {
	core *RegistryCore
}

// #region Entity CRUD

func (h *EndorHandlerRepository) List(session sdk.Session, entityType *sdk.EntityType) ([]sdk.EntityInterface, error) {
	dict, err := h.core.Dictionary(session)
	if err != nil {
		return []sdk.EntityInterface{}, err
	}
	entityEntityID := path.Join(h.core.module, h.core.version, "entity")
	entityActionEntityID := path.Join(h.core.module, h.core.version, "entity-action")
	entityList := make([]sdk.EntityInterface, 0, len(dict))
	for _, v := range dict {
		entityList = append(entityList, v.entity)
	}
	// filter by entity type
	filtered := make([]sdk.EntityInterface, 0, len(dict))
	for _, r := range entityList {
		if r.GetID() != entityEntityID && r.GetID() != entityActionEntityID {
			if r.GetCategoryType() == string(*entityType) {
				filtered = append(filtered, r)
			} else {
				if entityType == nil || *entityType == "" {
					filtered = append(filtered, r)
				}
			}
		}
	}
	return filtered, nil
}

func (h *EndorHandlerRepository) Instance(session sdk.Session, entityType *sdk.EntityType, dto sdk.ReadInstanceDTO) (*sdk.EntityInterface, error) {
	entry, err := h.core.DictionaryInstance(session, dto)
	if err != nil {
		return nil, err
	}
	if entityType == nil || *entityType == "" {
		return &entry.entity, nil
	}
	if entry.entity.GetCategoryType() == string(*entityType) {
		return &entry.entity, nil
	}
	return nil, sdk.NewNotFoundError(fmt.Errorf("entity %s not found", dto.Id)).WithTranslation("entities.entity.not_found", map[string]any{"id": dto.Id})
}

// EndorHandlerList returns all registered EndorHandlers.
// Used by the server to register routes and swagger configuration.
func (h *EndorHandlerRepository) EndorHandlerList() ([]sdk.EndorHandler, error) {
	return h.core.endorHandlerList()
}

// ActionRepository returns an EndorHandlerActionRepository backed by the same core.
func (h *EndorHandlerRepository) ActionRepository() *EndorHandlerActionRepository {
	return NewEndorHandlerActionRepository(h.core)
}

// #endregion
