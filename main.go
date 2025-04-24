package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/mattiabonardi/endor-sdk-go/internal"
	"github.com/mattiabonardi/endor-sdk-go/internal/handler"
)

type CreateUserPayload struct {
	Username string `json:"username"`
}

type LoginPayload struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func main() {
	services := []internal.EndorService{
		{
			Resource: "users",
			Version:  "v2",
			Apps:     []string{"app1", "app2"},
			Methods: map[string]internal.EndorServiceMethod{
				"create": internal.NewMethod(
					handler.AuthorizationHandler,
					handle,
				),
				"login": internal.NewMethod(
					handler.ValidationHandler,
					func(c *internal.EndorContext[LoginPayload]) {
						fmt.Println("Login:", c.Payload.Email)
						c.End(gin.H{"status": "login success"})
					},
				),
			},
		},
	}
	internal.Init("me-example", services)
}

func handle(c *internal.EndorContext[CreateUserPayload]) {
	fmt.Println("Create:", c.Payload.Username)
	c.End("User created")
}
