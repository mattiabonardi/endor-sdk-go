package sdk

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type ServerConfig struct {
	ServerPort        string
	EndorServiceDBUri string
	// @default false
	EndorResourceServiceEnabled bool
	// @default false
	EndorDynamicResourcesEnabled bool
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
	EndorServiceDBUri, exists := os.LookupEnv("ENDOR_RESOURCE_DB_URI")
	if !exists || EndorServiceDBUri == "" {
		EndorServiceDBUri = "mongodb://localhost:27017"
	}
	EndorResourceServiceEnabledStr, exists := os.LookupEnv("ENDOR_RESOURCE_SERVICE_ENABLED")
	EndorResourceServiceEnabled := true
	if !exists || EndorResourceServiceEnabledStr == "" || EndorResourceServiceEnabledStr == "false" {
		EndorResourceServiceEnabled = false
	}
	EndorDynamicResourcesEnabledStr, exists := os.LookupEnv("ENDOR_DYNAMIC_RESOURCES_ENABLED")
	EndorDynamicResourcesEnabled := true
	if !exists || EndorDynamicResourcesEnabledStr == "" || EndorDynamicResourcesEnabledStr == "false" {
		EndorDynamicResourcesEnabled = false
	}

	return ServerConfig{
		ServerPort:                   port,
		EndorServiceDBUri:            EndorServiceDBUri,
		EndorResourceServiceEnabled:  EndorResourceServiceEnabled,
		EndorDynamicResourcesEnabled: EndorDynamicResourcesEnabled,
	}
}
