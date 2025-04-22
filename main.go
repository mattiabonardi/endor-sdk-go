package main

import (
	"github.com/mattiabonardi/endor-sdk-go/models"
)

type Method1DTO struct {
	Data string
}

func main() {
	services := []models.EndorService{}
	userService := models.NewEndorService("users")

	userService.Handle("method1",
		func(c *models.EndorContext) {
			c.Data["name"] = "Endor"
			c.Next()
		},
		func(c *models.EndorContext) {
			name := c.Data["name"].(string)
			c.End(name)
		},
	)

	services = append(services, *userService)

	Init("me-example", services)
}
