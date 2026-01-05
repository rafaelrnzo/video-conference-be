package utility

import (
	"context"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client

func InitRedis() {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     Config.RedisAddr,
		Password: Config.RedisPassword, // no password set
		DB:       0,                    // use default DB
	})

	if err := RedisClient.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("failed to connect to redis: %v", err)
	}

	fmt.Println("Connected to Redis at", Config.RedisAddr)
}
