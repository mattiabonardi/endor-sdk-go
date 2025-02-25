package managers

import "github.com/mattiabonardi/endor-sdk-go/models"

func SessionToMap(session models.Session) map[string]string {
	sessionMap := make(map[string]string)
	sessionMap["id"] = session.Id
	sessionMap["user"] = session.User
	sessionMap["email"] = session.Email
	sessionMap["app"] = session.App
	return sessionMap
}

func SessionFromMap(sessionMap map[string]string) models.Session {
	session := models.Session{}
	session.Id = sessionMap["id"]
	session.User = sessionMap["user"]
	session.Email = sessionMap["email"]
	session.App = sessionMap["app"]
	return session
}
