package sdk_entity

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"path"

	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
)

// InitEndorEntityRepository initializes the RegistryCore singleton and returns
// a production-scoped EndorHandlerRepository (empty session).
func InitEndorEntityRepository(microServiceId string, module string, internalEndorHandlers *[]sdk.EndorHandlerInterface, logger *sdk.Logger, projectLocalesFS fs.FS) *EndorHandlerRepository {
	InitRegistryCore(microServiceId, module, internalEndorHandlers, logger, projectLocalesFS)
	return &EndorHandlerRepository{}
}

// GetEndorHandlerRepository returns a production-scoped EndorHandlerRepository (empty session).
func GetEndorHandlerRepository() *EndorHandlerRepository {
	return &EndorHandlerRepository{}
}

// EndorHandlerRepository is a lightweight per-request struct.
// session is baked in at construction; all methods access the RegistryCore singleton via GetRegistryCore().
type EndorHandlerRepository struct {
	session sdk.Session
}

// #region Entity CRUD

func (h *EndorHandlerRepository) List(entityType *sdk.EntityType) ([]sdk.Entity, error) {
	core := GetRegistryCore()
	dict, err := core.Dictionary(h.session)
	if err != nil {
		return []sdk.Entity{}, err
	}
	entityEntityID := path.Join(core.module, "entity")
	entityActionEntityID := path.Join(core.module, "entity-action")
	aggregatonID := path.Join(core.module, "aggregation")
	entityList := make([]sdk.Entity, 0, len(dict))
	for _, v := range dict {
		entityList = append(entityList, v.entity)
	}
	// filter by entity type
	filtered := make([]sdk.Entity, 0, len(dict))
	for _, r := range entityList {
		if r.ID != entityEntityID && r.ID != entityActionEntityID && r.ID != aggregatonID {
			if r.Type == string(*entityType) {
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

func (h *EndorHandlerRepository) Instance(entityType *sdk.EntityType, dto sdk.ReadInstanceDTO) (*sdk.Entity, error) {
	core := GetRegistryCore()
	entry, err := core.DictionaryInstance(h.session, dto)
	if err != nil {
		return nil, err
	}
	if entityType == nil || *entityType == "" {
		return &entry.entity, nil
	}
	if entry.entity.Type == string(*entityType) {
		return &entry.entity, nil
	}
	return nil, sdk.NewNotFoundError(fmt.Errorf("entity %s not found", dto.Id)).WithTranslation("sdk.entity.messages.not_found", map[string]any{"id": dto.Id})
}

func (h *EndorHandlerRepository) GetEntity() string {
	return "entity"
}

func (h *EndorHandlerRepository) GetSchema() *sdk.RootSchema {
	return sdk.NewSchema(&sdk.Entity{})
}

func (h *EndorHandlerRepository) FindReferences(_ context.Context, ids sdk.ReadInstancesDTO) (sdk.EntityReferenceGroupDescriptions, error) {
	core := GetRegistryCore()
	result := make(sdk.EntityReferenceGroupDescriptions, len(ids.Ids))
	for _, id := range ids.Ids {
		entry, err := core.DictionaryInstance(h.session, sdk.ReadInstanceDTO{Id: id})
		if err == nil {
			result[id] = entry.entity.Title
		}
	}
	return result, nil
}

func (h *EndorHandlerRepository) RawList(_ context.Context, _ sdk.ReadDTO) ([]map[string]interface{}, error) {
	core := GetRegistryCore()
	dict, err := core.Dictionary(h.session)
	if err != nil {
		return nil, err
	}
	result := make([]map[string]interface{}, 0, len(dict))
	for _, entry := range dict {
		data, err := json.Marshal(entry.entity)
		if err != nil {
			return nil, err
		}
		var m map[string]interface{}
		if err := json.Unmarshal(data, &m); err != nil {
			return nil, err
		}
		result = append(result, m)
	}
	return result, nil
}

// EndorHandlerList returns all registered EndorHandlers.
// Used by the server to register routes and swagger configuration.
func (h *EndorHandlerRepository) EndorHandlerList() ([]sdk.EndorHandler, error) {
	return GetRegistryCore().endorHandlerList()
}

// ActionRepository returns an EndorHandlerActionRepository backed by the same core.
func (h *EndorHandlerRepository) ActionRepository() *EndorHandlerActionRepository {
	return NewEndorHandlerActionRepository(GetRegistryCore())
}

// #endregion
