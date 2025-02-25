package middlewares

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mattiabonardi/endor-sdk-go/configuration"
	e "github.com/mattiabonardi/endor-sdk-go/errors"
	"github.com/mattiabonardi/endor-sdk-go/managers"
	"github.com/mattiabonardi/endor-sdk-go/models"
)

// get access token and verify user session
func AuthorizationMiddleware(c *gin.Context) {
	config := configuration.Load()

	if config.Env == "DEVELOPMENT" {
		// create dummy userSession
		userSession := models.Session{
			Id:    uuid.New().String(),
			User:  "659f27cce7fd9277b3cc4ef7",
			Email: "endor@endor.com",
			App:   "",
		}
		// set token data to context
		c.Set(models.USER_SESSION_CONTEXT_KEY, managers.SessionToMap(userSession))
		c.Next()
		return
	}

	// read the session id cookie
	cookie, err := c.Request.Cookie("sessionId")
	if err != nil {
		e.ThrowUnauthorize(c, err)
		return
	}
	// request authorization to identity provider
	payload := []byte(fmt.Sprintf(`{"path": "%s"}`, c.FullPath()))
	path := fmt.Sprintf("%s/api/v1/authentication/authorize", config.EndorAuthenticationServiceUrl)
	request, err := http.NewRequest("POST", path, bytes.NewBuffer(payload))
	if err != nil {
		e.ThrowInternalServerError(c, err)
		return
	}
	request.Header.Set("Content-Type", "application/json")
	request.AddCookie(cookie)

	client := http.DefaultClient

	response, err := client.Do(request)
	if err != nil {
		e.ThrowInternalServerError(c, err)
		return
	}
	defer response.Body.Close()

	// Read the response body
	jsonData, err := io.ReadAll(response.Body)
	if err != nil {
		e.ThrowInternalServerError(c, err)
		return
	}
	r := models.Response[models.Session]{}
	err = json.Unmarshal([]byte(jsonData), &r)
	if err != nil {
		e.ThrowInternalServerError(c, err)
		return
	}
	// return user session to next route
	c.Set(models.USER_SESSION_CONTEXT_KEY, r.Data)
	c.Next()
}
