package utility

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type AppConfig struct {
	Port          string
	DBDSN         string
	JWTSecret     string
	LivekitAPIKey string
	LivekitSecret string
	LivekitURL    string
}

var Config AppConfig

func LoadConfig() {
	_ = godotenv.Load()

	Config = AppConfig{
		Port:          getEnv("PORT", "4000"),
		DBDSN:         getEnv("DB_DSN", "host=localhost port=5432 user=postgres password=postgres dbname=postgres sslmode=disable TimeZone=Asia/Jakarta"),
		JWTSecret:     mustEnv("JWT_SECRET"),
		LivekitAPIKey: mustEnv("LIVEKIT_API_KEY"),
		LivekitSecret: mustEnv("LIVEKIT_API_SECRET"),
		LivekitURL:    mustEnv("LIVEKIT_SERVER_URL"),
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("missing env: %s", key)
	}
	return v
}
