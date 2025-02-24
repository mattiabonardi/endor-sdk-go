package managers

import "github.com/mattiabonardi/endor-sdk-go/models"

func SessionToMap(session models.Session) map[string]any {
	sessionMap := make(map[string]any)
	sessionMap["id"] = session.Id
	sessionMap["user"] = session.User
	sessionMap["email"] = session.Email
	sessionMap["app"] = session.App
	return sessionMap
}

func SessionFromMap(sessionMap map[string]any) models.Session {
	session := models.Session{}
	session.Id = sessionMap["id"].(string)
	session.User = sessionMap["user"].(string)
	session.Email = sessionMap["email"].(string)
	session.App = sessionMap["app"].(string)
	return session
}
