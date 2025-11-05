package sdk

import (
	"log"
	"os"
	"sync"

	"github.com/joho/godotenv"
)

type ServerConfig struct {
	ServerPort                   string
	EndorServiceDBUri            string
	EndorResourceServiceEnabled  bool
	EndorDynamicResourcesEnabled bool
	EndorDynamicResourceDBName   string
}

// Variabili globali per il singleton
var (
	instance *ServerConfig
	once     sync.Once
)

// GetConfig restituisce l'istanza singleton di ServerConfig
func GetConfig() *ServerConfig {
	once.Do(func() {
		instance = loadConfiguration()
	})
	return instance
}

// funzione privata per caricare la configurazione
func loadConfiguration() *ServerConfig {
	err := godotenv.Load(".env")
	if err != nil {
		log.Printf("Error loading .env file: %s. Ignore this in production.", err)
	}

	port := getEnv("PORT", "8080")
	dbUri := getEnv("ENDOR_RESOURCE_DB_URI", "mongodb://localhost:27017")

	EndorResourceServiceEnabled := getEnvAsBool("ENDOR_RESOURCE_SERVICE_ENABLED", false)
	EndorDynamicResourcesEnabled := getEnvAsBool("ENDOR_DYNAMIC_RESOURCES_ENABLED", false)

	return &ServerConfig{
		ServerPort:                   port,
		EndorServiceDBUri:            dbUri,
		EndorResourceServiceEnabled:  EndorResourceServiceEnabled,
		EndorDynamicResourcesEnabled: EndorDynamicResourcesEnabled,
	}
}

// Helpers
func getEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}
	return defaultVal
}

func getEnvAsBool(key string, defaultVal bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		if value == "true" || value == "1" {
			return true
		}
		return false
	}
	return defaultVal
}
