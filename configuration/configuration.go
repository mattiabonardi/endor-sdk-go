package configuration

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type ServerConfig struct {
	ServerDNS                     string
	ServerPort                    string
	EndorAuthenticationServiceUrl string
	Env                           string
}

func Load() ServerConfig {
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

	return ServerConfig{
		ServerPort:                    port,
		ServerDNS:                     os.Getenv("DNS"),
		EndorAuthenticationServiceUrl: endorAuthenticationServiceUrl,
		Env:                           os.Getenv("ENV"),
	}
}
