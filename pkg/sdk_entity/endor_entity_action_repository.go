package sdk_entity

import (
	"fmt"

	"github.com/mattiabonardi/endor-sdk-go/pkg/sdk"
)

// EndorHandlerActionRepository provides action-level access to the handler registry.
// It delegates all resolution to RegistryCore, passing the session so the core
// can transparently choose the production or ephemeral dev dictionary.
type EndorHandlerActionRepository struct {
	core *RegistryCore
}

// NewEndorHandlerActionRepository creates an EndorHandlerActionRepository backed by the given core.
func NewEndorHandlerActionRepository(core *RegistryCore) *EndorHandlerActionRepository {
	return &EndorHandlerActionRepository{core: core}
}

// #region Entity Action CRUD

func (r *EndorHandlerActionRepository) DictionaryActionMap(session sdk.Session) (map[string]EndorHandlerActionDictionary, error) {
	dict, err := r.core.Dictionary(session)
	if err != nil {
		return nil, err
	}
	actions := make(map[string]EndorHandlerActionDictionary)
	for entityName, entity := range dict {
		for actionName, endorHandlerAction := range entity.EndorHandler.Actions {
			action, err := r.core.createAction(entityName, entity.EndorHandler.Version, actionName, endorHandlerAction)
			if err == nil {
				actions[action.entityAction.ID] = *action
			}
		}
	}
	return actions, nil
}

func (r *EndorHandlerActionRepository) DictionaryActionInstance(session sdk.Session, dto sdk.ReadInstanceDTO) (*EndorHandlerActionDictionary, error) {
	actions, err := r.DictionaryActionMap(session)
	if err != nil {
		return nil, err
	}
	if action, ok := actions[dto.Id]; ok {
		return &action, nil
	}
	return nil, sdk.NewNotFoundError(fmt.Errorf("entity action not found")).WithTranslation("entities.entity.action_not_found", nil)
}

func (r *EndorHandlerActionRepository) EntityActionList(session sdk.Session) ([]sdk.EntityAction, error) {
	actions, err := r.DictionaryActionMap(session)
	if err != nil {
		return []sdk.EntityAction{}, err
	}
	actionList := make([]sdk.EntityAction, 0, len(actions))
	for _, action := range actions {
		actionList = append(actionList, action.entityAction)
	}
	return actionList, nil
}

// #endregion
