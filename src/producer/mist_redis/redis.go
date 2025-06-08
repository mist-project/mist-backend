package mist_redis

import (
	"fmt"
	"log"
	"mist/src/logging/logger"
	"os"
	"strconv"

	"github.com/redis/go-redis/v9"
)

var (
	username     = os.Getenv("REDIS_USERNAME")
	password     = os.Getenv("REDIS_PASSWORD")
	hostname     = os.Getenv("REDIS_HOSTNAME")
	port         = os.Getenv("REDIS_PORT")
	eventChannel = os.Getenv("REDIS_NOTIFICATION_CHANNEL")
)

func ConnectToRedis(dbStr string) *redis.Client {
	var client *redis.Client

	if eventChannel == "" {
		log.Fatal("REDIS_NOTIFICATION_CHANNEL environment variable is not set")
	}

	if username == "" || password == "" || hostname == "" || port == "" || dbStr == "" {
		log.Fatal("Redis connection details are not set in environment variables")
	}
	db, err := strconv.Atoi(dbStr)
	if err != nil {
		log.Fatalf("Invalid REDIS_DB value: %v", err)
	}

	for client == nil {
		logger.Debug("Initializing Redis client.", "SERVICE", "REDIS")
		client = redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%s", hostname, port),
			Username: username,
			Password: password,
			DB:       db, // use default DB
		})
	}

	return client
}
