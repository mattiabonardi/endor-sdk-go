package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mattiabonardi/endor-sdk-go/configuration"
	"github.com/mattiabonardi/endor-sdk-go/models"
)

// get access token and verify user session
func AuthorizeResource(c *gin.Context, resource string, methodKey string) (models.Session, error) {
	config := configuration.LoadConfiguration()
	app := c.Param("app")
	userSession := models.Session{}

	if config.Env == "DEVELOPMENT" {
		// create dummy userSession
		userSession.Id = uuid.New().String()
		userSession.User = "659f27cce7fd9277b3cc4ef7"
		userSession.Email = "endor@endor.com"
		userSession.App = app
		return userSession, nil
	}
	// read the session id cookie
	cookie, err := c.Request.Cookie("sessionId")
	if err != nil {
		return userSession, err
	}
	// request authorization to identity provider
	payload := []byte(fmt.Sprintf(`{"path": "%s"}`, fmt.Sprintf("%s/%s", resource, methodKey)))
	path := fmt.Sprintf("%s/api/%s/v1/authentication/authorize", config.EndorAuthenticationServiceUrl, app)
	request, err := http.NewRequest("POST", path, bytes.NewBuffer(payload))
	if err != nil {
		return userSession, err
	}
	request.Header.Set("Content-Type", "application/json")
	request.AddCookie(cookie)

	client := http.DefaultClient

	response, err := client.Do(request)
	if err != nil {
		return userSession, err
	}
	defer response.Body.Close()

	// Read the response body
	jsonData, err := io.ReadAll(response.Body)
	if err != nil {
		return userSession, err
	}
	r := models.Response[map[string]string]{}
	err = json.Unmarshal([]byte(jsonData), &r)
	if err != nil {
		return userSession, err
	}
	return userSession, nil
}
