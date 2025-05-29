package sdk

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type ServerConfig struct {
	ServerDNS                     string
	ServerPort                    string
	EndorAuthenticationServiceUrl string
	EndorResourceDBUri            string
	EndorResourceServiceEnabled   bool
	Env                           string
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
	endorAuthenticationServiceUrl, exists := os.LookupEnv("ENDOR_AUTHENTICATION_SERVICE_URL")
	if !exists || endorAuthenticationServiceUrl == "" {
		endorAuthenticationServiceUrl = "http://localhost:8000"
	}
	endorResourceDBUri, exists := os.LookupEnv("ENDOR_RESOURCE_DB_URI")
	if !exists || endorResourceDBUri == "" {
		endorResourceDBUri = "mongodb://localhost:27017"
	}
	endorResourceServiceEnabledStr, exists := os.LookupEnv("ENDOR_RESOURCE_SERVICE_ENBALED")
	endorResourceServiceEnabled := true
	if !exists || endorResourceServiceEnabledStr == "" || endorResourceServiceEnabledStr == "false" {
		endorResourceServiceEnabled = false
	}

	return ServerConfig{
		ServerPort:                    port,
		ServerDNS:                     os.Getenv("DNS"),
		EndorAuthenticationServiceUrl: endorAuthenticationServiceUrl,
		EndorResourceDBUri:            endorResourceDBUri,
		Env:                           os.Getenv("ENV"),
		EndorResourceServiceEnabled:   endorResourceServiceEnabled,
	}
}
