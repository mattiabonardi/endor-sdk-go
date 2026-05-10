package sdk_entity

import (
	"fmt"
	"io/fs"
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
func InitEndorHandlerRepository(microServiceId string, module string, internalEndorHandlers *[]sdk.EndorHandlerInterface, logger *sdk.Logger, projectLocalesFS fs.FS) *EndorHandlerRepository {
	endorHandlerRepositoryOnce.Do(func() {
		core := InitRegistryCore(microServiceId, module, internalEndorHandlers, logger, projectLocalesFS)
		endorHandlerRepositoryInstance = &EndorHandlerRepository{core: core}
	})
	return endorHandlerRepositoryInstance
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
	entityEntityID := path.Join(h.core.module, "entity")
	entityActionEntityID := path.Join(h.core.module, "entity-action")
	aggregatonID := path.Join(h.core.module, "aggregation")
	entityList := make([]sdk.EntityInterface, 0, len(dict))
	for _, v := range dict {
		entityList = append(entityList, v.entity)
	}
	// filter by entity type
	filtered := make([]sdk.EntityInterface, 0, len(dict))
	for _, r := range entityList {
		if r.GetID() != entityEntityID && r.GetID() != entityActionEntityID && r.GetID() != aggregatonID {
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
	return nil, sdk.NewNotFoundError(fmt.Errorf("entity %s not found", dto.Id)).WithTranslation("sdk.entity.messages.not_found", map[string]any{"id": dto.Id})
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
