package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"
	"github.com/mattiabonardi/endor-sdk-go/sdk"
)

// get access token and verify user session
func AuthorizationHandler[T any](c *sdk.EndorContext[T]) {
	config := sdk.LoadConfiguration()
	app := c.Session.App

	if config.Env == "DEVELOPMENT" {
		// create dummy userSession
		c.Session = sdk.Session{
			Id:    uuid.New().String(),
			User:  "659f27cce7fd9277b3cc4ef7",
			Email: "endor@endor.com",
			App:   app,
		}
		c.Next()
		return
	}

	// read the session id cookie
	cookie, err := c.GinContext.Request.Cookie("sessionId")
	if err != nil {
		c.Unauthorize(err)
		return
	}
	// request authorization to identity provider
	payload := []byte(fmt.Sprintf(`{"path": "%s"}`, c.GinContext.FullPath()))
	path := fmt.Sprintf("%s/api/%s/v1/authentication/authorize", config.EndorAuthenticationServiceUrl, app)
	request, err := http.NewRequest("POST", path, bytes.NewBuffer(payload))
	if err != nil {
		c.InternalServerError(err)
		return
	}
	request.Header.Set("Content-Type", "application/json")
	request.AddCookie(cookie)

	client := http.DefaultClient

	response, err := client.Do(request)
	if err != nil {
		c.InternalServerError(err)
		return
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode > 299 {
		c.Unauthorize(fmt.Errorf("Unhauthorized"))
		return
	}

	// Read the response body
	jsonData, err := io.ReadAll(response.Body)
	if err != nil {
		c.InternalServerError(err)
		return
	}
	r := sdk.Response[sdk.Session]{}
	err = json.Unmarshal([]byte(jsonData), &r)
	if err != nil {
		c.InternalServerError(err)
		return
	}
	c.Session = r.Data
	c.Next()
}
