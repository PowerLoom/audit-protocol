package redisutils

import (
	"context"
	"net"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	log "github.com/sirupsen/logrus"
)

func InitRedisClient(redisHost string, port int, redisDb int, poolSize int, password string, timeout time.Duration) *redis.Client {
	redisURL := net.JoinHostPort(redisHost, strconv.Itoa(port))

	log.Info("connecting to redis at:", redisURL)

	redisClient := redis.NewClient(&redis.Options{
		Addr:        redisURL,
		Password:    password,
		DB:          redisDb,
		PoolSize:    poolSize,
		ReadTimeout: timeout,
	})

	pong, err := redisClient.Ping(context.Background()).Result()
	if err != nil {
		log.WithField("addr", redisURL).Fatal("Unable to connect to redis")
	}

	log.Info("Connected successfully to Redis and received ", pong, " back")

	return redisClient
}
