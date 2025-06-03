package sdk

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type ServerConfig struct {
	ServerPort                  string
	EndorResourceDBUri          string
	EndorResourceServiceEnabled bool
}

func LoadConfiguration() ServerConfig {
	err := godotenv.Load(".env")
	if err != nil {
		log.Printf("Error loading .env file: %s. Ignore this in production.", err)
	}

	port, exists := os.LookupEnv("PORT")
	if !exists || port == "" {
		port = "8080"
	}
	endorResourceDBUri, exists := os.LookupEnv("ENDOR_RESOURCE_DB_URI")
	if !exists || endorResourceDBUri == "" {
		endorResourceDBUri = "mongodb://localhost:27017"
	}
	endorResourceServiceEnabledStr, exists := os.LookupEnv("ENDOR_RESOURCE_SERVICE_ENABLED")
	endorResourceServiceEnabled := true
	if !exists || endorResourceServiceEnabledStr == "" || endorResourceServiceEnabledStr == "false" {
		endorResourceServiceEnabled = false
	}

	return ServerConfig{
		ServerPort:                  port,
		EndorResourceDBUri:          endorResourceDBUri,
		EndorResourceServiceEnabled: endorResourceServiceEnabled,
	}
}
