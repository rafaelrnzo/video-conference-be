package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port         string
	KeycloakURL  string
	Realm        string
	ClientID     string
	ClientSecret string
	FrontendURL  string
}

func LoadConfig() *Config {
	_ = godotenv.Load()

	return &Config{
		Port:         getEnv("PORT", "8080"),
		KeycloakURL:  getEnv("KEYCLOAK_URL", "http://localhost:8090"),
		Realm:        getEnv("KEYCLOAK_REALM", "Meet"),
		ClientID:     getEnv("KEYCLOAK_CLIENT_ID", "meet-backend"),
		ClientSecret: getEnv("KEYCLOAK_CLIENT_SECRET", "3AoWIvt3azB1CvbE4P0swb5HpICcobzH"),
		FrontendURL:  getEnv("FRONTEND_URL", "http://localhost:3000"),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
