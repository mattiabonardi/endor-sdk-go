package managers

import "github.com/mattiabonardi/endor-sdk-go/models"

func SessionToMap(session models.Session) map[string]any {
	sessionMap := make(map[string]any)
	sessionMap["id"] = session.Id
	sessionMap["userId"] = session.UserId
	sessionMap["email"] = session.Email
	sessionMap["appId"] = session.AppId
	return sessionMap
}

func SessionFromMap(sessionMap map[string]any) models.Session {
	session := models.Session{}
	session.Id = sessionMap["id"].(string)
	session.UserId = sessionMap["userId"].(string)
	session.Email = sessionMap["email"].(string)
	session.AppId = sessionMap["appId"].(string)
	return session
}
