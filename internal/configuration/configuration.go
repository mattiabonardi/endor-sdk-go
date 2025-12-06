package configuration

import (
	"log"
	"os"
	"sync"

	"github.com/joho/godotenv"
)

type ServerConfig struct {
	ServerPort                    string
	DocumentDBUri                 string
	HybridResourcesEnabled        bool
	DynamicResourcesEnabled       bool
	DynamicResourceDocumentDBName string
	LogType                       string
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
	dbUri := getEnv("DOCUMENT_DB_URI", "mongodb://localhost:27017")

	hybridResourcesEnabled := getEnvAsBool("HYBRID_RESOURCES_ENABLED", false)
	dynamicResourcesEnabled := getEnvAsBool("DYNAMIC_RESOURCES_ENABLED", false)
	logType := getEnv("LOG_TYPE", "JSON")

	return &ServerConfig{
		ServerPort:              port,
		DocumentDBUri:           dbUri,
		HybridResourcesEnabled:  hybridResourcesEnabled,
		DynamicResourcesEnabled: dynamicResourcesEnabled,
		LogType:                 logType,
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
