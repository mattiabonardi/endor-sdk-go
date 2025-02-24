package main

import (
	"github.com/mattiabonardi/endor-sdk-go/models"
	"github.com/mattiabonardi/endor-sdk-go/server"
)

func main() {
	var handlers []models.Handler
	server.Init(handlers)
}
