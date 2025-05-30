package sdk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"
)

// get access token and verify user session
func AuthorizationHandler[T any](c *EndorContext[T]) {
	config := LoadConfiguration()

	if config.Env == "DEVELOPMENT" {
		// create dummy userSession
		c.Session = Session{
			Id:       uuid.New().String(),
			Username: "endor",
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
	path := fmt.Sprintf("%s/api/v1/authentication/authorize", config.EndorAuthenticationServiceUrl)
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
	r := Response[Session]{}
	err = json.Unmarshal([]byte(jsonData), &r)
	if err != nil {
		c.InternalServerError(err)
		return
	}
	c.Session = *r.Data
	c.Next()
}
