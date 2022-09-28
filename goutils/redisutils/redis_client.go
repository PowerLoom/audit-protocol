package redisutils

import (
	"context"

	"github.com/go-redis/redis/v8"
	log "github.com/sirupsen/logrus"
)

func InitRedisClient(redisURL string, redisDb int, poolSize int) *redis.Client {

	log.Info("Connecting to redis at:", redisURL)
	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisURL,
		Password: "",
		DB:       redisDb,
		PoolSize: poolSize,
	})
	pong, err := redisClient.Ping(context.Background()).Result()
	if err != nil {
		log.Error("Unable to connect to redis at:")
	}
	log.Info("Connected successfully to Redis and received ", pong, " back")
	return redisClient
}
