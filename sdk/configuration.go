package sdk

import (
	"fmt"
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

	return &ServerConfig{
		ServerPort:              port,
		DocumentDBUri:           dbUri,
		HybridResourcesEnabled:  hybridResourcesEnabled,
		DynamicResourcesEnabled: dynamicResourcesEnabled,
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

// Interface implementation methods for ConfigProviderInterface
// These methods implement the interfaces.ConfigProviderInterface contract

// GetServerPort returns the HTTP server port configuration.
func (c *ServerConfig) GetServerPort() string {
	return c.ServerPort
}

// GetDocumentDBUri returns the MongoDB connection URI.
func (c *ServerConfig) GetDocumentDBUri() string {
	return c.DocumentDBUri
}

// IsHybridResourcesEnabled returns whether hybrid resource functionality is enabled.
func (c *ServerConfig) IsHybridResourcesEnabled() bool {
	return c.HybridResourcesEnabled
}

// IsDynamicResourcesEnabled returns whether dynamic resource functionality is enabled.
func (c *ServerConfig) IsDynamicResourcesEnabled() bool {
	return c.DynamicResourcesEnabled
}

// GetDynamicResourceDocumentDBName returns the database name for dynamic resources.
func (c *ServerConfig) GetDynamicResourceDocumentDBName() string {
	return c.DynamicResourceDocumentDBName
}

// Reload forces configuration reload from sources.
func (c *ServerConfig) Reload() error {
	// Reload configuration from environment variables
	newConfig := loadConfiguration()

	// Update current instance fields
	c.ServerPort = newConfig.ServerPort
	c.DocumentDBUri = newConfig.DocumentDBUri
	c.HybridResourcesEnabled = newConfig.HybridResourcesEnabled
	c.DynamicResourcesEnabled = newConfig.DynamicResourcesEnabled
	c.DynamicResourceDocumentDBName = newConfig.DynamicResourceDocumentDBName

	return nil
}

// Validate performs configuration validation.
func (c *ServerConfig) Validate() error {
	if c.ServerPort == "" {
		return fmt.Errorf("server port cannot be empty")
	}
	if c.DocumentDBUri == "" {
		return fmt.Errorf("document DB URI cannot be empty")
	}
	return nil
}
