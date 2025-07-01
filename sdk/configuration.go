package sdk

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type ServerConfig struct {
	ServerPort                 string
	EndorServiceDBUri          string
	EndorServiceServiceEnabled bool
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
	EndorServiceServiceEnabledStr, exists := os.LookupEnv("ENDOR_RESOURCE_SERVICE_ENABLED")
	EndorServiceServiceEnabled := true
	if !exists || EndorServiceServiceEnabledStr == "" || EndorServiceServiceEnabledStr == "false" {
		EndorServiceServiceEnabled = false
	}

	return ServerConfig{
		ServerPort:                 port,
		EndorServiceDBUri:          EndorServiceDBUri,
		EndorServiceServiceEnabled: EndorServiceServiceEnabled,
	}
}
