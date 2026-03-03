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

	// MINIO / S3 utk egress
	MinioEndpoint string
	MinioBaseURL  string
	MinioAccess   string
	MinioSecret   string
	MinioBucket   string
	MinioRegion   string

	// Redis
	RedisAddr     string
	RedisPassword string

	// Recorder Service
	RecorderSvcURL string
}

var Config AppConfig

func LoadConfig() {
	dir, _ := os.Getwd()
	log.Printf("Current working directory: %s", dir)

	if err := godotenv.Load(); err != nil {
		log.Printf("Error loading .env file: %v", err)

		_ = godotenv.Load("../../.env")
	}

	Config = AppConfig{
		Port:          getEnv("PORT", "4000"),
		DBDSN:         getEnv("DB_DSN", "host=localhost port=5432 user=postgres password=postgres dbname=postgres sslmode=disable TimeZone=Asia/Jakarta"),
		JWTSecret:     mustEnv("JWT_SECRET"),
		LivekitAPIKey: mustEnv("LIVEKIT_API_KEY"),
		LivekitSecret: mustEnv("LIVEKIT_API_SECRET"),
		LivekitURL:    mustEnv("LIVEKIT_SERVER_URL"),
		MinioBaseURL:  mustEnv("MINIO_ENDPOINT"),

		MinioEndpoint: mustEnv("MINIO_ENDPOINT"),   // contoh: http://minio-pusbahasa.10.70.0.45.nip.io
		MinioAccess:   mustEnv("MINIO_ACCESS_KEY"), // admin
		MinioSecret:   mustEnv("MINIO_SECRET_KEY"), // Falah0918PRGM
		MinioBucket:   mustEnv("MINIO_BUCKET"),     // livekit-egress
		MinioRegion:   getEnv("MINIO_REGION", "us-east-1"),

		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),

		RecorderSvcURL: getEnv("RECORDER_SVC_URL", "http://localhost:4000"),
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
